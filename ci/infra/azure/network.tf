resource "azurerm_virtual_network" "virtual_network" {
  name                = "${var.stack_name}-virtual_network"
  address_space       = [var.cidr_block]
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

# Subnets

resource "azurerm_subnet" "bastionhost" {
  count                = var.create_bastionhost ? 1 : 0
  name                 = "AzureBastionSubnet"
  resource_group_name  = azurerm_resource_group.resource_group.name
  virtual_network_name = azurerm_virtual_network.virtual_network.name
  address_prefixes     = [var.bastionhost_subnet_cidr]
}

resource "azurerm_subnet" "nodes" {
  name                 = "${var.stack_name}-nodes"
  resource_group_name  = azurerm_resource_group.resource_group.name
  virtual_network_name = azurerm_virtual_network.virtual_network.name
  address_prefixes     = [var.private_subnet_cidr]
}

# Public IPs

resource "azurerm_public_ip" "bastionhost" {
  count               = var.create_bastionhost ? 1 : 0
  name                = "${var.stack_name}-bastionhost-ip"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_public_ip" "lb" {
  name                = "${var.stack_name}-lb-ip"
  domain_name_label   = var.stack_name
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  allocation_method   = "Static"
  sku                 = "standard"
}

resource "azurerm_public_ip" "master" {
  count               = var.create_bastionhost ? 0 : var.masters
  name                = "${var.stack_name}-master-${count.index}-ip"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  allocation_method   = "Static"
  sku                 = "Standard"
}


# DNS Zone

resource "azurerm_private_dns_zone" "dns_zone" {
  name                = var.dnsdomain
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "internal_dns" {
  name                  = "caasp"
  resource_group_name   = azurerm_resource_group.resource_group.name
  private_dns_zone_name = azurerm_private_dns_zone.dns_zone.name
  virtual_network_id    = azurerm_virtual_network.virtual_network.id
  registration_enabled  = true
}
