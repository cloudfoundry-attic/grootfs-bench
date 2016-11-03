package integration_test

import (
	"encoding/json"
	"os/exec"

	"code.cloudfoundry.org/grootfs-bench/bench"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Bench", func() {
	It("returns the output in plain text by default", func() {
		cmd := exec.Command(GrootFSBenchBin, "--gbin", FakeGrootFS, "--nospin", "--bundles", "10")
		buffer := gbytes.NewBuffer()
		cmd.Stdout = buffer
		err := cmd.Run()

		Expect(err).NotTo(HaveOccurred())
		Expect(buffer).Should(gbytes.Say("Total bundles requested"))
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

	Context("when grootfs fails", func() {
		It("returns the error message and the bundle number", func() {
			cmd := exec.Command(GrootFSBenchBin, "--gbin", FakeGrootFS, "--nospin", "--concurrency", "1", "--bundles", "1", "--image", "fail-this")
			sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).ShouldNot(gexec.Exit(0))

			Eventually(sess.Err).Should(gbytes.Say("could not create bundle 1: exit status 1, fake grootfs failed"))
		})
	})
})
