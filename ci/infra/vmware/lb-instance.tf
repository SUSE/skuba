variable "load-balancers" {
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


data "template_file" "lb_repositories" {
  count    = "${length(var.repositories)}"
  template = "${file("cloud-init/repository.tpl")}"

  vars {
    repository_url  = "${element(values(var.repositories[count.index]), 0)}"
    repository_name = "${element(keys(var.repositories[count.index]), 0)}"
  }
}

data "template_file" "haproxy_backends_master" {
  count    = "${var.masters}"
  template = "${file("cloud-init/haproxy-backends.tpl")}"

  vars = {
    fqdn = "${var.stack_name}-master-${count.index}"
    ip   = "${element(vsphere_virtual_machine.master.*.default_ip_address, count.index)}"
  }
}

data "template_file" "lb_cloud_init_metadata" {
  count    = "${var.load-balancers}"
  template = "${file("cloud-init/metadata.tpl")}"

  vars {
    network_config = "${base64gzip(data.local_file.network_cloud_init.content)}"
    instance_id    = "${var.stack_name}-lb"
  }
}

data "template_file" "lb_cloud_init_userdata" {
  count    = "${var.load-balancers}"
  template = "${file("cloud-init/lb.tpl")}"

  vars {
    backends        = "${join("      ", data.template_file.haproxy_backends_master.*.rendered)}"
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.lb_repositories.*.rendered)}"
    packages        = "${join("\n", formatlist("  - %s", var.packages))}"
    ntp_servers     = "${join("\n", formatlist ("    - %s", var.ntp_servers))}"
  }
}

resource "vsphere_virtual_machine" "lb" {
  count                = "${var.load-balancers}"
  name                 = "${var.stack_name}-lb-${count.index}"
  num_cpus             = "${var.lb_cpus}"
  memory               = "${var.lb_memory}"
  guest_id             = "${var.guest_id}"
  scsi_type            = "${data.vsphere_virtual_machine.template.scsi_type}"
  resource_pool_id     = "${data.vsphere_resource_pool.pool.id}"
  datastore_cluster_id = "${data.vsphere_datastore_cluster.datastore_cluster.id}"

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  disk {
    label        = "disk0"
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
}

resource "null_resource" "lb_wait_cloudinit" {
  count = "${var.load-balancers}"

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
