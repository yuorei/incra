terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.90"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  # S3バックエンド（バケット作成後にコメント解除）
  # backend "s3" {
  #   bucket  = "incra-terraform-state"
  #   key     = "global/oidc/terraform.tfstate"
  #   region  = "ap-northeast-1"
  #   encrypt = true
  # }
}

provider "aws" {
  region = var.region
}

data "tls_certificate" "github_actions" {
  url = "https://token.actions.githubusercontent.com/.well-known/openid-configuration"
}

resource "aws_iam_openid_connect_provider" "github_actions" {
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.github_actions.certificates[0].sha1_fingerprint]
}

resource "aws_iam_role" "github_actions" {
  name               = "github-actions"
  assume_role_policy = data.aws_iam_policy_document.github_actions.json
  description        = "IAM Role for GitHub Actions OIDC"
}

locals {
  allowed_github_repositories = [
    "incra",
  ]
  github_org_name = "yuorei"
  full_paths = [
    for repo in local.allowed_github_repositories : "repo:${local.github_org_name}/${repo}:*"
  ]
}

data "aws_iam_policy_document" "github_actions" {
  statement {
    actions = [
      "sts:AssumeRoleWithWebIdentity",
    ]

    principals {
      type = "Federated"
      identifiers = [
        aws_iam_openid_connect_provider.github_actions.arn
      ]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = local.full_paths
    }
  }
}

resource "aws_iam_role_policy_attachment" "admin" {
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
  role       = aws_iam_role.github_actions.name
}
