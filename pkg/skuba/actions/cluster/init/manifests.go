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
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
apiServer:
  certSANs:
    - {{.ControlPlane}}
  extraArgs:
    oidc-issuer-url: https://{{.ControlPlane}}:32000
    oidc-client-id: oidc
    oidc-ca-file: /etc/kubernetes/pki/ca.crt
    oidc-username-claim: email
    oidc-groups-claim: groups
clusterName: {{.ClusterName}}
controlPlaneEndpoint: {{.ControlPlane}}:6443
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

	// TODO: This needs to handle `insecure = true` somehow as well - might
	// need an additional flag
	criRegistriesV2Template = `# For more information on this configuration file, see containers-registries.conf(5).
#
# Registries to search for images that are not fully-qualified.
# i.e. foobar.com/my_image:latest vs my_image:latest
unqualified-search-registries = ["docker.io"]

{{ if (ne .RegistryMirror.SourceRegistry "") }}
[[registry]]
location = "{{ .RegistryMirror.SourceRegistry }}"
mirror = [
  { location = "{{ .RegistryMirror.MirrorRegistry }}", insecure = {{ .RegistryMirror.Insecure }} }
]
{{ end }}
{{ if (eq .RegistryMirror.Insecure true) }}
[[registry]]
location = "{{ .RegistryMirror.MirrorRegistry }}"
insecure = {{ .RegistryMirror.Insecure }}
{{ end }}
`

	masterConfTemplate = `apiVersion: kubeadm.k8s.io/v1beta1
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{.ControlPlane}}:6443
    unsafeSkipCAVerification: true
controlPlane:
  localAPIEndpoint:
    advertiseAddress: ""
`

	workerConfTemplate = `apiVersion: kubeadm.k8s.io/v1beta1
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{.ControlPlane}}:6443
    unsafeSkipCAVerification: true
`
)
