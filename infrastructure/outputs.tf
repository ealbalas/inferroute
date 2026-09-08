output "alb_dns" {
  description = "Public DNS of the Application Load Balancer."
  value       = "TODO: wire up after ALB resource is created"
}

output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint."
  value       = "TODO: wire up after RDS resource is created"
}

output "redis_endpoint" {
  description = "ElastiCache Redis endpoint."
  value       = "TODO: wire up after ElastiCache resource is created"
}
