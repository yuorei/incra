import { redirect, Form, Link, useNavigation } from "react-router";
import type { Route } from "./+types/clients.$clientId";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";

type Client = {
  client_id: string;
  name: string;
  slack_user_id?: string;
  slack_real_name?: string;
  email?: string;
  phone?: string;
  address?: string;
  bank_details?: string;
  registered_by: string;
  created_at: string;
};

export async function loader({ request, context, params }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const res = await apiFetch(env, session, `/clients/${params.clientId}`);
  if (!res.ok) throw redirect("/clients");
  const client: Client = await res.json();
  return { client, user: session };
}

export async function action({ request, context, params }: Route.ActionArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");

  const formData = await request.formData();
  const intent = formData.get("intent") as string;

  if (intent === "delete") {
    const res = await apiFetch(env, session, `/clients/${params.clientId}`, {
      method: "DELETE",
    });
    if (!res.ok) {
      const error = await res.text();
      return { error: `削除に失敗しました: ${error}` };
    }
    throw redirect("/clients");
  }

  if (intent === "update") {
    const body = {
      name: formData.get("name") as string,
      slack_user_id: (formData.get("slack_user_id") as string) || undefined,
      email: (formData.get("email") as string) || undefined,
      phone: (formData.get("phone") as string) || undefined,
      address: (formData.get("address") as string) || undefined,
      bank_details: (formData.get("bank_details") as string) || undefined,
    };

    const res = await apiFetch(env, session, `/clients/${params.clientId}`, {
      method: "PUT",
      body: JSON.stringify(body),
    });

    if (!res.ok) {
      const error = await res.text();
      return { error: `更新に失敗しました: ${error}` };
    }

    return { success: true };
  }

  return null;
}

export default function ClientDetail({ loaderData, actionData }: Route.ComponentProps) {
  const { client } = loaderData;
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
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-gray-700">取引先詳細</h2>
          <Link to="/clients" className="text-sm text-blue-600 hover:underline">一覧に戻る</Link>
        </div>
        {actionData?.error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
            {actionData.error}
          </div>
        )}
        {actionData?.success && (
          <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded mb-4">
            更新しました。
          </div>
        )}
        <Form method="post" className="bg-white shadow rounded-lg p-6 space-y-4">
          <input type="hidden" name="intent" value="update" />
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              会社名 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              required
              defaultValue={client.name}
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
              defaultValue={client.slack_user_id || ""}
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
              defaultValue={client.email || ""}
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
              defaultValue={client.phone || ""}
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
              defaultValue={client.address || ""}
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
              defaultValue={client.bank_details || ""}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={isSubmitting}
              className="bg-blue-600 text-white px-6 py-2 rounded text-sm hover:bg-blue-700 disabled:opacity-50"
            >
              {isSubmitting ? "更新中..." : "更新"}
            </button>
          </div>
        </Form>
        <Form method="post" className="mt-6">
          <input type="hidden" name="intent" value="delete" />
          <button
            type="submit"
            className="text-sm text-red-600 hover:text-red-800 underline"
            onClick={(e) => {
              if (!confirm("本当に削除しますか？")) e.preventDefault();
            }}
          >
            この取引先を削除
          </button>
        </Form>
      </main>
    </div>
  );
}
