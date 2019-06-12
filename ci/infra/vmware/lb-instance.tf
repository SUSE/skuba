variable "lbs" {
  default     = 1
  description = "Number of load-balancer nodes"
}

variable "lb_cpus" {
  default     = 1
  description = "Number of CPUs used on load-balancer node"
}

variable "lb_memory" {
  default     = 2048
  description = "Amount of memory used on load-balancer node"
}

variable "lb_repositories" {
  type        = "map"
  default     = {
    sle_server_pool     = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Products/SLE-Product-SLES/15-SP1/x86_64/product/"
    basesystem_pool     = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Products/SLE-Module-Basesystem/15-SP1/x86_64/product/"
    ha_pool             = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Products/SLE-Module-HA/15/x86_64/product/"
    sle_server_updates  = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Updates/SLE-Product-SLES/15-SP1/x86_64/update/"
    basesystem_updates  = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Updates/SLE-Module-Basesystem/15-SP1/x86_64/update/"
  }
}

data "template_file" "lb_repositories_template" {
  count    = "${length(var.lb_repositories)}"
  template = "${file("cloud-init/repository.tpl")}"

  vars {
    repository_url  = "${element(values(var.lb_repositories), count.index)}"
    repository_name = "${element(keys(var.lb_repositories), count.index)}"
  }
}

data "template_file" "haproxy_backends_master" {
  count    = "${var.masters}"
  template = "server $${fqdn} $${ip}:6443 check check-ssl verify none\n"

  vars = {
    fqdn = "${element(vsphere_virtual_machine.master.*.name, count.index)}"
    ip   = "${element(vsphere_virtual_machine.master.*.default_ip_address, count.index)}"
  }

  depends_on = [
    "vsphere_virtual_machine.master"
  ]
}

data "template_file" "lb_cloud_init_metadata" {
  count    = "${var.lbs}"
  template = "${file("cloud-init/metadata.tpl")}"

  vars {
    network_config = "${base64gzip(data.local_file.network_cloud_init.content)}"
    instance_id    = "${var.stack_name}-lb"
  }
}

data "template_file" "lb_cloud_init_userdata" {
  count    = "${var.lbs}"
  template = "${file("cloud-init/lb.tpl")}"

  vars {
    backends        = "${join("      ", data.template_file.haproxy_backends_master.*.rendered)}"
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.lb_repositories_template.*.rendered)}"
    packages        = "${join("\n", formatlist("  - %s", var.packages))}"
    ntp_servers     = "${join("\n", formatlist ("    - %s", var.ntp_servers))}"
  }
}

resource "vsphere_virtual_machine" "lb" {
  count            = "${var.lbs}"
  name             = "${var.stack_name}-lb-${count.index}"
  num_cpus         = "${var.lb_cpus}"
  memory           = "${var.lb_memory}"
  guest_id         = "${var.guest_id}"
  scsi_type        = "${data.vsphere_virtual_machine.template.scsi_type}"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  disk {
    label        = "disk0"
    datastore_id = "${data.vsphere_datastore.datastore.id}"
    size         = "${data.vsphere_virtual_machine.template.disks.0.size}"
  }

  extra_config {
    "guestinfo.metadata"          = "${base64gzip(data.template_file.lb_cloud_init_metadata.rendered)}"
    "guestinfo.metadata.encoding" = "gzip+base64"

    "guestinfo.userdata"          = "${base64gzip(data.template_file.lb_cloud_init_userdata.rendered)}"
    "guestinfo.userdata.encoding" = "gzip+base64"
  }

  network_interface {
    network_id = "${data.vsphere_network.network.id}"
  }

  depends_on = [
    "vsphere_virtual_machine.master"
  ]
}

resource "null_resource" "lb_wait_cloudinit" {
  count = "${var.lbs}"

  connection {
    host     = "${element(vsphere_virtual_machine.lb.*.guest_ip_addresses.0, count.index)}"
    user     = "${var.username}"
    type     = "ssh"
    agent    = true
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait",
    ]
  }
}
