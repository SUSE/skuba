variable "stack_name" {
  default     = "k8s"
  description = "identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "Extra tags used for the Azure resources created"
}

variable "location" {
  default     = "West Europe"
  description = "Name of the Azure location to be used"
}

variable "ami_name_pattern" {
  default     = "suse-sles-15-*"
  description = "Pattern for choosing the AMI image"
}

variable "create_bastionhost" {
  type        = bool
  description = "Enables creation of bastion host"
  default     = false
}

variable "bastionhost_subnet_cidr" {
  type        = string
  description = "CIDR blocks for the bastion host subnet"
  default     = "10.1.1.0/24"
}

variable "private_subnet_cidr" {
  type        = string
  description = "Private subnet of vnet"
  default     = "10.1.4.0/24"
}

variable "vnet_address_space" {
  type        = string
  description = "CIRD blocks for vnet"
  default     = "10.1.0.0/16"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "master_size" {
  default     = "Standard_D2s_v3"
  description = "Size of the master nodes"
}

variable "workers" {
  default     = 1
  description = "Number of worker nodes"
}

variable "worker_size" {
  default     = "Standard_D2s_v3"
  description = "Size of the worker nodes"
}

variable "master_use_spot_instance" {
  default     = false
  description = "Use spot instances for master nodes"
}

variable "worker_use_spot_instance" {
  default     = false
  description = "Use spot instances for worker nodes"
}

variable "admin_ssh_key" {
  type = string
  description = "ssh public key used by the admin user"
}

variable "admin_password" {
  type        = string
  default     = ""
  description = "password used by the admin user"
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

variable "dnsdomain" {
  default     = "caasp.cloud"
  description = "Name of DNS domain"
}
