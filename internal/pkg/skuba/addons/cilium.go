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
	"fmt"
	"strings"

	"github.com/SUSE/skuba/internal/pkg/skuba/cni"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

func init() {
	registerAddon(kubernetes.Cilium, renderCiliumTemplate, renderCiliumPreflightTemplate, ciliumCallbacks{}, normalPriority, []getImageCallback{GetCiliumInitImage, GetCiliumOperatorImage, GetCiliumImage})
}

func GetCiliumInitImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "cilium-init", imageTag)
}

func GetCiliumOperatorImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "cilium-operator", imageTag)
}

func GetCiliumImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "cilium", imageTag)
}

func (renderContext renderContext) CiliumInitImage() string {
	return GetCiliumInitImage(kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, renderContext.config.ClusterVersion).Version)
}

func (renderContext renderContext) CiliumOperatorImage() string {
	return GetCiliumOperatorImage(kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, renderContext.config.ClusterVersion).Version)
}

func (renderContext renderContext) CiliumImage() string {
	return GetCiliumImage(kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, renderContext.config.ClusterVersion).Version)
}

func renderCiliumTemplate(addonConfiguration AddonConfiguration) string {
	ciliumVersion := kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, addonConfiguration.ClusterVersion).Version
	switch {
	case strings.HasPrefix(ciliumVersion, "1.5"):
		return ciliumManifestv15
	case strings.HasPrefix(ciliumVersion, "1.6"):
		return ciliumManifestv16
	}
	panic(fmt.Sprintf("invalid cilium addon version: %s", ciliumVersion))
}

func renderCiliumPreflightTemplate(addonConfiguration AddonConfiguration) string {
	ciliumVersion := kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, addonConfiguration.ClusterVersion).Version
	switch {
	case strings.HasPrefix(ciliumVersion, "1.6"):
		return ciliumPreflightManifestv16
	}
	return ""
}

type ciliumCallbacks struct{}

func (ciliumCallbacks) beforeApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	ciliumVersion := kubernetes.AddonVersionForClusterVersion(kubernetes.Cilium, addonConfiguration.ClusterVersion).Version
	client, config, err := kubernetes.GetAdminClientSetWithConfig()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}
	if err := cni.CreateCiliumSecret(client, ciliumVersion); err != nil {
		return err
	}
	if err := cni.CreateOrUpdateCiliumConfigMap(client, ciliumVersion); err != nil {
		return err
	}

	// Handle migration from etcd to CRD when upgrading from Cilium 1.5.
	needsMigration, err := cni.NeedsEtcdToCrdMigration(client, ciliumVersion)
	if err != nil {
		return err
	}
	if !needsMigration {
		return nil
	}
	if err := cni.MigrateEtcdToCrd(client, config); err != nil {
		return err
	}
	if err := cni.RemoveEtcdConfig(client); err != nil {
		return err
	}

	return nil
}

func (ciliumCallbacks) afterApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

