package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"code.cloudfoundry.org/grootfs-bench/grootfspool"

	"code.cloudfoundry.org/commandrunner/linux_command_runner"
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
		}

		cmdRunner := linux_command_runner.New()
		pool := grootfspool.New(cmdRunner, grootfs, storePath, image, concurrency)

		if hasSpinner {
			spinner.Stop()
		}

		res := run(totalBundlesAmt, concurrency, pool)

		if jsonify {
			j, err := json.Marshal(res)
			if err != nil {
				return err
			}
			fmt.Println(string(j))
			return nil
		}

		fmt.Printf("Total bundles requested: %d\n", res.BundlesRequested)
		fmt.Printf("Concurrency factor.....: %d\n", res.ConcurrencyFactor)
		fmt.Printf("\r........................                     \n")
		fmt.Printf("Total duration.........: %s\n", res.TotalDuration.String())
		fmt.Printf("Bundles per second.....: %f\n", res.BundlesPerSecond)
		fmt.Printf("Average time per bundle: %f\n", res.AverageTimePerBundle)
		fmt.Printf("Total errors...........: %d\n", res.TotalErrorsAmt)
		fmt.Printf("Error Rate.............: %f\n", res.ErrorRate)

		return nil
	}

	bench.Run(os.Args)
}

func run(totalBundlesAmt int, concurrency int, pool *grootfspool.Pool) result {
	bundlesChan := pool.Start(totalBundlesAmt)

	start := time.Now()
	for bundle := 1; bundle <= totalBundlesAmt; bundle++ {
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

	totalErrorsAmt := 0
	for err := range pool.ErrorsChan() {
		totalErrorsAmt++
		fmt.Fprintf(os.Stderr, "Failures: %s\n", err.Error())
	}

	return result{
		TotalDuration:        totalDuration,
		BundlesPerSecond:     float64(createdBundles) / totalDuration.Seconds(),
		AverageTimePerBundle: averageTimePerBundle,
		TotalErrorsAmt:       totalErrorsAmt,
		ErrorRate:            float64(totalErrorsAmt*100) / float64(totalBundlesAmt),
		BundlesRequested:     totalBundlesAmt,
		ConcurrencyFactor:    concurrency,
	}
}

type result struct {
	TotalDuration        time.Duration `json:"total_duration"`
	BundlesPerSecond     float64       `json:"bundles_per_second"`
	AverageTimePerBundle float64       `json:"average_time_per_bundle"`
	TotalErrorsAmt       int           `json:"total_errors_amt"`
	ErrorRate            float64       `json:"error_rate"`
	BundlesRequested     int           `json:"bundles_requested"`
	ConcurrencyFactor    int           `json:"concurrency_factor"`
}
