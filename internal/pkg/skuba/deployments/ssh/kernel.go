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

var (
	modules = []string{
		"br_netfilter",
		"vxlan",
	}

	parameters = map[string]struct {
		Attribute string
		Value     string
	}{
		"net-ipv4-ip-forward": {
			Attribute: "net.ipv4.ip_forward",
			Value:     "1",
		},
		"net-bridge-bridge-nf-call-iptables": {
			Attribute: "net.bridge.bridge-nf-call-iptables",
			Value:     "1",
		},
	}
)

func init() {
	stateMap["kernel.load-modules"] = kernelLoadModules
	stateMap["kernel.configure-parameters"] = kernelConfigureParameters
}

func kernelLoadModules(t *Target, data interface{}) error {
	for _, module := range modules {
		if err := loadModule(t, module); err != nil {
			return err
		}
	}
	return nil
}

func kernelConfigureParameters(t *Target, data interface{}) error {
	for parameterName, parameter := range parameters {
		if err := configureParameter(t, parameterName, parameter.Attribute, parameter.Value); err != nil {
			return err
		}
	}
	return nil
}

func loadModule(t *Target, module string) error {
	if _, _, err := t.ssh(fmt.Sprintf("modprobe %s", module)); err != nil {
		return err
	}
	return t.UploadFileContents(fmt.Sprintf("/etc/modules-load.d/skuba-%s.conf", module), module)
}

func configureParameter(t *Target, name, attribute, value string) error {
	if _, _, err := t.ssh(fmt.Sprintf("sysctl -w %s=%s", attribute, value)); err != nil {
		return err
	}
	return t.UploadFileContents(fmt.Sprintf("/etc/sysctl.d/90-skuba-%s.conf", name), fmt.Sprintf("%s=%s", attribute, value))
}
