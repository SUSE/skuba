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

package cluster

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"k8s.io/klog"
)

type InitConfiguration struct {
	ClusterName  string
	ControlPlane string
}

// Init creates a cluster definition scaffold in the local machine, in the current
// folder, at a directory named after ClusterName provided in the InitConfiguration
// parameter
//
// FIXME: being this a part of the go API accept the toplevel directory instead of
//        using the PWD
// FIXME: error handling with `github.com/pkg/errors`; return errors
func Init(initConfiguration InitConfiguration) {
	if _, err := os.Stat(initConfiguration.ClusterName); err == nil {
		klog.Fatalf("cluster configuration directory %q already exists\n", initConfiguration.ClusterName)
	}
	if err := os.MkdirAll(initConfiguration.ClusterName, 0700); err != nil {
		klog.Fatalf("could not create cluster directory %q: %v\n", initConfiguration.ClusterName, err)
	}
	if err := os.Chdir(initConfiguration.ClusterName); err != nil {
		klog.Fatalf("could not change to cluster directory %q: %v\n", initConfiguration.ClusterName, err)
	}
	for _, file := range scaffoldFiles {
		filePath, _ := filepath.Split(file.Location)
		if filePath != "" {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				klog.Fatalf("could not create directory %q: %v\n", filePath, err)
			}
		}
		f, err := os.Create(file.Location)
		if err != nil {
			klog.Fatalf("could not create file %q: %v\n", file.Location, err)
		}
		if file.DoNotRender {
			f.WriteString(file.Content)
		} else {
			f.WriteString(renderTemplate(file.Content, initConfiguration))
		}
		f.Close()
	}

	if currentDir, err := os.Getwd(); err != nil {
		klog.Fatalf("could not get current directory %s\n", err)
	} else {
		fmt.Printf("[init] configuration files written to %s\n", currentDir)
	}
}

func renderTemplate(templateContents string, initConfiguration InitConfiguration) string {
	template, err := template.New("").Parse(templateContents)
	if err != nil {
		klog.Fatal("could not parse template")
	}
	var rendered bytes.Buffer
	if err := template.Execute(&rendered, initConfiguration); err != nil {
		klog.Fatal("could not render configuration")
	}
	return rendered.String()
}
