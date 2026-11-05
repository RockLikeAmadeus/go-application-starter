import { vi } from "vitest";
import type { components } from "../api/schema.gen";

type Feature = components["schemas"]["Feature"];

/**
 * A fake of the API for global fetch: GET /api/v1/entitlements answers a plan
 * with these features, by key, and anything else is a 404.
 */
export function entitlementsApi(features: Record<string, Feature>) {
  return vi.fn(async (input: Request) =>
    new URL(input.url).pathname === "/api/v1/entitlements"
      ? new Response(JSON.stringify({ plan: "free", limits: {}, features }), {
          headers: { "Content-Type": "application/json" },
        })
      : new Response(null, { status: 404 }),
  );
}
