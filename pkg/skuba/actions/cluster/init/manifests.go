/*
 * Copyright (c) 2019-2020 SUSE LLC.
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
	CriConfFolderReadme = `This folder provides CRI-O configuration for CaaSP nodes.
All the files (except this README) will be uploaded to the /etc/crio/crio.conf.d/ folder.
If you need any customization, please add a custom files with the name 99-custom.conf
	`
	criDockerDefaultsConf = `# PLEASE DON'T EDIT THIS FILE!
# This file is managed by CaaSP skuba.

## Path           : System/Management
## Description    : Extra cli switches for crio daemon
## Type           : string
## Default        : ""
## ServiceRestart : crio
#
CRIO_OPTIONS=--pause-image={{.PauseImage}}{{if not .StrictCapDefaults}} --default-capabilities="CHOWN,DAC_OVERRIDE,FSETID,FOWNER,NET_RAW,SETGID,SETUID,SETPCAP,NET_BIND_SERVICE,SYS_CHROOT,KILL,MKNOD,AUDIT_WRITE,SETFCAP"{{end}}`
	criDefaultsConf = `# PLEASE DON'T EDIT THIS FILE!
# This file is managed by CaaSP skuba.

[crio.runtime]

{{if not .StrictCapDefaults}}
default_capabilities = [
	"CHOWN",
	"DAC_OVERRIDE",
	"FSETID",
	"FOWNER",
	"NET_RAW",
	"SETGID",
	"SETUID",
	"SETPCAP",
	"NET_BIND_SERVICE",
	"SYS_CHROOT",
	"KILL",
	"MKNOD",
	"AUDIT_WRITE",
	"SETFCAP",
]
{{else}}
pids_limit = 32768
{{end}}

[crio.image]

pause_image = "{{.PauseImage}}"
`
)
