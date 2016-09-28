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

var (
	benchBinPath string
)

var dogURL = "https://app.datadoghq.com/api/v1/series?api_key=" + os.Getenv("DATADOG_API_KEY")

func main() {
	flag.StringVar(&benchBinPath, "benchBinPath", "", "")
	flag.Parse()

	args := fmt.Sprintf("%s", strings.Join(flag.Args(), " "))
	// We need to strip the first and last char because of `flag` package thinks
	// we're passing arguments
	args = args[1 : len(args)-1]

	// We cannot send one "big string" to exec.Cmd
	cmd := exec.Command(benchBinPath, strings.Split(args, " ")...)
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	fmt.Printf("sending the following metrics to datadog:\n%+v\n", string(out))

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		panic(err)
	}

	series := createMetricSeries(result)
	emitMetric(map[string]interface{}{
		"series": series,
	})
}

func createMetricSeries(result map[string]interface{}) []map[string]interface{} {
	series := make([]map[string]interface{}, len(result))

	for key, value := range result {
		now := float64(time.Now().Unix())

		//convert value into float
		metric, ok := value.(float64)
		if !ok {
			continue
		}

		series = append(series, map[string]interface{}{
			"metric": fmt.Sprintf("grootfs.benchmark-performance.%s", key),
			"points": [][]float64{
				{now, metric},
			},
			"tags": []string{"concourse"},
		})
	}

	return series
}

func emitMetric(req interface{}) {
	if os.Getenv("DATADOG_API_KEY") == "" {
		panic("datadog api key not specified")
	}
	buf, err := json.Marshal(req)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal data: %s", err))
		return
	}
	response, err := http.Post(dogURL, "application/json", bytes.NewReader(buf))
	if err != nil {
		panic(fmt.Errorf("Failed to send request to Datadog: %s", err))
		return
	}
	if response.StatusCode != http.StatusAccepted {
		panic("datadog is the worst :(")
		return
	}
}
