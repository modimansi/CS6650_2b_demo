# Use existing LabRole (for AWS Learner Lab / Academy)
# Cannot create IAM roles in Learner Lab, must use pre-existing LabRole
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Lambda function
resource "aws_lambda_function" "processor" {
  function_name = "${var.service_name}-payment-processor"
  role          = data.aws_iam_role.lab_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2" # Custom runtime for Go
  
  filename         = var.lambda_zip_path
  source_code_hash = filebase64sha256(var.lambda_zip_path)

  memory_size = var.memory_size
  timeout     = var.timeout

  # Note: AWS_REGION is automatically provided by Lambda runtime
  # No need to set environment variables for this function

  tags = {
    Name = "${var.service_name}-payment-processor"
  }
}

# SNS subscription for Lambda
resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = var.sns_topic_arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.processor.arn
}

# Allow SNS to invoke Lambda
resource "aws_lambda_permission" "allow_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = var.sns_topic_arn
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${aws_lambda_function.processor.function_name}"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "${var.service_name}-lambda-logs"
  }
}

