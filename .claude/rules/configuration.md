---
paths:
  - "internal/config/**"
  - "internal/bootstrap/**"
  - "cmd/**"
  - "docs/configuration.md"
---

# Settings (ADR 6, `docs/configuration.md`)

Settings come in layers: built in < a TOML settings file < `STARTER_*` env vars (and `PORT`) < CLI options. The CLI, TUI and desktop app read `config.toml` in the app's settings folder (`~/.config/starter/config.toml` on Linux; `config.LocalConfigFile`), or the file `STARTER_CONFIG_FILE` or `--config` names. The server reads a file only when `STARTER_CONFIG_FILE` names one, so your own settings file never changes `task dev`. An unknown key, bad TOML or a missing named file fails at start; a missing default file is fine.

`internal/config` never reads `os`: pass `config.Env` (`os.Getenv`, `os.UserConfigDir` in `cmd/*`, fakes in tests). A new setting goes in `envVars`, `Load` and `Config.Settings` (`internal/config/config.go`), and in `docs/configuration.md`.
