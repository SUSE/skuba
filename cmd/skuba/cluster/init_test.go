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
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	cluster "github.com/SUSE/skuba/pkg/skuba/actions/cluster/init"
)

func TestParseCrioRegistryConfiguration(t *testing.T) {
	type args struct {
		config string
	}
	tests := []struct {
		name string
		args args
		want *cluster.CrioMirrorConfiguration
	}{
		{"Parse a valid registry configuration", args{config: "registry.suse.com:test.registry.com:true"}, &cluster.CrioMirrorConfiguration{SourceRegistry: "registry.suse.com", MirrorRegistry: "test.registry.com", Insecure: true}},
		{"Parse an invalid registry configuration", args{config: "::"}, &cluster.CrioMirrorConfiguration{SourceRegistry: "", MirrorRegistry: "", Insecure: false}},
		{"Parse an registry configuration with security part", args{config: "registry.suse.com:test.registry.com:"}, &cluster.CrioMirrorConfiguration{SourceRegistry: "registry.suse.com", MirrorRegistry: "test.registry.com", Insecure: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCrioRegistryConfiguration(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCrioRegistryConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

type FileCheck struct {
	RelativePath string
	Contents     string
	Exists       bool
}

func (fc FileCheck) FileState(e error) bool {
	if fc.Exists {
		return os.IsNotExist(e)
	} else {
		return e == nil
	}
}

func TestInitCmdCreatesRegistryMirrorFiles(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		files []FileCheck
	}{
		{
			"Run init with mirror setup",
			[]string{"cluster", "--control-plane", "1.2.3.4", "--registry-mirror", "registry.suse.com:my.registry.com:false"},
			[]FileCheck{FileCheck{"addons/containers/registries.conf", "", true}},
		},
		{
			"Run init without mirror setup",
			[]string{"cluster", "--control-plane", "1.2.3.4"},
			[]FileCheck{FileCheck{"addons/containers/registries.conf", "", false}},
		},
	}
	for _, tt := range tests {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		clusterName := tt.args[0]
		os.Args = tt.args
		dir, err := ioutil.TempDir("/tmp", "skuba-init-test")
		if err != nil {
			t.Fail()
		}
		os.Chdir(dir)
		command := NewInitCmd()
		command.Run(command, tt.args)
		for _, relPath := range tt.files {
			fullPath := path.Join(dir, clusterName, relPath.RelativePath)
			if _, err := os.Stat(fullPath); relPath.FileState(err) {
				t.Errorf("%s incorrect state - expected: %v got %v.", fullPath, relPath.Exists, relPath.FileState(err))
			}
		}

	}
}
