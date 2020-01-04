data "template_file" "lb_repositories" {
  count    = "${length(var.lb_repositories)}"
  template = "${file("cloud-init/repository.tpl")}"

  vars {
    repository_url  = "${element(values(var.lb_repositories), count.index)}"
    repository_name = "${element(keys(var.lb_repositories), count.index)}"
  }
}

data "template_file" "haproxy_apiserver_backends_master" {
  count    = "${var.masters}"
  template = "server $${fqdn} $${ip}:6443\n"

  vars = {
    fqdn = "${var.stack_name}-master-${count.index}.${var.dns_domain}"
    ip   = "${cidrhost(var.network_cidr, 512 + count.index)}"
  }
}

data "template_file" "haproxy_gangway_backends_master" {
  count    = "${var.masters}"
  template = "server $${fqdn} $${ip}:32001\n"

  vars = {
    fqdn = "${var.stack_name}-master-${count.index}.${var.dns_domain}"
    ip   = "${cidrhost(var.network_cidr, 512 + count.index)}"
  }
}

data "template_file" "haproxy_dex_backends_master" {
  count    = "${var.masters}"
  template = "server $${fqdn} $${ip}:32000\n"

  vars = {
    fqdn = "${var.stack_name}-master-${count.index}.${var.dns_domain}"
    ip   = "${cidrhost(var.network_cidr, 512 + count.index)}"
  }
}

data "template_file" "lb_haproxy_cfg" {
  template = "${file("cloud-init/haproxy.cfg.tpl")}"

  vars {
    apiserver_backends = "${join("  ", data.template_file.haproxy_apiserver_backends_master.*.rendered)}"
    gangway_backends   = "${join("  ", data.template_file.haproxy_gangway_backends_master.*.rendered)}"
    dex_backends       = "${join("  ", data.template_file.haproxy_dex_backends_master.*.rendered)}"
  }
}

data "template_file" "lb_cloud_init_userdata" {
  template = "${file("cloud-init/lb.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", data.template_file.lb_repositories.*.rendered)}"
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
  name = "${var.stack_name}-lb-cloudinit-disk"
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

resource "null_resource" "lb_push_haproxy_cfg" {
  depends_on = ["null_resource.lb_wait_cloudinit"]
  count      = "${var.lbs}"

  triggers = {
    master_count = "${var.masters}"
  }

  connection {
    host  = "${element(libvirt_domain.lb.*.network_interface.0.addresses.0, count.index)}"
    user  = "${var.username}"
    type  = "ssh"
    agent = true
  }

  provisioner "file" {
    content     = "${data.template_file.lb_haproxy_cfg.rendered}"
    destination = "/tmp/haproxy.cfg"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/haproxy.cfg /etc/haproxy/haproxy.cfg",
      "sudo systemctl enable haproxy && sudo systemctl restart haproxy",
    ]
  }
}
