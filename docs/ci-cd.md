# CI/CD

How this project checks, builds and ships itself on GitHub, in plain terms. The README's [CI/CD section](../README.md#cicd) has the short version and the one-time Fly.io setup.

- **CI** (continuous integration): every push and pull request runs the project's checks on a GitHub machine, so a mistake is caught even if nobody ran `task check` locally.
- **CD** (continuous delivery): the same machines build what you ship (the server's Docker image, the CLI and desktop downloads) and deploy the server.

All of it is **GitHub Actions**, which is built into GitHub: no other service or hosting is needed, except a host for the server if you deploy it. Public repositories get it free with no limit; private ones get a monthly allowance of minutes (macOS machines count ten times).

## Words you'll see

| Word | Meaning |
|---|---|
| **Workflow** | One YAML file in `.github/workflows/`, describing what to do and when. |
| **Trigger** (`on:`) | The event that starts a workflow: a push, a pull request, a tag, another workflow finishing, or a click on **Run workflow**. |
| **Run** | One time a workflow ran. Each has a page on the **Actions** tab, with every job's log. |
| **Job** | A group of steps that runs on its own fresh machine. Jobs in one workflow run side by side unless one `needs:` another. |
| **Runner** | The machine a job runs on (`runs-on: ubuntu-24.04`). It starts empty and is thrown away afterwards, which is why every job begins by checking out the code and installing Go or Node. |
| **Step** | One shell command (`run:`) or one ready-made **action** someone published (`uses: actions/setup-go@v5`). |
| **`if:`** | A condition on a job or step. When it's false, the job shows as *skipped* (grey), not failed. |
| **Artifact** | A file a job saves before its machine is thrown away, downloadable from the run's page. |
| **Tag** | A permanent name for one commit, e.g. `v1.2.0`. Pushing a `v*` tag is how you make a release. |
| **Release** | The page on the repository where people download a version's files. |
| **Secret** / **variable** | Settings kept on GitHub, not in the code (**Settings → Secrets and variables → Actions**). A secret (a password, like `FLY_API_TOKEN`) is hidden and can't be read back; a variable (a plain setting, like `DEPLOY_ENABLED`) can. `GITHUB_TOKEN` is a temporary secret GitHub gives every run by itself. |
| **Permissions** | What a run's `GITHUB_TOKEN` may do. Each workflow starts with read-only and gives write access only to the job that needs it. |

## The workflows

| Workflow | Starts on | What it does |
|---|---|---|
| `ci.yml` | a pull request; a push to `main`; a `v*` tag | The checks `task check` runs, as five jobs: Go (lint, unit, acceptance, terminal journeys), Web (lint, types, format, unit), generated code is up to date, browser e2e, and then the Docker image. The image job waits for the other four, and publishes the server's image to the repository's **Packages** (GHCR) only on a push, not for a pull request. |
| `deploy.yml` | `ci.yml` finishing | Deploys the image CI just built to Fly.io: a push to `main` goes to **staging**, a `v*` tag to **production**, and then a smoke test asks the new server's `/readyz`. Only after a successful CI run, and only once deploys are switched on (below). |
| `release.yml` | a `v*` tag; **Run workflow** | On a tag or a click, two desktop jobs build the installers and save them as artifacts: the Linux `.deb`, `.rpm` and AppImage, and the Windows setup `.exe` (cross-compiled on Linux). On a tag only, GoReleaser builds the CLI for Linux, macOS and Windows and creates the GitHub Release, and then `publish` adds the installers to it. A click never publishes anything. |

So, for each thing you do:

| You… | What runs |
|---|---|
| open or update a pull request | CI (the image is built, not published) |
| push to `main` (or merge a pull request) | CI, then Deploy to staging |
| push a `v*` tag | CI, then Deploy to production; and Release, which publishes the GitHub Release ([below](#making-a-release)) |
| click **Run workflow** on Release | the desktop jobs only |

### Keeping CI in step with `task check`

`ci.yml` runs the same commands as `task check`, written out step by step rather than calling Task. When you add a check, a test folder or a browser to `task check` or `task setup`, add it to `ci.yml` too, or CI quietly stops covering it. (CI once lacked the WebKit and Firefox browsers the phone e2e tests use, and failed the first time it ran them.)

`ci.yml`'s branch must match your repository's default branch (`main`). Under another name, pushes run no CI at all, and nothing says so.

## Deploys are off until you switch them on

`deploy.yml` runs only when the repository variable `DEPLOY_ENABLED` is `true`. Until then, every Deploy run shows as skipped, so a repository with no host set up yet doesn't fail a deploy on every push. To switch it on, follow the README's [one-time Fly.io setup](../README.md#one-time-flyio-setup) (or replace `deploy.yml` for another host), then add the variable under **Settings → Secrets and variables → Actions → Variables**.

## Dependabot

Dependabot (`.github/dependabot.yml`) checks once a week for newer versions of the Go modules (`go.mod`, `tools/go.mod`), the web app's npm packages, the actions the workflows use, and the Docker base images. For each, it opens a pull request, and CI runs on it. Nothing changes on `main` until you merge one.

Before merging, check two things:

- **Is it really tested?** A pull request runs CI in pull-request mode. A step that only runs on a push (signing in to GHCR) or a workflow that only runs on a tag (`release.yml`) isn't exercised, so the pull request is green without proving anything. Merge those, then watch the next push or tag.
- **Is it a major version** (v4 → v7)? That's its author warning that something may break. The pull request includes the release notes: look for anything you use being removed.

Some versions are held back on purpose, and `dependabot.yml` tells Dependabot to leave them alone, with the reasons in its comments:

- **Wails**, in all three places (`go.mod`, `tools/go.mod`, `web/package.json`), which must move together; and CI builds neither the desktop nor the Android app, so green wouldn't prove an upgrade safe.
- **TypeScript 7**, until typescript-eslint supports it.
- **A new Node major in the Dockerfile**, which goes with `.nvmrc` (Dependabot doesn't read it). Odd-numbered Node majors are never long-term releases. Move both to the next even one by hand.

A new oapi-codegen (in `tools/go.mod`) can change the generated API code, which fails CI's "generated code is up to date" job. Run `task gen` on the pull request's branch and commit the result.

To merge: the **Merge pull request** button, or `gh pr merge <number> --squash`. After you merge one, Dependabot updates the others that touch the same file; wait for their checks before merging the next. To turn an update down, close its pull request: Dependabot won't propose that version again.

## Making a release

A release is a version with a page of downloads. You make one by pushing a tag:

```sh
git switch main && git pull     # tag what's on GitHub's main
git tag v1.2.0
git push origin v1.2.0
```

Use a plain `vMAJOR.MINOR.PATCH` tag. The desktop installers take their version from it (`1.2.0`), and any other tag, such as `v1.2.0-rc.1`, builds installers numbered `0.0.0`.

The tag starts two workflows:

- **Release**, on the Actions tab. `goreleaser`, `desktop-linux` and `desktop-windows` run side by side; GoReleaser creates the Release with the CLI downloads, its `checksums.txt` and a list of the changes since the last tag. When all three are green, `publish` adds the installers. If any of them fails, `publish` is skipped and the installers aren't added: fix the problem, then delete the Release and the tag (below) and push the tag again.
- **CI**, as for a push, which also publishes the server's Docker image tagged `1.2.0` to GHCR, and then **Deploy** to production, once deploys are switched on.

When Release is green, the repository's **Releases** page (on the right of its front page) shows the version with all its downloads. `checksums.txt` covers the CLI downloads only.

A tag means "this exact commit is version 1.2.0", so don't move it once people may have downloaded it: fix a bad release with the next version. To undo one nobody has used yet, or a test:

```sh
gh release delete v1.2.0 --cleanup-tag --yes   # the Release and the tag on GitHub
git tag -d v1.2.0                              # your local copy of the tag
```

The Docker image stays in GHCR: delete it under the repository's **Packages** if you want it gone.

## Building the desktop installers by hand

To build the desktop installers on GitHub without making a release, e.g. after changing `build/desktop/` or the Release workflow:

1. Push the change to a branch on GitHub. (The **Run workflow** button appears only once `release.yml` with `workflow_dispatch:` is on the default branch.)
2. Open the repository's **Actions** tab, choose **Release** in the list on the left, then **Run workflow** at the right of the list of runs.
3. Choose the branch to build (usually `main`), then **Run workflow**. The new run appears in the list after a few seconds.
4. Open the run. `goreleaser` and `publish` are skipped (they only run for a tag); `desktop-linux` and `desktop-windows` run side by side. Open one to follow each step's log. Linux takes several minutes, most of them on the AppImage; Windows is quicker.
5. When it's green, scroll to **Artifacts** at the bottom of the run's **Summary** page. Each is a zip: `desktop-linux` holds `starter-desktop_0.0.0_amd64.deb`, `starter-desktop-0.0.0-1.x86_64.rpm` and `starter-desktop-0.0.0-x86_64.AppImage`; `desktop-windows` holds `starter-desktop-0.0.0-setup.exe`. A run from a branch is version `0.0.0`; only a release tag gives a version. Artifacts are deleted after 14 days.

To try a package, install it as you would any other, e.g. `sudo apt install ./starter-desktop_0.0.0_amd64.deb`, or make the AppImage executable and run it. On Windows, run the setup program; it's unsigned, so SmartScreen warns first (**More info → Run anyway**).

From the command line, with the GitHub CLI: `gh workflow run release.yml --ref main`, then `gh run watch` to follow it and `gh run download <run id> -n desktop-linux` (or `-n desktop-windows`) to fetch the files.

## When a run fails

1. Open the run and the red job, then the red step: its log is open at the error. The end of it usually says what went wrong.
2. Something that passes locally but fails on GitHub is usually something your machine has and the runner doesn't: a system package, a browser, a tool. Install it in the workflow, as `task setup` or the README's Prerequisites do locally.
3. To try again without a new commit, use **Re-run jobs** on the run's page (**Re-run failed jobs** reruns only the red ones).
4. A grey, skipped job isn't a failure: its `if:` was false. Deploy is skipped until it's switched on; GoReleaser and `publish` are skipped on a manual run, and `publish` also when a build before it failed.
