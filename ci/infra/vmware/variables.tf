variable "template_name" {
}

variable "stack_name" {
}

variable "vsphere_datastore" {
  default = null
}

variable "vsphere_datastore_cluster" {
  default = null
}

variable "vsphere_datacenter" {
}

variable "vsphere_network" {
}

variable "vsphere_resource_pool" {
}

variable "vsphere_hardware_version" {
  default     = null
  description = "Hardware version of the virtual machine."
}

variable "authorized_keys" {
  type        = list(string)
  default     = []
  description = "SSH keys to inject into all the nodes"
}

variable "caasp_registry_code" {
  default     = ""
  description = "SUSE CaaSP Product Registration Code"
}

variable "firmware" {
  default     = "bios"
  description = "Firmware interface to use"
}

variable "guest_id" {
  default     = "sles15_64Guest"
  description = "Guest ID of the virtual machine"
}

variable "ntp_servers" {
  type        = list(string)
  default     = []
  description = "List of ntp servers to configure"
}

variable "packages" {
  type        = list(string)
  default     = []
  description = "List of additional packages to install"
}

variable "repositories" {
  type        = map(string)
  default     = {}
  description = "URLs of the repositories to mount via cloud-init"
}

variable "rmt_server_name" {
  default     = ""
  description = "SUSE Repository Mirroring Server Name"
}

variable "username" {
  default     = "sles"
  description = "Default user for the cluster nodes created by cloud-init default configuration for all SUSE SLES systems"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "workers" {
  default     = 1
  description = "Number of worker nodes"
}

variable "worker_cpus" {
  default     = 4
  description = "Number of CPUs used on worker node"
}

variable "worker_memory" {
  default     = 8192
  description = "Amount of memory used on worker node"
}

variable "worker_disk_size" {
  default     = 40
  description = "Size of the root disk in GB on worker node"
}

variable "master_cpus" {
  default     = 4
  description = "Number of CPUs used on master node"
}

variable "master_memory" {
  default     = 8192
  description = "Amount of memory used on master node"
}

variable "master_disk_size" {
  default     = 50
  description = "Size of the root disk in GB on master node"
}

variable "cpi_enable" {
  default     = false
  description = "Enable CPI integration with vSphere"
}
