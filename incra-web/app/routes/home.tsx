import { redirect, Form, Link } from "react-router";
import type { Route } from "./+types/home";
import { getSession, clearSessionCookie } from "../lib/session";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "incra" },
    { name: "description", content: "incra - 請求書生成システム" },
  ];
}

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (!session) throw redirect("/login");
  return { user: session };
}

export async function action({ request }: Route.ActionArgs) {
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "logout") {
    return new Response(null, {
      status: 302,
      headers: {
        Location: "/login",
        "Set-Cookie": clearSessionCookie(),
      },
    });
  }

  return null;
}

export default function Home({ loaderData }: Route.ComponentProps) {
  const { user } = loaderData;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-800">incra</h1>
          <div className="flex items-center gap-4">
            {user.avatarUrl && (
              <img
                src={user.avatarUrl}
                alt={user.name}
                className="w-8 h-8 rounded-full"
              />
            )}
            <span className="text-gray-700 text-sm">{user.name}</span>
            <Form method="post">
              <input type="hidden" name="intent" value="logout" />
              <button
                type="submit"
                className="text-sm text-gray-500 hover:text-gray-700 underline"
              >
                ログアウト
              </button>
            </Form>
          </div>
        </div>
      </header>
      <main className="max-w-4xl mx-auto px-4 py-8">
        <p className="text-gray-600 mb-6">ようこそ、{user.name} さん！</p>
        <div className="grid grid-cols-2 gap-4">
          <Link to="/invoices" className="block bg-white shadow rounded-lg p-6 hover:shadow-md transition">
            <h2 className="text-lg font-semibold text-gray-700">請求書管理</h2>
            <p className="text-sm text-gray-500 mt-1">請求書の作成・管理</p>
          </Link>
          <Link to="/clients" className="block bg-white shadow rounded-lg p-6 hover:shadow-md transition">
            <h2 className="text-lg font-semibold text-gray-700">取引先管理</h2>
            <p className="text-sm text-gray-500 mt-1">取引先の登録・管理</p>
          </Link>
        </div>
      </main>
    </div>
  );
}
