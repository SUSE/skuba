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

package addons

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

func TestAddonLegacyManifestMigration(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("unable to get current directory: %v", err)
		return
	}

	defer func() {
		// removes rendered addon folder
		dir := filepath.Join(pwd, "addons")
		if f, err := os.Stat(dir); !os.IsNotExist(err) && f.IsDir() {
			if err := os.RemoveAll(dir); err != nil {
				t.Errorf("unable to remove rendered addon folder: %v", err)
				return
			}
		}
	}()

	// create legacy addons manifest folder
	if err := os.Mkdir(filepath.Join(pwd, skubaconstants.AddonsDir()), 0700); err != nil {
		t.Errorf("unable to create directory %s: %v", skubaconstants.AddonsDir(), err)
		return
	}
	for _, addon := range Addons {
		addonDir := filepath.Join(pwd, addon.addonDir())
		if err := os.Mkdir(addonDir, 0700); err != nil {
			t.Errorf("unable to create directory %s: %v", addonDir, err)
			return
		}
		lagacyManifestPath := addon.legacyManifestPath(addonDir)
		if err := ioutil.WriteFile(lagacyManifestPath, []byte(""), 0600); err != nil {
			t.Errorf("unable to write legacy addon manifest: %v", err)
			return
		}
	}

	addonConfiguration := AddonConfiguration{
		ClusterVersion: kubernetes.LatestVersion(),
		ControlPlane:   "unit.test",
		ClusterName:    "unit-test",
	}
	// render new addons manifest folder
	for _, addon := range Addons {
		if err := addon.Write(addonConfiguration); err != nil {
			t.Errorf("expected no error, but got error: %v", err)
			return
		}
	}

	// check the legacy addons manifest gone
	for _, addon := range Addons {
		lagacyManifestPath := addon.legacyManifestPath(filepath.Join(pwd, addon.addonDir()))
		if _, err := os.Stat(lagacyManifestPath); !os.IsNotExist(err) {
			t.Error("expected legacy manifest not exists")
			return
		}
	}
}
