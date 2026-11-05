import { vi } from "vitest";

/**
 * A fake of the API for global fetch: GET /api/v1/connection says whether the
 * data needs a connection, and anything else is a 404.
 */
export function connectionApi(isRequired: boolean) {
  return vi.fn(async (input: Request) =>
    new URL(input.url).pathname === "/api/v1/connection"
      ? new Response(JSON.stringify({ isRequired }), {
          headers: { "Content-Type": "application/json" },
        })
      : new Response(null, { status: 404 }),
  );
}
