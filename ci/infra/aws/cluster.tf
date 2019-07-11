locals {
  tags = "${merge(
    map("Name", var.stack_name,
        "Environment", var.stack_name,
        format("kubernetes.io/cluster/%v", var.stack_name), "owned"),
    var.tags)}"
}

provider "aws" {
  region     = "${var.aws_region}"
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  profile    = "default"
}

###########################################
# SCC
###########################################

data "template_file" "register_scc" {
  # register with SCC iff an RMT has not been provided
  count = "${var.caasp_registry_code != "" && var.rmt_server_name == "" ? 1 : 0}"

  template = <<EOF
  - SUSEConnect -d
  - SUSEConnect --cleanup
  - SUSEConnect --url https://scc.suse.com -r ${var.caasp_registry_code}
  - SUSEConnect -p sle-module-containers/15.1/x86_64
  - SUSEConnect -p caasp/4.0/x86_64 -r ${var.caasp_registry_code}
EOF
}

data "template_file" "register_rmt" {
  count = "${var.rmt_server_name != "" ? 1 : 0}"

  template = <<EOF
  - curl --tlsv1.2 --silent --insecure --connect-timeout 10 https://${var.rmt_server_name}/rmt.crt --output /etc/pki/trust/anchors/rmt-server.pem && /usr/sbin/update-ca-certificates &> /dev/null
  - SUSEConnect --url https://${var.rmt_server_name}
  - SUSEConnect -p sle-module-containers/15.1/x86_64
  - SUSEConnect -p caasp/4.0/x86_64
EOF
}

###########################################
# images
###########################################

data "template_file" "repositories" {
  count    = "${length(var.repositories) == 0 ? 0 : length(var.repositories)}"
  template = "${file("cloud-init/repository.tpl")}"

  vars {
    repository_url  = "${element(values(var.repositories[count.index]), 0)}"
    repository_name = "${element(keys(var.repositories[count.index]), 0)}"
  }
}

data "template_file" "cloud-init" {
  template = "${file("cloud-init/cloud-init.yaml.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    packages        = "${join(" ", var.packages)}"
    repositories    = "${length(var.repositories) == 0 ? "\n" : join("\n", data.template_file.repositories.*.rendered)}"
    register_scc    = "${var.caasp_registry_code != "" && var.rmt_server_name == "" ? join("\n", data.template_file.register_scc.*.rendered) : "" }"
    register_rmt    = "${var.rmt_server_name != "" ? join("\n", data.template_file.register_rmt.*.rendered) : ""}"
  }
}

data "template_cloudinit_config" "cfg" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/cloud-config"
    content      = "${data.template_file.cloud-init.rendered}"
  }
}

data "aws_ami" "latest_ami" {
  filter {
    name   = "name"
    values = ["${var.ami_name_pattern}"]
  }

  most_recent = true
  owners      = ["${var.ami_owner}"]

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

###########################################
# networking
###########################################
resource "aws_vpc" "platform" {
  cidr_block           = "${var.vpc_cidr_block}"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags                 = "${merge(local.tags, map("Class", "VPC"))}"
}

// list of az which can be access from the current region
data "aws_availability_zones" "az" {
  state = "available"
}

resource "aws_vpc_dhcp_options" "platform" {
  domain_name         = "${var.aws_region}.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags                = "${merge(local.tags, map("Class", "VPCDHCP"))}"
}

resource "aws_vpc_dhcp_options_association" "dns_resolver" {
  dhcp_options_id = "${aws_vpc_dhcp_options.platform.id}"
  vpc_id          = "${aws_vpc.platform.id}"
}

resource "aws_internet_gateway" "platform" {
  tags       = "${merge(local.tags, map("Class", "Gateway"))}"
  vpc_id     = "${aws_vpc.platform.id}"
  depends_on = ["aws_vpc.platform"]
}

resource "aws_subnet" "public" {
  availability_zone       = "${element(data.aws_availability_zones.az.names, 0)}"
  cidr_block              = "${var.public_subnet}"
  depends_on              = ["aws_main_route_table_association.main"]
  map_public_ip_on_launch = true

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-subnet-public-${element(data.aws_availability_zones.az.names, 0)}",
    "Class", "VPC"))}"

  vpc_id = "${aws_vpc.platform.id}"
}

