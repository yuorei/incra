resource "aws_sqs_queue" "this" {
  name = var.queue_name
}

resource "aws_lambda_event_source_mapping" "this" {
  count = var.enable_event_source_mapping ? 1 : 0

  event_source_arn = aws_sqs_queue.this.arn
  function_name    = var.consumer_lambda_arn
  batch_size       = var.batch_size
  enabled          = true
}
