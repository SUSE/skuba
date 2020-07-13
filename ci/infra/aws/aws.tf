locals {
  # Do not add the special `kubernetes.io/cluster<cluster-name>` here,
  # this tag cannot be added to all our resources otherwise the CPI
  # will get confused when dealing with security rules objects.
  basic_tags = merge(
    {
      "Name"        = var.stack_name
      "Environment" = var.stack_name
    },
    var.tags,
  )

  tags = merge(
    local.basic_tags,
    {
      format("kubernetes.io/cluster/%v", var.stack_name) = "SUSE-terraform"
    },
  )
}

provider "aws" {
  profile = "default"
}

resource "aws_key_pair" "kube" {
  key_name   = "${var.stack_name}-keypair"
  public_key = element(var.authorized_keys, 0)

  tags = merge(
    local.basic_tags,
    {
      "Name"  = "${var.stack_name}-keypair"
      "Class" = "KeyPair"
    },
  )
}

resource "aws_resourcegroups_group" "kube" {
  name = "${var.stack_name}-resourcegroup"

  tags = merge(
    local.basic_tags,
    {
      "Name"  = "${var.stack_name}-resourcegroup"
      "Class" = "ResourceGroup"
    },
  )

  resource_query {
    query = jsonencode({
      "ResourceTypeFilters" : [
        "AWS::EC2::DHCPOptions",
        "AWS::EC2::EIP",
        "AWS::EC2::Instance",
        "AWS::EC2::InternetGateway",
        "AWS::EC2::NatGateway",
        "AWS::EC2::NetworkInterface",
        "AWS::EC2::RouteTable",
        "AWS::EC2::SecurityGroup",
        "AWS::EC2::Subnet",
        "AWS::EC2::VPC",
        "AWS::EC2::VPCPeeringConnection",
        "AWS::ElasticLoadBalancing::LoadBalancer",
        "AWS::ResourceGroups::Group"
      ],
      "TagFilters" : [
        {
          "Key" : "Environment",
          "Values" : [var.stack_name]
        }
      ]
    })
  }
}
