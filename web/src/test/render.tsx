import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, renderHook } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/** Renders UI with a fresh, retry-free QueryClient. */
export function renderWithClient(ui: ReactElement) {
  const client = newClient();
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

/** Renders a hook with a fresh, retry-free QueryClient. */
export function renderHookWithClient<T>(hook: () => T) {
  const client = newClient();
  return renderHook(hook, {
    wrapper: ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    ),
  });
}
