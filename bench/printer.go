package bench

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
)

type Printer interface {
	Print(summary Summary) error
}

func NewTextPrinter(out, err io.Writer) *TextPrinter {
	return &TextPrinter{out: out, err: err}
}

type TextPrinter struct {
	out io.Writer
	err io.Writer
}

func (p *TextPrinter) Print(summary Summary) error {
	printErrors(summary, p.err)

	tmplText := `
Total images requested: {{.TotalImages}}
Concurrency factor.....: {{.ConcurrencyFactor}}
Using quota?...........: {{.RanWithQuota}}
........................
Total duration.........: {{.TotalDuration}}
Images per second.....: {{printf "%.3f" .ImagesPerSecond}}
Average time per image: {{printf "%.3f" .AverageTimePerImage}}s
Total errors...........: {{.TotalErrorsAmt}}
Error Rate.............: {{printf "%.3f" .ErrorRate}}
`
	tmpl, err := template.New("groot").Parse(tmplText)
	if err != nil {
		return err
	}

	return tmpl.Execute(p.out, summary)
}

func NewJsonPrinter(out, err io.Writer) *JsonPrinter {
	return &JsonPrinter{out: out, err: err}
}

type JsonPrinter struct {
	out io.Writer
	err io.Writer
}

func (j *JsonPrinter) Print(summary Summary) error {
	printErrors(summary, j.err)

	return json.NewEncoder(j.out).Encode(summary)
}

func printErrors(summary Summary, buffer io.Writer) {
	if len(summary.ErrorMessages) > 0 {
		for _, message := range summary.ErrorMessages {
			fmt.Fprintf(buffer, message)
		}
	}
}
