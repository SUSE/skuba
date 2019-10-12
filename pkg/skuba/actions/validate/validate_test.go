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

package validate

import (
	"testing"
)

func TestNodeName(t *testing.T) {
	testCases := []struct {
		nodename  string
		expectErr bool
	}{
		{
			nodename:  "my-master-0",
			expectErr: false,
		},
		{
			nodename:  "my.worker.0",
			expectErr: false,
		},
		{
			nodename:  "my_master_0",
			expectErr: true,
		},
		{
			nodename:  "some-too-long-hostname-foo-bar-baz-fizz-buzz-xyz-zzy-kaboom-ayy-lmao",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		err := NodeName(tc.nodename)
		if tc.expectErr && (err == nil) {
			t.Errorf("expected error for node name \"%s\", but no error returned",
				tc.nodename)
		} else if !tc.expectErr && (err != nil) {
			t.Error(err)
		}
	}
}
