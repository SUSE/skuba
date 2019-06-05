package corefeatures

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

var _ = ginkgo.Describe("Create Skuba Cluster", func() {
	// ENV. parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	controlPlaneIP := os.Getenv("CONTROLPLANE") // ENV variable IP of controlplane
	master00IP := os.Getenv("MASTER00")         // IP of master 00
	worker00IP := os.Getenv("WORKER00")         // IP of worker 00

	// constants used by skuba
	clusterName := "e2e-cluster"
	master00Name := "master00"
	worker00Name := "worker00"

	// configuration of OS
	username := "sles"

	// Use an RPM binary provided by env variable otherwise use devel mode
	var skuba string
	skuba = os.Getenv("SKUBA_BIN_PATH")
	if len(skuba) == 0 {
		// use devel binary from gopath
		fmt.Println("Skuba binary path not specified: taking skuba from GOPATH")
		skuba = filepath.Join(os.Getenv("GOPATH"), "/bin/skuba")
	}

	// check binary exists
	if _, err := os.Stat(skuba); os.IsNotExist(err) {
		panic("skuba binary not found in GOPATH and ENV. variable: SKUBA_BIN_PATH !")
	}

	// wait 10 minutes max as timeout for completing command
	// the default timeout provided by ginkgo is 1 sec which is to low for us.
	gomega.SetDefaultEventuallyTimeout(600 * time.Second)
	gomega.SetDefaultEventuallyPollingInterval(5 * time.Second)
	gomega.SetDefaultConsistentlyDuration(600 * time.Second)
	gomega.SetDefaultConsistentlyPollingInterval(5 * time.Second)

	ginkgo.BeforeEach(func() {
		os.RemoveAll(clusterName)
	})

	ginkgo.It("00: Initialize cluster", func() {
		ginkgo.By("create configuration files")
		command := exec.Command(skuba, "cluster", "init", "--control-plane", controlPlaneIP, clusterName)
		session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
		gomega.Eventually(session.Out).Should(gbytes.Say(".*configuration files written to"))
		gomega.Expect(session).Should(gexec.Exit(), "configuration was not created")
		gomega.Expect(err).To(gomega.BeNil(), "configuration was not created")

		// change to created skuba directory
		err = os.Chdir(clusterName)
		if err != nil {
			panic(err)
		}

		ginkgo.By("add master00 to the cluster")
		command = exec.Command(skuba, "node", "bootstrap", "-v3", "--user", username, "--sudo", "--target", master00IP, master00Name)
		session, err = gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)

		// hack: we wait in this print until the command to finish. (if removed the following cmd fails because command hasn't finished)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba adding master00 failed")
		gomega.Expect(err).To(gomega.BeNil(), "skuba adding master00 failed")

		ginkgo.By("verify master00 with skuba status")
		command = exec.Command(skuba, "cluster", "status")
		session, err = gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)

		gomega.Eventually(session.Out).Should(gbytes.Say(".*" + master00Name))
		gomega.Expect(session).Should(gexec.Exit(0), "skuba status verify master00 failed")
		gomega.Expect(err).To(gomega.BeNil(), "skuba status verify master00 failed")

		ginkgo.By("add a worker00 to the cluster")
		command = exec.Command(skuba, "node", "join", "-v3", "--role", "worker", "--user", username, "--sudo", "--target", worker00IP, worker00Name)
		session, err = gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)

		// hack: we wait in this print until the command to finish. (if removed the following cmd fails because command hasn't finished)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba adding worker00 failed")
		gomega.Expect(err).To(gomega.BeNil(), "skuba adding worker00 failed")

		ginkgo.By("verify worker00 with skuba status")
		command = exec.Command(skuba, "cluster", "status")
		session, err = gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)

		gomega.Eventually(session.Out).Should(gbytes.Say(".*" + worker00Name))
		gomega.Expect(session).Should(gexec.Exit(0), "skuba status verify worker00 failed")
		gomega.Expect(err).To(gomega.BeNil(), "skuba status verify worker00 failed")

		ginkgo.By("verify all system pods are running")
		client, err := kubernetes.GetAdminClientSet()
		if err != nil {
			panic(err)
		}

		gomega.Eventually(func() []v1.Pod {
			podList, err := client.CoreV1().Pods("kube-system").List(metav1.ListOptions{FieldSelector: "status.phase!=Running"})
			if err != nil {
				panic(err)
			}
			return podList.Items
		}, 500*time.Second, 3*time.Second).ShouldNot(gomega.HaveLen(0), "Some system pods are not in 'running' state")
	})

})
