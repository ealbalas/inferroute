# IAM roles for ECS tasks — uncomment when beginning Phase 3.

# data "aws_iam_policy_document" "ecs_assume" {
#   statement {
#     actions = ["sts:AssumeRole"]
#     principals {
#       type        = "Service"
#       identifiers = ["ecs-tasks.amazonaws.com"]
#     }
#   }
# }
#
# resource "aws_iam_role" "ecs_execution" {
#   name               = "${local.name}-ecs-execution"
#   assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
#   managed_policy_arns = [
#     "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
#   ]
#   tags = local.tags
# }
#
# resource "aws_iam_role" "ecs_task" {
#   name               = "${local.name}-ecs-task"
#   assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
#   tags               = local.tags
# }
#
# Inline policy grants least-privilege access to SQS, SSM, and CloudWatch.
# resource "aws_iam_role_policy" "ecs_task_policy" {
#   name   = "inferroute-task-policy"
#   role   = aws_iam_role.ecs_task.id
#   policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [
#       {
#         Effect   = "Allow"
#         Action   = ["sqs:ReceiveMessage", "sqs:DeleteMessage", "sqs:SendMessage"]
#         Resource = aws_sqs_queue.events.arn
#       },
#       {
#         Effect   = "Allow"
#         Action   = ["ssm:GetParameter"]
#         Resource = "arn:aws:ssm:*:*:parameter/inferroute/*"
#       },
#     ]
#   })
# }
