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
)

type TestFilesystemContext struct {
	WorkDirectory    string
	OldWorkDirectory string
}

func TestInitAWSCloudProviderEnabled(t *testing.T) {
	clusterName := "testCluster"

	initConf, err := NewInitConfiguration(
		clusterName,
		"aws",
		"http://k8s.example.com",
		"",
		true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	ctx, err := switchToTemporaryDirectory()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	defer switchBackAndCleanFilesystem(ctx)

	if err = Init(initConf); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := checkInitConfig(ctx, clusterName, "kubeadm-init.conf", "aws"); err != nil {
		t.Errorf("kubeadm-init.conf: %v", err)
	}

	joinFiles := []string{
		filepath.Join("kubeadm-join.conf.d", "master.conf.template"),
		filepath.Join("kubeadm-join.conf.d", "worker.conf.template"),
	}
	for _, f := range joinFiles {
		if err := checkJoinConfig(ctx, clusterName, f, "aws"); err != nil {
			t.Errorf("Error while inspecting file %s: %v", f, err)
		}
	}
}

func TestInitAWSCloudProviderDisabled(t *testing.T) {
	clusterName := "testCluster"

	initConf, err := NewInitConfiguration(
		clusterName,
		"",
		"http://k8s.example.com",
		"",
		true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	ctx, err := switchToTemporaryDirectory()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	defer switchBackAndCleanFilesystem(ctx)

	if err = Init(initConf); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := checkInitConfig(ctx, clusterName, "kubeadm-init.conf", ""); err != nil {
		t.Errorf("kubeadm-init.conf: %v", err)
	}

	joinFiles := []string{
		filepath.Join("kubeadm-join.conf.d", "master.conf.template"),
		filepath.Join("kubeadm-join.conf.d", "worker.conf.template"),
	}
	for _, f := range joinFiles {
		if err := checkJoinConfig(ctx, clusterName, f, ""); err != nil {
			t.Errorf("Error while inspecting file %s: %v", f, err)
		}
	}
}

// check the init configuration hold inside of `file`. Path to file is built starting
// from the test `ctx` and the `clusterName`, plus the ending name of the `file`.
// `cloud` holds the name of the CPI - leave empty if CPI is disabled.
func checkInitConfig(ctx TestFilesystemContext, clusterName, file, cloud string) error {
	expected := cloud != ""
	kubeadmInitConfig, err := node.LoadInitConfigurationFromFile(
		filepath.Join(ctx.WorkDirectory, clusterName, file))
	if err != nil {
		return fmt.Errorf("Unexpected error: %v", err)
	}

	value, found := kubeadmInitConfig.NodeRegistration.KubeletExtraArgs["cloud-provider"]
	if err := checkMapEntry("aws", value, expected, found); err != nil {
		return fmt.Errorf("nodeRegistration - %v", err)
	}

	value, found = kubeadmInitConfig.ClusterConfiguration.APIServer.ExtraArgs["cloud-provider"]
	if err := checkMapEntry("aws", value, expected, found); err != nil {
		return fmt.Errorf("APIServer extraArgs - %v", err)
	}

	value, found = kubeadmInitConfig.ClusterConfiguration.ControllerManager.ExtraArgs["cloud-provider"]
	if err := checkMapEntry("aws", value, expected, found); err != nil {
		return fmt.Errorf("ControllerManager extraArgs - %v", err)
	}

	value, found = kubeadmInitConfig.ClusterConfiguration.ControllerManager.ExtraArgs["allocate-node-cidrs"]
	if err := checkMapEntry("false", value, expected, found); err != nil {
		return fmt.Errorf("ControllerManager extraArgs - %v", err)
	}

	return nil
}

// check the join configuration hold inside of `file`. Path to file is built starting
// from the test `ctx` and the `clusterName`, plus the ending name of the `file`.
// `cloud` holds the name of the CPI - leave empty if CPI is disabled.
func checkJoinConfig(ctx TestFilesystemContext, clusterName, file, cloud string) error {
	expected := cloud != ""
	kubeadmJoinConfig, err := node.LoadJoinConfigurationFromFile(
		filepath.Join(ctx.WorkDirectory, clusterName, file))
	if err != nil {
		return err
	}

	value, found := kubeadmJoinConfig.NodeRegistration.KubeletExtraArgs["cloud-provider"]
	return checkMapEntry("aws", value, expected, found)
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
