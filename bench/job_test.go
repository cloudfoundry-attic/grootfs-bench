package bench_test

import (
	"errors"
	"runtime"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"
	"code.cloudfoundry.org/grootfs-bench/bench"
	"code.cloudfoundry.org/grootfs-bench/bench/benchfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job", func() {

	var (
		fakeCmdRunner *fake_command_runner.FakeCommandRunner
		fakePrinter   *benchfakes.FakePrinter
	)

	BeforeEach(func() {
		fakeCmdRunner = fake_command_runner.New()
		fakePrinter = new(benchfakes.FakePrinter)
	})

	Describe("Run", func() {
		It("creates the number of bundles provided", func() {
			job := &bench.Job{
				Runner:         fakeCmdRunner,
				GrootFSBinPath: "/path/to/grootfs",
				StorePath:      "/store/path",
				LogLevel:       "debug",
				Image:          "docker:///busybox",
				TotalBundles:   11,
				Concurrency:    3,
			}
			job.Run(fakePrinter)

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
					Image:          "docker:///busybox",
					UseQuota:       true,
					TotalBundles:   11,
					Concurrency:    3,
				}
				job.Run(fakePrinter)

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

		Context("when not providing concurrency level", func() {
			It("sets the default to the # of cpus", func() {
				job := &bench.Job{}
				job.Run(fakePrinter)

				Expect(job.Concurrency).To(Equal(runtime.NumCPU()))
			})
		})

		It("prints the result", func() {
			job := &bench.Job{}
			job.Run(fakePrinter)

			Expect(fakePrinter.PrintCallCount()).To(Equal(1))
		})
	})

	Describe("SummarizeResults", func() {
		var (
			results       chan *bench.Result
			totalDuration time.Duration
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
			summary := bench.SummarizeResults(totalDuration, 2, false, results)

			Expect(summary.TotalBundles).To(Equal(2))
			Expect(summary.BundlesPerSecond).To(BeNumerically("~", 0.099, 0.1))
			Expect(summary.TotalDuration).To(BeNumerically("~", time.Second*20, time.Second*21))
			Expect(summary.AverageTimePerBundle).To(Equal(float64(10)))
			Expect(summary.ConcurrencyFactor).To(Equal(2))
		})

		Context("when there are 0 images created", func() {
			It("sets the average time per bundle to -1", func() {
				results = make(chan *bench.Result, 1)
				results <- &bench.Result{
					Err:      errors.New("failed"),
					Duration: 10 * time.Second,
				}
				close(results)
				summary := bench.SummarizeResults(totalDuration, 1, false, results)

				Expect(summary.AverageTimePerBundle).To(Equal(float64(-1)))
			})
		})

		Context("when command fails", func() {
			JustBeforeEach(func() {
				results <- &bench.Result{
					Err:      errors.New("failed to execute grootfs"),
					Duration: 20 * time.Second,
				}
			})

			It("returns the total errors", func() {
				close(results)
				summary := bench.SummarizeResults(totalDuration, 2, false, results)
				Expect(summary.TotalErrorsAmt).To(Equal(1))
			})

			It("returns the error rate", func() {
				close(results)
				summary := bench.SummarizeResults(totalDuration, 2, false, results)
				// 33.33 because we're creating 2 in the outer BeforeEach
				Expect(summary.ErrorRate).To(BeNumerically(">", 33.33))
			})

			It("ignores the the failures for AverageTimePerBundle metrics", func() {
				close(results)
				summary := bench.SummarizeResults(totalDuration, 2, false, results)
				// 10 because of the outer BeforeEach
				Expect(summary.AverageTimePerBundle).To(Equal(float64(10)))
			})

			It("sets RanWithQuota to true if quota was applied", func() {
				close(results)
				summary := bench.SummarizeResults(totalDuration, 2, true, results)
				// 10 because of the outer BeforeEach
				Expect(summary.RanWithQuota).To(BeTrue())
			})
		})
	})
})
