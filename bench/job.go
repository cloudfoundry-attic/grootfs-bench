package bench

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
)

type Job struct {
	Runner commandrunner.CommandRunner

	// Path to grootfs binary
	GrootFSBinPath string

	// The GrootFS store path, where blobs/bundles/cache are stored
	StorePath string

	// The image to be downloaded
	Image string

	// Run benchmark using quotas
	UseQuota bool

	// The number of concurrent workers to run
	Concurrency int

	// The total number of bundles to be created
	TotalBundles int

	// Hold the results for each iteration
	results chan *Result
}

// Run is a blocking operation. It blocks until all concurrent cmd execution is
// completed.
func (j *Job) Run(printer Printer) {
	if j.Concurrency == 0 {
		j.Concurrency = runtime.NumCPU()
	}

	j.results = make(chan *Result, j.TotalBundles)

	start := time.Now()
	j.runWorkers()
	totalDuration := time.Since(start)

	close(j.results)
	summary := SummarizeResults(totalDuration, j.Concurrency, j.UseQuota, j.results)
	fmt.Fprint(os.Stdout, string(printer.Print(summary)))
}

func (j *Job) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(j.TotalBundles)

	for i := 0; i < j.Concurrency; i++ {
		go func(conc int) {
			for i := 0; i < j.TotalBundles/j.Concurrency; i++ {
				defer wg.Done()
				j.run(conc)
			}
		}(i)
	}

	// Handle some leftovers (i.e: 11 % 3)
	for i := 0; i < j.TotalBundles%j.Concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			j.run(i)
		}(i)
	}

	wg.Wait()
}

func (j *Job) run(i int) {
	start := time.Now()
	cmd := j.grootfsCmd(i)

	err := j.Runner.Run(cmd)
	j.results <- &Result{
		Err:      err,
		Duration: time.Since(start),
	}
}

func (j *Job) grootfsCmd(workerId int) *exec.Cmd {
	args := []string{
		"--store",
		j.StorePath,
		"create",
	}

	if j.UseQuota {
		args = append(args, "--disk-limit-size-bytes", "1019430400")
	}

	args = append(args,
		j.Image,
		fmt.Sprintf("image-%d-%d", workerId, time.Now().UnixNano()),
	)

	return exec.Command(j.GrootFSBinPath, args...)
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
	BundlesPerSecond     float64       `json:"bundles_per_second"`
	RanWithQuota         bool          `json:"ran_with_quota"`
	AverageTimePerBundle float64       `json:"average_time_per_bundle"`
	TotalErrorsAmt       int           `json:"total_errors_amt"`
	ErrorRate            float64       `json:"error_rate"`
	TotalBundles         int           `json:"total_bundles"`
	ConcurrencyFactor    int           `json:"concurrency_factor"`
}

func SummarizeResults(totalDuration time.Duration, concurrency int, useQuota bool, results chan *Result) Summary {
	summary := Summary{
		ConcurrencyFactor: concurrency,
		TotalDuration:     totalDuration,
		RanWithQuota:      useQuota,
	}

	averageTimePerBundle := 0.0
	for res := range results {
		summary.TotalBundles++

		if res.Err != nil {
			summary.TotalErrorsAmt++
		} else {
			averageTimePerBundle += res.Duration.Seconds()
		}
	}

	createdBundles := float64(summary.TotalBundles - summary.TotalErrorsAmt)
	summary.BundlesPerSecond = createdBundles / totalDuration.Seconds()
	summary.ErrorRate = float64(summary.TotalErrorsAmt*100) / float64(summary.TotalBundles)
	summary.AverageTimePerBundle = averageTimePerBundle / createdBundles
	summary.TotalDuration = totalDuration

	return summary
}
