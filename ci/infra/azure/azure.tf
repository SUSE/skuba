provider "azurerm" {
  #version = "<= 1.33"

  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = "${var.stack_name}-resource-group"
  location = var.location
  tags     = var.tags
}
