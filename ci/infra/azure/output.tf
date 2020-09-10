output "username" {
  value = var.username
}

output "bastion_host" {
  value = var.create_bastionhost ? azurerm_bastion_host.bastionhost.0.dns_name : "creation disabled"
}

output "ip_load_balancer" {
  value = {
    "public_ip" : azurerm_public_ip.lb.ip_address,
    "fqdn" : azurerm_public_ip.lb.fqdn,
  }
}

output "masters_public_ip" {
  value = zipmap(
    azurerm_linux_virtual_machine.master.*.name,
    azurerm_linux_virtual_machine.master.*.public_ip_address,
  )
}

output "masters_private_ip" {
  value = zipmap(
    azurerm_linux_virtual_machine.master.*.name,
    azurerm_linux_virtual_machine.master.*.private_ip_address,
  )
}

output "workers_private_ip" {
  value = zipmap(
    azurerm_linux_virtual_machine.worker.*.name,
    azurerm_linux_virtual_machine.worker.*.private_ip_address
  )
}

output "route_table" {
  value = var.cpi_enable ? azurerm_route_table.nodes[0].name : null
}
