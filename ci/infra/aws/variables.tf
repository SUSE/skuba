variable "stack_name" {
  default     = "k8s"
  description = "identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "aws_region" {
  default     = "eu-north-1"
  description = "Name of the AWS region to be used"
}

variable "ami_name_pattern" {
  default     = "suse-sles-15-*"
  description = "Pattern for choosing the AMI image"
}

variable "authorized_keys" {
  type        = list(string)
  default     = []
  description = "ssh keys to inject into all the nodes. First key will be used for creating a keypair."
}

variable "public_subnet" {
  type        = string
  description = "CIDR blocks for each public subnet of vpc"
  default     = "10.1.1.0/24"
}

variable "private_subnet" {
  type        = string
  description = "Private subnet of vpc"
  default     = "10.1.4.0/24"
}

variable "vpc_cidr_block" {
  type        = string
  description = "CIRD blocks for vpc"
  default     = "10.1.0.0/16"
}

variable "aws_access_key" {
  default     = ""
  description = "AWS access key"
}

variable "aws_secret_key" {
  default     = ""
  description = "AWS secret key"
}

variable "master_size" {
  default     = "t2.large"
  description = "Size of the master nodes"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "worker_size" {
  default     = "t2.medium"
  description = "Size of the worker nodes"
}

variable "workers" {
  default     = 1
  description = "Number of worker nodes"
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Extra tags used for the AWS resources created"
}

variable "repositories" {
  type        = list(string)
  default     = []
  description = "List of extra repositories (as maps with '<name>'='<url>') to add via cloud-init"
}

variable "packages" {
  type = list(string)

  default = [
    "kmod",
    "-docker",
    "-containerd",
    "-docker-runc",
    "-docker-libnetwork",
  ]

  description = "list of additional packages to install"
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

variable "iam_profile_master" {
  default     = ""
  description = "IAM profile associated with the master nodes"
}

variable "iam_profile_worker" {
  default     = ""
  description = "IAM profile associated with the worker nodes"
}

variable "peer_vpc_ids" {
  type        = list(string)
  default     = []
  description = "IDs of a VPCs to connect to via a peering connection"
}

