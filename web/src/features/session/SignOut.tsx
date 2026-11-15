import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, unwrap } from "../../api/client";
import type { components } from "../../api/schema.gen";

/**
 * Who is signed in, and a button that ends their session. With the `browser`
 * method (the apps in remote mode) it calls `DELETE /api/v1/session` and
 * refetches the session, as the webview can't follow a redirect. Otherwise it is a plain form
 * posting to `/auth/sign-out`: the server ends the session, clears the cookie,
 * and sends the browser back to the app, which then asks them to sign in.
 */
export function SignOut({
  name,
  signInMethod,
}: {
  name: string;
  signInMethod: components["schemas"]["SignInMethod"];
}) {
  const queryClient = useQueryClient();
  const signOut = useMutation({
    mutationFn: async () => unwrap(await api.DELETE("/api/v1/session")),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["session"] }),
  });
  if (signInMethod === "browser") {
    return (
      <div className="flex flex-col gap-2 border-t border-sidebar-line px-2 pt-4 soft:border-t-0">
        <p className="text-sm text-sidebar-muted">Signed in as {name}.</p>
        <button
          type="button"
          onClick={() => signOut.mutate()}
          className="type-label self-start rounded-control border border-sidebar-line px-3 py-1.5 text-sm text-sidebar-muted hover:text-sidebar-fg terminal:text-xs"
        >
          Sign out
        </button>
      </div>
    );
  }
  return (
    <form
      method="post"
      action="/auth/sign-out"
      className="flex flex-col gap-2 border-t border-sidebar-line px-2 pt-4 soft:border-t-0"
    >
      <p className="text-sm text-sidebar-muted">Signed in as {name}.</p>
      <button
        type="submit"
        className="type-label self-start rounded-control border border-sidebar-line px-3 py-1.5 text-sm text-sidebar-muted hover:text-sidebar-fg terminal:text-xs"
      >
        Sign out
      </button>
    </form>
  );
}
