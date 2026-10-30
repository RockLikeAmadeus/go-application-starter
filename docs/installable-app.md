# Installing and testing the web app

The web app can be installed like a native app (SPEC.md §4.5). It opens in its own window, opens offline after one visit, and updates itself quietly. This page explains how to try that by hand. The browser scenarios in `web/e2e/features/install.feature` cover the same ground automatically, with `task e2e`.

## Run a build, not the dev server

The service worker only exists in a production build. It is off under `task dev` (Vite's dev server), so test installing against the Go server with the app embedded:

```sh
task build
./bin/server
```

This serves the app and the API on http://localhost:8080, with data in `./data`. Browsers treat `localhost` as secure, so it needs no HTTPS.

## Install it on a computer

Use Google Chrome, Chromium or Edge. Firefox can't install web apps on Linux or macOS. Safari on macOS uses *File › Add to Dock*.

1. Open http://localhost:8080.
2. Click the install icon at the right of the address bar, or open the ⋮ menu › *Cast, save and share* › *Install page as app…*.
3. The app opens in its own window and appears in your app launcher as "Starter".

To uninstall it, use the app window's ⋮ menu › *Uninstall Starter*, or go to `chrome://apps`.

## Install it on a phone

A service worker needs HTTPS on any address other than `localhost`. So a phone on your Wi-Fi can't install the app from `http://<your-computer's-IP>:8080`. Instead, make the phone's `localhost` reach your computer:

- **Android, over USB:** turn on USB debugging on the phone, then run `adb reverse tcp:8080 tcp:8080` on the computer. Open http://localhost:8080 in Chrome on the phone, then use ⋮ › *Add to Home screen* › *Install*. The Android emulator works the same way.
- **iPhone:** iOS has no `adb reverse`. Serve the app over HTTPS (a tunnel such as `cloudflared` or `ngrok`, or a deployed copy), open it in Safari, then use *Share* › *Add to Home Screen*.

## Things to check

| Check | How | Expected |
|---|---|---|
| Opens offline | Open the installed app once, stop `./bin/server` (Ctrl+C), then reopen the app | The app opens. Pages that need data show no data. |
| Offline notice | With the app open, turn off your network (Wi-Fi or cable), then turn it back on | "You're offline…" appears, then goes away and data loads. Stopping the server doesn't count as offline: the browser still has a network, so no notice appears. |
| Quiet update | Change something visible in `web/`, run `task build`, restart `./bin/server`, then open the app twice | The first open still shows the old version, with no prompt. The second shows the new one. |
| Fresh data | Add an item, go offline, reopen | The item isn't shown from a stored copy. Data only ever comes from the server. |

## Changing the icon or colours

- **Icons:** replace `web/public/icon.svg`, run `pnpm --dir web icons`, and commit the generated PNGs and `favicon.ico`.
- **Name:** `task rename` changes it along with the rest of the project (see [rename.md](rename.md)). A long name is also used as the short name, which phones may cut off under the home-screen icon. Shorten `short_name` in `web/vite.config.ts` by hand if needed.
- **Window colour:** set by `appBackground` in `web/vite.config.ts`, and repeated in the `theme-color` meta in `web/index.html`.

An installed app keeps its icon and name until the browser next checks the manifest. To see a change straight away, uninstall and reinstall the app.
