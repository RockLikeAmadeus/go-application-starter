---
paths:
  - "internal/config/**"
  - "internal/bootstrap/**"
  - "cmd/**"
  - "docs/configuration.md"
---

# Settings (ADR 6, `docs/configuration.md`)

Settings come in layers: built in < a TOML settings file < `STARTER_*` env vars (and `PORT`) < CLI options. The CLI, TUI and desktop app read `config.toml` in the app's settings folder (`~/.config/starter/config.toml` on Linux; `config.LocalConfigFile`), or the file `STARTER_CONFIG_FILE` or `--config` names. The server reads a file only when `STARTER_CONFIG_FILE` names one, so your own settings file never changes `task dev`. An unknown key, bad TOML or a missing named file fails at start; a missing default file is fine.

The desktop app, the CLI and the TUI create the default file when it's missing (`config.CreateDefaultSettingsFile`, called by `desktop.Open` and `cli`'s `settingsFileFor`), with the text of `config.DefaultSettingsFileText`. They never overwrite a file (`O_EXCL`), never create one when `STARTER_CONFIG_FILE` or `--config` names another, and carry on silently when they can't write it. The server and the Android app (`desktop.OpenWithoutDefaultSettingsFile`) never create one.

`internal/config` never reads `os`: pass `config.Env` (`os.Getenv`, `os.UserConfigDir` in `cmd/*`, fakes in tests). A new setting goes in `envVars`, `Load` and `Config.Settings` (`internal/config/config.go`), and in `docs/configuration.md`'s table. If the desktop app, the CLI or the TUI reads it, it also goes in `config.DefaultSettingsFileText`, with a note above it, set to its default or commented out when it has no useful default. The note is user-facing text: agree its wording with me first (`user-facing-text.md`). Then change the example file in `docs/configuration.md` to match.
