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
	CriConfFolderReadme = `This directory contains CRI-O configuration files that are uploaded to CaaSP nodes. All the files (except this README) will be uploaded to the /etc/crio/conf.d/ folder. If it is required user custom files name as 99-custom.conf
	`
	criDockerDefaultsConf = `## Path           : System/Management
## Description    : Extra cli switches for crio daemon
## Type           : string
## Default        : ""
## ServiceRestart : crio
#
CRIO_OPTIONS=--pause-image={{.PauseImage}}{{if not .StrictCapDefaults}} --default-capabilities="CHOWN,DAC_OVERRIDE,FSETID,FOWNER,NET_RAW,SETGID,SETUID,SETPCAP,NET_BIND_SERVICE,SYS_CHROOT,KILL,MKNOD,AUDIT_WRITE,SETFCAP"{{end}}`
	criDefaultsConf = `# PLEASE DON'T EDIT THIS FILE!
# This file is managed by CaaSP skuba.
{{if not .StrictCapDefaults}}[crio.runtime]

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
]{{end}}

[crio.image]

pause_image = "{{.PauseImage}}"
`
)
