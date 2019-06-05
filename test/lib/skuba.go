package lib

import (
	"fmt"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"os"
	"os/exec"
	"path/filepath"
)

func getEnvWithDefault(variable string, defaultValue string) string {
	value := os.Getenv(variable)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func getSkubaPath() (string, error) {
	// Use a binary provided by env variable otherwise use devel mode
	skuba := os.Getenv("SKUBA_BIN_PATH")
	if len(skuba) == 0 {
		//TODO: what's the best way to report this warning?
		fmt.Println("Skuba binary path not specified: taking skuba from GOPATH")
		skuba = filepath.Join(os.Getenv("GOPATH"), "/bin/skuba")
	}

	// check binary exists
	if _, err := os.Stat(skuba); os.IsNotExist(err) {
		return "", errors.New("skuba binary not found in GOPATH and ENV. variable:SKUBA_BIN_PATH")
	}

	return skuba, nil
}

type Skuba struct {
	clusterName  string
	controlPlane string
	debugLevel   string
	skuba        string
	username     string
}

func NewSkuba(skuba string, username string, clusterName string, controlPlane string, debugLevel string) *Skuba {
	return &Skuba{
		clusterName:  clusterName,
		controlPlane: controlPlane,
		debugLevel:   debugLevel,
		skuba:        skuba,
		username:     username,
	}

}

// NewSkubaFromEnv builds a Skuba reading values from environment variables
func NewSkubaFromEnv() (*Skuba, error) {
	controlPlane := os.Getenv("CONTROLPLANE") // ENV variable IP of controlplane
	if len(controlPlane) == 0 {
		return nil, errors.New("Env variable 'CONTROLPLANE' is required")
	}

	clusterName := getEnvWithDefault("CLUSTERNAME", "e2e-cluster")

	username := getEnvWithDefault("SKUBA_USERNAME", "sles")

	debugLevel := getEnvWithDefault("SKUBA_DEBUG", "3")

	skuba, err := getSkubaPath()
	if err != nil {
		return nil, err
	}

	s := &Skuba{
		clusterName:  clusterName,
		controlPlane: controlPlane,
		debugLevel:   debugLevel,
		skuba:        skuba,
		username:     username,
	}

	return s, nil
}

// Init initializes the cluster
func (s *Skuba) Init() (*gexec.Session, error) {
	// Clear cluster directory
	os.RemoveAll(s.clusterName)

	command := exec.Command(s.skuba, "cluster", "init", "--control-plane", s.controlPlane, s.clusterName)
	return gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
}

func (s *Skuba) Bootstrap(master00IP string, master00Name string) (*gexec.Session, error) {
	command := exec.Command(s.skuba, "node", "bootstrap", "-v", s.debugLevel, "--user", s.username, "--sudo", "--target", master00IP, master00Name)
	command.Dir = s.clusterName
	return gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
}

func (s *Skuba) JoinNode(nodeIP string, nodeName string, nodeRole string) (*gexec.Session, error) {
	command := exec.Command(s.skuba, "node", "join", "-v", s.debugLevel, "--role", nodeRole, "--user", s.username, "--sudo", "--target", nodeIP, nodeName)
	command.Dir = s.clusterName
	return gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
}

func (s *Skuba) Status() (*gexec.Session, error) {
	command := exec.Command(s.skuba, "cluster", "status")
	command.Dir = s.clusterName
	return gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
}

func (s *Skuba) ClusterName() string {
	return s.clusterName
}

// GetClient returns a Client for the cluster
func (s *Skuba) GetClient() (*clientset.Clientset, error) {
	client, err := kubeconfigutil.ClientSetFromFile(s.clusterName + "/" + skuba.KubeConfigAdminFile())
	if err != nil {
		return nil, errors.Wrap(err, "could not load admin kubeconfig file")
	}
	return client, nil
}
