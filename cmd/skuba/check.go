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

package main

import (
	kpug "github.com/rikatz/kubepug/lib"
	"github.com/spf13/cobra"
)

type initOptions struct {
	K8sVersion      string
	ForceDownload   bool
	APIWalk         bool
	SwaggerDir      string
	ShowDescription bool
}

// NewCheckCmd creates a new `skuba check` cobra command
func NewCheckCmd() *cobra.Command {
	initOptions := initOptions{}

	cmd := &cobra.Command{
		Use:   "check k8s-version=<version> swaggerDir=<directory> --api-walk=<true|fasle>",
		Short: "Print Check information",
		Run:   run,
		Args:  cobra.MinArgs(1),
	}
	cmd.PersistentFlags().BoolVar(&initOptions.APIWalk, "api-walk", true, "Wether to walk in the whole API, checking if all objects type still exists in the current swagger.json. May be IO intensive to APIServer. Defaults to true")
	cmd.PersistentFlags().BoolVar(&initOptions.ShowDescription, "description", true, "Wether to show the description of the deprecated object. The description may contain the solution for the deprecation. Defaults to true")
	cmd.PersistentFlags().StringVar(&initOptions.K8sVersion, "k8s-version", "master", "Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master")
	cmd.PersistentFlags().StringVar(&initOptions.SwaggerDir, "swagger-dir", "", "Where to keep swagger.json downloaded file. If not provided will use the system temporary directory")
	cmd.PersistentFlags().BoolVar(&initOptions.ForceDownload, "force-download", false, "Wether to force the download of a new swagger.json file even if one exists. Defaults to false")
	return cmd
}

func run(cmd *cobra.Command, args []string) {
	config := kpug.Config{
		K8sVersion:      initOptions.K8sVersion,
		ForceDownload:   initOptions.ForceDownload,
		APIWalk:         initOptions.APIWalk,
		SwaggerDir:      initOptions.SwaggerDir,
		ShowDescription: initOptions.ShowDescription,
	}

	kubepug := kpug.NewKubepug(config)

	result, err := kubepug.GetDeprecated()
	if err != nil {
		return err
	}
}
