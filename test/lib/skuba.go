package lib

import (
	"errors"
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
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
	// Use an RPM binary provided by env variable otherwise use devel mode
	var skuba string
	skuba = os.Getenv("SKUBA_BIN_PATH")
	if len(skuba) == 0 {
		//TODO: what's the best way to report this warning?
		fmt.Println("Skuba binary path not specified: taking skuba from GOPATH")
		skuba = filepath.Join(os.Getenv("GOPATH"), "/bin/skuba")
	}

	// check binary exists
	if _, err := os.Stat(skuba); os.IsNotExist(err) {
		return "", errors.New("skuba binary not found in GOPATH and ENV. variable:SKUBA_BIN_PATH !")
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

// SkubaFromEnv builds a Skuba reading values from environment variables
func NewSkubaFromEnv() (*Skuba, error) {
	controlPlane := os.Getenv("CONTROLPLANE") // ENV variable IP of controlplane
	if len(controlPlane) == 0 {
		return nil, errors.New("Env variable 'CONTROLPLANE' is required")
	}

	//TODO: add CLUSTERNAME as an env variable to documentation
	clusterName := getEnvWithDefault("CLUSTERNAME", "e2e-cluster")

	username := getEnvWithDefault("SKUBA_USERNAME", "sles")

	//TODO: add SKUBA_DEBUG as an env variable to documentation
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
