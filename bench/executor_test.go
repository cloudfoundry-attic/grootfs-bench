package bench_test

import (
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/commandrunner/fake_command_runner"
	"code.cloudfoundry.org/grootfs-bench/bench"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SlowFakeCommandRunner struct {
	Runner *fake_command_runner.FakeCommandRunner
}

func (r *SlowFakeCommandRunner) Background(cmd *exec.Cmd) error {
	return r.Runner.Background(cmd)
}

func (r *SlowFakeCommandRunner) Kill(cmd *exec.Cmd) error {
	return r.Runner.Kill(cmd)
}

func (r *SlowFakeCommandRunner) Run(cmd *exec.Cmd) error {
	err := r.Runner.Run(cmd)
	time.Sleep(1 * time.Second)
	return err
}

func (r *SlowFakeCommandRunner) Signal(cmd *exec.Cmd, signal os.Signal) error {
	return r.Runner.Signal(cmd, signal)
}

func (r *SlowFakeCommandRunner) Start(cmd *exec.Cmd) error {
	return r.Runner.Start(cmd)
}

func (r *SlowFakeCommandRunner) Wait(cmd *exec.Cmd) error {
	return r.Runner.Wait(cmd)
}

var _ = Describe("JobExecutor", func() {
	var fakeCmdRunner *fake_command_runner.FakeCommandRunner
	var executor bench.JobExecutor

	BeforeEach(func() {
		fakeCmdRunner = fake_command_runner.New()
		executor = bench.JobExecutor{}
	})

	Describe("Run", func() {
		Describe("when there are no jobs", func() {
			It("does not fail", func() {
				executor.Jobs = []*bench.Job{}
				Expect(executor.Run()).To(Equal(bench.Summary{}))
			})
		})

		Describe("when a set of jobs is given", func() {
			var fakeCmdRunner1 *fake_command_runner.FakeCommandRunner
			var fakeCmdRunner2 *fake_command_runner.FakeCommandRunner
			var fakeCmdRunner3 *fake_command_runner.FakeCommandRunner

			BeforeEach(func() {
				fakeCmdRunner1 = fake_command_runner.New()
				fakeCmdRunner2 = fake_command_runner.New()
				fakeCmdRunner3 = fake_command_runner.New()
				createRunner := &SlowFakeCommandRunner{
					Runner: fakeCmdRunner1,
				}

				executor.Jobs = []*bench.Job{
					&bench.Job{Command: "create", Runner: createRunner, TotalImages: 5, Concurrency: 1, BaseImages: []string{"image"}},
					&bench.Job{Command: "clean", Runner: fakeCmdRunner2, Interval: 2},
					&bench.Job{Command: "delete", Runner: fakeCmdRunner3, Interval: 1},
				}
			})

			It("executes them", func() {
				summary := executor.Run()
				Expect(summary).ToNot(BeNil())
				Expect(fakeCmdRunner1.ExecutedCommands()).NotTo(HaveLen(0), "create job ran no commands")
				Expect(fakeCmdRunner2.ExecutedCommands()).NotTo(HaveLen(0), "clean job ran no commands")
				Expect(fakeCmdRunner3.ExecutedCommands()).NotTo(HaveLen(0), "delete job ran no commands")
			})

			It("reports how many times delete and clean commands ran successfully", func() {
				summary := executor.Run()
				Expect(summary).ToNot(BeNil())
				Expect(summary.NumberOfCleans).To(Equal(len(fakeCmdRunner2.ExecutedCommands())))
				Expect(summary.NumberOfDeletes).To(Equal(len(fakeCmdRunner3.ExecutedCommands())))
			})
		})
	})
})
