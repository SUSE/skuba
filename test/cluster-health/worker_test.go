package clusterhealth

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("fake worker test", func() {
	/// TODO 03: this will run remote cmd via ssh

	It("0001: fake test passing", func() {
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})

	It("Ooo3 cluster status after 1 worker was added", func() {
		output, err := exec.Command("caaspctl cluster status").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})
})
