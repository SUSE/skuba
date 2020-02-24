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
	registerAddon(kubernetes.MetricsServer, renderMetricsServerTemplate, metricsServerCallbacks{}, normalPriority, []getImageCallback{GetMetricsServerImage})
}

func GetMetricsServerImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "metrics-server", imageTag)
}

func (renderContext renderContext) MetricsServerImage() string {
	return GetMetricsServerImage(kubernetes.AddonVersionForClusterVersion(kubernetes.MetricsServer, renderContext.config.ClusterVersion).Version)
}

func renderMetricsServerTemplate(addonConfiguration AddonConfiguration) string {
	return metricsServerManifest
}

type metricsServerCallbacks struct{}

func (metricsServerCallbacks) beforeApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

func (metricsServerCallbacks) afterApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

const (
	metricsServerManifest = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:aggregated-metrics-reader
  labels:
    rbac.authorization.k8s.io/aggregate-to-view: "true"
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
    rbac.authorization.k8s.io/aggregate-to-admin: "true"
rules:
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:metrics-server
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  - nodes/stats
  - namespaces
  - configmaps
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: metrics-server:system:auth-delegator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:metrics-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:metrics-server
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: metrics-server-auth-reader
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
- kind: ServiceAccount
  name: metrics-server
  namespace: kube-system
---
apiVersion: v1
kind: Service
metadata:
  name: metrics-server
  namespace: kube-system
  labels:
    app: metrics-server
    kubernetes.io/name: "Metrics-server"
    kubernetes.io/cluster-service: "true"
spec:
  type: ClusterIP
  selector:
    app: metrics-server
  ports:
  - port: 443
    protocol: TCP
    targetPort: https
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: metrics-server
  namespace: kube-system
  labels:
    app: metrics-server
spec:
  revisionHistoryLimit: 3
  replicas: 2
  selector:
    matchLabels:
      app: metrics-server
  template:
    metadata:
      name: metrics-server
      labels:
        app: metrics-server
      annotations:
        {{.AnnotatedVersion}}
    spec:
      serviceAccountName: metrics-server
      containers:
      - name: metrics-server
        image: {{.MetricsServerImage}}
        imagePullPolicy: IfNotPresent
        command:
          - metrics-server
          - --cert-dir=/tmp
          - --logtostderr
          - --secure-port=8443
          - --kubelet-preferred-address-types=InternalIP,InternalDNS,Hostname,ExternalIP,ExternalDNS
          - --kubelet-certificate-authority=/var/lib/kubelet/pki/kubelet-ca.crt
        ports:
        - containerPort: 8443
          name: https
        livenessProbe:
          httpGet:
            path: /healthz
            port: https
            scheme: HTTPS
          initialDelaySeconds: 20
        readinessProbe:
          httpGet:
            path: /healthz
            port: https
            scheme: HTTPS
          initialDelaySeconds: 20
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - all
          readOnlyRootFilesystem: true
          runAsGroup: 10001
          runAsNonRoot: true
          runAsUser: 10001
        volumeMounts:
        - name: tmp
          mountPath: /tmp
        - mountPath: /var/lib/kubelet/pki/kubelet-ca.crt
          name: pki
          readOnly: true
      # mount in tmp so we can safely use from-scratch images and/or read-only containers
      volumes:
      - name: tmp
        emptyDir: {}
      - hostPath:
          path: /var/lib/kubelet/pki/kubelet-ca.crt
          type: File
        name: pki
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
---
apiVersion: apiregistration.k8s.io/v1beta1
kind: APIService
metadata:
  name: v1beta1.metrics.k8s.io
spec:
  service:
    name: metrics-server
    namespace: kube-system
  group: metrics.k8s.io
  version: v1beta1
  insecureSkipTLSVerify: true
  groupPriorityMinimum: 100
  versionPriority: 100
`
)
