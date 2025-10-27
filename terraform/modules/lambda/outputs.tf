output "function_name" {
  value       = aws_lambda_function.processor.function_name
  description = "Lambda function name"
}

output "function_arn" {
  value       = aws_lambda_function.processor.arn
  description = "Lambda function ARN"
}

output "log_group_name" {
  value       = aws_cloudwatch_log_group.lambda.name
  description = "CloudWatch log group name"
}

output "role_arn" {
  value       = data.aws_iam_role.lab_role.arn
  description = "Lambda execution role ARN (LabRole)"
}

