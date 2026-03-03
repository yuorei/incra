# 共通バージョン定義（参照用）
# 各環境の main.tf にて terraform ブロック内で指定してください
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.90"
    }
  }
}
