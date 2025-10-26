# SNS Topic for order events
resource "aws_sns_topic" "orders" {
  name = "order-processing-events"

  tags = {
    Name = "order-processing-events"
  }
}

# SQS Queue for order processing
resource "aws_sqs_queue" "orders" {
  name                       = "order-processing-queue"
  visibility_timeout_seconds = 30     # 30 seconds
  message_retention_seconds  = 345600 # 4 days (4 * 24 * 60 * 60)
  receive_wait_time_seconds  = 20     # Long polling (20 seconds)

  tags = {
    Name = "order-processing-queue"
  }
}

# SQS Queue Policy to allow SNS to send messages
resource "aws_sqs_queue_policy" "orders" {
  queue_url = aws_sqs_queue.orders.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "SQS:SendMessage"
        Resource = aws_sqs_queue.orders.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.orders.arn
          }
        }
      }
    ]
  })
}

# Subscribe SQS queue to SNS topic
resource "aws_sns_topic_subscription" "orders" {
  topic_arn = aws_sns_topic.orders.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.orders.arn
}

