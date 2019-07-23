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

package cloud

import (
	"os"

	"k8s.io/klog"

	"github.com/SUSE/skuba/pkg/skuba"
)

// HasCloudIntegration checks if the cluster has a cloud integration
func HasCloudIntegration() bool {
	if _, err := os.Stat(skuba.OpenstackCloudConfFile()); err == nil {
		return true
	} else if _, err := os.Stat(skuba.OpenstackCloudConfTemplateFile()); err == nil {
		klog.Fatalf("%q file exists, but %q file does not. Please create this file with the expected contents to enable cloud integration", skuba.OpenstackCloudConfTemplateFile(), skuba.OpenstackCloudConfFile())
	}
	return false
}

// ConfigHasRestrictedPermissions returns true if
func ConfigHasRestrictedPermissions(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	mode := info.Mode()
	var mask os.FileMode = 077
	if (mode & mask) > 0 {
		return false
	}
	return true
}
