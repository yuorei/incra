// Extends the auto-generated Env interface from worker-configuration.d.ts
// Secrets are set via `wrangler secret put` and not stored in wrangler.jsonc
interface Env {
  SLACK_CLIENT_ID: string;
  SLACK_CLIENT_SECRET: string; // wrangler secret put SLACK_CLIENT_SECRET
  SESSION_SECRET: string; // wrangler secret put SESSION_SECRET
}
