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
	"fmt"
	"io/ioutil"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubeadm"
	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/SUSE/skuba/pkg/skuba/actions/cluster/upgrade"
	"github.com/rikatz/kubepug/lib"
	"github.com/rikatz/kubepug/pkg/formatter"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog"
)

type checkOptions struct {
	K8sVersion      string
	ForceDownload   bool
	APIWalk         bool
	SwaggerDir      string
	ShowDescription bool
}

var (
	format    string
	filename  string
	inputFile string
)

// newUpgradeCheckCmd creates a new `skuba check` cobra command
func newUpgradeCheckCmd() *cobra.Command {
	checkOptions := &checkOptions{}

	cmd := &cobra.Command{
		Use:   "check kubernetes-version=<version> swaggerDir=<directory> --api-walk=<true|fasle>",
		Short: "Print Upgrade Check information",
		RunE: func(cmd *cobra.Command, args []string) error {
			kubernetesConfigFlags := genericclioptions.NewConfigFlags(true)
			kubeAdminFile := skuba.KubeConfigAdminFile()
			kubernetesConfigFlags.KubeConfig = &kubeAdminFile

			config := lib.Config{
				K8sVersion:      checkOptions.K8sVersion,
				ForceDownload:   checkOptions.ForceDownload,
				APIWalk:         checkOptions.APIWalk,
				SwaggerDir:      checkOptions.SwaggerDir,
				ShowDescription: checkOptions.ShowDescription,
				ConfigFlags:     kubernetesConfigFlags,
				Input:           inputFile,
			}

			if config.K8sVersion == "" {
				client, err := kubernetes.GetAdminClientSet()
				if err != nil {
					return err
				}
				currentClusterVersion, err := kubeadm.GetCurrentClusterVersion(client)
				if err != nil {
					return err
				}
				availableVersions := kubernetes.AvailableVersions()
				upgradePath, err := upgrade.CalculateUpgradePath(currentClusterVersion, availableVersions)
				if err != nil {
					return err
				}

				var nextClusterVersion *version.Version
				if len(upgradePath) > 0 {
					nextClusterVersion = upgradePath[0]
				} else {
					klog.Warning("Already on the latest version, nothing to check.\nFor a specific version use the `kubernetes-version` flag.")
					return nil
				}
				config.K8sVersion = fmt.Sprintf("v%s", nextClusterVersion.String())
			}

			klog.Infof("Starting Kubepug with configs: %+v", config)
			kubepug := lib.NewKubepug(config)

			result, err := kubepug.GetDeprecated()
			if err != nil {
				return err
			}

			klog.Info("Starting deprecated objects printing")
			formatter := formatter.NewFormatter(format)
			bytes, err := formatter.Output(*result)
			if err != nil {
				return err
			}

			if filename != "" {
				err = ioutil.WriteFile(filename, bytes, 0644)
				if err != nil {
					return err
				}
			} else {
				fmt.Printf("%s", string(bytes))
			}

			return nil
		},
	}
	cmd.PersistentFlags().BoolVar(&checkOptions.APIWalk, "api-walk", false, "Whether to walk in the whole API, checking if all objects type still exists in the current swagger.json. May be IO intensive to APIServer.")
	cmd.PersistentFlags().BoolVar(&checkOptions.ShowDescription, "description", true, "Wether to show the description of the deprecated object. The description may contain the solution for the deprecation.")
	cmd.PersistentFlags().StringVar(&checkOptions.K8sVersion, "kubernetes-version", "", "Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects.")
	cmd.PersistentFlags().StringVar(&checkOptions.SwaggerDir, "swagger-dir", "", "Where to keep swagger.json downloaded file. If not provided will use the system temporary directory")
	cmd.PersistentFlags().BoolVar(&checkOptions.ForceDownload, "force-download", false, "Wether to force the download of a new swagger.json file even if one exists.")
	cmd.PersistentFlags().StringVar(&format, "format", "plain", "Format in which the list will be displayed [stdout, plain, json, yaml]")
	cmd.PersistentFlags().StringVar(&filename, "filename", "", "Name of the file the results will be saved to, if empty it will display to stdout")
	cmd.PersistentFlags().StringVar(&inputFile, "input-file", "", "Location of a file or directory containing k8s manifests to be analized")
	return cmd
}
