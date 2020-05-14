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
	"context"
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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
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

	// retriesPreflightPod is the number of retries for cilium preflight
	// pod. This value means that we wait 12.5 minutes for it to become
	// available.
	retriesPreflightPod = 50
	// timeoutPreflightPod is the time to wait during each retry for cilium
	// preflight pod.
	timeoutPreflightPod = 15 * time.Second
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
	CAFile    string   `json:"trusted-ca-file"`
	CertFile  string   `json:"cert-file"`
	KeyFile   string   `json:"key-file"`
}

type EtcdConfigLegacy struct {
	Endpoints []string `json:"endpoints"`
	CAFile    string   `json:"ca-file"`
	CertFile  string   `json:"cert-file"`
	KeyFile   string   `json:"key-file"`
}

func CreateCiliumSecret(client clientset.Interface, ciliumVersion string) error {
	caCert, caKey, err := pkiutil.TryLoadCertAndKeyFromDisk(etcdDir, "ca")
	if err != nil {
		return errors.Errorf("etcd generation retrieval failed %v", err)
	}
	cfg := pkiutil.CertConfig{
		Config: ciliumCertConfig,
	}
	cert, key, err := pkiutil.NewCertAndKey(caCert, caKey, &cfg)
	if err != nil {
		return errors.Errorf("error when creating etcd client certificate for cilium %v", err)
	}

	privateKey, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Errorf("etcd private key marshal failed %v", err)
	}

	var secretData map[string][]byte
	switch {
	case strings.HasPrefix(ciliumVersion, "1.5"):
		secretData = map[string][]byte{
			v1.TLSCertKey:       pkiutil.EncodeCertPEM(cert),
			v1.TLSPrivateKeyKey: privateKey,
			"ca.crt":            pkiutil.EncodeCertPEM(caCert),
		}
	case strings.HasPrefix(ciliumVersion, "1.6"):
		secretData = map[string][]byte{
			"etcd-client-ca.crt": pkiutil.EncodeCertPEM(caCert),
			"etcd-client.key":    privateKey,
			"etcd-client.crt":    pkiutil.EncodeCertPEM(cert),
		}
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ciliumSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: secretData,
	}

	if err = apiclient.CreateOrUpdateSecret(client, secret); err != nil {
		return errors.Errorf("error when creating cilium secret  %v", err)
	}
	return nil
}

func CiliumSecretExists(client clientset.Interface) (bool, error) {
	_, err := client.CoreV1().Secrets(metav1.NamespaceSystem).Get(context.TODO(), ciliumSecretName, metav1.GetOptions{})
	return kubernetes.DoesResourceExistWithError(err)
}

// NeedsEtcdToCrdMigration checks if the migration from etcd to CRD is needed,
// which is the case when upgrading from Cilium 1.5 to Cilium 1.6. Decision
// depends on the old Cilium ConfigMap. If that config map exists and contains
// the etcd config, migration has to be done. If not, it means that we have a
// fresh deployment of Cilium 1.6 configured to use CRD and no migration is
// needed.
func NeedsEtcdToCrdMigration(client clientset.Interface, ciliumVersion string) (bool, error) {
	// Migration has to be done only when upgrading from 1.5 to 1.6.
	if !strings.HasPrefix(ciliumVersion, "1.6") {
		return false, nil
	}
	configMap, err := client.CoreV1().ConfigMaps(
		metav1.NamespaceSystem).Get(
		context.TODO(),
		ciliumConfigMapName,
		metav1.GetOptions{},
	)
	if err != nil {
		// If the old config map is not found, etcd config and migration
		// to CRD are not needed.
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "could not retrieve old cilium configmap, although it exists")
	}
	_, ok := configMap.Data["etcd-config"]
	return ok, nil
}

func getEtcdEndpoints(client clientset.Interface) ([]string, error) {
	etcdEndpoints := []string{}
	apiEndpoints, err := kubeadm.GetAPIEndpointsFromConfigMap(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get api endpoints")
	}
	for _, endpoints := range apiEndpoints {
		etcdEndpoints = append(etcdEndpoints, fmt.Sprintf(etcdEndpointFmt, endpoints))
	}
	return etcdEndpoints, nil
}

func marshalEtcdConfig(etcdConfigData interface{}) ([]byte, error) {
	etcdConfigDataByte, err := yaml.Marshal(&etcdConfigData)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal etcd config")
	}
	return etcdConfigDataByte, nil
}

