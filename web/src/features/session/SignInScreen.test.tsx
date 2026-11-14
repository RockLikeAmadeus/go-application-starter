import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { renderWithClient } from "../../test/render";
import { SignInScreen } from "./SignInScreen";

type Call = { method: string; path: string; body: string };

/**
 * A fake of the in-process handler for global fetch. POST /api/v1/browser-sign-in
 * waits until `finish` is called with the response to answer; everything else
 * answers 204.
 */
function browserSignInApi() {
  const calls: Call[] = [];
  let answer: (response: Response) => void = () => {};
  const fetch = vi.fn(async (input: Request) => {
    const path = new URL(input.url).pathname;
    calls.push({ method: input.method, path, body: await input.clone().text() });
    if (input.method === "POST" && path === "/api/v1/browser-sign-in") {
      return new Promise<Response>((resolve) => {
        answer = resolve;
      });
    }
    if (path === "/api/v1/session") {
      return Response.json({ isSignInOn: true, name: "Alpha", signInMethod: "browser" });
    }
    return new Response(null, { status: 204 });
  });
  return {
    fetch,
    calls,
    finish: (response: Response) => answer(response),
  };
}

function problem(status: number, type: string, detail?: string) {
  return new Response(JSON.stringify({ type, title: "Problem", status, detail }), {
    status,
    headers: { "Content-Type": "application/problem+json" },
  });
}

describe("SignInScreen in browser mode", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("shows page-mode links and no browser note", () => {
    renderWithClient(<SignInScreen signInMethod="page" />);

    expect(screen.getByRole("link", { name: "Sign in" })).toHaveAttribute("href", "/auth/sign-in");
    expect(screen.queryByText(/Your browser will open/)).not.toBeInTheDocument();
  });

  it("offers buttons and a note instead of links", () => {
    renderWithClient(<SignInScreen signInMethod="browser" />);

    expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create an account" })).toBeInTheDocument();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    expect(
      screen.getByText("Your browser will open so that you can sign in or create an account."),
    ).toBeInTheDocument();
  });

  it("starts a browser sign-in and shows the waiting screen with Cancel", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    renderWithClient(<SignInScreen signInMethod="browser" />);

    await userEvent.setup().click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("heading", { name: "Signing in" })).toBeInTheDocument();
    expect(
      screen.getByText(
        "Please finish signing in in your browser. The application will continue once you have signed in.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
    expect(api.calls[0]).toMatchObject({ method: "POST", path: "/api/v1/browser-sign-in" });
    expect(JSON.parse(api.calls[0]?.body ?? "")).toEqual({ isSignUp: false });
  });

  it("starts creating an account and says so on the waiting screen", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    renderWithClient(<SignInScreen signInMethod="browser" />);

    await userEvent.setup().click(screen.getByRole("button", { name: "Create an account" }));

    expect(
      await screen.findByRole("heading", { name: "Creating your account" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        "Please finish creating your account in your browser. The application will continue once you have finished.",
      ),
    ).toBeInTheDocument();
    expect(JSON.parse(api.calls[0]?.body ?? "")).toEqual({ isSignUp: true });
  });

  it("refetches the session when the sign-in succeeds", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    const { client } = renderWithClient(<SignInScreen signInMethod="browser" />);
    const invalidate = vi.spyOn(client, "invalidateQueries");
    await userEvent.setup().click(screen.getByRole("button", { name: "Sign in" }));
    await screen.findByRole("heading", { name: "Signing in" });

    api.finish(Response.json({ isSignInOn: true, name: "Alpha", signInMethod: "browser" }));

    await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: ["session"] }));
  });

  it("cancels through DELETE and shows the cancelled message above the usual text", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    renderWithClient(<SignInScreen signInMethod="browser" />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Sign in" }));
    await user.click(await screen.findByRole("button", { name: "Cancel" }));

    expect(api.calls.some((c) => c.method === "DELETE")).toBe(true);
    api.finish(problem(409, "cancelled", "the sign-in was cancelled"));

    const cancelled = await screen.findByText("Sign-in was cancelled.");
    expect(cancelled).toHaveAttribute("role", "status");
    expect(cancelled).toHaveClass("text-muted");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(
      screen.getByText("Please sign in or create an account to continue."),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
  });

  it("says creating the account was cancelled after cancelling that", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    renderWithClient(<SignInScreen signInMethod="browser" />);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "Create an account" }));
    await user.click(await screen.findByRole("button", { name: "Cancel" }));
    api.finish(problem(409, "cancelled", "the sign-in was cancelled"));

    expect(await screen.findByText("Creating your account was cancelled.")).toBeInTheDocument();
  });

  it("shows a failure with the server's detail after the sentence", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    renderWithClient(<SignInScreen signInMethod="browser" />);
    await userEvent.setup().click(screen.getByRole("button", { name: "Sign in" }));
    await screen.findByRole("heading", { name: "Signing in" });

    api.finish(problem(500, "internal", "An unexpected error occurred on the server."));

    expect(
      await screen.findByText("Signing in failed. An unexpected error occurred on the server."),
    ).toBeInTheDocument();
  });

  it("ends a failure with no detail with Please try again", async () => {
    const api = browserSignInApi();
    vi.stubGlobal("fetch", api.fetch);
    renderWithClient(<SignInScreen signInMethod="browser" />);
    await userEvent.setup().click(screen.getByRole("button", { name: "Create an account" }));
    await screen.findByRole("heading", { name: "Creating your account" });

    api.finish(new Response(null, { status: 500 }));

    expect(
      await screen.findByText("Creating your account failed. Please try again."),
    ).toBeInTheDocument();
  });
});
