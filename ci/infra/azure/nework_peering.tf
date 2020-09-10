resource "azurerm_virtual_network_peering" "peering" {
  count                        = length(var.peer_virutal_network_id)
  name                         = "peering-to-${var.peer_virutal_network_id[count.index]}"
  resource_group_name          = azurerm_resource_group.resource_group.name
  virtual_network_name         = azurerm_virtual_network.virtual_network.name
  remote_virtual_network_id    = var.peer_virutal_network_id[count.index]
  allow_virtual_network_access = true
  allow_forwarded_traffic      = true

  # `allow_gateway_transit` must be set to false for vnet Global Peering
  allow_gateway_transit = false
}
