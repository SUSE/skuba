variable "image_name" {
  default     = ""
  description = "Name of the image to use"
}

variable "repositories" {
  type        = "list"
  default     = []
  description = "Urls of the repositories to mount via cloud-init"
}

variable "internal_net" {
  default     = ""
  description = "Name of the internal network to be created"
}

variable "subnet_cidr" {
  default     = ""
  description = "CIDR of the subnet for the internal network"
}

variable "dns_nameservers" {
  type        = "list"
  default     = []
  description = "DNS servers for the nodes"
}

variable "external_net" {
  default     = ""
  description = "Name of the external network to be used, the one used to allocate floating IPs"
}

variable "master_size" {
  default     = "m1.medium"
  description = "Size of the master nodes"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "worker_size" {
  default     = "m1.medium"
  description = "Size of the worker nodes"
}

variable "workers" {
  default     = 2
  description = "Number of worker nodes"
}

variable "workers_vol_enabled" {
  default     = 0
  description = "Attach persistent volumes to workers"
}

variable "workers_vol_size" {
  default     = 5
  description = "size of the volumes in GB"
}

variable "dnsdomain" {
  default     = ""
  description = "Name of DNS domain"
}

variable "dnsentry" {
  default     = 0
  description = "DNS Entry"
}

variable "stack_name" {
  default     = ""
  description = "Identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "SSH keys to inject into all the nodes"
}

variable "ntp_servers" {
  type        = "list"
  default     = []
  description = "List of ntp servers to configure"
}

variable "packages" {
  type = "list"

  default = [
    "kernel-default",
    "-kernel-default-base",
    "patterns-caasp-Node",
  ]

  description = "list of additional packages to install"
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
