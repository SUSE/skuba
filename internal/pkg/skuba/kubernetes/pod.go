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

package kubernetes

import (
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func getPodContainerImageTag(client clientset.Interface, namespace string, podName string) (string, error) {
	podObject, err := client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not retrieve pod object")
	}
	containerImageWithName := podObject.Spec.Containers[0].Image
	containerImageTag := strings.Split(containerImageWithName, ":")

	return containerImageTag[len(containerImageTag)-1], nil
}

// DeletePodWithLabelSelector function deletes pod(s) with the label selector
func DeletePodWithLabelSelector(client clientset.Interface, labelSelector string) error {
	listOptions := metav1.ListOptions{LabelSelector: labelSelector}
	return client.CoreV1().Pods(metav1.NamespaceSystem).DeleteCollection(&metav1.DeleteOptions{}, listOptions)
}
