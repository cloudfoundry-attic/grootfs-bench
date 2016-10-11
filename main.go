package main

import (
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
	bench.UsageText = "grootfs-bench --gbin <grootfs-bin> --store <btrfs-store> --bundles <n> --concurrency <c> --image <docker:///img>"
	bench.Version = "0.1.0"

	bench.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "gbin",
			Usage: "path to grootfs bin",
			Value: "grootfs",
		},
		cli.StringFlag{
			Name:  "bundles",
			Usage: "number of bundles to create",
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
			Name:  "image",
			Usage: "image to use",
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
		image := ctx.String("image")
		grootfs := ctx.String("gbin")
		totalBundlesAmt := ctx.Int("bundles")
		concurrency := ctx.Int("concurrency")
		nospin := ctx.Bool("nospin")
		jsonify := ctx.Bool("json")
		hasSpinner := !nospin

		var spinner *spinnerpkg.Spinner
		if hasSpinner {
			style := rand.New(rand.NewSource(time.Now().UnixNano())).Int() % 36
			spinner = spinnerpkg.New(spinnerpkg.CharSets[style], 100*time.Millisecond)
			spinner.Prefix = "Doing crazy maths "
			spinner.Color("green")
			spinner.Start()
			defer spinner.Stop()
		}

		var printer benchpkg.Printer
		printer = benchpkg.TextPrinter([]byte{})
		if jsonify {
			printer = benchpkg.JsonPrinter([]byte{})
		}

		cmdRunner := linux_command_runner.New()
		(&benchpkg.Job{
			Runner:         cmdRunner,
			GrootFSBinPath: grootfs,
			StorePath:      storePath,
			UseQuota:       ctx.Bool("with-quota"),
			Image:          image,
			Concurrency:    concurrency,
			TotalBundles:   totalBundlesAmt,
		}).Run(printer)

		return nil
	}

	bench.Run(os.Args)
}
