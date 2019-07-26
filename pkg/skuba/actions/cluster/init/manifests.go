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
CRIO_OPTIONS=--pause-image=registry.suse.de/devel/caasp/4.0/containers/containers/caasp/v4/pause:3.1 --default-capabilities="CHOWN,DAC_OVERRIDE,FSETID,FOWNER,NET_RAW,SETGID,SETUID,SETPCAP,NET_BIND_SERVICE,SYS_CHROOT,KILL,MKNOD,AUDIT_WRITE,SETFCAP"
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

	pspPrivManifest = `---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: suse.caasp.psp.privileged
  annotations:
    apparmor.security.beta.kubernetes.io/allowedProfileName: '*'
    apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
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
    # SELinux is unsed in CaaSP
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
`

	pspUnprivManifest = `---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: suse.caasp.psp.unprivileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: runtime/default
    seccomp.security.alpha.kubernetes.io/defaultProfileName: runtime/default
    apparmor.security.beta.kubernetes.io/allowedProfileName: runtime/default
    apparmor.security.beta.kubernetes.io/defaultProfileName: runtime/default
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

	ciliumManifest = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cilium
  namespace: kube-system
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: cilium
  namespace: kube-system
spec:
  updateStrategy:
    type: "RollingUpdate"
    rollingUpdate:
      # Specifies the maximum number of Pods that can be unavailable during the update process.
      # The current default value is 1 or 100% for daemonsets; Adding an explicit value here
      # to avoid confusion, as the default value is specific to the type (daemonset/deployment).
      maxUnavailable: "100%"
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
      - name: install-cni-conf
        image: {{.CiliumImage}}
        command:
          - /bin/sh
          - "-c"
          - "cp -f /etc/cni/net.d/10-cilium-cni.conf /host/etc/cni/net.d/10-cilium-cni.conf"
        volumeMounts:
        - name: host-cni-conf
          mountPath: /host/etc/cni/net.d
      - name: install-cni-bin
        image: {{.CiliumImage}}
        command:
          - /bin/sh
          - "-c"
          - "cp -f /usr/lib/cni/* /host/opt/cni/bin/"
        volumeMounts:
        - name: host-cni-bin
          mountPath: /host/opt/cni/bin/
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
          - "--disable-ipv4=$(DISABLE_IPV4)"
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
          - name: "DISABLE_IPV4"
            valueFrom:
              configMapKeyRef:
                name: cilium-config
                key: disable-ipv4
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
- apiGroups: ["extensions","apps"]
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
- apiGroups:     ["extensions"]
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
`
	dexManifest = `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: oidc-dex-config
  namespace: kube-system
data:
  config.yaml: |
    issuer: https://{{.ControlPlane}}:32000

    storage:
      type: kubernetes
      config:
        inCluster: true

    web:
      https: 0.0.0.0:32000
      tlsCert: /etc/dex/pki/tls.crt
      tlsKey: /etc/dex/pki/tls.key
      tlsClientCA: /etc/dex/pki/ca.crt

    frontend:
      theme: "caasp"
      dir: /usr/share/caasp-dex/web

    # This is a sample with LDAP as connector.
    # Requires a update to fulfill your environment.
    connectors:
    - type: ldap
      id: ldap
      name: openLDAP
      config:
        host: openldap.kube-system.svc.cluster.local:389
        insecureNoSSL: true
        insecureSkipVerify: true
        bindDN: cn=admin,dc=example,dc=com
        bindPW: admin
        usernamePrompt: Email Address
        userSearch:
          baseDN: cn=Users,dc=example,dc=com
          filter: "(objectClass=person)"
          username: mail
          idAttr: DN
          emailAttr: mail
          nameAttr: cn
        groupSearch:
          baseDN: cn=Groups,dc=example,dc=com
          filter: "(objectClass=group)"
          userAttr: DN
          groupAttr: member
          nameAttr: cn

    staticClients:
    - id: oidc
      redirectURIs:
      - 'https://{{.ControlPlane}}:32001/callback'
      name: 'OIDC'
      secret: {{.GangwayClientSecret}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oidc-dex
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: oidc-dex
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      name: oidc-dex
      labels:
        app: oidc-dex
    spec:
      serviceAccountName: oidc-dex
      containers:
      - name: oidc-dex
        image: {{.DexImage}}
        imagePullPolicy: IfNotPresent
        command:
          - /usr/bin/caasp-dex
          - serve
          - /etc/dex/cfg/config.yaml
        ports:
          - name: https
            containerPort: 32000
        volumeMounts:
          - name: dex-config-path
            mountPath: /etc/dex/cfg
          - name: dex-cert-path
            mountPath: /etc/dex/pki
      volumes:
      - name: dex-config-path
        configMap:
          name: oidc-dex-config
          items:
          - key: config.yaml
            path: config.yaml
      - name: dex-cert-path
        secret:
          secretName: oidc-dex-cert
---
apiVersion: v1
kind: Service
metadata:
  name: oidc-dex
  namespace: kube-system
spec:
  selector:
    app: oidc-dex
  type: NodePort
  ports:
  - name: https
    port: 32000
    targetPort: 32000
    nodePort: 32000
    protocol: TCP
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: oidc-dex
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: oidc-dex
  namespace: kube-system
rules:
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["create", "get", "list", "update", "watch"]
- apiGroups: ["dex.coreos.com"]
  resources: ["oauth2clients", "connectors", "passwords", "refreshtokens"]
  verbs: ["list"]
- apiGroups: ["dex.coreos.com"]
  resources: ["signingkeies"]
  verbs: ["create", "get", "list", "update"]
- apiGroups: ["dex.coreos.com"]
  resources: ["authcodes", "authrequests", "offlinesessionses"]
  verbs: ["create", "delete", "get", "list", "update"]
- apiGroups: ["dex.coreos.com"]
  resources: ["refreshtokens"]
  verbs: ["create", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: oidc-dex
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: oidc-dex
  apiGroup: rbac.authorization.k8s.io
subjects:
- kind: ServiceAccount
  name: oidc-dex
  namespace: kube-system
`
	gangwayManifest = `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: oidc-gangway-config
  namespace: kube-system
data:
  gangway.yaml: |
    clusterName: "skuba"

    redirectURL: "https://{{.ControlPlane}}:32001/callback"
    scopes: ["openid", "email", "groups", "profile", "offline_access"]

    serveTLS: true
    authorizeURL: "https://{{.ControlPlane}}:32000/auth"
    tokenURL: "https://{{.ControlPlane}}:32000/token"
    keyFile: /etc/gangway/pki/tls.key
    certFile: /etc/gangway/pki/tls.crt

    clientID: "oidc"
    clientSecret: "{{.GangwayClientSecret}}"
    usernameClaim: "email"
    apiServerURL: "https://{{.ControlPlane}}:6443"
    cluster_ca_path: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
    trustedCAPath: /etc/gangway/pki/ca.crt
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oidc-gangway
  namespace: kube-system
  labels:
    app: oidc-gangway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: oidc-gangway
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: oidc-gangway
    spec:
      serviceAccountName: oidc-gangway
      containers:
        - name: oidc-gangway
          image: {{.GangwayImage}}
          imagePullPolicy: IfNotPresent
          command: ["gangway", "-config", "/gangway/gangway.yaml"]
          env:
            - name: GANGWAY_SESSION_SECURITY_KEY
              valueFrom:
                secretKeyRef:
                  name: oidc-gangway-secret
                  key: session-key
          ports:
            - name: web
              containerPort: 8080
              protocol: TCP
          volumeMounts:
            - name: gangway-config-path
              mountPath: /gangway/
              readOnly: true
            - name: gangway-cert-path
              mountPath: /etc/gangway/pki
              readOnly: true
      volumes:
        - name: gangway-config-path
          configMap:
            name: oidc-gangway-config
            items:
              - key: gangway.yaml
                path: gangway.yaml
        - name: gangway-cert-path
          secret:
            secretName: oidc-gangway-cert
---
apiVersion: v1
kind: Service
metadata:
  name: oidc-gangway
  namespace: kube-system
spec:
  type: NodePort
  ports:
  - name: web
    port: 8080
    protocol: TCP
    targetPort: 8080
    nodePort: 32001
  selector:
    app: oidc-gangway
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: oidc-gangway
  namespace: kube-system
`
)
