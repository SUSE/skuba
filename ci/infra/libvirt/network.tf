resource "libvirt_network" "network" {
  name   = "${var.stack_name}-network"
  mode   = var.network_mode
  domain = var.dns_domain

  dns {
    enabled = true
  }

  addresses = [var.network_cidr]
}

