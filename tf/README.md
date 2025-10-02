# FILE:/terraform-aws-oidc/terraform-aws-oidc/README.md
# GitHub Actions を使用した Terraform AWS OIDC ログイン

このプロジェクトでは、GitHub Actions で OpenID Connect (OIDC) を使用して AWS ログインを設定します。これにより、GitHub Actions が AWS で認証してロールを引き受け、AWS リソースへの安全なアクセスが可能になります。

## プロジェクト構造

- `iam.tf`: GitHub Actions で OIDC を使用して AWS ログインを有効にするための IAM ロールとポリシー構成が含まれています。
- `main.tf`: Terraform 構成のメイン エントリ ポイントで、プロバイダーを初期化し、必要なリソースを含めます。
- `variables.tf`: GitHub リポジトリやロール名など、Terraform 構成の入力変数を定義します。
- `outputs.tf`: 構成を適用した後に Terraform によって返される出力値を指定します。
- `README.md`: Terraform 構成の設定と使用に関するドキュメント。

## 前提条件

- ローカルマシンに Terraform がインストールされている。

- IAM ロールとポリシーを作成するための適切な権限を持つ AWS アカウント。

- GitHub Actions が使用される GitHub リポジトリ。

## セットアップ手順

1. **リポジトリのクローンを作成する**
```bash
git clone <repository-url>
cd terraform-aws-oidc
```

2. **変数を構成する**
GitHub リポジトリの詳細とその他の必要なパラメータを使用して `variables.tf` ファイルを更新します。

3. **Terraform を初期化する**
Terraform 構成を初期化するには、次のコマンドを実行します:
```bash
terraform init
```

4. **デプロイメントを計画する**
実行プランを生成して、作成されるリソースを確認します:
```bash
terraform plan
```

5. **構成を適用する**
Terraform 構成を適用してリソースを作成します:
```bash
terraform apply
```

6. **セットアップを確認する**
適用後、IAM ロールが作成され、GitHub Actions 用に正しく構成されていることを確認します。

## 使用方法

セットアップが完了したら、GitHub Actions ワークフローで構成された IAM ロールを使用して、AWS で認証できます。ワークフローでロール ARN を参照するようにしてください。

## ライセンス

このプロジェクトは、MIT ライセンスに基づいてライセンスされています。詳細については、LICENSE ファイルを参照してください。