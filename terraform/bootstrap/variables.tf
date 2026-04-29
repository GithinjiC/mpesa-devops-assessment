variable "region" {
  description = "AWS region for the bootstrap resources."
  type        = string
  default     = "us-east-1"
}

variable "project" {
  description = "Project tag and resource-name prefix."
  type        = string
  default     = "test-ecs"
}

variable "ecr_repo_name" {
  description = "ECR repository name for the application image."
  type        = string
  default     = "test-ecs"
}
