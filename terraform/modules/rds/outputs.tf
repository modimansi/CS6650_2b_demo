output "db_instance_id" {
  description = "The RDS instance ID"
  value       = aws_db_instance.mysql.id
}

output "db_instance_endpoint" {
  description = "The connection endpoint (hostname:port)"
  value       = aws_db_instance.mysql.endpoint
}

output "db_instance_address" {
  description = "The hostname of the RDS instance"
  value       = aws_db_instance.mysql.address
}

output "db_instance_port" {
  description = "The port the database is listening on"
  value       = aws_db_instance.mysql.port
}

output "db_name" {
  description = "The name of the database"
  value       = aws_db_instance.mysql.db_name
}

output "db_username" {
  description = "The master username for the database"
  value       = aws_db_instance.mysql.username
  sensitive   = true
}

output "db_password" {
  description = "The master password for the database"
  value       = var.database_password != "" ? var.database_password : random_password.db_password[0].result
  sensitive   = true
}

output "db_security_group_id" {
  description = "Security group ID for the RDS instance"
  value       = aws_security_group.rds.id
}

output "connection_string" {
  description = "Full PostgreSQL connection string for the application"
  value       = "postgres://${aws_db_instance.mysql.username}:${urlencode(var.database_password != "" ? var.database_password : random_password.db_password[0].result)}@${aws_db_instance.mysql.endpoint}/${aws_db_instance.mysql.db_name}?sslmode=require"
  sensitive   = true
}

