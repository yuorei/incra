import { Form, Link } from "react-router";
import type { SessionUser } from "../lib/session";

type AuthHeaderProps = {
  user: SessionUser;
  actions?: React.ReactNode;
};

export function AuthHeader({ user, actions }: AuthHeaderProps) {
  return (
    <header className="bg-white dark:bg-gray-800 shadow-sm">
      <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
        <nav className="flex gap-4 items-center">
          <Link to="/" className="text-xl font-bold text-gray-800 dark:text-white">incra</Link>
          <Link to="/invoices" className="text-sm text-blue-600 dark:text-blue-400 hover:underline font-semibold">請求書</Link>
        </nav>
        <div className="flex items-center gap-4">
          {actions}
          {user.avatarUrl && (
            <img
              src={user.avatarUrl}
              alt={user.name}
              className="w-8 h-8 rounded-full"
            />
          )}
          <span className="text-gray-700 dark:text-gray-300 text-sm">{user.name}</span>
          <Form method="post" action="/logout">
            <button
              type="submit"
              className="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 underline"
            >
              ログアウト
            </button>
          </Form>
        </div>
      </div>
    </header>
  );
}
