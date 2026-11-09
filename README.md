# go-application-starter

A template for multi-front-end applications: a **hexagonal Go core** driven by a **React web app** (through a REST API), a **Cobra CLI**, a **Bubble Tea TUI**, and **desktop and Android apps** that run the same web app and API in-process (Wails), with swappable storage. It is built **BDD outside-in (Godog + playwright-bdd) with a TDD inner loop**, and set up for agentic development with Claude Code.

The app comes in two editions, from the same core ([ADR 5](docs/adr/0005-hosted-and-local-editions.md)):
- **Local** (works today): one user, no sign-in, data on their own device or on a server they host for themselves.
- **Hosted** (planned, see [the plan](docs/plans/editions-and-accounts.md)): many users signing in to their own workspaces, with features and limits set by paid plans.

<!-- sample:start -->
It ships as a working *walking skeleton*: one tiny `Item` domain wired end to end through every layer, tested at every level, and deployable from CI.
<!-- sample:end -->

```
 React SPA ──HTTP──┐
 CLI / TUI ──HTTP──┤ (--server mode)
                   ▼
            httpapi adapter ──► driving ports ◄── CLI / TUI (embedded mode)
                   ▲                   │
 Desktop / Android ┘              core (domain + app)
 (in-process, no port)                 │
                               driven ports
                           ┌──────────┼──────────┐
                        memory    jsonfile    sqlite   (← Postgres planned)

 Local edition:  CLI / TUI embedded, desktop and Android apps, a self-hosted server
 Hosted edition: the web app on a hosted server with sign-in, plus the CLI / TUI and
                 the desktop and Android apps connected to it (all planned)
```

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.27+ | https://go.dev/dl |
| Node | 24 (see `.nvmrc`) | `nvm install` |
| pnpm | 10 | `npm i -g pnpm@10` |
| Task | 3 | `go install github.com/go-task/task/v3/cmd/task@latest` |
| golangci-lint | 2 | https://golangci-lint.run/welcome/install/ |
| lefthook (optional) | 2 | `go install github.com/evilmartians/lefthook/v2@v2.2.2` |
| Docker (optional) | | for `task docker:*` and `task db:up`, usable without `sudo` (on Linux, `sudo usermod -aG docker $USER`, then log in again). Without a test database, Postgres tests skip |
| GitHub CLI (optional) | 2 | only for the Claude Code workflow, which opens a feature's pull request with `gh pr create`. https://cli.github.com, then `gh auth login` |
| Desktop build deps (optional) | | only for the desktop app (`task run:desktop`, `task build:desktop`). Linux (Ubuntu/Debian): `sudo apt install build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev`. macOS: `xcode-select --install`. Windows: nothing extra (Wails uses no cgo there, and the WebView2 runtime ships with Windows 10 and 11; the setup program installs it if missing). Tried on Windows on 2026-11-02 |
| Linux packaging tools (optional) | | only for the desktop app's Linux packages (`task package:desktop:linux`, `task test:package`), on top of the desktop build deps. `sudo apt install desktop-file-utils rpm libfuse2t64`: `desktop-file-validate` checks the desktop file, `rpm` reads the `.rpm` (its checks are skipped without it), and FUSE 2 runs an AppImage (on Ubuntu before 24.04, `libfuse2`). The first AppImage build downloads `linuxdeploy` from GitHub. Build the AppImage where WebKitGTK 4.1 isn't installed (`sudo apt remove libwebkit2gtk-4.1-dev` if nothing else needs it), or it bundles that too, about 90 MB the app never uses |
| Windows packaging tools (optional) | | only for the desktop app's Windows setup program (`task package:desktop:windows`, `task test:package`), built on Linux. `sudo apt install nsis 7zip`: `makensis` builds the setup `.exe`, and 7-Zip checks the files inside it (that check is skipped without it). Cross-compiling the Windows program needs no desktop build deps |
| Android tools (optional) | JDK 17+, Android SDK 35, NDK 26 | only for the mobile app (`task mobile:android:*`). A JDK (`sudo apt install openjdk-21-jdk-headless`), then the SDK command-line tools and `sdkmanager` packages listed in [docs/mobile.md](docs/mobile.md), plus an emulator or a phone with USB debugging |

