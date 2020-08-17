output "username" {
  value = var.username
}

output "elb_address" {
  value = aws_elb.elb.dns_name
}

output "vpc_id" {
  value = aws_vpc.platform.id
}

output "public_subnets" {
  value = aws_subnet.public.*.id
}

output "public_cidrs" {
  value = aws_subnet.public.*.cidr_block
}

output "private_subnets" {
  value = aws_subnet.private.*.id
}

output "private_cidrs" {
  value = aws_subnet.private.*.cidr_block
}

output "masters_public_ip" {
  value = zipmap(
    aws_instance.master.*.id,
    aws_instance.master.*.public_ip,
  )
}

output "masters_private_dns" {
  value = zipmap(
    aws_instance.master.*.id,
    aws_instance.master.*.private_dns,
  )
}

output "workers_public_ip" {
  value = zipmap(
    aws_instance.worker.*.id,
    aws_instance.worker.*.public_ip,
  )
}

output "workers_private_dns" {
  value = zipmap(
    aws_instance.worker.*.id,
    aws_instance.worker.*.private_dns
  )
}
