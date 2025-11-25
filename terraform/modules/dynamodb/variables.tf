variable "service_name" {
  type        = string
  description = "Name of the service (used for resource naming)"
}

variable "enable_point_in_time_recovery" {
  type        = bool
  default     = false
  description = "Enable point-in-time recovery for the DynamoDB table"
}

variable "create_metrics_table" {
  type        = bool
  default     = false
  description = "Create a separate table for storing performance metrics"
}

