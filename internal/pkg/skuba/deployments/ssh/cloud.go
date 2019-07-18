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

package ssh

import (
	"bytes"
	"io/ioutil"
	"text/template"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/cmd/kubeadm/app/images"
)

func init() {
	stateMap["cloud.deploy-controller-manager"] = cloudDeployControllerManager
}

func cloudDeployControllerManager(t *Target, data interface{}) error {
	template, err := template.New("").Parse(kubernetes.OpenstackCloudControllerManagerTemplate)
	if err != nil {
		return err
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered,
		struct {
			Image      string
			SecretName string
		}{
			Image:      images.GetGenericImage(skuba.ImageRepository, "hyperkube", kubernetes.LatestVersion().String()),
			SecretName: kubernetes.OpenstackCloudControllerManagerSecretName,
		},
	)
	if err != nil {
		return err
	}

	defer t.ssh("rm -rf /tmp/cloud.d")
	if err := t.target.UploadFileContents("/tmp/cloud.d/openstack-cloud-controller-manager.yaml", rendered.String()); err != nil {
		return err
	}

	cloudConfContents, err := ioutil.ReadFile(skuba.OpenstackCloudConfFile())
	if err != nil {
		return err
	}

	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Secrets(metav1.NamespaceSystem).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubernetes.OpenstackCloudControllerManagerSecretName,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string][]byte{
			"cloud.conf": cloudConfContents,
		},
	})
	if err != nil {
		return err
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cloud.d")
	return err
}
