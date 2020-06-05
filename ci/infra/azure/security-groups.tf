locals {
  security_rules_commons = [
    {
      "name" : "ssh",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Tcp",
      "source_port_range" : "*",
      "destination_port_range" : "22",
      "source_address_prefix" : "*",
      "destination_address_prefix" : "*"
    },
    {
      "name" : "cilium-health-check",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Tcp",
      "source_port_range" : "*",
      "destination_port_range" : "4240",
      "source_address_prefix" : var.private_subnet_cidr,
      "destination_address_prefix" : "*"
    },
    {
      "name" : "cilium-vxlan",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Udp",
      "source_port_range" : "*",
      "destination_port_range" : "8472",
      "source_address_prefix" : var.private_subnet_cidr,
      "destination_address_prefix" : "*"
    },
    {
      "name" : "api-server_to_kubelet",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Tcp",
      "source_port_range" : "*",
      "destination_port_range" : "10250",
      "source_address_prefix" : var.private_subnet_cidr,
      "destination_address_prefix" : "*"
    },
    {
      "name" : "kubeproxy-health-check",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Tcp",
      "source_port_range" : "*",
      "destination_port_range" : "10256",
      "source_address_prefix" : var.private_subnet_cidr,
      "destination_address_prefix" : "*"
    },
    {
      "name" : "services-NodePort-tcp",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Tcp",
      "source_port_range" : "*",
      "destination_port_range" : "30000-32767",
      "source_address_prefix" : "*",
      "destination_address_prefix" : "*"
    },
    {
      "name" : "services-NodePort-udp",
      "direction" : "Inbound",
      "access" : "Allow",
      "protocol" : "Udp",
      "source_port_range" : "*",
      "destination_port_range" : "30000-32767",
      "source_address_prefix" : "*",
      "destination_address_prefix" : "*"
    }
  ]

  security_rules_master = concat(
    local.security_rules_commons,
    [
      {
        "name" : "etcd-internal",
        "direction" : "Inbound",
        "access" : "Allow",
        "protocol" : "Tcp",
        "source_port_range" : "*",
        "destination_port_range" : "2379-2380",
        "source_address_prefix" : var.private_subnet_cidr,
        "destination_address_prefix" : "*"
      },
      {
        "name" : "api-server",
        "direction" : "Inbound",
        "access" : "Allow",
        "protocol" : "Tcp",
        "source_port_range" : "*",
        "destination_port_range" : "6443",
        "source_address_prefix" : "*",
        "destination_address_prefix" : "*"
      }
  ])

  # Right now there's no need for special rules for worker nodes
  security_rules_worker = local.security_rules_commons
}

resource "azurerm_network_security_group" "master" {
  name                = "${var.stack_name}-master"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azurerm_network_security_rule" "master" {
  count                       = length(local.security_rules_master)
  resource_group_name         = azurerm_resource_group.resource_group.name
  network_security_group_name = azurerm_network_security_group.master.name
  name                        = element(local.security_rules_master, count.index)["name"]
  priority                    = (100 + count.index)
  direction                   = element(local.security_rules_master, count.index)["direction"]
  access                      = element(local.security_rules_master, count.index)["access"]
  protocol                    = element(local.security_rules_master, count.index)["protocol"]
  source_port_range           = element(local.security_rules_master, count.index)["source_port_range"]
  destination_port_range      = element(local.security_rules_master, count.index)["destination_port_range"]
  source_address_prefix       = element(local.security_rules_master, count.index)["source_address_prefix"]
  destination_address_prefix  = element(local.security_rules_master, count.index)["destination_address_prefix"]
}

resource "azurerm_network_security_group" "worker" {
  name                = "${var.stack_name}-worker"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
}

resource "azurerm_network_security_rule" "worker" {
  count                       = length(local.security_rules_worker)
  resource_group_name         = azurerm_resource_group.resource_group.name
  network_security_group_name = azurerm_network_security_group.worker.name
  name                        = element(local.security_rules_worker, count.index)["name"]
  priority                    = (100 + count.index)
  direction                   = element(local.security_rules_worker, count.index)["direction"]
  access                      = element(local.security_rules_worker, count.index)["access"]
  protocol                    = element(local.security_rules_worker, count.index)["protocol"]
  source_port_range           = element(local.security_rules_worker, count.index)["source_port_range"]
  destination_port_range      = element(local.security_rules_worker, count.index)["destination_port_range"]
  source_address_prefix       = element(local.security_rules_worker, count.index)["source_address_prefix"]
  destination_address_prefix  = element(local.security_rules_worker, count.index)["destination_address_prefix"]
}
