# Incra - 請求書管理システム

請求書の作成、発行、ステータス管理、PDF生成、取引先管理を行うフルスタックアプリケーションです。

## システム構成

### 1. API Server (incra_api_server/)

Go製のバックエンドAPIサーバー。AWS Lambda上で動作します。

- **技術スタック**: Go 1.23, Echo Framework, AWS Lambda, DynamoDB
- **アーキテクチャ**: クリーンアーキテクチャ（domain / usecase / ui / infrastructure）
- **主な機能**:
  - 請求書CRUD（作成・一覧・詳細・更新・削除）
  - ステータス管理（draft → sent → paid / cancelled）
  - 取引先マスタ管理
  - 請求書番号の自動採番（INV-YYYY-NNNN）
  - SQS経由のPDF生成トリガー
  - Slackスラッシュコマンド対応
- **デプロイ**: AWS Lambda + API Gateway

### 2. Reminder Lambda (incra_api_server/cmd/reminder/)

Go製のリマインダーLambda関数。EventBridge Schedulerから毎日9:00 JSTに起動します。

- **主な機能**:
  - 支払い期限が3日以内または超過した請求書を検知
  - 発行者のSlackへDM通知

### 3. PDF Generator (pdf_generate/)

Python製のPDF生成Lambda関数。請求書のPDF出力を担当します。

- **技術スタック**: Python 3.10, ReportLab
- **主な機能**:
  - SQSからフルインボイスデータを受信してPDF生成
  - Cloudflare R2へのアップロード
  - DynamoDBへのPDF URL更新
- **デプロイ**: AWS Lambda

### 4. Web Frontend (incra-web/)

React Router製のフロントエンドアプリケーション。Cloudflare Workersで動作します。

- **技術スタック**: React 19, React Router v7, TailwindCSS v4, TypeScript
- **主な機能**:
  - 請求書の作成・一覧・詳細・編集・ステータス遷移UI
  - 取引先の登録・一覧・詳細・編集・削除UI
  - Slack OAuthログイン
  - SSR (Server-Side Rendering) 対応
- **デプロイ**: Cloudflare Workers

### 5. Infrastructure (infra/)

Terraform構成をモジュール化・環境分離して一元管理しています。

- `infra/modules/` - 再利用可能なTerraformモジュール（lambda, api_gateway, dynamodb, sqs, eventbridge, iam）
- `infra/environments/prod/` - 本番環境のTerraform設定（全モジュール呼び出し）
- `infra/global/oidc/` - GitHub Actions OIDC認証（AWS IAM設定）

## クイックスタート

### 前提条件

- Go 1.23以上
- Python 3.10以上
- Node.js 18以上
- AWS CLI（AWS認証設定済み）
- Terraform
- Wrangler CLI（Cloudflare認証設定済み）

### 開発環境のセットアップ

```bash
# API Server
cd incra_api_server
make dev

# PDF Generator
cd pdf_generate
pip install -r src/requirements.txt

# Web Frontend
cd incra-web
npm install
npm run dev
```

## API エンドポイント

### 請求書
| メソッド | パス | 説明 |
|---------|------|------|
| POST | /invoices | 請求書作成（draft） |
| GET | /invoices | 一覧取得（?status=&limit=&last_key=） |
| GET | /invoices/{id} | 詳細取得 |
| PUT | /invoices/{id} | 更新（draftのみ） |
| PATCH | /invoices/{id}/status | ステータス遷移 |
| DELETE | /invoices/{id} | 削除（draftのみ） |

### 取引先
| メソッド | パス | 説明 |
|---------|------|------|
| POST | /clients | 取引先登録 |
| GET | /clients | 一覧取得 |
| GET | /clients/{id} | 詳細取得 |
| PUT | /clients/{id} | 更新 |
| DELETE | /clients/{id} | 削除 |

## E2Eフロー

1. 取引先登録（Web UI or API）
2. 請求書作成（draft状態）
3. draft → sent ステータス遷移（PDF生成SQS送信）
4. PDF生成Lambda実行 → R2アップロード → DynamoDB pdf_url更新
5. sent → paid ステータス遷移
6. Web UIで一覧・詳細・PDF確認
7. リマインダーLambdaが期限間近の請求書をSlack通知

## デプロイ

各コンポーネントのデプロイ方法については、それぞれのディレクトリ内のドキュメントを参照してください。

## CI/CD

GitHub Actionsを使用したCI/CDパイプラインが設定されています:

### PR時（CI）
- `infra/` 変更時にTerraform planを自動実行し、PRコメントに結果を表示
- `/apply` コメントで手動Terraform apply（レビュー後の即時適用）

### mainマージ時（CD）
- `incra_api_server/`, `pdf_generate/`, `infra/` の変更を検知して自動デプロイ
- Go Lambda（API Server + Reminder）とPython Lambda（PDF Generator）をビルド
- OIDC認証でAWSにアクセスし、Terraform apply -auto-approveを実行
- Terraform stateはS3（`incra-terraform-state`）で管理、DynamoDB（`incra-terraform-locks`）でロック

### Webフロントエンド
- Cloudflare Workersは `wrangler deploy` でデプロイ

## 開発ガイド

開発者向けの詳細なガイドラインは [CLAUDE.md](CLAUDE.md) を参照してください。

## ライセンス

このプロジェクトはプライベートリポジトリです。

## サポート

問題が発生した場合は、GitHubのIssuesで報告してください。
