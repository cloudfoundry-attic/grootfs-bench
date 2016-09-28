package grootfspool

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner"
)

type Pool struct {
	cmdRunner  commandrunner.CommandRunner
	grootfsBin string
	storePath  string
	image      string

	errorsChan   chan error
	durationChan chan time.Duration

	concurrency int
	wg          sync.WaitGroup

	startedAt time.Time
}

func New(cmdRunner commandrunner.CommandRunner, grootfsBin, storePath, image string, concurrency int) *Pool {
	pool := &Pool{
		cmdRunner:   cmdRunner,
		grootfsBin:  grootfsBin,
		storePath:   storePath,
		image:       image,
		concurrency: concurrency,
	}
	pool.wg.Add(concurrency)

	return pool
}

func (p *Pool) Wait() {
	p.wg.Wait()
	close(p.durationChan)
	close(p.errorsChan)
}

func (p *Pool) Start(totalBundles int) chan int {
	p.durationChan = make(chan time.Duration, totalBundles)
	p.errorsChan = make(chan error, totalBundles)
	jobs := make(chan int, totalBundles)

	for w := 0; w < p.concurrency; w++ {
		go p.worker(w, jobs, p.durationChan, p.errorsChan)
	}

	return jobs
}

func (p *Pool) DurationChan() chan time.Duration {
	return p.durationChan
}

func (p *Pool) ErrorsChan() chan error {
	return p.errorsChan
}

func (p *Pool) worker(workerId int, jobs <-chan int, results chan time.Duration, errors chan error) {
	defer p.wg.Done()

	for i := range jobs {
		start := time.Now()
		outputBuffer := bytes.NewBuffer([]byte{})

		cmd := exec.Command(
			p.grootfsBin,
			"--store",
			p.storePath,
			"create",
			p.image,
			fmt.Sprintf("image-%d-%d-%d", workerId, i, time.Now().UnixNano()))

		cmd.Stderr = outputBuffer
		cmd.Stdout = outputBuffer
		err := p.cmdRunner.Run(cmd)

		duration := time.Since(start)
		if err != nil {
			errors <- fmt.Errorf("%s: %s", err, outputBuffer.String())
		} else {
			results <- duration
		}
	}
}
