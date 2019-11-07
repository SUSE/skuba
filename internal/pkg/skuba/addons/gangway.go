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
	"github.com/SUSE/skuba/internal/pkg/skuba/gangway"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

func init() {
	registerAddon(kubernetes.Gangway, renderGangwayTemplate, gangwayCallbacks{}, normalPriority, []getImageCallback{GetGangwayImage})
}

func GetGangwayImage(imageTag string) string {
	return images.GetGenericImage(skubaconstants.ImageRepository, "gangway", imageTag)
}

func (renderContext renderContext) GangwayImage() string {
	return GetGangwayImage(kubernetes.AddonVersionForClusterVersion(kubernetes.Gangway, renderContext.config.ClusterVersion).Version)
}

func renderGangwayTemplate(addonConfiguration AddonConfiguration) string {
	return gangwayManifest
}

type gangwayCallbacks struct{}

func (gangwayCallbacks) beforeApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}
	gangwaySecretExists, err := gangway.GangwaySecretExists(client)
	if err != nil {
		return errors.Wrap(err, "unable to determine if gangway secret exists")
	}
	gangwayCertExists, err := gangway.GangwayCertExists(client)
	if err != nil {
		return errors.Wrap(err, "unable to determine if gangway cert exists")
	}
	if !gangwaySecretExists {
		key, err := gangway.GenerateSessionKey()
		if err != nil {
			return errors.Wrap(err, "unable to generate gangway session key")
		}
		err = gangway.CreateOrUpdateSessionKeyToSecret(client, key)
		if err != nil {
			return errors.Wrap(err, "unable to create/update gangway session key to secret")
		}
	}
	err = gangway.CreateCert(client, skubaconstants.PkiDir(), skubaconstants.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrap(err, "unable to create gangway certificate")
	}
	if gangwayCertExists {
		if err := gangway.RestartPods(client); err != nil {
			return err
		}
	}
	return nil
}

func (gangwayCallbacks) afterApply(addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	return nil
}

const (
	gangwayManifest = `---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: oidc-gangway
  namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: oidc-gangway-config
  namespace: kube-system
data:
  gangway.yaml: |
    clusterName: {{.ClusterName}}

    redirectURL: "https://{{.ControlPlaneHost}}:32001/callback"
    scopes: ["openid", "email", "groups", "profile", "offline_access"]

    serveTLS: true
    authorizeURL: "https://{{.ControlPlaneHost}}:32000/auth"
    tokenURL: "https://{{.ControlPlaneHost}}:32000/token"
    keyFile: /etc/gangway/pki/tls.key
    certFile: /etc/gangway/pki/tls.crt

    clientID: "oidc"
    clientSecret: "{{.GangwayClientSecret}}"
    usernameClaim: "email"
    apiServerURL: "https://{{.ControlPlaneHostAndPort}}"
    cluster_ca_path: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
    trustedCAPath: /etc/gangway/pki/ca.crt
    customHTMLTemplatesDir: /usr/share/caasp-gangway/web/templates/caasp
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oidc-gangway
  namespace: kube-system
  labels:
    app: oidc-gangway
spec:
  replicas: 3
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
      annotations:
        {{.AnnotatedVersion}}
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
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
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
`
)
