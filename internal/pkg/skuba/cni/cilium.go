/*
 * Copyright (c) 2019 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cni

import (
	"crypto/x509"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"sigs.k8s.io/yaml"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

const (
	ciliumSecretName      = "cilium-secret"
	ciliumConfigMapName   = "cilium-config"
	ciliumUpdateLabelsFmt = `{"spec":{"template":{"metadata":{"labels":{"caasp.suse.com/skuba-updated-at":"%v"}}}}}`
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
	etcdDir = filepath.Join("pki", "etcd")
)

type EtcdConfig struct {
	Endpoints []string `json:"endpoints"`
	CAFile    string   `json:"ca-file"`
	CertFile  string   `json:"cert-file"`
	KeyFile   string   `json:"key-file"`
}

func CreateCiliumSecret(client clientset.Interface) error {
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

	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Errorf("error when creating cilium secret  %v", err)
	}
	return nil
}

func CiliumSecretExists(client clientset.Interface) (bool, error) {
	_, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(ciliumSecretName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}

func CreateOrUpdateCiliumConfigMap(client clientset.Interface) error {
	etcdEndpoints := []string{}
	apiEndpoints, err := kubeadm.GetAPIEndpointsFromConfigMap(client)
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

	if err := apiclient.CreateOrUpdateConfigMap(client, ciliumConfigMap); err != nil {
		return errors.Wrap(err, "error when creating cilium config ")
	}

	return nil
}

func CiliumUpdateConfigMap(client clientset.Interface) error {
	if err := CreateOrUpdateCiliumConfigMap(client); err != nil {
		return err
	}
	return annotateCiliumDaemonsetWithCurrentTimestamp(client)
}

func CiliumConfigMapExists(client clientset.Interface) (bool, error) {
	_, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(ciliumConfigMapName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}

func annotateCiliumDaemonsetWithCurrentTimestamp(client clientset.Interface) error {
	patch := fmt.Sprintf(ciliumUpdateLabelsFmt, time.Now().Unix())
	_, err := client.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch("cilium", types.StrategicMergePatchType, []byte(patch))
	if err != nil {
		return err
	}

	klog.V(1).Info("successfully annotated cilium daemonset with current timestamp, which will restart all cilium pods")
	return nil
}
