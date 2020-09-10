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

package ssh

// ZypperInstall runs a zypper command to install an arbitrary list of packages,
// wrapped with the right userdata and parameters
func (t *Target) ZypperInstall(packages ...string) (stdout string, stderr string, error error) {
	var cliArgs []string
	cliArgs = append(cliArgs, "--userdata", "skuba", "-i", "--non-interactive", "install", "--")
	cliArgs = append(cliArgs, packages...)
	return t.ssh("zypper", cliArgs...)
}
