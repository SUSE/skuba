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

package cluster

const (
	kubeadmInitConf = `apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
bootstrapTokens: []
localAPIEndpoint:
  advertiseAddress: ""
{{- if eq .CloudProvider "aws" }}
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: "aws"
{{- end }}
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
apiServer:
  certSANs:
    - {{.ControlPlaneHost}}
  extraArgs:
    oidc-issuer-url: https://{{.ControlPlaneHost}}:32000
    oidc-client-id: oidc
    oidc-ca-file: /etc/kubernetes/pki/ca.crt
    oidc-username-claim: email
    oidc-groups-claim: groups
{{- if eq .CloudProvider "aws" }}
    cloud-provider: "aws"
controllerManager:
  extraArgs:
    cloud-provider: "aws"
    allocate-node-cidrs: "false"
{{- end }}
clusterName: {{.ClusterName}}
controlPlaneEndpoint: {{.ControlPlaneHostAndPort}}
dns:
  imageRepository: {{.ImageRepository}}
  imageTag: {{.CoreDNSImageTag}}
  type: CoreDNS
etcd:
  local:
    imageRepository: {{.ImageRepository}}
    imageTag: {{.EtcdImageTag}}
imageRepository: {{.ImageRepository}}
kubernetesVersion: {{.KubernetesVersion}}
networking:
  podSubnet: 10.244.0.0/16
  serviceSubnet: 10.96.0.0/12
useHyperKubeImage: true
`
	criDockerDefaultsConf = `## Path           : System/Management
## Description    : Extra cli switches for crio daemon
## Type           : string
## Default        : ""
## ServiceRestart : crio
#
CRIO_OPTIONS=--pause-image={{.PauseImage}}{{if not .StrictCapDefaults}} --default-capabilities="CHOWN,DAC_OVERRIDE,FSETID,FOWNER,NET_RAW,SETGID,SETUID,SETPCAP,NET_BIND_SERVICE,SYS_CHROOT,KILL,MKNOD,AUDIT_WRITE,SETFCAP"{{end}}
`

	masterConfTemplate = `apiVersion: kubeadm.k8s.io/v1beta1
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{.ControlPlaneHostAndPort}}
    unsafeSkipCAVerification: true
{{- if eq .CloudProvider "aws" }}
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: "aws"
{{- end }}
controlPlane:
  localAPIEndpoint:
    advertiseAddress: ""
`

	workerConfTemplate = `apiVersion: kubeadm.k8s.io/v1beta1
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{.ControlPlaneHostAndPort}}
    unsafeSkipCAVerification: true
{{- if eq .CloudProvider "aws" }}
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: "aws"
{{- end }}
`
)
