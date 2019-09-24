resource "aws_vpc_peering_connection" "tunnel" {
  count       = "${length(var.peer_vpc_ids)}"
  peer_vpc_id = "${var.peer_vpc_ids[count.index]}"
  vpc_id      = "${aws_vpc.platform.id}"
  auto_accept = true
  tags        = "${merge(local.tags, map("Class", "VPC-peering-connection"))}"

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

data "aws_vpc" "peer" {
  count = "${length(var.peer_vpc_ids)}"
  id    = "${var.peer_vpc_ids[count.index]}"
}

resource "aws_route" "peer_to_k8s" {
  count                     = "${length(var.peer_vpc_ids)}"
  route_table_id            = "${element(data.aws_vpc.peer.*.main_route_table_id, count.index)}"
  destination_cidr_block    = "${aws_vpc.platform.cidr_block}"
  vpc_peering_connection_id = "${element(aws_vpc_peering_connection.tunnel.*.id, count.index)}"
}

resource "aws_route" "k8s_to_peer" {
  count                     = "${length(var.peer_vpc_ids)}"
  route_table_id            = "${aws_route_table.public.id}"
  destination_cidr_block    = "${element(data.aws_vpc.peer.*.cidr_block, count.index)}"
  vpc_peering_connection_id = "${element(aws_vpc_peering_connection.tunnel.*.id, count.index)}"
}
