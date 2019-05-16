data "template_file" "worker_repositories" {
  template = "${file("cloud-init/repository.tpl")}"
  count    = "${length(var.repositories)}"

  vars {
    repository_url  = "${element(values(var.repositories[count.index]), 0)}"
    repository_name = "${element(keys(var.repositories[count.index]), 0)}"
  }
}

data "template_file" "worker_registration" {
  template = "${file("cloud-init/registration.tpl")}"
  count    = "${var.caasp_registry_code == "" ? 0 : 1}"

  vars {
    caasp_registry_code = "${var.caasp_registry_code}"
    packages            = "${join(", ", var.packages)}"
  }
}

data "template_file" "worker_commands" {
  template = "${file("cloud-init/commands.tpl")}"

  vars {
    packages = "${join(", ", var.packages)}"
  }
}

data "template_file" "worker-cloud-init" {
  template = "${file("cloud-init/worker.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.worker_repositories.*.rendered)}"
    registration    = "${join("\n", data.template_file.worker_registration.*.rendered)}"
    commands        = "${join("\n", data.template_file.worker_commands.*.rendered)}"
    username        = "${var.username}"
    password        = "${var.password}"
    ntp_servers     = "${join("\n", formatlist ("    - %s", var.ntp_servers))}"
  }
}

resource "openstack_blockstorage_volume_v2" "worker_vol" {
  count = "${var.workers_vol_enabled ? "${var.workers}" : 0 }"
  size  = "${var.workers_vol_size}"
  name  = "vol_${element(openstack_compute_instance_v2.worker.*.name, count.index)}"
}

resource "openstack_compute_volume_attach_v2" "worker_vol_attach" {
  count       = "${var.workers_vol_enabled ? "${var.workers}" : 0 }"
  instance_id = "${element(openstack_compute_instance_v2.worker.*.id, count.index)}"
  volume_id   = "${element(openstack_blockstorage_volume_v2.worker_vol.*.id, count.index)}"
}

resource "openstack_compute_instance_v2" "worker" {
  count      = "${var.workers}"
  name       = "caasp-worker-${var.stack_name}-${count.index}"
  image_name = "${var.image_name}"

  depends_on = [
    "openstack_networking_network_v2.network",
    "openstack_networking_subnet_v2.subnet",
  ]

  flavor_name = "${var.worker_size}"

  network {
    name = "${var.internal_net}"
  }

  security_groups = [
    "${openstack_compute_secgroup_v2.secgroup_base.name}",
    "${openstack_compute_secgroup_v2.secgroup_worker.name}",
  ]

  user_data = "${data.template_file.worker-cloud-init.rendered}"
}

resource "openstack_networking_floatingip_v2" "worker_ext" {
  count = "${var.workers}"
  pool  = "${var.external_net}"
}

resource "openstack_compute_floatingip_associate_v2" "worker_ext_ip" {
  count       = "${var.workers}"
  floating_ip = "${element(openstack_networking_floatingip_v2.worker_ext.*.address, count.index)}"
  instance_id = "${element(openstack_compute_instance_v2.worker.*.id, count.index)}"
}

resource "null_resource" "worker_wait_cloudinit" {
  count = "${var.workers}"

  connection {
    host     = "${element(openstack_compute_floatingip_associate_v2.worker_ext_ip.*.floating_ip, count.index)}"
    user     = "${var.username}"
    password = "${var.password}"
    type     = "ssh"
  }

  depends_on = ["openstack_compute_instance_v2.worker"]

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait",
    ]
  }
}
