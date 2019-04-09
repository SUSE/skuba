package clusterhealth

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

// gomega provide assertions/matchers and control on cmd execution

// Tests goals;
// this test test some basic caaspctl basic:
// 1) help message utilty
// 2) caaspctl command should create a valid kubeadm configuration

var _ = Describe("CLI/caaspctl: test versions", func() {
	// parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	controlPlane := os.Getenv("CONTROLPLANE")  // ENV variable IP of controlplane
	caaspctlPath := os.Getenv("GOPATH")        // devel mode . For release this can be changed.
	caaspctl := caaspctlPath + "/bin/caaspctl" // this is the binary ( it can come from devel/release/rpm sources since path isn't hardcoded)

	It("run caaspctl version command with exit 0", func() {
		command := exec.Command(caaspctl, "version")
		_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		fmt.Printf("PASSING test, doing other stuff later, bye!")
	})

	// FAIL for fun: assert version output equal to 3000
	It("caaspctl version command must be equal to version 3000", func() {
		command := exec.Command(caaspctl, "version")
		session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Eventually(session.Out).Should(gbytes.Say("version 3000 sinatra-space"))
	})

	// TEARDOWN for each test-case: since each test can be executed in parallel-random order, you need to think about the initial condition
	JustAfterEach(func() {
		fmt.Println("[CLEANUP]: running after each testcase. Cleanup-things for idempotency ----- control-plane fake=destroyed:" + controlPlane)
	})
})
