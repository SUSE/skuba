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
    password        = var.password
    ntp_servers     = join("\n", formatlist("    - %s", var.ntp_servers))
  }
}

resource "libvirt_volume" "worker" {
  name           = "${var.stack_name}-worker-volume-${count.index}"
  pool           = var.pool
  size           = var.disk_size
  base_volume_id = libvirt_volume.img.id
  count          = var.workers
}

resource "libvirt_cloudinit_disk" "worker" {
  # needed when 0 worker nodes are defined
  count     = var.workers
  name      = "${var.stack_name}-worker-cloudinit-disk-${count.index}"
  pool      = var.pool
  user_data = data.template_file.worker-cloud-init.rendered
}

resource "libvirt_domain" "worker" {
  count      = var.workers
  name       = "${var.stack_name}-worker-domain-${count.index}"
  memory     = var.worker_memory
  vcpu       = var.worker_vcpu
  cloudinit  = element(libvirt_cloudinit_disk.worker.*.id, count.index)
  depends_on = [libvirt_domain.lb]

  cpu = {
    mode = "host-passthrough"
  }

  disk {
    volume_id = element(libvirt_volume.worker.*.id, count.index)
  }

  network_interface {
    network_id     = libvirt_network.network.id
    hostname       = "${var.stack_name}-worker-${count.index}"
    addresses      = [cidrhost(var.network_cidr, 768 + count.index)]
    wait_for_lease = true
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }
}

resource "null_resource" "worker_wait_cloudinit" {
  depends_on = [libvirt_domain.worker]
  count      = var.workers

  connection {
    host = element(
      libvirt_domain.worker.*.network_interface.0.addresses.0,
      count.index,
    )
    user     = var.username
    password = var.password
    type     = "ssh"
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
        libvirt_domain.worker.*.network_interface.0.addresses.0,
        count.index,
      )
    }

    command = <<EOT
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $user@$host sudo reboot || :
# wait for ssh ready after reboot
until nc -zv $host 22; do sleep 5; done
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -oConnectionAttempts=60 $user@$host /usr/bin/true
EOT

  }
}

