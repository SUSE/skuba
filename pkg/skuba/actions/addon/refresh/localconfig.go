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

package refresh

import (
	"fmt"

	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/SUSE/skuba/internal/pkg/skuba/addons"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
)

// AddonsBaseManifest implements the `skuba addon refresh localconfig` command.
func AddonsBaseManifest(client clientset.Interface) error {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return err
	}

	clusterConfiguration, err := kubeadm.GetClusterConfiguration(client)
	if err != nil {
		return errors.Wrap(err, "Could not fetch cluster configuration")
	}

	// re-render all addons manifest
	addonConfiguration := addons.AddonConfiguration{
		ClusterVersion: currentClusterVersion,
		ControlPlane:   clusterConfiguration.ControlPlaneEndpoint,
		ClusterName:    clusterConfiguration.ClusterName,
	}
	for addonName, addon := range addons.Addons {
		if !addon.IsPresentForClusterVersion(currentClusterVersion) {
			continue
		}
		if err := addon.Write(addonConfiguration); err != nil {
			return errors.Wrapf(err, "failed to refresh addon %s manifest", string(addonName))
		}
	}

	fmt.Println("Successfully refreshed addons base manifests")
	return nil
}
