package bench_test

import (
	"code.cloudfoundry.org/grootfs-bench/bench"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("TextPrinter", func() {
	var (
		summary bench.Summary
	)

	BeforeEach(func() {
		summary = bench.Summary{}
	})

	Describe("Print", func() {
		It("prints the summary in plain text", func() {
			summary := bench.TextPrinter([]byte{}).Print(summary)

			buffer := gbytes.BufferWithBytes(summary)
			Expect(buffer).Should(gbytes.Say("Total bundles requested"))
			Expect(buffer).Should(gbytes.Say("Concurrency factor"))
			Expect(buffer).Should(gbytes.Say("Total duration"))
			Expect(buffer).Should(gbytes.Say("Bundles per second"))
			Expect(buffer).Should(gbytes.Say("Average time per bundle"))
			Expect(buffer).Should(gbytes.Say("Total errors"))
			Expect(buffer).Should(gbytes.Say("Error Rate"))
		})
	})
})

var _ = Describe("JsonPrinter", func() {
	var (
		summary bench.Summary
	)

	BeforeEach(func() {
		summary = bench.Summary{}
	})

	Describe("Print", func() {
		It("prints the summary in json", func() {
			summary := bench.JsonPrinter([]byte{}).Print(summary)
			Expect(summary).To(MatchJSON(`{"total_duration":0,"bundles_per_second":0,"ran_with_quota":false,"average_time_per_bundle":0,"total_errors_amt":0,"error_rate":0,"total_bundles":0,"concurrency_factor":0}`))
		})
	})
})
