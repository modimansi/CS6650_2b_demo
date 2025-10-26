variable "service_name" {
  type        = string
  description = "Base name for ALB resources"
}

variable "container_port" {
  type        = number
  description = "Port the target group forwards to"
}

variable "health_check_path" {
  type        = string
  description = "Health check path for targets"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for ALB"
}

variable "public_subnet_ids" {
  type        = list(string)
  description = "Public subnet IDs for ALB"
}

variable "alb_security_group_id" {
  type        = string
  description = "Security group ID for ALB"
}

