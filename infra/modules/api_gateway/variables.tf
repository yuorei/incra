variable "api_name" {
  description = "Name of the API Gateway REST API"
  type        = string
}

variable "description" {
  description = "Description of the API Gateway"
  type        = string
  default     = ""
}

variable "stage_name" {
  description = "Deployment stage name"
  type        = string
}

variable "lambda_invoke_arn" {
  description = "Invoke ARN of the backend Lambda function"
  type        = string
}

variable "lambda_function_name" {
  description = "Name of the backend Lambda function"
  type        = string
}
