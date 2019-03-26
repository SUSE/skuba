package clusterhealth

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	//	. "suse.com/caaspctl/test/lib"
)

// TESTS spec:
//Add 2 master nodes to the cluster and check cluster-healt + logs

// TODO 02: this need  to be adapted better ( wip): e.g better detection of ip master etc.
// working on this now. ( to be though if using env. vars or files.

// use init function for reading IPs of cluster. This should be read all ips if possible.
// func init() {
// masterHost := os.Getenv("MASTER")
// if masterHost == "" {
// 	panic("MASTER IP not set")
// }
//}

// 00: add 1 worker and check status
var _ = Describe("Add 1 worker node to cluster", func() {
	/// TODO 03: this will run remote cmd via ssh

	// It("Add 1 worker node to cluster", func() {
	// 	output, err := RunCmd("localhost", "caaspctl cluster status")
	// 	if err != nil {
	// 		fmt.Printf("[ERROR]: Cluster is not healthy")
	// 		panic(err)
	// 	}
	// 	fmt.Println(string(output))
	// })

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
