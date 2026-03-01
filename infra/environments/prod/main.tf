terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.90"
    }
  }

  backend "s3" {
    bucket         = "incra-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "ap-northeast-1"
    encrypt        = true
    dynamodb_table = "incra-terraform-locks"
  }
}

provider "aws" {
  region = var.region
}

# ---------------------------------------------------------------------------
# 1. IAM Roles
# ---------------------------------------------------------------------------

module "api_server_iam" {
  source = "../../modules/iam"

  role_name          = "lambda_exec_role"
  policy_name        = "api-server-lambda-policy"
  policy_description = "API Server Lambda: SQS send + DynamoDB full access"

  policy_statements = [
    {
      effect  = "Allow"
      actions = ["kms:Decrypt"]
      resources = [
        "arn:aws:kms:${var.region}:438037648687:key/f5f98cb3-d309-4acc-9cd0-1258177f781f"
      ]
    },
    {
      effect  = "Allow"
      actions = ["sqs:SendMessage"]
      resources = [
        module.pdf_queue.queue_arn
      ]
    },
    {
      effect = "Allow"
      actions = [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem",
        "dynamodb:Query",
        "dynamodb:Scan",
        "dynamodb:BatchGetItem",
        "dynamodb:BatchWriteItem"
      ]
      resources = [
        module.invoices_table.table_arn,
        "${module.invoices_table.table_arn}/index/*",
        module.clients_table.table_arn,
        "${module.clients_table.table_arn}/index/*",
        module.counter_table.table_arn
      ]
    }
  ]
}

module "reminder_iam" {
  source = "../../modules/iam"

  role_name          = "reminder_lambda_exec_role"
  policy_name        = "reminder-lambda-policy"
  policy_description = "Reminder Lambda: DynamoDB read access on invoices"

  policy_statements = [
    {
      effect = "Allow"
      actions = [
        "dynamodb:GetItem",
        "dynamodb:Query",
        "dynamodb:Scan"
      ]
      resources = [
        module.invoices_table.table_arn,
        "${module.invoices_table.table_arn}/index/*"
      ]
    }
  ]
}

module "pdf_generate_iam" {
  source = "../../modules/iam"

  role_name          = "python_lambda_role"
  policy_name        = "pdf-generate-lambda-policy"
  policy_description = "PDF Generate Lambda: SQS consume + DynamoDB write on invoices"

  policy_statements = [
    {
      effect = "Allow"
      actions = [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ]
      resources = [
        module.pdf_queue.queue_arn
      ]
    },
    {
      effect = "Allow"
      actions = [
        "dynamodb:UpdateItem",
        "dynamodb:GetItem"
      ]
      resources = [
        "arn:aws:dynamodb:${var.region}:*:table/${var.invoice_table_name}"
      ]
    }
  ]
}

# ---------------------------------------------------------------------------
# 2. DynamoDB Tables
# ---------------------------------------------------------------------------

module "invoices_table" {
  source = "../../modules/dynamodb"

  table_name = var.invoice_table_name
  hash_key   = "invoice_id"

  attributes = [
    { name = "invoice_id", type = "S" },
    { name = "issuer_slack_user_id", type = "S" },
    { name = "created_at", type = "S" }
  ]

  global_secondary_indexes = [
    {
      name            = "issuer_slack_user_id-created_at-index"
      hash_key        = "issuer_slack_user_id"
      range_key       = "created_at"
      projection_type = "ALL"
    }
  ]
}

module "clients_table" {
  source = "../../modules/dynamodb"

  table_name = var.client_table_name
  hash_key   = "client_id"

  attributes = [
    { name = "client_id", type = "S" },
    { name = "slack_user_id", type = "S" }
  ]

  global_secondary_indexes = [
    {
      name            = "slack_user_id-index"
      hash_key        = "slack_user_id"
      projection_type = "ALL"
    }
  ]
}

module "counter_table" {
  source = "../../modules/dynamodb"

  table_name = var.counter_table_name
  hash_key   = "counter_name"

  attributes = [
    { name = "counter_name", type = "S" }
  ]
}

# ---------------------------------------------------------------------------
# 3. SQS
# ---------------------------------------------------------------------------

module "pdf_queue" {
  source = "../../modules/sqs"

  queue_name                  = "incra-pdf-generate-queue"
  enable_event_source_mapping = true
  consumer_lambda_arn         = module.pdf_generate_lambda.function_arn
  batch_size                  = 10
}

# ---------------------------------------------------------------------------
# 4. Lambda Functions
# ---------------------------------------------------------------------------

module "api_server_lambda" {
  source = "../../modules/lambda"

  function_name = "api-server"
  filename      = "./lambda/bootstrap.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  role_arn      = module.api_server_iam.role_arn

  environment_variables = {
    STAGE               = "prod"
    QUEUE_URL           = module.pdf_queue.queue_url
    SLACK_TOKEN         = var.slack_token
    SLACK_CLIENT_ID     = var.slack_client_id
    SLACK_CLIENT_SECRET = var.slack_client_secret
    SESSION_SECRET      = var.session_secret
    DYNAMODB_REGION     = var.region
    INVOICE_TABLE_NAME  = var.invoice_table_name
    CLIENT_TABLE_NAME   = var.client_table_name
    COUNTER_TABLE_NAME  = var.counter_table_name
    WEB_BASE_URL        = var.web_base_url
  }
}

module "reminder_lambda" {
  source = "../../modules/lambda"

  function_name = "incra-reminder-prod"
  filename      = "./lambda/reminder.zip"
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  role_arn      = module.reminder_iam.role_arn

  environment_variables = {
    INVOICE_TABLE_NAME = var.invoice_table_name
    SLACK_TOKEN        = var.slack_token
    WEB_BASE_URL       = var.web_base_url
  }
}

module "pdf_generate_lambda" {
  source = "../../modules/lambda"

  function_name = "python_handler"
  filename      = "./lambda/python_lambda.zip"
  handler       = "handler.lambda_handler"
  runtime       = "python3.10"
  role_arn      = module.pdf_generate_iam.role_arn

  environment_variables = {
    ENV                = "production"
    FONT_NAME          = var.font_name
    FONT_PATH          = var.font_path
    INVOICE_TABLE_NAME = var.invoice_table_name
    SLACK_TOKEN        = var.slack_token
  }
}

# ---------------------------------------------------------------------------
# 5. API Gateway
# ---------------------------------------------------------------------------

module "api_gateway" {
  source = "../../modules/api_gateway"

  api_name             = "incra-api-server-prod"
  description          = "golang echo api server prod"
  stage_name           = "api"
  lambda_invoke_arn    = module.api_server_lambda.invoke_arn
  lambda_function_name = module.api_server_lambda.function_name
}

# API Gateway CloudWatch Log Group
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/incra-api-server-prod"
  retention_in_days = 14
}

# ---------------------------------------------------------------------------
# 6. EventBridge
# ---------------------------------------------------------------------------

module "reminder_schedule" {
  source = "../../modules/eventbridge"

  schedule_name       = "incra-reminder-daily"
  schedule_expression = "cron(0 0 * * ? *)"
  target_lambda_arn   = module.reminder_lambda.function_arn
}
