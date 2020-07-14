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

package cluster

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SUSE/skuba/pkg/skuba"
)

func TestCriMigrate(t *testing.T) {
	// First of all, create the new working directory as it will be taken by
	// skuba.

	dir, err := ioutil.TempDir("/tmp", "skuba-cri-test")
	if err != nil {
		t.Fatalf("Could not initialize test: %v", err)
	}
	defer os.RemoveAll(dir)
	destPath := filepath.Join(dir, "addons", "cri")
	err = os.MkdirAll(destPath, os.ModePerm)
	if err != nil {
		t.Fatalf("Could not initialize test: %v", err)
	}

	// Move the old `default_flags` we have in `testdata` into the new temporary
	// working directory.

	b, err := ioutil.ReadFile(filepath.Join("testdata", "addons", "cri", "default_flags"))
	if err != nil {
		t.Fatalf("Could not initialize test: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(destPath, "default_flags"), b, 0644)
	if err != nil {
		t.Fatalf("Could not initialize test: %v", err)
	}

	// Now it's safe to change the working directory.

	wd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(wd)
	}()
	_ = os.Chdir(dir)

	// Test: before calling the function, check that the expected file exists
	// there. Then, upon calling the tested function, we should have the new
	// configuration schema (with `default_capabilities` as given in the old
	// configuration).

	_, err = os.Stat(skuba.CriDockerDefaultsConfFile())
	if err != nil {
		t.Fatalf("File should exist, got '%v' instead", err)
	}

	err = CriMigrate()
	if err != nil {
		t.Fatalf("Function should've run correctly")
	}

	_, err = os.Stat(skuba.CriDockerDefaultsConfFile())
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("File should not exist, got '%v' instead", err)
	}

	b, err = ioutil.ReadFile(filepath.Join(destPath, "conf.d", "01-caasp.conf"))
	if err != nil {
		t.Fatalf("File should be readable, got '%v' instead", err)
	}
	if !strings.Contains(string(b), "default_capabilities") {
		t.Fatalf("New file should include default capabilities")
	}
}
