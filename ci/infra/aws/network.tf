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
}

resource "aws_route" "public_to_everywhere" {
  route_table_id         = "${aws_route_table.public.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.platform.id}"
}

resource "aws_route_table" "private" {
  vpc_id = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-route-table-private",
    "Class", "RouteTable"))}"
}

resource "aws_route" "private_to_everywhere" {
  route_table_id         = "${aws_route_table.private.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.platform.id}"
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
