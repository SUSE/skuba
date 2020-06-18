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
  offer     = "sles-15-sp1-chost-byos"
  sku       = "gen2"
}

resource "azurerm_image" "sles_chost_byos" {
  # name                      = "SLES15-SP2-CHOST-BYOS.x86_64-0.9.12-Azure-Build1.4.vhd"
  name                      = "sles-15-sp2-chost-byos"
  location                  = "West Europe"
  resource_group_name       = "openqa-upload"

    os_disk {
    os_type  = "Linux"
    os_state = "Generalized"
    # blob_uri = "{blob_uri}"
    # blob_uri = https://<StorageAcctName>.blob.core.windows.net/vhds/<osdiskName>.vhd
    blob_uri = "https://openqa.blob.core.windows.net/vhds/SLES15-SP2-CHOST-BYOS.x86_64-0.9.12-Azure-Build1.4.vhd"
    # blob_uri = "/subscriptions/c011786b-59d7-4817-880c-7cd8a6ca4b19/resourceGroups/openqa-upload/providers/Microsoft.Compute/disks/SLES15-SP2-CHOST-BYOS.x86_64-0.9.12-Azure-Build1.4.vhd"
    blob_uri       = "https://${var.storage_account_name}.blob.core.windows.net/vhds/${var.hostname}-osdisk.vhd"
    size_gb  = 30
  }
}
