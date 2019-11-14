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
	registerAddon(kubernetes.HelmTiller, renderHelmTillerTemplate, nil, normalPriority, []getImageCallback{GetHelmTillerImage})
}

func GetHelmTillerImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "helm-tiller", imageTag)
}

func (renderContext renderContext) HelmTillerImage() string {
	return GetHelmTillerImage(kubernetes.AddonVersionForClusterVersion(kubernetes.HelmTiller, renderContext.config.ClusterVersion).Version)
}

func renderHelmTillerTemplate(addonConfiguration AddonConfiguration) string {
	return helmTillerManifest
}

const (
	helmTillerManifest = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiller
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: tiller
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tiller
  namespace: kube-system
  labels:
    app: helm
    name: tiller
spec:
  replicas: 1
  selector:
    matchLabels:
      app: helm
      name: tiller
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: helm
        name: tiller
    spec:
      automountServiceAccountToken: true
      containers:
      - env:
        - name: TILLER_NAMESPACE
          value: kube-system
        - name: TILLER_HISTORY_MAX
          value: "0"
        image: {{.HelmTillerImage}}
        imagePullPolicy: IfNotPresent
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /liveness
            port: 44135
            scheme: HTTP
          initialDelaySeconds: 1
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: tiller
        ports:
        - containerPort: 44134
          name: tiller
          protocol: TCP
        - containerPort: 44135
          name: http
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /readiness
            port: 44135
            scheme: HTTP
          initialDelaySeconds: 1
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      serviceAccount: tiller
      serviceAccountName: tiller
      terminationGracePeriodSeconds: 30
`
)
