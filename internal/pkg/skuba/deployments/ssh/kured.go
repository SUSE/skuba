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
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
)

func init() {
	stateMap["kured.deploy"] = kuredDeploy
	stateMap["kured.lock"] = kuredLock
	stateMap["kured.unlock"] = kuredUnlock
}

const (
	kuredDSName             = "kured"
	kuredLockAnnotationJson = `{"metadata":{"annotations":{"weave.works/kured-node-lock":"'{\"nodeID\":\"manual\"}'"}}}`
)

func kuredDeploy(t *Target, data interface{}) error {
	kuredFiles, err := ioutil.ReadDir(skuba.KuredDir())
	if err != nil {
		return errors.Wrap(err, "could not read local kured directory")
	}

	defer t.ssh("rm -rf /tmp/kured.d")

	for _, f := range kuredFiles {
		if err := t.target.UploadFile(filepath.Join(skuba.KuredDir(), f.Name()), filepath.Join("/tmp/kured.d", f.Name())); err != nil {
			return err
		}
	}

	_, _, err = t.ssh("kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/kured.d")
	return err
}

func kuredLock(t *Target, data interface{}) error {
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}
	_, err = clientSet.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch(kuredDSName, types.StrategicMergePatchType, []byte(kuredLockAnnotationJson))
	if err != nil {
		return errors.Wrap(err, "unable to patch daemonset with kured locking annotation")
	}
	klog.V(1).Info("successfully annotated daemonset with kured locking annotation")
	return err
}

func kuredUnlock(t *Target, data interface{}) error {
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return errors.Wrap(err, "unable to get admin client set")
	}
	// jsonpatch expects a ~1 escape sequence for a forward slash '/'
	// the annotation we want to remove is 'weave.works/kured-node-lock'
	payload := []byte(`[{"op":"remove","path":"/metadata/annotations/weave.works~1kured-node-lock"}]`)
	_, err = clientSet.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch(kuredDSName, types.JSONPatchType, payload)
	if err != nil {
		return errors.Wrap(err, "unable to patch daemonset with kured unlocking annotation")
	}
	klog.V(1).Info("successfully removed kured locking annotation")
	return err
}
