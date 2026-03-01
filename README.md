# Incra - 請求書生成システム

請求書の作成、PDF生成、管理を行うためのフルスタックアプリケーションです。

## システム構成

このリポジトリは3つの主要コンポーネントで構成されています:

### 1. API Server (incra_api_server/)

Go製のバックエンドAPIサーバー。AWS Lambda上で動作します。

- **技術スタック**: Go 1.x, Echo Framework, AWS Lambda
- **アーキテクチャ**: クリーンアーキテクチャ
- **主な機能**:
  - 請求書データの管理
  - SQSとの連携
  - Slack通知機能
- **デプロイ**: AWS Lambda (Function URL)

詳細は [incra_api_server/README.md](incra_api_server/README.md) を参照してください。

### 2. PDF Generator (pdf_generate/)

Python製のPDF生成Lambda関数。請求書のPDF出力を担当します。

- **技術スタック**: Python 3.x, ReportLab
- **主な機能**:
  - 請求書PDFの生成
  - Cloudflare R2へのアップロード
- **デプロイ**: AWS Lambda

詳細は [pdf_generate/README.md](pdf_generate/README.md) を参照してください。

### 3. Web Frontend (incra-web/)

React Router製のフロントエンドアプリケーション。Cloudflare Workersで動作します。

- **技術スタック**: React 19, React Router v7, TailwindCSS v4, TypeScript
- **主な機能**:
  - 請求書の作成・閲覧UI
  - SSR (Server-Side Rendering) 対応
  - レスポンシブデザイン
- **デプロイ**: Cloudflare Workers

詳細は [incra-web/README.md](incra-web/README.md) を参照してください。

## クイックスタート

### 前提条件

- Go 1.x以上
- Python 3.x以上
- Node.js 18以上
- AWS CLI (AWS認証設定済み)
- Terraform
- Wrangler CLI (Cloudflare認証設定済み)

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

## デプロイ

各コンポーネントのデプロイ方法については、それぞれのREADMEを参照してください。

### インフラストラクチャ

Terraformを使用してAWSリソースを管理しています:

- `incra_api_server/terraform/` - APIサーバー用Lambda関数
- `pdf_generate/terraform/` - PDF生成用Lambda関数
- `tf/` - AWS OIDC認証設定 (GitHub Actions用)

## CI/CD

GitHub Actionsを使用した自動デプロイが設定されています:

- PRマージ時にTerraform planを実行
- 承認後、自動でTerraform applyを実行
- Cloudflare Workersは `wrangler deploy` で自動デプロイ

## 開発ガイド

開発者向けの詳細なガイドラインは [CLAUDE.md](CLAUDE.md) を参照してください。

## ライセンス

このプロジェクトはプライベートリポジトリです。

## サポート

問題が発生した場合は、GitHubのIssuesで報告してください。
