import { redirect, Form, Link, useNavigation } from "react-router";
import { useState } from "react";
import type { Route } from "./+types/invoices.new";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";
import { SlackUserSelect, type SlackUser } from "../components/slack-user-select";
import { AuthHeader } from "../components/auth-header";

type ItemRow = {
  key: number;
  date: string;
  description: string;
  quantity: number;
  unit_price: number;
  memo: string;
};

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  let slackUsers: SlackUser[] = [];
  try {
    const res = await apiFetch(env, session, "/slack/users");
    if (res.ok) {
      slackUsers = await res.json();
    }
  } catch {
    // Don't block the page if Slack users fetch fails
  }

  return { slackUsers, user: session };
}

export async function action({ request, context }: Route.ActionArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const formData = await request.formData();
  const itemCount = parseInt(formData.get("item_count") as string, 10) || 0;
  const items = [];
  for (let i = 0; i < itemCount; i++) {
    const quantity = parseInt(formData.get(`items[${i}].quantity`) as string, 10) || 0;
    const unitPrice = parseInt(formData.get(`items[${i}].unit_price`) as string, 10) || 0;
    items.push({
      date: formData.get(`items[${i}].date`) as string,
      description: formData.get(`items[${i}].description`) as string,
      quantity,
      unit_price: unitPrice,
      amount: quantity * unitPrice,
      memo: (formData.get(`items[${i}].memo`) as string) || "",
    });
  }

  const billingSlackUserId = formData.get("billing_slack_user_id") as string;
  if (!billingSlackUserId) {
    return { error: "請求先を選択してください" };
  }

  const body = {
    billing_slack_user_id: billingSlackUserId,
    billing_client_name: formData.get("billing_client_name") as string,
    due_date: formData.get("due_date") as string,
    bank_details: (formData.get("bank_details") as string) || "",
    additional_info: (formData.get("additional_info") as string) || undefined,
    items,
  };

  const res = await apiFetch(env, session, "/invoices", {
    method: "POST",
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const error = await res.text();
    return { error: `作成に失敗しました: ${error}` };
  }

  const created = (await res.json()) as { invoice_id: string };
  throw redirect(`/invoices/${created.invoice_id}`);
}

let nextKey = 1;

export default function InvoicesNew({ loaderData, actionData }: Route.ComponentProps) {
  const { slackUsers } = loaderData;
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";
  const [selectedUser, setSelectedUser] = useState<SlackUser | null>(null);

  const [items, setItems] = useState<ItemRow[]>([
    { key: nextKey++, date: "", description: "", quantity: 1, unit_price: 0, memo: "" },
  ]);

  const addItem = () => {
    setItems([...items, { key: nextKey++, date: "", description: "", quantity: 1, unit_price: 0, memo: "" }]);
  };

  const removeItem = (key: number) => {
    if (items.length <= 1) return;
    setItems(items.filter((item) => item.key !== key));
  };

  const updateItem = (key: number, field: keyof ItemRow, value: string | number) => {
    setItems(items.map((item) => (item.key === key ? { ...item, [field]: value } : item)));
  };

  const total = items.reduce((sum, item) => sum + item.quantity * item.unit_price, 0);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <AuthHeader user={loaderData.user} />
      <main className="max-w-4xl mx-auto px-4 py-8">
        <h2 className="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-6">請求書新規作成</h2>
        {actionData?.error && (
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 px-4 py-3 rounded mb-4">
            {actionData.error}
          </div>
        )}
        <Form method="post" className="bg-white dark:bg-gray-800 shadow rounded-lg p-6 space-y-6">
          <input type="hidden" name="item_count" value={items.length} />
          <input type="hidden" name="billing_slack_user_id" value={selectedUser?.id ?? ""} />
          <input type="hidden" name="billing_client_name" value={selectedUser?.real_name ?? ""} />
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                請求先 <span className="text-red-500">*</span>
              </label>
              <SlackUserSelect
                users={slackUsers}
                excludeUserId={loaderData.user.id}
                onSelect={setSelectedUser}
              />
            </div>
            <div>
              <label htmlFor="due_date" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                支払期限 <span className="text-red-500">*</span>
              </label>
              <input
                type="date"
                id="due_date"
                name="due_date"
                required
                className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div>
            <label htmlFor="bank_details" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              振込先情報
            </label>
            <input
              type="text"
              id="bank_details"
              name="bank_details"
              className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label htmlFor="additional_info" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              備考
            </label>
            <textarea
              id="additional_info"
              name="additional_info"
              rows={2}
              className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">品目</h3>
              <button
                type="button"
                onClick={addItem}
                className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
              >
                + 行を追加
              </button>
            </div>
            <div className="space-y-3">
              {items.map((item, index) => (
                <div key={item.key} className="border border-gray-200 dark:border-gray-600 rounded p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-500 dark:text-gray-400">品目 {index + 1}</span>
                    {items.length > 1 && (
                      <button
                        type="button"
                        onClick={() => removeItem(item.key)}
                        className="text-xs text-red-500 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300"
                      >
                        削除
                      </button>
                    )}
                  </div>
                  <div className="grid grid-cols-5 gap-2">
                    <div>
                      <label className="block text-xs text-gray-500 dark:text-gray-400 mb-0.5">日付</label>
                      <input
                        type="date"
                        name={`items[${index}].date`}
                        value={item.date}
                        onChange={(e) => updateItem(item.key, "date", e.target.value)}
                        className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-2 py-1 text-sm"
                      />
                    </div>
                    <div className="col-span-2">
                      <label className="block text-xs text-gray-500 dark:text-gray-400 mb-0.5">内容</label>
                      <input
                        type="text"
                        name={`items[${index}].description`}
                        value={item.description}
                        onChange={(e) => updateItem(item.key, "description", e.target.value)}
                        required
                        className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-2 py-1 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-gray-500 dark:text-gray-400 mb-0.5">数量</label>
                      <input
                        type="number"
                        name={`items[${index}].quantity`}
                        value={item.quantity}
                        onChange={(e) => updateItem(item.key, "quantity", parseInt(e.target.value, 10) || 0)}
                        min={1}
                        step="1"
                        required
                        className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-2 py-1 text-sm"
                      />
                    </div>
                    <div>
                      <label className="block text-xs text-gray-500 dark:text-gray-400 mb-0.5">単価</label>
                      <input
                        type="number"
                        name={`items[${index}].unit_price`}
                        value={item.unit_price}
                        onChange={(e) => updateItem(item.key, "unit_price", parseInt(e.target.value, 10) || 0)}
                        min={0}
                        step="1"
                        required
                        className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-2 py-1 text-sm"
                      />
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="flex-1 mr-2">
                      <label className="block text-xs text-gray-500 dark:text-gray-400 mb-0.5">メモ</label>
                      <input
                        type="text"
                        name={`items[${index}].memo`}
                        value={item.memo}
                        onChange={(e) => updateItem(item.key, "memo", e.target.value)}
                        className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded px-2 py-1 text-sm"
                      />
                    </div>
                    <div className="text-sm text-gray-700 dark:text-gray-300 font-medium pt-4">
                      {"\u00a5"}{(item.quantity * item.unit_price).toLocaleString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 text-right text-lg font-semibold text-gray-800 dark:text-white">
              合計: {"\u00a5"}{total.toLocaleString()}
            </div>
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="bg-blue-600 text-white px-6 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-50"
            >
              {isSubmitting ? "作成中..." : "作成"}
            </button>
            <Link to="/invoices" className="px-6 py-2 rounded text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 border border-gray-300 dark:border-gray-600">
              キャンセル
            </Link>
          </div>
        </Form>
      </main>
    </div>
  );
}
