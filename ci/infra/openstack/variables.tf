variable "image_name" {
  default     = "openSUSE-Leap-15.0-OpenStack.x86_64"
  description = "Name of the image to use"
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

variable "internal_net" {
  default     = "testing-net"
  description = "Name of the internal network to be created"
}

variable "subnet_cidr" {
  default     = "172.28.0.0/24"
  description = "CIDR of the subnet for the internal network"
}

variable "dns_nameservers" {
  type = "list"

  default = [
    "172.28.0.2",
    "8.8.8.8",
    "8.8.8.4",
  ]

  description = "DNS servers for the nodes"
}

variable "external_net" {
  default     = "floating"
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
  default     = 1
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
  default     = "testing.qa.caasp.suse.net"
  description = "TBD - leftover?"
}

variable "dnsentry" {
  default     = 0
  description = "TBD - leftover?"
}

variable "stack_name" {
  default     = "testing"
  description = "identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes"
}

variable "ntp_servers" {
  type        = "list"
  default     = []
  description = "list of ntp servers to configure"
}

variable "packages" {
  type = "list"

  default = [
    "kubernetes-kubeadm",
    "kubernetes-kubelet",
    "kubernetes-client",
    "cri-o",
    "cni-plugins",
    "-docker",
    "-containerd",
    "-docker-runc",
    "-docker-libnetwork",
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
