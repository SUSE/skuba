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

## fixme: see issue https://github.com/SUSE/avant-garde/issues/91
variable "img_source_url" {
  type    = "string"
  default = "https://download.opensuse.org/repositories/Cloud:/Images:/Leap_15.0/images/openSUSE-Leap-15.0-OpenStack.x86_64-0.0.4-Buildlp150.12.136.qcow2"
}

variable "repositories" {
  type = "list"

  default = [
    {
      caasp_devel_leap15 = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
    },
  ]

  description = "Urls of the repositories to mount via cloud-init"
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
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes"
}

variable "packages" {
  type = "list"

  default = [
    "patterns-caasp-Node",
  ]

  description = "list of additional packages to install"
}

variable "username" {
  default     = "opensuse"
  description = "Username for the cluster nodes"
}

variable "password" {
  default     = "linux"
  description = "Password for the cluster nodes"
}

# Extend disk size to 24G (JeOS-KVM default size) because we use
# JeOS-OpenStack instead of JeOS-KVM image with libvirt provider
variable "disk_size" {
  default     = "25769803776"
  description = "disk size (in bytes)"
}
