# DynamoDB Table for Shopping Carts
# Design optimized for shopping cart access patterns

resource "aws_dynamodb_table" "shopping_carts" {
  name           = "${var.service_name}-shopping-carts"
  billing_mode   = "PAY_PER_REQUEST"  # On-demand pricing for variable load
  hash_key       = "cart_id"          # Partition key for even distribution
  
  # Partition Key: cart_id
  # Choice: cart_id ensures even distribution since cart IDs are sequential/random
  # Alternative considered: customer_id (rejected - would create hot partitions for active customers)
  attribute {
    name = "cart_id"
    type = "S"  # String type (we'll use UUID for better distribution)
  }
  
  # Global Secondary Index for customer lookup
  # Access Pattern: "Get all carts for a customer"
  # Trade-off: Additional storage cost, but enables customer history queries
  attribute {
    name = "customer_id"
    type = "N"  # Number type
  }
  
  global_secondary_index {
    name            = "CustomerIndex"
    hash_key        = "customer_id"
    projection_type = "ALL"  # Project all attributes for complete cart data
    # Note: With PAY_PER_REQUEST, capacity is managed automatically
  }
  
  # Time-To-Live for automatic cart expiration
  # Abandoned carts auto-delete after 7 days
  ttl {
    attribute_name = "expiration_time"
    enabled        = true
  }
  
  # Point-in-time recovery for data protection
  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }
  
  # Server-side encryption
  server_side_encryption {
    enabled = true
  }
  
  # Tags for resource management
  tags = {
    Name        = "${var.service_name}-shopping-carts"
    Environment = "assignment"
    ManagedBy   = "terraform"
    Database    = "DynamoDB"
  }
}

# DynamoDB table for performance comparison metrics
# Optional: Store test results for comparison
resource "aws_dynamodb_table" "performance_metrics" {
  count = var.create_metrics_table ? 1 : 0
  
  name         = "${var.service_name}-cart-metrics"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "test_id"
  range_key    = "timestamp"
  
  attribute {
    name = "test_id"
    type = "S"
  }
  
  attribute {
    name = "timestamp"
    type = "N"
  }
  
  ttl {
    attribute_name = "expiration_time"
    enabled        = true
  }
  
  tags = {
    Name        = "${var.service_name}-cart-metrics"
    Environment = "assignment"
    ManagedBy   = "terraform"
  }
}

