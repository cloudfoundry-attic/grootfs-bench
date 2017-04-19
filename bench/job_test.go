package bench_test

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"
	"code.cloudfoundry.org/grootfs-bench/bench"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job", func() {
	Describe("Run", func() {
		It("creates the number of images provided", func() {
			job := createJob()
			job.TotalImages = 11
			job.Concurrency = 3

			fakeCmdRunner := job.Runner.(*fake_command_runner.FakeCommandRunner)

			job.Run()

			executedCommands := fakeCmdRunner.ExecutedCommands()
			Expect(len(executedCommands)).To(Equal(11))

			for _, cmd := range executedCommands {
				Expect(cmd.Args[0]).To(Equal("/path/to/grootfs"))
				Expect(cmd.Args[1]).To(Equal("--store"))
				Expect(cmd.Args[2]).To(Equal("/store/path"))
				Expect(cmd.Args[3]).To(Equal("--log-level"))
				Expect(cmd.Args[4]).To(Equal("debug"))
				Expect(cmd.Args[5]).To(Equal("--driver"))
				Expect(cmd.Args[6]).To(Equal("btrfs"))
				Expect(cmd.Args[7]).To(Equal("create"))
				Expect(cmd.Args[8]).To(Equal("docker:///busybox"))
			}
		})

		Context("when UseQuota is true", func() {
			It("runs grootfs with --disk-limit-size-bytes flag", func() {
				job := createJob()
				job.TotalImages = 11
				job.Concurrency = 3
				job.UseQuota = true

				fakeCmdRunner := job.Runner.(*fake_command_runner.FakeCommandRunner)

				job.Run()

				executedCommands := fakeCmdRunner.ExecutedCommands()
				Expect(len(executedCommands)).To(Equal(11))

				for _, cmd := range fakeCmdRunner.ExecutedCommands() {
					Expect(cmd.Args[0]).To(Equal("/path/to/grootfs"))
					Expect(cmd.Args[1]).To(Equal("--store"))
					Expect(cmd.Args[2]).To(Equal("/store/path"))
					Expect(cmd.Args[3]).To(Equal("--log-level"))
					Expect(cmd.Args[4]).To(Equal("debug"))
					Expect(cmd.Args[5]).To(Equal("--driver"))
					Expect(cmd.Args[6]).To(Equal("btrfs"))
					Expect(cmd.Args[7]).To(Equal("create"))
					Expect(cmd.Args[8]).To(Equal("--disk-limit-size-bytes"))
					Expect(cmd.Args[9]).To(Equal("1019430400"))
					Expect(cmd.Args[10]).To(Equal("docker:///busybox"))
				}
			})
		})

		Context("when something fails", func() {
			var job *bench.Job

			BeforeEach(func() {
				job = createJob()
				job.TotalImages = 10

				fakeCmdRunner := job.Runner.(*fake_command_runner.FakeCommandRunner)
				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{}, func(cmd *exec.Cmd) error {
					cmd.Stderr.Write([]byte("groot failed to make a image"))
					return errors.New("exit status 1")
				})
			})

			It("returns the errors", func() {
				summary := job.Run()

				Expect(summary.ErrorMessages).To(HaveLen(10))
				for _, message := range summary.ErrorMessages {
					Expect(message).To(ContainSubstring("groot failed to make a image"))
				}
			})
		})

		Context("when not providing concurrency level", func() {
			It("sets the default to the # of cpus", func() {
				job := createJob()
				job.Concurrency = 0

				summary := job.Run()

				Expect(summary.ConcurrencyFactor).To(Equal(runtime.NumCPU()))
			})
		})
	})

	Describe("clean", func() {
		It("runs the clean command", func() {
			job := cleanJob()
			fakeCmdRunner := job.Runner.(*fake_command_runner.FakeCommandRunner)

			jobAssassin(job)
			job.Run()

			Eventually(fakeCmdRunner.ExecutedCommands, 5*time.Second).ShouldNot(BeEmpty())

			for _, cmd := range fakeCmdRunner.ExecutedCommands() {
				Expect(cmd.Args[0]).To(Equal("/path/to/grootfs"))
				Expect(cmd.Args[1]).To(Equal("--store"))
				Expect(cmd.Args[2]).To(Equal("/store/path"))
				Expect(cmd.Args[3]).To(Equal("--log-level"))
				Expect(cmd.Args[4]).To(Equal("debug"))
				Expect(cmd.Args[5]).To(Equal("--driver"))
				Expect(cmd.Args[6]).To(Equal("btrfs"))
				Expect(cmd.Args[7]).To(Equal("clean"))
			}
		})
	})

	Describe("Delete", func() {
		var (
			job           *bench.Job
			fakeCmdRunner *fake_command_runner.FakeCommandRunner
		)

		JustBeforeEach(func() {
			job = deleteJob()
			job.CreatedImages <- "image-0"
			job.CreatedImages <- "image-1"
			job.CreatedImages <- "image-2"
			job.CreatedImages <- "image-3"
			close(job.CreatedImages)

			fakeCmdRunner = job.Runner.(*fake_command_runner.FakeCommandRunner)
		})

		It("deletes the expected images", func() {
			jobAssassin(job)
			job.Run()

			Eventually(fakeCmdRunner.ExecutedCommands, 5*time.Second).Should(HaveLen(4))

			for i, cmd := range fakeCmdRunner.ExecutedCommands() {
				Expect(cmd.Args[0]).To(Equal("/path/to/grootfs"))
				Expect(cmd.Args[1]).To(Equal("--store"))
				Expect(cmd.Args[2]).To(Equal("/store/path"))
				Expect(cmd.Args[3]).To(Equal("--log-level"))
				Expect(cmd.Args[4]).To(Equal("debug"))
				Expect(cmd.Args[5]).To(Equal("--driver"))
				Expect(cmd.Args[6]).To(Equal("btrfs"))
				Expect(cmd.Args[7]).To(Equal("delete"))
				Expect(cmd.Args[8]).To(Equal(fmt.Sprintf("image-%d", i)))
			}
		})

		Describe("when the interval is specified", func() {
			It("deletes every n seconds", func() {
				job.Interval = 2

				Expect(fakeCmdRunner.ExecutedCommands()).Should(HaveLen(0))
				jobAssassin(job)
				job.Run()
				Eventually(fakeCmdRunner.ExecutedCommands, 5*time.Second).Should(HaveLen(3))
			})
		})
	})

	Describe("Job summaries", func() {
		It("returns the results summarized", func() {
			job := createJob()
			job.Concurrency = 2
			job.TotalImages = 2
			summary := job.Run()

			Expect(summary.TotalImages).To(Equal(2))
			Expect(summary.TotalDuration).To(BeNumerically("~", time.Second*20, time.Second*21))
			Expect(summary.ConcurrencyFactor).To(Equal(2))
		})

		Context("when there are 0 images created", func() {
			var job *bench.Job

			BeforeEach(func() {
				job = createJob()
				job.TotalImages = 10

				fakeCmdRunner := job.Runner.(*fake_command_runner.FakeCommandRunner)
				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{}, func(cmd *exec.Cmd) error {
					cmd.Stderr.Write([]byte("groot failed to make a image"))
					return errors.New("exit status 1")
				})
			})

			It("sets the average time per image to -1", func() {
				summary := job.Run()

				Expect(summary.AverageTimePerImage).To(Equal(float64(-1)))
			})
		})

		Context("when command fails", func() {
			var job *bench.Job

			BeforeEach(func() {
				job = createJob()
				job.TotalImages = 10

				fakeCmdRunner := job.Runner.(*fake_command_runner.FakeCommandRunner)
				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{}, func(cmd *exec.Cmd) error {
					cmd.Stderr.Write([]byte("groot failed to make a image"))
					return errors.New("exit status 1")
				})
			})

			It("returns the total errors", func() {
				summary := job.Run()
				Expect(summary.TotalErrorsAmt).To(Equal(10))
			})

			It("returns the error rate", func() {
				summary := job.Run()
				// 33.33 because we're creating 2 in the outer BeforeEach
				Expect(summary.ErrorRate).To(BeNumerically(">", 33.33))
			})

			It("sets RanWithQuota to true if quota was applied", func() {
				job.UseQuota = true
				summary := job.Run()
				Expect(summary.RanWithQuota).To(BeTrue())
			})
		})
	})
})

func createJob() *bench.Job {
	job := genericJob()
	job.Command = "create"
	return job
}

func deleteJob() *bench.Job {
	job := genericJob()
	job.Command = "delete"
	return job
}

func cleanJob() *bench.Job {
	job := genericJob()
	job.Command = "clean"
	return job
}

func genericJob() *bench.Job {
	return &bench.Job{
		Runner:         fake_command_runner.New(),
		GrootFSBinPath: "/path/to/grootfs",
		StorePath:      "/store/path",
		Driver:         "btrfs",
		LogLevel:       "debug",
		BaseImages:     []string{"docker:///busybox"},
		Interval:       1,
		Concurrency:    1,
		CreatedImages:  make(chan string, 100),
		Done:           make(chan bool),
		Mutex:          new(sync.Mutex),
	}
}

func jobAssassin(job *bench.Job) {
	go func() {
		time.Sleep(5 * time.Second)
		close(job.Done)
	}()
}
