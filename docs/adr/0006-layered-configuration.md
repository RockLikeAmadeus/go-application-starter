# 6. Settings come in layers: a settings file under the environment

- Status: accepted
- Date: 2026-11-02

## Context

Every setting came from an environment variable (`STARTER_STORAGE`, `STARTER_DATA_DIR`, …), with a few CLI options on top. That suits a server: Docker, Fly.io, CI and secret stores all set environment variables. It doesn't suit the local edition (ADR 5). Desktop users can't set environment variables for an app they open from a menu, yet they must be able to choose their data folder. And the CLI, TUI and desktop app share one data folder, but had no shared place for settings. The TODO item asked whether to replace the environment variables with a config file.

## Decision

- **Add a settings file as a layer; don't replace the environment variables.** The order is built-in defaults < settings file < environment variables < command-line options. The same keys exist in every layer (see [docs/configuration.md](../configuration.md)).
- **The environment overrides the file.** This was a deliberate choice: the first instinct was the file over the environment. Viper, koanf, git, the AWS CLI, npm and Docker all put the environment above the file, and it keeps three things working:
  - a one-off override (`STARTER_DATA_DIR=/tmp/x starter …`) without editing the file;
  - deployments and secrets, set as environment variables, beating a file baked into an image;
  - tests, which isolate themselves through the environment.

  The cost is that a forgotten environment variable silently beats the file. `starter info` answers it by showing where each setting came from.
- **TOML**, read with **koanf** (`github.com/knadh/koanf/v2`, its file provider and TOML parser). TOML has comments and no indentation rules, and is common for app settings (Cargo, Hugo, Fly.io).
- **One file for the local front ends:** `config.toml` in the app's folder in the user's settings folder, beside the default data folder. `STARTER_CONFIG_FILE` or `--config` names another. The default file may be missing; a named one may not.
- **The server reads a file only when `STARTER_CONFIG_FILE` names one.** It never reads the user's settings folder, so a developer's own file can't change `task dev`'s server, and a deployment stays configured by what it sets.
- **Strict:** an unknown key or bad TOML stops the app at start, naming the file and the key, as a bad plans file does. A typo never leaves the app running on a setting it ignored.
- **`internal/config` reads nothing from the process.** `config.Env` hands it the environment and the user's settings folder, so every rule is unit-tested without touching the real ones.

Rejected:
- **Replacing the environment variables with the file:** every deployment target and CI would need a file written into it, and secrets would end up in files.
- **The file over the environment:** see above.
- **Viper:** a large dependency tree, global state by default, and case-insensitive keys.
- **A hand-written parser:** strictness, sources and TOML's edge cases are what a library already does well.
- **YAML or JSON:** YAML's indentation and implicit types trip people up in a file they edit by hand; JSON has no comments.

## Consequences

- Desktop users can choose their data folder (SPEC.md §4.6) by writing one line in the settings file.
- Adding a setting means adding it to every layer: a field, a key, an environment variable and a line in `starter info` (see [docs/configuration.md](../configuration.md)).
- The editions plan's `starter login` (step 9) will need to remember a server and a session. This file is the user's, so whether the app writes those here or in a file of its own is decided then.
