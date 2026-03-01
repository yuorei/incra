resource "aws_iam_role" "scheduler" {
  name = "${var.schedule_name}-scheduler-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "scheduler.amazonaws.com"
        },
      },
    ],
  })
}

resource "aws_iam_policy" "scheduler_invoke" {
  name = "${var.schedule_name}-invoke-lambda"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = "lambda:InvokeFunction",
        Resource = var.target_lambda_arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "scheduler_invoke" {
  role       = aws_iam_role.scheduler.name
  policy_arn = aws_iam_policy.scheduler_invoke.arn
}

resource "aws_scheduler_schedule" "this" {
  name = var.schedule_name

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = var.schedule_expression

  target {
    arn      = var.target_lambda_arn
    role_arn = aws_iam_role.scheduler.arn
  }
}
