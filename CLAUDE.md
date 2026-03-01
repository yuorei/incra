# CLAUDE.md

自動で更新をして下さい。
- README.md
- CLAUDE.md
- .github/copilot-instructions.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

このリポジトリは請求書生成システムで、3つの主要コンポーネントで構成されています:

1. **incra_api_server/** - Go製のAPIサーバー（Lambda関数として動作）
   - クリーンアーキテクチャに基づく構造:
     - `src/domain/` - ビジネスロジックとリポジトリインターフェース
     - `src/usecase/` - ユースケース層（オーケストレーション）
     - `src/ui/` - HTTPハンドラー（Echo framework）
     - `src/infrastructure/` - 外部サービス統合（SQS、Slackなど）
   - `api/v1/generated.go` - OpenAPI仕様（`petstore.yaml`）から自動生成
   - AWS Lambda上でEcho serverを実行（aws-lambda-go-api-proxy使用）

2. **pdf_generate/** - Python製のPDF生成Lambda関数
   - `src/handler.py` がLambdaエントリーポイント
   - `src/invoice_generator.py` で実際のPDF生成処理
   - R2（Cloudflare Object Storage）へのアップロード機能

3. **incra-web/** - React Router製のフロントエンドアプリケーション（Cloudflare Workersで動作）
   - React Router v7を使用したSSR対応のWebアプリケーション
   - `app/routes/` - ファイルシステムベースルーティング
   - `workers/app.ts` - Cloudflare Workers エントリーポイント
   - TailwindCSS v4でスタイリング
   - Viteでビルド、Wranglerでデプロイ

4. **tf/** - AWS OIDC認証のTerraform設定（GitHub Actions用）
   - 各サービスのTerraformは各ディレクトリ内の`terraform/`配下に存在

## Development Commands

### Go API Server (incra_api_server/)

```bash
# ローカル開発環境起動（Docker Compose）
cd incra_api_server && make dev

# OpenAPI仕様からコード生成
cd incra_api_server && make gen

# コードフォーマット + go.mod整理
cd incra_api_server && make fmt

# テスト実行
cd incra_api_server && go test ./...

# Lambda用バイナリをビルド（terraform/lambda/bootstrap.zipを生成）
cd incra_api_server && ./build.sh
```

**重要:** `build.sh`は`GOOS=linux GOARCH=amd64`と`-tags lambda.norpc`でビルドします。これらの設定は必須です。

### PDF Generator (pdf_generate/)

```bash
# 依存関係インストール
pip install -r pdf_generate/src/requirements.txt

# ローカルでの動作確認
python pdf_generate/src/handler.py

# Lambda用zipパッケージ作成（terraform/lambda/python_lambda.zipを生成）
cd pdf_generate/src && ./build.sh
```

### Web Frontend (incra-web/)

```bash
# 依存関係インストール
cd incra-web && npm install

# 開発サーバー起動（http://localhost:5173）
cd incra-web && npm run dev

# 型チェック実行
cd incra-web && npm run typecheck

# プロダクションビルド
cd incra-web && npm run build

# ローカルでプロダクションビルドをプレビュー
cd incra-web && npm run preview

# Cloudflare Workersへデプロイ（ビルド+デプロイを自動実行）
cd incra-web && npm run deploy
```

**重要:** デプロイには Cloudflare アカウントと Wrangler CLI の認証設定が必要です。プレビューURLデプロイは`npx wrangler versions upload`を使用してください。

### Terraform

```bash
# API Server用
cd incra_api_server/terraform
terraform init
terraform fmt        # フォーマット
terraform validate   # 検証
terraform plan       # 実行計画

# PDF Generator用
cd pdf_generate/terraform
terraform init
terraform plan

# AWS OIDC設定用
cd tf
terraform init
terraform plan
```

**注意:** Terraform stateファイルはPRに含めない。変更が必要な場合はインフラ担当者と調整すること。

## Code Generation & Generated Files

- `incra_api_server/api/v1/generated.go` は`petstore.yaml`から自動生成されるため、直接編集しない
- APIスキーマを変更する場合は`petstore.yaml`を編集後、`make gen`を実行
- 生成ファイルは`api/`配下に隔離されている

## Testing

- Goテスト: `*_test.go`ファイルをソースと同じディレクトリに配置
- テーブル駆動テストを推奨（特にusecase層）
- Pythonテスト: `pdf_generate/tests/`配下に配置（必要に応じて作成）し、`python -m unittest`で実行

## Build Tags & Lambda Deployment

- Go Lambda関数は`lambda.norpc`ビルドタグが必須
- `build.sh`スクリプトがビルドとzip化を自動実行
- 出力先は各サービスの`terraform/lambda/`ディレクトリ
- CI/CDでも同じビルド設定を使用すること

## Secrets Management

- 環境変数: `SLACK_TOKEN`, `QUEUE_URL`など
- AWS Systems Manager Parameter Storeまたは Secrets Manager を使用
- `.env`ファイルをコミットしない
- GitHub Actionsでは OIDC経由でAWS認証（IAM roleをassume）

## CI/CD Workflows

- `.github/workflows/incra_api_server_plan.yaml` - API Serverの Terraform plan
- `.github/workflows/pdf_generate_plan.yaml` - PDF Generatorの Terraform plan
- PRマージ時に自動でTerraform apply実行
- Terraform fmt checkが含まれる（`terraform fmt -check -recursive`）

## Coding Conventions

- **Go**: gofmt準拠、タブインデント、エクスポートされた識別子はPascalCase
- **パッケージ名**: 短い小文字の名詞（例: `domain`, `usecase`）
- **Python**: PEP 8準拠、4スペースインデント、snake_case関数名
- **TypeScript/React**:
  - 2スペースインデント（prettier準拠）
  - React Hooks使用推奨、関数コンポーネント優先
  - コンポーネント名はPascalCase、ファイル名はkebab-case
  - TailwindCSSのユーティリティクラスを使用
- **Terraform**: 明示的なリソース名（例: `aws_lambda_function.invoice_api`）
