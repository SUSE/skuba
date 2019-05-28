package cni

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

const (
	ciliumSecretName      = "cilium-secret"
	ciliumConfigMapName   = "cilium-config"
	ciliumUpdateLabelsFmt = `{"spec":{"template":{"metadata":{"labels":{"skuba-updated-at":"%v"}}}}}`
	etcdEndpointFmt       = "https://%s:2379"
	etcdCAFileName        = "/tmp/cilium-etcd/ca.crt"
	etcdCertFileName      = "/tmp/cilium-etcd/tls.crt"
	etcdKeyFileName       = "/tmp/cilium-etcd/tls.key"
)

var (
	ciliumCertConfig = certutil.Config{
		CommonName:   "cilium-etcd-client",
		Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
)

type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints"`
	CAFile    string   `yaml:"ca-file"`
	CertFile  string   `yaml:"cert-file"`
	KeyFile   string   `yaml:"key-file"`
}

type ciliumConfiguration struct {
	CiliumImage         string
	CiliumInitImage     string
	CiliumOperatorImage string
}

func renderCiliumTemplate(ciliumConfig ciliumConfiguration, file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Errorf("could not create file %s", file)
	}

	template, err := template.New("").Parse(string(content))
	if err != nil {
		return errors.Errorf("could not parse template")
	}

	var rendered bytes.Buffer
	if err := template.Execute(&rendered, ciliumConfig); err != nil {
		return errors.Errorf("could not render configuration")
	}

	if err := ioutil.WriteFile(file, rendered.Bytes(), 0644); err != nil {
		return errors.Errorf("could not write to %s: %s", file, err)
	}

	return nil
}

func CreateCiliumSecret() error {
	etcdDir := filepath.Join("pki", "etcd")
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(etcdDir, "ca")
	if err != nil {
		return errors.Errorf("etcd generation retrieval failed %v", err)
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &ciliumCertConfig)
	if err != nil {
		return errors.Errorf("error when creating etcd client certificate for cilium %v", err)
	}

	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Errorf("etcd private key marshal failed %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ciliumSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			v1.TLSCertKey:       pkiutil.EncodeCertPEM(cert),
			v1.TLSPrivateKeyKey: privateKey,
			"ca.crt":            pkiutil.EncodeCertPEM(caCert),
		},
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}
	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Errorf("error when creating cilium secret  %v", err)
	}
	return nil
}

func AnnotateCiliumDaemonsetWithCurrentTimestamp() error {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return err
	}
	patch := fmt.Sprintf(ciliumUpdateLabelsFmt, time.Now().Unix())
	_, err = client.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch("cilium", types.StrategicMergePatchType, []byte(patch))
	if err != nil {
		return err
	}

	klog.V(1).Info("successfully annotated cilium daemonset with current timestamp, which will restart all cilium pods")
	return nil
}

func CreateOrUpdateCiliumConfigMap() error {
	etcdEndpoints := []string{}
	apiEndpoints, err := kubeadm.GetAPIEndpointsFromConfigMap()
	if err != nil {
		return errors.Wrap(err, "unable to get api endpoints")
	}
	for _, endpoints := range apiEndpoints {
		etcdEndpoints = append(etcdEndpoints, fmt.Sprintf(etcdEndpointFmt, endpoints))
	}
	etcdConfigData := EtcdConfig{
		Endpoints: etcdEndpoints,
		CAFile:    etcdCAFileName,
		CertFile:  etcdCertFileName,
		KeyFile:   etcdKeyFileName,
	}

	etcdConfigDataByte, err := yaml.Marshal(&etcdConfigData)
	if err != nil {
		return err
	}
	ciliumConfigMapData := map[string]string{
		"debug":        "false",
		"disable-ipv4": "false",
		"etcd-config":  string(etcdConfigDataByte),
	}
	ciliumConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ciliumConfigMapName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: ciliumConfigMapData,
	}
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}
	if err := apiclient.CreateOrUpdateConfigMap(client, ciliumConfigMap); err != nil {
		return errors.Wrap(err, "error when creating cilium config ")
	}

	return nil
}

func FillCiliumManifestFile(target, file string) error {
	ciliumImage := images.GetGenericImage(skuba.ImageRepository, "cilium",
		kubernetes.CurrentAddonVersion(kubernetes.Cilium))
	ciliumInitImage := images.GetGenericImage(skuba.ImageRepository, "cilium-init",
		kubernetes.CurrentAddonVersion(kubernetes.Cilium))
	ciliumOperatorImage := images.GetGenericImage(skuba.ImageRepository, "cilium-operator",
		kubernetes.CurrentAddonVersion(kubernetes.Cilium))
	ciliumConfig := ciliumConfiguration{
		CiliumImage:         ciliumImage,
		CiliumInitImage:     ciliumInitImage,
		CiliumOperatorImage: ciliumOperatorImage,
	}

	return renderCiliumTemplate(ciliumConfig, filepath.Join("addons", "cni", "cilium.yaml"))
}
