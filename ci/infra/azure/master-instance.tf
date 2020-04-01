resource "azurerm_network_interface" "master" {
  count               = var.masters
  name                = "${var.stack_name}-master-${count.index}-nic"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.nodes.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = var.create_bastionhost ? null : element(azurerm_public_ip.master.*.id, count.index)
  }
}

resource "azurerm_network_interface_security_group_association" "master" {
  count                     = var.masters
  network_interface_id      = element(azurerm_network_interface.master.*.id, count.index)
  network_security_group_id = azurerm_network_security_group.master.id
}

resource "azurerm_linux_virtual_machine" "master" {
  count                 = var.masters
  name                  = "${var.stack_name}-master-${count.index}"
  computer_name         = "master-${count.index}"
  resource_group_name   = azurerm_resource_group.rg.name
  location              = azurerm_resource_group.rg.location
  size                  = var.master_size
  network_interface_ids = [element(azurerm_network_interface.master.*.id, count.index)]

  source_image_reference {
    publisher = "SUSE"
    offer     = "sles-15-sp1-chost-byos"
    sku       = "gen2"
    version   = data.azurerm_platform_image.sles_chost_byos.version
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  priority        = var.master_use_spot_instance ? "Spot" : "Regular"
  eviction_policy = var.master_use_spot_instance ? "Deallocate" : null

  admin_username = "sles"
  admin_ssh_key {
    username = "sles"
    public_key = var.admin_ssh_key
  }
  admin_password = var.admin_password
  disable_password_authentication = (var.admin_password == "") ? true : false

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [source_image_id]
  }
}

resource "azurerm_virtual_machine_extension" "master" {
  count                = var.masters
  name                 = "${var.stack_name}-master-${count.index}"
  virtual_machine_id   = element(azurerm_linux_virtual_machine.master.*.id, count.index)
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
    {
        "script": "${base64encode(data.template_file.init.rendered)}"
    }
SETTINGS

  tags = var.tags
}
