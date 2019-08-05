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

package deployments

import (
	"errors"
	"fmt"
	"io/ioutil"

	"k8s.io/klog"
)

type Actionable interface {
	Apply(data interface{}, states ...string) error
	UploadFileContents(targetPath, contents string) error
	DownloadFileContents(sourcePath string) (string, error)
	IsServiceEnabled(serviceName string) (bool, error)
}

type TargetCache struct {
	OsRelease map[string]string
}

type Target struct {
	Actionable
	Target   string
	Nodename string
	Cache    TargetCache
}

func (t *Target) Apply(data interface{}, states ...string) error {
	filteredStates := []string{}
	for _, s := range states {
		if s == "" {
			continue
		}
		filteredStates = append(filteredStates, s)
	}

	return t.Actionable.Apply(data, filteredStates...)
}

func (t *Target) UploadFile(sourcePath, targetPath string) error {
	klog.V(1).Infof("uploading local file %q to remote file %q", sourcePath, targetPath)
	if contents, err := ioutil.ReadFile(sourcePath); err == nil {
		return t.UploadFileContents(targetPath, string(contents))
	}
	return errors.New(fmt.Sprintf("could not find file %s", sourcePath))
}

func (t *Target) UploadFileContents(targetPath, contents string) error {
	return t.Actionable.UploadFileContents(targetPath, contents)
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	return t.Actionable.DownloadFileContents(sourcePath)
}

func (t *Target) IsServiceEnabled(serviceName string) (bool, error) {
	return t.Actionable.IsServiceEnabled(serviceName)
}
