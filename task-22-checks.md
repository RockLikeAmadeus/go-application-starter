# Task 22: checks by hand (temporary)

The steps for task 22 of `docs/wip/desktop-and-mobile-remote-mode.md`: the desktop app against your own computer, and the Android app on the emulator or a phone. Delete this file when you're done. It isn't committed.

A note on A7 before you start. The offline notice doesn't appear when the server stops. It appears when the window itself loses its network connection, because it follows the browser's own online or offline signal (`navigator.onLine` in `web/src/features/offline/useIsOnline.ts`). So in remote mode, stopping the server will probably show an error instead of the notice. A7 checks both, so we learn which happens.

## Before you start

- Run `nvm use` in each terminal you use. Your shell may still default to Node 22, and the web build needs Node 24.
- **Make sure no old server holds port 8080.** On 2026-11-15 a server left over from 2026-11-11 (`task run:server`) was still running, so `task dev` couldn't get the port. Check with `ss -ltnp | grep :8080`. After A1, `curl -s localhost:8080/api/v1/session` must print `{"isSignInOn":true,"signInMethod":"page"}`. Without `signInMethod`, it's an old server: stop it, with Ctrl+C in its terminal or `kill <pid>`, and start A1 again.
- Desktop: the Wails and GTK dependencies from the README have to be installed.
- Android: the SDK and an emulator from `docs/mobile.md`. A real phone works too; see B1.

## Part A: the desktop app against `task dev`

**A1. Start the server with the dev sign-in.** Terminal 1:

```sh
STARTER_SIGN_IN=dev STARTER_STORAGE=memory task dev
```

- This starts the API on `:8080` with the "type any name" sign-in, and opens the web app in your browser. You can close that tab.
- Memory storage forgets every account when the server restarts, which keeps each attempt clean.

**A2. Open the desktop app in remote mode.** Terminal 2:

```sh
STARTER_SERVER_URL=http://localhost:8080 STARTER_CREDENTIAL_STORE=file task run:desktop
```

- `STARTER_SERVER_URL` puts the app in remote mode for this run only, without changing your settings file.
- `STARTER_CREDENTIAL_STORE=file` keeps the app's token in a file you can look at: `~/.config/starter/app-tokens.json`. Leave it out to try the system's credential store instead.
- **Expect:** a window showing "Please sign in or create an account to continue.", with **Sign in** and **Create an account** buttons. Under them: "Your browser will open so that you can sign in or create an account."

**A3. Sign in.**

1. Press **Sign in**.
2. **Expect in the window:** the heading "Signing in", the text "Please finish signing in in your browser. The application will continue once you have signed in.", and a **Cancel** button.
3. **Expect in your browser:** a new tab with the dev sign-in form. Type a name, such as `ada`, and submit it.
4. **Expect in the browser:** "Sign-in complete. Please close this window and return to the application that initiated the sign-in."
5. **Expect in the window, without touching it:** your workspace, signed in as `ada`.
6. Optional: run `cat ~/.config/starter/app-tokens.json`. It should hold one token, for `http://localhost:8080`.

**A4. Sign out.**

1. Press **Sign out** in the window.
2. **Expect:** the sign-in screen again, straight away. This used to give a blank page, which tasks 21b to 21f fixed. If you see a blank page, the app or the server is an old build.
3. **Expect:** `app-tokens.json` no longer holds the `localhost:8080` token.

**A5. Cancel a sign-in.**

1. Press **Sign in**, and when the browser tab opens, don't submit the form.
2. Press **Cancel** in the window.
3. **Expect:** the sign-in screen, with "Sign-in was cancelled." above the usual text, in muted grey rather than red.
4. Now submit the form in that old browser tab. **Expect:** the window stays signed out, because that attempt was cancelled. The tab will probably show an error page; note what it says.

**A6. Create an account.**

1. Press **Create an account**.
2. **Expect in the window:** the heading "Creating your account", the text "Please finish creating your account in your browser. The application will continue once you have finished.", and **Cancel**.
3. With the dev sign-in, sign-up shows the same name form. Type a new name, such as `grace`, and submit it.
4. **Expect:** the window opens `grace`'s workspace, which is empty and separate from `ada`'s.

