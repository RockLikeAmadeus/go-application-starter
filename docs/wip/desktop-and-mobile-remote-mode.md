# Use the desktop and mobile apps on a hosted server (remote mode)

Step 10 of [the editions plan](../plans/editions-and-accounts.md), and the TODO item "Desktop and mobile remote mode".

- Spec: [SPEC.md §4.20](../../SPEC.md#420-use-the-desktop-and-mobile-apps-on-a-hosted-server)
- Scenarios: [`features/remote-mode.feature`](../../features/remote-mode.feature) (`@remote-mode`, on the http and oidc drivers; 10 scenarios, 8 of them outlines over the desktop and mobile apps, each tagged `@wip`)
- Decisions: [ADR 4](../adr/0004-native-shells.md) (the window and the in-process handler stay), [ADR 9](../adr/0009-sign-in-and-sessions.md) (native sign-in, bearer tokens), and the plan's step 10 notes

## Decisions

Agreed with the user (2026-11-13):
- **Separate sessions.** The apps keep their own session token for a server, apart from the command line's (GitHub Desktop and `gh`, not Docker Desktop and `docker`). So `starter login` doesn't sign the desktop app in, and signing out of either leaves the other alone. `tokenstore` gains the program's name: the command line keeps today's keyring service and `tokens.json`, and the apps use their own service name and file (the names are picked in task 1), in the same settings folder.
- **The web app shows a waiting screen.** In remote mode, the web app's sign-in screen starts a browser sign-in through the in-process handler and shows that it is waiting for the browser, with a Cancel button, in its own styles and themes, like the TUI's waiting screen. The hosted web app never shows it.
- **A built-in server address.** A build can set the server address (`-ldflags -X …/internal/config.BuiltInServerURL=…`, like `bootstrap.LocalPlan`). It is the lowest layer of `server_url` (ADR 6), so a setting or `STARTER_SERVER_URL` still wins, and the CLI and TUI use it too. It is how the Android app goes remote, as nobody can edit its settings. An empty setting doesn't override it, so a build with one has no local mode. This takes the plan's follow-up "A server address built into the program" into this step.

Decided in planning:
- **Reuse step 9's client sign-in.** The apps sign in through `clientsignin.Client`, `browsersignin` (system browser, loopback port, PKCE), the server's `/auth/native` routes, and `httpclient` sending the token as a bearer token. Nothing new on the hosted server, except sign-up through `/auth/native` (below). The remote core is `bootstrap.NewRemote`, as for the CLI.
- **Remote mode is chosen in `desktop.Open`,** and so in `mobile.Open`: with a server set (setting, environment, or built in), it puts `bootstrap.NewRemote` behind the same in-process handler, opens no data folder or storage (so no lock, and the `storage` setting is ignored), and leaves `IsDataLocal` unset, so the offline notice comes back. Otherwise it opens local data as today. The desktop app shares `server_url` with the command line through the settings file they already share (§4.12). A change takes effect at the next opening.
- **`bootstrap.NewClientSignIn` takes the settings folder and the program,** rather than calling `os.UserConfigDir` itself, so the desktop app (its settings folder) and the Android app (its private storage) reuse it. `NewRemote`'s token source does the same, so the app's requests send the app's token.
- **The in-process handler signs in through an interface `httpapi` declares** (`Options.ClientSignIn`, satisfied by a small adapter over `*clientsignin.Client` in bootstrap), so `httpapi` imports no client adapter. With it set:
  - `GET /api/v1/session` answers from the hosted server's session, through the client sign-in.
  - `POST /api/v1/browser-sign-in` (body: whether to create an account) opens the system browser, waits until the sign-in finishes, and answers the session. A failure or a cancel is a problem response. It blocks rather than being polled: the Android bridge has no request cancelling, so cancelling is its own call.
  - `DELETE /api/v1/browser-sign-in` cancels the sign-in in progress. One sign-in at a time: starting another cancels the first.
  - `POST /auth/sign-out` signs out through the client sign-in (forgetting the token even when the server can't be reached) and redirects to `/`, so the web app's sign-out form doesn't change.
  - The hosted server, with no client sign-in, answers both `browser-sign-in` operations 404.
- **The session says how to sign in.** `Session` in `api/openapi.yaml` gains `signInMethod`: `page` (the window goes to `/auth/sign-in`, as on the hosted web app) or `browser` (the app opens the system browser and waits). The server answers `page`; the in-process handler in remote mode answers `browser`.
- **Creating an account through the browser.** `browsersignin.Start` takes whether to sign up, which adds `sign-up=1` to the `/auth/native` address, and `/auth/native` then sends the browser to `/auth/sign-up` instead of `/auth/sign-in`. `clientsignin` gains `SignUp` beside `SignIn`. The command line could use it later (not in this feature).
- **Cancelling lives in `clientsignin`:** the client holds the sign-in in progress and cancels its context, so no token is saved. The TUI keeps cancelling through its own context.
- **Opening the browser is passed in.** `desktop.Open` and `mobile.Open` take the opener. `cmd/desktop` and `cmd/mobile` pass Wails' `Browser.OpenURL`, which works on each operating system, including Android, where `browsersignin.Open` (xdg-open and others) doesn't. Tests pass an opener that drives the test browser.
- **The Android app keeps its token in a file in its private storage** (`credential_store = file`, set by `mobile.Open`). Android has no Secret Service, and go-keyring's Linux code would try D-Bus first. Only the app can read its private storage.
- **The window doesn't say where the token is kept.** The command line says so when it falls back to a file; the apps don't, and `docs/sign-in.md` explains it. This is a possible follow-up.
- **No browser journey (Playwright).** Playwright runs the web app against the server, which never answers `browser`. The waiting screen is checked by Vitest, and the real windows by hand (task 22).
- **Acceptance steps:** the apps are opened with `desktop.Open` and `mobile.Open` on a temporary settings folder, with `STARTER_SERVER_URL` set to the driver's server, the file token store, and an opener that finishes the sign-in in the test browser (`newBrowsersOverHTTP`). The mobile app's requests go over the bridge, as in the mobile driver. The command line "on this computer" is `bootstrap.NewClientSignIn` on the same folder, unlike `command_line_steps.go`'s in-memory store, so the separate-session scenarios mean something. The built-in address scenario sets `config.BuiltInServerURL` for the scenario and restores it after.
- **Template first:** all of this is template code. The sample has one thin scenario outline (an item added in an app is in the web app), in a sample block.
- **User-facing text to agree in its cycle:** the waiting screen and its Cancel button, the sign-in screen's buttons in browser mode, a failed or cancelled sign-in, a server with sign-in off (if shown), and the `credential_store` note in the default settings file.

## Tasks

Each is one `/red` → `/green` → `/refactor` cycle.

Keeping the apps' tokens apart:
- [x] 1. `tokenstore`: a token saved under one program's name isn't read under another's, in both the file store and the system store (the keyring service name). The command line's name keeps today's keyring service and `tokens.json`. Pick the apps' names.
  - Red: `TestAppStores_KeepTheirTokensApartFromTheCommandLines` (app Read found the command line's token). Green: `File` takes a file name, `Keyring` a service name; `NewAppFile` (`app-tokens.json`) and `NewAppKeyring` (`go-application-starter-app`); the command line's names are unchanged. Refactor: nothing worth changing. Task 2 passes the apps' stores through bootstrap (`System` gains an app variant there).
- [x] 2. `bootstrap.NewClientSignIn` and `NewRemote` take the settings folder and the program: the apps' client keeps its token apart from the command line's. The CLI and TUI pass their folder and name, with no change in behaviour.
  - Red: `TestNewClientSignInFor_SendsTheProgramsOwnTokenFromASharedFolder` (both programs sent the command line's token). Green: `Program` (`CommandLine`, `DesktopOrMobileApp`), `NewClientSignInFor(cfg, folder, program)`, `NewRemote` takes the program, `tokenstore.NewAppSystem`. Refactor: nothing to change. `NewClientSignIn(cfg)` stays as the command line's wrapper, so the CLI and TUI factories and their tests are untouched.

Creating an account through the browser:
- [x] 3. `browsersignin.Start` for a sign-up opens the `/auth/native` address with `sign-up=1`.
  - Red: `TestStart_ForASignUpAsksTheNativeSignInAddressForSignUp` (no `sign-up` in the address). Green: `Start` takes `isSignUp` and adds `sign-up=1`; callers pass `false`. Refactor: nothing to change.
- [x] 4. `httpapi`: `/auth/native` with `sign-up=1` sends a signed-out browser to `/auth/sign-up`, returning to `/auth/native/finish`.
  - Red: `TestANativeSignInRequestToSignUpWithNoSessionIsSentToSignUp` redirected to `/auth/sign-in`. Green: `startNativeSignIn` picks the page from `sign-up`. Refactor: nothing worth changing.
- [x] 5. `clientsignin.SignUp` signs in through the browser's sign-up, and keeps the token as `SignIn` does.
  - Red: `TestSignUp_SignsUpThroughTheBrowserAndKeepsTheToken` failed (stub). Green: `BrowserSignIn` takes `isSignUp`; `SignIn` and `SignUp` share one helper. Refactor: nothing more.

Cancelling:
- [x] 6. `clientsignin`: cancelling the sign-in in progress makes it return a cancelled error and keeps no token. Red: `TestCancel_EndsTheSignInInProgressAndKeepsNoToken` with an empty `Cancel`, failing "SignIn was still waiting 2s after Cancel"; green: `Client.Cancel` cancels the sign-in's context with cause `ErrCancelled`. Refactor: nothing worth changing.
- [ ] 6b. `clientsignin`: starting another sign-in cancels the first.

The built-in server address:
- [ ] 7. `config`: `BuiltInServerURL` is `server_url`'s built-in value, below the settings file and `STARTER_SERVER_URL`; an empty setting doesn't override it. Add a Taskfile variable for the desktop and Android builds, and a step for template users under the starter TODO item.

The API:
- [ ] 8. `api/openapi.yaml`: `Session` gains `signInMethod` (`page`, `browser`); the server answers `page`. `task gen`.
- [ ] 9. `api/openapi.yaml`: `POST` and `DELETE /api/v1/browser-sign-in`; a server with no client sign-in answers both 404. `task gen`.
- [ ] 10. `httpapi`: with `Options.ClientSignIn`, `GET /api/v1/session` answers the hosted server's session (name, sign-in on), with `signInMethod: browser`.
- [ ] 11. `httpapi`: `POST /api/v1/browser-sign-in` signs in (or up) through the client sign-in, handing it the opener, and answers the session; a failure is a problem response.
- [ ] 12. `httpapi`: `DELETE /api/v1/browser-sign-in` cancels the sign-in in progress, and the waiting `POST` answers a cancelled problem.
- [ ] 13. `httpapi`: with `Options.ClientSignIn`, `POST /auth/sign-out` signs out through it and redirects to `/`, also when the server can't end the session.

The apps:
- [ ] 14. `desktop.Open` with a server set opens remote mode: the core is `NewRemote` (no data folder is opened, nothing locked), and `GET /api/v1/connection` says a connection is required. Acceptance: define the steps for opening the apps on the server; "An app can't see anything without signing in" passes for the desktop app.
- [ ] 15. `desktop.Open` takes the opener and wires the client sign-in into the handler (the app's own token store, in its settings folder). Acceptance: signing in, the browser offer, signing out, and the sample scenario pass for the desktop app.
- [ ] 16. `mobile.Open` in remote mode: through `desktop`'s, with the file token store in private storage, and the opener. Acceptance: the same scenarios pass for the mobile app.
- [ ] 17. Acceptance: creating an account and cancelling a sign-in pass for both apps (steps only, if tasks 5, 6, 11, and 12 hold).
- [ ] 18. Acceptance: the command line "on this computer" through `bootstrap.NewClientSignIn` on the apps' settings folder; the two separate-session scenarios pass.
- [ ] 19. Acceptance: the built-in address scenario passes for both apps (`config.BuiltInServerURL` set for the scenario).
- [ ] 20. `cmd/desktop` and `cmd/mobile` pass Wails' `Browser.OpenURL`. `task build:desktop`, and the Android build, compile. (No unit test: wiring only.)

The web app:
- [ ] 21. `useSession` reads `signInMethod`; with `browser`, the sign-in screen's sign-in and create-account controls start `POST /api/v1/browser-sign-in` instead of following links, show the waiting screen with Cancel (`DELETE`), refetch on success, and show the error on failure. One Vitest test per behaviour. Agree the wording first.

Finish:
- [ ] 22. By hand: the desktop app against `task dev` (dev sign-in), signing in, cancelling, creating an account, and signing out, and the offline notice with the server stopped; the Android app built with the built-in address of a server the phone reaches, signing in through the phone's browser and coming back to the app (the loopback reply on Android is the risk to check).
- [ ] 23. Docs and finish: ADR 4 (remote mode built), ADR 9 (the apps' sign-in, separate sessions, `signInMethod`, browser sign-in routes), `docs/sign-in.md` (the apps, sign-up through `/auth/native`, where the apps keep tokens), `docs/configuration.md` (the built-in address; the `credential_store` row naming the apps), the `credential_store` note in `config.DefaultSettingsFileText` (wording agreed), `docs/mobile.md`, `docs/data-locations.md`, `.claude/rules/` (`desktop.md`, `mobile.md`, `server.md`, `web.md`, `configuration.md`), CLAUDE.md (the editions list, the architecture lines for `httpclient` and `clientsignin`), and README if it mentions remote mode. Remove the `@wip` tags once every scenario passes, and delete this file. Mark step 10 done in the plan, with a summary and links. If step 5's by-hand check is still open, the plan stays; otherwise it was the last step, so move the plan's open follow-ups (without the built-in address, done here) into `TODO.md`, tick the "Hosted and local editions" and "Desktop and mobile remote mode" items, and delete the plan. Run `task test:init`.

## Changes to the plan

- Task 6 split: 6 (cancel) and 6b (a new sign-in cancels the first), one test each.
