resource "aws_resourcegroups_group" "kube" {
  count = var.enable_resource_group ? 1 : 0
  name  = "${var.stack_name}-resourcegroup"

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

  tags = merge(
    local.basic_tags,
    {
      "Name"  = "${var.stack_name}-resourcegroup"
      "Class" = "ResourceGroup"
    },
  )
}