**A7. Lose the connection.** While signed in:

1. Stop the server: press Ctrl+C in Terminal 1.
2. In the window, move to another page, or create an item. **Note what you see:** the offline notice ("You're offline. Data will load when the connection comes back."), an error message, or nothing.
3. Start the server again (A1). Memory storage forgot every account and session, so the app's token is no longer valid. **Note** whether the app returns you to the sign-in screen or shows an error.
4. Optional, to see the real offline notice: turn off Wi-Fi or unplug the network cable while the app is open. Then turn it back on.

**A8. Sign out while the server is down.**

1. Sign in again.
2. Stop the server.
3. Press **Sign out**.
4. **Expect:** the sign-in screen, and the token gone from `app-tokens.json`, even though the server couldn't end the session.

## Part B: the Android app built with the server's address

This is the risk task 22 exists for:

1. The phone's browser signs in on the server.
2. The server then sends the browser to `http://127.0.0.1:<port>/…`, a port the app itself listens on.
3. We need to know whether that reply reaches the app.

Android might pause the app while the browser is in front, or the browser might refuse the address. The acceptance tests can't cover this.

**B1. Choose the server's address as the phone sees it.**

- **Emulator:** `http://10.0.2.2:8080`. `10.0.2.2` is the emulator's name for your computer.
- **Real phone on the same Wi-Fi:** your computer's address on the network, from `hostname -I` (the first one, for example `http://192.168.1.20:8080`). The server listens on every network interface, but your firewall may block port 8080. If the phone's browser can't open `http://<that address>:8080/healthz`, allow the port with `sudo ufw allow 8080/tcp`, and remove the rule again afterwards with `sudo ufw delete allow 8080/tcp`.

**B2. Start the server** as in A1, and leave it running.

**B3. Build the app with the address, install it, and open it.**

```sh
BUILT_IN_SERVER_URL=http://10.0.2.2:8080 task mobile:android:run
```

- Use the real phone's address from B1 instead of `10.0.2.2` if you're using a phone.
- This starts the first emulator if no phone is connected.
- **Expect:** the same sign-in screen as on the desktop.

**B4. Sign in through the phone's browser.**

1. Press **Sign in**.
2. **Expect:** the phone's browser opens on the dev sign-in form. Type a name and submit it.
3. **The point of the test:** does the browser show "Sign-in complete. Please close this window…"? Switch back to the app. Does it show the workspace?
4. If it fails, note where it stops:
   - Does the browser open at all?
   - Does the form load?
   - Does the page after signing in show an error, and what does it say?
   - What does the app show?

   Run `adb logcat -s GoLog` while you try. That's a best guess at where the Go side's messages go; if it shows nothing, `adb logcat | grep -i starter` shows more. Keep whatever looks related.

**B5. If B4 worked,** repeat A4 (sign out), A5 (cancel), and A6 (create an account) on the phone. The Android app always keeps its token in its private storage. A debug build lets you check it with `adb shell run-as com.example.starter cat files/starter/app-tokens.json`. The path is worked out from `mobile.Open`'s settings folder, so if the file isn't there, `adb shell run-as com.example.starter ls -R files` shows where it is.

## What to send back

Paste this back with your notes:

```
A2 window as expected:            yes
A3 sign in:                       yes
A4 sign out, token gone:          yes
A5 cancel message, old tab:       yes / no:   old tab showed:
A6 create an account:             yes / no:
A7 server stopped, I saw:
A7 server restarted, I saw:
A8 sign out while down:           yes / no:
B  emulator or phone:
B4 browser reply reached the app: yes / no:   where it stopped:
B5 sign out / cancel / create:
```

If something doesn't work, the fix goes in as its own TDD cycle before task 22 is ticked. A7 might show that the offline notice can't appear without a server. If so, you'll be asked whether to fix it here or add it to the backlog.
