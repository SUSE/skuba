data "template_file" "worker_repositories" {
  template = file("cloud-init/repository.tpl")
  count    = length(var.repositories)

  vars = {
    repository_url  = element(values(var.repositories), count.index)
    repository_name = element(keys(var.repositories), count.index)
  }
}

data "template_file" "worker_register_scc" {
  template = file("cloud-init/register-scc.tpl")
  count    = var.caasp_registry_code == "" ? 0 : 1

  vars = {
    caasp_registry_code = var.caasp_registry_code
  }
}

data "template_file" "worker_register_rmt" {
  template = file("cloud-init/register-rmt.tpl")
  count    = var.rmt_server_name == "" ? 0 : 1

  vars = {
    rmt_server_name = var.rmt_server_name
  }
}

data "template_file" "worker_commands" {
  template = file("cloud-init/commands.tpl")
  count    = join("", var.packages) == "" ? 0 : 1

  vars = {
    packages = join(", ", var.packages)
  }
}

data "template_file" "worker-cloud-init" {
  template = file("cloud-init/common.tpl")

  vars = {
    authorized_keys = join("\n", formatlist("  - %s", var.authorized_keys))
    repositories    = join("\n", data.template_file.worker_repositories.*.rendered)
    register_scc    = join("\n", data.template_file.worker_register_scc.*.rendered)
    register_rmt    = join("\n", data.template_file.worker_register_rmt.*.rendered)
    commands        = join("\n", data.template_file.worker_commands.*.rendered)
    username        = var.username
    ntp_servers     = join("\n", formatlist("    - %s", var.ntp_servers))
  }
}

resource "openstack_blockstorage_volume_v2" "worker_vol" {
  count = var.workers_vol_enabled ? var.workers : 0
  size  = var.workers_vol_size
  name  = "vol_${element(openstack_compute_instance_v2.worker.*.name, count.index)}"
}

resource "openstack_compute_volume_attach_v2" "worker_vol_attach" {
  count       = var.workers_vol_enabled ? var.workers : 0
  instance_id = element(openstack_compute_instance_v2.worker.*.id, count.index)
  volume_id = element(
    openstack_blockstorage_volume_v2.worker_vol.*.id,
    count.index,
  )
}

resource "openstack_compute_instance_v2" "worker" {
  count      = var.workers
  name       = "caasp-worker-${var.stack_name}-${count.index}"
  image_name = var.image_name
  key_pair   = var.key_pair

  depends_on = [
    openstack_networking_network_v2.network,
    openstack_networking_subnet_v2.subnet,
  ]

  flavor_name = var.worker_size

  network {
    name = var.internal_net
  }

  security_groups = [
    openstack_networking_secgroup_v2.common.name,
  ]

  user_data = data.template_file.worker-cloud-init.rendered
}

resource "openstack_networking_floatingip_v2" "worker_ext" {
  count = var.workers
  pool  = var.external_net
}

resource "openstack_compute_floatingip_associate_v2" "worker_ext_ip" {
  depends_on = [openstack_compute_instance_v2.worker]
  count = var.workers
  floating_ip = element(
    openstack_networking_floatingip_v2.worker_ext.*.address,
    count.index,
  )
  instance_id = element(openstack_compute_instance_v2.worker.*.id, count.index)
}

resource "null_resource" "worker_wait_cloudinit" {
  depends_on = [
    openstack_compute_instance_v2.worker,
    openstack_compute_floatingip_associate_v2.worker_ext_ip,
  ]
  count      = var.workers

  connection {
    host = element(
      openstack_compute_floatingip_associate_v2.worker_ext_ip.*.floating_ip,
      count.index,
    )
    user = var.username
    type = "ssh"
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null",
    ]
  }
}

resource "null_resource" "worker_reboot" {
  depends_on = [null_resource.worker_wait_cloudinit]
  count      = var.workers

  provisioner "local-exec" {
    environment = {
      user = var.username
      host = element(
        openstack_compute_floatingip_associate_v2.worker_ext_ip.*.floating_ip,
        count.index,
      )
    }

    command = <<EOT
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $user@$host sudo reboot || :
# wait for ssh ready after reboot
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -oConnectionAttempts=60 $user@$host /usr/bin/true
EOT

  }
}

