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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDisarmKubeletJobName(t *testing.T) {
	testCases := []struct {
		node         *v1.Node
		expectedName string
	}{
		{
			node:         &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "my-master-0"}},
			expectedName: "caasp-kubelet-disarm-5a5d6b47201c0dab0034cb6959d24ddb35e56546",
		},
		{
			node:         &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "my-worker-0"}},
			expectedName: "caasp-kubelet-disarm-6db327a52a7ce5599572759bc60bdcbe63664630",
		},
		{
			node:         &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "foo-bar"}},
			expectedName: "caasp-kubelet-disarm-db7329d5a3f381875ea6ce7e28fe1ea536d0acaf",
		},
		{
			node:         &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "ayy-lmao"}},
			expectedName: "caasp-kubelet-disarm-dd07f282f4cf6821a365a3d01b43038d20d5c7fe",
		},
	}

	for _, tc := range testCases {
		name := disarmKubeletJobName(tc.node)
		if name != tc.expectedName {
			t.Errorf("expected name \"%s\", but instead got \"%s\"",
				tc.expectedName, name)
		}
	}
}
