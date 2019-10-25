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

package addons

import "github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"

func init() {
	registerAddon(kubernetes.PSP, renderPSPTemplate, nil, highPriority)
}

func renderPSPTemplate(addonConfiguration AddonConfiguration) string {
	return pspManifest
}

const (
	pspManifest = `---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: suse.caasp.psp.privileged
  annotations:
    {{.AnnotatedVersion}}
{{- if .KubernetesVersionAtLeast "1.15.0" }}
    apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
{{- end }}
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: '*'
    seccomp.security.alpha.kubernetes.io/defaultProfileName: runtime/default
spec:
  # Privileged
  privileged: true
  # Volumes and File Systems
  volumes:
    # Kubernetes Pseudo Volume Types
    - configMap
    - secret
    - emptyDir
    - downwardAPI
    - projected
    - persistentVolumeClaim
    # Kubernetes Host Volume Types
    - hostPath
    # Networked Storage
    - nfs
    - rbd
    - cephFS
    - glusterfs
    - fc
    - iscsi
    # Cloud Volumes
    - cinder
    - gcePersistentDisk
    - awsElasticBlockStore
    - azureDisk
    - azureFile
    - vsphereVolume
  #allowedHostPaths: []
  readOnlyRootFilesystem: false
  # Users and groups
  runAsUser:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  # Privilege Escalation
  allowPrivilegeEscalation: true
  defaultAllowPrivilegeEscalation: true
  # Capabilities
  allowedCapabilities:
    - '*'
  defaultAddCapabilities: []
  requiredDropCapabilities: []
  # Host namespaces
  hostPID: true
  hostIPC: true
  hostNetwork: true
  hostPorts:
  - min: 0
    max: 65535
  seLinux:
    # SELinux is unused in CaaSP
    rule: 'RunAsAny'
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: suse:caasp:psp:privileged
rules:
  - apiGroups: ['extensions']
    resources: ['podsecuritypolicies']
    resourceNames: ['suse.caasp.psp.privileged']
    verbs: ['use']
---
# Allow CaaSP nodes to use the privileged
# PodSecurityPolicy.
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: suse:caasp:psp:privileged
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: suse:caasp:psp:privileged
subjects:
# Only authenticated users
- kind: Group
  apiGroup: rbac.authorization.k8s.io
  name: system:authenticated
# All ServiceAccounts in the 'kube-system' namespace
- kind: Group
  apiGroup: rbac.authorization.k8s.io
  name: system:serviceaccounts:kube-system
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: suse.caasp.psp.unprivileged
  annotations:
    {{.AnnotatedVersion}}
{{- if .KubernetesVersionAtLeast "1.15.0" }}
    apparmor.security.beta.kubernetes.io/allowedProfileNames: runtime/default
    apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
{{- end }}
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: runtime/default
    seccomp.security.alpha.kubernetes.io/defaultProfileName: runtime/default
spec:
  # Privileged
  privileged: false
  # Volumes and File Systems
  volumes:
    # Kubernetes Pseudo Volume Types
    - configMap
    - secret
    - emptyDir
    - downwardAPI
    - projected
    - persistentVolumeClaim
    # Networked Storage
    - nfs
    - rbd
    - cephFS
    - glusterfs
    - fc
    - iscsi
    # Cloud Volumes
    - cinder
    - gcePersistentDisk
    - awsElasticBlockStore
    - azureDisk
    - azureFile
    - vsphereVolume
  allowedHostPaths:
    # Note: We don't allow hostPath volumes above, but set this to a path we
    # control anyway as a belt+braces protection. /dev/null may be a better
    # option, but the implications of pointing this towards a device are
    # unclear.
    - pathPrefix: /opt/kubernetes-hostpath-volumes
  readOnlyRootFilesystem: false
  # Users and groups
  runAsUser:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  # Privilege Escalation
  allowPrivilegeEscalation: false
  defaultAllowPrivilegeEscalation: false
  # Capabilities
  allowedCapabilities: []
  defaultAddCapabilities: []
  requiredDropCapabilities: []
  # Host namespaces
  hostPID: false
  hostIPC: false
  hostNetwork: false
  hostPorts:
  - min: 0
    max: 65535
  # SELinux
  seLinux:
    # SELinux is unused in CaaSP
    rule: 'RunAsAny'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: suse:caasp:psp:unprivileged
rules:
  - apiGroups: ['extensions']
    resources: ['podsecuritypolicies']
    verbs: ['use']
    resourceNames: ['suse.caasp.psp.unprivileged']
---
# Allow all users and serviceaccounts to use the unprivileged
# PodSecurityPolicy
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: suse:caasp:psp:default
roleRef:
  kind: ClusterRole
  name: suse:caasp:psp:unprivileged
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: Group
  apiGroup: rbac.authorization.k8s.io
  name: system:serviceaccounts
- kind: Group
  apiGroup: rbac.authorization.k8s.io
  name: system:authenticated
`
)
