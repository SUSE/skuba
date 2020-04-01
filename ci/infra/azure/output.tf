output "bastion_host" {
  value = var.create_bastionhost ? azurerm_bastion_host.bastionhost.0.dns_name : "creation disabled"
}

output "loadbalancer" {
  value = {
    "public_ip": azurerm_public_ip.lb.ip_address,
    "fqdn":      azurerm_public_ip.lb.fqdn,
  }
}

output "masters" {
  value = { for vm in azurerm_linux_virtual_machine.master:
    "${vm.computer_name}.${var.dnsdomain}" => {
      "public_ip":  vm.public_ip_address,
      "private_ip": vm.private_ip_address,
    }
  }
}

output "workers" {
  value = { for vm in azurerm_linux_virtual_machine.worker:
    "${vm.computer_name}.${var.dnsdomain}" => {
      "private_ip": vm.private_ip_address,
    }
  }
}
