#####################
# libvirt variables #
#####################

variable "libvirt_uri" {
  default     = "qemu:///system"
  description = "libvirt connection url - default to localhost"
}

variable "pool" {
  default     = "default"
  description = "pool to be used to store all the volumes"
}

#####################
# Cluster variables #
#####################

variable "img_source_url" {
  type        = "string"
  default     = "https://download.opensuse.org/repositories/Cloud:/Images:/Leap_15.0/images/openSUSE-Leap-15.0-OpenStack.x86_64-0.0.4-Buildlp150.12.127.qcow2"
}

variable "repo_baseurl" {
  type        = "string"
  default     = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
}

variable "lb_memory" {
  default     = 2048
  description = "The amount of RAM for a load balancer node"
}

variable "lb_vcpu" {
  default     = 1
  description = "The amount of virtual CPUs for a load balancer node"
}

variable "master_count" {
  default     = 1
  description = "Number of masters to be created"
}

variable "master_memory" {
  default     = 2048
  description = "The amount of RAM for a master"
}

variable "master_vcpu" {
  default     = 2
  description = "The amount of virtual CPUs for a master"
}

variable "worker_count" {
  default     = 2
  description = "Number of workers to be created"
}

variable "worker_memory" {
  default     = 2048
  description = "The amount of RAM for a worker"
}

variable "worker_vcpu" {
  default     = 2
  description = "The amount of virtual CPUs for a worker"
}

variable "name_prefix" {
  type        = "string"
  default     = "ag-"
  description = "Optional prefix to be able to have multiple clusters on one host"
}

variable "domain_name" {
  type        = "string"
  default     = "test.net"
  description = "The domain name"
}

variable "net_mode" {
  type        = "string"
  default     = "nat"
  description = "Network mode used by the cluster"
}

variable "network" {
  type        = "string"
  default     = "10.17.0.0/22"
  description = "Network used by the cluster"
}

variable "authorized_keys" {
  type = "list"
  default = []
  description = "ssh keys to inject into all the nodes"
}
