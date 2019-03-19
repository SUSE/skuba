data "template_file" "master-cloud-init" {
  template = "${file("cloud-init/master.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repo_baseurl = "${var.repo_baseurl}"
  }
}

resource "openstack_compute_instance_v2" "master" {
  count      = "${var.masters}"
  name       = "ag-master-${var.stack_name}-${count.index}"
  image_name = "${var.image_name}"

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
