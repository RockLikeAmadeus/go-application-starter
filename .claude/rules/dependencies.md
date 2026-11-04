---
paths:
  - "go.mod"
  - "tools/**"
  - "web/package.json"
  - ".github/dependabot.yml"
  - "Taskfile.yml"
  - ".nvmrc"
---

# Dependencies and dev tools

- Dev-only Go tools (oapi-codegen for `task gen`, Air for `task dev`) live in `tools/go.mod` so their dependencies stay out of the app's `go.mod`. Add one with `cd tools && go get -tool <pkg>@<version>`, and run it with `go tool -modfile=tools/go.mod <name>`.
- Keep `@wailsio/runtime` in `web/package.json` (pinned exactly), the Wails module in `go.mod` and wails3 in `tools/go.mod` on the same version. Dependabot ignores all three (`.github/dependabot.yml`), so upgrade them together by hand. On a Wails upgrade, also regenerate `mobile/android` (`docs/mobile.md`) and replace `build/desktop/windows/wails_tools.nsh`.
- SQLite uses `modernc.org/sqlite`, a pure-Go driver: keep it that way, because the Windows tasks set `CGO_ENABLED=0`.
- Node ≥ 24 is required (see `.nvmrc`). TypeScript is pinned to 6.0.x until typescript-eslint supports 7. Dependabot ignores TypeScript 7 until then; drop that rule in `.github/dependabot.yml` when unpinning.
