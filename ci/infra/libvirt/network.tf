resource "libvirt_network" "network" {
  count  = var.network_name == "" ? 1 : 0
  name   = "${var.stack_name}-network"
  mode   = var.network_mode
  domain = var.dns_domain

  dns {
    enabled = true
  }

  addresses = [var.network_cidr]
}

