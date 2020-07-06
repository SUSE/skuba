% skuba-node-upgrade-apply(1) # skuba node upgrade apply - Apply the upgrade plan

# NAME

apply - Applies the upgrade plan for the given node

# SYNOPSIS
**apply**
[**--help**|**-h**] [**--port**|**-p**] [**--sudo**|**-s**] [**--target**|**-t**]
[**--bastion] [**--bastion-user**] [**--bastion-port**]
[**--user**|**-u**]
*apply* *-t <fqdn>* [-hs] [-u user] [-p port]

# DESCRIPTION
**apply** Evaluates the upgrade plan and it also applies it for the given node

# OPTIONS

**--help, -h**
  Print usage statement.

**--target, -t**
  IP or host name of the node to connect to using SSH

**--user, -u**
  User identity used to connect to target (required)

**--port, -p**
  Port to connect to using SSH

**--sudo, -s**
  Run remote command via sudo (defaults to ssh connection user identity)

**--bastion**
  IP or FQDN of the bastion to connect to the other nodes using SSH

**--bastion-user**
  User identity used to connect to the bastion using SSH (defaults to target user)

**--bastion-port**
  Port to connect to the bastion using SSH (default 22)
