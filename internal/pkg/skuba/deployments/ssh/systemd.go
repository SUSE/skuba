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
	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
)

// IsServiceEnabled returns if a service is enabled
func (t *Target) IsServiceEnabled(serviceName string) (bool, error) {
	klog.V(2).Infof("checking if %s is enabled", serviceName)
	isEnabled := true
	_, _, err := t.silentSsh("systemctl", "is-enabled", serviceName)
	if err != nil {
		t, isExitError := err.(*ssh.ExitError)
		if isExitError && t.ExitStatus() == 1 {
			isEnabled = false
			// the error is sane
			err = nil
		}
	}
	return isEnabled, err
}
