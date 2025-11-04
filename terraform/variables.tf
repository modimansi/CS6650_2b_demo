# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "ecr_service"
}

variable "service_name" {
  type    = string
  default = "CS6650L2"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type        = number
  default     = 1
  description = "Number of ECS tasks to run"
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Number of concurrent payment processing workers
variable "worker_count" {
  type        = number
  default     = 1
  description = "Number of concurrent payment processing workers (goroutines). Increase to scale throughput."
}

# RDS Database Configuration
variable "database_name" {
  type        = string
  default     = "shopping"
  description = "Name of the database to create"
}

variable "database_username" {
  type        = string
  default     = "dbadmin"
  description = "Master username for the database"
}

variable "database_password" {
  type        = string
  default     = ""
  description = "Master password for the database (leave empty for auto-generated password)"
  sensitive   = true
}