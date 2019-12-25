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
	"crypto/sha1"
	"fmt"
	"os"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktest "k8s.io/client-go/testing"

	"github.com/SUSE/skuba/pkg/skuba"
)

func TestGenerateKubeletRootCert(t *testing.T) {
	t.Run("generate kubelet root cert", func(t *testing.T) {
		if err := GenerateKubeletRootCert(); err != nil {
			t.Errorf("error not expected, but error reported generating pki: %v", err)
		}

		// test generate kubelet root cert when already exist
		if err := GenerateKubeletRootCert(); err != nil {
			t.Errorf("error not expected, but error reported when pki exist: %v", err)
		}

		pkiDir := skuba.PkiDir()
		if err := os.RemoveAll(pkiDir); err != nil {
			t.Errorf("error not expected, but error reported when removing %v: %v", pkiDir, err)
		}
	})
}

func TestDisarmKubelet(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()
	fakeClientset.PrependReactor("get", "jobs", func(action ktest.Action) (bool, runtime.Object, error) {
		obj := &batchv1.Job{
			Status: batchv1.JobStatus{
				Active:    0,
				Succeeded: 1,
				Failed:    0,
			},
		}
		return true, obj, nil
	})

	fakeNode := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake",
		},
	}

	t.Run("disarm kubelet", func(t *testing.T) {
		if err := DisarmKubelet(fakeClientset, &fakeNode, LatestVersion()); err != nil {
			t.Errorf("error not expected, but error reported generating pki: %v", err)
		}
	})
}

func TestDisarmKubeletJobName(t *testing.T) {
	tests := []struct {
		nodeName string
	}{
		{
			nodeName: "my-master-0",
		},
		{
			nodeName: "my-workder-0",
		},
		{
			nodeName: "foo-bar",
		},
		{
			nodeName: "ayy-lmao",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("disarm job name when node is: %v", tt.nodeName), func(t *testing.T) {
			fakeNode := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "my-master-0"}}
			expect := fmt.Sprintf("caasp-kubelet-disarm-%x", sha1.Sum([]byte(fakeNode.ObjectMeta.Name)))
			actual := disarmKubeletJobName(fakeNode)
			if actual != expect {
				t.Errorf("expected name \"%s\", but instead got \"%s\"", expect, actual)
			}
		})
	}
}
