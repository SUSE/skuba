package corefeatures

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
        "github.com/onsi/gomega/gbytes"
	"os"
        "time"
    	"os/exec"
)

var _ = Describe("Create Caaspctl Cluster", func() {

        // TODO: all this variables can be refactored in a pkg where we read them one time and then pass to other
        // files.

	// ENV. parameters , for convenience here but they can be global parameter, configurable and passed to testsuite
	controlPlaneIP := os.Getenv("CONTROLPLANE")  // ENV variable IP of controlplane
        master00IP := os.Getenv("MASTER00") // IP of master 00
        worker00IP := os.Getenv("WORKER00") // IP of worker 00      
 
        // constants
        clusterName := "e2e-cluster"
        master00Name := "master00"
        worker00Name := "worker00"

        // configuration
        username := "sles"

        // binary variables[init] configuration files written to
	caaspctlPath := os.Getenv("GOPATH")        // devel mode . For release this can be changed.
 	caaspctl := caaspctlPath + "/bin/caaspctl" // this is the binary ( it can come from devel/release/rpm sources since path isn't hardcoded)
        
        // wait  30 minutes max as timeout for completing command
      	SetDefaultEventuallyTimeout(1800 * time.Second)
        SetDefaultEventuallyPollingInterval(5 * time.Second)
        SetDefaultConsistentlyDuration(1800 * time.Second)
        SetDefaultConsistentlyPollingInterval(5 * time.Second)

	It("00: Initialize cluster", func() {
                By("create configuration files")
		command := exec.Command(caaspctl, "cluster", "init", "--control-plane", controlPlaneIP, clusterName)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
                Eventually(session.Out).Should(gbytes.Say(".*configuration files written to"))
                Expect(session).Should(gexec.Exit(), "configuration was not created")
                Expect(err).To(BeNil(), "configuration was not created")
                // change to created caaspctl directory
                err = os.Chdir(clusterName)
                if err != nil {
                  panic(err)
                }

                By("add master00 to the cluster")
         	command = exec.Command(caaspctl, "node", "bootstrap", "--user", username, "--sudo", "--target", master00IP, master00Name)
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
                
                // cmd should be execute without error
                Expect(session.Wait().Out.Contents()).Should(ContainSubstring("kubeadm.init applied successfully"))
                Expect(session).Should(gexec.Exit(), "caaspctl adding master00 failed:")
                Expect(err).To(BeNil(), "caaspctl adding master00 failed:")	
	
                By("verify master00 with caaspctl status")
         	command = exec.Command(caaspctl, "cluster", "status")
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
                // check if master is present
                Eventually(session.Out).Should(gbytes.Say(".*"+ master00Name))
                Expect(session).Should(gexec.Exit(), "caaspctl status verify master00 failed")
                Expect(err).To(BeNil(), "caaspctl status verify master00 failed")

                By("add a worker00 to the cluster")
 	 	command = exec.Command(caaspctl, "node", "join", "--role", "worker", "--user", username, "--sudo", "--target", worker00IP, worker00Name)
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
                // cmd should be execute without error
                Eventually(session.Out).Should(gbytes.Say(".*state kubeadm.join applied successfully"))
                Expect(session).Should(gexec.Exit(), "caaspctl adding worker00 failed:")
                Expect(err).To(BeNil(), "caaspctl adding worker00 failed:")	
        
                By("verify worker00 with caaspctl status")
         	command = exec.Command(caaspctl, "cluster", "status")
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
                // check if master is present
                Eventually(session.Out).Should(gbytes.Say(".*"+ worker00Name))
                Expect(session).Should(gexec.Exit(), "caaspctl status verify worker00 failed")
                Expect(err).To(BeNil(), "caaspctl status verify worker00 failed")

       	})

})
