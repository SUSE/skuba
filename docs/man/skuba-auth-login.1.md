% skuba-auth-login(1) # skuba auth login - Login to a cluster

# NAME
login - Authenticate to a cluster and authorized with kubeconfig 

# SYNOPSIS
**login**
[**--help**|**-h**] [**--server|-s**] [**--username|-u**]
[**--password|-p**] [**--auth-connector|-a**] [**--root-ca**|**-r**]
[**--insecure**|**-k**] [**--cluster-name**|**-n**] [**--kubeconfig**|**-c**]
*login* [--server https://<ip/fqdn>:<port>] [--username username] [--password password]

# DESCRIPTION
**login** lets you login to a cluster and authorized with kubeconfig

# OPTIONS

**--help, -h**
  Print usage statement.

**--server, -s**
  (Required) The OIDC dex server url (https://<controller plane IP/FQDN>:<port>)

**--username, -u**
  The authentication username

**--password, -p**
  The authentication password

**--auth-connector, -a**
  The authentication connector ID

**--root-ca, -r**
  The cluster root certificate authority chain file

**--insecure, -k**
  Insecure SSL/TLS connection to OIDC dex server and further kube apiserver (true|false)

**--cluster-name, -n**
  The cluster name (default=local)

**--kubeconfig, -c**
  The path to stores kubeconfig (default=kubeconf.txt)