resource "aws_subnet" "private" {
  availability_zone       = "${element(data.aws_availability_zones.az.names, 0)}"
  cidr_block              = "${var.private_subnet}"
  map_public_ip_on_launch = true

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-subnet-private-${element(data.aws_availability_zones.az.names, 0)}",
    "Class", "Subnet"))}"

  vpc_id = "${aws_vpc.platform.id}"
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-route-table-public",
    "Class", "RouteTable"))}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.platform.id}"
  }
}

resource "aws_route_table" "private" {
  vpc_id = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-route-table-private",
    "Class", "RouteTable"))}"

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = "${aws_nat_gateway.nat_gw.id}"
  }
}

resource "aws_main_route_table_association" "main" {
  route_table_id = "${aws_route_table.public.id}"
  vpc_id         = "${aws_vpc.platform.id}"
}

resource "aws_route_table_association" "private" {
  route_table_id = "${aws_route_table.private.id}"
  subnet_id      = "${aws_subnet.private.id}"
}

resource "aws_route_table_association" "public" {
  route_table_id = "${aws_route_table.public.id}"
  subnet_id      = "${aws_subnet.public.id}"
}

resource "aws_eip" "nat_eip" {
  vpc        = true
  depends_on = ["aws_internet_gateway.platform"]
}

resource "aws_nat_gateway" "nat_gw" {
  allocation_id = "${aws_eip.nat_eip.id}"
  subnet_id     = "${aws_subnet.public.id}"
  depends_on    = ["aws_eip.nat_eip"]
}

###########################################
# load balancer
###########################################
resource "aws_elb" "kube_api" {
  connection_draining       = false
  cross_zone_load_balancing = true
  idle_timeout              = 400
  instances                 = ["${aws_instance.control_plane.id}"]
  name                      = "${var.stack_name}-elb"
  security_groups           = ["${aws_security_group.elb.id}"]
  subnets                   = ["${aws_subnet.public.0.id}"]

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    interval            = 30
    target              = "TCP:6443"
    timeout             = 3
    unhealthy_threshold = 6
  }
}

output "elb_address" {
  value = "${aws_elb.kube_api.dns_name}"
}

###########################################
# security
###########################################
resource "aws_security_group" "ssh" {
  description = "allow ssh traffic"
  name        = "${var.stack_name}-ssh"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-ssh",
    "Class", "SecurityGroup"))}"

  // allow traffic for TCP 22 from anywhere
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "lbports" {
  description = "allow load balancers to hit high ports"
  name        = "${var.stack_name}-lbports"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-lbport",
    "Class", "SecurityGroup"))}"

  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }
}

resource "aws_security_group" "icmp" {
  description = "allow ping between instances"
  name        = "${var.stack_name}-icmp"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-icmp",
    "Class", "SecurityGroup"))}"

  ingress {
    from_port       = -1
    to_port         = -1
    protocol        = "icmp"
    security_groups = []
    self            = true
  }

  egress {
    from_port       = -1
    to_port         = -1
    protocol        = "icmp"
    security_groups = []
    cidr_blocks     = ["${var.vpc_cidr_block}"]
  }
}

resource "aws_security_group" "egress" {
  description = "egress traffic"
  name        = "${var.stack_name}-egress"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-egress",
    "Class", "SecurityGroup"))}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "allow_https_apiserver" {
  description = "give access to 6443 port on the API servers"
  name        = "${var.stack_name}-allow-https-to-kubeapi"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-https",
    "Class", "SecurityGroup"))}"

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "allow_control_plane_traffic" {
  description = "give access to some traffic on the control plane hosts"
  name        = "${var.stack_name}-allow-control-plane-traffic"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-control-plane",
    "Class", "SecurityGroup"))}"

  ingress {
    from_port   = 2380
    to_port     = 2380
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 8285
    to_port     = 8285
    protocol    = "udp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 30000
    to_port     = 32768
    protocol    = "udp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }
}

