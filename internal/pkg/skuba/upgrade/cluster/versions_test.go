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
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/version"
)

func TestNextAvailableVersionsForVersion(t *testing.T) {
	var versions = []struct {
		currentClusterVersion             *version.Version
		availableVersions                 []*version.Version
		expectedPatchNextAvailableVersion *version.Version
		expectedMinorNextAvailableVersion *version.Version
		expectedMajorNextAvailableVersion *version.Version
	}{
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.13.9"),
			},
			expectedPatchNextAvailableVersion: nil,
			expectedMinorNextAvailableVersion: nil,
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.0"),
			},
			expectedPatchNextAvailableVersion: nil,
			expectedMinorNextAvailableVersion: nil,
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.1"),
			expectedMinorNextAvailableVersion: nil,
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.2"),
			expectedMinorNextAvailableVersion: nil,
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.2"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.0"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
			},
			expectedPatchNextAvailableVersion: nil,
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.0"),
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.2"),
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.0"),
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.15.0"),
			},
			expectedPatchNextAvailableVersion: nil,
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.0"),
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.15.2"),
			},
			expectedPatchNextAvailableVersion: nil,
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.2"),
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.15.2"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.2"),
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.2"),
			expectedMajorNextAvailableVersion: nil,
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v2.0.0"),
			},
			expectedPatchNextAvailableVersion: nil,
			expectedMinorNextAvailableVersion: nil,
			expectedMajorNextAvailableVersion: version.MustParseSemantic("v2.0.0"),
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.15.2"),
				version.MustParseSemantic("v2.0.0"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.2"),
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.2"),
			expectedMajorNextAvailableVersion: version.MustParseSemantic("v2.0.0"),
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.15.2"),
				version.MustParseSemantic("v2.0.0"),
				version.MustParseSemantic("v2.1.0"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.2"),
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.2"),
			expectedMajorNextAvailableVersion: version.MustParseSemantic("v2.0.0"),
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.15.2"),
				version.MustParseSemantic("v2.0.0"),
				version.MustParseSemantic("v2.1.0"),
				version.MustParseSemantic("v3.0.0"),
				version.MustParseSemantic("v3.1.0"),
			},
			expectedPatchNextAvailableVersion: version.MustParseSemantic("v1.14.2"),
			expectedMinorNextAvailableVersion: version.MustParseSemantic("v1.15.2"),
			expectedMajorNextAvailableVersion: version.MustParseSemantic("v2.0.0"),
		},
	}
	for _, tt := range versions {
		t.Run(fmt.Sprintf("Upgrade %s with %v available versions", tt.currentClusterVersion.String(), tt.availableVersions), func(t *testing.T) {
			nextPatch, nextMinor, nextMajor, err := nextAvailableVersionsForVersion(tt.currentClusterVersion, tt.availableVersions)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(nextPatch, tt.expectedPatchNextAvailableVersion) {
				t.Errorf("next patch version (%v) does not match expected version (%v)", nextPatch, tt.expectedPatchNextAvailableVersion)
			}
			if !reflect.DeepEqual(nextMinor, tt.expectedMinorNextAvailableVersion) {
				t.Errorf("next minor version (%v) does not match expected version (%v)", nextMinor, tt.expectedMinorNextAvailableVersion)
			}
			if !reflect.DeepEqual(nextMajor, tt.expectedMajorNextAvailableVersion) {
				t.Errorf("next major version (%v) does not match expected version (%v)", nextMajor, tt.expectedMajorNextAvailableVersion)
			}
		})
	}
}

func TestUpgradePathWithAvailableVersions(t *testing.T) {
	var versions = []struct {
		currentClusterVersion *version.Version
		availableVersions     []*version.Version
		expectedUpgradePath   []*version.Version
	}{
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.13.9"),
			},
			expectedUpgradePath: []*version.Version{},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.0"),
			},
			expectedUpgradePath: []*version.Version{},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.14.2"),
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.2"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.0"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.15.0"),
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.16.0"),
				version.MustParseSemantic("v1.16.1"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.16.1"),
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.16.0"),
				version.MustParseSemantic("v1.16.1"),
				version.MustParseSemantic("v1.16.2"),
				version.MustParseSemantic("v1.17.0"),
				version.MustParseSemantic("v1.17.1"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.16.2"),
				version.MustParseSemantic("v1.17.1"),
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			availableVersions: []*version.Version{
				version.MustParseSemantic("v1.14.1"),
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.0"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.16.0"),
				version.MustParseSemantic("v1.16.1"),
				version.MustParseSemantic("v1.17.0"),
				version.MustParseSemantic("v1.17.1"),
				version.MustParseSemantic("v2.0.0"),
				version.MustParseSemantic("v2.1.0"),
				version.MustParseSemantic("v2.1.1"),
				version.MustParseSemantic("v2.2.0"),
				version.MustParseSemantic("v2.2.1"),
			},
			expectedUpgradePath: []*version.Version{
				version.MustParseSemantic("v1.14.2"),
				version.MustParseSemantic("v1.15.1"),
				version.MustParseSemantic("v1.16.1"),
				version.MustParseSemantic("v1.17.1"),
				version.MustParseSemantic("v2.0.0"),
				version.MustParseSemantic("v2.1.1"),
				version.MustParseSemantic("v2.2.1"),
			},
		},
	}
	for _, tt := range versions {
		t.Run(fmt.Sprintf("Upgrade path from %s with %v available versions", tt.currentClusterVersion.String(), tt.availableVersions), func(t *testing.T) {
			upgradePath, err := UpgradePathWithAvailableVersions(tt.currentClusterVersion, tt.availableVersions)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(upgradePath, tt.expectedUpgradePath) {
				t.Errorf("upgrade path (%v) does not match expected upgrade path (%v)", upgradePath, tt.expectedUpgradePath)
			}
		})
	}
}
