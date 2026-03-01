import { redirect, Form, Link, useNavigation } from "react-router";
import type { Route } from "./+types/invoices.$invoiceId";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";

type InvoiceItem = {
  date: string;
  description: string;
  quantity: number;
  unit_price: number;
  amount: number;
  memo?: string;
};

type HistoryEntry = {
  changed_at: string;
  old_status: string;
  new_status: string;
  changed_by: string;
};

type Invoice = {
  invoice_id: string;
  billing_client_id: string;
  billing_client_name?: string;
  total_amount: number;
  due_date: string;
  status: string;
  pdf_url?: string;
  bank_details?: string;
  additional_info?: string;
  created_at: string;
  items?: InvoiceItem[];
  history?: HistoryEntry[];
};

const STATUS_LABELS: Record<string, string> = {
  draft: "下書き",
  sent: "送信済み",
  paid: "支払い済み",
  cancelled: "キャンセル",
};

const STATUS_COLORS: Record<string, string> = {
  draft: "bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300",
  sent: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300",
  paid: "bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300",
  cancelled: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
};

function formatYen(amount: number): string {
  return `\u00a5${amount.toLocaleString()}`;
}

export async function loader({ request, context, params }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const res = await apiFetch(env, session, `/invoices/${params.invoiceId}`);
  if (!res.ok) throw redirect("/invoices");
  const invoice: Invoice = await res.json();
  return { invoice, user: session };
}

