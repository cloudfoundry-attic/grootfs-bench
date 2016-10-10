package main

import (
	"errors"
	"flag"
	"os"
)

var (
	benchBinPath string
	eventTitle   string
	eventMessage string
	mode         string
)

func init() {
	flag.StringVar(&benchBinPath, "benchBinPath", "", "")
	flag.StringVar(&eventTitle, "eventTitle", "", "")
	flag.StringVar(&eventMessage, "eventMessage", "", "")
	flag.StringVar(&mode, "mode", "", "")
	flag.Parse()
}

func main() {
	if os.Getenv("DATADOG_API_KEY") == "" {
		panic(errors.New("datadog api key not specified"))
	}

	if os.Getenv("DATADOG_APPLICATION_KEY") == "" {
		panic(errors.New("datadog application key not specified"))
	}

	var err error
	if mode == "event" {
		err = eventCommand()
	} else {
		err = benchmarkCommand()
	}

	if err != nil {
		panic(err)
	}
}
