/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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

package kubernetes

const (
	OpenstackCloudControllerManagerSecretName = "openstack-cloud-config"

	OpenstackCloudControllerManagerTemplate = `---
 apiVersion: v1
 kind: ServiceAccount
 metadata:
   name: cloud-controller-manager
   namespace: kube-system
 ---
 kind: ClusterRoleBinding
 apiVersion: rbac.authorization.k8s.io/v1
 metadata:
   name: system:cloud-controller-manager
 roleRef:
   apiGroup: rbac.authorization.k8s.io
   kind: ClusterRole
   name: cluster-admin
 subjects:
 - kind: ServiceAccount
   name: cloud-controller-manager
   namespace: kube-system
 ---
 kind: ClusterRoleBinding
 apiVersion: rbac.authorization.k8s.io/v1
 metadata:
   name: system:cloud-node-controller
 roleRef:
   apiGroup: rbac.authorization.k8s.io
   kind: ClusterRole
   name: cluster-admin
 subjects:
 - kind: ServiceAccount
   name: cloud-node-controller
   namespace: kube-system
 ---
 kind: ClusterRoleBinding
 apiVersion: rbac.authorization.k8s.io/v1
 metadata:
   name: system:pvl-controller
 roleRef:
   apiGroup: rbac.authorization.k8s.io
   kind: ClusterRole
   name: cluster-admin
 subjects:
 - kind: ServiceAccount
   name: pvl-controller
   namespace: kube-system
 ---
 apiVersion: apps/v1
 kind: DaemonSet
 metadata:
   labels:
	 k8s-app: cloud-controller-manager
   name: cloud-controller-manager
   namespace: kube-system
 spec:
   selector:
	 matchLabels:
	   k8s-app: cloud-controller-manager
   template:
	 metadata:
	   labels:
		 k8s-app: cloud-controller-manager
	 spec:
	   serviceAccountName: cloud-controller-manager
	   hostNetwork: true
	   containers:
	   - name: cloud-controller-manager
		 image: {{.Image}}
		 command:
		 - /usr/local/bin/cloud-controller-manager
		 - --cloud-config=$(CLOUD_CONFIG)
		 - --cloud-provider=openstack
		 - --leader-elect=true
		 - --use-service-account-credentials
		 - --allocate-node-cidrs=true
		 - --configure-cloud-routes=true
		 - --cluster-cidr=172.17.0.0/16
		 env:
		   - name: CLOUD_CONFIG
			 value: /etc/config/cloud.conf
		 volumeMounts:
		 - mountPath: /etc/kubernetes/pki
		   name: k8s-certs
		   readOnly: true
		 - mountPath: /etc/ssl/certs
		   name: ca-certs
		   readOnly: true
		 - mountPath: /etc/config
		   name: openstack-cloud-config-volume
		   readOnly: true
	   tolerations:
	   - key: node.cloudprovider.kubernetes.io/uninitialized
		 value: "true"
		 effect: NoSchedule
	   - key: node-role.kubernetes.io/master
		 effect: NoSchedule
	   nodeSelector:
		 node-role.kubernetes.io/master: ""
	   volumes:
	   - hostPath:
		   path: /etc/kubernetes/pki
		   type: DirectoryOrCreate
		 name: k8s-certs
	   - hostPath:
		   path: /etc/ssl/certs
		   type: DirectoryOrCreate
		 name: ca-certs
	   - name: openstack-cloud-config-volume
		 secret:
		   secretName: {{.SecretName}}
 `
)
