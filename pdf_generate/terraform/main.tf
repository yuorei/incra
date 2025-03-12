provider "aws" {
  region = "ap-northeast-1"

  # assume_role {
  #   role_arn     = "arn:aws:iam::${var.aws_account_id}:role/github-actions"
  #   session_name = "GitHubActionsTerraform"
  # }
}

variable "aws_account_id" {
  description = "AWS Account ID"
  type        = string
}

resource "aws_sqs_queue" "my_queue" {
  name = "my-queue"
}

variable "font_name" {
  description = "generate pdf font name"
  type        = string
}

variable "font_path" {
  description = "generate pdf font path"
  type        = string
}

variable "r2_endpoint_url" {
  description = "r2 endpoint url"
  type        = string
}

variable "region_name" {
  description = "region name"
  type        = string
}

variable "bucket_name" {
  description = "bucket name"
  type        = string
}

variable "r2_access_key_id" {
  description = "r2 access key id"
  type        = string
}

variable "r2_secret_access_key" {
  description = "r2 secret access key"
  type        = string
}

# Lambda用のIAMロール作成
resource "aws_iam_role" "lambda_role" {
  name = "python_lambda_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Action = "sts:AssumeRole",
      Effect = "Allow",
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

# CloudWatch Logs用のポリシーアタッチ
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# SQSアクセス用のカスタムポリシー作成
resource "aws_iam_policy" "sqs_policy" {
  name        = "lambda_sqs_policy"
  description = "Allow Lambda to read from SQS queue"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect = "Allow",
      Action = [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ],
      Resource = aws_sqs_queue.my_queue.arn
    }]
  })
}

# カスタムポリシーをIAMロールにアタッチ
resource "aws_iam_role_policy_attachment" "lambda_sqs_attach" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.sqs_policy.arn
}

resource "aws_lambda_function" "python_lambda" {
  function_name = "python_handler"
  filename      = "./lambda/python_lambda.zip"
  handler       = "handler.lambda_handler"
  runtime       = "python3.10"
  role          = aws_iam_role.lambda_role.arn

  environment {
    variables = {
      ENV             = "production"
      FONT_NAME       = var.font_name
      FONT_PATH       = var.font_path
      R2_ENDPOINT_URL = var.r2_endpoint_url
      REGION_NAME     = var.region_name
      BUCKET_NAME     = var.bucket_name
      R2_ACCESS_KEY_ID    = var.r2_access_key_id
      R2_SECRET_ACCESS_KEY = var.r2_secret_access_key
    }
  }
}

resource "aws_lambda_event_source_mapping" "sqs_mapping" {
  event_source_arn = aws_sqs_queue.my_queue.arn
  function_name    = aws_lambda_function.python_lambda.arn
  batch_size       = 10
  enabled          = true
}
