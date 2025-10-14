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


