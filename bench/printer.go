package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
)

//go:generate counterfeiter . Printer
type Printer interface {
	Print(summary Summary) []byte
}

type TextPrinter []byte

func (TextPrinter) Print(summary Summary) []byte {
	buffer := bytes.NewBuffer([]byte{})

	buffer.WriteString(fmt.Sprintf("\nTotal bundles requested: %d\n", summary.TotalBundles))
	buffer.WriteString(fmt.Sprintf("Concurrency factor.....: %d\n", summary.ConcurrencyFactor))
	buffer.WriteString(fmt.Sprintf("Using quota?...........: %t\n", summary.RanWithQuota))
	buffer.WriteString(fmt.Sprintf("\r........................                     \n"))
	buffer.WriteString(fmt.Sprintf("Total duration.........: %s\n", summary.TotalDuration))
	buffer.WriteString(fmt.Sprintf("Bundles per second.....: %.3f\n", summary.BundlesPerSecond))
	buffer.WriteString(fmt.Sprintf("Average time per bundle: %.3fs\n", summary.AverageTimePerBundle))
	buffer.WriteString(fmt.Sprintf("Total errors...........: %d\n", summary.TotalErrorsAmt))
	buffer.WriteString(fmt.Sprintf("Error Rate.............: %.3f\n", summary.ErrorRate))

	return buffer.Bytes()
}

type JsonPrinter []byte

func (JsonPrinter) Print(summary Summary) []byte {
	sum, err := json.Marshal(summary)
	if err != nil {
		return []byte{}
	}
	return sum
}
