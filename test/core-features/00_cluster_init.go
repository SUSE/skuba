package corefeatures

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
	"os/exec"
)

var _ = Describe("Create Cluster", func() {
	// parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	controlPlane := os.Getenv("CONTROLPLANE")  // ENV variable IP of controlplane
	caaspctlPath := os.Getenv("GOPATH")        // devel mode . For release this can be changed.
	caaspctl := caaspctlPath + "/bin/caaspctl" // this is the binary ( it can come from devel/release/rpm sources since path isn't hardcoded)

	It("00-run caaspctl cluster init", func() {
		command := exec.Command(caaspctl, "cluster", "init", "--control-plane", controlPlane, "e2e-cluster")
		_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

})
