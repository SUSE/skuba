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
	"github.com/SUSE/skuba/internal/pkg/skuba/dex"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

var gangwayClientSecret string

func init() {
	registerAddon(kubernetes.Dex, renderDexTemplate, dexCallbacks{}, normalPriority)
}

func (renderContext renderContext) DexImage() string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "dex",
		kubernetes.AddonVersionForClusterVersion(kubernetes.Dex, renderContext.config.ClusterVersion).Version)
}

func (renderContext renderContext) GangwayClientSecret() string {
	if len(gangwayClientSecret) == 0 {
		gangwayClientSecret = dex.GenerateClientSecret()
	}
	return gangwayClientSecret
}

func renderDexTemplate(addonConfiguration AddonConfiguration) string {
	return dexManifest
}

type dexCallbacks struct{}

func (dexCallbacks) beforeApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}
	dexCertExists, err := dex.DexCertExists(client)
	if err != nil {
		return errors.Wrap(err, "unable to determine if dex certificate exists")
	}
	err = dex.CreateCert(client, skubaconstants.PkiDir(), skubaconstants.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrap(err, "unable to create dex certificate")
	}
	if dexCertExists {
		if err := dex.RestartPods(client); err != nil {
			return err
		}
	}
	return nil
}

func (dexCallbacks) afterApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

const (
	dexManifest = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: oidc-dex
  namespace: kube-system
---
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
      issuer: "SUSE CaaS Platform"
      theme: "caasp"
      dir: /usr/share/dex/web

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

    oauth2:
      skipApprovalScreen: true

    staticClients:
    - id: oidc
      redirectURIs:
      - 'https://{{.ControlPlane}}:32001/callback'
      name: 'OIDC'
      secret: {{.GangwayClientSecret}}
      trustedPeers:
      - oidc-cli
    - id: oidc-cli
      public: true
      redirectURIs:
      - 'urn:ietf:wg:oauth:2.0:oob'
      name: 'OIDC CLI'
      secret: swac7qakes7AvucH8bRucucH
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oidc-dex
  namespace: kube-system
spec:
  replicas: 3
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
      annotations:
        {{.AnnotatedVersion}}
    spec:
      serviceAccountName: oidc-dex
      containers:
      - name: oidc-dex
        image: {{.DexImage}}
        imagePullPolicy: IfNotPresent
        command:
          - /usr/bin/dex
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
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
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
  verbs: ["get", "list"]
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
)
