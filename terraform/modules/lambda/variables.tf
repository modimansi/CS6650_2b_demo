variable "service_name" {
  type        = string
  description = "Base name for resources"
}

variable "lambda_zip_path" {
  type        = string
  description = "Path to Lambda deployment package (function.zip)"
}

variable "sns_topic_arn" {
  type        = string
  description = "SNS topic ARN to subscribe to"
}

variable "memory_size" {
  type        = number
  default     = 512
  description = "Lambda memory in MB (128-10240)"
}

variable "timeout" {
  type        = number
  default     = 10
  description = "Lambda timeout in seconds (max 900)"
}

variable "region" {
  type        = string
  description = "AWS region"
}

variable "log_retention_days" {
  type        = number
  default     = 7
  description = "CloudWatch log retention in days"
}

