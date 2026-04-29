output "alb_dns_name" {
  description = "Public DNS name of the ALB"
  value       = aws_lb.app.dns_name
}

output "ecr_repository_url" {
  description = "ECR repo URL"
  value       = data.aws_ecr_repository.app.repository_url
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.app.name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.app.name
}

output "ecs_task_family" {
  description = "Task-definition family"
  value       = aws_ecs_task_definition.app.family
}

output "github_actions_role_arn" {
  description = "Role ARN that GitHub Actions assumes via OIDC"
  value       = aws_iam_role.github_actions.arn
}

output "db_secret_arn" {
  description = "Secrets Manager ARN holding DB credentials"
  value       = aws_secretsmanager_secret.db.arn
}

output "log_group_name" {
  description = "CloudWatch log group name for the app container"
  value       = aws_cloudwatch_log_group.app.name
}
