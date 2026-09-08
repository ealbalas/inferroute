variable "aws_region" {
  description = "AWS region for all resources."
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Deployment environment (dev | staging | prod)."
  type        = string
  default     = "dev"
}

variable "db_password" {
  description = "Password for the RDS PostgreSQL instance. Set via TF_VAR_db_password — never commit."
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret for dashboard auth. Set via TF_VAR_jwt_secret."
  type        = string
  sensitive   = true
}
