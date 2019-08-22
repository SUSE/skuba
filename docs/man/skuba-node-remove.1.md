% skuba-node-remove(1) # skuba node remove - remove a node to a state prior to join or bootstrap

# NAME
remove - remove a node to a state prior to join or bootstrap

# SYNOPSIS
**remove**
[**--help**|**-h**]
*remove* *node-name*

# DESCRIPTION
**remove** will permanently remove a node from the cluster via kubernetes.  Note that this node 
cannot be added back to the cluster or any other skuba-initiated kubernetes cluster without 
reinstalling first.

# OPTIONS

**--help, -h**
  Print usage statement.
