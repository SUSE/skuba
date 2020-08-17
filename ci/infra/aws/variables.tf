variable "aws_region" {
  type = string
  # default     = "eu-north-1"
  description = "Name of the AWS region to be used"
}

variable "aws_availability_zones" {
  type = list(string)
  # default     = ["us-west-2a", "us-west-2b", "us-west-2c"]
  description = "List of Availability Zones (e.g. `['us-east-1a', 'us-east-1b', 'us-east-1c']`)"
}

variable "enable_resource_group" {
  type        = bool
  default     = false
  description = "Use this to enable resource group"
}

variable "cidr_block" {
  type        = string
  default     = "10.1.0.0/16"
  description = "CIDR blocks for vpc"
}

variable "stack_name" {
  default     = "k8s"
  description = "Identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Extra tags used for the AWS resources created"
}

variable "authorized_keys" {
  type        = list(string)
  default     = []
  description = "SSH keys to inject into all the nodes"
}

variable "key_pair" {
  default     = ""
  description = "SSH key stored in openstack to create the nodes with"
}

variable "ntp_servers" {
  type        = list(string)
  default     = []
  description = "List of NTP servers to configure"
}

variable "dns_nameservers" {
  type        = list(string)
  default     = []
  description = "List of Name servers to configure"
}

variable "repositories" {
  type        = map(string)
  default     = {}
  description = "Maps of repositories with '<name>'='<url>' to add via cloud-init"
}

variable "packages" {
  type = list(string)

  default = [
    "kmod",
    "-docker",
    "-containerd",
    "-containerd-ctr",
    "-docker-runc",
    "-docker-libnetwork",
    "-docker-img-store-setup",
  ]

  description = "List of packages to install"
}

variable "username" {
  default     = "sles"
  description = "Username for the cluster nodes"
}

variable "password" {
  default     = "linux"
  description = "Password for the cluster nodes"
}

variable "caasp_registry_code" {
  default     = ""
  description = "SUSE CaaSP Product Registration Code"
}

variable "rmt_server_name" {
  default     = ""
  description = "SUSE Repository Mirroring Server Name"
}

variable "suma_server_name" {
  default     = ""
  description = "SUSE Manager Server Name"
}

variable "peer_vpc_ids" {
  type        = list(string)
  default     = []
  description = "IDs of a VPCs to connect to via a peering connection"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "master_instance_type" {
  default     = "t2.medium"
  description = "Instance type of the master nodes"
}

variable "master_volume_size" {
  default     = 20
  description = "Size of the EBS volume, in GB"
}

variable "iam_profile_master" {
  default     = ""
  description = "IAM profile associated with the master nodes"
}

variable "workers" {
  default     = 2
  description = "Number of worker nodes"
}

variable "worker_instance_type" {
  default     = "t2.medium"
  description = "Instance type of the worker nodes"
}

variable "worker_volume_size" {
  default     = 20
  description = "Size of the EBS volume, in GB"
}

variable "iam_profile_worker" {
  default     = ""
  description = "IAM profile associated with the worker nodes"
}
