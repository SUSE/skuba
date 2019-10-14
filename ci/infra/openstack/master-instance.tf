locals {
  master_repositories = [for i in range(length(var.repositories)) : templatefile("cloud-init/repository.tpl",
    {
      repository_url  = "${element(values(var.repositories), i)}",
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

  master_cloud_init = templatefile("cloud-init/common.tpl", {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", local.master_repositories.*)}"
    register_scc    = "${join("\n", local.master_register_scc.*)}"
    register_rmt    = "${join("\n", local.master_register_rmt.*)}"
    commands        = "${join("\n", local.master_commands.*)}"
    username        = "${var.username}"
    ntp_servers     = "${join("\n", formatlist("    - %s", var.ntp_servers))}"
  })
}

resource "openstack_compute_instance_v2" "master" {
  count      = "${var.masters}"
  name       = "caasp-master-${var.stack_name}-${count.index}"
  image_name = "${var.image_name}"
  key_pair   = "${var.key_pair}"

  depends_on = [
    "openstack_networking_network_v2.network",
    "openstack_networking_subnet_v2.subnet",
  ]

  flavor_name = "${var.master_size}"

  network {
    name = "${var.internal_net}"
  }

  security_groups = [
    "${openstack_networking_secgroup_v2.common.name}",
    "${openstack_networking_secgroup_v2.master_nodes.name}",
  ]

  user_data = "${local.master_cloud_init}"
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
  depends_on = ["openstack_compute_instance_v2.master"]
  count      = "${var.masters}"

  connection {
    host = "${element(openstack_compute_floatingip_associate_v2.master_ext_ip.*.floating_ip, count.index)}"
    user = "${var.username}"
    type = "ssh"
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null",
    ]
  }
}

resource "null_resource" "master_reboot" {
  depends_on = ["null_resource.master_wait_cloudinit"]
  count      = "${var.masters}"

  provisioner "local-exec" {
    environment = {
      user = "${var.username}"
      host = "${element(openstack_compute_floatingip_associate_v2.master_ext_ip.*.floating_ip, count.index)}"
    }

    command = <<EOT
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $user@$host sudo reboot || :
# wait for ssh ready after reboot
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -oConnectionAttempts=60 $user@$host /usr/bin/true
EOT
  }
}
