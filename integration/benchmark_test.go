package integration_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Benchmark", func() {
	Context("when --json is provided", func() {
		It("returns a json formatted bench summary", func() {
			cmd := exec.Command(GrootFSBenchBin, "--gbin", FakeGrootFS, "--nospin", "--bundles", "10", "--json")
			out, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())

			var result interface{}
			err = json.Unmarshal(out, &result)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
