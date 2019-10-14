variable "lbs" {
  default     = 1
  description = "Number of load-balancer nodes"
}

variable "lb_cpus" {
  default     = 1
  description = "Number of CPUs used on load-balancer node"
}

variable "lb_memory" {
  default     = 2048
  description = "Amount of memory used on load-balancer node"
}

variable "lb_disk_size" {
  default     = 40
  description = "Size of the root disk in GB on load-balancer node"
}

variable "lb_repositories" {
  type = "map"

  default = {
    sle_server_pool    = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Products/SLE-Product-SLES/15-SP1/x86_64/product/"
    basesystem_pool    = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Products/SLE-Module-Basesystem/15-SP1/x86_64/product/"
    ha_pool            = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Products/SLE-Product-HA/15-SP1/x86_64/product/"
    ha_updates         = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Updates/SLE-Product-HA/15-SP1/x86_64/update/"
    sle_server_updates = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Updates/SLE-Product-SLES/15-SP1/x86_64/update/"
    basesystem_updates = "http://ibs-mirror.prv.suse.net/ibs/SUSE/Updates/SLE-Module-Basesystem/15-SP1/x86_64/update/"
  }
}

locals {
  lb_repositories_template = [for i in range(length(var.lb_repositories)) : templatefile("cloud-init/repository.tpl",
    {
      repository_url  = "${element(values(var.lb_repositories), i)}",
      repository_name = "${element(keys(var.lb_repositories), i)}"
  })]

  # TODO: 
  # depends_on = [
  #   "vsphere_virtual_machine.master",
  # ]
  haproxy_apiserver_backends_master = <<EOT
    %{for i in range(var.masters)~}
    %{for fqdn in vsphere_virtual_machine.master.*.name~}
    %{for ip in vsphere_virtual_machine.master.*.default_ip_address~}
    server ${fqdn} ${ip}:6443
    %{endfor~}
    %{endfor~}
    %{endfor~}
  EOT

  haproxy_gangway_backends_master = <<EOT
    %{for i in range(var.masters)~}
    %{for fqdn in vsphere_virtual_machine.master.*.name~}
    %{for ip in vsphere_virtual_machine.master.*.default_ip_address~}
    server ${fqdn} ${ip}:32001
    %{endfor~}
    %{endfor~}
    %{endfor~}
  EOT

  haproxy_dex_backends_master = <<EOT
    %{for i in range(var.masters)~}
    %{for fqdn in vsphere_virtual_machine.master.*.name~}
    %{for ip in vsphere_virtual_machine.master.*.default_ip_address~}
    server ${fqdn} ${ip}:32000
    %{endfor~}
    %{endfor~}
    %{endfor~}
  EOT

  lb_cloud_init_metadata = templatefile("cloud-init/metadata.tpl", {
    network_config = "${base64gzip(data.local_file.network_cloud_init.content)}"
    instance_id    = "${var.stack_name}-lb"
  })

  lb_haproxy_cfg = templatefile("cloud-init/haproxy.cfg.tpl", {
    apiserver_backends = "${join("  ", local.haproxy_apiserver_backends_master.*)}"
    gangway_backends   = "${join("  ", local.haproxy_gangway_backends_master.*)}"
    dex_backends       = "${join("  ", local.haproxy_dex_backends_master.*)}"
  })

  lb_cloud_init_userdata = templatefile("cloud-init/lb.tpl", {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repositories    = "${join("\n", local.lb_repositories_template.*)}"
    packages        = "${join("\n", formatlist("  - %s", var.packages))}"
    ntp_servers     = "${join("\n", formatlist("    - %s", var.ntp_servers))}"
  })
}

resource "vsphere_virtual_machine" "lb" {
  count            = "${var.lbs}"
  name             = "${var.stack_name}-lb-${count.index}"
  num_cpus         = "${var.lb_cpus}"
  memory           = "${var.lb_memory}"
  guest_id         = "${var.guest_id}"
  firmware         = "${var.firmware}"
  scsi_type        = "${data.vsphere_virtual_machine.template.scsi_type}"
  resource_pool_id = "${data.vsphere_resource_pool.pool.id}"
  datastore_id     = "${data.vsphere_datastore.datastore.id}"

  clone {
    template_uuid = "${data.vsphere_virtual_machine.template.id}"
  }

  disk {
    label = "disk0"
    size  = "${var.lb_disk_size}"
  }

  extra_config = {
    "guestinfo.metadata"          = "${base64gzip(local.lb_cloud_init_metadata)}"
    "guestinfo.metadata.encoding" = "gzip+base64"

    "guestinfo.userdata"          = "${base64gzip(local.lb_cloud_init_userdata)}"
    "guestinfo.userdata.encoding" = "gzip+base64"
  }

  network_interface {
    network_id = "${data.vsphere_network.network.id}"
  }

  depends_on = [
    "vsphere_virtual_machine.master",
  ]
}

resource "null_resource" "lb_wait_cloudinit" {
  depends_on = ["vsphere_virtual_machine.lb"]
  count      = "${var.lbs}"

  connection {
    host  = "${element(vsphere_virtual_machine.lb.*.guest_ip_addresses.0, count.index)}"
    user  = "${var.username}"
    type  = "ssh"
    agent = true
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
    host  = "${element(vsphere_virtual_machine.lb.*.guest_ip_addresses.0, count.index)}"
    user  = "${var.username}"
    type  = "ssh"
    agent = true
  }

  provisioner "file" {
    content     = "${local.lb_haproxy_cfg}"
    destination = "/tmp/haproxy.cfg"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mv /tmp/haproxy.cfg /etc/haproxy/haproxy.cfg",
      "sudo systemctl enable haproxy && sudo systemctl restart haproxy",
    ]
  }
}
