import createClient from "openapi-fetch";
import { apiFetch, type BridgeCall, type BridgeWindow } from "./bridge";
import type { components, paths } from "./schema.gen";

// sample:start
export type Item = components["schemas"]["Item"];
// sample:end
export type Problem = components["schemas"]["Problem"];

/**
 * Loads the Wails runtime, only on Android, and calls mobile.Serve through the
 * service cmd/mobile binds. The import is dynamic so the browser never
 * downloads the runtime.
 */
async function loadBridgeCall(): Promise<BridgeCall> {
  const { Call } = await import("@wailsio/runtime");
  return (method, path, body) => Call.ByName("main.Bridge.Serve", method, path, body);
}

/**
 * Typed client generated from api/openapi.yaml. Same-origin in prod; proxied by
 * Vite in dev; over the Wails bridge in the Android app (see apiFetch).
 */
export const api = createClient<paths>({
  baseUrl: globalThis.location?.origin ?? "",
  // Only Android's Wails window has `wails`, so globalThis doesn't declare it.
  fetch: apiFetch(globalThis as BridgeWindow, loadBridgeCall),
});

/** Error thrown for non-2xx responses, carrying the server's problem document. */
export class ApiError extends Error {
  readonly problem: Problem;

  constructor(problem: Problem) {
    super(problem.detail ?? problem.title);
    this.name = "ApiError";
    this.problem = problem;
  }
}

/** Unwraps an openapi-fetch result, throwing ApiError on failure. */
export function unwrap<T>(result: { data?: T; error?: Problem; response: Response }): T {
  if (result.error !== undefined || !result.response.ok) {
    throw new ApiError(
      result.error ?? { title: result.response.statusText, status: result.response.status },
    );
  }
  return result.data as T;
}
