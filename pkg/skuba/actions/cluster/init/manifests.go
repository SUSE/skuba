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

const (
	criDockerDefaultsConf = `## Path           : System/Management
## Description    : Extra cli switches for crio daemon
## Type           : string
## Default        : ""
## ServiceRestart : crio
#
### BUG [ CRIO v1.18.0rc1 ] - string parsing issue based on comma separators
CRIO_OPTIONS=--pause-image={{.PauseImage}}{{if not .StrictCapDefaults}} --default-capabilities CHOWN --default-capabilities DAC_OVERRIDE --default-capabilities FSETID --default-capabilities FOWNER --default-capabilities NET_RAW --default-capabilities SETGID --default-capabilities SETUID --default-capabilities SETPCAP --default-capabilities NET_BIND_SERVICE --default-capabilities SYS_CHROOT --default-capabilities KILL --default-capabilities MKNOD --default-capabilities AUDIT_WRITE --default-capabilities SETFCAP{{end}}
`
)
