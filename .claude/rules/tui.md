---
paths:
  - "internal/adapters/tui/**"
  - "test/terminal/**"
---

# TUI conventions

- Vim navigation is required (see SPEC.md §5). Every list or selectable view binds `j`/`k` alongside `down`/`up`. Bind them only while browsing, never while a text input has focus, so users can still type those letters.
- Cover the `j`/`k` bindings in each new view's tests.
- Shared colours, styles, and the helpers every screen uses (`errorLine`, `selectedRow`) live in `internal/adapters/tui/styles.go`, outside sample blocks. Give a style a small helper with its own test (`styles_test.go`), so it stays in use, and passes golangci-lint's `unused` check, when the sample is removed.
- Sign-in (`sign_in.go`, `docs/sign-in.md`): with `Ports.SignIn` set (a server with sign-in on), `signInGate` wraps the first screen. No session shows the "Signed out" screen (`l` signs in through `clientsignin`, with a waiting screen and Esc to cancel); signed in, a screen implementing `signOutScreen` gets the "Signed in to <url> as <name>." line and `o` to sign out. Both keys only while no text input has focus. A message implementing `failedMsg` whose error is `ErrNoWorkspace` brings back the signed-out screen. No keys or text on local data or with sign-in off.
<!-- sample:start -->
- The Items screen (`internal/adapters/tui/items.go`) and its tests (`items_test.go`) show both.
<!-- sample:end -->
