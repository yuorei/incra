# CLAUDE.md

自動で更新をして下さい。
- README.md
- CLAUDE.md
- .github/copilot-instructions.md

実装する時はそれぞれの専門家を呼び出して下さい。そして並列で効率的に作業を進めて下さい。

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

このリポジトリは請求書管理システムで、以下の主要コンポーネントで構成されています:

1. **incra_api_server/** - Go製のAPIサーバー（Lambda関数として動作）
   - クリーンアーキテクチャに基づく構造:
     - `src/domain/` - ビジネスロジック（Invoice, Client）とリポジトリインターフェース
     - `src/usecase/` - ユースケース層（InvoiceUseCase, ClientUseCase）
     - `src/ui/` - HTTPハンドラー（Echo framework）、Slackハンドラー、認証ミドルウェア
     - `src/infrastructure/` - 外部サービス統合（DynamoDB, SQS, Slack DM通知）
   - `api/v1/generated.go` - OpenAPI仕様（`petstore.yaml`）から自動生成
   - AWS Lambda上でEcho serverを実行（aws-lambda-go-api-proxy使用）
   - DynamoDB: `incra-invoices`, `incra-clients`, `incra-counter` テーブル使用

2. **incra_api_server/cmd/reminder/** - Go製のリマインダーLambda関数
   - EventBridge Schedulerから毎日9:00 JSTに起動
   - 支払い期限が近い/超過した請求書のSlack DM通知

3. **pdf_generate/** - Python製のPDF生成Lambda関数
   - `src/handler.py` がLambdaエントリーポイント
   - SQSからフルインボイスデータを受信してPDF生成
   - `src/invoice_generator.py` で実際のPDF生成処理
   - R2（Cloudflare Object Storage）へのアップロード後、DynamoDBのpdf_urlを更新

4. **incra-web/** - React Router製のフロントエンドアプリケーション（Cloudflare Workersで動作）
   - React Router v7を使用したSSR対応のWebアプリケーション
   - `app/routes/` - ファイルシステムベースルーティング
     - 請求書管理: `invoices._index`, `invoices.new`, `invoices.$invoiceId`, `invoices.$invoiceId.edit`
     - 取引先管理: `clients._index`, `clients.new`, `clients.$clientId`
   - `app/lib/api.ts` - APIフェッチヘルパー（X-Slack-User-Idヘッダー付与）
   - `workers/app.ts` - Cloudflare Workers エントリーポイント
   - TailwindCSS v4でスタイリング
   - Viteでビルド、Wranglerでデプロイ

5. **infra/** - Terraform構成（モジュール化・環境分離）
   - `modules/` - 再利用可能なTerraformモジュール（lambda, api_gateway, dynamodb, sqs, eventbridge, iam）
   - `environments/prod/` - 本番環境のTerraform設定（全モジュール呼び出し）
   - `global/oidc/` - GitHub Actions OIDC認証（AWS IAM設定）

## API Endpoints

### 請求書（Invoices）
- `POST /invoices` - 請求書作成（draft状態）
- `GET /invoices` - 一覧取得（?status=&limit=&last_key=）
- `GET /invoices/{invoice_id}` - 詳細取得
- `PUT /invoices/{invoice_id}` - 更新（draftのみ）
- `PATCH /invoices/{invoice_id}/status` - ステータス遷移
- `DELETE /invoices/{invoice_id}` - 削除（draftのみ）

ステータス遷移: draft→sent（PDF生成SQS送信 + 取引先Slack DM通知）、sent→paid/cancelled、draft→cancelled

### 取引先（Clients）
- `POST /clients` - 取引先登録
- `GET /clients` - 一覧取得
- `GET /clients/{client_id}` - 詳細取得
- `PUT /clients/{client_id}` - 更新
- `DELETE /clients/{client_id}` - 削除

### Slack
- `POST /slack/events` - Slackイベント受信
- `POST /slack/slashs` - Slackスラッシュコマンド（請求書作成モーダル → 即sent遷移）
- `POST /slack/interactions` - Slackモーダル送信処理
- `GET /slack/users` - Slackワークスペースユーザー一覧取得

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

# Lambda用バイナリをビルド（bootstrap.zip + reminder.zipを生成）
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

### Terraform (infra/)

```bash
# 本番環境
cd infra/environments/prod
terraform init
terraform fmt -check -recursive ../../
terraform validate
terraform plan

# OIDC設定
cd infra/global/oidc
terraform init
terraform plan
```

**注意:** Terraform stateファイルはPRに含めない。シークレット値は `-var` フラグまたは環境変数で渡す。`terraform.tfvars` には非シークレット値のみ記載。

## DynamoDB Tables

- **incra-invoices** - 請求書テーブル（PK: invoice_id, GSI: issuer_slack_user_id-created_at-index）
- **incra-clients** - 取引先テーブル（PK: client_id, GSI: slack_user_id-index）
- **incra-counter** - 採番テーブル（PK: counter_name, アトミックインクリメントでINV-YYYY-NNNNフォーマット）

## Code Generation & Generated Files

- `incra_api_server/api/v1/generated.go` は`petstore.yaml`から自動生成されるため、直接編集しない
- APIスキーマを変更する場合は`petstore.yaml`を編集後、`make gen`を実行
- 生成ファイルは`api/`配下に隔離されている

## Testing

- Goテスト: `*_test.go`ファイルをソースと同じディレクトリに配置
- テーブル駆動テストを推奨（特にusecase層）
- Pythonテスト: `pdf_generate/tests/`配下に配置（必要に応じて作成）し、`python -m unittest`で実行
- Web型チェック: `cd incra-web && npm run typecheck`

## Build Tags & Lambda Deployment

- Go Lambda関数は`lambda.norpc`ビルドタグが必須
- `build.sh`スクリプトがAPI ServerとReminder Lambdaのビルドとzip化を自動実行
- 出力先は各サービスの`terraform/lambda/`ディレクトリ
- CI/CDでも同じビルド設定を使用すること

## Secrets Management

- 環境変数: `SLACK_TOKEN`, `QUEUE_URL`, `INVOICE_TABLE_NAME`, `CLIENT_TABLE_NAME`, `COUNTER_TABLE_NAME`, `WEB_BASE_URL`など
- AWS Systems Manager Parameter Storeまたは Secrets Manager を使用
- `.env`ファイルをコミットしない
- GitHub Actionsでは OIDC経由でAWS認証（IAM roleをassume）

## CI/CD Workflows

- `.github/workflows/incra_api_server_plan.yaml` - PRで `infra/` 変更時にTerraform plan実行
- `.github/workflows/pdf_generate_plan.yaml` - PRで `infra/` 変更時にTerraform plan実行（Python Lambda含む）
- `.github/workflows/pdf_generate_apply.yaml` - `/apply` コメントでTerraform apply実行（手動トリガー）
- `.github/workflows/deploy.yaml` - **mainブランチへのpush時に自動デプロイ**
  - トリガー: `incra_api_server/`, `pdf_generate/`, `infra/` のいずれかが変更された場合
  - Go Lambda（API Server + Reminder）とPython Lambda（PDF Generator）をビルド
  - AWS OIDC認証 → Terraform apply -auto-approve
  - Terraform stateはS3バックエンド（`incra-terraform-state`）+ DynamoDBロック（`incra-terraform-locks`）
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
