import { redirect, Link } from "react-router";
import type { Route } from "./+types/clients._index";
import { getSession } from "../lib/session";
import { apiFetch } from "../lib/api";

type Client = {
  client_id: string;
  name: string;
  slack_user_id?: string;
  slack_real_name?: string;
  email?: string;
  phone?: string;
  registered_by: string;
  created_at: string;
};

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");
  const res = await apiFetch(env, session, "/clients");
  const clients: Client[] = res.ok ? await res.json() : [];
  return { clients, user: session };
}

export default function ClientsIndex({ loaderData }: Route.ComponentProps) {
  const { clients } = loaderData;
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-white dark:bg-gray-800 shadow-sm">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <nav className="flex gap-4 items-center">
            <Link to="/" className="text-xl font-bold text-gray-800 dark:text-white">incra</Link>
            <Link to="/invoices" className="text-sm text-blue-600 dark:text-blue-400 hover:underline">請求書</Link>
            <Link to="/clients" className="text-sm text-blue-600 dark:text-blue-400 hover:underline font-semibold">取引先</Link>
          </nav>
          <Link to="/clients/new" className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">
            新規登録
          </Link>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-4 py-8">
        <h2 className="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-4">取引先一覧</h2>
        {clients.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400">取引先が登録されていません。</p>
        ) : (
          <div className="bg-white dark:bg-gray-800 shadow rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-900 border-b dark:border-gray-700">
                <tr>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">会社名</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">Slack</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">メール</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">電話</th>
                  <th className="px-4 py-3 text-left text-gray-600 dark:text-gray-400">操作</th>
                </tr>
              </thead>
              <tbody>
                {clients.map((c) => (
                  <tr key={c.client_id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{c.name}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{c.slack_real_name || c.slack_user_id || "-"}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{c.email || "-"}</td>
                    <td className="px-4 py-3 text-gray-600 dark:text-gray-400">{c.phone || "-"}</td>
                    <td className="px-4 py-3">
                      <Link to={`/clients/${c.client_id}`} className="text-blue-600 dark:text-blue-400 hover:underline">
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
