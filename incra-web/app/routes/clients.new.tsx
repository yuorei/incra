import { redirect, Form, Link, useNavigation } from "react-router";
import type { Route } from "./+types/clients.new";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");
  return { user: session };
}

export async function action({ request, context }: Route.ActionArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const formData = await request.formData();
  const body = {
    name: formData.get("name") as string,
    slack_user_id: (formData.get("slack_user_id") as string) || undefined,
    email: (formData.get("email") as string) || undefined,
    phone: (formData.get("phone") as string) || undefined,
    address: (formData.get("address") as string) || undefined,
    bank_details: (formData.get("bank_details") as string) || undefined,
  };

  const res = await apiFetch(env, session, "/clients", {
    method: "POST",
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const error = await res.text();
    return { error: `登録に失敗しました: ${error}` };
  }

  throw redirect("/clients");
}

export default function ClientsNew({ actionData }: Route.ComponentProps) {
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <nav className="flex gap-4 items-center">
            <Link to="/" className="text-xl font-bold text-gray-800">incra</Link>
            <Link to="/invoices" className="text-sm text-blue-600 hover:underline">請求書</Link>
            <Link to="/clients" className="text-sm text-blue-600 hover:underline font-semibold">取引先</Link>
          </nav>
        </div>
      </header>
      <main className="max-w-2xl mx-auto px-4 py-8">
        <h2 className="text-lg font-semibold text-gray-700 mb-6">取引先新規登録</h2>
        {actionData?.error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
            {actionData.error}
          </div>
        )}
        <Form method="post" className="bg-white shadow rounded-lg p-6 space-y-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              会社名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              required
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label htmlFor="slack_user_id" className="block text-sm font-medium text-gray-700 mb-1">
              Slack User ID
            </label>
            <input
              type="text"
              id="slack_user_id"
              name="slack_user_id"
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-1">
              メールアドレス
            </label>
            <input
              type="email"
              id="email"
              name="email"
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-1">
              電話番号
            </label>
            <input
              type="tel"
              id="phone"
              name="phone"
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label htmlFor="address" className="block text-sm font-medium text-gray-700 mb-1">
              住所
            </label>
            <textarea
              id="address"
              name="address"
              rows={2}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label htmlFor="bank_details" className="block text-sm font-medium text-gray-700 mb-1">
              振込先情報
            </label>
            <textarea
              id="bank_details"
              name="bank_details"
              rows={2}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="bg-blue-600 text-white px-6 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-50"
            >
              {isSubmitting ? "登録中..." : "登録"}
            </button>
            <Link to="/clients" className="px-6 py-2 rounded text-sm text-gray-600 hover:text-gray-800 border border-gray-300">
              キャンセル
            </Link>
          </div>
        </Form>
      </main>
    </div>
  );
}