Then:

```sh
task setup     # Go modules, pnpm install, Playwright browser, git hooks
task check     # codegen drift, lint, unit, acceptance, e2e: everything CI runs
```

## Everyday commands

```sh
task dev                         # API on :8080 + Vite on :5173, opened in your browser; both reload on save
task run:cli -- --help           # CLI, embedded core, data in the desktop app's folder (task run:cli -- info names it; add --server http://localhost:8080 to use a running server)
task run:tui                     # terminal UI, on the same data as the CLI (also accepts --server)
task run:desktop                 # desktop app in a native window, data in your settings folder (needs the desktop build deps)
task mobile:android:run          # Android app on the emulator or a phone, data on the device (needs the Android tools, see docs/mobile.md)
task test                        # unit tests (Go + web)
task accept                      # Godog acceptance tests
task bench                       # Go benchmarks, only when you run them
task e2e                         # browser tests against the real server binary
task e2e:terminal                # terminal journeys: the built CLI and TUI in a terminal (Unix only)
task gen                         # regenerate code after editing api/openapi.yaml
task build                       # ./bin/server (with UI embedded) and ./bin/starter
task build:desktop               # ./bin/starter-desktop (needs the desktop build deps)
task package:desktop:linux       # .deb, .rpm and AppImage of the desktop app in ./bin (VERSION=v1.2.3 for a release)
task package:desktop:windows     # Windows setup .exe of the desktop app in ./bin, built on Linux (needs nsis)
task mobile:android:build        # ./bin/starter.apk, a debug build for phones and the emulator (needs the Android tools)
./bin/server                     # the built app on :8080, installable from Chrome (see docs/installable-app.md)
task docker:up                   # production image locally on :8080
task db:up                       # Postgres for tests on :5432 (task db:down stops it); run Postgres tests with
                                 # STARTER_TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable'
task rename -- --name "New Name" # rebrand the project; --dry-run previews it (see docs/rename.md)
```

Settings come from an optional TOML settings file, overridden by environment variables (see `.env.example`), overridden in turn by command-line options. The CLI, TUI and desktop app share `~/.config/starter/config.toml` (on Linux), which they create, with the defaults, when it's missing; the server reads a file only when `STARTER_CONFIG_FILE` names one. `starter info` shows where each setting came from. Every setting, and the rules, are in [docs/configuration.md](docs/configuration.md) ([ADR 6](docs/adr/0006-layered-configuration.md)). Where each front end keeps its data, on each OS, is in [docs/data-locations.md](docs/data-locations.md).

## Development workflow (with Claude Code)

| Step | Command | Result |
|---|---|---|
| Outer loop | `/feature <desc>` | a `feature/<slug>` branch from `main`, SPEC.md section, `@wip` Gherkin in `features/`, shown failing, TDD task list ending in a docs-and-finish task; committed and pushed |
| Red | `/red <behaviour>` | exactly one failing unit test |
| Green | `/green` | the minimum code to pass, then stop |
| Refactor | `/refactor` | cleaner code, all tests still green |
| One cycle | `/iterate [behaviour]` | red, green and refactor for the next open task, stopping only for questions; then commits and pushes it, if all went cleanly; after the last task, opens the pull request |
| Every open task | `/iterate-all [feature]` | `/iterate` in a loop through a feature's open tasks, each in a fresh subagent with one commit after a short review, pausing only for questions; stops when the tasks are done (opening the pull request) or a cycle can't finish cleanly |

Repeat red → green → refactor until `task accept` is green. The last task drops the `@wip` tags and the `docs/wip/` file. See `CLAUDE.md` for the architecture rules Claude follows, `.claude/rules/` for its topic-specific rules (loaded when it works on matching files), and `.claude/settings.json` for permissions and the auto-format hook.

### Three layers of Gherkin

