package bench_test

import (
	"time"

	"code.cloudfoundry.org/grootfs-bench/bench"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Printer", func() {
	var summary bench.Summary

	BeforeEach(func() {
		summary = bench.Summary{
			TotalDuration:        time.Millisecond,
			ImagesPerSecond:     0.88,
			RanWithQuota:         true,
			AverageTimePerImage: 2,
			TotalErrorsAmt:       3,
			ErrorRate:            4,
			TotalImages:         5,
			ConcurrencyFactor:    6,
			ErrorMessages:        []string{"o noes"},
		}
	})

	Describe("TextPrinter", func() {
		Describe("Print", func() {
			It("prints the summary in plain text", func() {
				errBuffer := gbytes.NewBuffer()
				outBuffer := gbytes.NewBuffer()

				printer := bench.NewTextPrinter(outBuffer, errBuffer)
				Expect(printer.Print(summary)).To(Succeed())

				Expect(outBuffer).Should(gbytes.Say("Total images requested([.]*): 5"))
				Expect(outBuffer).Should(gbytes.Say("Concurrency factor([.]*): 6"))
				Expect(outBuffer).Should(gbytes.Say("Total duration([.]*): 1ms"))
				Expect(outBuffer).Should(gbytes.Say("Images per second([.]*): 0.880"))
				Expect(outBuffer).Should(gbytes.Say("Average time per image([.]*): 2.000s"))
				Expect(outBuffer).Should(gbytes.Say("Total errors([.]*): 3"))
				Expect(outBuffer).Should(gbytes.Say("Error Rate([.]*): 4.000"))
			})

			It("prints the error messages if something went wrong", func() {
				outBuffer := gbytes.NewBuffer()
				errBuffer := gbytes.NewBuffer()

				printer := bench.NewTextPrinter(outBuffer, errBuffer)
				Expect(printer.Print(summary)).To(Succeed())

				Expect(errBuffer).Should(gbytes.Say("o noes"))
			})
		})
	})

	Describe("JsonPrinter", func() {
		Describe("Print", func() {
			It("prints the summary in json", func() {
				outBuffer := gbytes.NewBuffer()
				errBuffer := gbytes.NewBuffer()

				printer := bench.NewJsonPrinter(outBuffer, errBuffer)
				Expect(printer.Print(summary)).To(Succeed())

				Expect(outBuffer.Contents()).To(MatchJSON(`{"total_duration":1000000,"images_per_second":0.88,"ran_with_quota":true,"average_time_per_image":2,"total_errors_amt":3,"error_rate":4,"total_images":5,"concurrency_factor":6}`))
			})

			It("prints the error messages in plain text", func() {
				outBuffer := gbytes.NewBuffer()
				errBuffer := gbytes.NewBuffer()

				printer := bench.NewJsonPrinter(outBuffer, errBuffer)
				Expect(printer.Print(summary)).To(Succeed())

				Expect(errBuffer).Should(gbytes.Say("o noes"))
			})
		})
	})
})
