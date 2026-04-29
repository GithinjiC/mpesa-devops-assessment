resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${local.name}/app"
  retention_in_days = var.log_retention_days

  tags = { Name = "${local.name}-app-logs" }
}
