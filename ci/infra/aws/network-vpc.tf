# https://www.terraform.io/docs/providers/aws/r/vpc.html
resource "aws_vpc" "platform" {
  cidr_block           = var.cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-vpc"
      "Class" = "VPC"
    },
  )
}

resource "aws_internet_gateway" "platform" {
  vpc_id     = aws_vpc.platform.id
  depends_on = [aws_vpc.platform,]

  tags = merge(
    local.tags,
    {
      "Class" = "Gateway"
    },
  )
}

resource "aws_main_route_table_association" "main" {
  route_table_id = aws_route_table.public.id
  vpc_id         = aws_vpc.platform.id
}

resource "aws_vpc_dhcp_options" "platform" {
  domain_name         = "${var.aws_region}.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = merge(
    local.tags,
    {
      "Class" = "VPCDHCP"
    },
  )
}

resource "aws_vpc_dhcp_options_association" "platform" {
  vpc_id          = aws_vpc.platform.id
  dhcp_options_id = aws_vpc_dhcp_options.platform.id
}
