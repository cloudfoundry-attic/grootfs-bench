package integration_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	GrootFSBenchBin string
	FakeGrootFS     string
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)

	SynchronizedBeforeSuite(func() []byte {
		var err error
		bins := make(map[string]string)

		bins["grootfsBenchBin"], err = gexec.Build("code.cloudfoundry.org/grootfs-bench")
		Expect(err).NotTo(HaveOccurred())

		bins["fakeGrootFSBin"], err = gexec.Build("code.cloudfoundry.org/grootfs-bench/integration/fakegrootfs")
		Expect(err).NotTo(HaveOccurred())

		data, err := json.Marshal(bins)
		Expect(err).NotTo(HaveOccurred())

		return data

	}, func(data []byte) {
		bins := make(map[string]string)
		Expect(json.Unmarshal(data, &bins)).To(Succeed())

		GrootFSBenchBin = bins["grootfsBenchBin"]
		FakeGrootFS = bins["fakeGrootFSBin"]
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		gexec.CleanupBuildArtifacts()
	})

	RunSpecs(t, "Integration Suite")
}
