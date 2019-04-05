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

package ssh

func init() {
	stateMap["kernel.load-modules"] = kernelLoadModules
	stateMap["kernel.configure-parameters"] = kernelConfigureParameters
}

func kernelLoadModules(t *Target, data interface{}) error {
	if _, _, err := t.ssh("modprobe br_netfilter"); err != nil {
		return err
	}
	err := t.UploadFileContents("/etc/modules-load.d/br_netfilter.conf", "br_netfilter")
	return err
}

func kernelConfigureParameters(t *Target, data interface{}) error {
	if _, _, err := t.ssh("sysctl -w net.ipv4.ip_forward=1"); err != nil {
		return err
	}
	_, _, err := t.ssh("sysctl -w net.bridge.bridge-nf-call-iptables=1")
	return err
}
