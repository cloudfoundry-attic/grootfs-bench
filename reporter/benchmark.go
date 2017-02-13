package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var metricsEndpoint = "https://app.datadoghq.com/api/v1/series?api_key=" + os.Getenv("DATADOG_API_KEY")

func benchmarkCommand() error {

	args := fmt.Sprintf("%s", strings.Join(flag.Args(), " "))
	// We need to strip the first and last char because of `flag` package thinks
	// we're passing arguments
	args = args[1 : len(args)-1]
	// We cannot send one "big string" to exec.Cmd
	cmd := exec.Command(benchBinPath, strings.Split(args, " ")...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Printf("sending the following metrics to datadog:\n%+v\n", string(out))

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return err
	}

	series := createMetricSeries(metricPrefix, result)
	emitMetric(map[string]interface{}{
		"series": series,
	})

	return nil
}

func createMetricSeries(prefix string, result map[string]interface{}) []map[string]interface{} {
	series := make([]map[string]interface{}, len(result))

	for key, value := range result {
		now := float64(time.Now().Unix())

		//convert value into float
		metric, ok := value.(float64)
		if !ok {
			continue
		}

		series = append(series, map[string]interface{}{
			"metric": fmt.Sprintf("%s.grootfs.benchmark-performance.%s", prefix, key),
			"points": [][]float64{
				{now, metric},
			},
			"tags": []string{"concourse"},
		})
	}

	return series
}

func emitMetric(req interface{}) {
	buf, err := json.Marshal(req)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal data: %s", err))
		return
	}
	response, err := http.Post(metricsEndpoint, "application/json", bytes.NewReader(buf))
	if err != nil {
		panic(fmt.Errorf("Failed to send request to Datadog: %s", err))
		return
	}
	if response.StatusCode != http.StatusAccepted {
		panic("datadog is the worst :(")
		return
	}
}
