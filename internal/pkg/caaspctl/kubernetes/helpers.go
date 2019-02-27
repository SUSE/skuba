package kubernetes

import (
	v1 "k8s.io/api/core/v1"
)

func VolumeMount(name, mount string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:      name,
		MountPath: mount,
	}
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