export async function action({ request, context, params }: Route.ActionArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const formData = await request.formData();
  const intent = formData.get("intent") as string;

  if (intent === "transition") {
    const newStatus = formData.get("new_status") as string;
    const res = await apiFetch(env, session, `/invoices/${params.invoiceId}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status: newStatus }),
    });

    if (!res.ok) {
      const error = await res.text();
      return { error: `ステータス変更に失敗しました: ${error}` };
    }

    return redirect(`/invoices/${params.invoiceId}`);
  }

  return null;
}

export default function InvoiceDetail({ loaderData, actionData }: Route.ComponentProps) {
  const { invoice } = loaderData;
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-white dark:bg-gray-800 shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <nav className="flex gap-4 items-center">
            <Link to="/" className="text-xl font-bold text-gray-800 dark:text-white">incra</Link>
            <Link to="/invoices" className="text-sm text-blue-600 dark:text-blue-400 hover:underline font-semibold">請求書</Link>
            <Link to="/clients" className="text-sm text-blue-600 dark:text-blue-400 hover:underline">取引先</Link>
          </nav>
        </div>
      </header>
      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-gray-700 dark:text-gray-300">請求書詳細</h2>
          <div className="flex gap-3">
            {invoice.status === "draft" && (
              <Link
                to={`/invoices/${invoice.invoice_id}/edit`}
                className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
              >
                編集
              </Link>
            )}
            <Link to="/invoices" className="text-sm text-blue-600 dark:text-blue-400 hover:underline">一覧に戻る</Link>
          </div>
        </div>
        {actionData?.error && (
          <div className="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 text-red-700 dark:text-red-300 px-4 py-3 rounded mb-4">
            {actionData.error}
          </div>
        )}

        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6 mb-6">
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <span className="text-xs text-gray-500 dark:text-gray-400">請求書ID</span>
              <p className="text-sm font-mono text-gray-900 dark:text-gray-100">{invoice.invoice_id}</p>
            </div>
            <div>
              <span className="text-xs text-gray-500 dark:text-gray-400">ステータス</span>
              <p>
                <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[invoice.status] || "bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300"}`}>
                  {STATUS_LABELS[invoice.status] || invoice.status}
                </span>
              </p>
            </div>
            <div>
              <span className="text-xs text-gray-500 dark:text-gray-400">取引先</span>
              <p className="text-sm text-gray-900 dark:text-gray-100">{invoice.billing_client_name || invoice.billing_client_id}</p>
            </div>
            <div>
              <span className="text-xs text-gray-500 dark:text-gray-400">支払期限</span>
              <p className="text-sm text-gray-900 dark:text-gray-100">{invoice.due_date}</p>
            </div>
            <div>
              <span className="text-xs text-gray-500 dark:text-gray-400">合計金額</span>
              <p className="text-lg font-semibold text-gray-900 dark:text-gray-100">{formatYen(invoice.total_amount)}</p>
            </div>
            {invoice.bank_details && (
              <div>
                <span className="text-xs text-gray-500 dark:text-gray-400">振込先</span>
                <p className="text-sm text-gray-900 dark:text-gray-100">{invoice.bank_details}</p>
              </div>
            )}
          </div>
          {invoice.additional_info && (
            <div className="mb-4">
              <span className="text-xs text-gray-500 dark:text-gray-400">備考</span>
              <p className="text-sm text-gray-900 dark:text-gray-100 whitespace-pre-wrap">{invoice.additional_info}</p>
            </div>
          )}
          {invoice.pdf_url && (
            <div className="mb-4">
              <a
                href={invoice.pdf_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
              >
                PDFを表示
              </a>
            </div>
          )}

          <div className="flex gap-2 pt-4 border-t dark:border-gray-700">
            {invoice.status === "draft" && (
              <>
                <Form method="post">
                  <input type="hidden" name="intent" value="transition" />
                  <input type="hidden" name="new_status" value="sent" />
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-50"
                  >
                    PDF生成・送信
                  </button>
                </Form>
                <Form method="post">
                  <input type="hidden" name="intent" value="transition" />
                  <input type="hidden" name="new_status" value="cancelled" />
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="border border-red-300 dark:border-red-600 text-red-600 dark:text-red-400 px-4 py-2 rounded text-sm hover:bg-red-50 dark:hover:bg-red-900/30 disabled:opacity-50"
                    onClick={(e) => {
                      if (!confirm("キャンセルしますか？")) e.preventDefault();
                    }}
                  >
                    キャンセル
                  </button>
                </Form>
              </>
            )}
            {invoice.status === "sent" && (
              <>
                <Form method="post">
                  <input type="hidden" name="intent" value="transition" />
                  <input type="hidden" name="new_status" value="paid" />
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 disabled:opacity-50"
                  >
                    支払い完了
                  </button>
                </Form>
                <Form method="post">
                  <input type="hidden" name="intent" value="transition" />
                  <input type="hidden" name="new_status" value="cancelled" />
                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="border border-red-300 dark:border-red-600 text-red-600 dark:text-red-400 px-4 py-2 rounded text-sm hover:bg-red-50 dark:hover:bg-red-900/30 disabled:opacity-50"
                    onClick={(e) => {
                      if (!confirm("キャンセルしますか？")) e.preventDefault();
                    }}
                  >
                    キャンセル
                  </button>
                </Form>
              </>
            )}
          </div>
        </div>

        {invoice.items && invoice.items.length > 0 && (
          <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6 mb-6">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">品目</h3>
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-900 border-b dark:border-gray-700">
                <tr>
                  <th className="px-3 py-2 text-left text-gray-600 dark:text-gray-400">日付</th>
                  <th className="px-3 py-2 text-left text-gray-600 dark:text-gray-400">内容</th>
                  <th className="px-3 py-2 text-right text-gray-600 dark:text-gray-400">数量</th>
                  <th className="px-3 py-2 text-right text-gray-600 dark:text-gray-400">単価</th>
                  <th className="px-3 py-2 text-right text-gray-600 dark:text-gray-400">金額</th>
                  <th className="px-3 py-2 text-left text-gray-600 dark:text-gray-400">メモ</th>
                </tr>
              </thead>
              <tbody>
                {invoice.items.map((item, index) => (
                  <tr key={index} className="border-b dark:border-gray-700">
                    <td className="px-3 py-2 text-gray-600 dark:text-gray-400">{item.date}</td>
                    <td className="px-3 py-2 text-gray-900 dark:text-gray-100">{item.description}</td>
                    <td className="px-3 py-2 text-right text-gray-600 dark:text-gray-400">{item.quantity}</td>
                    <td className="px-3 py-2 text-right text-gray-600 dark:text-gray-400">{formatYen(item.unit_price)}</td>
                    <td className="px-3 py-2 text-right font-medium text-gray-900 dark:text-gray-100">{formatYen(item.amount)}</td>
                    <td className="px-3 py-2 text-gray-500 dark:text-gray-400">{item.memo || ""}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {invoice.history && invoice.history.length > 0 && (
          <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">履歴</h3>
            <div className="space-y-3">
              {invoice.history.map((entry, index) => (
                <div key={index} className="flex items-start gap-3">
                  <div className="w-2 h-2 rounded-full bg-blue-400 dark:bg-blue-500 mt-1.5 shrink-0" />
                  <div>
                    <p className="text-sm text-gray-900 dark:text-gray-100">
                      <span className={`inline-block px-1.5 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[entry.old_status] || "bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300"}`}>
                        {STATUS_LABELS[entry.old_status] || entry.old_status}
                      </span>
                      <span className="mx-1 text-gray-400">&rarr;</span>
                      <span className={`inline-block px-1.5 py-0.5 rounded text-xs font-medium ${STATUS_COLORS[entry.new_status] || "bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300"}`}>
                        {STATUS_LABELS[entry.new_status] || entry.new_status}
                      </span>
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                      {entry.changed_at} - {entry.changed_by}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
