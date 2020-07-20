% skuba-node-join(1) # skuba node join - join a node to a cluster

# NAME
join - join a node to a cluster

# SYNOPSIS
**join**
[**--help**|**-h**] [**--target**|**-t**] [**--user**|**-u**] [**--role**|**-r**]
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
  User identity used to connect to target (default=root)

**--role, -r**
  (required) Role that this node will have in the cluster (master|worker)

**--sudo, -s**
  Run remote command via sudo (defaults to ssh connection user identity)

**--port, -p**
  Port to connect to using SSH

**--ignore-preflight-errors**
  A list of checks whose errors will be shown as warnings. Value 'all' ignores errors from all checks.