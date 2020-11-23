resource "azurerm_network_interface" "worker" {
  count               = var.workers
  name                = "${var.stack_name}-worker-${count.index}-nic"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  depends_on          = [azurerm_subnet.nodes,]

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
  depends_on                = [azurerm_network_interface.worker, azurerm_network_security_group.worker, azurerm_linux_virtual_machine.worker,]
}

resource "azurerm_linux_virtual_machine" "worker" {
  count                 = var.workers
  name                  = "${var.stack_name}-worker-${count.index}-vm"
  resource_group_name   = azurerm_resource_group.resource_group.name
  location              = azurerm_resource_group.resource_group.location
  zone                  = var.enable_zone ? var.azure_availability_zones[count.index % length(var.azure_availability_zones)] : null
  size                  = var.worker_vm_size
  network_interface_ids = [element(azurerm_network_interface.worker.*.id, count.index),]
  depends_on            = [azurerm_network_interface.worker,]

  admin_username = var.username
  admin_ssh_key {
    username   = var.username
    public_key = var.authorized_keys.0
  }
  admin_password                  = var.password
  disable_password_authentication = (var.password == "") ? true : false
  custom_data                     = data.template_cloudinit_config.cfg.rendered
  os_disk {
    name                 = "${var.stack_name}-worker-${count.index}-disk"
    caching              = "ReadOnly"
    storage_account_type = var.worker_storage_account_type
    disk_size_gb         = var.worker_disk_size
  }

  source_image_reference {
    publisher = "SUSE"
    offer     = "sles-15-sp2-chost-byos"
    sku       = "gen2"
    version   = data.azurerm_platform_image.sles_chost_byos.version
  }
  priority        = var.worker_use_spot_instance ? "Spot" : "Regular"
  eviction_policy = var.worker_use_spot_instance ? "Deallocate" : null

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [source_image_id]
  }

  dynamic "identity" {
    for_each = range(var.cpi_enable ? 1 : 0)
    content {
      type = "SystemAssigned"
    }
  }
}

locals {
  worker_principal_ids = var.cpi_enable ? azurerm_linux_virtual_machine.worker.*.identity.0.principal_id : []
}

resource "azurerm_role_assignment" "worker" {
  count              = var.cpi_enable ? var.workers : 0
  scope              = data.azurerm_subscription.current.id
  role_definition_id = "${data.azurerm_subscription.current.id}${data.azurerm_role_definition.contributor.id}"
  principal_id       = local.worker_principal_ids[count.index]

  depends_on = [azurerm_linux_virtual_machine.worker,]
}
