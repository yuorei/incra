variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

# --- DynamoDB table names ---

variable "invoice_table_name" {
  description = "DynamoDB table name for invoices"
  type        = string
  default     = "incra-invoices"
}

variable "counter_table_name" {
  description = "DynamoDB table name for counter"
  type        = string
  default     = "incra-counter"
}

# --- Slack ---

variable "slack_token" {
  description = "Slack Bot Token (xoxb-...)"
  type        = string
  sensitive   = true
}

variable "slack_client_id" {
  description = "Slack App Client ID (OAuth)"
  type        = string
}

variable "slack_client_secret" {
  description = "Slack App Client Secret (OAuth)"
  type        = string
  sensitive   = true
}

variable "session_secret" {
  description = "Session encryption secret key"
  type        = string
  sensitive   = true
}

variable "web_base_url" {
  description = "Base URL of the web application"
  type        = string
  default     = ""
}

# --- PDF Generate Lambda secrets ---

variable "font_name" {
  description = "PDF generation font name"
  type        = string
}

variable "font_path" {
  description = "PDF generation font path"
  type        = string
}

