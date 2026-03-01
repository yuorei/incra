import { redirect } from "react-router";
import type { Route } from "./+types/auth.slack.callback";
import { createSessionCookie } from "../lib/session";
import { getAndClearStateCookie } from "../lib/oauth-state";

export async function loader({ request, context }: Route.LoaderArgs) {
  const { env } = context.cloudflare;
  const url = new URL(request.url);

  const code = url.searchParams.get("code");
  const state = url.searchParams.get("state");
  const slackError = url.searchParams.get("error");

  const { state: savedState, clearCookie } = getAndClearStateCookie(request);

  if (slackError) {
    return new Response(null, {
      status: 302,
      headers: { Location: "/login?error=slack_error", "Set-Cookie": clearCookie },
    });
  }

  if (!state || !savedState || state !== savedState) {
    return new Response(null, {
      status: 302,
      headers: { Location: "/login?error=state_mismatch", "Set-Cookie": clearCookie },
    });
  }

  if (!code) {
    return new Response(null, {
      status: 302,
      headers: { Location: "/login?error=slack_error", "Set-Cookie": clearCookie },
    });
  }

  const redirectUri = new URL(request.url);
  redirectUri.pathname = "/auth/slack/callback";
  redirectUri.search = "";

  // Exchange authorization code for tokens (server-side only)
  const tokenRes = await fetch("https://slack.com/api/openid.connect.token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "authorization_code",
      client_id: env.SLACK_CLIENT_ID,
      client_secret: env.SLACK_CLIENT_SECRET,
      code,
      redirect_uri: redirectUri.toString(),
    }),
  });

  const tokenData = (await tokenRes.json()) as {
    ok: boolean;
    access_token?: string;
    error?: string;
  };

  console.log("[slack callback] token response:", JSON.stringify(tokenData));

  if (!tokenData.ok || !tokenData.access_token) {
    return new Response(null, {
      status: 302,
      headers: { Location: "/login?error=token_error", "Set-Cookie": clearCookie },
    });
  }

  // Fetch user info
  const userRes = await fetch("https://slack.com/api/openid.connect.userInfo", {
    headers: { Authorization: `Bearer ${tokenData.access_token}` },
  });

  const userInfo = (await userRes.json()) as {
    ok: boolean;
    sub?: string;
    name?: string;
    email?: string;
    picture?: string;
  };

  if (!userInfo.ok) {
    return new Response(null, {
      status: 302,
      headers: { Location: "/login?error=token_error", "Set-Cookie": clearCookie },
    });
  }

  const sessionCookie = await createSessionCookie(
    {
      id: userInfo.sub ?? "",
      name: userInfo.name ?? "",
      email: userInfo.email ?? "",
      avatarUrl: userInfo.picture ?? "",
    },
    env.SESSION_SECRET
  );

  return new Response(null, {
    status: 302,
    headers: new Headers([
      ["Location", "/"],
      ["Set-Cookie", clearCookie],
      ["Set-Cookie", sessionCookie],
    ]),
  });
}
