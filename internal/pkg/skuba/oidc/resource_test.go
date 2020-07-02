/*
 * Copyright (c) 2020 SUSE LLC.
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

package oidc

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestSecret(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		namespace  string
		key        string
		value      []byte
	}{
		{
			name:       "oidc-dex certificate exist",
			secretName: "oidc-dex-cert",
			key:        "a",
			value:      []byte("1"),
		},
		{
			name:       "oidc-gangway certificate exist",
			secretName: "oidc-gangway-cert",
			key:        "b",
			value:      []byte("2"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			exist, err := IsSecretExist(client, tt.secretName)
			if exist != false && err != nil {
				t.Error("expected certificate not exist and without error")
				return
			}

			err = CreateOrUpdateToSecret(client, tt.secretName, tt.key, tt.value)
			if err != nil {
				t.Errorf("expected not error, but error reported %v", err)
				return
			}

			exist, err = IsSecretExist(client, tt.secretName)
			if exist != true && err != nil {
				t.Error("expected certificate exist and without error")
				return
			}
		})
	}
}
