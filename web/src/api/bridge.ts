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
    return new Response(reply.body, {
      status: reply.status,
      headers: { "Content-Type": reply.contentType },
    });
  };
}
