# VPC + subnets + NAT gateway
# Uncomment and fill in when beginning Phase 3 (cloud deployment).

# resource "aws_vpc" "main" {
#   cidr_block           = "10.0.0.0/16"
#   enable_dns_hostnames = true
#   tags = merge(local.tags, { Name = "${local.name}-vpc" })
# }
#
# resource "aws_internet_gateway" "main" {
#   vpc_id = aws_vpc.main.id
#   tags   = merge(local.tags, { Name = "${local.name}-igw" })
# }
#
# resource "aws_subnet" "public" {
#   count                   = 2
#   vpc_id                  = aws_vpc.main.id
#   cidr_block              = cidrsubnet("10.0.0.0/16", 8, count.index)
#   availability_zone       = data.aws_availability_zones.available.names[count.index]
#   map_public_ip_on_launch = true
#   tags = merge(local.tags, { Name = "${local.name}-public-${count.index}" })
# }
#
# resource "aws_subnet" "private" {
#   count             = 2
#   vpc_id            = aws_vpc.main.id
#   cidr_block        = cidrsubnet("10.0.0.0/16", 8, count.index + 10)
#   availability_zone = data.aws_availability_zones.available.names[count.index]
#   tags = merge(local.tags, { Name = "${local.name}-private-${count.index}" })
# }
#
# data "aws_availability_zones" "available" {}
