---
paths:
  - "web/e2e/**"
  - "web/playwright.config.ts"
---

# Browser e2e (playwright-bdd)

- `task e2e` builds the SPA and starts the Go server itself.
- `web/e2e/features/` holds only a few journeys that prove the wiring; business scenarios belong in `features/`.
- Scenarios run in parallel against one server, so use unique names and don't assert on whole lists.
- Tag phone browser scenarios `@phone`. They run on an emulated Pixel (Chromium), an emulated iPhone (WebKit, the local stand-in for Safari) and a phone-sized Firefox window. Untagged scenarios run on desktop Chromium.
- Playwright's WebKit (the `iphone` project) needs system libraries that need sudo. `task setup` installs them with `--with-deps`. If it was skipped, run `sudo pnpm --dir web exec playwright install-deps webkit` in a real terminal: Claude Code's `!` can't answer sudo's password prompt.
- Playwright's WebKit doesn't show WebKitGTK's `@starting-style` bug (see `.claude/rules/web-styles.md`).
