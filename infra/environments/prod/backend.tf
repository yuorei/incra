# S3バックエンド設定
# バケット作成後に main.tf の backend ブロックのコメントを解除してください。
#
# バケット作成コマンド:
#   aws s3api create-bucket \
#     --bucket incra-terraform-state \
#     --region ap-northeast-1 \
#     --create-bucket-configuration LocationConstraint=ap-northeast-1
#
#   aws s3api put-bucket-versioning \
#     --bucket incra-terraform-state \
#     --versioning-configuration Status=Enabled
