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

package kubectl

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"

	skubaconstants "github.com/SUSE/skuba/pkg/skuba"
)

func Apply(manifest string) error {
	cmd := exec.Command("kubectl", "apply", "--kubeconfig", skubaconstants.KubeConfigAdminFile(), "-f", "-")
	cmd.Stdin = bytes.NewBuffer([]byte(manifest))
	if combinedOutput, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "failed to run kubectl apply: %s", combinedOutput)
	}
	return nil
}
