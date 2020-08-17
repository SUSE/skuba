resource "aws_subnet" "public" {
  count                   = length(var.aws_availability_zones)
  vpc_id                  = aws_vpc.platform.id
  availability_zone       = element(var.aws_availability_zones, count.index)
  cidr_block              = cidrsubnet(var.cidr_block, 8, count.index)
  map_public_ip_on_launch = true
  # depends_on              = [aws_main_route_table_association.platform,]

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-subnet-public-${element(var.aws_availability_zones, count.index)}"
      "Class" = "VPC"
    },
  )
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-route-table-public"
      "Class" = "RouteTable"
    },
  )
}

resource "aws_route_table_association" "public" {
  count = length(var.aws_availability_zones)

  route_table_id = element(aws_route_table.public.*.id, count.index)
  subnet_id      = element(aws_subnet.public.*.id, count.index)
}

resource "aws_route" "public" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.platform.id
}

resource "aws_eip" "eip" {
  count = length(var.aws_availability_zones)
  vpc   = true
  # depends_on = [aws_internet_gateway.platform,]

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-eip-eip"
      "Class" = "ElasticIP"
    },
  )
}

resource "aws_nat_gateway" "public" {
  count         = length(var.aws_availability_zones)
  subnet_id     = element(aws_subnet.public.*.id, count.index)
  allocation_id = element(aws_eip.eip.*.id, count.index)
  # depends_on = [aws_eip.eip,]

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-nat_gateway-${element(var.aws_availability_zones, count.index)}"
      "Class" = "NatGateway"
    },
  )
}
