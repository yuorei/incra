import { redirect } from "react-router";
import type { Route } from "./+types/login";
import { getSession } from "../lib/session";
import { generateState, setStateCookie } from "../lib/oauth-state";

export function meta({}: Route.MetaArgs) {
  return [{ title: "ログイン - incra" }];
}

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const session = await getSession(request, env.SESSION_SECRET);
  if (session) throw redirect("/invoices");

  const url = new URL(request.url);
  const error = url.searchParams.get("error");
  return { error };
}

export async function action({ request, context }: Route.ActionArgs) {
  const { env } = context.cloudflare;
  const state = generateState();

  const redirectUri = new URL(request.url);
  redirectUri.pathname = "/auth/slack/callback";
  redirectUri.search = "";

  const authorizeUrl = new URL("https://slack.com/openid/connect/authorize");
  authorizeUrl.searchParams.set("client_id", env.SLACK_CLIENT_ID);
  authorizeUrl.searchParams.set("scope", "openid profile email");
  authorizeUrl.searchParams.set("redirect_uri", redirectUri.toString());
  authorizeUrl.searchParams.set("response_type", "code");
  authorizeUrl.searchParams.set("state", state);

  return new Response(null, {
    status: 302,
    headers: {
      Location: authorizeUrl.toString(),
      "Set-Cookie": setStateCookie(state),
    },
  });
}

const ERROR_MESSAGES: Record<string, string> = {
  state_mismatch: "認証エラーが発生しました。再度お試しください。",
  slack_error: "Slack 認証に失敗しました。再度お試しください。",
  token_error: "トークンの取得に失敗しました。再度お試しください。",
};

export default function Login({ loaderData }: Route.ComponentProps) {
  const { error } = loaderData;

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-8 w-full max-w-sm">
        <h1 className="text-2xl font-bold text-center mb-6 text-gray-800 dark:text-white">
          incra にサインイン
        </h1>
        {error && (
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-700 dark:text-red-400 text-sm">
            {ERROR_MESSAGES[error] ?? "エラーが発生しました。再度お試しください。"}
          </div>
        )}
        <form method="post">
          <button
            type="submit"
            className="w-full flex items-center justify-center gap-3 py-3 px-4 rounded font-semibold text-white transition-opacity hover:opacity-90"
            style={{ backgroundColor: "#4A154B" }}
          >
            <SlackIcon />
            Sign in with Slack
          </button>
        </form>
      </div>
    </div>
  );
}

function SlackIcon() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 54 54"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M19.712.133a5.381 5.381 0 0 0-5.376 5.387 5.381 5.381 0 0 0 5.376 5.386h5.376V5.52A5.381 5.381 0 0 0 19.712.133m0 14.365H5.376A5.381 5.381 0 0 0 0 19.884a5.381 5.381 0 0 0 5.376 5.387h14.336a5.381 5.381 0 0 0 5.376-5.387 5.381 5.381 0 0 0-5.376-5.386"
        fill="#36C5F0"
      />
      <path
        d="M53.76 19.884a5.381 5.381 0 0 0-5.376-5.386 5.381 5.381 0 0 0-5.376 5.386v5.387h5.376a5.381 5.381 0 0 0 5.376-5.387m-14.336 0V5.52A5.381 5.381 0 0 0 34.048.133a5.381 5.381 0 0 0-5.376 5.387v14.364a5.381 5.381 0 0 0 5.376 5.387 5.381 5.381 0 0 0 5.376-5.387"
        fill="#2EB67D"
      />
      <path
        d="M34.048 54a5.381 5.381 0 0 0 5.376-5.387 5.381 5.381 0 0 0-5.376-5.386h-5.376v5.386A5.381 5.381 0 0 0 34.048 54m0-14.365h14.336a5.381 5.381 0 0 0 5.376-5.386 5.381 5.381 0 0 0-5.376-5.387H34.048a5.381 5.381 0 0 0-5.376 5.387 5.381 5.381 0 0 0 5.376 5.386"
        fill="#ECB22E"
      />
      <path
        d="M0 34.249a5.381 5.381 0 0 0 5.376 5.386 5.381 5.381 0 0 0 5.376-5.386v-5.387H5.376A5.381 5.381 0 0 0 0 34.249m14.336 0v14.364A5.381 5.381 0 0 0 19.712 54a5.381 5.381 0 0 0 5.376-5.387V34.249a5.381 5.381 0 0 0-5.376-5.387 5.381 5.381 0 0 0-5.376 5.387"
        fill="#E01E5A"
      />
    </svg>
  );
}
