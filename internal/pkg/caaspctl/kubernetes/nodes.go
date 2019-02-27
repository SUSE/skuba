package kubernetes

import (
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	masterLabel = "node-role.kubernetes.io/master"
)

func GetMasterNodes() (*v1.NodeList, error) {
	return GetAdminClientSet().CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=", masterLabel),
	})
}

func IsMaster(node *v1.Node) bool {
	_, isMaster := node.ObjectMeta.Labels[masterLabel]
	return isMaster
}

func DisarmKubelet(node *v1.Node) error {
	return CreateAndWaitForJob(disarmKubelet(node))
}

func disarmKubelet(node *v1.Node) (string, batchv1.JobSpec) {
	return disarmKubeletJobName(node), disarmKubeletJobSpec(node)
}

func disarmKubeletJobName(node *v1.Node) string {
	return fmt.Sprintf("caasp-kubelet-disarm-%s", node.ObjectMeta.Name)
}

func disarmKubeletJobSpec(node *v1.Node) batchv1.JobSpec {
	privilegedJob := true
	return batchv1.JobSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: fmt.Sprintf("caasp-kubelet-disarm-%s", node.ObjectMeta.Name),
						// This can be simplified to use `go-systemd` or `godbus` and embedding this calling logic in a possible separate smally containerized binary
						Image: "ereslibre/opensuse-tooling:latest",
						Command: []string{
							"/bin/bash", "-c",
							strings.Join(
								[]string{
									"rm -rf /etc/kubernetes/*",
									"dbus-send --system --print-reply --dest=org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager.DisableUnitFiles array:string:'kubelet.service' boolean:false",
									"dbus-send --system --print-reply --dest=org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager.MaskUnitFiles array:string:'kubelet.service' boolean:false boolean:true",
								},
								" && ",
							),
						},
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "etc-kubernetes",
								MountPath: "/etc/kubernetes",
							},
							{
								Name:      "var-run-dbus",
								MountPath: "/var/run/dbus",
							},
						},
						SecurityContext: &v1.SecurityContext{
							Privileged: &privilegedJob,
						},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
				Volumes: []v1.Volume{
					hostMount("etc-kubernetes", "/etc/kubernetes"),
					hostMount("var-run-dbus", "/var/run/dbus"),
				},
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": node.ObjectMeta.Name,
				},
				Tolerations: []v1.Toleration{
					{
						Operator: v1.TolerationOpExists,
					},
				},
			},
		},
	}
}

func hostMount(name, mount string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: mount,
			},
		},
	}
}
