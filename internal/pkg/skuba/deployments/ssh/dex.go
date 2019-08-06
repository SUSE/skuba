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

package ssh

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/SUSE/skuba/internal/pkg/skuba/dex"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	stateMap["dex.deploy"] = dexDeploy
	stateMap["dex.cert.renew"] = dexRenewCertificate
}

func dexDeploy(t *Target, data interface{}) error {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}
	err = dex.CreateCert(client, skuba.PkiDir(), skuba.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrap(err, "unable to create dex certificate")
	}

	dexFiles, err := ioutil.ReadDir(skuba.DexDir())
	if err != nil {
		return errors.Wrap(err, "could not read local dex directory")
	}

	defer t.ssh("rm -rf /tmp/dex.d")

	for _, f := range dexFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.DexDir(), f.Name()), filepath.Join("/tmp/dex.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/dex.d")
	return err
}

func dexRenewCertificate(t *Target, data interface{}) error {
	client, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "could not get admin client set")
	}

	err = dex.CreateCert(client, skuba.PkiDir(), skuba.KubeadmInitConfFile())
	if err != nil {
		return errors.Wrap(err, "unable to create dex certificate")
	}

	listOptions := metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", dex.CertCommonName)}
	err = client.CoreV1().Pods(metav1.NamespaceSystem).DeleteCollection(&metav1.DeleteOptions{}, listOptions)
	if err != nil {
		return errors.Wrap(err, "unable to delete dex pod")
	}

	return err
}
