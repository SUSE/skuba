package cni

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/kubernetes"
	"github.com/SUSE/caaspctl/pkg/caaspctl"
)

const (
	ciliumSecretName = "cilium-secret"
)

var (
	ciliumCertConfig = certutil.Config{
		CommonName:   "cilium-etcd-client",
		Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
)

type ciliumConfiguration struct {
	EtcdServer  string
	CiliumImage string
}

func renderCiliumTemplate(ciliumConfig ciliumConfiguration, file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not create file %s", file)
	}

	template, err := template.New("").Parse(string(content))
	if err != nil {
		return fmt.Errorf("could not parse template")
	}

	var rendered bytes.Buffer
	if err := template.Execute(&rendered, ciliumConfig); err != nil {
		return fmt.Errorf("could not render configuration")
	}

	if err := ioutil.WriteFile(file, rendered.Bytes(), 0644); err != nil {
		return fmt.Errorf("could not write to %s: %s", file, err)
	}

	return nil
}

func FillCiliumManifestFile(target, file string) error {
	ciliumImage := images.GetGenericImage(caaspctl.ImageRepository, "cilium",
		kubernetes.CurrentComponentVersion(kubernetes.Cilium))
	ciliumConfig := ciliumConfiguration{EtcdServer: target, CiliumImage: ciliumImage}

	etcdDir := filepath.Join("pki", "etcd")
	renderCiliumTemplate(ciliumConfig, filepath.Join("addons", "cni", "cilium.yaml"))
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(etcdDir, "ca")
	if err != nil {
		return fmt.Errorf("etcd generation retrieval failed %v", err)
	}

	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &ciliumCertConfig)
	if err != nil {
		return fmt.Errorf("error when creating etcd client certificate for cilium %v", err)
	}

	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return fmt.Errorf("etcd private key marshal failed %v", err)
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

	client := kubernetes.GetAdminClientSet()
	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return fmt.Errorf("error when creating cilium secret  %v", err)
	}

	return nil
}
