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

package kubernetes

import (
	"encoding/json"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func Test_VolumeMount(t *testing.T) {
	tests := []struct {
		name                string
		volumeName          string
		volumeMountPath     string
		volumeMountReadOnly bool
		expect              corev1.VolumeMount
	}{
		{
			name:                "volume mount read and write",
			volumeName:          "test",
			volumeMountPath:     "/etc/test",
			volumeMountReadOnly: false,
			expect: corev1.VolumeMount{
				Name:      "test",
				MountPath: "/etc/test",
			},
		},
		{
			name:                "volume mount read only",
			volumeName:          "test",
			volumeMountPath:     "/etc/test",
			volumeMountReadOnly: true,
			expect: corev1.VolumeMount{
				Name:      "test",
				MountPath: "/etc/test",
				ReadOnly:  true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			volumeMountMode := VolumeMountReadWrite
			if tt.volumeMountReadOnly {
				volumeMountMode = VolumeMountReadOnly
			}
			actual := VolumeMount(tt.volumeName, tt.volumeMountPath, volumeMountMode)
			if !reflect.DeepEqual(actual, tt.expect) {
				actualData, err := json.Marshal(actual)
				if err != nil {
					t.Errorf("error not expected while convert returned result to json data (%v)", err)
					return
				}
				expectData, err := json.Marshal(tt.expect)
				if err != nil {
					t.Errorf("error not expected while convert returned result to json data (%v)", err)
					return
				}
				t.Errorf("returned result (%s) does not match the expected one (%s)", actualData, expectData)
				return
			}
		})
	}
}

func Test_HostMount(t *testing.T) {
	tests := []struct {
		name          string
		volumeName    string
		hostMountPath string
		expect        corev1.Volume
	}{
		{
			name:          "mount host volume",
			volumeName:    "test",
			hostMountPath: "/etc/test",
			expect: corev1.Volume{
				Name: "test",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/etc/test",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			actual := HostMount(tt.volumeName, tt.hostMountPath)
			if !reflect.DeepEqual(actual, tt.expect) {
				actualData, err := json.Marshal(actual)
				if err != nil {
					t.Errorf("error not expected while convert returned result to json data (%v)", err)
					return
				}
				expectData, err := json.Marshal(tt.expect)
				if err != nil {
					t.Errorf("error not expected while convert returned result to json data (%v)", err)
					return
				}
				t.Errorf("returned result (%s) does not match the expected one (%s)", actualData, expectData)
				return
			}
		})
	}
}