func CreateOrUpdateCiliumConfigMap(client clientset.Interface, ciliumVersion string) error {
	var ciliumConfigMapData map[string]string
	switch {
	case strings.HasPrefix(ciliumVersion, "1.5"):
		etcdEndpoints, err := getEtcdEndpoints(client)
		if err != nil {
			return err
		}
		etcdConfigData := EtcdConfigLegacy{
			Endpoints: etcdEndpoints,
			CAFile:    "/tmp/cilium-etcd/ca.crt",
			CertFile:  "/tmp/cilium-etcd/tls.crt",
			KeyFile:   "/tmp/cilium-etcd/tls.key",
		}
		etcdConfigDataByte, err := marshalEtcdConfig(etcdConfigData)
		if err != nil {
			return err
		}
		ciliumConfigMapData = map[string]string{
			"debug":       "false",
			"enable-ipv4": "true",
			"enable-ipv6": "false",
			"etcd-config": string(etcdConfigDataByte),
		}
	case strings.HasPrefix(ciliumVersion, "1.6"):
		ciliumConfigMapData = map[string]string{
			"bpf-ct-global-tcp-max":    "524288",
			"bpf-ct-global-any-max":    "262144",
			"debug":                    "false",
			"enable-ipv4":              "true",
			"enable-ipv6":              "false",
			"identity-allocation-mode": "crd",
			"preallocate-bpf-maps":     "false",
		}

		needsEtcdConfig, err := NeedsEtcdToCrdMigration(client, ciliumVersion)
		if err != nil {
			return err
		}
		if needsEtcdConfig {
			etcdEndpoints, err := getEtcdEndpoints(client)
			if err != nil {
				return err
			}
			etcdConfigData := EtcdConfig{
				Endpoints: etcdEndpoints,
				CAFile:    "/var/lib/etcd-secrets/etcd-client-ca.crt",
				CertFile:  "/var/lib/etcd-secrets/etcd-client.crt",
				KeyFile:   "/var/lib/etcd-secrets/etcd-client.key",
			}
			etcdConfigDataByte, err := marshalEtcdConfig(etcdConfigData)
			if err != nil {
				return err
			}
			ciliumConfigMapData["etcd-config"] = string(etcdConfigDataByte)
			ciliumConfigMapData["identity-allocation-mode"] = "kvstore"
			ciliumConfigMapData["kvstore"] = "etcd"
			ciliumConfigMapData["kvstore-opt"] = "{\"etcd.config\": \"/var/lib/etcd-config/etcd.config\"}"
		}
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

// waitForCiliumPod waits until any Cilium preflight pod is running.
func waitForCiliumPod(client clientset.Interface) (string, error) {
	klog.Info("waiting for any running Cilium preflight pod")

	// Find any Cilium preflight pod.
	for i := 0; i < retriesPreflightPod; i++ {
		time.Sleep(timeoutPreflightPod)
		klog.V(4).Infof("waiting for cilium preflight pod - %d iteration", i)

		pods, err := client.CoreV1().Pods(metav1.NamespaceSystem).List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: "k8s-app=cilium-pre-flight-check",
			},
		)
		if err != nil {
			klog.Error("api call failed")
			return "", errors.Wrap(err, "api call for getting cilium preflight pods failed")
		}
		if len(pods.Items) < 1 {
			klog.Info("pods not found, continuing")
			continue
		}

		pod := &pods.Items[0]
		podName := pod.GetName()
		// Wait until the Cilium pod is not in the pending status and
		// check whether it's running.
		klog.Infof("checking availability of cilium preflight pod %s", podName)

		switch pod.Status.Phase {
		case v1.PodPending:
			klog.Info("cilium preflight pod status: pending")
		case v1.PodRunning:
			klog.Info("found cilium preflight pod")
			return podName, nil
		default:
			klog.Errorf("cilium preflight pod status: %s", pod.Status.Phase)
			return "", ErrCiliumPodUnsuccessful
		}
	}

	klog.V(4).Info("outside the waiting loop, returning an error")
	return "", ErrCiliumNotFound
}

// MigrateEtcdToCrd performs the migration of Cilium internal data from etcd
// cluster to CRD during upgrade from Cilium 1.5 to Cilium 1.6. This step is not
// mandatory, without it, Cilium is going to regenerate data from scratch which
// might result in service downtimes. If the automated migration is not
// successful, the upgrade will be continued without migration and user will be
// warned about downtime of services.
func MigrateEtcdToCrd(client clientset.Interface, config *rest.Config) error {
	klog.Info("starting migration from etcd to CRD as a data store for cilium")

	podName, err := waitForCiliumPod(client)
	if err != nil {
		return err
	}

	// Perform the migration.
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
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
		// If migration failed, log a warning instead of returning an
		// error.
		klog.Warningf(
			"Could not migrate data from etcd to CRD. Addons upgrade will be continued without it, which will result in temporary connection loss for currently existing pods and services. Error: %v; stdout: %v; stderr: %v",
			err, stdout.String(), stderr.String())
		return nil
	}

	klog.Info("successfully migrated from etcd to CRD")

	return nil
}

// RemoveEtcdConfig removes the etcd configuration from Cilium configmap and
// switched to CRD as the identity allocation backend. That function should
// be used after the migration from etcd to CRD is done.
func RemoveEtcdConfig(client clientset.Interface) error {
	cm, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(context.TODO(), ciliumConfigMapName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "could not get cilium config map")
	}
	cm.Data["identity-allocation-mode"] = "crd"
	delete(cm.Data, "etcd-config")
	delete(cm.Data, "kvstore")
	delete(cm.Data, "kvstore-opt")
	if _, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
		return errors.Wrap(err, "could not update cilium config map")
	}
	return nil
}

func CiliumUpdateConfigMap(client clientset.Interface, ciliumVersion string) error {
	if err := CreateOrUpdateCiliumConfigMap(client, ciliumVersion); err != nil {
		return err
	}
	return annotateCiliumDaemonsetWithCurrentTimestamp(client)
}

func annotateCiliumDaemonsetWithCurrentTimestamp(client clientset.Interface) error {
	patch := fmt.Sprintf(ciliumUpdateLabelsFmt, time.Now().Unix())
	_, err := client.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch(context.TODO(), "cilium", types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	klog.V(1).Info("successfully annotated cilium daemonset with current timestamp, which will restart all cilium pods")
	return nil
}
