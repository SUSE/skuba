package corefeatures

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCoreFeatures(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core-features caaspctl Suite")
}
