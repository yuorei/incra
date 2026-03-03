日本語でお願いします

# Copilot Instructions

## アーキテクチャ

請求書管理システム（incra）で、以下のコンポーネントで構成:

1. **incra_api_server/** - Go APIサーバー（Lambda）、クリーンアーキテクチャ
   - `src/domain/` - ドメインモデル（Invoice）とリポジトリインターフェース
   - `src/usecase/` - ユースケース層（InvoiceUseCase）
   - `src/ui/` - HTTPハンドラー（Echo framework）、Slackハンドラー、認証ミドルウェア
   - `src/infrastructure/` - DynamoDB, SQS, Slack DM通知実装
   - `api/v1/generated.go` - `petstore.yaml`から自動生成（直接編集禁止）
   - `cmd/reminder/` - リマインダーLambda（毎日Slack通知）
   - Slack連携: ユーザー一覧API（`GET /slack/users`）、Block Kitボタン付きDM通知、モーダルで即sent遷移、ボタンアクション処理（支払い報告・確認・差し戻し）
   - 二段階支払い確認フロー: draft→sent→paid→confirmed（権限ベースのステータス遷移バリデーション）
   - 請求先はSlackユーザーIDベースで管理（`billing_slack_user_id`フィールド）
   - 発行済み・受領済み請求書の一覧取得対応（`type=issued|received`クエリパラメータ）

2. **pdf_generate/** - Python PDF生成Lambda
   - SQSからインボイスデータ受信 → PDF生成 → Slack DMで請求先ユーザーへPDFファイル送信

3. **incra-web/** - React Router v7フロントエンド（Cloudflare Workers）
   - `app/routes/home` - パブリックランディングページ（`/`、認証不要）
   - `app/routes/logout` - ログアウト（セッションクリア）
   - `app/routes/invoices.*` - 請求書管理ページ群（発行済み・受領済みタブ切替）
   - `app/components/auth-header.tsx` - 共通認証ヘッダー（ユーザー情報・ログアウト）
   - `app/lib/api.ts` - 認証ヘッダー付きAPIフェッチヘルパー
   - ログイン後のリダイレクト先は`/invoices`（`/`ではなく）

4. **infra/** - Terraform構成（モジュール化・環境分離）
   - `modules/` - 再利用可能なTerraformモジュール（lambda, api_gateway, dynamodb, sqs, eventbridge, iam）
   - `environments/prod/` - 本番環境のTerraform設定（全モジュール呼び出し）
   - `global/oidc/` - GitHub Actions OIDC認証（AWS IAM設定）

## 主要コマンド

```bash
cd incra_api_server && make gen       # petstore.yamlから再生成
cd incra_api_server && go test ./...  # テスト
cd incra_api_server && ./build.sh     # API + reminder Lambdaビルド
cd incra-web && npm run typecheck     # 型チェック
cd incra-web && npm run dev           # 開発サーバー

# Terraform（本番環境）
cd infra/environments/prod && terraform init && terraform plan

# Terraform（OIDC設定）
cd infra/global/oidc && terraform init && terraform plan
```

## DynamoDBテーブル

- `incra-invoices` (PK: invoice_id, GSI: issuer_slack_user_id-created_at-index, GSI: billing_slack_user_id-created_at-index)
- `incra-counter` (PK: counter_name, アトミック採番 INV-YYYY-NNNN)

## CI/CD

- `.github/workflows/deploy.yaml` - mainへのpush時にAWS自動デプロイ
  - `incra_api_server/`, `pdf_generate/`, `infra/` の変更で発火
  - Go/Python Lambdaビルド → OIDC認証 → Terraform apply
  - Terraform state: S3（`incra-terraform-state`）+ DynamoDBロック（`incra-terraform-locks`）
- PR時: 全Lambda（Go + Python）ビルド → Terraform plan自動実行 + PRコメント

## コーディング規約

- Go: gofmt準拠、PascalCaseエクスポート、lambda.norpcビルドタグ必須
- Python: PEP 8、snake_case
- TypeScript: 2スペース、関数コンポーネント、TailwindCSS
- Terraform: 明示的リソース名
