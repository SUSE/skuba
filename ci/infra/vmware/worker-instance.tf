data "template_file" "worker_repositories" {
  template = "${file("cloud-init/repository.tpl")}"
  count    = "${length(var.repositories)}"

  vars {
    repository_url  = "${element(values(var.repositories[count.index]), 0)}"
    repository_name = "${element(keys(var.repositories[count.index]), 0)}"
  }
}

data "template_file" "worker-cloud-init" {
  template = "${file("cloud-init/worker.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.worker_repositories.*.rendered)}"
    packages        = "${join("\n", formatlist("  - %s", var.packages))}"
    username        = "${var.username}"
    password        = "${var.password}"
    ntp_servers     = "${join("\n", formatlist ("    - %s", var.ntp_servers))}"
  }
}

resource "vsphere_virtual_machine" "worker" {
  count            = "${var.workers}"
  name             = "${var.stack_name}-worker-${count.index}"
  num_cpus         = "${var.worker_cpus}"
  memory           = "${var.worker_memory}"
  guest_id         = "sles12_64Guest"
  scsi_type        = "lsilogic"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  network_interface {
    network_id = "${data.vsphere_network.network.id}"
  }

  disk {
    label        = "disk0"
    datastore_id = "${data.vsphere_datastore.datastore.id}"
    size         = "${data.vsphere_virtual_machine.template.disks.0.size}"
  }

  cdrom {
    datastore_id = "${data.vsphere_datastore.datastore.id}"
    path         = "${var.stack_name}/cc-worker.iso"
  }

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  depends_on = ["vsphere_file.upload_cc_worker_iso"]
}

resource "null_resource" "worker_wait_cloudinit" {
  count = "${var.workers}"

  connection {
    host     = "${element(vsphere_virtual_machine.worker.*.guest_ip_addresses.0, count.index)}"
    user     = "${var.username}"
    password = "${var.password}"
    type     = "ssh"
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait",
    ]
  }
}

resource "null_resource" "local_gen-cc-worker-iso" {
  provisioner "local-exec" {
    command = "./gen-cloud-init-iso.sh worker '${data.template_file.worker-cloud-init.rendered}'"
  }
}

resource "vsphere_file" "upload_cc_worker_iso" {
  datacenter         = "${data.vsphere_datacenter.dc.name}"
  datastore          = "${data.vsphere_datastore.datastore.name}"
  source_file        = "./cc-worker.iso"
  create_directories = true
  destination_file   = "${var.stack_name}/cc-worker.iso"
  depends_on         = ["null_resource.local_gen-cc-worker-iso"]
}
