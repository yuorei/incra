import { redirect, Link, useSearchParams } from "react-router";
import type { Route } from "./+types/invoices._index";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";

type Invoice = {
  invoice_id: string;
  billing_client_id: string;
  billing_client_name?: string;
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
  cancelled: "キャンセル",
};

const STATUS_COLORS: Record<string, string> = {
  draft: "bg-gray-100 text-gray-700",
  sent: "bg-blue-100 text-blue-700",
  paid: "bg-green-100 text-green-700",
  cancelled: "bg-red-100 text-red-700",
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
  const params = new URLSearchParams();
  if (status) params.set("status", status);
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

  const tabs = [
    { label: "全て", value: "" },
    { label: "下書き", value: "draft" },
    { label: "送信済み", value: "sent" },
    { label: "支払い済み", value: "paid" },
    { label: "キャンセル", value: "cancelled" },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <nav className="flex gap-4 items-center">
            <Link to="/" className="text-xl font-bold text-gray-800">incra</Link>
            <Link to="/invoices" className="text-sm text-blue-600 hover:underline font-semibold">請求書</Link>
            <Link to="/clients" className="text-sm text-blue-600 hover:underline">取引先</Link>
          </nav>
          <Link to="/invoices/new" className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">
            新規作成
          </Link>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-4 py-8">
        <h2 className="text-lg font-semibold text-gray-700 mb-4">請求書一覧</h2>
        <div className="flex gap-2 mb-6">
          {tabs.map((tab) => (
            <Link
              key={tab.value}
              to={tab.value ? `/invoices?status=${tab.value}` : "/invoices"}
              className={`px-3 py-1.5 rounded text-sm ${
                currentStatus === tab.value
                  ? "bg-blue-600 text-white"
                  : "bg-white text-gray-600 border border-gray-300 hover:bg-gray-50"
              }`}
            >
              {tab.label}
            </Link>
          ))}
        </div>
        {invoices.length === 0 ? (
          <p className="text-gray-500">請求書がありません。</p>
        ) : (
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-4 py-3 text-left text-gray-600">請求書ID</th>
                  <th className="px-4 py-3 text-left text-gray-600">取引先</th>
                  <th className="px-4 py-3 text-right text-gray-600">合計金額</th>
                  <th className="px-4 py-3 text-left text-gray-600">期限</th>
                  <th className="px-4 py-3 text-left text-gray-600">ステータス</th>
                  <th className="px-4 py-3 text-left text-gray-600">操作</th>
                </tr>
              </thead>
              <tbody>
                {invoices.map((inv) => (
                  <tr key={inv.invoice_id} className="border-b hover:bg-gray-50">
                    <td className="px-4 py-3 text-gray-900 font-mono text-xs">{inv.invoice_id.slice(0, 8)}</td>
                    <td className="px-4 py-3 text-gray-900">{inv.billing_client_name || inv.billing_client_id}</td>
                    <td className="px-4 py-3 text-right text-gray-900 font-medium">{formatYen(inv.total_amount)}</td>
                    <td className="px-4 py-3 text-gray-600">{inv.due_date}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[inv.status] || "bg-gray-100 text-gray-700"}`}>
                        {STATUS_LABELS[inv.status] || inv.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <Link to={`/invoices/${inv.invoice_id}`} className="text-blue-600 hover:underline">
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
