resource "openstack_networking_secgroup_v2" "load_balancer" {
  name        = "${var.stack_name}-caasp_lb_secgroup"
  description = "Common security group for CaaSP load balancer"
}

resource "openstack_networking_secgroup_rule_v2" "lb_api_server" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.load_balancer.id
}

# Needed to allow access from the LB to dex (32000) and gangway (32001)
resource "openstack_networking_secgroup_rule_v2" "lb_k8s_auth_svcs" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 32000
  port_range_max    = 32001
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.load_balancer.id
}

