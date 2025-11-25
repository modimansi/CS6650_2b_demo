output "table_name" {
  description = "Name of the DynamoDB shopping carts table"
  value       = aws_dynamodb_table.shopping_carts.name
}

output "table_arn" {
  description = "ARN of the DynamoDB shopping carts table"
  value       = aws_dynamodb_table.shopping_carts.arn
}

output "table_id" {
  description = "ID of the DynamoDB shopping carts table"
  value       = aws_dynamodb_table.shopping_carts.id
}

output "customer_index_name" {
  description = "Name of the customer GSI"
  value       = "CustomerIndex"
}

output "metrics_table_name" {
  description = "Name of the performance metrics table (if created)"
  value       = var.create_metrics_table ? aws_dynamodb_table.performance_metrics[0].name : ""
}

