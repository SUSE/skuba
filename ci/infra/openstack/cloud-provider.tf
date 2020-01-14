resource "null_resource" "generate_cloud_provider_conf" {
  depends_on = [
    null_resource.master_reboot,
    null_resource.worker_reboot,
  ]
  count = var.cpi_enable ? 1 : 0

  provisioner "local-exec" {
    environment = {
      CA_FILE              = var.ca_file
      TR_STACK             = var.stack_name
      TR_USERNAME          = var.username
      TR_LB_IP             = openstack_networking_floatingip_v2.lb_ext.address
      TR_MASTER_IPS        = join(" ", openstack_networking_floatingip_v2.master_ext.*.address)
      TR_WORKER_IPS        = join(" ", openstack_networking_floatingip_v2.worker_ext.*.address)
      OS_PRIVATE_SUBNET_ID = openstack_networking_subnet_v2.subnet.id
      OS_PUBLIC_NET_ID     = data.openstack_networking_network_v2.external_network.id
    }

    command = "bash generate-cpi-conf.sh"
  }
}

