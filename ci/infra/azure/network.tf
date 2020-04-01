resource "azurerm_virtual_network" "vnet" {
  name                = "${var.stack_name}-vnet"
  address_space       = [var.vnet_address_space]
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  tags                = var.tags
}

# Subnets

resource "azurerm_subnet" "bastionhost" {
  count                = var.create_bastionhost ? 1: 0
  name                 = "AzureBastionSubnet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefix       = var.bastionhost_subnet_cidr
}

resource "azurerm_subnet" "nodes" {
  name                 = "${var.stack_name}-nodes"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefix       = var.private_subnet_cidr
}

# Public IPs

resource "azurerm_public_ip" "bastionhost" {
  count               = var.create_bastionhost ? 1: 0
  name                = "${var.stack_name}-bastionhost-ip"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_public_ip" "master" {
  count               = var.create_bastionhost ? 0 : var.masters
  name                = "${var.stack_name}-master-${count.index}-ip"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_public_ip" "lb" {
  name                = "control-plane"
  domain_name_label   = var.stack_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "standard"
}
