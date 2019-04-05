package cilium

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCilium(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cilium Suite")
}
