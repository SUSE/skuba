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

package remove

import (
	"testing"
 
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_RemoveMasterNode(t *testing.T) {
	node1 := v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "master-1", Labels: map[string]string{"node-role.kubernetes.io/master": "" }}}

	test := []struct {
		name			string
		target			string
		clientset		*fake.Clientset
		errorExpected	bool
		errorMessage	string
	}{
		{
			name:			"remove last master from cluster",
			target: 		"master-1",
			clientset:		fake.NewSimpleClientset(&v1.NodeList{Items: []v1.Node{node1}}),
			errorExpected: 	true,
			errorMessage: 	"could not remove last master of the cluster",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			actual := Remove(tt.clientset, tt.target, 0)
			if tt.errorExpected && actual.Error() == "" {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.errorExpected && actual.Error() != "" {
				t.Errorf("error not expected on %s, but an error was reported (%s)", tt.name, actual.Error())
				return
			} else if tt.errorExpected && actual.Error() != tt.errorMessage {
				t.Errorf("(%v) expected on %s, but different error message reported (%v)", tt.errorMessage, tt.name, actual.Error())
				return
			}
		})
	}
}

