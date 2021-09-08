resource "aws_subnet" "private" {
  count                   = length(var.aws_availability_zones)
  vpc_id                  = aws_vpc.platform.id
  availability_zone       = element(var.aws_availability_zones, count.index)
  cidr_block              = cidrsubnet(var.cidr_block, 8, count.index + length(var.aws_availability_zones))
  map_public_ip_on_launch = false

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-subnet-private-${element(var.aws_availability_zones, count.index)}"
      "Class" = "Subnet"
    },
  )
}

resource "aws_route_table" "private" {
  count  = length(var.aws_availability_zones)
  vpc_id = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-route-table-private-${element(var.aws_availability_zones, count.index)}"
      "Class" = "RouteTable"
    },
  )
}

resource "aws_route_table_association" "private" {
  count = length(var.aws_availability_zones)

  route_table_id = element(aws_route_table.private.*.id, count.index)
  subnet_id      = element(aws_subnet.private.*.id, count.index)
}

resource "aws_route" "private" {
  count                  = length(var.aws_availability_zones)
  route_table_id         = element(aws_route_table.private.*.id, count.index)
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = element(aws_nat_gateway.public.*.id, count.index)
}
