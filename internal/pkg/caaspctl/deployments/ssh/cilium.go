package ssh

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	pkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

var (
	ciliumSecretName = "cilium-secret"

	ciliumCertConfig = certutil.Config{
		CommonName: "cilium-etcd-client",
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
)

type ciliumConfiguration struct {
	EtcdServer string
}

func renderCiliumTemplate(t *Target, file string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("could not create file %s\n", file)
	}

	template, err := template.New("").Parse(string(content))
	if err != nil {
		return fmt.Errorf("could not parse template")
	}

	var rendered bytes.Buffer
	if err := template.Execute(&rendered, ciliumConfiguration{EtcdServer: t.target.Target}); err != nil {
		return fmt.Errorf("could not render configuration")
	}

	if err := ioutil.WriteFile(file, rendered.Bytes(), 0644); err != nil {
		return fmt.Errorf("could not write to %s: %s", file, err)
	}

	return nil
}

func fillCiliumManifestFile(t *Target, file string) error {
	etcdDir := filepath.Join("./pki/", "etcd")
	renderCiliumTemplate(t, "addons/cni/cilium.yaml")
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(etcdDir, "ca")
	if err != nil {
		return fmt.Errorf("etcd generation retrieval failed %v", err)
	}

	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &ciliumCertConfig)
	if err != nil {
		return fmt.Errorf("error when creating etcd client certificate for cilium %v", err)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ciliumSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Type: v1.SecretTypeTLS,
		Data: map[string][]byte{
			v1.TLSCertKey:       certutil.EncodeCertPEM(cert),
			v1.TLSPrivateKeyKey: certutil.EncodePrivateKeyPEM(key),
			"ca.crt":            certutil.EncodeCertPEM(caCert),
		},
	}

	client, err := kubeconfigutil.ClientSetFromFile("./admin.conf")
	if err != nil {
		return err
	}

	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return fmt.Errorf("error when creating cilium secret  %v", err)
	}

	return nil
}
