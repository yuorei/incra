resource "aws_cloudwatch_log_group" "this" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = var.log_retention_days
}

resource "aws_lambda_function" "this" {
  filename      = var.filename
  function_name = var.function_name
  role          = var.role_arn
  handler       = var.handler
  runtime       = var.runtime

  environment {
    variables = var.environment_variables
  }
}
