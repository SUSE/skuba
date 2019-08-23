resource "openstack_networking_secgroup_v2" "load_balancer" {
  name        = "${var.stack_name}-caasp_common_secgroup"
  description = "Common security group for CaaSP load balancer"
}

resource "openstack_networking_secgroup_rule_v2" "lb_api_server" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.load_balancer.id}"
}

# Range of ports used by kubernetes when allocating services of type `NodePort`
resource "openstack_networking_secgroup_rule_v2" "lb_kubernetes_services_tcp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 30000
  port_range_max    = 32767
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.load_balancer.id}"
}

resource "openstack_networking_secgroup_rule_v2" "lb_kubernetes_services_udp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = 30000
  port_range_max    = 32767
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.load_balancer.id}"
}
