package clusterhealth

import (
	"log"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// Tests goals;
// this test test some basic caaspctl basic:
// 1) help message utilty
// 2) caaspctl command should create a valid kubeadm configuration

var _ = Describe("caaspctl: CLI basics", func() {
	// parameters , for convenience here but they should be global parameter, configurable.
	clusterName := "caaspci"
	controlPlane := os.Getenv("CONTROLPLANE")  // ENV variable IP of controlplane
	caaspctlPath := os.Getenv("GOPATH")        // devel mode . For release this can be changed.
	caaspctl := caaspctlPath + "/bin/caaspctl" // this is the binary ( it can come from devel/release/rpm sources since path isn't hardcoded)

	It("run caaspct cluster help without errors", func() {
		command := exec.Command(caaspctl, "cluster", "-h")
		_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

	// We have more power with gexec
	// https://onsi.github.io/gomega/#gexec-testing-external-processes

	It("create caaspctl init configuration", func() {
		command := exec.Command(caaspctl, "cluster", "init", "--control-plane", controlPlane, clusterName)
		_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

	// TEARDOWN: ( this run after every IT test in this context)
	// since each test can be executed in parallel-random order, you need to think about the initial condition
	JustAfterEach(func() {

		// get current dir where we execute caaspctl so we remove the configuration dir
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		// we remove without checking for error because not all tests create the dir
		command := exec.Command("rm", "-rf", pwd+"/"+clusterName)
		gexec.Start(command, GinkgoWriter, GinkgoWriter)
	})
})
