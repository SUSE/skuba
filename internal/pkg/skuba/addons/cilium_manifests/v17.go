/*
 * Copyright (c) 2020 SUSE LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package cilium_manifests

const (
	Manifestv17 = `---
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
# Source: cilium/install/kubernetes/cilium/charts/agent/templates/clusterrole.yaml
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
  - discovery.k8s.io
  resources:
  - endpointslices
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
  - ciliumclusterwidenetworkpolicies
  - ciliumclusterwidenetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
---
# Source: install/kubernetes/cilium/charts/operator/templates/clusterrole.yaml
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
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  # to automatically read from k8s and import the node's pod CIDR to cilium's
  # etcd so all nodes know how to reach another pod running in in a different
  # node.
  - nodes
  # to perform the translation of a CNP that contains 'ToGroup' to its endpoints
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
  - ciliumclusterwidenetworkpolicies
  - ciliumclusterwidenetworkpolicies/status
  - ciliumendpoints
  - ciliumendpoints/status
  - ciliumnodes
  - ciliumnodes/status
  - ciliumidentities
  - ciliumidentities/status
  verbs:
  - '*'
# For cilium-operator running in HA mode.
#
# Cilium operator running in HA mode requires the use of ResourceLock for Leader Election
# between mulitple running instances.
# The preferred way of doing this is to use LeasesResourceLock as edits to Leases are less
# common and fewer objects in the cluster watch "all Leases".
# The support for leases was introduced in coordination.k8s.io/v1 during Kubernetes 1.14 release.
# In Cilium we currently don't support HA mode for K8s version < 1.14. This condition make sure
# that we only authorize access to leases resources in supported K8s versions.
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - create
  - get
  - update
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
            host: '127.0.0.1'
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
)
