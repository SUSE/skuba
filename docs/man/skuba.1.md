% skuba(1) # skuba - tool to manage the full lifecycle of a Kubernetes cluster
# NAME
skuba - tool to manage the full lifecycle of a Kubernetes cluster

# SYNOPSIS
**skuba**
[**-h**|**--help**] [**-v**|**--verbosity**]
*command* [*args*]

# DESCRIPTION
**skuba** is a tool that allows for Kubernetes cluster creation and
reconfiguration in an easy way.

# GLOBAL OPTIONS

**-h, --help**
  Print usage statement.

**-v, --verbosity**
  Log level [0-5]. 0 (Only Error and Warning) to 5 (Maximum detail).

# COMMANDS

**cluster**
  Cluster initialization and handling commands.

**completion**
  Command line completion commands.

**node**
  Node handling commands.

**addon**
  Addon handling commands.

# SEE ALSO
**skuba-auth-login**(1),
**skuba-cluster-images**(1),
**skuba-cluster-init**(1),
**skuba-cluster-status**(1),
**skuba-cluster-upgrade-plan**(1),
**skuba-node-bootstrap**(1),
**skuba-node-join**(1),
**skuba-node-remove**(1),
**skuba-node-upgrade-plan**(1),
**skuba-node-upgrade-apply**(1),
**skuba-addon-upgrade-plan**(1),
**skuba-addon-upgrade-apply**(1)
