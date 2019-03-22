#####################
# libvirt variables #
#####################

variable "libvirt_uri" {
  default     = "qemu:///system"
  description = "libvirt connection url - default to localhost"
}

variable "pool" {
  default     = "home_libvirtd"
  description = "pool to be used to store all the volumes"
}

#####################
# Cluster variables #
#####################

variable "img_source_url" {
  type        = "string"
  default     = "https://download.opensuse.org/repositories/Cloud:/Images:/Leap_15.0/images/openSUSE-Leap-15.0-OpenStack.x86_64-0.0.4-Buildlp150.12.127.qcow2"
}

variable "repo_baseurl" {
  type        = "string"
  default     = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
}

variable "lb_memory" {
  default     = 2048
  description = "The amount of RAM for a load balancer node"
}

variable "lb_vcpu" {
  default     = 1
  description = "The amount of virtual CPUs for a load balancer node"
}

variable "master_count" {
  default     = 1
  description = "Number of masters to be created"
}

variable "master_memory" {
  default     = 2048
  description = "The amount of RAM for a master"
}

variable "master_vcpu" {
  default     = 2
  description = "The amount of virtual CPUs for a master"
}

variable "worker_count" {
  default     = 2
  description = "Number of workers to be created"
}

variable "worker_memory" {
  default     = 2048
  description = "The amount of RAM for a worker"
}

variable "worker_vcpu" {
  default     = 2
  description = "The amount of virtual CPUs for a worker"
}

variable "name_prefix" {
  type        = "string"
  default     = "ag-"
  description = "Optional prefix to be able to have multiple clusters on one host"
}

variable "domain_name" {
  type        = "string"
  default     = "test.net"
  description = "The domain name"
}

variable "net_mode" {
  type        = "string"
  default     = "nat"
  description = "Network mode used by the cluster"
}

variable "network" {
  type        = "string"
  default     = "10.17.0.0/22"
  description = "Network used by the cluster"
}

#######################
# Cluster declaration #
#######################

provider "libvirt" {
  uri = "${var.libvirt_uri}"
}

# This is the CaaSP kvm image that has been created by IBS
resource "libvirt_volume" "img" {
  name   = "${var.name_prefix}${basename(var.img_source_url)}"
  source = "${var.img_source_url}"
  pool   = "${var.pool}"
}

##############
# Networking #
##############
resource "libvirt_network" "network" {
    name      = "${var.name_prefix}net"
    mode      = "${var.net_mode}"
    domain    = "${var.name_prefix}${var.domain_name}"
    addresses = ["${var.network}"]
}

######################
# Load Balancer node #
######################
resource "libvirt_volume" "lb" {
  name           = "${var.name_prefix}lb.qcow2"
  pool           = "${var.pool}"
  base_volume_id = "${libvirt_volume.img.id}"
}

data "template_file" "haproxy_backends_master" {
  count    = "${var.master_count}"
  template = "${file("cloud-init/haproxy-backends.tpl")}"
  vars = {
    fqdn = "${var.name_prefix}master-${count.index}.${var.name_prefix}${var.domain_name}"
    ip = "${cidrhost("${var.network}", 512 + count.index)}"
  }
}

data "template_file" "lb_cloud_init_user_data" {
  template = "${file("cloud-init/lb.cfg.tpl")}"
  vars = {
    hostname = "${var.name_prefix}lb-${count.index}"
    fqdn = "${var.name_prefix}lb-${count.index}.${var.name_prefix}${var.domain_name}"
    backends = "${join("      ", data.template_file.haproxy_backends_master.*.rendered)}"
  }
}

resource "libvirt_cloudinit_disk" "lb" {
  name      = "${var.name_prefix}lb_cloud_init.iso"
  pool      = "${var.pool}"

  user_data = "${data.template_file.lb_cloud_init_user_data.rendered}"
}

resource "libvirt_domain" "lb" {
  name      = "${var.name_prefix}lb"
  memory    = "${var.lb_memory}"
  vcpu      = "${var.lb_vcpu}"
  metadata  = "lb.${var.domain_name},lb,${count.index}"
  cloudinit = "${libvirt_cloudinit_disk.lb.id}"

  cpu {
    mode = "host-passthrough"
  }


  disk {
    volume_id = "${libvirt_volume.lb.id}"
  }

  network_interface {
    network_id     = "${libvirt_network.network.id}"
    hostname       = "${var.name_prefix}lb"
    addresses      = ["${cidrhost("${var.network}", 256)}"]
    wait_for_lease = 1
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }

  connection {
    type     = "ssh"
    user     = "root"
    password = "linux"
  }
}

