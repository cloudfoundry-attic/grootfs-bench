package grootfspool_test

import (
	"errors"
	"os/exec"
	"regexp"
	"time"

	"code.cloudfoundry.org/grootfs-bench/grootfspool"

	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool", func() {
	var (
		fakeCmdRunner *fake_command_runner.FakeCommandRunner
	)

	BeforeEach(func() {
		fakeCmdRunner = fake_command_runner.New()
	})

	Describe("Start", func() {
		Context("number of bundles", func() {
			It("creates the number of bundles provided", func() {
				runAndWait(fakeCmdRunner, "/grootfs", "/store", "ubuntu", 1, 10)
				Expect(len(fakeCmdRunner.ExecutedCommands())).To(Equal(10))
			})
		})

		Context("concurrency", func() {
			BeforeEach(func() {
				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/grootfs",
				}, func(cmd *exec.Cmd) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				})
			})

			It("creates the bundles in parallel", func() {
				runAndWait(fakeCmdRunner, "/grootfs", "/store", "ubuntu", 3, 6)

				workerIds := map[string]bool{}

				for _, cmd := range fakeCmdRunner.ExecutedCommands() {
					imageName := cmd.Args[5]
					regexp := regexp.MustCompile(`image-(\d)-\d-\d`)
					workerId := regexp.FindStringSubmatch(imageName)[1]
					workerIds[workerId] = true
				}

				Expect(workerIds["0"]).To(BeTrue())
				Expect(workerIds["1"]).To(BeTrue())
				Expect(workerIds["2"]).To(BeTrue())
			})
		})
	})

	Context("ErrorChan", func() {
		BeforeEach(func() {
			fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{
				Path: "/grootfs",
			}, func(cmd *exec.Cmd) error {
				return errors.New("failed here")
			})
		})

		It("returns a channel with runtime errors", func() {
			pool := runAndWait(fakeCmdRunner, "/grootfs", "/store", "ubuntu", 1, 2)
			errorsChan := pool.ErrorsChan()

			Expect(len(errorsChan)).To(Equal(2))
			for err := range errorsChan {
				Expect(err).To(MatchError(ContainSubstring("failed here")))
			}
		})
	})

	Context("DurationChan", func() {

		Context("when all the commands succeed", func() {
			BeforeEach(func() {
				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/grootfs",
				}, func(cmd *exec.Cmd) error {
					time.Sleep(15 * time.Millisecond)
					return nil
				})
			})

			It("returns a channel with execution durations", func() {
				pool := runAndWait(fakeCmdRunner, "/grootfs", "/store", "ubuntu", 1, 2)
				durationChan := pool.DurationChan()

				Expect(len(durationChan)).To(Equal(2))

				for duration := range durationChan {
					Expect(duration.Nanoseconds()).To(BeNumerically(">", 15000000))
				}
			})
		})

		Context("when a commands fails", func() {
			BeforeEach(func() {
				failedCommands := 0

				fakeCmdRunner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/grootfs",
				}, func(cmd *exec.Cmd) error {
					if failedCommands == 0 {
						failedCommands++
						return errors.New("failed")
					}

					return nil
				})
			})

			It("only counts the duration of the successful commands", func() {
				pool := runAndWait(fakeCmdRunner, "/grootfs", "/store", "ubuntu", 1, 2)
				durationChan := pool.DurationChan()

				Expect(len(durationChan)).To(Equal(1))
			})
		})

	})
})

func runAndWait(fakeCmdRunner *fake_command_runner.FakeCommandRunner, grootfs, storePath, image string, concurrency, totalBundles int) *grootfspool.Pool {
	pool := grootfspool.New(fakeCmdRunner, grootfs, storePath, image, concurrency)

	jobsChan := pool.Start(totalBundles)
	for i := 0; i < totalBundles; i++ {
		jobsChan <- i
	}

	close(jobsChan)
	pool.Wait()

	return pool
}
