provider "vsphere" {
  version = "~> 1.17"
}

data "vsphere_resource_pool" "pool" {
  name          = var.vsphere_resource_pool
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_datastore" "datastore" {
  count         = var.vsphere_datastore == "null" || var.vsphere_datastore == null ? 0 : 1
  name          = var.vsphere_datastore
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_datastore_cluster" "datastore" {
  count         = var.vsphere_datastore_cluster == "null" || var.vsphere_datastore_cluster == null ? 0 : 1
  name          = var.vsphere_datastore_cluster
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_datacenter" "dc" {
  name = var.vsphere_datacenter
}

data "vsphere_network" "network" {
  name          = var.vsphere_network
  datacenter_id = data.vsphere_datacenter.dc.id
}

data "vsphere_virtual_machine" "template" {
  name          = var.template_name
  datacenter_id = data.vsphere_datacenter.dc.id
}

resource "vsphere_folder" "folder" {
  count         = var.cpi_enable == true ? 1 : 0
  path          = "${var.stack_name}-cluster"
  type          = "vm"
  datacenter_id = data.vsphere_datacenter.dc.id
}
