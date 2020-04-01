data "azurerm_platform_image" "sles_chost_byos" {
  location  = azurerm_resource_group.rg.location
  publisher = "SUSE"
  offer     = "sles-15-sp1-chost-byos"
  sku       = "gen2"
}
