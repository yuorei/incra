import { redirect } from "react-router";
import type { Route } from "./+types/logout";
import { clearSessionCookie } from "../lib/session";

export async function action({}: Route.ActionArgs) {
  return new Response(null, {
    status: 302,
    headers: {
      Location: "/login",
      "Set-Cookie": clearSessionCookie(),
    },
  });
}

export async function loader({}: Route.LoaderArgs) {
  throw redirect("/");
}
