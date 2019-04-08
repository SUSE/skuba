package clusterhealth

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	//	. "suse.com/caaspctl/test/lib"
)

// TESTS spec:
//Add 2 master nodes to the cluster and check cluster-healt + logs

var _ = Describe("00-caaspctl-init: basics ", func() {
	// parameters , for convenience here but they should be global parameter, configurable.
	clusterName := "caaspci"
	controlPlane := os.Getenv("CONTROLPLANE")
	caaspctlPath := os.Getenv("GOPATH") //devel mode . For release this can be changed.
	caaspctl := caaspctlPath + "/bin/caaspctl"

	It("run caaspctl help", func() {
		output, err := exec.Command(caaspctl, "cluster", "-h").Output()
		if err != nil {
			fmt.Printf("[ERROR]: caaspctl cluster cluster help failed")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})
	// TODO use goexec from gomega here for better handling o errors
	// https://onsi.github.io/gomega/#gexec-testing-external-processes

	It("run cluster-init", func() {
		output, err := exec.Command(caaspctl, "cluster", "init", "--control-plane", controlPlane, clusterName).Output()
		if err != nil {
			fmt.Printf("[ERROR]: caaspctl cluster init failed")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})

	// TODO: add tear-down function

})
