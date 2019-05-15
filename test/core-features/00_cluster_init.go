package corefeatures

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
        "github.com/onsi/gomega/gbytes"
	"os"
        "fmt"
	"os/exec"
)

var _ = Describe("Create Caaspctl Cluster", func() {

        // TODO: all this variables can be refactored in a pkg where we read them one time and then pass to other
        // files.

	// parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	controlPlaneIP := os.Getenv("CONTROLPLANE")  // ENV variable IP of controlplane
        master00IP := os.Getenv("MASTER-00") // IP of master 00
 
        clusterName := "e2e-cluster"
        master00Name := "master00"

        // configuration
        username := "sles"

        // binary variables[init] configuration files written to
	caaspctlPath := os.Getenv("GOPATH")        // devel mode . For release this can be changed.
 	caaspctl := caaspctlPath + "/bin/caaspctl" // this is the binary ( it can come from devel/release/rpm sources since path isn't hardcoded)
      
	It("00: Initialize cluster", func() {
		command := exec.Command(caaspctl, "cluster", "init", "--control-plane", controlPlaneIP, clusterName)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
                fmt.Println(session.Out)
		Expect(err).ShouldNot(HaveOccurred())
                Eventually(session.Out).Should(gbytes.Say("This is a BETA release and NOT intended for production usage"))
                // change to created caaspctl directory
                err = os.Chdir(clusterName)
                if err != nil {
                  panic(err)
                }
                // add master
		command = exec.Command(caaspctl, "node", "bootstrap", "--user", username, "--sudo", "--target", master00IP, master00Name)
		_, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

       	})

})
