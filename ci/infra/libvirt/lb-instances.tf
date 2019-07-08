data "template_file" "lb_repositories" {
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
    fqdn = "${var.stack_name}-master-${count.index}.${var.dns_domain}"
    ip   = "${cidrhost(var.network_cidr, 512 + count.index)}"
  }
}

data "template_file" "lb_cloud_init_userdata" {
  count    = "${var.lbs}"
  template = "${file("cloud-init/lb.tpl")}"

  vars {
    backends        = "${join("      ", data.template_file.haproxy_backends_master.*.rendered)}"
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.lb_repositories.*.rendered)}"
    packages        = "${join("\n", formatlist("  - %s", var.packages))}"
    username        = "${var.username}"
    password        = "${var.password}"
    ntp_servers     = "${join("\n", formatlist ("    - %s", var.ntp_servers))}"
  }
}

resource "libvirt_volume" "lb" {
  name           = "${var.stack_name}-lb-volume"
  pool           = "${var.pool}"
  size           = "${var.disk_size}"
  base_volume_id = "${libvirt_volume.img.id}"
}

resource "libvirt_cloudinit_disk" "lb" {
  name = "${var.stack_name}-lib-cloudinit-disk"
  pool = "${var.pool}"

  user_data = "${data.template_file.lb_cloud_init_userdata.rendered}"
}

resource "libvirt_domain" "lb" {
  name      = "${var.stack_name}-lb-domain"
  memory    = "${var.lb_memory}"
  vcpu      = "${var.lb_vcpu}"
  cloudinit = "${libvirt_cloudinit_disk.lb.id}"

  cpu {
    mode = "host-passthrough"
  }

  disk {
    volume_id = "${libvirt_volume.lb.id}"
  }

  network_interface {
    network_id     = "${libvirt_network.network.id}"
    hostname       = "${var.stack_name}-lb"
    addresses      = ["${cidrhost(var.network_cidr, 256)}"]
    wait_for_lease = 1
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }
}

resource "null_resource" "lb_wait_cloudinit" {
  depends_on = ["libvirt_domain.lb"]
  count      = "${var.lbs}"

  connection {
    host     = "${element(libvirt_domain.lb.*.network_interface.0.addresses.0, count.index)}"
    user     = "${var.username}"
    password = "${var.password}"
    type     = "ssh"
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null",
    ]
  }
}

resource "null_resource" "lb_reboot" {
  depends_on = ["null_resource.lb_wait_cloudinit"]
  count      = "${var.lbs}"

  provisioner "local-exec" {
    environment = {
      user = "${var.username}"
      host = "${element(libvirt_domain.lb.*.network_interface.0.addresses.0, count.index)}"
    }

    command = <<EOT
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $user@$host sudo reboot || :
# wait for ssh ready after reboot
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -oConnectionAttempts=60 $user@$host /usr/bin/true
EOT
  }
}
