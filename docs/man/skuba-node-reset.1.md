% skuba-node-reset(1) # skuba node reset - Reset a node to a state prior to join or bootstrap

# NAME
reset - Reset a node to a state prior to join or bootstrap

# SYNOPSIS
**reset**
[**--help**|**-h**] [**--target**|**-t**] [**--user**|**-u**]
[**--sudo**|**-s**] [**--port**|**-p**] [**--ignore-preflight-errors**]
*reset* *-t <fqdn>* [-hsp] [-u user] [-p port] [-v level]

# DESCRIPTION
**reset** is a command that enables you to reset a node 
to the state prior to join or bootstrap being run.

# OPTIONS

**--help, -h**
  Print usage statement.

**-v**
  set verbosity level.

**--target, -t**
  (required) IP or host name of the node to connect to using SSH

**--user, -u**
  User identity used to connect to target (default=root)

**--sudo, -s**
  Run remote command via sudo (defaults to ssh connection user identity)

**--port, -p**
  Port to connect to using SSH

**--ignore-preflight-errors**
  A list of checks whose errors will be shown as warnings. Value 'all' ignores errors from all checks.
