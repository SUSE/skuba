resource "aws_vpc" "main" {
  cidr_block                       = "10.0.0.0/16"
  enable_dns_support               = true
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = false

  tags = {
    Name        = "${var.stack_name}"
    Environment = "${var.stack_name}"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.main.id}"

  tags {
    Name        = "${var.stack_name}"
    Environment = "${var.stack_name}"
  }
}

resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }

  tags {
    Name        = "${var.stack_name}"
    Environment = "${var.stack_name}"
  }
}

resource "aws_subnet" "public" {
  count                   = "${length(var.availability_zones)}"
  vpc_id                  = "${aws_vpc.main.id}"
  cidr_block              = "${cidrsubnet("10.0.0.0/16", 8, count.index)}"
  availability_zone       = "${element(var.availability_zones, count.index)}"
  map_public_ip_on_launch = true

  tags {
    Name        = "${var.stack_name}-az-${element(var.availability_zones, count.index)}"
    Environment = "${var.stack_name}"
  }
}

resource "aws_route_table_association" "public" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.public.*.id, count.index)}"
  route_table_id = "${aws_route_table.r.id}"
}