resource "aws_security_group" "allow_workers_traffic" {
  description = "give access to some traffic on the workers"
  name        = "${var.stack_name}-allow-workers-traffic"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-control-plane",
    "Class", "SecurityGroup"))}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 8081
    to_port     = 8081
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 2380
    to_port     = 2380
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 8285
    to_port     = 8285
    protocol    = "udp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 30000
    to_port     = 32768
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  ingress {
    from_port   = 30000
    to_port     = 32768
    protocol    = "udp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }
}

# A security group for the ELB so it is accessible via the web
resource "aws_security_group" "elb" {
  name        = "${var.stack_name}-elb"
  description = "give access to kube api server"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-elb",
    "Class", "SecurityGroup"))}"

  # HTTP access from anywhere
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_key_pair" "kube" {
  key_name   = "${var.stack_name}-keypair"
  public_key = "${element(var.authorized_keys, 0)}"
}

###########################################
# control plane
###########################################
resource "aws_instance" "control_plane" {
  ami                         = "${data.aws_ami.latest_ami.id}"
  associate_public_ip_address = true
  count                       = "${var.masters}"
  instance_type               = "${var.master_size}"
  key_name                    = "${aws_key_pair.kube.key_name}"
  source_dest_check           = false
  subnet_id                   = "${aws_subnet.public.0.id}"
  user_data                   = "${data.template_cloudinit_config.cfg.rendered}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-master-${count.index}",
    "Class", "Instance"))}"

  vpc_security_group_ids = [
    "${aws_security_group.ssh.id}",
    "${aws_security_group.icmp.id}",
    "${aws_security_group.egress.id}",
    "${aws_security_group.allow_https_apiserver.id}",
    "${aws_security_group.allow_control_plane_traffic.id}",
  ]

  lifecycle {
    create_before_destroy = true

    # ignore_changes = ["associate_public_ip_address"]
  }

  root_block_device {
    volume_type           = "gp2"
    volume_size           = 20
    delete_on_termination = true
  }
}

resource "null_resource" "control_plane_wait_cloudinit" {
  depends_on = ["aws_instance.control_plane"]
  count      = "${var.masters}"

  connection {
    host  = "${element(aws_instance.control_plane.*.public_ip, count.index)}"
    user  = "ec2-user"
    type  = "ssh"
    agent = true
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null || /bin/true",
    ]
  }
}

output "control_plane.public_ip" {
  value = "${aws_instance.control_plane.*.public_ip}"
}

output "control_plane.private_dns" {
  value = "${aws_instance.control_plane.*.private_dns}"
}

###########################################
# workers
###########################################

resource "aws_instance" "nodes" {
  ami                         = "${data.aws_ami.latest_ami.id}"
  associate_public_ip_address = true
  count                       = "${var.workers}"
  instance_type               = "${var.worker_size}"
  key_name                    = "${aws_key_pair.kube.key_name}"
  source_dest_check           = false

  # TODO: remove from the public network once we have bastion hosts
  subnet_id = "${aws_subnet.public.0.id}"
  user_data = "${data.template_cloudinit_config.cfg.rendered}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-node-${count.index}",
    "Class", "Instance"))}"

  security_groups = [
    "${aws_security_group.ssh.id}",
    "${aws_security_group.icmp.id}",
    "${aws_security_group.egress.id}",
    "${aws_security_group.allow_workers_traffic.id}",
    "${aws_security_group.lbports.id}",
  ]

  lifecycle {
    create_before_destroy = true

    # ignore_changes = ["associate_public_ip_address"]
  }

  root_block_device {
    volume_type           = "gp2"
    volume_size           = 20
    delete_on_termination = true
  }
}

resource "null_resource" "nodes_wait_cloudinit" {
  depends_on = ["aws_instance.nodes"]
  count      = "${var.workers}"

  connection {
    host  = "${element(aws_instance.nodes.*.public_ip, count.index)}"
    user  = "ec2-user"
    type  = "ssh"
    agent = true
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null || /bin/true",
    ]
  }
}

output "nodes.public_ip" {
  value = "${aws_instance.nodes.*.public_ip}"
}

output "nodes.private_dns" {
  value = "${aws_instance.nodes.*.private_dns}"
}
