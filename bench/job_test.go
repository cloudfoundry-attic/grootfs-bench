package bench_test

import (
	"errors"
	"os/exec"
	"runtime"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"
	"code.cloudfoundry.org/grootfs-bench/bench"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job", func() {
	var fakeCmdRunner *fake_command_runner.FakeCommandRunner

	BeforeEach(func() {
		fakeCmdRunner = fake_command_runner.New()
	})

	Describe("Run", func() {
		It("creates the number of images provided", func() {
			job := &bench.Job{
				Runner:         fakeCmdRunner,
				GrootFSBinPath: "/path/to/grootfs",
				StorePath:      "/store/path",
				LogLevel:       "debug",
				BaseImage:      "docker:///busybox",
				TotalImages:   11,
				Concurrency:    3,
			}
			job.Run()

			executedCommands := fakeCmdRunner.ExecutedCommands()
			Expect(len(executedCommands)).To(Equal(11))

			for _, cmd := range fakeCmdRunner.ExecutedCommands() {
				Expect(cmd.Args[0]).To(Equal("/path/to/grootfs"))
				Expect(cmd.Args[2]).To(Equal("/store/path"))
				Expect(cmd.Args[4]).To(Equal("debug"))
				Expect(cmd.Args[5]).To(Equal("create"))
				Expect(cmd.Args[6]).To(Equal("docker:///busybox"))
			}
		})

		Context("when UseQuota is true", func() {
			It("runs grootfs with --disk-limit-size-bytes flag", func() {
				job := &bench.Job{
					Runner:         fakeCmdRunner,
					GrootFSBinPath: "/path/to/grootfs",
					StorePath:      "/store/path",
					LogLevel:       "debug",
					BaseImage:      "docker:///busybox",
					UseQuota:       true,
					TotalImages:   11,
					Concurrency:    3,
				}
				job.Run()

				executedCommands := fakeCmdRunner.ExecutedCommands()
				Expect(len(executedCommands)).To(Equal(11))

				for _, cmd := range fakeCmdRunner.ExecutedCommands() {
					Expect(cmd.Args[0]).To(Equal("/path/to/grootfs"))
					Expect(cmd.Args[2]).To(Equal("/store/path"))
					Expect(cmd.Args[4]).To(Equal("debug"))
					Expect(cmd.Args[5]).To(Equal("create"))
					Expect(cmd.Args[6]).To(Equal("--disk-limit-size-bytes"))
					Expect(cmd.Args[7]).To(Equal("1019430400"))
					Expect(cmd.Args[8]).To(Equal("docker:///busybox"))
				}
			})
		})

		Context("when something fails", func() {
			BeforeEach(func() {
				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{}, func(cmd *exec.Cmd) error {
					cmd.Stderr.Write([]byte("groot failed to make a image"))
					return errors.New("exit status 1")
				})
			})

			It("returns the errors", func() {
				job := &bench.Job{
					Runner:       fakeCmdRunner,
					TotalImages: 10,
				}
				summary := job.Run()

				Expect(summary.ErrorMessages).To(HaveLen(10))
				for _, message := range summary.ErrorMessages {
					Expect(message).To(ContainSubstring("groot failed to make a image"))
				}
			})
		})

		Context("when not providing concurrency level", func() {
			It("sets the default to the # of cpus", func() {
				job := &bench.Job{
					Runner: fakeCmdRunner,
				}
				summary := job.Run()

				Expect(summary.ConcurrencyFactor).To(Equal(runtime.NumCPU()))
			})
		})
	})

	Describe("SummarizeResults", func() {
		var (
			results       chan *bench.Result
			totalDuration time.Duration
			spec          bench.SumSpec
		)

		BeforeEach(func() {
			results = make(chan *bench.Result, 100)
			totalDuration = time.Second * 20
		})

		JustBeforeEach(func() {
			r := &bench.Result{
				Err:      nil,
				Duration: 10 * time.Second,
			}
			results <- r
			results <- r
		})

		It("returns the results summarized", func() {
			close(results)

			spec = bench.SumSpec{
				Duration:    totalDuration,
				Concurrency: 2,
				UseQuota:    false,
				ResChan:     results,
			}
			summary := bench.SummarizeResults(spec)

			Expect(summary.TotalImages).To(Equal(2))
			Expect(summary.ImagesPerSecond).To(BeNumerically("~", 0.099, 0.1))
			Expect(summary.TotalDuration).To(BeNumerically("~", time.Second*20, time.Second*21))
			Expect(summary.AverageTimePerImage).To(Equal(float64(10)))
			Expect(summary.ConcurrencyFactor).To(Equal(2))
		})

		Context("when there are 0 images created", func() {
			It("sets the average time per image to -1", func() {
				results = make(chan *bench.Result, 1)
				results <- &bench.Result{
					Err:      errors.New("failed"),
					Duration: 10 * time.Second,
				}
				close(results)
				spec = bench.SumSpec{
					Duration:    totalDuration,
					Concurrency: 2,
					UseQuota:    false,
					ResChan:     results,
				}
				summary := bench.SummarizeResults(spec)

				Expect(summary.AverageTimePerImage).To(Equal(float64(-1)))
			})
		})

		Context("when command fails", func() {
			BeforeEach(func() {
				spec = bench.SumSpec{
					Duration:    totalDuration,
					Concurrency: 2,
					UseQuota:    false,
					ResChan:     results,
				}
			})

			JustBeforeEach(func() {
				results <- &bench.Result{
					Err:      errors.New("failed to execute grootfs"),
					Duration: 20 * time.Second,
				}
			})

			It("returns the total errors", func() {
				close(results)
				summary := bench.SummarizeResults(spec)
				Expect(summary.TotalErrorsAmt).To(Equal(1))
			})

			It("returns the error rate", func() {
				close(results)
				summary := bench.SummarizeResults(spec)
				// 33.33 because we're creating 2 in the outer BeforeEach
				Expect(summary.ErrorRate).To(BeNumerically(">", 33.33))
			})

			It("ignores the the failures for AverageTimePerImage metrics", func() {
				close(results)
				summary := bench.SummarizeResults(spec)
				// 10 because of the outer BeforeEach
				Expect(summary.AverageTimePerImage).To(Equal(float64(10)))
			})

			It("sets RanWithQuota to true if quota was applied", func() {
				close(results)
				spec = bench.SumSpec{
					Duration:    totalDuration,
					Concurrency: 2,
					UseQuota:    true,
					ResChan:     results,
				}
				summary := bench.SummarizeResults(spec)
				// 10 because of the outer BeforeEach
				Expect(summary.RanWithQuota).To(BeTrue())
			})
		})
	})
})
