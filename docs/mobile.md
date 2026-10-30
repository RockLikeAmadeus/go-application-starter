# The Android app

The mobile app (SPEC.md §4.7) is the web app as an Android app, with the core and its data on the phone. It needs no server and works offline. It uses Wails v3, like the desktop app; ADR 4 explains why. This page covers setting up the Android tools, running the app on the emulator or a phone, debugging it, and how it works inside.

iPhone, remote mode, release signing and an app icon aren't done yet: see `TODO.md`.

## Set up the Android tools (once)

`task setup` doesn't install these, because only the mobile app needs them. Tested on Ubuntu with the versions below. On macOS the same steps should work with the macOS downloads, but they haven't been tried.

1. **A JDK, 17 or newer** (Gradle needs it; 21 is tested):

   ```sh
   sudo apt install openjdk-21-jdk-headless
   ```

2. **The Android SDK command-line tools.** Download "Command line tools only" from https://developer.android.com/studio#command-tools and unpack it so that `sdkmanager` ends up at `~/Android/Sdk/cmdline-tools/latest/bin/sdkmanager`:

   ```sh
   mkdir -p ~/Android/Sdk/cmdline-tools && cd ~/Android/Sdk/cmdline-tools
   unzip ~/Downloads/commandlinetools-linux-*_latest.zip
   mv cmdline-tools latest
   ```

   Android Studio installs the same SDK, also in `~/Android/Sdk` on Linux. Use it instead if you prefer.

3. **Accept the licences, then install the packages.** Read each licence before you accept it:

   ```sh
   export ANDROID_HOME=~/Android/Sdk
   $ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager --licenses
   $ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager \
     "platform-tools" "platforms;android-35" "build-tools;35.0.0" \
     "ndk;26.3.11579264" "emulator" "system-images;android-35;google_apis;x86_64"
   ```

   The NDK compiles the Go side of the app. The emulator and the system image are only needed for the emulator. On an Apple-silicon Mac, use the `arm64-v8a` system image instead of `x86_64`.

4. **Tell the Taskfile where the SDK is**, if it isn't in `~/Android/Sdk` (for example Android Studio's `~/Library/Android/sdk` on macOS): `export ANDROID_HOME=…`. The newest NDK in the SDK is used, unless `ANDROID_NDK_HOME` names another.

## Create an emulator

```sh
echo no | $ANDROID_HOME/cmdline-tools/latest/bin/avdmanager create avd \
  -n pixel -k "system-images;android-35;google_apis;x86_64" -d pixel_7
```

On Linux the emulator needs KVM. If it complains about `/dev/kvm`, add yourself to the `kvm` group (`sudo usermod -aG kvm $USER`) and log in again.

## Build and run

```sh
task mobile:android:run      # build, install on the phone or emulator, and open the app
task mobile:android:build    # just build bin/starter.apk
```

- `task mobile:android:build` builds the web app, compiles the Go side for phones (`arm64-v8a`) and the emulator (`x86_64`), and runs Gradle. The result is a debug APK in `bin/starter.apk`, about 19 MB.
- `task mobile:android:run` installs it on the connected device and opens it. If none is connected, it starts the first emulator (`AVD=pixel` picks one) and waits for it to boot, which takes under a minute.
- **A real phone:** turn on *Developer options* › *USB debugging*, plug it in, and accept the prompt on the phone. `adb devices` should list it. With more than one device connected, set `ANDROID_SERIAL` to the one you want.
- Reinstalling keeps the app's data (`adb install -r`), as a store update would. Uninstalling the app deletes it.

The data folder is in the app's private storage: `/data/data/com.example.starter/files/starter`. A debug build can read it with `adb shell run-as com.example.starter ls files/starter`.

## Debug it

- **The web app:** open `chrome://inspect` in Chrome on your computer. The app's WebView appears under the device, with the usual DevTools (console, network, elements). This works because the APK is a debug build.
- **The Go side:** Go's own output goes nowhere on Android. Wails sends its logs, and the app's startup errors, to logcat: `adb logcat -s Wails`.
- **Harmless noise:** Gradle prints two Java deprecation "Note:" lines. Logcat shows `error serving asset /wails/custom.js`, an optional script the Wails runtime looks for, and `/sw.js` answers 404 on purpose (no service worker, as on the desktop).

