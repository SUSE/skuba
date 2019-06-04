variable "template_name" {}
variable "stack_name" {}
variable "vsphere_datastore" {}
variable "vsphere_datacenter" {}
variable "vsphere_network" {}
variable "vsphere_resource_pool" {}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes"
}

variable "guest_id" {
  default     = "sles15_64Guest"
  description = "Guest ID of the virtual machine"
}

variable "ntp_servers" {
  type        = "list"
  default     = []
  description = "list of ntp servers to configure"
}

variable "packages" {
  type        = "list"
  default     = []
  description = "list of additional packages to install"
}

variable "repositories" {
  type        = "map"
  default     = {}
  description = "Urls of the repositories to mount via cloud-init"
}

variable "username" {
  default     = "sles"
  description = "Username for the cluster nodes"
}

variable "password" {
  default     = "sles"
  description = "Password for the cluster nodes"
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

variable "master_cpus" {
  default     = 4
  description = "Number of CPUs used on master node"
}

variable "master_memory" {
  default     = 8192
  description = "Amount of memory used on master node"
}

variable "caasp_registry_code" {
  default     = ""
  description = "SUSE CaaSP Product Registration Code"
}

variable "rmt_server_name" {
  default     = ""
  description = "SUSE Repository Mirroring Server Name"
}

#### To be moved to separate vsphere.tf? ####

provider "vsphere" {}

data "vsphere_resource_pool" "pool" {
  name          = "${var.vsphere_resource_pool}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_datastore" "datastore" {
  name          = "${var.vsphere_datastore}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_datacenter" "dc" {
  name = "${var.vsphere_datacenter}"
}

data "vsphere_network" "network" {
  name          = "${var.vsphere_network}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}

data "vsphere_virtual_machine" "template" {
  name          = "${var.template_name}"
  datacenter_id = "${data.vsphere_datacenter.dc.id}"
}
