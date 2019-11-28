% skuba(1) # skuba - tool to manage the full lifecycle of a Kubernetes cluster
# NAME
skuba - tool to manage the full lifecycle of a Kubernetes cluster

# SYNOPSIS
**skuba**
[**--help**|**-h**] [**-v**]
*command* [*args*]

# DESCRIPTION
**skuba** is a tool that allows for Kubernetes cluster creation and
reconfiguration in an easy way.

# GLOBAL OPTIONS

**--help, -h**
  Print usage statement.

**-v**
  number for the log level verbosity [0-10].

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
