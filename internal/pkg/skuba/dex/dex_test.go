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

package dex

import (
	"strings"
	"testing"
)

func Test_GenerateClientSecret(t *testing.T) {
	tests := []struct {
		name      string
		inputLen  int
		expectLen int
	}{
		{
			name:      "normal case",
			expectLen: 24,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateClientSecret()
			gotLen := len(got)

			// Compare length
			if gotLen != tt.expectLen {
				t.Errorf("got len %d != expect len %d", gotLen, tt.expectLen)
			}

			// check string is all 0 or not
			if strings.Trim(got, "0") == "" {
				t.Error("got is not randomly")
			}
		})
	}
}
