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

package util

import (
	"fmt"

	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
)

// ControlPlaneHost parses a control plane address of the form "host:port", "ipv4:port", "[ipv6]:port" into host;
// ":port" can be eventually omitted.
// If the string is not a valid representation of network address, empty control plane returns.
func ControlPlaneHost(cp string) string {
	controlPlane, _, err := kubeadmutil.ParseHostPort(cp)
	if err != nil {
		return ""
	}
	return controlPlane
}

// ControlPlaneHostAndPort parses a control plane address of the form "host:port", "ipv4:port", "[ipv6]:port" into host and port;
// ":port" can be eventually omitted.
// If the port is empty, a default control plane port 6443 added.
// If the string is not a valid representation of network address, empty control plane returns.
func ControlPlaneHostAndPort(cp string) string {
	controlPlaneHost, controlPlanePort, err := kubeadmutil.ParseHostPort(cp)
	if err != nil {
		return ""
	}
	if controlPlanePort == "" {
		controlPlanePort = "6443"
	}
	return fmt.Sprintf("%s:%s", controlPlaneHost, controlPlanePort)
}
