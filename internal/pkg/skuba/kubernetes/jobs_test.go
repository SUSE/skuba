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
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktest "k8s.io/client-go/testing"
)

func Test_CreateJob(t *testing.T) {
	tests := []struct {
		name          string
		errExpected   bool
		errMessage    string
		fakeClientset *fake.Clientset
		jobName       string
		jobSpec       batchv1.JobSpec
	}{
		{
			name:          "create job",
			fakeClientset: fake.NewSimpleClientset(),
			jobName:       "create-job",
			jobSpec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "fake",
								Image: "fake",
							},
						},
					},
				},
			},
			errExpected: false,
		},
		{
			name: "should fail when job exist",
			fakeClientset: fake.NewSimpleClientset(
				&batchv1.Job{
					TypeMeta: metav1.TypeMeta{Kind: "Job"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "create-job",
						Namespace: metav1.NamespaceSystem,
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "fake",
										Image: "fake",
									},
								},
								RestartPolicy: corev1.RestartPolicyOnFailure,
							},
						},
					},
					Status: batchv1.JobStatus{
						Active:    0,
						Succeeded: 1,
						Failed:    0,
					},
				},
			),
			jobName: "create-job",
			jobSpec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "fake",
								Image: "fake",
							},
						},
						RestartPolicy: corev1.RestartPolicyNever,
					},
				},
			},
			errExpected: true,
			errMessage:  "jobs.batch \"create-job\" already exists",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateJob(tt.fakeClientset, tt.jobName, tt.jobSpec)
			if tt.errExpected {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.errMessage {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.errMessage)
					return
				}
			} else if !tt.errExpected {
				if err != nil {
					t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
					return
				}
				_, err = tt.fakeClientset.BatchV1().Jobs(metav1.NamespaceSystem).Get(tt.jobName, metav1.GetOptions{})
				if err != nil {
					t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
					return
				}
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := []struct {
		name         string
		jobName      string
		jobNamespace string
		expectErrMsg string
	}{
		{
			name:         "delete job",
			jobName:      "valid",
			jobNamespace: metav1.NamespaceSystem,
		},
		{
			name:         "delete job when job exist in other namespace",
			jobName:      "valid",
			jobNamespace: metav1.NamespaceDefault,
			expectErrMsg: "jobs.batch \"valid\" not found",
		},
		{
			name:         "delete job when job not exist",
			jobName:      "invalid",
			jobNamespace: metav1.NamespaceSystem,
			expectErrMsg: "jobs.batch \"invalid\" not found",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(
				&batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "valid",
						Namespace: tt.jobNamespace,
					},
				},
			)

			err := DeleteJob(fakeClientset, tt.jobName)
			if tt.expectErrMsg != "" {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.expectErrMsg {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.expectErrMsg)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
				return
			}
		})
	}
}

func Test_CreateAndWaitForJob(t *testing.T) {
	tests := []struct {
		name           string
		timeout        int
		reactor        bool
		reactActive    int32
		reactSucceeded int32
		reactFailed    int32
		expectErrMsg   string
	}{
		{
			name:           "create and wait for job",
			timeout:        5,
			reactor:        true,
			reactSucceeded: 1,
		},
		{
			name:         "create and wait for job when job pend active",
			timeout:      2,
			reactor:      true,
			reactActive:  1,
			expectErrMsg: "failed waiting for job test",
		},
		{
			name:         "create and wait for job when job failed",
			timeout:      2,
			reactor:      true,
			reactFailed:  1,
			expectErrMsg: "failed waiting for job test",
		},
		{
			name:         "create and wait for job when failed to get job",
			timeout:      2,
			expectErrMsg: "failed waiting for job test",
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset()
			if tt.reactor {
				fakeClientset.PrependReactor("get", "jobs", func(action ktest.Action) (bool, runtime.Object, error) {
					obj := &batchv1.Job{
						Status: batchv1.JobStatus{
							Active:    tt.reactActive,
							Succeeded: tt.reactSucceeded,
							Failed:    tt.reactFailed,
						},
					}
					return true, obj, nil
				})
			}

			err := CreateAndWaitForJob(fakeClientset, "test", batchv1.JobSpec{}, tt.timeout)
			if tt.expectErrMsg != "" {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
					return
				}
				if err.Error() != tt.expectErrMsg {
					t.Errorf("returned error (%v) does not match the expected one (%v)", err.Error(), tt.expectErrMsg)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err.Error())
				return
			}
		})
	}
}
