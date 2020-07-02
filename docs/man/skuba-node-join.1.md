% skuba-node-join(1) # skuba node join - join a node to a cluster

# NAME
join - join a node to a cluster

# SYNOPSIS
**join**
[**--help**|**-h**] [**--target**|**-t**] [**--user**|**-u**] [**--role**|**-r**]
[**--bastion] [**--bastion-user**] [**--bastion-port**]
[**--sudo**|**-s**] [**--port**|**-p**] [**--ignore-preflight-errors**]
*join* *<node-name>* *-t <fqdn>* [-hsp] [-r master] [-u user] [-p port]

# DESCRIPTION
**join** lets you join a new node to the cluster

# OPTIONS

**--help, -h**
  Print usage statement.

**--target, -t**
  IP or host name of the node to connect to using SSH

**--user, -u**
  User identity used to connect to target

**--port, -p**
  Port to connect to using SSH

**--sudo, -s**
  Run remote command via sudo (defaults to ssh connection user identity)

**--role, -r**
  (required) Role that this node will have in the cluster (master|worker)

**--ignore-preflight-errors**
  A list of checks whose errors will be shown as warnings. Value 'all' ignores errors from all checks.

**--bastion**
  IP or FQDN of the bastion to connect to the other nodes using SSH

**--bastion-user**
  User identity used to connect to the bastion using SSH (default to target user)

**--bastion-port**
  Port to connect to the bastion using SSH (default 22)
