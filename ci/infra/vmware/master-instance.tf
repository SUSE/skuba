data "template_file" "master_repositories" {
  template = file("cloud-init/repository.tpl")
  count    = length(var.repositories)

  vars = {
    repository_url  = element(values(var.repositories), count.index)
    repository_name = element(keys(var.repositories), count.index)
  }
}

data "template_file" "master_register_scc" {
  template = file("cloud-init/register-scc.tpl")
  count    = var.caasp_registry_code == "" ? 0 : 1

  vars = {
    caasp_registry_code = var.caasp_registry_code
  }
}

data "template_file" "master_register_rmt" {
  template = file("cloud-init/register-rmt.tpl")
  count    = var.rmt_server_name == "" ? 0 : 1

  vars = {
    rmt_server_name = var.rmt_server_name
  }
}

data "template_file" "master_commands" {
  template = file("cloud-init/commands.tpl")
  count    = join("", var.packages) == "" ? 0 : 1

  vars = {
    packages = join(", ", var.packages)
  }
}

data "template_file" "master_cloud_init_metadata" {
  template = file("cloud-init/metadata.tpl")

  vars = {
    network_config = base64gzip(data.local_file.network_cloud_init.content)
    instance_id    = "${var.stack_name}-master"
  }
}

data "template_file" "master_cloud_init_userdata" {
  template = file("cloud-init/common.tpl")
  count    = var.masters

  vars = {
    authorized_keys    = join("\n", formatlist("  - %s", var.authorized_keys))
    repositories       = join("\n", data.template_file.master_repositories.*.rendered)
    register_scc       = join("\n", data.template_file.master_register_scc.*.rendered)
    register_rmt       = join("\n", data.template_file.master_register_rmt.*.rendered)
    commands           = join("\n", data.template_file.master_commands.*.rendered)
    ntp_servers        = join("\n", formatlist("    - %s", var.ntp_servers))
    hostname           = "${var.stack_name}-master-${count.index}"
    hostname_from_dhcp = var.hostname_from_dhcp == true ? "yes" : "no"
  }
}

resource "vsphere_virtual_machine" "master" {
  depends_on = [vsphere_folder.folder]

  count                = var.masters
  name                 = "${var.stack_name}-master-${count.index}"
  num_cpus             = var.master_cpus
  memory               = var.master_memory
  guest_id             = var.guest_id
  firmware             = var.firmware
  scsi_type            = data.vsphere_virtual_machine.template.scsi_type
  resource_pool_id     = data.vsphere_resource_pool.pool.id
  datastore_id         = (var.vsphere_datastore == null ? null : data.vsphere_datastore.datastore[0].id)
  datastore_cluster_id = (var.vsphere_datastore_cluster == null ? null : data.vsphere_datastore_cluster.datastore[0].id)
  folder               = var.cpi_enable == true ? vsphere_folder.folder[0].path : null

  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
  }

  hardware_version = var.vsphere_hardware_version

  disk {
    label            = "disk0"
    size             = var.master_disk_size
    eagerly_scrub    = data.vsphere_virtual_machine.template.disks.0.eagerly_scrub
    thin_provisioned = data.vsphere_virtual_machine.template.disks.0.thin_provisioned
  }

  extra_config = {
    "guestinfo.metadata"          = base64gzip(data.template_file.master_cloud_init_metadata.rendered)
    "guestinfo.metadata.encoding" = "gzip+base64"
    "guestinfo.userdata"          = base64gzip(data.template_file.master_cloud_init_userdata[count.index].rendered)
    "guestinfo.userdata.encoding" = "gzip+base64"
  }
  enable_disk_uuid = var.cpi_enable == true ? true : false

  network_interface {
    network_id = data.vsphere_network.network.id
  }
}

resource "null_resource" "master_wait_cloudinit" {
  depends_on = [vsphere_virtual_machine.master]
  count      = var.masters

  connection {
    host = element(
      vsphere_virtual_machine.master.*.guest_ip_addresses.0,
      count.index,
    )
    user  = var.username
    type  = "ssh"
    agent = true
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null",
    ]
  }
}
