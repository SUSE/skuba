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
	"strings"
	"testing"
)

func TestRandomGenerateWithLength(t *testing.T) {
	tests := []struct {
		name     string
		inputLen int
	}{
		{
			name:     "length=12",
			inputLen: 12,
		},
		{
			name:     "length=32",
			inputLen: 32,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomGenerateWithLength(tt.inputLen)
			if err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
			}

			// Compare length
			gotLen := len(got)
			if gotLen != tt.inputLen {
				t.Errorf("got len %d != expect len %d", gotLen, tt.inputLen)
			}

			// check string is all 0 or not
			if strings.Trim(string(got), "0") == "" {
				t.Error("got is not randomly")
			}
		})
	}
}
