package join

import (
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeadmconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config/strict"
	kubeadmtokenphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/bootstraptoken/node"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"suse.com/caaspctl/pkg/caaspctl"
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/salt"
)

type JoinConfiguration struct {
	Target salt.Target
	Role   caaspctl.Role
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

	if joinConfiguration.Role == caaspctl.MasterRole {
		statesToApply = append([]string{"kubernetes.upload-secrets"}, statesToApply...)
		pillar.Join.Kubernetes = &salt.Kubernetes{
			AdminConfPath: salt.SaltPath(caaspctl.KubeConfigAdminFile()),
			SecretsPath:   salt.SaltPath(caaspctl.PkiDir()),
		}
	}

	salt.Apply(joinConfiguration.Target, masterConfig, pillar, statesToApply...)
}

func configPath(role caaspctl.Role, target string) string {
	configPath := caaspctl.MachineConfFile(target)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = caaspctl.TemplatePathForRole(role)
	}

	joinConfiguration, err := joinConfigFileAndDefaultsToInternalConfig(configPath)
	if err != nil {
		log.Fatalf("error parsing configuration: %v", err)
	}
	addFreshTokenToJoinConfiguration(target, joinConfiguration)
	addTargetInformationToJoinConfiguration(target, role, joinConfiguration)
	finalJoinConfigurationContents, err := kubeadmconfigutil.MarshalKubeadmConfigObject(joinConfiguration)
	if err != nil {
		log.Fatal("could not marshal configuration")
	}

	if err := ioutil.WriteFile(caaspctl.MachineConfFile(target), finalJoinConfigurationContents, 0600); err != nil {
		log.Fatal("error writing specific machine configuration")
	}

	return salt.SaltPath(caaspctl.MachineConfFile(target))
}

func addFreshTokenToJoinConfiguration(target string, joinConfiguration *kubeadmapi.JoinConfiguration) {
	if joinConfiguration.Discovery.BootstrapToken == nil {
		joinConfiguration.Discovery.BootstrapToken = &kubeadmapi.BootstrapTokenDiscovery{}
	}
	joinConfiguration.Discovery.BootstrapToken.Token = createBootstrapToken(target)
	joinConfiguration.Discovery.TLSBootstrapToken = ""
}

func addTargetInformationToJoinConfiguration(target string, role caaspctl.Role, joinConfiguration *kubeadmapi.JoinConfiguration) {
	if ip := net.ParseIP(target); ip != nil {
		if joinConfiguration.NodeRegistration.KubeletExtraArgs == nil {
			joinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{}
		}
		joinConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = target
	}
}

func createBootstrapToken(target string) string {
	client, err := kubeconfigutil.ClientSetFromFile(caaspctl.KubeConfigAdminFile())
	if err != nil {
		log.Fatal("could not load admin kubeconfig file")
	}

	internalCfg, err := kubeadmconfigutil.ConfigFileAndDefaultsToInternalConfig(caaspctl.KubeadmInitConfFile(), nil)
	if err != nil {
		log.Fatal("could not load init configuration")
	}

	bootstrapTokenRaw, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		log.Fatal("could not generate a new boostrap token")
	}

	bootstrapToken, err := kubeadmapi.NewBootstrapTokenString(bootstrapTokenRaw)
	if err != nil {
		log.Fatal("could not generate a new boostrap token")
	}

	internalCfg.BootstrapTokens = []kubeadmapi.BootstrapToken{
		kubeadmapi.BootstrapToken{
			Token: bootstrapToken,
		},
	}

	if err := kubeadmtokenphase.CreateNewTokens(client, internalCfg.BootstrapTokens); err != nil {
		log.Fatal("could not create new bootstrap token")
	}

	return bootstrapTokenRaw
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
