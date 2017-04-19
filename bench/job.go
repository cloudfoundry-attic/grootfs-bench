package bench

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
)

type JobExecutor struct {
	Jobs []*Job
}

func (e *JobExecutor) Run() Summary {
	if len(e.Jobs) == 0 {
		return Summary{}
	}

	var wg sync.WaitGroup
	wg.Add(len(e.Jobs))

	summaryChannel := make(chan Summary, len(e.Jobs))
	doneChannel := make(chan bool)
	createdImagesChannel := make(chan string, e.Jobs[0].TotalImages)

	for _, job := range e.Jobs {
		job.Done = doneChannel
		job.Mutex = &sync.Mutex{}
		job.CreatedImages = createdImagesChannel
		go func(job *Job) {
			defer wg.Done()
			summary := job.Run()
			if job.Command == "create" {
				summaryChannel <- *summary
			}
		}(job)
	}

	wg.Wait()
	finalSummary := <-summaryChannel

	for _, job := range e.Jobs {
		if job.Command == "clean" {
			finalSummary.NumberOfCleans = job.RunCounter
		}

		if job.Command == "delete" {
			finalSummary.NumberOfDeletes = job.RunCounter
		}
	}
	return finalSummary
}

type Result struct {
	// Original error from grootfs if it occurrs
	Err error

	// Duration took by grootfs bin to run
	Duration time.Duration
}

// Summary represents some metrics while running grootfs with given input
type Summary struct {
	TotalDuration        time.Duration `json:"total_duration"`
	ImagesPerSecond      float64       `json:"images_per_second"`
	RanWithQuota         bool          `json:"ran_with_quota"`
	RanWithParallelClean bool          `json:"ran_with_parallel_clean"`
	NumberOfCleans       int           `json:"number_of_cleans"`
	NumberOfDeletes      int           `json:"number_of_deletes"`
	AverageTimePerImage  float64       `json:"average_time_per_image"`
	TotalErrorsAmt       int           `json:"total_errors_amt"`
	ErrorRate            float64       `json:"error_rate"`
	TotalImages          int           `json:"total_images"`
	ConcurrencyFactor    int           `json:"concurrency_factor"`
	ErrorMessages        []string      `json:"-"`
}

type Job struct {
	Runner         commandrunner.CommandRunner
	GrootFSBinPath string
	StorePath      string
	Driver         string
	LogLevel       string
	Command        string
	BaseImages     []string
	Interval       int
	UseQuota       bool
	Concurrency    int
	TotalImages    int
	CreatedImages  chan string
	Done           chan bool
	Results        chan *Result
	StartTime      time.Time
	Duration       time.Duration

	RunCounter int
	Mutex      *sync.Mutex
}

func (j *Job) Run() *Summary {
	if j.Concurrency == 0 {
		j.Concurrency = runtime.NumCPU()
	}

	j.StartTime = time.Now()
	if j.Command == "create" {
		j.runWorkers()
		close(j.Done)
	} else {
		j.runLoop(j.Done)
	}
	j.Duration = time.Since(j.StartTime)

	return j.summarizeResults()
}

func (j *Job) summarizeResults() *Summary {
	if j.Command != "create" {
		return nil
	}

	summary := Summary{
		ConcurrencyFactor: j.Concurrency,
		TotalDuration:     j.Duration,
		RanWithQuota:      j.UseQuota,
	}

	errors := []string{}

	averageTimePerImage := 0.0

	for res := range j.Results {
		summary.TotalImages++

		if res.Err != nil {
			summary.TotalErrorsAmt++
			errors = append(errors, fmt.Sprintf("could not create image %d: %s\n", summary.TotalImages, res.Err))
		} else {
			averageTimePerImage += res.Duration.Seconds()
		}
	}

	createdImages := float64(summary.TotalImages - summary.TotalErrorsAmt)
	summary.ImagesPerSecond = createdImages / j.Duration.Seconds()
	summary.ErrorRate = float64(summary.TotalErrorsAmt*100) / float64(summary.TotalImages)
	if createdImages == float64(0) {
		summary.AverageTimePerImage = float64(-1)
	} else {
		summary.AverageTimePerImage = averageTimePerImage / createdImages
	}
	summary.TotalDuration = j.Duration
	summary.ErrorMessages = errors
	return &summary
}

func (j *Job) runLoop(done chan bool) {
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				cmd := j.grootfsCmd("")
				if cmd != nil {
					j.runCommand(cmd)
					j.Mutex.Lock()
					j.RunCounter++
					j.Mutex.Unlock()
				}
				time.Sleep(time.Second * time.Duration(j.Interval))
			}
		}
	}()
}

func (j *Job) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(j.Concurrency)

	cmds := make(chan *exec.Cmd, j.TotalImages)
	n := 0
	for i := 0; i < j.TotalImages; i++ {
		n = i % len(j.BaseImages)
		cmd := j.grootfsCmd(j.BaseImages[n])
		if cmd != nil {
			cmds <- cmd
		}
	}
	close(cmds)

	j.Results = make(chan *Result, j.TotalImages)

	for i := 0; i < j.Concurrency; i++ {
		go func(number int) {
			defer wg.Done()
			for cmd := range cmds {
				j.runCommand(cmd)
			}
		}(i)
	}

	wg.Wait()

	close(j.Results)
}

func (j *Job) runCommand(cmd *exec.Cmd) {
	start := time.Now()

	buffer := bytes.NewBuffer([]byte{})
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	var cmdErr error
	if err := j.Runner.Run(cmd); err != nil {
		cmdErr = fmt.Errorf("%s, %s", err, buffer.String())
	}

	if j.Command == "create" {
		imageName := cmd.Args[len(cmd.Args)-1]
		j.CreatedImages <- imageName

		j.Results <- &Result{
			Err:      cmdErr,
			Duration: time.Since(start),
		}
	}
}

func (j *Job) grootfsCmd(baseImage string) *exec.Cmd {
	args := []string{
		"--store",
		j.StorePath,
		"--log-level",
		j.LogLevel,
	}

	if j.Driver != "" {
		args = append(args, "--driver", j.Driver)
	}

	args = append(args, j.Command)

	if j.Command == "create" {
		if j.UseQuota {
			args = append(args, "--disk-limit-size-bytes", "1019430400")
		}
		imageName := fmt.Sprintf("base-image-%d", time.Now().UnixNano())
		args = append(args,
			baseImage,
			imageName,
		)

	} else if j.Command == "delete" {
		imageName := <-j.CreatedImages
		if imageName == "" {
			return nil
		}
		args = append(args, imageName)
	}

	return exec.Command(j.GrootFSBinPath, args...)
}
