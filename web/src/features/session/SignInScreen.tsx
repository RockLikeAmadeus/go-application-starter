/**
 * What someone sees while signed out of a server with sign-in on: no sidebar
 * and no page, only a link to `/auth/sign-in`, which hands them to the
 * server's sign-in.
 */
export function SignInScreen() {
  return (
    <main className="flex min-h-screen items-center justify-center px-4 py-8">
      <section className="w-full max-w-sm space-y-6 rounded-panel border border-line bg-panel p-8 terminal:space-y-4 terminal:p-5 soft:border-transparent">
        <h1 className="text-2xl font-semibold tracking-tight soft:text-3xl soft:font-bold">
          Sign in
        </h1>
        <p className="text-muted">Please sign in to continue.</p>
        <a
          href="/auth/sign-in"
          className="type-label inline-block rounded-control bg-accent px-4 py-2 font-medium text-accent-fg hover:opacity-90 terminal:text-sm soft:px-5"
        >
          Sign in
        </a>
      </section>
    </main>
  );
}
