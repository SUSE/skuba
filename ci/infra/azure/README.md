# Introduction

This terraform project creates the infrastructure needed to run a
cluster on top of Azure.

Once the infrastructure is up and running nothing special has to be done
to deploy CaaS Platform on top of it.

This document focuses on the key aspects of the infrastructure created
by terraform.

# Cluster layout

## Setup service principal for terraform login credential

Following the guide [Creating a Service Principal in the Azure Portal](https://www.terraform.io/docs/providers/azurerm/guides/service_principal_client_secret.html#creating-a-service-principal-in-the-azure-portal), and set up `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `and ARM_TENANT_ID` in `container-openrc.sh`.  Source `container-openrc.sh` before deploying terraform script.

## Machines

As usual the cluster is based on two types of nodes: master and worker nodes.

All the nodes are created using the SLES 15 SP1 container host image built
and maintained by the SUSE Public Cloud team.

Right now users **must** bring their own license into the public cloud to be
able to access SLE and SUSE CaaS Platform packages.

The machines are automatically registered at boot time against SUSE Customer
Center, RMT or a SUSE Manager instance depending on which one of the following
variables has been set:

  * `caasp_registry_code`
  * `rmt_server_name`
  * `suma_server_name`

The SLES images [do not yet support cloud-init](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/using-cloud-init)
(it will probably be supported starting from SLE15 SP2). In the meantime
terraform leverages the Azure Linux Extension capabilities provided by
the [Azure Linux Agent](https://docs.microsoft.com/en-us/azure/virtual-machines/extensions/agent-linux).

### Using spot instances

It's possible to create a [spot VMs](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/spot-vms)
both for the master and worker nodes.

This can be done by setting these variables to `true` (they are set to
`false` by default):

  * `master_use_spot_instance`
  * `worker_use_spot_instance`

## Network layout

All of the infrastructure is created inside of a user specified AZURE region.
The resources are currently all located inside of the user specified availability
zones. All the nodes are placed inside of the same virtual network, within the same
subnet.

Worker nodes are never exposed to the public internet. On the opposite
each master nodes has a public IP address by default. This allows users to
connect to them via ssh from their computers.

It's also possible disable this behaviour and make **all** the nodes private.
This can be done setting the `create_bastionhost` variable to `true`.

When this variable is set all the master nodes cease to have a public IP address.
An [Azure Bastion](https://docs.microsoft.com/en-us/azure/bastion/bastion-overview)
instance is created which becomes the only way to ssh into the cluster.

Terraform creates also an internal DNS zone with the domain specified via the
`dnsdomain` variable. This allows all the nodes to reach each other using
their FQDN.

### Security groups

Terraform automatically creates security groups for the master and worker nodes
that are going to allow connections only to the allowed services. The security
rules are a 1:1 mapping of what we describe inside of SUSE CaaS Platform
documentation.

## Load balancer

Terraform automatically creates a load balancer with a public IP that exposes
the following services running on the control plane nodes:

  * kubernetes API server: port 6443
  * dex: port 32000
  * gangway: port 32001

This is exactly the same behaviour used by other deployment platforms.

## Accessing the nodes

A default `sles` user is created on each node of the cluster. The user has
administrator capabilities by using the `sudo` utility.

By default password based authentication is disabled. It's possible to log
using the ssh key specified via the `admin_ssh_key` variable.
It's also possible to enable password based authentication by specifying a
value for the `admin_password` variable. Note well: Azure has some security
checks in place to avoid the usage of weak passwords.

When the bastion host creation is disabled the access to the master nodes of
the cluster is just a matter of doing a ssh against their public IP address.

Accessing a cluster through an Azure Bastion requires a different procedure.

### Using Azure Bastion

Azure bastion supports only RSA keys with PEM format. These can be created by
doing:

```
ssh-keygen -f azure -t rsa -b 4096 -m pem
```

This will create a public and private key named `azure`. The **private** key has
to be provided later to Azure, hence it's strongly recommended to create
a dedicate key pair.

Once the whole infrastructure is created you can connect into any node of the
cluster by doing the following steps:

  1. Log into Azure portal
  2. Choose one of the nodes of the cluster
  3. Click "connect" and select "bastion" as option
  4. Enter all the required fields

Once this is done a new browser tab will be open with a shell session running
inside of the desired node.

It's recommended to use Chrome or Chromium during this process.

You can ssh into the first bootstrapped master node to download the kubeconfig 
file to operate the cluster without having to go through the bastion host.

Caveats of Azure Bastion:

  * As of June 2020, the [Azure Bastion service](https://docs.microsoft.com/en-us/azure/bastion/bastion-overview#regions) is not available in all Azure regions.
  * By design it's not possible to leverage the bastion host without using the
    ssh session embedded into the browser. This makes impossible to use tools like
    `sftp` or `scp`.
  * You have to rely on copy and paste to share data (like the `admin.conf` file
    generated by skuba) between the remote nodes and your local system.
    You can "rely" on `cat`, `base64` and a lot of copy and paste...
  * `skuba` requires a private ssh key to connect to all the nodes of the cluster.
    You have to upload the private key you specified at cluster creation
    or create a new one inside of the first master node and copy that
    around the cluster.

## Virtual Network Peering Support

It is possible to join existing network to the cluster.  It can be setup by adding a list of network id to `peer_virutal_network_id`.

## Enable Multiple Zones

It is possible to enable multiple zone.  It can be set `enable_zone` to `true` and master/worker node will distribute sequentially based on zones defined in `azure_availability_zones`.

  * As of June 2020, the [Azure Availability Zones](https://docs.microsoft.com/en-us/azure/availability-zones/az-region) is not available in all Azure regions.
