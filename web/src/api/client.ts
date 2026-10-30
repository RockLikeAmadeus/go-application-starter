import createClient from "openapi-fetch";
import type { components, paths } from "./schema.gen";

// sample:start
export type Item = components["schemas"]["Item"];
// sample:end
export type Problem = components["schemas"]["Problem"];

/** Typed client generated from api/openapi.yaml. Same-origin in prod; proxied by Vite in dev. */
export const api = createClient<paths>({
  baseUrl: globalThis.location?.origin ?? "",
  // Resolve fetch per call so tests can stub it after this module loads.
  fetch: (request) => globalThis.fetch(request),
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
