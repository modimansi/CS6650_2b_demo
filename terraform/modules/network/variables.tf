variable "service_name" {
  description = "Base name for network resources"
  type        = string
}

variable "container_port" {
  description = "Port to expose on the security groups"
  type        = number
}
