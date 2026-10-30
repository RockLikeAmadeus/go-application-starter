# go-application-starter

A template for multi-front-end applications: a **hexagonal Go core** driven by a **React web app** (through a REST API), a **Cobra CLI**, and a **Bubble Tea TUI**, with swappable storage. It is built **BDD outside-in (Godog + playwright-bdd) with a TDD inner loop**, and set up for agentic development with Claude Code.

<!-- sample:start -->
It ships as a working *walking skeleton*: one tiny `Item` domain wired end to end through every layer, tested at every level, and deployable from CI.
<!-- sample:end -->

```
 React SPA ──HTTP──┐
 CLI / TUI ──HTTP──┤ (--server mode)
                   ▼
            httpapi adapter ──► driving ports ◄── CLI / TUI (embedded mode)
                                       │
                                  core (domain + app)
                                       │
                               driven ports
                              ┌────────┴────────┐
                           memory            jsonfile   (← add SQL/NoSQL here)
```

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.27+ | https://go.dev/dl |
| Node | 24 (see `.nvmrc`) | `nvm install` |
| pnpm | 10 | `npm i -g pnpm@10` |
| Task | 3 | `go install github.com/go-task/task/v3/cmd/task@latest` |
| golangci-lint | 2 | https://golangci-lint.run/welcome/install/ |
| lefthook (optional) | | `go install github.com/evilmartians/lefthook@latest` |
| Docker (optional) | | for `task docker:*` |

Then:

```sh
task setup     # Go modules, pnpm install, Playwright browser, git hooks
task check     # codegen drift, lint, unit, acceptance, e2e: everything CI runs
```

## Everyday commands

```sh
task dev                         # API on :8080 + Vite on :5173, opened in your browser; both reload on save
task run:cli -- --help           # CLI, embedded core, data in ./data (add --server http://localhost:8080 to use a running server)
task run:tui                     # terminal UI (also accepts --server)
task test                        # unit tests (Go + web)
task accept                      # Godog acceptance tests
task e2e                         # browser tests against the real server binary
task gen                         # regenerate code after editing api/openapi.yaml
task build                       # ./bin/server (with UI embedded) and ./bin/starter
task docker:up                   # production image locally on :8080
```

Configuration is through environment variables (see `.env.example`): `STARTER_STORAGE` (`jsonfile` or `memory`), `STARTER_DATA_DIR`, `PORT`, `STARTER_LOG_FORMAT`, and `STARTER_SERVER_URL`.

## Development workflow (with Claude Code)

| Step | Command | Result |
|---|---|---|
| Outer loop | `/feature <desc>` | SPEC.md section, `@wip` Gherkin in `features/`, shown failing, TDD task list |
| Red | `/red <behaviour>` | exactly one failing unit test |
| Green | `/green` | the minimum code to pass, then stop |
| Refactor | `/refactor` | cleaner code, all tests still green |

Repeat red → green → refactor until `task accept` is green, then drop the `@wip` tag. See `CLAUDE.md` for the architecture rules Claude follows, and `.claude/settings.json` for permissions and the auto-format hook.

### Two layers of Gherkin

- `features/*.feature` (**Godog**) holds the business rules in domain language. Each scenario runs against three drivers (in-memory core, jsonfile storage, and the full HTTP server+client), all in-process and fast.
- `web/e2e/features/*.feature` (**playwright-bdd**) holds a thin set of browser journeys proving the UI → API → core wiring.

Keep the business rules in the first layer. Browser tests are slow and fragile, so they only prove the wiring.

## CI/CD

GitHub Actions, in `.github/workflows/`:

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | PRs, `main`, `v*` tags | Go lint/unit/acceptance, web lint/typecheck/unit, codegen drift, playwright-bdd e2e. Then builds the Docker image and, on push, publishes it to GHCR (`ghcr.io/<owner>/<repo>:sha-<sha>`) |
| `deploy.yml` | CI success on `main` → **staging**; on `v*` tag → **production** | `flyctl deploy --image …`, then a `/readyz` smoke test |
| `release.yml` | `v*` tag | GoReleaser publishes CLI/TUI binaries for Linux, macOS and Windows |

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

Fly must be able to pull the image from GHCR. The simplest option is to make the package public (**Package settings → Change visibility**). To keep it private, change `deploy.yml` to `flyctl deploy --remote-only` (Fly builds from source) instead of `--image`.

> ⚠️ The `jsonfile` adapter requires **one machine per volume** (it takes an exclusive lock). Don't `fly scale count` above 1 until you've moved to a database adapter.

### Docker

`Dockerfile` is a multi-stage build: Node builds the SPA, Go builds a static server with the SPA embedded, and the result is a small distroless image. `docker-compose.yml` runs it locally with a named volume.

<!-- template:start -->
## Using this as a template

1. Create a repo from this template, or copy it.
2. Run `task init -- --module github.com/you/your-app --name "Your App"`. Add `--remove-sample` to drop the sample `Item` domain too. See [docs/template.md](docs/template.md).
3. Fill in `SPEC.md`, then build your first capability with `/feature`.
<!-- template:end -->

## Ideas not included (yet)

- **Devcontainer** (`.devcontainer/`) so the toolchain is reproducible in Codespaces.
- **SQLite adapter** as a second storage example. Run the contract suite, then switch it in `bootstrap`. This also removes the single-machine limit if you move to LiteFS or Postgres.
- **OpenTelemetry** tracing and metrics around the driving ports.
- **Auth** (OIDC) as HTTP middleware, plus a `Principal` passed through `context`.
- **Conventional commits + release-please** for automatic changelogs and versions.
- **Contract tests against the deployed API** by reusing the playwright-bdd suite with `E2E_BASE_URL=https://…`.
