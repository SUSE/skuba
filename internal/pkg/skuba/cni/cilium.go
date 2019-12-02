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
	"bufio"
	"bytes"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"sigs.k8s.io/yaml"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubectl"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
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

// IsMigrationToCrdNeeded checks whether Cilium kvstore needs to be migrated from
// etcd to CRD. The migration needs to be done if the Cilium config map before
// upgrade contains etcd configuration.
func IsMigrationToCrdNeeded(client clientset.Interface) (bool, error) {
	configMap, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(ciliumConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Errorf("could not get the cilium configmap: %v", err)
	}
	_, ok := configMap.Data["etcd-config"]
	return ok, nil
}

// MigrateEtcdToCrd performs the migration of data from etcd to CRD when
// upgrading Cilium from 1.5 to 1.6.
func MigrateEtcdToCrd(client clientset.Interface, config *rest.Config, manifest string) error {
	// Create ConfigMap for Cilium preflight deployment.
	if err := CreateOrUpdateCiliumConfigMap(client, true); err != nil {
		return err
	}
	// Apply Cilium preflight deployment.
	if err := kubectl.Apply(manifest); err != nil {
		return err
	}

	// Delete preflight deployment after migration is done (regardless whether
	// successful or not).
	defer func() {
		deletePolicy := metav1.DeletePropagationForeground
		if err := client.AppsV1().DaemonSets(metav1.NamespaceSystem).Delete(
			"cilium-pre-flight-check",
			&metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			},
		); err != nil {
			klog.Errorf("unable to delete cilium preflight deployment: %s", err)
		}
	}()

	// Find the Cilium preflight pod.
	var ciliumPreflightPod string
	if err := util.RetryOnError(retry.DefaultRetry, IsErrPreflightNotFound, func() error {
		pods, err := client.CoreV1().Pods(metav1.NamespaceSystem).List(metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, pod := range pods.Items {
			podName := pod.GetName()
			if strings.HasPrefix(podName, "cilium-pre-flight") {
				ciliumPreflightPod = podName
				break
			}
		}
		if ciliumPreflightPod == "" {
			return ErrPreflightNotFound
		}
		// Wait until the Cilium preflight pod is not in the pending status and
		// check whether status is successful.
		var pod *v1.Pod
		for {
			var err error
			pod, err = client.CoreV1().Pods(metav1.NamespaceSystem).Get(ciliumPreflightPod, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if pod.Status.Phase != v1.PodPending {
				break
			}
		}
		if pod.Status.Phase != v1.PodSucceeded {
			return ErrPreflightUnsuccessful
		}
		return nil
	}); err != nil {
		return err
	}

	// Perform the migration.
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(ciliumPreflightPod).
		Namespace(metav1.NamespaceSystem).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: []string{"cilium", "preflight", "migrate-identity"},
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	var stdout, stderr bytes.Buffer
	bStdout := bufio.NewWriter(&stdout)
	bStderr := bufio.NewWriter(&stderr)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: bStdout,
		Stderr: bStderr,
	})
	bStdout.Flush()
	bStderr.Flush()
	if err != nil {
		return errors.Errorf("could not migrate data from etcd to CRD: %v; stdout: %v; stderr: %v",
			err, stdout.String(), stderr.String())
	}

	return nil
}

func CreateOrUpdateCiliumConfigMap(client clientset.Interface, preflight bool) error {
	ciliumConfigMapData := map[string]string{
		"identity-allocation-mode": "crd",
		"debug":                    "true",
		"enable-ipv4":              "true",
		"enable-ipv6":              "false",
		"monitor-aggregation":      "medium",
		"bpf-ct-global-tcp-max":    "524288",
		"bpf-ct-global-any-max":    "262144",
		"preallocate-bpf-maps":     "false",
		"tunnel":                   "vxlan",
		// That value is relevant only when creating a mesh of clusters.
		// TODO(mrostecki): Support clustermesh in skuba.
		"cluster-name":          "default",
		"tofqdns-enable-poller": "false",
		"wait-bpf-mount":        "false",
		// This setting refers to the "workloads" functionality in Cilium
		// which is going to be removed from next releases. It can read
		// CRI-spefic data from the direct access to the CRI socket. We
		// don't need that.
		"container-runtime": "none",
		"masquerade":        "true",
		// "host-reachable-services-protos": "tcp",
		// We can probably disable that.
		"install-iptables-rules":  "true",
		"auto-direct-node-routes": "false",
		"enable-node-port":        "false",
	}

	if preflight {
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
		ciliumConfigMapData["etcd-config"] = string(etcdConfigDataByte)
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
	if err := CreateOrUpdateCiliumConfigMap(client, false); err != nil {
		return err
	}
	return annotateCiliumDaemonsetWithCurrentTimestamp(client)
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
