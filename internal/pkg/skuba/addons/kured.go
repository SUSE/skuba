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

import (
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

func init() {
	registerAddon(kubernetes.Kured, renderKuredTemplate, nil, normalPriority, []getImageCallback{GetKuredImage})
}

func GetKuredImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "kured", imageTag)
}

func (renderContext renderContext) KuredImage() string {
	return GetKuredImage(kubernetes.AddonVersionForClusterVersion(kubernetes.Kured, renderContext.config.ClusterVersion).Version)
}

func renderKuredTemplate(addonConfiguration AddonConfiguration) string {
	return kuredManifest
}

const (
	kuredManifest = `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kured
rules:
# Allow kured to read spec.unschedulable
# Allow kubectl to drain/uncordon
#
# NB: These permissions are tightly coupled to the bundled version of kubectl; the ones below
# match https://github.com/kubernetes/kubernetes/blob/v1.12.1/pkg/kubectl/cmd/drain.go
#
- apiGroups: [""]
  resources: ["nodes"]
  verbs:     ["get", "patch"]
- apiGroups: [""]
  resources: ["pods"]
  verbs:     ["list","delete","get"]
- apiGroups: ["apps"]
  resources: ["daemonsets"]
  verbs:     ["get"]
- apiGroups: [""]
  resources: ["pods/eviction"]
  verbs:     ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kured
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kured
subjects:
- kind: ServiceAccount
  name: kured
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: kube-system
  name: kured
rules:
# Allow kured to lock/unlock itself
- apiGroups:     ["apps"]
  resources:     ["daemonsets"]
  resourceNames: ["kured"]
  verbs:         ["update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: kube-system
  name: kured
subjects:
- kind: ServiceAccount
  namespace: kube-system
  name: kured
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kured
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: suse:caasp:psp:kured
roleRef:
  kind: ClusterRole
  name: suse:caasp:psp:privileged
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: kured
  namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kured
  namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kured            # Must match "--ds-name"
  namespace: kube-system # Must match "--ds-namespace"
spec:
  selector:
    matchLabels:
      name: kured
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        name: kured
      annotations:
        {{.AnnotatedVersion}}
    spec:
      serviceAccountName: kured
      tolerations:
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
      hostPID: true # Facilitate entering the host mount namespace via init
      restartPolicy: Always
      containers:
        - name: kured
          image: {{.KuredImage}}
          imagePullPolicy: IfNotPresent
          securityContext:
            privileged: true # Give permission to nsenter /proc/1/ns/mnt
          env:
            # Pass in the name of the node on which this pod is scheduled
            # for use with drain/uncordon operations and lock acquisition
            - name: KURED_NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          command:
            - /usr/bin/kured
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
`
)
