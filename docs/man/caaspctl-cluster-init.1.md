% caaspctl-cluster-init(1) # caaspctl cluster init - Initialize caaspctl structure for cluster deployment

# NAME
init - Initialize caaspctl structure for cluster deployment

# SYNOPSIS
**init**
[**--help**|**-h**] [**--control-plane**]
*init* *<node-name>* [--control-plane fqdn]

# DESCRIPTION
**init** Lets you Initialize the files required for cluster deployment

# OPTIONS

**--help, -h**
  Print usage statement.

**--control-plane**
  (Required) The control plane location (IP/FQDN) that will load balance the master nodes
