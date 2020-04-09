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

package kubernetes

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/SUSE/skuba/pkg/skuba"
)

const (
	// KubeletCertAndKeyDir defines kubelet's CA certificate and key directory
	KubeletCertAndKeyDir = "/var/lib/kubelet/pki"

	// KubeletCACertAndKeyBaseName defines kubelet's CA certificate and key base name
	KubeletCACertAndKeyBaseName = "kubelet-ca"
	// KubeletCACertName defines kubelet's CA certificate name
	KubeletCACertName = "kubelet-ca.crt"
	// KubeletCAKeyName defines kubelet's CA key name
	KubeletCAKeyName = "kubelet-ca.key"

	// KubeletServerCertAndKeyBaseName defines kubelet server certificate and key base name
	KubeletServerCertAndKeyBaseName = "kubelet"
	// KubeletServerCertName defines kubelet server certificate name
	KubeletServerCertName = "kubelet.crt"
	// KubeletServerKeyName defines kubelet server key name
	KubeletServerKeyName = "kubelet.key"
)

// GenerateKubeletRootCert generates kubelet root CA certificate and key
// and save the generated file locally.
func GenerateKubeletRootCert() error {
	caCertFilePath := filepath.Join(skuba.PkiDir(), KubeletCACertName)
	caKeyFilePath := filepath.Join(skuba.PkiDir(), KubeletCAKeyName)
	if canReadCertAndKey, _ := certutil.CanReadCertAndKey(caCertFilePath, caKeyFilePath); canReadCertAndKey {
		klog.V(1).Info("kubelet root ca cert and key already exists")
		return nil
	}

	cfg := &certutil.Config{
		CommonName: "kubelet-ca",
	}
	caCert, caKey, err := pkiutil.NewCertificateAuthority(cfg)
	if err != nil {
		return errors.Wrap(err, "couldn't generate kubelet CA certificate")
	}

	if err := pkiutil.WriteCertAndKey(skuba.PkiDir(), KubeletCACertAndKeyBaseName, caCert, caKey); err != nil {
		return errors.Wrap(err, "failure while saving kubelet CA certificate and key")
	}

	return nil
}

func DisarmKubelet(client clientset.Interface, node *v1.Node, clusterVersion *version.Version) error {
	return CreateAndWaitForJob(
		client,
		disarmKubeletJobName(node),
		disarmKubeletJobSpec(node, clusterVersion),
		TimeoutWaitForJob,
	)
}

func disarmKubeletJobName(node *v1.Node) string {
	return fmt.Sprintf("caasp-kubelet-disarm-%x",
		sha1.Sum([]byte(node.ObjectMeta.Name)))
}

func disarmKubeletJobSpec(node *v1.Node, clusterVersion *version.Version) batchv1.JobSpec {
	privilegedJob := true
	return batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				ServiceAccountName: "kube-disarm",
				Containers: []v1.Container{
					{
						Name:  disarmKubeletJobName(node),
						Image: ComponentContainerImageForClusterVersion(Tooling, clusterVersion),
						Command: []string{
							"/bin/bash", "-c",
							strings.Join(
								[]string{
									"rm -rf /etc/kubernetes/*",
									"rm -rf /var/lib/kubelet/pki/*",
									"rm -f /var/lib/kubelet/{config.yaml,kubeadm-flags.env}",
									"rm -rf /var/lib/etcd/*",
									"dbus-send --system --print-reply --dest=org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager.DisableUnitFiles array:string:'kubelet.service' boolean:false",
									"dbus-send --system --print-reply --dest=org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager.MaskUnitFiles array:string:'kubelet.service' boolean:false boolean:true",
								},
								" && ",
							),
						},
						VolumeMounts: []v1.VolumeMount{
							VolumeMount("etc-kubernetes", "/etc/kubernetes", VolumeMountReadWrite),
							VolumeMount("var-lib-kubelet", "/var/lib/kubelet", VolumeMountReadWrite),
							VolumeMount("var-lib-etcd", "/var/lib/etcd", VolumeMountReadWrite),
							VolumeMount("var-run-dbus", "/var/run/dbus", VolumeMountReadWrite),
						},
						SecurityContext: &v1.SecurityContext{
							Privileged: &privilegedJob,
						},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
				Volumes: []v1.Volume{
					HostMount("etc-kubernetes", "/etc/kubernetes"),
					HostMount("var-lib-kubelet", "/var/lib/kubelet"),
					HostMount("var-lib-etcd", "/var/lib/etcd"),
					HostMount("var-run-dbus", "/var/run/dbus"),
				},
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": node.ObjectMeta.Name,
				},
				Tolerations: []v1.Toleration{
					{
						Operator: v1.TolerationOpExists,
					},
				},
			},
		},
	}
}
