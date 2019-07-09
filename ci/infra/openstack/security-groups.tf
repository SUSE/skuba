resource "openstack_compute_secgroup_v2" "secgroup_base" {
  name        = "${var.stack_name}-base"
  description = "Basic security group"

  rule {
    from_port   = -1
    to_port     = -1
    ip_protocol = "icmp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 22
    to_port     = 22
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 2379
    to_port     = 2379
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 4240
    to_port     = 4240
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 8472
    to_port     = 8472
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master" {
  name        = "${var.stack_name}-master"
  description = "security group for masters"

  rule {
    from_port   = 2380
    to_port     = 2380
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 6443
    to_port     = 6444
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 8285
    to_port     = 8285
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_worker" {
  name        = "${var.stack_name}-worker"
  description = "security group for workers"

  rule {
    from_port   = 80
    to_port     = 80
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 443
    to_port     = 443
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 8080
    to_port     = 8080
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 8081
    to_port     = 8081
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 2380
    to_port     = 2380
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 10250
    to_port     = 10250
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 8285
    to_port     = 8285
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master_lb" {
  name        = "${var.stack_name}-master-lb"
  description = "security group for master load balancers"

  rule {
    from_port   = 6443
    to_port     = 6443
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 32001
    to_port     = 32001
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  rule {
    from_port   = 32000
    to_port     = 32000
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }
}
