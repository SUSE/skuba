locals {
  worker_repositories = [for i in range(length(var.repositories)) : templatefile("cloud-init/repository.tpl",
    {
      repository_url  = "${element(values(var.repositories), i)}",
      repository_name = "${element(keys(var.repositories), i)}"
  })]

  worker_register_scc = [for c in(var.caasp_registry_code == "" ? [] : [1]) : templatefile("cloud-init/register-scc.tpl", {
    caasp_registry_code = "${var.caasp_registry_code}"
  })]

  worker_register_rmt = [for c in(var.rmt_server_name == "" ? [] : [1]) : templatefile("cloud-init/register-rmt.tpl", {
    rmt_server_name = "${var.rmt_server_name}"
  })]

  worker_commands = [for c in(var.packages == "" ? [] : [1]) : templatefile("cloud-init/commands.tpl", {
    packages = "${join(", ", var.packages)}"
  })]

  worker_cloud_init = templatefile("cloud-init/common.tpl", {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", local.worker_repositories.*)}"
    register_scc    = "${join("\n", local.worker_register_scc.*)}"
    register_rmt    = "${join("\n", local.worker_register_rmt.*)}"
    commands        = "${join("\n", local.worker_commands.*)}"
    username        = "${var.username}"
    ntp_servers     = "${join("\n", formatlist("    - %s", var.ntp_servers))}"
  })
}

resource "openstack_blockstorage_volume_v2" "worker_vol" {
  count = "${var.workers_vol_enabled ? "${var.workers}" : 0}"
  size  = "${var.workers_vol_size}"
  name  = "vol_${element(openstack_compute_instance_v2.worker.*.name, count.index)}"
}

resource "openstack_compute_volume_attach_v2" "worker_vol_attach" {
  count       = "${var.workers_vol_enabled ? "${var.workers}" : 0}"
  instance_id = "${element(openstack_compute_instance_v2.worker.*.id, count.index)}"
  volume_id   = "${element(openstack_blockstorage_volume_v2.worker_vol.*.id, count.index)}"
}

resource "openstack_compute_instance_v2" "worker" {
  count      = "${var.workers}"
  name       = "caasp-worker-${var.stack_name}-${count.index}"
  image_name = "${var.image_name}"
  key_pair   = "${var.key_pair}"

  depends_on = [
    "openstack_networking_network_v2.network",
    "openstack_networking_subnet_v2.subnet",
  ]

  flavor_name = "${var.worker_size}"

  network {
    name = "${var.internal_net}"
  }

  security_groups = [
    "${openstack_networking_secgroup_v2.common.name}",
  ]

  user_data = "${local.worker_cloud_init}"
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
  depends_on = ["openstack_compute_instance_v2.worker"]
  count      = "${var.workers}"

  connection {
    host = "${element(openstack_compute_floatingip_associate_v2.worker_ext_ip.*.floating_ip, count.index)}"
    user = "${var.username}"
    type = "ssh"
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null",
    ]
  }
}

resource "null_resource" "worker_reboot" {
  depends_on = ["null_resource.worker_wait_cloudinit"]
  count      = "${var.workers}"

  provisioner "local-exec" {
    environment = {
      user = "${var.username}"
      host = "${element(openstack_compute_floatingip_associate_v2.worker_ext_ip.*.floating_ip, count.index)}"
    }

    command = <<EOT
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $user@$host sudo reboot || :
# wait for ssh ready after reboot
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -oConnectionAttempts=60 $user@$host /usr/bin/true
EOT
  }
}
