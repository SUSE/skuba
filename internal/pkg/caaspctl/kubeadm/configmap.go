package kubeadm

import (
	"log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"

	"suse.com/caaspctl/internal/pkg/caaspctl/kubernetes"
)

func RemoveAPIEndpointFromConfigMap(node *v1.Node) error {
	kubeadmConfig, err := kubernetes.GetAdminClientSet().CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(kubeadmconstants.KubeadmConfigConfigMap, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("could not retrieve the kubeadm-config configmap to change the apiEndpoints\n")
	}
	clusterStatus := &kubeadmapi.ClusterStatus{}
	if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), []byte(kubeadmConfig.Data[kubeadmconstants.ClusterStatusConfigMapKey]), clusterStatus); err != nil {
		log.Fatalf("could not unmarshal cluster status from kubeadm-config configmap\n")
	}
	delete(clusterStatus.APIEndpoints, node.ObjectMeta.Name)
	clusterStatusYaml, err := configutil.MarshalKubeadmConfigObject(clusterStatus)
	if err != nil {
		log.Fatalf("could not marshal modified cluster status\n")
	}
	_, err = kubernetes.GetAdminClientSet().CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeadmconstants.KubeadmConfigConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			kubeadmconstants.ClusterConfigurationConfigMapKey: kubeadmConfig.Data[kubeadmconstants.ClusterConfigurationConfigMapKey],
			kubeadmconstants.ClusterStatusConfigMapKey:        string(clusterStatusYaml),
		},
	})
	if err != nil {
		log.Fatalf("could not update kubeadm-config configmap\n")
	}
	return nil
}
