variable "service_name" {
  type        = string
  description = "Name of the service (used for resource naming)"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where RDS will be deployed"
}

variable "private_subnet_ids" {
  type        = list(string)
  description = "List of private subnet IDs for RDS (must be in different AZs)"
}

variable "ecs_security_group_id" {
  type        = string
  description = "Security group ID of ECS tasks (to allow connection to RDS)"
}

variable "database_name" {
  type        = string
  description = "Initial database name"
  default     = "shopping"
}

variable "database_username" {
  type        = string
  description = "Master username for the database"
  default     = "dbadmin"
}

variable "database_password" {
  type        = string
  description = "Master password for the database (leave empty to auto-generate)"
  default     = ""
  sensitive   = true
}

