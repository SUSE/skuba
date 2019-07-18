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
	"strings"
)

var (
	cloudReadme = strings.ReplaceAll(`# Cloud provider integration
 
 This directory contains all supported cloud provider integration configurations
 and instructions.
 
 - [~Openstack~](openstack/README.md)
 
 ## Important
 
 If you don't make any change on this directories the cluster will be deployed
 with no specific cloud provider integration.
 
 Only one cloud provider integration is supported for a given cluster.
 `, "~", "`")

	openstackReadme = strings.ReplaceAll(`# ~Openstack~ integration
 
 Create a file inside this directory named ~cloud.conf~, with the [supported contents](https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/provider-configuration.md):
 
 ~~~
 [Global]
 username=user
 password=pass
 auth-url=https://<keystone_ip>/identity/v3
 tenant-id=c869168a828847f39f7f06edd7305637
 domain-id=2a73b8f597c04551a0fdc8e95544be8a
 ~~~
 
 You can find a template named cloud.conf.template inside this directory.
 
 If this file exists the cloud integration for ~Openstack~ will be automatically
 enabled when you bootstrap the cluster.
 
 ## Important
 
 When the cloud provider integration is enabled, it's very important to bootstrap
 and join nodes with the same node names that they have inside ~Openstack~, as
 this name will be used by the ~Openstack~ cloud controller manager to reconcile
 node metadata.
 `, "~", "`")

	openstackCloudConfTemplate = `[Global]
 username=user
 password=pass
 auth-url=https://<keystone_ip>/identity/v3
 tenant-id=c869168a828847f39f7f06edd7305637
 domain-id=2a73b8f597c04551a0fdc8e95544be8a
 `
)