## How it works

The phone runs the desktop app's wiring. `mobile.Open` (in `internal/adapters/mobile`) calls `desktop.Open` with the app's private storage as the settings folder, so the core, jsonfile storage and the REST handler are the same as on the desktop. `cmd/mobile/main.go` (build tag `android`) only wraps it in Wails. Four things differ from the desktop.

**The API goes over the Wails bridge.** On Android, Wails hands every page request to Go as a `GET` with no body, and treats any status but 200 as "not found". So the web app can't reach the REST API the normal way: a `POST` would arrive as a `GET`, and a 404 or 422 would be lost. Instead:

1. On Android, `web/src/api/client.ts` gives the API client `apiFetch`, which sends each request through `bridgeFetch` (`web/src/api/bridge.ts`).
2. `bridgeFetch` calls the Go method `main.Bridge.Serve(method, path, body)` with the Wails runtime's `Call.ByName`.
3. `Bridge.Serve` (in `cmd/mobile`) passes it to `mobile.Serve`, which runs the request through the REST handler and returns the status, content type and body.
4. `bridgeFetch` turns that back into a normal `Response` (with no body for a 204). The web app's code doesn't know the difference.

Wails names a bound method after its Go package and type. `cmd/mobile` is package `main`, so the name is `main.Bridge.Serve`, and it survives `task rename`. Renaming the `Bridge` type, or moving it to another package, breaks the app.

**`/api` on the page path answers 404.** `mobile.AssetHandler` serves the web app for Wails' page path. Any `/api/…` request that reaches it has missed the bridge, so it fails loudly instead of being answered as a `GET`. Without this, a spike build showed a saved item as saved when nothing had been stored. The handler also serves `/` when Android asks for `/index.html`, because Go's file server would answer with a redirect, which Wails treats as "not found".

**The bridge check runs once, before the runtime loads.** Android's bridge shows up as `window.wails.invokeAsync`. Importing the Wails runtime replaces `window.wails`, so `apiFetch` checks for the bridge once, when the API client is created, and imports the runtime only when it's there. The runtime is its own chunk, and the service worker doesn't precache it, so browsers never download it.

**`main` runs once per process.** Wails calls `main` each time Android creates the app's screen. That happens again in the same process after a system theme change or Back and reopen, while the first app is still running, because on Android `app.Run` never returns. So `main` is wrapped in a `sync.Once`; otherwise the second copy would fail on the data-folder lock. For the same reason `opened.Close()` never runs, which is fine: jsonfile writes each change straight away.

The acceptance tests' `mobile` driver (`test/acceptance/drivers.go`) runs every feature's scenarios across the same `mobile.Serve` bridge, without a phone.

## Upgrading Wails

Wails mobile is experimental: it's outside Wails' stability promise, and the bridge depends on how Wails' Android code behaves today. After an upgrade, run the app on the emulator and save, change and delete something before trusting it.

- **Keep the versions in step.** The Go module (`go.mod`) and the `@wailsio/runtime` npm package (`web/package.json`, pinned exactly) must be the same version. The wails3 CLI in `tools/go.mod` should match too.
- **Regenerate the Android project** if the new version changes it. `mobile/android/` is Wails' generated Gradle project with a few edits. Generate a fresh one and compare:

  ```sh
  go tool -modfile=tools/go.mod wails3 generate build-assets -dir /tmp/wails-assets \
    -name starter -productname Starter -productidentifier com.example.starter
  diff -r /tmp/wails-assets/android mobile/android
  ```

  Copy over everything except `Taskfile.yml`, `scripts/` (both for wails3's own build) and `main_android.go` (`cmd/mobile/main.go` already does its job). Then reapply the edits: `applicationId`, `app_name` and `rootProject.name`; the comment on `namespace`; the trimmed permissions in `AndroidManifest.xml`; and `chmod +x gradlew`. Keep `mobile/android/.gitignore`.
- **Keep the Java package `com.wails.app`.** The Go library's JNI functions are named after it, so only `applicationId` changes per app.
- Wails 3.0.0 exposes the Android bridge to every frame in the WebView; a fix is planned for 3.0.1. The app embeds no third-party pages, so this doesn't matter yet.
