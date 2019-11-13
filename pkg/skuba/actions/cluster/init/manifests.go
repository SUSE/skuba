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
{{ if .ComponentExtraArgs .Kubelet -}}
nodeRegistration:
  kubeletExtraArgs:
{{- range $key, $value := .ComponentExtraArgs .Kubelet }}
    {{ $key }}: "{{ $value }}"
{{- end -}}
{{ end }}
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
{{- if .ComponentExtraArgs .APIServer -}}
{{- range $key, $value := .ComponentExtraArgs .APIServer }}
    {{ $key }}: "{{ $value }}"
{{- end -}}
{{ end -}}
{{ if .ComponentExtraVolumes .APIServer }}
  extraVolumes:
{{- range .ComponentExtraVolumes .APIServer }}
    - name: {{.Name}}
      hostPath: {{.HostPath}}
      mountPath: {{.MountPath}}
      readOnly: {{.ReadOnly}}
      pathType: {{.PathType}}
{{- end -}}
{{ end }}
{{ if or (.ComponentExtraArgs .ControllerManager) (.ComponentExtraVolumes .ControllerManager) -}}
controllerManager:
{{- if .ComponentExtraArgs .ControllerManager }}
  extraArgs:
{{- range $key, $value := .ComponentExtraArgs .ControllerManager }}
    {{ $key }}: "{{ $value }}"
{{- end -}}
{{ end -}}
{{ if .ComponentExtraVolumes .ControllerManager }}
  extraVolumes:
{{- range .ComponentExtraVolumes .ControllerManager }}
    - name: {{.Name}}
      hostPath: {{.HostPath}}
      mountPath: {{.MountPath}}
      readOnly: {{.ReadOnly}}
      pathType: {{.PathType}}
{{- end -}}
{{ end -}}
{{ end }}
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
{{ if .ComponentExtraArgs .Kubelet -}}
nodeRegistration:
  kubeletExtraArgs:
{{- range $key, $value := .ComponentExtraArgs .Kubelet }}
    {{ $key }}: "{{ $value }}"
{{- end -}}
{{ end }}
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
{{ if .ComponentExtraArgs .Kubelet -}}
nodeRegistration:
  kubeletExtraArgs:
{{- range $key, $value := .ComponentExtraArgs .Kubelet }}
    {{ $key }}: "{{ $value }}"
{{- end -}}
{{ end }}
`
)
