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
	"fmt"
)

const bpffsFstabEntry = "bpffs /sys/fs/bpf bpf defaults 0 0"

func init() {
	stateMap["bpffs.mount"] = bpffsMount
}

func bpffsMount(t *Target, data interface{}) error {
	if _, _, err := t.ssh("mount | grep bpf"); err != nil {
		if _, _, err := t.ssh("mount bpffs /sys/fs/bpf -t bpf"); err != nil {
			return fmt.Errorf("Could not mount the BPFFS filesystem: %s", err)
		}
	}
	if _, _, err := t.ssh(fmt.Sprintf("if ! grep bpf /etc/fstab; then echo \"%s\" >> /etc/fstab; fi", bpffsFstabEntry)); err != nil {
		return fmt.Errorf("Could not enable the BPFFS mount in fstab")
	}
	return nil
}
