resource "azurerm_network_interface" "worker" {
  count               = var.workers
  name                = "${var.stack_name}-worker-${count.index}-nic"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name

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
  name                  = "${var.stack_name}-worker-${count.index}-vm"
  resource_group_name   = azurerm_resource_group.resource_group.name
  location              = azurerm_resource_group.resource_group.location
  zone                  = var.enable_zone ? var.azure_availability_zones[count.index % length(var.azure_availability_zones)] : null
  size                  = var.worker_vm_size
  network_interface_ids = [element(azurerm_network_interface.worker.*.id, count.index)]

  admin_username = var.username
  admin_ssh_key {
    username   = var.username
    public_key = var.authorized_keys.0
  }
  admin_password                  = var.password
  disable_password_authentication = (var.password == "") ? true : false

  os_disk {
    caching              = "ReadOnly"
    storage_account_type = var.worker_storage_account_type
    disk_size_gb         = var.worker_disk_size
  }

  source_image_reference {
    publisher = "SUSE"
    offer     = "sles-15-sp1-chost-byos"
    sku       = "gen2"
    version   = data.azurerm_platform_image.sles_chost_byos.version
  }
  priority        = var.worker_use_spot_instance ? "Spot" : "Regular"
  eviction_policy = var.worker_use_spot_instance ? "Deallocate" : null

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [source_image_id]
  }
}

resource "azurerm_virtual_machine_extension" "worker" {
  count                = var.workers
  name                 = "${var.stack_name}-worker-${count.index}-vm-extension"
  virtual_machine_id   = element(azurerm_linux_virtual_machine.worker.*.id, count.index)
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
    {
        "script": "${base64encode(data.template_file.cloud-init.rendered)}"
    }
SETTINGS
}
