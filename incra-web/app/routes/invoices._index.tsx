import { redirect, Link, useSearchParams } from "react-router";
import type { Route } from "./+types/invoices._index";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";
import { AuthHeader } from "../components/auth-header";

type Invoice = {
  invoice_id: string;
  billing_client_id?: string;
  billing_slack_user_id?: string;
  billing_client_name?: string;
  issuer_slack_user_id?: string;
  issuer_slack_real_name?: string;
  total_amount: number;
  due_date: string;
  status: string;
  pdf_url?: string;
  created_at: string;
};

const STATUS_LABELS: Record<string, string> = {
  draft: "下書き",
  sent: "送信済み",
  paid: "支払い済み",
  confirmed: "承認済み",
  cancelled: "キャンセル",
};

const STATUS_COLORS: Record<string, string> = {
  draft: "bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300",
  sent: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300",
  paid: "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300",
  confirmed: "bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300",
  cancelled: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
};

function formatYen(amount: number): string {
  return `\u00a5${amount.toLocaleString()}`;
}

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const url = new URL(request.url);
  const status = url.searchParams.get("status");
  const type = url.searchParams.get("type");
  const params = new URLSearchParams();
  if (status) params.set("status", status);
  if (type === "received") params.set("type", "received");
  const query = params.toString() ? `?${params.toString()}` : "";

  const res = await apiFetch(env, session, `/invoices${query}`);
  if (!res.ok) return { invoices: [] as Invoice[], user: session };
  const data = await res.json() as { invoices?: Invoice[]; next_key?: string };
  const invoices: Invoice[] = data.invoices || [];
  return { invoices, user: session };
}

export default function InvoicesIndex({ loaderData }: Route.ComponentProps) {
  const { invoices } = loaderData;
  const [searchParams] = useSearchParams();
  const currentStatus = searchParams.get("status") || "";
  const currentType = searchParams.get("type") || "";
  const isReceived = currentType === "received";

  const typeTabs = [
    { label: "発行した請求書", value: "" },
    { label: "受け取った請求書", value: "received" },
  ];

  const statusTabs = [
    { label: "全て", value: "" },
    { label: "下書き", value: "draft" },
    { label: "送信済み", value: "sent" },
    { label: "支払い済み", value: "paid" },
    { label: "承認済み", value: "confirmed" },
    { label: "キャンセル", value: "cancelled" },
  ];

  function buildUrl(opts: { type?: string; status?: string }) {
    const p = new URLSearchParams();
    const t = opts.type !== undefined ? opts.type : currentType;
    const s = opts.status !== undefined ? opts.status : currentStatus;
    if (t) p.set("type", t);
    if (s) p.set("status", s);
    const q = p.toString();
    return q ? `/invoices?${q}` : "/invoices";
  }

  const { user } = loaderData;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <AuthHeader
        user={user}
        actions={
          <Link to="/invoices/new" className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">
            新規作成
          </Link>
        }
      />
      <main className="max-w-6xl mx-auto px-4 py-8">
        <h2 className="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-4">請求書一覧</h2>
        <div className="flex gap-2 mb-4">
          {typeTabs.map((tab) => (
            <Link
              key={tab.value}
              to={buildUrl({ type: tab.value, status: "" })}
              className={`px-3 py-1.5 rounded text-sm ${
                currentType === tab.value
                  ? "bg-blue-600 text-white"
                  : "bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700"
              }`}
            >
              {tab.label}
            </Link>
          ))}
        </div>
        <div className="flex gap-2 mb-6">
          {statusTabs.map((tab) => (
            <Link
              key={tab.value}
              to={buildUrl({ status: tab.value })}
              className={`px-3 py-1.5 rounded text-sm ${
                currentStatus === tab.value
                  ? "bg-blue-600 text-white"
                  : "bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700"
              }`}
            >
              {tab.label}
            </Link>
          ))}
        </div>
        {invoices.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400">請求書がありません。</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 shadow rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-900 border-b dark:border-gray-700">
                <tr>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">請求書ID</th>
                  {isReceived ? (
                    <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">発行者</th>
                  ) : (
                    <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">請求先</th>
                  )}
                  <th className="px-4 py-3 text-right text-gray-600 dark:text-gray-400">合計金額</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">期限</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">ステータス</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">操作</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv) => (
                  <tr key={inv.invoice_id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-4 py-3 text-gray-900 dark:text-gray-100 font-mono text-xs">{inv.invoice_id.slice(0, 8)}</td>
                    <td className="px-4 py-3 text-gray-900 dark:text-gray-100">
                      {isReceived
                        ? (inv.issuer_slack_real_name || inv.issuer_slack_user_id || "-")
                        : (inv.billing_client_name || inv.billing_slack_user_id || inv.billing_client_id || "-")
                      }
                    </td>
                    <td className="px-4 py-3 text-right text-gray-900 dark:text-gray-100 font-medium">{formatYen(inv.total_amount)}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{inv.due_date}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[inv.status] || "bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300"}`}>
                        {STATUS_LABELS[inv.status] || inv.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <Link to={`/invoices/${inv.invoice_id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
                        詳細
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </main>
    </div>
  );
}
