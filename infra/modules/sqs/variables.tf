variable "queue_name" {
  description = "Name of the SQS queue"
  type        = string
}

variable "consumer_lambda_arn" {
  description = "ARN of the Lambda function to consume messages"
  type        = string
  default     = ""
}

variable "enable_event_source_mapping" {
  description = "Whether to create the SQS event source mapping"
  type        = bool
  default     = false
}

variable "batch_size" {
  description = "SQS event source mapping batch size"
  type        = number
  default     = 10
}
