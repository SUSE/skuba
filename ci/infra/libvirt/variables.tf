variable "libvirt_uri" {
  default     = "qemu:///system"
  description = "URL of libvirt connection - default to localhost"
}

variable "libvirt_keyfile" {
  default     = ""
  description = "The private key file used for libvirt connection - default to none"
}

variable "pool" {
  default     = "default"
  description = "Pool to be used to store all the volumes"
}

variable "image_uri" {
  default     = ""
  description = "URL of the image to use"
}

variable "repositories" {
  type        = map(string)
  default     = {}
  description = "Urls of the repositories to mount via cloud-init"
}

variable "stack_name" {
  default     = ""
  description = "Identifier to make all your resources unique and avoid clashes with other users of this terraform project"
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

variable "packages" {
  type = list(string)

  default = [
    "kernel-default",
    "-kernel-default-base",
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

variable "ha_registry_code" {
  default     = ""
  description = "SUSE Linux Enterprise High Availability Extension Registration Code"
}

variable "rmt_server_name" {
  default     = ""
  description = "SUSE Repository Mirroring Server Name"
}

variable "dns_domain" {
  type        = string
  default     = "caasp.local"
  description = "Name of DNS Domain"
}

variable "network_cidr" {
  type        = string
  default     = "10.17.0.0/22"
  description = "Network used by the cluster"
}

variable "network_mode" {
  type        = string
  default     = "nat"
  description = "Network mode used by the cluster"
}

variable "network_name" {
  default     = ""
  description = "The virtual network name to use. If provided just use the given one (not managed by terraform), otherwise terraform creates a new virtual network resource"
}

variable "create_lb" {
  type        = bool
  default     = true
  description = "Create load balancer node exposing master nodes"
}

variable "lb_memory" {
  default     = 4096
  description = "Amount of RAM for a load balancer node"
}

variable "lb_vcpu" {
  default     = 1
  description = "Amount of virtual CPUs for a load balancer node"
}

variable "lb_disk_size" {
  default     = "25769803776"
  description = "Disk size (in bytes)"
}

variable "lb_repositories" {
  type = map(string)

  default = {
    sle_server_pool    = "http://download.suse.de/ibs/SUSE/Products/SLE-Product-SLES/15-SP1/x86_64/product/"
    basesystem_pool    = "http://download.suse.de/ibs/SUSE/Products/SLE-Module-Basesystem/15-SP1/x86_64/product/"
    ha_pool            = "http://download.suse.de/ibs/SUSE/Products/SLE-Product-HA/15-SP1/x86_64/product/"
    ha_updates         = "http://download.suse.de/ibs/SUSE/Updates/SLE-Product-HA/15-SP1/x86_64/update/"
    sle_server_updates = "http://download.suse.de/ibs/SUSE/Updates/SLE-Product-SLES/15-SP1/x86_64/update/"
    basesystem_updates = "http://download.suse.de/ibs/SUSE/Updates/SLE-Module-Basesystem/15-SP1/x86_64/update/"
  }
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "master_memory" {
  default     = 4096
  description = "Amount of RAM for a master"
}

variable "master_vcpu" {
  default     = 2
  description = "Amount of virtual CPUs for a master"
}

variable "master_disk_size" {
  default     = "25769803776"
  description = "Disk size (in bytes)"
}

variable "workers" {
  default     = 2
  description = "Number of worker nodes"
}

variable "worker_memory" {
  default     = 4096
  description = "Amount of RAM for a worker"
}

variable "worker_vcpu" {
  default     = 2
  description = "Amount of virtual CPUs for a worker"
}

variable "worker_disk_size" {
  default     = "25769803776"
  description = "Disk size (in bytes)"
}

variable "hostname_from_dhcp" {
  default     = true
  description = "Set node's hostname from DHCP server"
}
