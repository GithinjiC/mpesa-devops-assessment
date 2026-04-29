# URL-safe random password
resource "random_password" "db" {
  length           = 32
  special          = true
  override_special = "-_~"
}

resource "aws_secretsmanager_secret" "db" {
  name                    = "${local.name}/db"
  description             = "Postgres credentials and connection metadata for ${local.name}."
  recovery_window_in_days = 0

  tags = { Name = "${local.name}-db-secret" }
}

resource "aws_secretsmanager_secret_version" "db" {
  secret_id = aws_secretsmanager_secret.db.id

  secret_string = jsonencode({
    username = var.db_username
    password = random_password.db.result
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    dbname   = var.db_name

    DATABASE_URL = "postgres://${var.db_username}:${random_password.db.result}@${aws_db_instance.main.address}:${aws_db_instance.main.port}/${var.db_name}?sslmode=require"
  })
}
