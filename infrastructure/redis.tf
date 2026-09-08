# ElastiCache Redis 7 — uncomment when beginning Phase 3.

# resource "aws_elasticache_subnet_group" "main" {
#   name       = local.name
#   subnet_ids = aws_subnet.private[*].id
# }
#
# resource "aws_elasticache_replication_group" "redis" {
#   replication_group_id = local.name
#   description          = "InferRoute Redis cache and rate limiter"
#   node_type            = "cache.t4g.small"
#   num_cache_clusters   = 1
#   engine_version       = "7.1"
#   subnet_group_name    = aws_elasticache_subnet_group.main.name
#   security_group_ids   = [aws_security_group.redis.id]
#   at_rest_encryption_enabled = true
#   transit_encryption_enabled = true
#   tags                 = local.tags
# }
