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

package addons

import (
	"fmt"

	versionutil "k8s.io/apimachinery/pkg/util/version"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
)

// KubernetesVersionAtLeast will panic if the given version is not a semantic
// one. Thus, before calling this function, make sure that this version comes
// from a safe place (i.e skuba source code).
func (renderContext renderContext) KubernetesVersionAtLeast(version string) bool {
	return renderContext.config.ClusterVersion.AtLeast(versionutil.MustParseSemantic(version))
}

func (renderContext renderContext) ClusterName() string {
	return renderContext.config.ClusterName
}

func (renderContext renderContext) ControlPlane() string {
	return renderContext.config.ControlPlane
}

func (renderContext renderContext) ControlPlaneHost() string {
	controlPlaneHost, _, err := kubeadmutil.ParseHostPort(renderContext.ControlPlane())
	if err != nil {
		return ""
	}
	return controlPlaneHost
}

func (renderContext renderContext) ControlPlaneHostAndPort() string {
	controlPlaneHost, controlPlanePort, err := kubeadmutil.ParseHostPort(renderContext.ControlPlane())
	if err != nil {
		return ""
	}
	if controlPlanePort == "" {
		controlPlanePort = "6443"
	}
	return fmt.Sprintf("%s:%s", controlPlaneHost, controlPlanePort)
}
