# ECS cluster + task definitions for api, router, worker services
# Uncomment when beginning Phase 3.

# resource "aws_ecs_cluster" "main" {
#   name = local.name
#   setting {
#     name  = "containerInsights"
#     value = "enabled"
#   }
#   tags = local.tags
# }
#
# resource "aws_ecs_task_definition" "api" {
#   family                   = "${local.name}-api"
#   requires_compatibilities = ["FARGATE"]
#   network_mode             = "awsvpc"
#   cpu                      = 256
#   memory                   = 512
#   execution_role_arn       = aws_iam_role.ecs_execution.arn
#   task_role_arn            = aws_iam_role.ecs_task.arn
#
#   container_definitions = jsonencode([{
#     name      = "api"
#     image     = "${aws_ecr_repository.api.repository_url}:latest"
#     portMappings = [{ containerPort = 8080, protocol = "tcp" }]
#     environment = [
#       { name = "PORT",         value = "8080" },
#       { name = "ROUTER_URL",   value = "http://router.${local.name}.local:8081" },
#     ]
#     secrets = [
#       { name = "DATABASE_URL", valueFrom = aws_ssm_parameter.db_url.arn },
#       { name = "REDIS_URL",    valueFrom = aws_ssm_parameter.redis_url.arn },
#       { name = "JWT_SECRET",   valueFrom = aws_ssm_parameter.jwt_secret.arn },
#     ]
#     logConfiguration = {
#       logDriver = "awslogs"
#       options = {
#         awslogs-group         = "/ecs/${local.name}/api"
#         awslogs-region        = var.aws_region
#         awslogs-stream-prefix = "ecs"
#       }
#     }
#   }])
# }
