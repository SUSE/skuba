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
- [~AWS~](aws/README.md)
 
## Important
 
If you don't make any change on this directories the cluster will be deployed
with no specific cloud provider integration.
 
Only one cloud provider integration is supported for a given cluster.

## Cleanup

By enabling CPI providers your kubernetes cluster will be able to provision cloud
resources on its own (eg: Load Balancers, Persistent Volumes). You will have to manually
clean these resources before you destroy the cluster with terraform.
`, "~", "`")

	openstackReadme = strings.ReplaceAll(`# ~Openstack~ integration
 
Create a file inside this directory named ~openstack.conf~, with the [supported contents](https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/provider-configuration.md):
 
You can find a template named openstack.conf.template inside this directory.

Warning!!! Please replace option values with variables from OpenStack RC file.

~~~
[Global]
auth-url=<OS_AUTH_URL>
username=<OS_USERNAME>
password=<OS_PASSWORD>
tenant-id=<OS_PROJECT_ID>
domain-name=<OS_USER_DOMAIN_NAME>
region=<OS_REGION_NAME>
[LoadBalancer]
lb-version=v2
subnet-id=<PRIVATE_SUBNET_ID>
floating-network-id=<PUBLIC_NET_ID>
create-monitor=yes
monitor-delay=1m
monitor-timeout=30s
monitor-max-retries=3
[BlockStorage]
trust-device-path=false
bs-version=v2
ignore-volume-az=true
~~~
 
If this file exists the cloud integration for ~Openstack~ will be automatically
enabled when you bootstrap the cluster.

## Important
 
When the cloud provider integration is enabled, it's very important to bootstrap
and join nodes with the same node names that they have inside ~Openstack~, as
this name will be used by the ~Openstack~ cloud controller manager to reconcile
node metadata.
`, "~", "`")

	openstackCloudConfTemplate = `[Global]
auth-url=<OS_AUTH_URL>
username=<OS_USERNAME>
password=<OS_PASSWORD>
tenant-id=<OS_PROJECT_ID>
domain-name=<OS_USER_DOMAIN_NAME>
region=<OS_REGION_NAME>
[LoadBalancer]
lb-version=v2
subnet-id=<PRIVATE_SUBNET_ID>
floating-network-id=<PUBLIC_NET_ID>
create-monitor=yes
monitor-delay=1m
monitor-timeout=30s
monitor-max-retries=3
[BlockStorage]
trust-device-path=false
bs-version=v2
ignore-volume-az=true
`

	awsReadme = strings.ReplaceAll(`# ~AWS~ integration
 
No special operation needs to be performed if you deployed the whole
infrastructure using the terraform state files provided by SUSE.
`, "~", "`")
)
