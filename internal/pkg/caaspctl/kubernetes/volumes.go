package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

const (
	VolumeMountReadOnly  = iota
	VolumeMountReadWrite = iota
)

type VolumeMountMode uint32

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
