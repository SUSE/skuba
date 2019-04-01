#######################
# Cluster declaration #
#######################

provider "lxd" {
  generate_client_certificates = true
  accept_remote_certificate    = true
}

locals {
  authorized_keys_file = <<EOT
${join("\n", formatlist("%s", var.authorized_keys))}
${file("${path.module}/../id_shared.pub")}
EOT
}

##############
# Base image #
##############

resource "null_resource" "base_image" {
  # make sure we have an opensuse-caasp image
  # if that is not the case, build one with the help of distrobuilder
  provisioner "local-exec" {
    command = "./support/build-image.sh --img '${var.img}' --force '${var.force_img}'"
  }
}

#####################
### Cluster masters #
#####################

resource "lxd_container" "master" {
  count      = "${var.master_count}"
  name       = "${var.name_prefix}master-${count.index}"
  image      = "${var.img}"
  depends_on = ["null_resource.base_image"]

  connection {
    type     = "ssh"
    user     = "${var.ssh_user}"
    password = "${var.ssh_pass}"
  }

  provisioner "file" {
    content     = "${local.authorized_keys_file}"
    destination = "/root/.ssh/authorized_keys"
  }
}

output "masters" {
  value = ["${lxd_container.master.*.ip_address}"]
}

####################
## Cluster workers #
####################

resource "lxd_container" "worker" {
  count      = "${var.worker_count}"
  name       = "${var.name_prefix}worker-${count.index}"
  image      = "${var.img}"
  depends_on = ["lxd_container.master"]

  connection {
    type     = "ssh"
    user     = "${var.ssh_user}"
    password = "${var.ssh_pass}"
  }

  provisioner "file" {
    content     = "${local.authorized_keys_file}"
    destination = "/root/.ssh/authorized_keys"
  }
}

output "workers" {
  value = ["${lxd_container.worker.*.ip_address}"]
}

######################
# Load Balancer node #
######################
data "template_file" "haproxy_backends_master" {
  count    = "${var.master_count}"
  template = "${file("templates/haproxy-backends.tpl")}"

  vars = {
    fqdn = "${var.name_prefix}master-${count.index}.${var.name_prefix}${var.domain_name}"
    ip   = "${element(lxd_container.master.*.ip_address, count.index)}"
  }
}

data "template_file" "haproxy_cfg" {
  template = "${file("templates/haproxy.cfg.tpl")}"

  vars = {
    backends = "${join("      ", data.template_file.haproxy_backends_master.*.rendered)}"
  }
}

resource "lxd_container" "lb" {
  name       = "${var.name_prefix}lb"
  image      = "${var.img}"
  depends_on = ["lxd_container.master"]

  connection {
    type     = "ssh"
    user     = "${var.ssh_user}"
    password = "${var.ssh_pass}"
  }

  provisioner "file" {
    content     = "${local.authorized_keys_file}"
    destination = "/root/.ssh/authorized_keys"
  }

  provisioner "file" {
    content     = "${data.template_file.haproxy_cfg.rendered}"
    destination = "/etc/haproxy/haproxy.cfg"
  }

  provisioner "remote-exec" {
    inline = "systemctl enable --now haproxy"
  }
}

output "ip_lb" {
  value = "${lxd_container.lb.ip_address}"
}
