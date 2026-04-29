variable "region" {
  description = "AWS region to deploy into."
  type        = string
  default     = "us-east-1"
}

variable "project" {
  description = "Project tag and resource-name prefix."
  type        = string
  default     = "test-ecs"
}

variable "environment" {
  description = "Environment tag (test, staging, prod)."
  type        = string
  default     = "test"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC."
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for the two public subnets, one per AZ."
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for the two private subnets, one per AZ."
  type        = list(string)
  default     = ["10.0.10.0/24", "10.0.20.0/24"]
}

variable "container_port" {
  description = "Port the application listens on inside the container."
  type        = number
  default     = 3000
}

variable "task_cpu" {
  description = "Fargate task CPU units."
  type        = number
  default     = 256
}

variable "task_memory" {
  description = "Fargate task memory in MiB."
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "Number of ECS task replicas."
  type        = number
  default     = 1
}

variable "log_retention_days" {
  description = "Days to retain application logs in CloudWatch."
  type        = number
  default     = 7
}

variable "db_instance_class" {
  description = "RDS instance class."
  type        = string
  default     = "db.t4g.micro"
}

variable "db_allocated_storage" {
  description = "RDS allocated storage in GB."
  type        = number
  default     = 20
}

variable "db_name" {
  description = "Initial database name."
  type        = string
  default     = "jobs"
}

variable "db_username" {
  description = "Master DB username."
  type        = string
  default     = "jobs"
}

variable "ecr_repository_name" {
  description = "Name of the ECR repository created by the bootstrap layer."
  type        = string
  default     = "test-ecs"
}

variable "placeholder_image" {
  description = "Image used on first apply, before CI pushes the real one."
  type        = string
  default     = "public.ecr.aws/nginx/nginx:stable"
}

variable "image_tag" {
  description = "Tag of the application image to deploy."
  type        = string
  default     = ""
}

variable "github_org" {
  description = "GitHub organisation/user that owns the repo (for OIDC trust)."
  type        = string
  default     = "GithinjiC"
}

variable "github_repo" {
  description = "GitHub repository name (for OIDC trust)."
  type        = string
  default     = "mpesa-devops-assessment"
}

variable "github_branch" {
  description = "Branch allowed to assume the deployment role via OIDC."
  type        = string
  default     = "main"
}

variable "github_deploy_environment" {
  description = "Name of the GitHub Actions environment used by the deploy job."
  type        = string
  default     = "production"
}
