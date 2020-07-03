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
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/oidc"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	"github.com/SUSE/skuba/internal/pkg/skuba/util"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	registerAddon(kubernetes.Gangway, GenericAddOn, renderGangwayTemplate, nil, gangwayCallbacks{}, normalPriority, []getImageCallback{GetGangwayImage})
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

func (gangwayCallbacks) beforeApply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
	// handles gangway session key
	exist, err := oidc.IsSecretExist(client, oidc.GangwaySecretName)
	if err != nil {
		return errors.Wrap(err, "unable to determine if gangway secret exists")
	}
	if !exist {
		// generates session key with length=32
		sessionKey, err := oidc.RandomGenerateWithLength(32)
		if err != nil {
			return errors.Wrap(err, "unable to generate gangway session key")
		}
		err = oidc.CreateOrUpdateToSecret(client, oidc.GangwaySecretName, oidc.GangwaySecret_SessionKey, sessionKey)
		if err != nil {
			return err
		}
	}

	// handles gangway certificate
	exist, err = oidc.IsSecretExist(client, oidc.GangwayCertSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to determine if oidc gangway cert exists")
	}
	if !exist {
		// generate certificate if not present
		if err = oidc.CreateServerCert(client, skubaconstants.PkiDir(), oidc.GangwayCertCN, util.ControlPlaneHost(addonConfiguration.ControlPlane), oidc.GangwayCertSecretName); err != nil {
			return err
		}
	}

	return nil
}

func (gangwayCallbacks) afterApply(client clientset.Interface, addonConfiguration AddonConfiguration, skubaConfiguration *skuba.SkubaConfiguration) error {
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
    usernameClaim: "email"
    apiServerURL: "https://{{.ControlPlaneHostAndPort}}"
    clusterCAPath: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
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
    caasp.suse.com/skuba-replica-ha: "true"
spec:
  replicas: 3
  revisionHistoryLimit: 3
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
            - name: GANGWAY_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: oidc-client-secret
                  key: gangway
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
