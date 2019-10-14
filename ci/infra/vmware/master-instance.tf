locals {
  master_repositories = [for i in range(length(var.repositories)) : templatefile("cloud-init/repository.tpl", {
    repository_url  = "${element(values(var.repositories), i)}"
    repository_name = "${element(keys(var.repositories), i)}"
  })]

  master_register_scc = [for c in(var.caasp_registry_code == "" ? [] : [1]) : templatefile("cloud-init/register-scc.tpl", {
    caasp_registry_code = "${var.caasp_registry_code}"
  })]

  master_register_rmt = [for c in(var.rmt_server_name == "" ? [] : [1]) : templatefile("cloud-init/register-rmt.tpl", {
    rmt_server_name = "${var.rmt_server_name}"
  })]

  master_commands = [for _ in(join("", var.packages) == "" ? [] : [1]) : templatefile("cloud-init/commands.tpl", {
    packages = "${join(", ", var.packages)}"
  })]

  master_cloud_init_metadata = templatefile("cloud-init/metadata.tpl", {
    network_config = "${base64gzip(data.local_file.network_cloud_init.content)}"
    instance_id    = "${var.stack_name}-master"
  })

  master_cloud_init_userdata = templatefile("cloud-init/common.tpl", {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", local.master_repositories.*)}"
    register_scc    = "${join("\n", local.master_register_scc.*)}"
    register_rmt    = "${join("\n", local.master_register_rmt.*)}"
    commands        = "${join("\n", local.master_commands.*)}"
    ntp_servers     = "${join("\n", formatlist("    - %s", var.ntp_servers))}"
  })
}

resource "vsphere_virtual_machine" "master" {
  count            = "${var.masters}"
  name             = "${var.stack_name}-master-${count.index}"
  num_cpus         = "${var.master_cpus}"
  memory           = "${var.master_memory}"
  guest_id         = "${var.guest_id}"
  firmware         = "${var.firmware}"
  scsi_type        = "${data.vsphere_virtual_machine.template.scsi_type}"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  disk {
    label = "disk0"
    size  = "${var.master_disk_size}"
  }

  extra_config = {
    "guestinfo.metadata"          = "${base64gzip(local.master_cloud_init_metadata)}"
    "guestinfo.metadata.encoding" = "gzip+base64"

    "guestinfo.userdata"          = "${base64gzip(local.master_cloud_init_userdata)}"
    "guestinfo.userdata.encoding" = "gzip+base64"
  }

  network_interface {
    network_id = "${data.vsphere_network.network.id}"
  }
}

resource "null_resource" "master_wait_cloudinit" {
  count = "${var.masters}"

  connection {
    host  = "${element(vsphere_virtual_machine.master.*.guest_ip_addresses.0, count.index)}"
    user  = "${var.username}"
    type  = "ssh"
    agent = true
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null",
    ]
  }
}
