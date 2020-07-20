% skuba-node-upgrade-apply(1) # skuba node upgrade apply - Apply the upgrade plan

# NAME

apply - Applies the upgrade plan for the given node

# SYNOPSIS
**apply**
[**--help**|**-h**] [**--port**|**-p**] [**--sudo**|**-s**] [**--target**|**-t**]
[**--user**|**-u**]
*apply* *-t <fqdn>* [-hs] [-u user] [-p port]

# DESCRIPTION
**apply** Evaluates the upgrade plan and it also applies it for the given node

# OPTIONS

**--help, -h**
  Print usage statement.

**--port, -p**
  Port to connect to using SSH

**--sudo, -s**
  Run remote command via sudo (defaults to ssh connection user identity)

**--target, -t**
  IP or host name of the node to connect to using SSH

**--user, -u**
  User identity used to connect to target (default=root)
