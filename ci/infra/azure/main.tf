provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "resource_group" {
  name     = "${var.stack_name}-resource-group"
  location = var.azure_location
}

data "azurerm_platform_image" "sles_chost_byos" {
  location  = azurerm_resource_group.resource_group.location
  publisher = "SUSE"
  offer     = "sles-15-sp2-chost-byos"
  sku       = "gen2"
}

data "azurerm_subscription" "current" {
}

data "azurerm_role_definition" "contributor" {
  name = "Contributor"
}
