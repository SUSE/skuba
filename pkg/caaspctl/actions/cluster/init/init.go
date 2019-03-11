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
	"log"
	"os"
	"path/filepath"
	"text/template"
)

type InitConfiguration struct {
	ClusterName  string
	ControlPlane string
}

func Init(initConfiguration InitConfiguration) {
	if err := os.MkdirAll(initConfiguration.ClusterName, 0700); err != nil {
		log.Fatalf("could not create directory %s\n", initConfiguration.ClusterName)
	}
	if err := os.Chdir(initConfiguration.ClusterName); err != nil {
		log.Fatalf("could not change to directory %s\n", initConfiguration.ClusterName)
	}
	for _, file := range scaffoldFiles {
		filePath, _ := filepath.Split(file.Location)
		if filePath != "" {
			if err := os.MkdirAll(filePath, 0700); err != nil {
				log.Fatalf("could not create directory %s\n", filePath)
			}
		}
		f, err := os.Create(file.Location)
		if err != nil {
			log.Fatalf("could not create file %s\n", file.Location)
		}
		f.WriteString(renderTemplate(file.Content, initConfiguration))
		f.Close()
	}
}

func renderTemplate(templateContents string, initConfiguration InitConfiguration) string {
	template, err := template.New("").Parse(templateContents)
	if err != nil {
		log.Fatal("could not parse template")
	}
	var rendered bytes.Buffer
	if err := template.Execute(&rendered, initConfiguration); err != nil {
		log.Fatal("could not render configuration")
	}
	return rendered.String()
}
