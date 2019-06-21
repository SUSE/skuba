package corefeatures

import (
	"fmt"
	"os"
	"time"

	testlib "github.com/SUSE/skuba/test/lib"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("Create Skuba Cluster", func() {

	// ENV. parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	master00IP := os.Getenv("MASTER00") // IP of master 00
	worker00IP := os.Getenv("WORKER00") // IP of worker 00

	master00Name := "master00"
	worker00Name := "worker00"

	skuba, err := testlib.NewSkubaFromEnv()
	// TODO: at this point it is not possible to use gomega assertions
	// Find a better way to report/handle this error
	if err != nil {
		panic(err)
	}

	// wait 10 minutes max as timeout for completing command
	// the default timeout provided by ginkgo is 1 sec which is to low for us.
	gomega.SetDefaultEventuallyTimeout(600 * time.Second)
	gomega.SetDefaultEventuallyPollingInterval(5 * time.Second)
	gomega.SetDefaultConsistentlyDuration(600 * time.Second)
	gomega.SetDefaultConsistentlyPollingInterval(5 * time.Second)

	ginkgo.It("00: Initialize cluster", func() {
		ginkgo.By("create configuration files")
		session, err := skuba.Init()
		gomega.Eventually(session.Out).Should(gbytes.Say(".*configuration files written to"))

		//TODO: this assertion should be Eventually and wait for Exit(0)
		gomega.Expect(err).Should(gomega.BeNil(), "configuration was not created")

		_, err = os.Stat(skuba.ClusterName())
		gomega.Expect(os.IsNotExist(err)).ShouldNot(gomega.BeTrue())

		ginkgo.By("add master00 to the cluster")
		session, err = skuba.Bootstrap(master00IP, master00Name)

		// hack: we wait in this print until the command to finish. (if removed the following cmd fails because command hasn't finished)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba adding master00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba adding master00 failed")

		ginkgo.By("verify master00 with skuba status")
		session, err = skuba.Status()

		gomega.Eventually(session.Out).Should(gbytes.Say(".*" + master00Name))
		gomega.Expect(session).Should(gexec.Exit(0), "skuba status verify master00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba status verify master00 failed")

		ginkgo.By("add a worker00 to the cluster")
		session, err = skuba.JoinNode(worker00IP, worker00Name, "worker")
		// hack: we wait in this print until the command to finish. (if removed the following cmd fails because command hasn't finished)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba adding worker00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba adding worker00 failed")

		ginkgo.By("verify worker00 with skuba status")
		session, err = skuba.Status()

		gomega.Eventually(session.Out).Should(gbytes.Say(".*" + worker00Name))
		gomega.Expect(session).Should(gexec.Exit(0), "skuba status verify worker00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba status verify worker00 failed")

		ginkgo.By("verify all system pods are running")
		client, err := skuba.GetClient()
		if err != nil {
			panic(err)
		}
		// retrieve debug info of failedPods for printing in case of failure
		var failedPods []v1.Pod
		gomega.Eventually(func() []v1.Pod {
			podList, err := client.CoreV1().Pods("kube-system").List(metav1.ListOptions{FieldSelector: "status.phase!=Running"})
			if err != nil {
				panic(err)
			}
			failedPods = podList.Items
			return podList.Items
			// the timeout 700 must be generous enough because the startup time of different pods can vary a lot, causing flakiness if small timeout
		}, 700*time.Second, 3*time.Second).ShouldNot(gomega.HaveLen(0), "Some system pods are not in 'running' state: %#v ", failedPods)

		ginkgo.By("remove worker00 with skuba")
		session, err = skuba.RemoveNode(worker00Name)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba removing worker00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba removing worker00 failed")

		ginkgo.By("reset worker00 with skuba")
		session, err = skuba.ResetNode(worker00IP)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba reset worker00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba reset worker00 failed")

		ginkgo.By("verify worker00 is not present with skuba status")
		session, err = skuba.Status()
		fmt.Println(session.Wait().Out.Contents())
		gomega.Eventually(session.Out).ShouldNot(gbytes.Say(".*" + worker00Name))
		gomega.Expect(session).Should(gexec.Exit(0), "skuba status exited with error")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba status returned errors, worker00 not removed correctly")

		ginkgo.By("remove master00 with skuba")
		session, err = skuba.RemoveNode(master00Name)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba removing master00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba removing master00 failed")

		ginkgo.By("reset master00 with skuba")
		session, err = skuba.ResetNode(master00IP)
		fmt.Println(session.Wait().Out.Contents())
		gomega.Expect(session).Should(gexec.Exit(0), "skuba reset master00 failed")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba reset master00 failed")

		ginkgo.By("verify master00 is not present with skuba status")
		session, err = skuba.Status()
		fmt.Println(session.Wait().Out.Contents())
		gomega.Eventually(session.Out).ShouldNot(gbytes.Say(".*" + master00Name))
		gomega.Expect(session).Should(gexec.Exit(0), "skuba status exited with error, master00 not removed correctly")
		gomega.Expect(err).Should(gomega.BeNil(), "skuba status returned errors, master00 not removed correctly")

	})

})
