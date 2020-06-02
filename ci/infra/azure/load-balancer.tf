resource "azurerm_lb" "lb" {
  name                = "${var.stack_name}-lb"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  sku                 = "standard"

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.lb.id
  }
}

resource "azurerm_lb_backend_address_pool" "masters" {
  resource_group_name = azurerm_resource_group.resource_group.name
  loadbalancer_id     = azurerm_lb.lb.id
  name                = "master-nodes"
}

resource "azurerm_lb_probe" "kube_api" {
  resource_group_name = azurerm_resource_group.resource_group.name
  loadbalancer_id     = azurerm_lb.lb.id
  name                = "kube-apisever-running-probe"
  port                = 6443
}

resource "azurerm_lb_rule" "kube_api" {
  resource_group_name            = azurerm_resource_group.resource_group.name
  loadbalancer_id                = azurerm_lb.lb.id
  name                           = "kube-api-server"
  protocol                       = "Tcp"
  frontend_port                  = 6443
  backend_port                   = 6443
  frontend_ip_configuration_name = azurerm_lb.lb.frontend_ip_configuration[0].name
  probe_id                       = azurerm_lb_probe.kube_api.id
  backend_address_pool_id        = azurerm_lb_backend_address_pool.masters.id
}

resource "azurerm_lb_probe" "kube_dex" {
  resource_group_name = azurerm_resource_group.resource_group.name
  loadbalancer_id     = azurerm_lb.lb.id
  name                = "kube-dex-running-probe"
  port                = 32000
}

resource "azurerm_lb_rule" "kube_dex" {
  resource_group_name            = azurerm_resource_group.resource_group.name
  loadbalancer_id                = azurerm_lb.lb.id
  name                           = "kube-dex"
  protocol                       = "Tcp"
  frontend_port                  = 32000
  backend_port                   = 32000
  frontend_ip_configuration_name = azurerm_lb.lb.frontend_ip_configuration[0].name
  probe_id                       = azurerm_lb_probe.kube_dex.id
  backend_address_pool_id        = azurerm_lb_backend_address_pool.masters.id
}

resource "azurerm_lb_probe" "kube_gangway" {
  resource_group_name = azurerm_resource_group.resource_group.name
  loadbalancer_id     = azurerm_lb.lb.id
  name                = "kube-gangway-running-probe"
  port                = 32001
}

resource "azurerm_lb_rule" "kube_gangway" {
  resource_group_name            = azurerm_resource_group.resource_group.name
  loadbalancer_id                = azurerm_lb.lb.id
  name                           = "kube-gangway"
  protocol                       = "Tcp"
  frontend_port                  = 32001
  backend_port                   = 32001
  frontend_ip_configuration_name = azurerm_lb.lb.frontend_ip_configuration[0].name
  probe_id                       = azurerm_lb_probe.kube_gangway.id
  backend_address_pool_id        = azurerm_lb_backend_address_pool.masters.id
}

resource "azurerm_network_interface_backend_address_pool_association" "kube_api" {
  count                   = var.masters
  backend_address_pool_id = azurerm_lb_backend_address_pool.masters.id
  ip_configuration_name   = "internal"
  network_interface_id    = element(azurerm_network_interface.master.*.id, count.index)
}
