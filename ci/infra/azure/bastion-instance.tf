resource "azurerm_bastion_host" "bastionhost" {
  count               = var.create_bastionhost ? 1 : 0
  name                = "${var.stack_name}-bastion"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name

  ip_configuration {
    name                 = "${var.stack_name}-configuration"
    subnet_id            = azurerm_subnet.bastionhost.0.id
    public_ip_address_id = azurerm_public_ip.bastionhost.0.id
  }
}