- `features/*.feature` (**Godog**) holds the business rules in domain language. Each scenario runs against up to six drivers (in-memory core, jsonfile storage, SQLite storage, the full HTTP server+client, the desktop app's wiring, and the mobile app's wiring across its bridge), all in-process and fast.
- `web/e2e/features/*.feature` (**playwright-bdd**) holds a thin set of browser journeys proving the UI → API → core wiring.
- `test/terminal/features/*.feature` (**Godog**, with `creack/pty` and `charmbracelet/x/vt`) holds a thin set of terminal journeys: the built `starter` program, in a real pseudo-terminal, proving the CLI's options, output and exit status, and the TUI's keys and screen, reach the core and back (also through `--server`).

Keep the business rules in the first layer. Browser and terminal journeys are slower and more fragile, so they only prove the wiring.

## CI/CD

GitHub Actions, in `.github/workflows/`. [docs/ci-cd.md](docs/ci-cd.md) explains how they fit together, Dependabot, and how to build the desktop installers on GitHub by hand.

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | PRs, `main`, `v*` tags | Go lint/unit/acceptance/terminal journeys, web lint/typecheck/unit, codegen drift, playwright-bdd e2e. Then builds the Docker image and, on push, publishes it to GHCR (`ghcr.io/<owner>/<repo>:sha-<sha>`) |
| `deploy.yml` | CI success on `main` → **staging**; on `v*` tag → **production**. Only once `DEPLOY_ENABLED` is `true` (see below) | `flyctl deploy --image …`, then a `/readyz` smoke test |
| `release.yml` | `v*` tag; **Run workflow** | On a tag, creates the GitHub Release: GoReleaser's CLI/TUI binaries for Linux, macOS and Windows, and the desktop app's Linux and Windows installers. By hand, builds only the installers, as artifacts (see [docs/ci-cd.md](docs/ci-cd.md#making-a-release)) |

Dependabot keeps Go, npm, Actions and Docker dependencies current.

### One-time Fly.io setup

```sh
brew install flyctl   # or: curl -L https://fly.io/install.sh | sh
fly auth login

# For each environment (staging shown; repeat with fly.production.toml):
fly apps create go-application-starter-staging
fly volumes create app_data --size 1 --region iad --config fly.staging.toml --yes

# Let CI deploy:
fly tokens create org          # copy the token
```

Then, in GitHub, open **Settings → Environments** and:

1. Create `staging` and `production`. Add required reviewers to `production` if you want a manual approval gate.
2. Add the secret `FLY_API_TOKEN` (at repo level, or in each environment).
3. Make sure Actions may publish packages (**Settings → Actions → General → Workflow permissions**).
4. Switch deploys on: add the repository variable `DEPLOY_ENABLED` with the value `true` (**Settings → Secrets and variables → Actions → Variables**). Until then, `deploy.yml` skips instead of failing.

Fly must be able to pull the image from GHCR. The simplest option is to make the package public (**Package settings → Change visibility**). To keep it private, change `deploy.yml` to `flyctl deploy --remote-only` (Fly builds from source) instead of `--image`.

> ⚠️ Both file-based adapters need **one machine per volume**: `jsonfile` takes an exclusive lock, and a `sqlite` database can't be shared between machines (several processes on one machine are fine; see [ADR 7](docs/adr/0007-sqlite-storage.md)). Don't `fly scale count` above 1 until you've moved to a database server such as Postgres (planned).

### Docker

`Dockerfile` is a multi-stage build: Node builds the SPA, Go builds a static server with the SPA embedded, and the result is a small distroless image. `docker-compose.yml` runs it locally with a named volume.

<!-- template:start -->
## Using this as a template

1. Create a repo from this template, or copy it.
2. Run `task init -- --module github.com/you/your-app --name "Your App" --description "What it does" --publisher "Your Name" --email you@example.com`. Add `--remove-sample` to drop the sample `Item` domain too. See [docs/template.md](docs/template.md).
3. Fill in `SPEC.md`, then build your first capability with `/feature`.
<!-- template:end -->

## Ideas not included (yet)

- **Devcontainer** (`.devcontainer/`) so the toolchain is reproducible in Codespaces.
- **OpenTelemetry** tracing and metrics around the driving ports.
- **Auth** (OIDC) as HTTP middleware, plus a `Principal` passed through `context`.
- **Conventional commits + release-please** for automatic changelogs and versions.
- **Contract tests against the deployed API** by reusing the playwright-bdd suite with `E2E_BASE_URL=https://…`.
