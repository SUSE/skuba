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
	"time"

	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

func CreateJob(client clientset.Interface, name string, spec batchv1.JobSpec) (*batchv1.Job, error) {
	_, err := client.BatchV1().Jobs(metav1.NamespaceSystem).Create(&batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
		},
		Spec: spec,
	})
	return nil, err
}

func DeleteJob(client clientset.Interface, name string) error {
	return client.BatchV1().Jobs(metav1.NamespaceSystem).Delete(name, &metav1.DeleteOptions{})
}

func CreateAndWaitForJob(client clientset.Interface, name string, spec batchv1.JobSpec) error {
	_, err := CreateJob(client, name, spec)
	defer DeleteJob(client, name)
	if err != nil {
		return err
	}
	for i := 0; i < 300; i++ {
		job, err := client.BatchV1().Jobs(metav1.NamespaceSystem).Get(name, metav1.GetOptions{})

		if err != nil {
			klog.V(1).Infof("failed to get status for job %s, continuing...", name)
		} else {
			if job.Status.Active > 0 {
				klog.V(1).Infof("job %s is active, waiting...", name)
			} else {
				if job.Status.Succeeded > 0 {
					klog.V(1).Infof("job %s executed successfully", name)
					return nil
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	return errors.New(fmt.Sprintf("failed waiting for job %s", name))
}
