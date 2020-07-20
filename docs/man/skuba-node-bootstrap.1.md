% skuba-node-bootstrap(1) # skuba node bootstrap - Bootstraps the first master node of the cluster

# NAME
bootstrap - Bootstraps the first master node of the cluster

# SYNOPSIS
**bootstrap**
[**--help**|**-h**] [**--target**|**-t**] [**--user**|**-u**]
[**--sudo**|**-s**] [**--port**|**-p**] [**--ignore-preflight-errors**]
*bootstrap* *<node-name>* *-t <fqdn>* [-hsp] [-u user] [-p port]

# DESCRIPTION
**bootstrap** is a command that lets you bootstrap 
the first node of a cluster

# OPTIONS

**--help, -h**
  Print usage statement.

**--target, -t**
  IP or host name of the node to connect to using SSH

**--user, -u**
  User identity used to connect to target (default=root)

**--sudo, -s**
  Run remote command via sudo (defaults to ssh connection user identity)

**--port, -p**
  Port to connect to using SSH

**--ignore-preflight-errors**
  A list of checks whose errors will be shown as warnings. Value 'all' ignores errors from all checks.