resource "azurerm_network_security_group" "master" {
  name                = "${var.stack_name}-master"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  tags                = var.tags
}

resource "azurerm_network_security_rule" "master" {
  count                       = length(local.security_rules_master)
  resource_group_name         = azurerm_resource_group.rg.name
  network_security_group_name = azurerm_network_security_group.master.name
  name                        = element(local.security_rules_master, count.index)["name"]
  priority                    = (100 + count.index)
  direction                   = element(local.security_rules_master, count.index)["direction"]
  access                      = element(local.security_rules_master, count.index)["access"]
  protocol                    = element(local.security_rules_master, count.index)["protocol"]
  source_port_range           = element(local.security_rules_master, count.index)["source_port_range"]
  destination_port_range      = element(local.security_rules_master, count.index)["destination_port_range"]
  source_address_prefix       = element(local.security_rules_master, count.index)["source_address_prefix"]
  destination_address_prefix  = element(local.security_rules_master, count.index)["destination_address_prefix"]
}
