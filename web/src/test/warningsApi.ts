import { vi } from "vitest";

/**
 * A fake of the API for global fetch: GET /api/v1/warnings answers these
 * messages, in order, and anything else is a 404.
 */
export function warningsApi(...messages: string[]) {
  return vi.fn(async (input: Request) =>
    new URL(input.url).pathname === "/api/v1/warnings"
      ? new Response(JSON.stringify({ warnings: messages.map((message) => ({ message })) }), {
          headers: { "Content-Type": "application/json" },
        })
      : new Response(null, { status: 404 }),
  );
}
