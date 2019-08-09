resource "openstack_compute_secgroup_v2" "secgroup_base" {
  name        = "caasp-base-${var.stack_name}"
  description = "Basic security group"

  # Allow ping and cilium health checks as well
  rule {
    from_port   = -1
    to_port     = -1
    ip_protocol = "icmp"
    cidr        = "0.0.0.0/0"
  }

  # sshd
  rule {
    from_port   = 22
    to_port     = 22
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # cilium health check
  rule {
    from_port   = 4240
    to_port     = 4240
    ip_protocol = "tcp"
    cidr        = "${var.subnet_cidr}"
  }

  # cilium VXLAN
  rule {
    from_port   = 8472
    to_port     = 8472
    ip_protocol = "udp"
    cidr        = "${var.subnet_cidr}"
  }

  # kubelet - API server -> kubelet communication
  rule {
    from_port   = 10250
    to_port     = 10250
    ip_protocol = "tcp"
    cidr        = "${var.subnet_cidr}"
  }

  # kubeproxy health check
  rule {
    from_port   = 10256
    to_port     = 10256
    ip_protocol = "tcp"
    cidr        = "${var.subnet_cidr}"
  }

  # Range of ports used by kubernetes when
  # allocating services of type `NodePort`
  rule {
    from_port   = 30000
    to_port     = 32767
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }
  rule {
    from_port   = 30000
    to_port     = 32767
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master" {
  name        = "caasp-master-${var.stack_name}"
  description = "security group for masters"

  # etcd - client communication
  rule {
    from_port   = 2379
    to_port     = 2379
    ip_protocol = "tcp"
    cidr        = "${var.subnet_cidr}"
  }

  # etcd - server-to-server communication
  rule {
    from_port   = 2380
    to_port     = 2380
    ip_protocol = "tcp"
    cidr        = "${var.subnet_cidr}"
  }

  # API server
  rule {
    from_port   = 6443
    to_port     = 6444
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master_lb" {
  name        = "caasp-master-lb-${var.stack_name}"
  description = "security group for master load balancers"

  # API server
  rule {
    from_port   = 6443
    to_port     = 6443
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # dex - OIDC connect
  rule {
    from_port   = 32000
    to_port     = 32000
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # gangway (RBAC authenticate)
  rule {
    from_port   = 32001
    to_port     = 32001
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }
}
