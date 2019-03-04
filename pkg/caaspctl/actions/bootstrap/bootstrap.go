package bootstrap

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"

	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
	"suse.com/caaspctl/pkg/caaspctl"
)

func Bootstrap(target deployments.Target) error {
	initConfiguration, err := configFileAndDefaultsToInternalConfig(caaspctl.KubeadmInitConfFile())
	if err != nil {
		return fmt.Errorf("Could not parse %s file: %v", caaspctl.KubeadmInitConfFile(), err)
	}
	addTargetInformationToInitConfiguration(target, initConfiguration)
	finalInitConfigurationContents, err := kubeadmconfigutil.MarshalInitConfigurationToBytes(initConfiguration, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: "v1beta1",
	})
	if err != nil {
		return fmt.Errorf("Could not marshal configuration: %v", err)
	}

	if err := ioutil.WriteFile(caaspctl.KubeadmInitConfFile(), finalInitConfigurationContents, 0600); err != nil {
		return fmt.Errorf("Error writing init configuration: %v", err)
	}

	err = target.Apply(
		nil,
		"kubelet.configure",
		"kubelet.enable",
		"kubeadm.init",
		"cni.deploy",
	)
	if err != nil {
		return err
	}

	return downloadSecrets(target)
}

func downloadSecrets(target deployments.Target) error {
	os.MkdirAll(path.Join("pki", "etcd"), 0700)

	for _, secretLocation := range deployments.Secrets {
		secretData, err := target.DownloadFileContents(path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(secretLocation, []byte(secretData), 0600); err != nil {
			return err
		}
	}

	return nil
}

func addTargetInformationToInitConfiguration(target deployments.Target, initConfiguration *kubeadmapi.InitConfiguration) {
	if ip := net.ParseIP(target.Node()); ip != nil {
		if initConfiguration.NodeRegistration.KubeletExtraArgs == nil {
			initConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
		}
		initConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = target.Node()
	}
}

func configFileAndDefaultsToInternalConfig(cfgPath string) (*kubeadmapi.InitConfiguration, error) {
	internalcfg := &kubeadmapi.InitConfiguration{}

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	internalcfg, err = kubeadmconfigutil.BytesToInternalConfig(b)
	if err != nil {
		return nil, err
	}

	return internalcfg, nil
}
