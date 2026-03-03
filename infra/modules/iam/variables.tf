variable "role_name" {
  description = "Name of the IAM role"
  type        = string
}

variable "assume_role_service" {
  description = "AWS service principal for the assume role policy"
  type        = string
  default     = "lambda.amazonaws.com"
}

variable "policy_name" {
  description = "Name of the IAM policy"
  type        = string
}

variable "policy_description" {
  description = "Description of the IAM policy"
  type        = string
  default     = ""
}

variable "policy_statements" {
  description = "List of IAM policy statements"
  type = list(object({
    effect    = string
    actions   = list(string)
    resources = list(string)
  }))
}
