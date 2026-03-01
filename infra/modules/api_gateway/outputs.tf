output "api_url" {
  description = "URL of the deployed API"
  value       = aws_api_gateway_stage.this.invoke_url
}

output "execution_arn" {
  description = "Execution ARN of the REST API"
  value       = aws_api_gateway_rest_api.this.execution_arn
}

output "api_id" {
  description = "ID of the REST API"
  value       = aws_api_gateway_rest_api.this.id
}
