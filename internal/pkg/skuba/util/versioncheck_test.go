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

package util

import (
	"testing"
)

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		exp        bool
	}{
		{
			name:       "equal, should match",
			version:    "1.7.5",
			constraint: "1.7.5",
			exp:        true,
		},
		{
			name:       "greater or equal, should match",
			version:    "1.7.5",
			constraint: ">=1.7.5",
			exp:        true,
		},
		{
			name:       "lower or equal, should match",
			version:    "1.7.5",
			constraint: "<=1.7.5",
			exp:        true,
		},
		{
			name:       "greater, should match",
			version:    "1.7.5",
			constraint: ">1.7.0",
			exp:        true,
		},
		{
			name:       "lower, should match",
			version:    "1.7.5",
			constraint: "<1.8.0",
			exp:        true,
		},
		{
			name:       "equal, should not match, is lower",
			version:    "1.7.5",
			constraint: "1.5.3",
			exp:        false,
		},
		{
			name:       "equal, should not match, is greater",
			version:    "1.7.5",
			constraint: "1.8.0",
			exp:        false,
		},
		{
			name:       "greater, should not match",
			version:    "1.7.5",
			constraint: ">1.7.5",
			exp:        false,
		},
		{
			name:       "lower, should not match",
			version:    "1.7.5",
			constraint: "<1.7.5",
			exp:        false,
		},
		{
			name:       "greater or equal, should not match",
			version:    "1.7.5",
			constraint: ">=1.8.0",
			exp:        false,
		},
		{
			name:       "lower or equal, should not match",
			version:    "1.7.5",
			constraint: "<=1.5.0",
			exp:        false,
		},
		{
			name:       "greater, with rev",
			version:    "1.6.6-rev5",
			constraint: ">1.6.0",
			exp:        true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmp := VersionCompare(tt.version, tt.constraint)
			if cmp != tt.exp {
				t.Errorf("version comparison does not match, expected %v, got %v", tt.exp, cmp)
			}
		})
	}
}
