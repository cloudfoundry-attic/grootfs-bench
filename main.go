package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"code.cloudfoundry.org/commandrunner/linux_command_runner"
	benchpkg "code.cloudfoundry.org/grootfs-bench/bench"
	spinnerpkg "github.com/briandowns/spinner"
	"github.com/urfave/cli"
)

func main() {
	bench := cli.NewApp()
	bench.Name = "grootfs-bench"
	bench.Usage = "grootfs awesome benchmarking tool"
	bench.UsageText = "grootfs-bench --gbin <grootfs-bin> --store <store-path> --log-level <debug|info|warn> --images <n> --concurrency <c> --base-image <docker:///img>"
	bench.Version = "0.1.0"

	bench.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "gbin",
			Usage: "path to grootfs bin",
			Value: "grootfs",
		},
		cli.StringFlag{
			Name:  "images",
			Usage: "number of images to create",
			Value: "500",
		},
		cli.StringFlag{
			Name:  "concurrency",
			Usage: "what the name says",
			Value: "5",
		},
		cli.StringFlag{
			Name:  "store",
			Usage: "store path",
			Value: "/var/lib/grootfs",
		},
		cli.StringFlag{
			Name:  "driver",
			Usage: "filesystem driver",
		},
		cli.StringFlag{
			Name:  "log-level",
			Usage: "what the name says",
			Value: "debug",
		},
		cli.StringSliceFlag{
			Name:  "base-image",
			Usage: "base image to use",
		},
		cli.BoolFlag{
			Name:  "with-quota",
			Usage: "add quotas to the image creation",
		},
		cli.BoolFlag{
			Name:  "nospin",
			Usage: "turn off the awesome spinner, you monster",
		},
		cli.BoolFlag{
			Name:  "json",
			Usage: "return the result in json format",
		},
		cli.BoolFlag{
			Name:  "parallel-clean",
			Usage: "run a concurrent clean operation",
		},
		cli.IntFlag{
			Name:  "parallel-clean-interval",
			Usage: "interval at which to call clean during concurrent operations in seconds. parallel-clean must also be set",
			Value: 6,
		},
		cli.IntFlag{
			Name:  "parallel-delete-interval",
			Usage: "interval at which to call delete during concurrent operations in seconds. parallel-clean must also be set",
			Value: 3,
		},
	}

	bench.Action = func(ctx *cli.Context) error {
		storePath := ctx.String("store")
		fsDriver := ctx.String("driver")
		logLevel := ctx.String("log-level")
		baseImages := ctx.StringSlice("base-image")
		grootfs := ctx.String("gbin")
		totalImagesAmt := ctx.Int("images")
		concurrency := ctx.Int("concurrency")
		withQuota := ctx.Bool("with-quota")
		withParallelClean := ctx.Bool("parallel-clean")
		parallelCleanInterval := ctx.Int("parallel-clean-interval")
		parallelDeleteInterval := ctx.Int("parallel-delete-interval")
		jsonify := ctx.Bool("json")
		hasSpinner := !ctx.Bool("nospin")

		var spinner *spinnerpkg.Spinner
		if hasSpinner {
			now := time.Now().Format("15:04:05")
			style := rand.New(rand.NewSource(time.Now().UnixNano())).Int() % 36
			spinner = spinnerpkg.New(spinnerpkg.CharSets[style], 100*time.Millisecond)
			spinner.Prefix = fmt.Sprintf("Doing crazy maths since %v (images: %d, conc: %d, parallel commands? %v) ", now, totalImagesAmt, concurrency, withParallelClean)
			must(spinner.Color("green"))
			spinner.Start()
			defer spinner.Stop()
		}

		var printer benchpkg.Printer
		printer = benchpkg.NewTextPrinter(os.Stdout, os.Stderr)
		if jsonify {
			printer = benchpkg.NewJsonPrinter(os.Stdout, os.Stderr)
		}

		cmdRunner := linux_command_runner.New()
		executor := &benchpkg.JobExecutor{
			Jobs: []*benchpkg.Job{
				&benchpkg.Job{
					Command:        "create",
					Runner:         cmdRunner,
					GrootFSBinPath: grootfs,
					StorePath:      storePath,
					Driver:         fsDriver,
					LogLevel:       logLevel,
					UseQuota:       withQuota,
					BaseImages:     baseImages,
					Concurrency:    concurrency,
					TotalImages:    totalImagesAmt,
				},
			},
		}
		if withParallelClean {
			executor.Jobs = append(executor.Jobs,
				&benchpkg.Job{
					Command:        "clean",
					Runner:         cmdRunner,
					GrootFSBinPath: grootfs,
					StorePath:      storePath,
					Driver:         fsDriver,
					LogLevel:       logLevel,
					Interval:       parallelCleanInterval,
				})
			executor.Jobs = append(executor.Jobs,
				&benchpkg.Job{
					Command:        "delete",
					Runner:         cmdRunner,
					GrootFSBinPath: grootfs,
					StorePath:      storePath,
					Driver:         fsDriver,
					LogLevel:       logLevel,
					Interval:       parallelDeleteInterval,
				})
		}

		summary := executor.Run()
		summary.RanWithParallelClean = withParallelClean
		if err := printer.Print(summary); err != nil {
			return err
		}

		if summary.TotalErrorsAmt > 0 {
			return fmt.Errorf("%s failed %d times\n", grootfs, summary.TotalErrorsAmt)
		}

		return nil
	}

	if err := bench.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
