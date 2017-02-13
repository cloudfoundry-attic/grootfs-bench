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
	bench.UsageText = "grootfs-bench --gbin <grootfs-bin> --store <btrfs-store> --log-level <debug|info|warn> --images <n> --concurrency <c> --base-image <docker:///img>"
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
		cli.StringFlag{
			Name:  "base-image",
			Usage: "base image to use",
			Value: "docker:///busybox:latest",
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
	}

	bench.Action = func(ctx *cli.Context) error {
		storePath := ctx.String("store")
		fsDriver := ctx.String("driver")
		logLevel := ctx.String("log-level")
		baseImage := ctx.String("base-image")
		grootfs := ctx.String("gbin")
		totalImagesAmt := ctx.Int("images")
		concurrency := ctx.Int("concurrency")
		nospin := ctx.Bool("nospin")
		jsonify := ctx.Bool("json")
		hasSpinner := !nospin

		var spinner *spinnerpkg.Spinner
		if hasSpinner {
			style := rand.New(rand.NewSource(time.Now().UnixNano())).Int() % 36
			spinner = spinnerpkg.New(spinnerpkg.CharSets[style], 100*time.Millisecond)
			spinner.Prefix = "Doing crazy maths "
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
		summary := (&benchpkg.Job{
			Runner:         cmdRunner,
			GrootFSBinPath: grootfs,
			StorePath:      storePath,
			Driver:         fsDriver,
			LogLevel:       logLevel,
			UseQuota:       ctx.Bool("with-quota"),
			BaseImage:      baseImage,
			Concurrency:    concurrency,
			TotalImages:    totalImagesAmt,
		}).Run()

		if err := printer.Print(summary); err != nil {
			return err
		}

		if summary.TotalErrorsAmt > 0 {
			return fmt.Errorf("%s failed %d times", grootfs, summary.TotalErrorsAmt)
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
