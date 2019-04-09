package clusterhealth

import (
	"fmt"
	"os"
	"os/exec"

	// 00 ginkgo library
	. "github.com/onsi/ginkgo"

	// 01 extension: assertion/matchers
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

// gomega provide assertions/matchers and control on cmd execution

// 1)  run caaspctl version command without error
// 2)  check output with a value "Metallica next concert" run caaspctl version command
// 3)  for 1 and 2 use a teardown functions. ( 1 and 2 are executed in random order)

var _ = Describe("CLI/caaspctl: test versions", func() {
	// parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	controlPlane := os.Getenv("CONTROLPLANE")  // ENV variable IP of controlplane
	caaspctlPath := os.Getenv("GOPATH")        // devel mode . For release this can be changed.
	caaspctl := caaspctlPath + "/bin/caaspctl" // this is the binary ( it can come from devel/release/rpm sources since path isn't hardcoded)

	It("run caaspctl version command with exit 0", func() {
		command := exec.Command(caaspctl, "version")
		_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("caaspctl version command must be equal to version Metallica", func() {
		command := exec.Command(caaspctl, "version")
		session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Eventually(session.Out).Should(gbytes.Say("Metallica next concert"))
	})

	// TEARDOWN for each test-case: since each test can be executed in parallel-random order, you need to think about the initial condition
	JustAfterEach(func() {
		fmt.Println("[CLEANUP]: running after each testcase. Cleanup-things for idempotency ----- control-plane fake=destroyed:" + controlPlane)
	})
})
