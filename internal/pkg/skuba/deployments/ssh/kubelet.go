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

package ssh

import (
	"github.com/SUSE/skuba/internal/pkg/skuba/deployments/ssh/assets"
)

func init() {
	stateMap["kubelet.configure"] = kubeletConfigure
	stateMap["kubelet.enable"] = kubeletEnable
}

func kubeletConfigure(t *Target, data interface{}) error {
	isSUSE, err := t.target.IsSUSEOS()
	if err != nil {
		return err
	}
	if isSUSE {
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/etc/sysconfig/kubelet", assets.KubeletSysconfig); err != nil {
			return err
		}
	} else {
		if err := t.UploadFileContents("/lib/systemd/system/kubelet.service", assets.KubeletService); err != nil {
			return err
		}
		if err := t.UploadFileContents("/etc/systemd/system/kubelet.service.d/10-kubeadm.conf", assets.KubeadmService); err != nil {
			return err
		}
	}
	_, _, err = t.ssh("systemctl", "daemon-reload")
	return err
}

func kubeletEnable(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "enable", "kubelet")
	return err
}
