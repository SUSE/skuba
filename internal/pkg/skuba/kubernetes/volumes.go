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

package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

type VolumeMountMode uint

const (
	VolumeMountReadOnly  VolumeMountMode = iota
	VolumeMountReadWrite VolumeMountMode = iota
)

func VolumeMount(name, mount string, mode VolumeMountMode) v1.VolumeMount {
	res := v1.VolumeMount{
		Name:      name,
		MountPath: mount,
	}
	if mode == VolumeMountReadOnly {
		res.ReadOnly = true
	}
	return res
}

func HostMount(name, mount string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: mount,
			},
		},
	}
}
