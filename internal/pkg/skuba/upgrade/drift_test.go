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

package upgrade

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/version"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
)

func TestDriftedNodesWithVersions(t *testing.T) {
	var versions = []struct {
		currentClusterVersion *version.Version
		nodesVersionInfo      []kubernetes.NodeVersionInfo
		expectedDriftedNodes  []kubernetes.NodeVersionInfo
	}{
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			nodesVersionInfo: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "a-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
			},
			expectedDriftedNodes: []kubernetes.NodeVersionInfo{},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			nodesVersionInfo: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "a-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
				{
					Nodename:       "another-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
			},
			expectedDriftedNodes: []kubernetes.NodeVersionInfo{},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.5"),
			nodesVersionInfo: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "a-node",
					KubeletVersion: version.MustParseSemantic("v1.14.5"),
				},
				{
					Nodename:       "slightly-drifted-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
			},
			expectedDriftedNodes: []kubernetes.NodeVersionInfo{},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			nodesVersionInfo: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "a-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
				{
					Nodename:       "drifted-node",
					KubeletVersion: version.MustParseSemantic("v1.13.0"),
				},
			},
			expectedDriftedNodes: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "drifted-node",
					KubeletVersion: version.MustParseSemantic("v1.13.0"),
				},
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.5"),
			nodesVersionInfo: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "a-node",
					KubeletVersion: version.MustParseSemantic("v1.14.5"),
				},
				{
					Nodename:       "slightly-drifted-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
				{
					Nodename:       "drifted-node",
					KubeletVersion: version.MustParseSemantic("v1.13.0"),
				},
			},
			expectedDriftedNodes: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "drifted-node",
					KubeletVersion: version.MustParseSemantic("v1.13.0"),
				},
			},
		},
		{
			currentClusterVersion: version.MustParseSemantic("v1.14.0"),
			nodesVersionInfo: []kubernetes.NodeVersionInfo{
				{
					Nodename:       "a-node",
					KubeletVersion: version.MustParseSemantic("v1.14.0"),
				},
				{
					Nodename:       "drifted-unschedulable-node",
					KubeletVersion: version.MustParseSemantic("v1.13.0"),
					Unschedulable:  true,
				},
			},
			expectedDriftedNodes: []kubernetes.NodeVersionInfo{},
		},
	}
	for _, tt := range versions {
		t.Run(fmt.Sprintf("Test upgrade feasible for %s with %v nodes", tt.currentClusterVersion.String(), tt.nodesVersionInfo), func(t *testing.T) {
			driftedNodes := driftedNodesWithVersions(tt.currentClusterVersion, tt.nodesVersionInfo)
			if !reflect.DeepEqual(driftedNodes, tt.expectedDriftedNodes) {
				t.Errorf("reported drifted nodes (%v) do not match expectation (%v)", driftedNodes, tt.expectedDriftedNodes)
			}
		})
	}
}
