import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, renderHook } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";

function newClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

/**
 * Renders UI with a fresh, retry-free QueryClient, returned beside the
 * render result so a test can wait for its queries to settle.
 */
export function renderWithClient(ui: ReactElement) {
  const client = newClient();
  return { ...render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>), client };
}

/**
 * Renders a hook with a fresh, retry-free QueryClient, or with the one given
 * (to mount a second component on the same cache), returned beside the
 * render result.
 */
export function renderHookWithClient<T>(hook: () => T, client = newClient()) {
  return {
    ...renderHook(hook, {
      wrapper: ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      ),
    }),
    client,
  };
}
