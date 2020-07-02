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

import "testing"

func TestControlPlaneHost(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane string
		expected     string
	}{
		{
			name:         "IP",
			controlPlane: "1.1.1.1",
			expected:     "1.1.1.1",
		},
		{
			name:         "IP:PORT",
			controlPlane: "1.1.1.1:6443",
			expected:     "1.1.1.1",
		},
		{
			name:         "DNS",
			controlPlane: "a.b.c",
			expected:     "a.b.c",
		},
		{
			name:         "DNS:PORT",
			controlPlane: "a.b.c:6443",
			expected:     "a.b.c",
		},
		{
			name:         "invalid control plane",
			controlPlane: "6443:1.1.1.1",
			expected:     "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cp := ControlPlaneHost(tt.controlPlane)
			if cp != tt.expected {
				t.Errorf("expect %s not equal to got %s", tt.expected, cp)
			}
		})
	}
}

func TestControlPlaneHostAndPort(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane string
		expected     string
	}{
		{
			name:         "IP",
			controlPlane: "1.1.1.1",
			expected:     "1.1.1.1:6443",
		},
		{
			name:         "IP:6443",
			controlPlane: "1.1.1.1:6443",
			expected:     "1.1.1.1:6443",
		},
		{
			name:         "IP:8443",
			controlPlane: "1.1.1.1:8443",
			expected:     "1.1.1.1:8443",
		},
		{
			name:         "DNS",
			controlPlane: "a.b.c",
			expected:     "a.b.c:6443",
		},
		{
			name:         "DNS:6443",
			controlPlane: "a.b.c:6443",
			expected:     "a.b.c:6443",
		},
		{
			name:         "DNS:8443",
			controlPlane: "a.b.c:8443",
			expected:     "a.b.c:8443",
		},
		{
			name:         "invalid control plane",
			controlPlane: "6443:1.1.1.1",
			expected:     "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cp := ControlPlaneHostAndPort(tt.controlPlane)
			if cp != tt.expected {
				t.Errorf("expect %s not equal to got %s", tt.expected, cp)
			}
		})
	}
}
