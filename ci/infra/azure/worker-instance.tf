resource "azurerm_network_interface" "worker" {
  count               = var.workers
  name                = "${var.stack_name}-worker-${count.index}-nic"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.nodes.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_interface_security_group_association" "worker" {
  count                     = var.workers
  network_interface_id      = element(azurerm_network_interface.worker.*.id, count.index)
  network_security_group_id = azurerm_network_security_group.worker.id
}

resource "azurerm_linux_virtual_machine" "worker" {
  count                 = var.workers
  name                  = "${var.stack_name}-worker-${count.index}"
  computer_name         = "worker-${count.index}"
  resource_group_name   = azurerm_resource_group.rg.name
  location              = azurerm_resource_group.rg.location
  size                  = var.worker_size
  network_interface_ids = [element(azurerm_network_interface.worker.*.id, count.index)]

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

  priority        = var.worker_use_spot_instance ? "Spot" : "Regular"
  eviction_policy = var.worker_use_spot_instance ? "Deallocate" : null

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

resource "azurerm_virtual_machine_extension" "worker" {
  count                = var.workers
  name                 = "${var.stack_name}-worker-${count.index}"
  virtual_machine_id   = element(azurerm_linux_virtual_machine.worker.*.id, count.index)
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
