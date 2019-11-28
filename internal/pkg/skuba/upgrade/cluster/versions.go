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

package cluster

import (
	"k8s.io/apimachinery/pkg/util/version"
	clientset "k8s.io/client-go/kubernetes"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

func nextAvailableVersionsForVersion(currentClusterVersion *version.Version, availableVersions []*version.Version) (nextPatch *version.Version, nextMinor *version.Version, nextMajor *version.Version, err error) {
	nextPatch, nextMinor, nextMajor = nil, nil, nil

	for _, availableVersion := range availableVersions {
		versionCompare, err := availableVersion.Compare(currentClusterVersion.String())
		if err != nil {
			return nil, nil, nil, err
		}
		if versionCompare <= 0 {
			continue
		}
		if currentClusterVersion.Major() == availableVersion.Major() {
			if currentClusterVersion.Minor() == availableVersion.Minor() {
				// Allow to skip patch versions to the latest patch in the interval [Major, Minor]
				nextPatch = availableVersion
			} else if (currentClusterVersion.Minor() + 1) == availableVersion.Minor() {
				// Allow to skip patch versions to the latest patch in the interval [Major, Minor + 1]
				nextMinor = availableVersion
			}
		} else if nextMajor == nil {
			nextMajor = availableVersion
		}
	}

	return nextPatch, nextMinor, nextMajor, nil
}

// NextAvailableVersions return the next patch version available (if any) for
// the current minor version, the next minor version (if any) for the current
// major version, and the next major version (if any)
func NextAvailableVersions(client clientset.Interface) (nextPatch *version.Version, nextMinor *version.Version, nextMajor *version.Version, err error) {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return nil, nil, nil, err
	}
	return nextAvailableVersionsForVersion(currentClusterVersion, kubernetes.AvailableVersions())
}

// UpgradePathWithAvailableVersions returns the list of versions the cluster
// need to go through in order to upgrade to the latest available version in the
// provided list of available versions
func UpgradePathWithAvailableVersions(currentClusterVersion *version.Version, availableVersions []*version.Version) ([]*version.Version, error) {
	upgradePath := []*version.Version{}

	nextPatch, nextMinor, nextMajor, err := nextAvailableVersionsForVersion(currentClusterVersion, availableVersions)
	if err != nil {
		return upgradePath, err
	}

	if nextPatch != nil {
		newPath, err := UpgradePathWithAvailableVersions(nextPatch, availableVersions)
		if err != nil {
			return upgradePath, err
		}
		upgradePath = append(upgradePath, nextPatch)
		upgradePath = append(upgradePath, newPath...)
	} else if nextMinor != nil {
		newPath, err := UpgradePathWithAvailableVersions(nextMinor, availableVersions)
		if err != nil {
			return upgradePath, err
		}
		upgradePath = append(upgradePath, nextMinor)
		upgradePath = append(upgradePath, newPath...)
	} else if nextMajor != nil {
		newPath, err := UpgradePathWithAvailableVersions(nextMajor, availableVersions)
		if err != nil {
			return upgradePath, err
		}
		upgradePath = append(upgradePath, nextMajor)
		upgradePath = append(upgradePath, newPath...)
	}

	return upgradePath, nil
}

// UpgradePath returns the list of versions the cluster needs to go through
// in order to upgrade to the latest available version
func UpgradePath(client clientset.Interface) ([]*version.Version, error) {
	currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
	if err != nil {
		return []*version.Version{}, err
	}
	return UpgradePathWithAvailableVersions(currentClusterVersion, kubernetes.AvailableVersions())
}
