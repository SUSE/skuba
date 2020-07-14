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
	"strings"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	clusterinit "github.com/SUSE/skuba/pkg/skuba/actions/cluster/init"
)

// CriMigrate migrates the old configuration of cri-o < 1.17 to the new format.
func CriMigrate() error {
	_, err := os.Stat(skuba.CriDockerDefaultsConfFile())
	if os.IsNotExist(err) {
		return nil
	}

	if err := criGenerateLocalConfiguration(); err != nil {
		return err
	}

	if err := criRemoveLocalOldFile(); err != nil {
		return err
	}
	return nil
}

func criGenerateLocalConfiguration() error {
	cfg := clusterinit.InitConfiguration{
		PauseImage:        kubernetes.ComponentContainerImageForClusterVersion(kubernetes.Pause, kubernetes.LatestVersion()),
		StrictCapDefaults: !criHadStrictCapDefaults(),
	}

	files := clusterinit.CriScaffoldFiles["criconfig"]
	for _, file := range files {
		// Note well: this code will remove the whole local cri configuration
		// directory if something goes wrong. This is done so to avoid weird
		// intermediary states in case a file presented a problem. Ideally this
		// shouldn't happen (it's more of a all or nothing scenario), but take
		// it into account if you want to reuse this code.
		if err := clusterinit.WriteScaffoldFile(file, cfg); err != nil {
			_ = os.RemoveAll(skuba.CriConfDir())
			return err
		}
	}
	return nil
}

func criHadStrictCapDefaults() bool {
	data, err := ioutil.ReadFile(skuba.CriDockerDefaultsConfFile())
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "--default-capabilities")
}

func criRemoveLocalOldFile() error {
	_, err := os.Stat(skuba.CriDockerDefaultsConfFile())
	if os.IsNotExist(err) {
		return nil
	}
	return os.Remove(skuba.CriDockerDefaultsConfFile())
}
