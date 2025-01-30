provider "aws" {
  region     = "ap-northeast-1"
}

# CloudWatch Logsへのアクセス権限
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# CloudWatch Logsのロググループ
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/incra-api-server-prod"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/incra-api-server-prod"
  retention_in_days = 14
}

# Lambda関数
resource "aws_lambda_function" "incra-api-server-prod" {
  filename         = "./lambda/bootstrap.zip"  # Goアプリケーションのビルド済みバイナリ
  function_name    = "api-server"
  role            = aws_iam_role.lambda_exec.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"

  environment {
    variables = {
      STAGE = "prod"
    }
  }
}

## Lambdaロール
resource "aws_iam_role" "lambda_exec" {
  name = "lambda_exec_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com"
        },
      },
    ],
  })
}

## APIGateway
resource "aws_api_gateway_rest_api" "incra-api-server-prod" {
  name        = "incra-api-server-prod"
  description = "golang echo api server prod"
}

resource "aws_api_gateway_resource" "incra-api-server-prod" {
  rest_api_id = aws_api_gateway_rest_api.incra-api-server-prod.id
  parent_id   = aws_api_gateway_rest_api.incra-api-server-prod.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "incra-api-server-prod" {
  rest_api_id   = aws_api_gateway_rest_api.incra-api-server-prod.id
  resource_id   = aws_api_gateway_resource.incra-api-server-prod.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "incra-api-server-prod" {
  rest_api_id = aws_api_gateway_rest_api.incra-api-server-prod.id
  resource_id = aws_api_gateway_resource.incra-api-server-prod.id
  http_method = aws_api_gateway_method.incra-api-server-prod.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.incra-api-server-prod.invoke_arn
}

resource "aws_api_gateway_deployment" "incra-api-server-prod" {
  depends_on = [aws_api_gateway_integration.incra-api-server-prod]

  rest_api_id = aws_api_gateway_rest_api.incra-api-server-prod.id
  stage_name  = "api"
}

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.incra-api-server-prod.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.incra-api-server-prod.execution_arn}/*/*/*"
}

variable "region" {
  description = "The AWS region for the resources."
  type        = string
  default     = "ap-northeast-1"  # Default region value, you can change this if needed
}

# API GatewayのエンドポイントURLを出力
output "api_gateway_url" {
  value = "https://${aws_api_gateway_rest_api.incra-api-server-prod.id}.execute-api.${var.region}.amazonaws.com/api"
}
