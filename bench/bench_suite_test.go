package bench_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBench(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bench Suite")
}
