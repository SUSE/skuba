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

package util

import (
	"reflect"
	"testing"
)

func TestUniqueStringSlice(t *testing.T) {
	tests := []struct {
		Name          string
		Slice         []string
		ExpectedSlice []string
	}{
		{
			Name:          "no duplicates",
			Slice:         []string{"one", "two", "three"},
			ExpectedSlice: []string{"one", "two", "three"},
		},
		{
			Name:          "some duplicates",
			Slice:         []string{"one", "two", "two", "three", "one", "three"},
			ExpectedSlice: []string{"one", "two", "three"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			gotSlice := UniqueStringSlice(tt.Slice)
			if !reflect.DeepEqual(gotSlice, tt.ExpectedSlice) {
				t.Errorf("slice %v does not match expected slice: %v", gotSlice, tt.ExpectedSlice)
			}
		})
	}
}
