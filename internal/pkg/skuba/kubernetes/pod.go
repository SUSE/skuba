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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func getPodContainerImageTag(client clientset.Interface, namespace string, podName string) (string, error) {
	podObject, err := client.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "could not retrieve pod object")
	}
	return getPodContainerImageTagFromPodObject(podObject), nil
}

// return the image version tag without calling the api, just with the Pod object
func getPodContainerImageTagFromPodObject(podObject *v1.Pod) string {
	containerImageWithName := podObject.Spec.Containers[0].Image
	containerImageTag := strings.Split(containerImageWithName, ":")

	return containerImageTag[len(containerImageTag)-1]
}

// return a pod object matched from a podList
func getPodFromPodList(list *v1.PodList, name string) (*v1.Pod, error) {
	for _, pod := range list.Items {
		if name == pod.Name {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("could not find pod %s in pod list", name)
}
