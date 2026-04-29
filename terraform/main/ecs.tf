resource "aws_ecs_cluster" "app" {
  name = "${local.name}-cluster"

  setting {
    name  = "containerInsights"
    value = "disabled" # switch to enabled for richer observability
  }
}

resource "aws_ecs_cluster_capacity_providers" "app" {
  cluster_name       = aws_ecs_cluster.app.name
  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
    base              = 1
  }
}

# Task definition
locals {
  # Use the placeholder image on first apply (CI hasn't pushed yet).
  app_image = var.image_tag == "" ? var.placeholder_image : "${data.aws_ecr_repository.app.repository_url}:${var.image_tag}"
}

resource "aws_ecs_task_definition" "app" {
  family                   = "${local.name}-app"
  cpu                      = tostring(var.task_cpu)
  memory                   = tostring(var.task_memory)
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]

  execution_role_arn = aws_iam_role.task_exec.arn
  task_role_arn      = aws_iam_role.task.arn

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }

  container_definitions = jsonencode([
    {
      name      = "app"
      image     = local.app_image
      essential = true

      portMappings = [
        {
          containerPort = var.container_port
          protocol      = "tcp"
        }
      ]

      environment = [
        { name = "PORT", value = tostring(var.container_port) },
        { name = "MIGRATIONS_DIR", value = "migrations" },
        { name = "LOG_LEVEL", value = "info" },
        { name = "SERVICE_NAME", value = local.name },
        { name = "HTTP_LISTEN_HOST", value = "0.0.0.0" },
        { name = "HTTP_READ_HEADER_TIMEOUT", value = "10s" },
        { name = "HTTP_SHUTDOWN_TIMEOUT", value = "15s" },
        { name = "HTTP_REQUEST_TIMEOUT", value = "60s" },
        # LOG_FILE_PATH intentionally unset → stdout-only for CloudWatch.
      ]

      secrets = [
        {
          name      = "DATABASE_URL"
          valueFrom = "${aws_secretsmanager_secret.db.arn}:DATABASE_URL::"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.app.name
          awslogs-region        = local.region
          awslogs-stream-prefix = "app"
        }
      }
    }
  ])

  # CI updates the image
  lifecycle {
    ignore_changes = [container_definitions]
  }
}

# Service
resource "aws_ecs_service" "app" {
  name            = "${local.name}-svc"
  cluster         = aws_ecs_cluster.app.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  enable_execute_command = true

  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.ecs.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "app"
    container_port   = var.container_port
  }

  # CI rolls task-definition revisions
  lifecycle {
    ignore_changes = [task_definition, desired_count]
  }

  depends_on = [aws_lb_listener.http]
}
