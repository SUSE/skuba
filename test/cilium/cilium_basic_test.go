package cilium

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	//	. "suse.com/caaspctl/test/lib"
)

// 00: add 1 worker and check status
var _ = Describe("Add 1 worker node to cluster", func() {
	/// TODO 03: this will run remote cmd via ssh

	It("000: fake test passing", func() {
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})

	It("Omg i'm failing again", func() {
		output, err := exec.Command("caaspctl cluster status").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})
})

// 00: add 1 worker and check status
var _ = Describe("awesome-feature", func() {
	/// TODO 03: this will run remote cmd via ssh

	It("000: fake test passing", func() {
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})

	It("i'm not stable enough :( ", func() {
		output, err := exec.Command("caaspctl cluster status").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})
})

// 00: add 1 worker and check status
var _ = Describe("002-fake-scenario", func() {
	/// TODO 03: this will run remote cmd via ssh

	It("000: fake test passing", func() {
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})

	It("Jason SMITH WAS HERE", func() {
		output, err := exec.Command("caaspctl cluster status").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})
})

// 00: add 1 worker and check status
var _ = Describe("0003-fake scenario", func() {
	/// TODO 03: this will run remote cmd via ssh

	It("000: fake test passing", func() {
		output, err := exec.Command("uptime").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})

	It("Check cluster status after 1 worker was added", func() {
		output, err := exec.Command("caaspctl cluster status").Output()
		if err != nil {
			fmt.Printf("[ERROR]: Cluster is not healthy")
			fmt.Println(string(output))
			panic(err)
		}
		fmt.Println(string(output))
	})
})
