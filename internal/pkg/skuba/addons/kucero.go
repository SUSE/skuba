/*
 * Copyright (c) 2020 SUSE LLC.
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

import (
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	registerAddon(kubernetes.Kucero, renderKuceroTemplate, nil, kuceroCallbacks{}, normalPriority, []getImageCallback{GetKuceroImage})
}

func GetKuceroImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "kucero", imageTag)
}

func (renderContext renderContext) KuceroImage() string {
	return GetKuceroImage(kubernetes.AddonVersionForClusterVersion(kubernetes.Kucero, renderContext.config.ClusterVersion).Version)
}

func renderKuceroTemplate(addonConfiguration AddonConfiguration) string {
	return kuceroManifest
}

type kuceroCallbacks struct{}

func (kuceroCallbacks) beforeApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

func (kuceroCallbacks) afterApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

const (
	kuceroManifest = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kucero
  namespace: kube-system
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: kucero
spec:
  allowedHostPaths:
  - pathPrefix: /etc/kubernetes/pki
    readOnly: true
  - pathPrefix: /var/lib/kubelet/pki
    readOnly: true
  fsGroup:
    rule: RunAsAny
  hostPID: true
  privileged: true
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
  - secret
  - hostPath
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kucero
  namespace: kube-system
rules:
- apiGroups:
  - apps
  resourceNames:
  - kucero
  resources:
  - daemonsets
  verbs:
  - update
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - configmaps/status
  verbs:
  - get
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kucero
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - patch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - delete
  - get
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - pods/eviction
  verbs:
  - create
- apiGroups:
  - extensions
  resourceNames:
  - kucero
  resources:
  - podsecuritypolicies
  verbs:
  - use
- apiGroups:
  - certificates.k8s.io
  resourceNames:
  - kubernetes.io/kubelet-serving
  resources:
  - signers
  verbs:
  - approve
  - sign
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests/approval
  verbs:
  - create
  - update
- apiGroups:
  - certificates.k8s.io
  resources:
  - certificatesigningrequests/status
  verbs:
  - patch
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kucero
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kucero
subjects:
- kind: ServiceAccount
  name: kucero
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kucero
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kucero
subjects:
- kind: ServiceAccount
  name: kucero
  namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kucero
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: kucero
  revisionHistoryLimit: 3
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        name: kucero
      annotations:
        {{.AnnotatedVersion}}
    spec:
      serviceAccountName: kucero
      tolerations:
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      nodeSelector:
        node-role.kubernetes.io/master: ""
      hostPID: true
      restartPolicy: Always
      containers:
        - name: kucero
          image: {{.KuceroImage}}
          imagePullPolicy: IfNotPresent
          securityContext:
            privileged: true # Give permission to nsenter /proc/1/ns/mnt
          env:
            # Pass in the name of the node on which this pod is scheduled
            # for use with drain/uncordon operations and lock acquisition
            - name: KUCERO_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          command:
            - /usr/bin/kucero
            - --ca-cert-path=/var/lib/kubelet/pki/kubelet-ca.crt
            - --ca-key-path=/var/lib/kubelet/pki/kubelet-ca.key
          volumeMounts:
            - mountPath: /var/lib/kubelet/pki/kubelet-ca.crt
              name: ca-crt
              readOnly: true
            - mountPath: /var/lib/kubelet/pki/kubelet-ca.key
              name: ca-key
              readOnly: true
          livenessProbe:
            httpGet:
              path: /metrics
              port: 8080
            # The initial delay for the liveness probe is intentionally large to
            # avoid an endless kill & restart cycle if in the event that the initial
            # bootstrapping takes longer than expected.
            initialDelaySeconds: 120
            failureThreshold: 10
            periodSeconds: 60
          readinessProbe:
            httpGet:
              path: /metrics
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 60
      volumes:
        - name: ca-crt
          hostPath:
            path: /var/lib/kubelet/pki/kubelet-ca.crt
            type: File
        - name: ca-key
          hostPath:
            path: /var/lib/kubelet/pki/kubelet-ca.key
            type: File
`
)
