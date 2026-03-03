# Incra - 請求書管理システム

請求書の作成、発行、ステータス管理、PDF生成を行うフルスタックアプリケーションです。SlackユーザーIDベースで請求先を管理し、発行済み・受領済み請求書の両方を閲覧できます。

## システム構成

### 1. API Server (incra_api_server/)

Go製のバックエンドAPIサーバー。AWS Lambda上で動作します。

- **技術スタック**: Go 1.23, Echo Framework, AWS Lambda, DynamoDB
- **アーキテクチャ**: クリーンアーキテクチャ（domain / usecase / ui / infrastructure）
- **主な機能**:
  - 請求書CRUD（作成・一覧・詳細・更新・削除）
  - 発行済み請求書と受領済み請求書の一覧取得（`type=issued|received`）
  - ステータス管理（draft → sent → paid → confirmed / cancelled）
  - 二段階支払い確認フロー（受取人が支払い報告 → 発行者が確認/差し戻し）
  - 権限ベースのステータス遷移バリデーション
  - 請求書番号の自動採番（INV-YYYY-NNNN）
  - SQS経由のPDF生成トリガー
  - sent遷移時の請求先Slack DM通知（`billing_slack_user_id`ベース）
  - Slackスラッシュコマンド対応（即sent遷移）
  - Slackワークスペースユーザー一覧API
- **デプロイ**: AWS Lambda + API Gateway

### 2. Reminder Lambda (incra_api_server/cmd/reminder/)

Go製のリマインダーLambda関数。EventBridge Schedulerから毎日9:00 JSTに起動します。

- **主な機能**:
  - 支払い期限が3日以内または超過した請求書を検知
  - 発行者のSlackへDM通知

### 3. PDF Generator (pdf_generate/)

Python製のPDF生成Lambda関数。請求書のPDF出力を担当します。

- **技術スタック**: Python 3.10, ReportLab, slack_sdk
- **主な機能**:
  - SQSからフルインボイスデータを受信してPDF生成
  - Slack DMで請求先ユーザーへPDFファイル送信
- **デプロイ**: AWS Lambda

### 4. Web Frontend (incra-web/)

React Router製のフロントエンドアプリケーション。Cloudflare Workersで動作します。

- **技術スタック**: React 19, React Router v7, TailwindCSS v4, TypeScript
- **主な機能**:
  - パブリックランディングページ（`/`）- アプリ紹介・特徴・使い方
  - 請求書の作成・一覧・詳細・編集・ステータス遷移UI
  - 発行済み・受領済みタブ切替による請求書一覧表示
  - Slack OAuthログイン（認証後は`/invoices`へリダイレクト）
  - 共通認証ヘッダー（ユーザー情報・ログアウト機能）
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
| GET | /invoices | 一覧取得（?status=&limit=&last_key=&type=issued\|received） |
| GET | /invoices/{id} | 詳細取得 |
| PUT | /invoices/{id} | 更新（draftのみ） |
| PATCH | /invoices/{id}/status | ステータス遷移（sent時に請求先DM通知） |
| DELETE | /invoices/{id} | 削除（draftのみ） |

### Slack
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /slack/users | ワークスペースユーザー一覧 |
| POST | /slack/events | Slackイベント受信 |
| POST | /slack/slashs | スラッシュコマンド（請求書作成モーダル） |
| POST | /slack/interactions | モーダル送信・ボタンアクション処理 |

## E2Eフロー

1. 請求書作成（draft状態）- 請求先はSlackユーザーIDで指定（`billing_slack_user_id`）
2. A(発行者): draft → sent ステータス遷移（PDF生成SQS送信 + 請求先Slack DM通知「支払った」ボタン付き）
3. PDF生成Lambda実行 → Slack DMで請求先ユーザーへPDFファイル送信
4. B(受取人): sent → paid（支払い報告）→ A宛にSlack DM「確認/差し戻し」ボタン付き
5. A: paid → confirmed（支払い確認）→ B宛に確認完了DM / または paid → sent（差し戻し）→ B宛に差し戻しDM
6. Web UIで一覧・詳細・PDF確認（ロールに応じたボタン表示）
7. 受領済み請求書は「受領済み」タブで確認可能
8. リマインダーLambdaが期限間近の請求書をSlack通知

### Slackフロー（スラッシュコマンド）

1. `/invoice-gen` → モーダル表示（請求先担当者:任意）
2. 送信 → 請求書作成（draft） → 即sent遷移
3. PDF生成（SQS送信）
4. 発行者にDM: 「請求書を作成・送付しました」
5. 請求先担当者にDM: 「請求書が届きました」+「支払った」ボタン（選択されていれば）

### Slackフロー（ボタンアクション）

1. B が「支払った」ボタン → sent → paid → A にDM（「確認」「差し戻し」ボタン付き）
2. A が「確認」ボタン → paid → confirmed → B にDM（確認完了通知）
3. A が「差し戻し」ボタン → paid → sent → B にDM（差し戻し通知 +「支払った」ボタン再表示）

## デプロイ

各コンポーネントのデプロイ方法については、それぞれのディレクトリ内のドキュメントを参照してください。

## CI/CD

GitHub Actionsを使用したCI/CDパイプラインが設定されています:

### PR時（CI）
- `infra/` 変更時に全Lambda（Go + Python）をビルドしTerraform planを自動実行、PRコメントに結果を表示
- `/apply` コメントで手動Terraform apply（レビュー後の即時適用）

### mainマージ時（CD）
- `incra_api_server/`, `pdf_generate/`, `infra/` の変更を検知して自動デプロイ
- Go Lambda（API Server + Reminder）とPython Lambda（PDF Generator）をビルド
- OIDC認証でAWSにアクセスし、Terraform apply -auto-approveを実行
- Terraform stateはS3（`incra-terraform-state`）で管理、DynamoDB（`incra-terraform-locks`）でロック

### Webフロントエンド
- Cloudflare Workersは `npm run deploy` でデプロイ

## 開発ガイド

開発者向けの詳細なガイドラインは [CLAUDE.md](CLAUDE.md) を参照してください。

## サポート

問題が発生した場合は、GitHubのIssuesで報告してください。
