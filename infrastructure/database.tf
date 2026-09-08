# RDS PostgreSQL 16 — uncomment when beginning Phase 3.

# resource "aws_db_subnet_group" "main" {
#   name       = local.name
#   subnet_ids = aws_subnet.private[*].id
#   tags       = local.tags
# }
#
# resource "aws_db_instance" "postgres" {
#   identifier             = local.name
#   engine                 = "postgres"
#   engine_version         = "16"
#   instance_class         = "db.t4g.small"
#   allocated_storage      = 20
#   db_name                = "inferroute"
#   username               = "inferroute"
#   password               = var.db_password
#   db_subnet_group_name   = aws_db_subnet_group.main.name
#   vpc_security_group_ids = [aws_security_group.rds.id]
#   skip_final_snapshot    = true
#   deletion_protection    = false
#   storage_encrypted      = true
#   tags                   = local.tags
# }
