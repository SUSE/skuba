resource "openstack_compute_secgroup_v2" "secgroup_base_external" {
  name        = "caasp-base-external-${var.stack_name}"
  description = "Basic external security group"

  # ping
  rule {
    from_port   = -1
    to_port     = -1
    ip_protocol = "icmp"
    cidr        = "0.0.0.0/0"
  }

  # ssh
  rule {
    from_port   = 22
    to_port     = 22
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_base_internal" {
  name        = "caasp-base-internal-${var.stack_name}"
  description = "Basic internal security group"

  # etcd client requests
  rule {
    from_port   = 2379
    to_port     = 2379
    ip_protocol = "tcp"
    self        = true
  }

  # Cilium health checks
  rule {
    from_port   = 4240
    to_port     = 4240
    ip_protocol = "tcp"
    self        = true
  }

  # Cilium VXLAN overlay
  rule {
    from_port   = 8472
    to_port     = 8472
    ip_protocol = "udp"
    self        = true
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master_external" {
  name        = "caasp-master-external-${var.stack_name}"
  description = "External security group for masters"

  # Node ports (TCP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # Node ports (UDP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master_internal" {
  name        = "caasp-master-internal-${var.stack_name}"
  description = "Internal security group for masters"

  # etcd peer
  rule {
    from_port   = 2380
    to_port     = 2380
    ip_protocol = "tcp"
    self        = true
  }

  # Kubernetes API server
  rule {
    from_port   = 6443
    to_port     = 6443
    ip_protocol = "tcp"
    self        = true
  }

  # Kubelet API
  rule {
    from_port   = 10250
    to_port     = 10250
    ip_protocol = "tcp"
    self        = true
  }

  # Node ports (TCP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "tcp"
    self        = true
  }

  # Node ports (UDP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "udp"
    self        = true
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_worker_external" {
  name        = "caasp-worker-external-${var.stack_name}"
  description = "External security group for workers"

  # Node ports (TCP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }

  # Node ports (UDP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "udp"
    cidr        = "0.0.0.0/0"
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_worker_internal" {
  name        = "caasp-worker-internal-${var.stack_name}"
  description = "Internal security group for workers"

  # etcd peer
  rule {
    from_port   = 2380
    to_port     = 2380
    ip_protocol = "tcp"
    self        = true
  }

  # Kubelet API
  rule {
    from_port   = 10250
    to_port     = 10250
    ip_protocol = "tcp"
    self        = true
  }

  # Node ports (TCP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "tcp"
    self        = true
  }

  # Node ports (UDP)
  rule {
    from_port   = 30000
    to_port     = 32768
    ip_protocol = "udp"
    self        = true
  }
}

resource "openstack_compute_secgroup_v2" "secgroup_master_lb" {
  name        = "caasp-master-lb-${var.stack_name}"
  description = "security group for master load balancers"

  # Kubernetes API server
  rule {
    from_port   = 6443
    to_port     = 6443
    ip_protocol = "tcp"
    cidr        = "0.0.0.0/0"
  }
}
