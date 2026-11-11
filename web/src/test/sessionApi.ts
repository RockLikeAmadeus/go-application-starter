import { vi } from "vitest";
import type { components } from "../api/schema.gen";

type Session = components["schemas"]["Session"];

/**
 * A fake of the API for global fetch: GET /api/v1/session answers this
 * session, and anything else is a 404.
 */
export function sessionApi(session: Session) {
  return vi.fn(async (input: Request) =>
    new URL(input.url).pathname === "/api/v1/session"
      ? new Response(JSON.stringify(session), {
          headers: { "Content-Type": "application/json" },
        })
      : new Response(null, { status: 404 }),
  );
}
