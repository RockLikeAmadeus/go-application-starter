/**
 * Who is signed in, and a button that ends their session. It is a plain form
 * posting to `/auth/sign-out`: the server ends the session, clears the cookie,
 * and sends the browser back to the app, which then asks them to sign in.
 */
export function SignOutForm({ name }: { name: string }) {
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
