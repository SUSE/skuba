% skuba-cluster-init(1) # skuba cluster init - Initialize skuba structure for cluster deployment

# NAME
init - Initialize skuba structure for cluster deployment

# SYNOPSIS
**init**
[**--help**|**-h**] [**--control-plane**] [**--cloud-provider**]
*init* *<node-name>* [--control-plane fqdn]

# DESCRIPTION
**init** Lets you Initialize the files required for cluster deployment

# OPTIONS

**--help, -h**
  Print usage statement.

**--control-plane**
  (Required) The control plane location (IP/FQDN) that will load balance the master nodes

**--cloud-provider string**
  Enable cloud provider integration with the chosen cloud. Valid values: aws, openstack

**--strict-capability-defaults**
  All the containers will start with CRI-O default capabilities
