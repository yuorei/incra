output "api_gateway_url" {
  description = "API Gateway endpoint URL"
  value       = module.api_gateway.api_url
}

output "api_server_function_name" {
  description = "API Server Lambda function name"
  value       = module.api_server_lambda.function_name
}

output "reminder_function_name" {
  description = "Reminder Lambda function name"
  value       = module.reminder_lambda.function_name
}

output "pdf_generate_function_name" {
  description = "PDF Generate Lambda function name"
  value       = module.pdf_generate_lambda.function_name
}

output "pdf_queue_url" {
  description = "PDF generation SQS queue URL"
  value       = module.pdf_queue.queue_url
}
