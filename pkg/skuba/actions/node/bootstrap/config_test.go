/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
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

package bootstrap

import (
	"testing"
)

func Test_LoadInitConfigurationFromFile(t *testing.T) {
	tests := []struct {
		name          string
		cfgPath       string
		expectedError bool
	}{
		{
			name:    "normal",
			cfgPath: "testdata/init.conf",
		},
		{
			name:    "cluster configuration only",
			cfgPath: "testdata/cluster.conf",
		},
		{
			name:          "config path not exist",
			cfgPath:       "testdata/not-exist.conf",
			expectedError: true,
		},
		{
			name:          "invalid api version",
			cfgPath:       "testdata/invalid.conf",
			expectedError: true,
		},
		{
			name:          "not init or cluster configuration",
			cfgPath:       "testdata/join.conf",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadInitConfigurationFromFile(tt.cfgPath)
			if tt.expectedError && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}
