import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { api, ApiError, unwrap } from "../../api/client";
import type { components } from "../../api/schema.gen";

type SignInMethod = components["schemas"]["SignInMethod"];

const primaryButtonClasses =
  "type-label inline-block rounded-control bg-accent px-4 py-2 font-medium text-accent-fg hover:opacity-90 terminal:text-sm soft:px-5";
const secondaryButtonClasses =
  "type-label inline-block rounded-control border border-line px-4 py-2 font-medium hover:bg-sunken terminal:text-sm soft:border-transparent soft:px-5";

/**
 * What someone sees while signed out of a server with sign-in on: no sidebar
 * and no page. With the `page` method, links to `/auth/sign-in` and
 * `/auth/sign-up`, which hand them to the server's sign-in. With `browser`
 * (the desktop and mobile apps in remote mode), buttons that ask the app to
 * open the system browser, and a waiting screen until that finishes.
 */
export function SignInScreen({ signInMethod }: { signInMethod: SignInMethod }) {
  if (signInMethod === "browser") {
    return <BrowserSignInScreen />;
  }
  return (
    <SignInPanel heading="Sign in">
      <p className="text-muted">Please sign in or create an account to continue.</p>
      <div className="flex flex-wrap gap-3">
        <a href="/auth/sign-in" className={primaryButtonClasses}>
          Sign in
        </a>
        <a href="/auth/sign-up" className={secondaryButtonClasses}>
          Create an account
        </a>
      </div>
    </SignInPanel>
  );
}

function SignInPanel({ heading, children }: { heading: string; children: React.ReactNode }) {
  return (
    <main className="flex min-h-screen items-center justify-center px-4 py-8">
      <section className="w-full max-w-sm space-y-6 rounded-panel border border-line bg-panel p-8 terminal:space-y-4 terminal:p-5 soft:border-transparent">
        <h1 className="text-2xl font-semibold tracking-tight soft:text-3xl soft:font-bold">
          {heading}
        </h1>
        {children}
      </section>
    </main>
  );
}

function BrowserSignInScreen() {
  const queryClient = useQueryClient();
  const [message, setMessage] = useState<{ text: string; isFailure: boolean }>();
  const [isSignUp, setIsSignUp] = useState(false);
  const signIn = useMutation({
    mutationFn: async (isSignUp: boolean) =>
      unwrap(await api.POST("/api/v1/browser-sign-in", { body: { isSignUp } })),
    onMutate: (isSignUp) => {
      setIsSignUp(isSignUp);
      setMessage(undefined);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["session"] }),
    onError: (error, isSignUp) => {
      const action = isSignUp ? "Creating your account" : "Sign-in";
      if (error instanceof ApiError && error.problem.type === "cancelled") {
        setMessage({ text: `${action} was cancelled.`, isFailure: false });
        return;
      }
      const failed = isSignUp ? "Creating your account failed." : "Sign-in failed.";
      const detail = error instanceof ApiError ? error.problem.detail : undefined;
      setMessage({ text: `${failed} ${detail ?? "Please try again."}`, isFailure: true });
    },
  });
  const cancel = useMutation({
    mutationFn: async () => unwrap(await api.DELETE("/api/v1/browser-sign-in")),
  });

  if (signIn.isPending) {
    return (
      <SignInPanel heading={isSignUp ? "Creating your account" : "Signing in"}>
        <p className="text-muted">
          {isSignUp
            ? "Please finish creating your account in your browser. The application will continue once you have finished."
            : "Please finish signing in in your browser. The application will continue once you have signed in."}
        </p>
        <button type="button" className={secondaryButtonClasses} onClick={() => cancel.mutate()}>
          Cancel
        </button>
      </SignInPanel>
    );
  }
  return (
    <SignInPanel heading="Sign in">
      {message !== undefined && (
        <p
          role={message.isFailure ? "alert" : "status"}
          className={message.isFailure ? "text-sm text-danger" : "text-sm text-muted"}
        >
          {message.text}
        </p>
      )}
      <p className="text-muted">Please sign in or create an account to continue.</p>
      <div className="flex flex-wrap gap-3">
        <button type="button" className={primaryButtonClasses} onClick={() => signIn.mutate(false)}>
          Sign in
        </button>
        <button
          type="button"
          className={secondaryButtonClasses}
          onClick={() => signIn.mutate(true)}
        >
          Create an account
        </button>
      </div>
      <p className="text-sm text-muted">
        Your browser will open so that you can sign in or create an account.
      </p>
    </SignInPanel>
  );
}
