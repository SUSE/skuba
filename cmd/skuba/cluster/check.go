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
	"os"

	clientset "github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/rikatz/kubepug/lib"
	"github.com/rikatz/kubepug/pkg/formatter"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	errorOnDeprecated bool
	errorOnDeleted    bool
)

// NewCheckCmd creates a new `skuba check` cobra command
func NewCheckCmd() *cobra.Command {
	checkOptions := &checkOptions{}

	cmd := &cobra.Command{
		Use:   "check k8s-version=<version> swaggerDir=<directory> --api-walk=<true|fasle>",
		Short: "Print Check information",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, cnf, err := clientset.GetAdminClientSetWithConfig()
			if err != nil {
				klog.Errorf("unable to get admin client set: %s", err)
				os.Exit(1)
			}

			kubernetesConfigFlags := genericclioptions.NewConfigFlags(true)
			kubernetesConfigFlags.APIServer = &cnf.Host
			kubernetesConfigFlags.BearerToken = &cnf.BearerToken
			kubernetesConfigFlags.CAFile = &cnf.CAFile
			kubernetesConfigFlags.CertFile = &cnf.CertFile
			kubernetesConfigFlags.KeyFile = &cnf.KeyFile
			kubernetesConfigFlags.Username = &cnf.Username
			kubernetesConfigFlags.Insecure = &cnf.Insecure
			kubernetesConfigFlags.Password = &cnf.Password
			kubernetesConfigFlags.TLSServerName = &cnf.TLSClientConfig.ServerName
			timeout := cnf.Timeout.String()
			kubernetesConfigFlags.Timeout = &timeout

			config := lib.Config{
				K8sVersion:      checkOptions.K8sVersion,
				ForceDownload:   checkOptions.ForceDownload,
				APIWalk:         checkOptions.APIWalk,
				SwaggerDir:      checkOptions.SwaggerDir,
				ShowDescription: checkOptions.ShowDescription,
				ConfigFlags:     kubernetesConfigFlags,
			}

			lvl, err := logrus.ParseLevel("info")
			if err != nil {
				return err
			}
			logrus.SetLevel(lvl)

			log.SetFormatter(&log.TextFormatter{
				DisableColors: true,
				FullTimestamp: true,
			})
			if lvl == log.DebugLevel {
				log.SetReportCaller(true)
			}

			log.Debugf("Starting Kubepug with configs: %+v", config)
			kubepug := lib.NewKubepug(config)

			result, err := kubepug.GetDeprecated()
			if err != nil {
				return err
			}

			log.Debug("Starting deprecated objects printing")
			formatter := formatter.NewFormatter("plain")
			bytes, err := formatter.Output(*result)
			if err != nil {
				return err
			}

			fmt.Printf("%s", string(bytes))

			if (errorOnDeleted && len(result.DeletedAPIs) > 0) || (errorOnDeprecated && len(result.DeprecatedAPIs) > 0) {
				return fmt.Errorf("found %d Deleted APIs and %d Deprecated APIs", len(result.DeletedAPIs), len(result.DeprecatedAPIs))
			}
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}
	cmd.PersistentFlags().BoolVar(&checkOptions.APIWalk, "api-walk", true, "Wether to walk in the whole API, checking if all objects type still exists in the current swagger.json. May be IO intensive to APIServer. Defaults to true")
	cmd.PersistentFlags().BoolVar(&checkOptions.ShowDescription, "description", true, "Wether to show the description of the deprecated object. The description may contain the solution for the deprecation. Defaults to true")
	cmd.PersistentFlags().StringVar(&checkOptions.K8sVersion, "k8s-version", "master", "Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master")
	cmd.PersistentFlags().StringVar(&checkOptions.SwaggerDir, "swagger-dir", "", "Where to keep swagger.json downloaded file. If not provided will use the system temporary directory")
	cmd.PersistentFlags().BoolVar(&checkOptions.ForceDownload, "force-download", false, "Wether to force the download of a new swagger.json file even if one exists. Defaults to false")
	return cmd
}