output "ip_lb" {
  value = "${libvirt_domain.lb.network_interface.0.addresses[0]}"
}

#####################
### Cluster masters #
#####################

resource "libvirt_volume" "master" {
  name           = "${var.name_prefix}master_${count.index}.qcow2"
  pool           = "${var.pool}"
  base_volume_id = "${libvirt_volume.img.id}"
  count          = "${var.master_count}"
}

data "template_file" "master_cloud_init_user_data" {
  # needed when 0 master nodes are defined
  count    = "${var.master_count}"
  template = "${file("cloud-init/master.cfg.tpl")}"
  vars = {
    hostname = "${var.name_prefix}master-${count.index}"
    fqdn = "${var.name_prefix}master-${count.index}.${var.name_prefix}${var.domain_name}"
    repo_baseurl = "${var.repo_baseurl}"
  }
}

resource "libvirt_cloudinit_disk" "master" {
  # needed when 0 master nodes are defined
  count     = "${var.master_count}"
  name      = "${var.name_prefix}master_cloud_init_${count.index}.iso"
  pool      = "${var.pool}"
  user_data = "${element(data.template_file.master_cloud_init_user_data.*.rendered, count.index)}"
}

resource "libvirt_domain" "master" {
  count      = "${var.master_count}"
  name       = "${var.name_prefix}master_${count.index}"
  memory     = "${var.master_memory}"
  vcpu       = "${var.master_vcpu}"
  cloudinit  = "${element(libvirt_cloudinit_disk.master.*.id, count.index)}"
  metadata   = "master-${count.index}.${var.domain_name},master,${count.index},${var.name_prefix}"

  cpu {
    mode = "host-passthrough"
  }

  disk {
    volume_id = "${element(libvirt_volume.master.*.id, count.index)}"
  }

  network_interface {
    network_id     = "${libvirt_network.network.id}"
    hostname       = "${var.name_prefix}master-${count.index}"
    addresses      = ["${cidrhost("${var.network}", 512 + count.index)}"]
    wait_for_lease = 1
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }

  connection {
    type     = "ssh"
    user     = "root"
    password = "linux"
  }

  depends_on = ["libvirt_domain.lb"]
}

output "masters" {
  value = ["${libvirt_domain.master.*.network_interface.0.addresses[0]}"]
}

####################
## Cluster workers #
####################

resource "libvirt_volume" "worker" {
  name           = "${var.name_prefix}worker_${count.index}.qcow2"
  pool           = "${var.pool}"
  base_volume_id = "${libvirt_volume.img.id}"
  count          = "${var.worker_count}"
}

data "template_file" "worker_cloud_init_user_data" {
  # needed when 0 worker nodes are defined
  count    = "${var.worker_count}"
  template = "${file("cloud-init/worker.cfg.tpl")}"
  vars = {
    hostname = "${var.name_prefix}worker-${count.index}"
    fqdn = "${var.name_prefix}worker-${count.index}.${var.name_prefix}${var.domain_name}"
    repo_baseurl = "${var.repo_baseurl}"
  }
}

resource "libvirt_cloudinit_disk" "worker" {
  # needed when 0 worker nodes are defined
  count     = "${var.worker_count}"
  name      = "${var.name_prefix}worker_cloud_init_${count.index}.iso"
  pool      = "${var.pool}"
  user_data = "${element(data.template_file.worker_cloud_init_user_data.*.rendered, count.index)}"
}

resource "libvirt_domain" "worker" {
  count      = "${var.worker_count}"
  name       = "${var.name_prefix}worker_${count.index}"
  memory     = "${var.worker_memory}"
  vcpu       = "${var.worker_vcpu}"
  cloudinit  = "${element(libvirt_cloudinit_disk.worker.*.id, count.index)}"
  metadata   = "worker-${count.index}.${var.domain_name},worker,${count.index},${var.name_prefix}"

  cpu {
    mode = "host-passthrough"
  }


  disk {
    volume_id = "${element(libvirt_volume.worker.*.id, count.index)}"
  }

  network_interface {
    network_id     = "${libvirt_network.network.id}"
    hostname       = "${var.name_prefix}worker-${count.index}"
    addresses      = ["${cidrhost("${var.network}", 768 + count.index)}"]
    wait_for_lease = 1
  }

  graphics {
    type        = "vnc"
    listen_type = "address"
  }

  connection {
    type     = "ssh"
    user     = "root"
    password = "linux"
  }

  depends_on = ["libvirt_domain.master"]
}

output "workers" {
  value = ["${libvirt_domain.worker.*.network_interface.0.addresses[0]}"]
}
