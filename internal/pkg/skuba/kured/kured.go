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

package kured

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	kuredDSName             = "kured"
	kuredLockAnnotationJson = `{"metadata":{"annotations":{"weave.works/kured-node-lock":"'{\"nodeID\":\"manual\"}'"}}}`
)

func LockExists(client clientset.Interface) (bool, error) {
	kuredDaemonSet, err := client.AppsV1().DaemonSets(metav1.NamespaceSystem).Get("kured", metav1.GetOptions{})
	if err != nil {
		return false, errors.Wrap(err, "unable to get kured daemonset")
	}
	_, ok := kuredDaemonSet.GetAnnotations()["weave.works/kured-node-lock"]
	return ok, nil
}

func Lock(client clientset.Interface) error {
	_, err := client.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch(kuredDSName, types.StrategicMergePatchType, []byte(kuredLockAnnotationJson))
	if err != nil {
		return errors.Wrap(err, "unable to patch daemonset with kured locking annotation")
	}
	klog.V(1).Info("successfully annotated daemonset with kured locking annotation")
	return nil
}

func Unlock(client clientset.Interface) error {
	// jsonpatch expects a ~1 escape sequence for a forward slash '/'
	// the annotation we want to remove is 'weave.works/kured-node-lock'
	payload := []byte(`[{"op":"remove","path":"/metadata/annotations/weave.works~1kured-node-lock"}]`)
	_, err := client.AppsV1().DaemonSets(metav1.NamespaceSystem).Patch(kuredDSName, types.JSONPatchType, payload)
	if err != nil {
		return errors.Wrap(err, "unable to patch daemonset with kured unlocking annotation")
	}
	klog.V(1).Info("successfully removed kured locking annotation")
	return nil
}
