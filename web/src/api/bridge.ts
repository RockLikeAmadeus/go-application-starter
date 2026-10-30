/** mobile.Reply: the Go handler's answer to one bridged request. */
export type BridgeReply = { status: number; contentType: string; body: string };

/** Calls mobile.Serve over the Wails bridge. */
export type BridgeCall = (method: string, path: string, body: string) => Promise<BridgeReply>;

/**
 * A fetch that sends each request over the Wails bridge through call. On
 * Android, Wails' asset path turns every request into a GET with no body, so
 * the mobile app's API requests go over the bridge instead (see mobile.Serve).
 */
export function bridgeFetch(call: BridgeCall): (request: Request) => Promise<Response> {
  return async (request) => {
    const { pathname, search } = new URL(request.url);
    const reply = await call(request.method, pathname + search, await request.text());
    // Go sends an empty body for 204, but a 204 Response can't have one at all.
    const body = reply.status === 204 ? null : reply.body;
    return new Response(body, {
      status: reply.status,
      headers: { "Content-Type": reply.contentType },
    });
  };
}

/** The global object, as far as finding Android's Wails bridge goes. */
export type BridgeWindow = { wails?: { invokeAsync?: unknown } };

/**
 * The fetch for the API client: over the bridge on Android, else plain fetch.
 * It looks for Android's bridge once, now, because importing the Wails runtime
 * (loadCall) replaces window.wails. Plain fetch is looked up per request, so
 * tests can stub it after the API client loads.
 */
export function apiFetch(
  win: BridgeWindow,
  loadCall: () => Promise<BridgeCall>,
): (request: Request) => Promise<Response> {
  if (!win.wails?.invokeAsync) {
    return (request) => globalThis.fetch(request);
  }
  return async (request) => bridgeFetch(await loadCall())(request);
}