const (
	ciliumManifestv15 = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: suse:caasp:psp:cilium
roleRef:
  kind: ClusterRole
  name: suse:caasp:psp:privileged
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: cilium
  namespace: kube-system
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: cilium
  namespace: kube-system
spec:
  revisionHistoryLimit: 3
  updateStrategy:
    type: "RollingUpdate"
    rollingUpdate:
      # Specifies the maximum number of Pods that can be unavailable during the update process.
      maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: cilium
      kubernetes.io/cluster-service: "true"
  template:
    metadata:
      labels:
        k8s-app: cilium
        kubernetes.io/cluster-service: "true"
      annotations:
        {{.AnnotatedVersion}}
        # This annotation plus the CriticalAddonsOnly toleration makes
        # cilium to be a critical pod in the cluster, which ensures cilium
        # gets priority scheduling.
        # https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
        scheduler.alpha.kubernetes.io/critical-pod: ''
        scheduler.alpha.kubernetes.io/tolerations: >-
          [{"key":"dedicated","operator":"Equal","value":"master","effect":"NoSchedule"}]
    spec:
      serviceAccountName: cilium
      initContainers:
      # TODO(mrostecki): Using /usr/lib/cni is not recommended anymore as
      # every project (cri-o, Cilium, other CNI plugins) are moving away to
      # /opt/cni/bin. The /usr/lib/cni dir is left there only to support old
      # Cilium 1.5 + cri-o < 1.16.3 for backwards compatibility. As soon as
      # cri-o < 1.16.3 disappears from CaaSP, /usr/lib/cni can be removed.
      - name: install-cni-bin
        image: {{.CiliumImage}}
        command:
          - /bin/sh
          - "-c"
          - "cp -f /usr/lib/cni/* /host/opt/cni/bin/"
          - "cp -f /usr/lib/cni/* /host/usr/lib/cni/"
        volumeMounts:
        - name: host-cni-bin
          mountPath: /host/opt/cni/bin/
        - name: host-usr-cni-bin
          mountPath: /host/usr/lib/cni/
      - name: install-cni-conf
        image: {{.CiliumImage}}
        command:
          - /bin/sh
          - "-c"
          - "cp -f /etc/cni/net.d/10-cilium-cni.conf /host/etc/cni/net.d/10-cilium-cni.conf"
        volumeMounts:
        - name: host-cni-conf
          mountPath: /host/etc/cni/net.d
      - name: clean-cilium-state
        image: {{.CiliumInitImage}}
        imagePullPolicy: IfNotPresent
        command:
        - cilium-init
        env:
        - name: CLEAN_CILIUM_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-state
              name: cilium-config
              optional: true
        - name: CLEAN_CILIUM_BPF_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-bpf-state
              name: cilium-config
              optional: true
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          privileged: true
        volumeMounts:
        - mountPath: /var/run/cilium
          name: cilium-run
      containers:
      - image: {{.CiliumImage}}
        imagePullPolicy: IfNotPresent
        name: cilium-agent
        command: [ "cilium-agent" ]
        args:
          - "--debug=$(CILIUM_DEBUG)"
          - "--disable-envoy-version-check"
          - "-t=vxlan"
          - "--kvstore=etcd"
          - "--kvstore-opt=etcd.config=/var/lib/etcd-config/etcd.config"
          - "--enable-ipv4=$(ENABLE_IPV4)"
          - "--enable-ipv6=$(ENABLE_IPV6)"
          - "--container-runtime=crio"
        ports:
          - name: prometheus
            containerPort: 9090
        lifecycle:
          preStop:
            exec:
              command:
                - "rm -f /host/etc/cni/net.d/10-cilium-cni.conf /host/opt/cni/bin/cilium-cni"
        env:
          - name: "K8S_NODE_NAME"
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: "CILIUM_DEBUG"
            valueFrom:
              configMapKeyRef:
                name: cilium-config
                key: debug
          - name: "ENABLE_IPV4"
            valueFrom:
              configMapKeyRef:
                name: cilium-config
                key: enable-ipv4
          - name: "ENABLE_IPV6"
            valueFrom:
              configMapKeyRef:
                name: cilium-config
                key: enable-ipv6
        livenessProbe:
          exec:
            command:
            - cilium
            - status
          # The initial delay for the liveness probe is intentionally large to
          # avoid an endless kill & restart cycle if in the event that the initial
          # bootstrapping takes longer than expected.
          initialDelaySeconds: 120
          failureThreshold: 10
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - cilium
            - status
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
          - name: cilium-run
            mountPath: /var/run/cilium
          - name: host-cni-bin
            mountPath: /host/opt/cni/bin/
          - name: host-cni-conf
            mountPath: /host/etc/cni/net.d
          - name: container-socket
            mountPath: /var/run/crio/crio.sock
            readOnly: true
          - name: etcd-config-path
            mountPath: /var/lib/etcd-config
            readOnly: true
          - name: cilium-etcd-secret-mount
            mountPath: /tmp/cilium-etcd
          - name: lib-modules
            mountPath: /lib/modules
            readOnly: true
        securityContext:
          capabilities:
            add:
              - "NET_ADMIN"
              - "SYS_MODULE"
          privileged: true
      hostNetwork: true
      volumes:
          # To keep state between restarts / upgrades
        - name: cilium-run
          hostPath:
            path: /var/run/cilium
          # To keep state between restarts / upgrades
          # To read crio events from the node
        - name: container-socket
          hostPath:
            path: /var/run/crio/crio.sock
          # To install cilium cni plugin in the host
        - name: host-cni-bin
          hostPath:
            path: /opt/cni/bin
          # To install cilium cni plugin in the host
        - name: host-usr-cni-bin
          hostPath:
            path: /usr/lib/cni
          # To install cilium cni configuration in the host
        - name: host-cni-conf
          hostPath:
            path: /etc/cni/net.d
          # To be able to load kernel modules
        - name: lib-modules
          hostPath:
            path: /lib/modules
          # To read the etcd config stored in config maps
        - name: etcd-config-path
          configMap:
            name: cilium-config
            items:
            - key: etcd-config
              path: etcd.config
        - name: cilium-etcd-secret-mount
          secret:
            secretName: cilium-secret
      restartPolicy: Always
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      - effect: NoSchedule
        key: node.cloudprovider.kubernetes.io/uninitialized
        value: "true"
      # Mark cilium's pod as critical for rescheduling
      - key: CriticalAddonsOnly
        operator: "Exists"
      - key: node.kubernetes.io/not-ready
        operator: Exists
        effect: NoSchedule
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.cilium/app: operator
    name: cilium-operator
  name: cilium-operator
  namespace: kube-system
spec:
  revisionHistoryLimit: 3
  replicas: 1
  selector:
    matchLabels:
      io.cilium/app: operator
      name: cilium-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        io.cilium/app: operator
        name: cilium-operator
    spec:
      containers:
      - args:
        - --debug=$(CILIUM_DEBUG)
        - --kvstore=etcd
        - --kvstore-opt=etcd.config=/var/lib/etcd-config/etcd.config
        command:
        - cilium-operator
        env:
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_DEBUG
          valueFrom:
            configMapKeyRef:
              key: debug
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_NAME
          valueFrom:
            configMapKeyRef:
              key: cluster-name
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              key: cluster-id
              name: cilium-config
              optional: true
        - name: CILIUM_DISABLE_ENDPOINT_CRD
          valueFrom:
            configMapKeyRef:
              key: disable-endpoint-crd
              name: cilium-config
              optional: true
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              key: AWS_ACCESS_KEY_ID
              name: cilium-aws
              optional: true
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              key: AWS_SECRET_ACCESS_KEY
              name: cilium-aws
              optional: true
        - name: AWS_DEFAULT_REGION
          valueFrom:
            secretKeyRef:
              key: AWS_DEFAULT_REGION
              name: cilium-aws
              optional: true
        image: {{.CiliumOperatorImage}}
        imagePullPolicy: Always
        name: cilium-operator
        volumeMounts:
        - mountPath: /var/lib/etcd-config
          name: etcd-config-path
          readOnly: true
        - mountPath: /tmp/cilium-etcd
          name: etcd-secrets
          readOnly: true
      dnsPolicy: ClusterFirst
      priorityClassName: system-node-critical
      restartPolicy: Always
      serviceAccount: cilium-operator
      serviceAccountName: cilium-operator
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      volumes:
      # To read the etcd config stored in config maps
      - configMap:
          defaultMode: 420
          items:
          - key: etcd-config
            path: etcd.config
          name: cilium-config
        name: etcd-config-path
        # To read the k8s etcd secrets in case the user might want to use TLS
      - name: etcd-secrets
        secret:
          secretName: cilium-secret
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium-operator
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium-operator
rules:
- apiGroups:
  - ""
  resources:
  # to get k8s version and status
  - componentstatuses
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  # to automatically delete [core|kube]dns pods so that are starting to being
  # managed by Cilium
  - pods
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - ""
  resources:
  # to automatically read from k8s and import the node's pod CIDR to cilium's
  # etcd so all nodes know how to reach another pod running in in a different
  # node.
  - nodes
  # to perform the translation of a CNP that contains ToGroup to its endpoints
  - services
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium-operator
subjects:
- kind: ServiceAccount
  name: cilium-operator
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cilium
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium
subjects:
- kind: ServiceAccount
  name: cilium
  namespace: kube-system
- kind: Group
  name: system:nodes
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cilium
rules:
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  - services
  - nodes
  - endpoints
  - componentstatuses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - get
  - list
  - watch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  verbs:
  - '*'
`

	ciliumManifestv16 = `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: suse:caasp:psp:cilium
roleRef:
  kind: ClusterRole
  name: suse:caasp:psp:privileged
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: cilium
  namespace: kube-system
---
# Source: cilium/charts/agent/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium
  namespace: kube-system
---
# Source: cilium/charts/operator/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium-operator
  namespace: kube-system
---
# Source: cilium/charts/agent/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium
rules:
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  - services
  - nodes
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - get
  - list
  - watch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
# Source: cilium/charts/operator/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium-operator
rules:
- apiGroups:
  - ""
  resources:
  # to automatically delete [core|kube]dns pods so that are starting to being
  # managed by Cilium
  - pods
  verbs:
  - get
  - list
  - watch
  - delete
- apiGroups:
  - ""
  resources:
  # to automatically read from k8s and import the node's pod CIDR to cilium's
  # etcd so all nodes know how to reach another pod running in in a different
  # node.
  - nodes
  # to perform the translation of a CNP that contains ToGroup to its endpoints
  - services
  - endpoints
  # to check apiserver connectivity
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
# Source: cilium/charts/agent/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium
subjects:
- kind: ServiceAccount
  name: cilium
  namespace: kube-system
---
# Source: cilium/charts/operator/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium-operator
subjects:
- kind: ServiceAccount
  name: cilium-operator
  namespace: kube-system
---
# Source: cilium/charts/agent/templates/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: cilium
    kubernetes.io/cluster-service: "true"
  name: cilium
  namespace: kube-system
spec:
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      k8s-app: cilium
      kubernetes.io/cluster-service: "true"
  template:
    metadata:
      annotations:
        {{.AnnotatedVersion}}
        # This annotation plus the CriticalAddonsOnly toleration makes
        # cilium to be a critical pod in the cluster, which ensures cilium
        # gets priority scheduling.
        # https://kubernetes.io/docs/tasks/administer-cluster/guaranteed-scheduling-critical-addon-pods/
        scheduler.alpha.kubernetes.io/critical-pod: ""
        scheduler.alpha.kubernetes.io/tolerations: '[{"key":"dedicated","operator":"Equal","value":"master","effect":"NoSchedule"}]'
      labels:
        k8s-app: cilium
        kubernetes.io/cluster-service: "true"
    spec:
      containers:
      - args:
        - --config-dir=/tmp/cilium/config-map
        command:
        - cilium-agent
        env:
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_K8S_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: CILIUM_FLANNEL_MASTER_DEVICE
          valueFrom:
            configMapKeyRef:
              key: flannel-master-device
              name: cilium-config
              optional: true
        - name: CILIUM_FLANNEL_UNINSTALL_ON_EXIT
          valueFrom:
            configMapKeyRef:
              key: flannel-uninstall-on-exit
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTERMESH_CONFIG
          value: /var/lib/cilium/clustermesh/
        - name: CILIUM_CNI_CHAINING_MODE
          valueFrom:
            configMapKeyRef:
              key: cni-chaining-mode
              name: cilium-config
              optional: true
        - name: CILIUM_CUSTOM_CNI_CONF
          valueFrom:
            configMapKeyRef:
              key: custom-cni-conf
              name: cilium-config
              optional: true
        image: {{.CiliumImage}}
        imagePullPolicy: IfNotPresent
        lifecycle:
          postStart:
            exec:
              command:
              - /bin/sh
              - "-c"
              - "cp -f /usr/lib/cni/cilium-cni /host/opt/cni/bin/ && cp -f /etc/cni/net.d/10-cilium-cni.conf /host/etc/cni/net.d/"
          preStop:
            exec:
              command:
              - cilium-cni-uninstall
        livenessProbe:
          exec:
            command:
            - cilium
            - status
            - --brief
          failureThreshold: 10
          # The initial delay for the liveness probe is intentionally large to
          # avoid an endless kill & restart cycle if in the event that the initial
          # bootstrapping takes longer than expected.
          initialDelaySeconds: 120
          periodSeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
        name: cilium-agent
        readinessProbe:
          exec:
            command:
            - cilium
            - status
            - --brief
          failureThreshold: 3
          initialDelaySeconds: 5
          periodSeconds: 30
          successThreshold: 1
          timeoutSeconds: 5
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
          privileged: true
        volumeMounts:
        - mountPath: /var/run/cilium
          name: cilium-run
        - mountPath: /host/opt/cni/bin
          name: cni-path
        - mountPath: /host/etc/cni/net.d
          name: etc-cni-netd
        - mountPath: /var/run/crio/crio.sock
          name: crio-socket
        - mountPath: /tmp/cilium/config-map
          name: cilium-config-path
          readOnly: true
          # Needed to be able to load kernel modules
        - mountPath: /lib/modules
          name: lib-modules
          readOnly: true
        - mountPath: /run/xtables.lock
          name: xtables-lock
      hostNetwork: true
      initContainers:
      - command:
        - cilium-init
        env:
        - name: CILIUM_ALL_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-state
              name: cilium-config
              optional: true
        - name: CILIUM_BPF_STATE
          valueFrom:
            configMapKeyRef:
              key: clean-cilium-bpf-state
              name: cilium-config
              optional: true
        - name: CILIUM_WAIT_BPF_MOUNT
          valueFrom:
            configMapKeyRef:
              key: wait-bpf-mount
              name: cilium-config
              optional: true
        image: {{.CiliumImage}}
        imagePullPolicy: IfNotPresent
        name: clean-cilium-state
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
          privileged: true
        volumeMounts:
        - mountPath: /var/run/cilium
          name: cilium-run
      restartPolicy: Always
      priorityClassName: system-node-critical
      serviceAccount: cilium
      serviceAccountName: cilium
      terminationGracePeriodSeconds: 1
      tolerations:
      - operator: Exists
      volumes:
        # To keep state between restarts / upgrades
      - hostPath:
          path: /var/run/cilium
          type: DirectoryOrCreate
        name: cilium-run
      # To read container runtime events from the node
      - hostPath:
          path: /var/run/crio/crio.sock
          type: Socket
        name: crio-socket
      # To install cilium cni plugin in the host
      - hostPath:
          path:  /opt/cni/bin
          type: DirectoryOrCreate
        name: cni-path
        # To install cilium cni configuration in the host
      - hostPath:
          path: /etc/cni/net.d
          type: DirectoryOrCreate
        name: etc-cni-netd
        # To be able to load kernel modules
      - hostPath:
          path: /lib/modules
        name: lib-modules
        # To access iptables concurrently with other processes (e.g. kube-proxy)
      - hostPath:
          path: /run/xtables.lock
          type: FileOrCreate
        name: xtables-lock
        # To read the configuration from the config map
      - configMap:
          name: cilium-config
        name: cilium-config-path
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
---
# Source: cilium/charts/operator/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    io.cilium/app: operator
    name: cilium-operator
  name: cilium-operator
  namespace: kube-system
spec:
  revisionHistoryLimit: 3
  replicas: 1
  selector:
    matchLabels:
      io.cilium/app: operator
      name: cilium-operator
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      annotations:
      labels:
        io.cilium/app: operator
        name: cilium-operator
    spec:
      containers:
      - args:
        - --debug=$(CILIUM_DEBUG)
        - --identity-allocation-mode=$(CILIUM_IDENTITY_ALLOCATION_MODE)
        command:
        - cilium-operator
        env:
        - name: CILIUM_K8S_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        - name: CILIUM_DEBUG
          valueFrom:
            configMapKeyRef:
              key: debug
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_NAME
          valueFrom:
            configMapKeyRef:
              key: cluster-name
              name: cilium-config
              optional: true
        - name: CILIUM_CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              key: cluster-id
              name: cilium-config
              optional: true
        - name: CILIUM_IPAM
          valueFrom:
            configMapKeyRef:
              key: ipam
              name: cilium-config
              optional: true
        - name: CILIUM_DISABLE_ENDPOINT_CRD
          valueFrom:
            configMapKeyRef:
              key: disable-endpoint-crd
              name: cilium-config
              optional: true
        - name: CILIUM_KVSTORE
          valueFrom:
            configMapKeyRef:
              key: kvstore
              name: cilium-config
              optional: true
        - name: CILIUM_KVSTORE_OPT
          valueFrom:
            configMapKeyRef:
              key: kvstore-opt
              name: cilium-config
              optional: true
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              key: AWS_ACCESS_KEY_ID
              name: cilium-aws
              optional: true
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              key: AWS_SECRET_ACCESS_KEY
              name: cilium-aws
              optional: true
        - name: AWS_DEFAULT_REGION
          valueFrom:
            secretKeyRef:
              key: AWS_DEFAULT_REGION
              name: cilium-aws
              optional: true
        - name: CILIUM_IDENTITY_ALLOCATION_MODE
          valueFrom:
            configMapKeyRef:
              key: identity-allocation-mode
              name: cilium-config
              optional: true
        image: {{.CiliumOperatorImage}}
        imagePullPolicy: IfNotPresent
        name: cilium-operator
        livenessProbe:
          httpGet:
            path: /healthz
            port: 9234
            scheme: HTTP
          initialDelaySeconds: 60
          periodSeconds: 10
          timeoutSeconds: 3
      hostNetwork: true
      restartPolicy: Always
      serviceAccount: cilium-operator
      serviceAccountName: cilium-operator
`

	ciliumPreflightManifestv16 = `---
# Source: cilium/charts/preflight/templates/cilium-preflight-clusterrole-binding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cilium-preflight-check
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cilium-preflight-check
subjects:
- kind: ServiceAccount
  name: cilium-preflight-check
  namespace: kube-system
---
# Source: cilium/charts/preflight/templates/cilium-preflight-serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium-preflight-check
  namespace: kube-system
---
# Source: cilium/charts/agent/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cilium-preflight-check
rules:
- apiGroups:
  - networking.k8s.io
  resources:
  - networkpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - namespaces
  - services
  - nodes
  - endpoints
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - nodes
  verbs:
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - get
  - list
  - watch
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - cilium.io
  resources:
  - ciliumnetworkpolicies
  - ciliumnetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
# Source: cilium/charts/preflight/templates/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: cilium-pre-flight-check
  namespace: kube-system
spec:
  selector:
    matchLabels:
      k8s-app: cilium-pre-flight-check
      kubernetes.io/cluster-service: "true"
  template:
    metadata:
      labels:
        k8s-app: cilium-pre-flight-check
        kubernetes.io/cluster-service: "true"
    spec:
      affinity:
        podAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: "k8s-app"
                operator: In
                values:
                - cilium
            topologyKey: "kubernetes.io/hostname"
      initContainers:
        - name: clean-cilium-state
          image: {{.CiliumImage}}
          imagePullPolicy: IfNotPresent
          command: ["/bin/echo"]
          args:
          - "hello"
      containers:
        - image: {{.CiliumImage}}
          imagePullPolicy: IfNotPresent
          name: cilium-pre-flight-check
          command: ["/bin/sh"]
          args:
          - -c
          - "touch /tmp/ready; sleep 1h"
          livenessProbe:
            exec:
              command:
              - cat
              - /tmp/ready
            initialDelaySeconds: 5
            periodSeconds: 5
          readinessProbe:
            exec:
              command:
              - cat
              - /tmp/ready
            initialDelaySeconds: 5
            periodSeconds: 5
          volumeMounts:
          - mountPath: /var/run/cilium
            name: cilium-run
          - mountPath: /var/lib/etcd-config
            name: etcd-config-path
            readOnly: true
          - mountPath: /var/lib/etcd-secrets
            name: etcd-secrets
            readOnly: true
      hostNetwork: true
      # This is here to seamlessly allow migrate-identity to work with
      # etcd-operator setups. The assumption is that other cases would also
      # work since the cluster DNS would forward the request on.
      # This differs from the cilium-agent daemonset, where this is only
      # enabled when global.etcd.managed=true
      dnsPolicy: ClusterFirstWithHostNet
      restartPolicy: Always
      serviceAccount: cilium-preflight-check
      serviceAccountName: cilium-preflight-check
      tolerations:
        - effect: NoSchedule
          key: node.kubernetes.io/not-ready
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
        - effect: NoSchedule
          key: node.cloudprovider.kubernetes.io/uninitialized
          value: "true"
        - key: CriticalAddonsOnly
          operator: "Exists"
      volumes:
        # To keep state between restarts / upgrades
      - hostPath:
          path: /var/run/cilium
          type: DirectoryOrCreate
        name: cilium-run
      - hostPath:
          path: /sys/fs/bpf
          type: DirectoryOrCreate
        name: bpf-maps
        # To read the etcd config stored in config maps
      - configMap:
          defaultMode: 420
          items:
          - key: etcd-config
            path: etcd.config
          name: cilium-config
        name: etcd-config-path
        # To read the k8s etcd secrets in case the user might want to use TLS
      - name: etcd-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-secret
`
)
