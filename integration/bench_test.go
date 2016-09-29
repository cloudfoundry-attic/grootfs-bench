package integration_test

import (
	"encoding/json"
	"os/exec"

	"code.cloudfoundry.org/grootfs-bench/bench"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Bench", func() {

	It("returns the output in plain text by default", func() {
		cmd := exec.Command(GrootFSBenchBin, "--gbin", FakeGrootFS, "--nospin", "--bundles", "10")
		buffer := gbytes.NewBuffer()
		cmd.Stdout = buffer
		err := cmd.Run()

		Expect(err).NotTo(HaveOccurred())
		Expect(buffer).Should(gbytes.Say("Total bundles requested"))
		Expect(buffer).Should(gbytes.Say("Concurrency factor"))
		Expect(buffer).Should(gbytes.Say("Total duration"))
		Expect(buffer).Should(gbytes.Say("Bundles per second"))
		Expect(buffer).Should(gbytes.Say("Average time per bundle"))
		Expect(buffer).Should(gbytes.Say("Total errors"))
		Expect(buffer).Should(gbytes.Say("Error Rate"))
	})

	Context("when --json is provided", func() {
		It("returns a json formatted summary", func() {
			cmd := exec.Command(GrootFSBenchBin, "--gbin", FakeGrootFS, "--nospin", "--bundles", "10", "--json")
			out, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())

			var result bench.Result
			err = json.Unmarshal(out, &result)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
