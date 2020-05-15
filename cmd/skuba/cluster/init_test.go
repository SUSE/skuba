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

package cluster

import "testing"

func Test_isValidClusterName(t *testing.T) {
	type args struct {
		clustername string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty string is invalid",
			args: args{clustername: ""},
			want: false,
		},
		{
			name: "ascii with only spaces is invalid",
			args: args{clustername: " "},
			want: false,
		},
		{
			name: "non-ascii is invalid",
			args: args{clustername: "🦍"},
			want: false,
		},
		{
			name: "mixed ascii/non-ascii is invalid",
			args: args{clustername: "monkey🦍"},
			want: false,
		},
		{
			name: "word are valid",
			args: args{clustername: "monkey"},
			want: true,
		},
		{
			name: "letters numbers dashes underscores are valid",
			args: args{clustername: "my_cluster-01"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidClusterName(tt.args.clustername); got != tt.want {
				t.Errorf("isValidClusterName() = %v, want %v", got, tt.want)
			}
		})
	}
}
