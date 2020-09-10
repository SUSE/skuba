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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/oidc"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	registerAddon(kubernetes.Dex, GenericAddOn, renderDexTemplate, nil, dexCallbacks{}, normalPriority, []getImageCallback{GetDexImage})
}

func GetDexImage(clusterVersion *version.Version, imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository(clusterVersion), "caasp-dex", imageTag)
}

func (renderContext renderContext) DexImage() string {
	return GetDexImage(renderContext.config.ClusterVersion, kubernetes.AddonVersionForClusterVersion(kubernetes.Dex, renderContext.config.ClusterVersion).Version)
}

func renderDexTemplate(addonConfiguration AddonConfiguration) string {
	return dexManifest
}

type dexCallbacks struct{}

func (dexCallbacks) beforeApply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	// handles oidc client-secret
	exist, err := oidc.IsSecretExist(client, oidc.ClientSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to determine if oidc client-secret exists")
	}
	if !exist {
		// generate client secret with length=12
		// client secret is used by auth client (gangway) to authenticate to auth server (dex)
		clientSecret, err := oidc.RandomGenerateWithLength(12)
		if err != nil {
			return errors.Wrap(err, "unable to generate oidc client-secret")
		}
		err = oidc.CreateOrUpdateToSecret(client, oidc.ClientSecretName, oidc.ClientSecretKey_Gangway, clientSecret)
		if err != nil {
			return err
		}
	}

	// handles dex certificate
	exist, err = oidc.IsSecretExist(client, oidc.DexCertSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to determine if oidc dex cert exists")
	}
	if !exist {
		// try to use local server certificate if present
		if err := oidc.TryToUseLocalServerCert(client, oidc.DexServerCertAndKeyBaseFileName, oidc.DexCertSecretName); err == nil {
			return nil
		}

		// sign the server certificate by cluster CA or custom OIDC CA if server certificate not present
		if err := oidc.SignServerWithLocalCACertAndKey(client, oidc.DexCertCN, util.ControlPlaneHost(addonConfiguration.ControlPlane), oidc.DexCertSecretName); err != nil {
			return err
		}
	}

	return nil
}

func (dexCallbacks) afterApply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
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
    issuer: https://{{.ControlPlaneHost}}:32000

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

    oauth2:
      skipApprovalScreen: true

    staticClients:
    - id: oidc
      redirectURIs:
      - 'https://{{.ControlPlaneHost}}:32001/callback'
      name: 'OIDC'
      # the secretEnv supports when dex >= 2.23.0
      secretEnv: OIDC_GANGWAY_CLIENT_SECRET
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
  labels:
    app: oidc-dex
    caasp.suse.com/skuba-replica-ha: "true"
spec:
  replicas: 3
  revisionHistoryLimit: 3
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
          - /usr/bin/caasp-dex
          - serve
          - /etc/dex/cfg/config.yaml
        env:
          - name: OIDC_GANGWAY_CLIENT_SECRET
            valueFrom:
              secretKeyRef:
                name: oidc-client-secret
                key: gangway
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
# Follow upstream example https://github.com/dexidp/dex/blob/4bede5eb80822fc3a7fc9edca0ed2605cd339d17/examples/k8s/dex.yaml#L116-L126
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["create"]
- apiGroups: ["dex.coreos.com"]
  resources: ["*"]
  verbs: ["*"]
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
