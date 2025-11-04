# RDS MySQL Database for Shopping Cart Service

# DB Subnet Group - RDS requires at least 2 subnets in different AZs
resource "aws_db_subnet_group" "main" {
  name       = "${lower(var.service_name)}-db-subnet-group"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.service_name}-db-subnet-group"
  }
}

# Security Group for RDS - Only allow access from ECS tasks
resource "aws_security_group" "rds" {
  name        = "${var.service_name}-rds-sg"
  description = "Security group for RDS PostgreSQL - only accessible from ECS tasks"
  vpc_id      = var.vpc_id

  # Allow PostgreSQL traffic from ECS tasks only
  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.ecs_security_group_id]
    description     = "Allow PostgreSQL from ECS tasks"
  }

  # RDS doesn't need outbound rules typically, but adding for completeness
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound (for updates/patches)"
  }

  tags = {
    Name = "${var.service_name}-rds-sg"
  }
}

# RDS PostgreSQL Instance (for shopping cart)
resource "aws_db_instance" "mysql" {
  # Instance settings
  identifier     = "${lower(var.service_name)}-postgres-db"
  engine         = "postgres"
  engine_version = "15"  # PostgreSQL 15 (let AWS choose latest minor version)
  instance_class = "db.t3.micro"  # Free tier eligible

  # Storage
  allocated_storage     = 20  # GB - Free tier allows up to 20GB
  max_allocated_storage = 0   # Disable autoscaling for assignment
  storage_type          = "gp2"  # General Purpose SSD
  storage_encrypted     = false  # Encryption not required for assignment

  # Database configuration
  db_name  = var.database_name
  username = var.database_username
  password = var.database_password != "" ? var.database_password : random_password.db_password[0].result
  port     = 5432

  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false  # Private subnet only

  # High availability - disabled for free tier
  multi_az = false

  # Backup configuration - for assignment, keep minimal
  backup_retention_period = 0  # No automated backups (saves cost/time)
  backup_window           = null
  maintenance_window      = "sun:04:00-sun:05:00"  # Sunday 4-5 AM UTC

  # Assignment-friendly settings
  skip_final_snapshot       = true   # Don't create snapshot on destroy
  deletion_protection       = false  # Allow terraform destroy
  delete_automated_backups  = true   # Clean up backups on destroy
  copy_tags_to_snapshot     = false
  
  # Performance and monitoring
  monitoring_interval          = 0      # Disable enhanced monitoring (not free tier)
  enabled_cloudwatch_logs_exports = []  # Disable CloudWatch logs (not free tier)
  performance_insights_enabled = false  # Disable (not free tier)

  # Parameter group - use default for PostgreSQL 15
  parameter_group_name = "default.postgres15"

  # Apply changes immediately (for development/assignment)
  apply_immediately = true

  # Allow major version upgrades (if needed)
  allow_major_version_upgrade = true
  auto_minor_version_upgrade  = false  # Control when updates happen

  tags = {
    Name        = "${var.service_name}-postgres-db"
    Environment = "assignment"
    ManagedBy   = "terraform"
  }
}

# Random password generation (if not provided)
resource "random_password" "db_password" {
  count   = var.database_password == "" ? 1 : 0
  length  = 16
  special = true
  # Avoid characters that might cause issues in connection strings
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

