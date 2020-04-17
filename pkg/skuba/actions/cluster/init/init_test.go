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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SUSE/skuba/internal/pkg/skuba/node"
	constants "github.com/SUSE/skuba/pkg/skuba"
)

type TestFilesystemContext struct {
	WorkDirectory    string
	OldWorkDirectory string
}

func TestInitCloudProvider(t *testing.T) {
	tests := []struct {
		name                string
		cloudProvider       string
		providerReadmeFiles []string
	}{
		{
			name:          "init with no cloud provider",
			cloudProvider: "",
			providerReadmeFiles: []string{
				constants.CloudReadmeFile(),
			},
		},
		{
			name:          "init with aws cloud provider",
			cloudProvider: "aws",
			providerReadmeFiles: []string{
				constants.CloudReadmeFile(),
				constants.AWSReadmeFile(),
			},
		},
		{
			name:          "init with opendstack cloud provider",
			cloudProvider: "openstack",
			providerReadmeFiles: []string{
				constants.CloudReadmeFile(),
				constants.OpenstackReadmeFile(),
			},
		},
		{
			name:          "init with vsphere cloud provider",
			cloudProvider: "vsphere",
			providerReadmeFiles: []string{
				constants.CloudReadmeFile(),
				constants.VSphereReadmeFile(),
			},
		},
	}

	clusterName := "testCluster"
	controlPlane := "http://k8s.example.com"
	k8sDesiredVersion := ""
	strictCapDefaults := true
	initConfig := "kubeadm-init.conf"
	joinFiles := []string{
		filepath.Join("kubeadm-join.conf.d", "master.conf.template"),
		filepath.Join("kubeadm-join.conf.d", "worker.conf.template"),
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			initConf, err := NewInitConfiguration(
				clusterName,
				tt.cloudProvider,
				controlPlane,
				k8sDesiredVersion,
				strictCapDefaults)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			ctx, err := switchToTemporaryDirectory()
			if err != nil {
				t.Errorf("unexpect error: %v", err)
			}
			defer switchBackAndCleanFilesystem(ctx)

			if err = Init(initConf); err != nil {
				t.Errorf("unexpect error: %v", err)
			}

			switch tt.cloudProvider {
			case "":
				for _, file := range tt.providerReadmeFiles {
					found, err := doesFileExist(ctx, clusterName, file)
					if err != nil {
						t.Errorf("unexpected error while checking %s presence: %v",
							file, err)
					}
					if found {
						t.Errorf("unexpected file %s found", file)
					}
				}
			default:
				for _, file := range tt.providerReadmeFiles {
					found, err := doesFileExist(ctx, clusterName, file)
					if err != nil {
						t.Errorf("unexpected error while checking %s presence: %v",
							file, err)
					}
					if !found {
						t.Errorf("expected %s to exist", file)
					}
				}
			}

			if err := checkInitConfig(ctx, clusterName, initConfig, tt.cloudProvider); err != nil {
				t.Errorf("%s: %v", initConfig, err)
			}

			for _, f := range joinFiles {
				if err := checkJoinConfig(ctx, clusterName, f, tt.cloudProvider); err != nil {
					t.Errorf("error while inspecting file %s: %v", f, err)
				}
			}
		})
	}
}

// check the init configuration hold inside of `file`. Path to file is built starting
// from the test `ctx` and the `clusterName`, plus the ending name of the `file`.
// `cloud` holds the name of the CPI - leave empty if CPI is disabled.
func checkInitConfig(ctx TestFilesystemContext, clusterName, file, provider string) error {
	expected := len(provider) > 0
	kubeadmInitConfig, err := node.LoadInitConfigurationFromFile(
		filepath.Join(ctx.WorkDirectory, clusterName, file))
	if err != nil {
		return fmt.Errorf("unexpected error: %v", err)
	}

	value, found := kubeadmInitConfig.NodeRegistration.KubeletExtraArgs["cloud-provider"]
	if err := checkMapEntry(provider, value, expected, found); err != nil {
		return fmt.Errorf("nodeRegistration  - %v", err)
	}

	value, found = kubeadmInitConfig.ClusterConfiguration.APIServer.ExtraArgs["cloud-provider"]
	if err := checkMapEntry(provider, value, expected, found); err != nil {
		return fmt.Errorf("apiServer.extraArgs - %v", err)
	}

	value, found = kubeadmInitConfig.ClusterConfiguration.ControllerManager.ExtraArgs["cloud-provider"]
	if err := checkMapEntry(provider, value, expected, found); err != nil {
		return fmt.Errorf("controllerManager.extraArgs - %v", err)
	}

	if provider == "aws" {
		value, found = kubeadmInitConfig.ClusterConfiguration.ControllerManager.ExtraArgs["allocate-node-cidrs"]
		if err := checkMapEntry("false", value, expected, found); err != nil {
			return fmt.Errorf("controllerManager.extraArgs - %v", err)
		}
	}

	return nil
}

// check the join configuration hold inside of `file`. Path to file is built starting
// from the test `ctx` and the `clusterName`, plus the ending name of the `file`.
// `cloud` holds the name of the CPI - leave empty if CPI is disabled.
func checkJoinConfig(ctx TestFilesystemContext, clusterName, file, provider string) error {
	expected := provider != ""
	kubeadmJoinConfig, err := node.LoadJoinConfigurationFromFile(
		filepath.Join(ctx.WorkDirectory, clusterName, file))
	if err != nil {
		return err
	}

	value, found := kubeadmJoinConfig.NodeRegistration.KubeletExtraArgs["cloud-provider"]
	return checkMapEntry(provider, value, expected, found)
}

// Compares the `value` found inside of a map against its `expectedValue`
// return an error when the check fails
func checkMapEntry(expectedValue, value string, expected, found bool) error {
	if found {
		if !expected {
			return fmt.Errorf("cloud-provider integration is accidentally enabled")
		}
		if value != expectedValue {
			return fmt.Errorf("wrong cloud-provider value, expected %s, got %s", expectedValue, value)
		}
	} else if expected {
		return fmt.Errorf("nodeRegistration - couldn't find cloud-provider value")
	}

	return nil
}

// Create a new temporary directory and switches into it
func switchToTemporaryDirectory() (TestFilesystemContext, error) {
	var ctx TestFilesystemContext

	tmp, err := os.Getwd()
	if err != nil {
		return ctx, err
	}
	ctx.OldWorkDirectory = tmp

	tmp, err = ioutil.TempDir("", "skuba-cluster-init-test")
	if err != nil {
		return ctx, err
	}
	ctx.WorkDirectory = tmp

	if err = os.Chdir(ctx.WorkDirectory); err != nil {
		return ctx, err
	}

	return ctx, nil
}

// Remove the temporary directory previously created and
// moves back to the work directory used at the beginning of
// the test
func switchBackAndCleanFilesystem(ctx TestFilesystemContext) {
	if err := os.Chdir(ctx.OldWorkDirectory); err != nil {
		fmt.Printf("Cannot chdir back to original work directory: %v", err)
	}
	if err := os.RemoveAll(ctx.WorkDirectory); err != nil {
		fmt.Printf("Cannot remove temporary directory: %v", err)
	}
}

// Checks if the specified file exists. Raises an error if something
// unexpected happens.
func doesFileExist(ctx TestFilesystemContext, clusterName, file string) (bool, error) {
	fullpath := filepath.Join(ctx.WorkDirectory, clusterName, file)
	if _, err := os.Stat(fullpath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
