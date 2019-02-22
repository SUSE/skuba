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

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

var (
	secrets = []string{
		"pki/ca.crt",
		"pki/ca.key",
		"pki/sa.key",
		"pki/sa.pub",
		"pki/front-proxy-ca.crt",
		"pki/front-proxy-ca.key",
		"pki/etcd/ca.crt",
		"pki/etcd/ca.key",
		"admin.conf",
	}
)

type BootstrapConfiguration struct {
	Target salt.Target
}

func Bootstrap(bootstrapConfiguration BootstrapConfiguration, masterConfig salt.MasterConfig) error {
	initConfiguration, err := configFileAndDefaultsToInternalConfig("kubeadm-init.conf")
	if err != nil {
		return fmt.Errorf("Could not parse kubeadm-init.conf file: %v", err)
	}
	addTargetInformationToInitConfiguration(bootstrapConfiguration.Target.Node, initConfiguration)
	finalInitConfigurationContents, err := kubeadmconfigutil.MarshalInitConfigurationToBytes(initConfiguration, schema.GroupVersion{
		Group:   "kubeadm.k8s.io",
		Version: "v1beta1",
	})
	if err != nil {
		return fmt.Errorf("Could not marshal configuration: %v", err)
	}

	if err := ioutil.WriteFile("kubeadm-init.conf", finalInitConfigurationContents, 0600); err != nil {
		return fmt.Errorf("Error writing init configuration: %v", err)
	}

	err = salt.Apply(
		bootstrapConfiguration.Target,
		masterConfig,
		&salt.Pillar{
			Bootstrap: &salt.Bootstrap{
				salt.Kubeadm{
					ConfigPath: "salt://kubeadm-init.conf",
				},
				salt.Cni{
					ConfigPath: "salt://addons/cni/flannel.yaml",
				},
			},
		},
		"kubelet.enable",
		"kubeadm.init",
		"cni.deploy",
	)

	if err != nil {
		return err
	}

	return downloadSecrets(bootstrapConfiguration.Target, masterConfig)
}

func downloadSecrets(target salt.Target, masterConfig salt.MasterConfig) error {
	os.MkdirAll(path.Join("pki", "etcd"), 0700)

	for _, secretLocation := range secrets {
		secretData, err := salt.DownloadFile(
			target,
			masterConfig,
			path.Join("/etc/kubernetes", secretLocation))
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(secretLocation, []byte(secretData), 0600); err != nil {
			return err
		}
	}

	return nil
}

func addTargetInformationToInitConfiguration(target string, initConfiguration *kubeadmapi.InitConfiguration) {
	if ip := net.ParseIP(target); ip != nil {
		// Node registration information
		if initConfiguration.NodeRegistration.KubeletExtraArgs == nil {
			initConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
		}
		initConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = target
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
