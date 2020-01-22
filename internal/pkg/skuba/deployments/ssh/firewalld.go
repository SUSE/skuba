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

package ssh

import (
	"k8s.io/klog"
)

func init() {
	stateMap["firewalld.disable"] = firewalldDisable
}

func firewalldDisable(t *Target, data interface{}) error {
	_, _, err := t.ssh("systemctl", "cat", "firewalld")
	if err == nil {
		_, _, err := t.ssh("systemctl", "disable", "--now", "firewalld")
		return err
	}
	klog.V(4).Info("=== Could not find firewalld.service ===")
	return nil
}
