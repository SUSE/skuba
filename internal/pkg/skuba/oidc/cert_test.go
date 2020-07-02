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

func TestCreateServerCert(t *testing.T) {
	tests := []struct {
		name          string
		pkiPath       string
		expectedError bool
	}{
		{
			name:    "normal case",
			pkiPath: "testdata",
		},
		{
			name:          "cannot load ca cert/key",
			pkiPath:       "lol",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := CreateServerCert(fake.NewSimpleClientset(), tt.pkiPath, "test-cn", "1.1.1.1", "test-secret")
			if tt.expectedError {
				if err == nil {
					t.Errorf("error expected on %s, but no error reported", tt.name)
				}
				return
			} else if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}
