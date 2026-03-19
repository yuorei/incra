import { Link } from "react-router";
import type { Route } from "./+types/_index";
import { getSession } from "../lib/session";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "incra - Slack連携の請求書管理システム" },
    { name: "description", content: "Slackで完結する請求書管理システム。自動PDF生成、Slack通知、ワンクリック承認。" },
  ];
}

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  return { isLoggedIn: !!session };
}

export default function Home({ loaderData }: Route.ComponentProps) {
  const { isLoggedIn } = loaderData;

  return (
    <div className="min-h-screen bg-white dark:bg-gray-900">
      {/* ナビバー */}
      <header className="border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold text-gray-800 dark:text-white">incra</Link>
          {isLoggedIn ? (
            <Link
              to="/invoices"
              className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
            >
              ダッシュボードへ
            </Link>
          ) : (
            <Link
              to="/login"
              className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700"
            >
              サインイン
            </Link>
          )}
        </div>
      </header>

      {/* ヒーローセクション */}
      <section className="py-20 px-4">
        <div className="max-w-4xl mx-auto text-center">
          <h1 className="text-4xl md:text-5xl font-bold text-gray-900 dark:text-white mb-6">
            Slackで完結する<br />請求書管理システム
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-400 mb-8 max-w-2xl mx-auto">
            請求書の作成からPDF生成、送信、支払い確認まで、すべてをSlackと連携して管理。
            チームの請求業務をシンプルに。
          </p>
          {isLoggedIn ? (
            <Link
              to="/invoices"
              className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg text-lg font-semibold hover:bg-blue-700 transition"
            >
              ダッシュボードへ
            </Link>
          ) : (
            <Link
              to="/login"
              className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg text-lg font-semibold hover:bg-blue-700 transition"
            >
              今すぐ始める
            </Link>
          )}
          <div className="mt-12">
            <img
              src="/images/hero-dashboard.png"
              alt="incraのダッシュボード画面。請求書一覧が表示されており、ステータスごとのフィルタータブ、請求書ID・請求先・金額・期限・ステータスの列が並ぶテーブルが確認できる。"
              className="rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 mx-auto max-w-full"
            />
          </div>
        </div>
      </section>

      {/* 特徴セクション */}
      <section className="py-20 px-4 bg-gray-50 dark:bg-gray-800">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-center text-gray-900 dark:text-white mb-12">
            主な特徴
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-white dark:bg-gray-900 rounded-lg p-6 shadow-sm">
              <img
                src="/images/feature-slack.png"
                alt="Slack連携のイメージ。Slackのチャンネル内で請求書通知メッセージと「支払った」「確認」アクションボタンが表示されている様子。"
                className="w-full h-40 object-cover rounded mb-4 bg-gray-100 dark:bg-gray-700"
              />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">Slack連携</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                請求書の通知をSlack DMで自動送信。「支払った」「確認」ボタンでワンクリック対応が可能です。
              </p>
            </div>
            <div className="bg-white dark:bg-gray-900 rounded-lg p-6 shadow-sm">
              <img
                src="/images/feature-pdf.png"
                alt="PDF生成のイメージ。請求書のフォーム入力画面から矢印が伸び、生成されたPDFドキュメントのプレビューが表示されている。"
                className="w-full h-40 object-cover rounded mb-4 bg-gray-100 dark:bg-gray-700"
              />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">自動PDF生成</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                フォームに入力するだけでPDFを自動生成。生成されたPDFはSlack DMで請求先へ直接送信されます。
              </p>
            </div>
            <div className="bg-white dark:bg-gray-900 rounded-lg p-6 shadow-sm">
              <img
                src="/images/feature-status.png"
                alt="ステータス管理のイメージ。下書き→送信済み→支払い済み→承認済みの4つのステータスが矢印で順番につながったフローチャート。"
                className="w-full h-40 object-cover rounded mb-4 bg-gray-100 dark:bg-gray-700"
              />
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">ステータス管理</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                下書き → 送信済み → 支払い済み → 承認済み。請求書のライフサイクルを一目で把握できます。
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* 使い方セクション */}
      <section className="py-20 px-4">
        <div className="max-w-4xl mx-auto">
          <h2 className="text-3xl font-bold text-center text-gray-900 dark:text-white mb-12">
            使い方
          </h2>
          <div className="space-y-12">
            <div className="flex items-start gap-6">
              <div className="flex-shrink-0 w-12 h-12 bg-blue-600 text-white rounded-full flex items-center justify-center text-xl font-bold">
                1
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  請求書を作成
                </h3>
                <p className="text-gray-600 dark:text-gray-400">
                  Webフォームから請求書を作成。Slackスラッシュコマンドからも直接作成できます。
                  請求先、品目、金額、支払期限を入力するだけ。
                </p>
              </div>
            </div>
            <div className="flex items-start gap-6">
              <div className="flex-shrink-0 w-12 h-12 bg-blue-600 text-white rounded-full flex items-center justify-center text-xl font-bold">
                2
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  PDF生成・送信
                </h3>
                <p className="text-gray-600 dark:text-gray-400">
                  送信ボタンを押すと、PDFが自動生成され、請求先のSlack DMへファイルとして送信されます。
                  「支払った」ボタン付きの通知も一緒に届きます。
                </p>
              </div>
            </div>
            <div className="flex items-start gap-6">
              <div className="flex-shrink-0 w-12 h-12 bg-blue-600 text-white rounded-full flex items-center justify-center text-xl font-bold">
                3
              </div>
              <div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  支払い・承認
                </h3>
                <p className="text-gray-600 dark:text-gray-400">
                  請求先がSlack上で「支払った」ボタンをクリック。発行者に通知が届き、
                  「確認」または「差し戻し」をワンクリックで対応できます。
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTAセクション */}
      <section className="py-20 px-4 bg-blue-600">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl font-bold text-white mb-4">
            今すぐ始めましょう
          </h2>
          <p className="text-blue-100 mb-8">
            Slackアカウントでサインインして、請求書管理をもっとシンプルに。
          </p>
          {isLoggedIn ? (
            <Link
              to="/invoices"
              className="inline-block bg-white text-blue-600 px-8 py-3 rounded-lg text-lg font-semibold hover:bg-blue-50 transition"
            >
              ダッシュボードへ
            </Link>
          ) : (
            <Link
              to="/login"
              className="inline-block bg-white text-blue-600 px-8 py-3 rounded-lg text-lg font-semibold hover:bg-blue-50 transition"
            >
              サインイン
            </Link>
          )}
        </div>
      </section>

      {/* フッター */}
      <footer className="py-8 px-4 border-t border-gray-200 dark:border-gray-700">
        <div className="max-w-6xl mx-auto text-center text-sm text-gray-500 dark:text-gray-400">
          &copy; incra
        </div>
      </footer>
    </div>
  );
}
