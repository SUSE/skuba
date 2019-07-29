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
	"k8s.io/kubernetes/cmd/kubeadm/app/images"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetKuredImage() string {
	return images.GetGenericImage(skuba.ImageRepository, "kured",
		kubernetes.CurrentAddonVersion(kubernetes.Kured))
}

func KuredLockExists() (bool, error) {
	clientSet, err := kubernetes.GetAdminClientSet()
	if err != nil {
		return false, errors.Wrap(err, "unable to get admin client set")
	}
	kuredDaemonSet, err := clientSet.AppsV1().DaemonSets(metav1.NamespaceSystem).Get("kured", metav1.GetOptions{})
	if err != nil {
		return false, errors.Wrap(err, "unable to get kured daemonset")
	}
	_, ok := kuredDaemonSet.GetAnnotations()["weave.works/kured-node-lock"]
	return ok, nil
}
