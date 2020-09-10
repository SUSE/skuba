variable "azure_location" {
  type        = string
  default     = ""
  description = "Name of the AZURE location to be used (eg: 'West Europe')"
}

variable "enable_zone" {
  type        = bool
  default     = false
  description = "Use this if only zone is available in the deploying region"
}

variable "azure_availability_zones" {
  type        = list(string)
  default     = []
  description = "List of Availability Zones (e.g. [\"1\", \"2\", \"3\"])"
}

variable "cidr_block" {
  type        = string
  default     = "10.1.0.0/16"
  description = "CIDR blocks for virtual_network"
}

variable "bastionhost_subnet_cidr" {
  type        = string
  default     = "10.1.1.0/24"
  description = "CIDR blocks for the bastion host subnet"
}

variable "private_subnet_cidr" {
  type        = string
  default     = "10.1.4.0/24"
  description = "Private subnet of virtual_network"
}

variable "stack_name" {
  type        = string
  default     = "k8s"
  description = "Identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "dnsdomain" {
  type        = string
  default     = ""
  description = "Name of DNS domain"
}

variable "authorized_keys" {
  type        = list(string)
  default     = []
  description = "SSH keys to inject into all the nodes"
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
  type        = list(string)
  default     = []
  description = "List of packages to install"
}

variable "username" {
  type        = string
  default     = "sles"
  description = "Username for the cluster nodes"
}

variable "password" {
  type        = string
  default     = ""
  description = "Password for the cluster nodes. Warning: password based authentication is a security risk, please use key-based authentication instead."
}

variable "caasp_registry_code" {
  type        = string
  default     = ""
  description = "SUSE CaaSP Product Registration Code"
}

variable "rmt_server_name" {
  type        = string
  default     = ""
  description = "SUSE Repository Mirroring Server Name"
}

variable "suma_server_name" {
  type        = string
  default     = ""
  description = "SUSE Manager Server Name"
}

variable "create_bastionhost" {
  type        = bool
  default     = false
  description = "Enables creation of bastion host"
}

variable "masters" {
  default     = 3
  description = "Number of master nodes"
}

variable "master_use_spot_instance" {
  type        = bool
  default     = false
  description = "Use spot instances for master nodes"
}

variable "master_vm_size" {
  type        = string
  default     = "Standard_D2s_v3"
  description = "Virtual machine size of the master nodes"
}

variable "master_storage_account_type" {
  type        = string
  default     = "Standard_LRS"
  description = "Storage account type"
}

variable "master_disk_size" {
  default     = 30
  description = "Size of the disk size, in Gb"
}

variable "workers" {
  default     = 2
  description = "Number of worker nodes"
}

variable "worker_use_spot_instance" {
  type        = bool
  default     = false
  description = "Use spot instances for worker nodes"
}

variable "worker_vm_size" {
  type        = string
  default     = "Standard_D2s_v3"
  description = "Virtual machine size of the worker nodes"
}

variable "worker_storage_account_type" {
  type        = string
  default     = "Standard_LRS"
  description = "Storage account type"
}

variable "worker_disk_size" {
  default     = 30
  description = "Size of the disk size, in Gb"
}

variable "peer_virutal_network_id" {
  type        = list(string)
  default     = []
  description = "IDs of a Virtual Network to connect to via a peering connection"
}

variable "cpi_enable" {
  default     = false
  description = "Enable CPI integration with Azure"
}
