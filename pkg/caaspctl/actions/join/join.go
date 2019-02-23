package join

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config/strict"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

const (
	MasterRole = iota
	WorkerRole = iota
)

type Role int

type JoinConfiguration struct {
	Target salt.Target
	Role   Role
}

func Join(joinConfiguration JoinConfiguration, masterConfig salt.MasterConfig) {
	statesToApply := []string{"kubelet.enable", "kubeadm.join"}

	pillar := &salt.Pillar{
		Join: &salt.Join{
			Kubeadm: salt.Kubeadm{
				ConfigPath: configPath(joinConfiguration.Role, joinConfiguration.Target.Node),
			},
		},
	}

	if joinConfiguration.Role == MasterRole {
		statesToApply = append([]string{"kubernetes.upload-secrets"}, statesToApply...)
		pillar.Join.Kubernetes = &salt.Kubernetes{
			AdminConfPath: "salt://admin.conf",
			SecretsPath:   "salt://pki",
		}
	}

	salt.Apply(joinConfiguration.Target, masterConfig, pillar, statesToApply...)
}

func specificPath(target string) string {
	return path.Join(
		"kubeadm-join.conf.d",
		fmt.Sprintf("%s.conf", target),
	)
}

func templatePath(role Role) string {
	switch role {
	case MasterRole:
		return path.Join("kubeadm-join.conf.d", "master.conf.template")
	case WorkerRole:
		return path.Join("kubeadm-join.conf.d", "worker.conf.template")
	}
	return ""
}

func configPath(role Role, target string) string {
	configPath := specificPath(target)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = templatePath(role)
	}

	joinConfiguration, err := joinConfigFileAndDefaultsToInternalConfig(configPath)
	if err != nil {
		log.Fatal("error parsing configuration: %v", err)
	}
	addFreshTokenToJoinConfiguration(target, joinConfiguration)
	addTargetInformationToJoinConfiguration(target, role, joinConfiguration)
	finalJoinConfigurationContents, err := kubeadmconfigutil.MarshalKubeadmConfigObject(joinConfiguration)
	if err != nil {
		log.Fatal("could not marshal configuration")
	}

	if err := ioutil.WriteFile(specificPath(target), finalJoinConfigurationContents, 0600); err != nil {
		log.Fatal("error writing specific machine configuration")
	}

	return fmt.Sprintf("salt://%s", specificPath(target))
}

func addFreshTokenToJoinConfiguration(target string, joinConfiguration *kubeadmapi.JoinConfiguration) {
	if joinConfiguration.Discovery.BootstrapToken == nil {
		joinConfiguration.Discovery.BootstrapToken = &kubeadmapi.BootstrapTokenDiscovery{}
	}
	joinConfiguration.Discovery.BootstrapToken.Token = createBootstrapToken(target)
	joinConfiguration.Discovery.TLSBootstrapToken = ""
}

func addTargetInformationToJoinConfiguration(target string, role Role, joinConfiguration *kubeadmapi.JoinConfiguration) {
	if ip := net.ParseIP(target); ip != nil {
		if joinConfiguration.NodeRegistration.KubeletExtraArgs == nil {
			joinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
		}
		joinConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = target
	}
}

// FIXME: do not shell out
func createBootstrapToken(target string) string {
	cmd := exec.Command(
		"kubeadm", "token", "create", "--kubeconfig", "admin.conf",
		"--description", fmt.Sprintf("Bootstrap token for machine %s'", target),
		"--ttl", "15m",
	)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	stdout := stdOut.String()
	stderr := stdErr.String()
	if err != nil {
		log.Fatalf("could not create token for joining a new node; stdout: %q, stderr: %q\n", stdout, stderr)
	}
	return strings.TrimSpace(stdout)
}

func joinConfigFileAndDefaultsToInternalConfig(cfgPath string) (*kubeadmapi.JoinConfiguration, error) {
	internalcfg := &kubeadmapi.JoinConfiguration{}

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	if err := kubeadmconfigutil.DetectUnsupportedVersion(b); err != nil {
		return nil, err
	}

	gvkmap, err := kubeadmutil.SplitYAMLDocuments(b)
	if err != nil {
		return nil, err
	}

	joinBytes := []byte{}
	for gvk, bytes := range gvkmap {
		if gvk.Kind == constants.JoinConfigurationKind {
			joinBytes = bytes
			// verify the validity of the YAML
			strict.VerifyUnmarshalStrict(bytes, gvk)
		}
	}

	if len(joinBytes) == 0 {
		return nil, errors.New("invalid config")
	}

	if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), joinBytes, internalcfg); err != nil {
		return nil, err
	}

	return internalcfg, nil
}
