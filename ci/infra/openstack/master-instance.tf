data "template_file" "master_repositories" {
  template = "${file("cloud-init/repository.tpl")}"
  count    = "${length(var.repositories)}"

  vars {
    repository_url  = "${element(values(var.repositories[count.index]), 0)}"
    repository_name = "${element(keys(var.repositories[count.index]), 0)}"
  }
}

data "template_file" "master-cloud-init" {
  template = "${file("cloud-init/master.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.master_repositories.*.rendered)}"
    packages = "${join("\n", formatlist("  - %s", var.packages))}"
    username = "${var.username}"
    password = "${var.password}"
  }
}

resource "openstack_compute_instance_v2" "master" {
  count      = "${var.masters}"
  name       = "caasp-master-${var.stack_name}-${count.index}"
  image_name = "${var.image_name}"
  depends_on = [
    "openstack_networking_network_v2.network",
    "openstack_networking_subnet_v2.subnet"
  ]
  flavor_name = "${var.master_size}"

  network {
    name = "${var.internal_net}"
  }

  security_groups = [
    "${openstack_compute_secgroup_v2.secgroup_base.name}",
    "${openstack_compute_secgroup_v2.secgroup_master.name}",
  ]

  user_data = "${data.template_file.master-cloud-init.rendered}"
}

resource "openstack_networking_floatingip_v2" "master_ext" {
  count = "${var.masters}"
  pool  = "${var.external_net}"
}

resource "openstack_compute_floatingip_associate_v2" "master_ext_ip" {
  count       = "${var.masters}"
  floating_ip = "${element(openstack_networking_floatingip_v2.master_ext.*.address, count.index)}"
  instance_id = "${element(openstack_compute_instance_v2.master.*.id, count.index)}"
}

resource "null_resource" "master_wait_cloudinit" {
  count = "${var.masters}"
  connection {
    host     = "${element(openstack_compute_floatingip_associate_v2.master_ext_ip.*.floating_ip, count.index)}"
    user     = "${var.username}"
    password = "${var.password}"
    type     = "ssh"
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait"
    ]
  }
}
