package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"code.cloudfoundry.org/grootfs-bench/grootfspool"

	"github.com/briandowns/spinner"
	"github.com/cloudfoundry/gunk/command_runner/linux_command_runner"
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
			Value: "1000",
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
	}

	bench.Action = func(ctx *cli.Context) error {
		style := rand.New(rand.NewSource(time.Now().UnixNano())).Int() % 36
		s := spinner.New(spinner.CharSets[style], 100*time.Millisecond)
		s.Prefix = "Doing crazy maths "
		s.Color("green")
		s.Start()

		storePath := ctx.String("store")
		image := ctx.String("image")
		grootfs := ctx.String("gbin")
		totalBundles := ctx.Int("bundles")
		concurrency := ctx.Int("concurrency")

		cmdRunner := linux_command_runner.New()
		pool := grootfspool.New(cmdRunner, grootfs, storePath, image, concurrency)
		bundlesChan := pool.Start(totalBundles)

		start := time.Now()
		for bundle := 1; bundle <= totalBundles; bundle++ {
			bundlesChan <- bundle
		}
		close(bundlesChan)
		pool.Wait()
		totalDuration := time.Since(start)

		createdBundles := 0
		averageTimePerBundle := 0.0
		for result := range pool.DurationChan() {
			createdBundles++
			averageTimePerBundle += result.Seconds()
		}
		averageTimePerBundle = averageTimePerBundle / float64(createdBundles)

		totalErrors := 0
		for err := range pool.ErrorsChan() {
			totalErrors++
			fmt.Fprintf(os.Stderr, "Failures: %s\n", err.Error())
		}

		s.Stop()
		fmt.Printf("\r........................                     \n")
		fmt.Printf("Total duration.........: %s\n", totalDuration.String())
		fmt.Printf("Bundles per second.....: %f\n", float64(createdBundles)/totalDuration.Seconds())
		fmt.Printf("Average time per bundle: %f\n", averageTimePerBundle)
		fmt.Printf("Total errors...........: %d\n", totalErrors)
		fmt.Printf("Error Rate.............: %f\n", float64(totalErrors*100)/float64(totalBundles))

		return nil
	}

	bench.Run(os.Args)
}
