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

package ssh

import (
	"encoding/base64"
	"fmt"
	"path"

	"k8s.io/klog"
)

func (t *Target) UploadFileContents(targetPath, contents string) error {
	klog.V(1).Infof("uploading to remote file %q with contents", targetPath)
	dir, _ := path.Split(targetPath)
	encodedContents := base64.StdEncoding.EncodeToString([]byte(contents))
	if _, _, err := t.silentSsh("mkdir", "-p", dir); err != nil {
		return err
	}
	_, _, err := t.silentSshWithStdin(encodedContents, "base64", "-d", "-w0", fmt.Sprintf("> %s", targetPath))
	return err
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	klog.V(1).Infof("downloading remote file %q contents", sourcePath)
	if stdout, _, err := t.silentSsh("base64", "-w0", sourcePath); err == nil {
		decodedStdout, err := base64.StdEncoding.DecodeString(stdout)
		if err != nil {
			return "", err
		}
		return string(decodedStdout), nil
	} else {
		return "", err
	}
}
