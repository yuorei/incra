variable "schedule_name" {
  description = "Name of the EventBridge schedule"
  type        = string
}

variable "schedule_expression" {
  description = "Schedule expression (e.g. cron or rate)"
  type        = string
}

variable "target_lambda_arn" {
  description = "ARN of the target Lambda function"
  type        = string
}
