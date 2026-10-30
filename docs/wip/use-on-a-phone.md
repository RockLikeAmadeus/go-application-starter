# WIP: Use on a phone

- Spec: [SPEC.md §4.4](../../SPEC.md#44-use-on-a-phone-web-only)
- Scenarios: [`web/e2e/features/phone.feature`](../../web/e2e/features/phone.feature) · steps: [`web/e2e/steps/phone.steps.ts`](../../web/e2e/steps/phone.steps.ts)
- Parent plan: step 1 of the native front ends work (phone layout → installable web app → Wails desktop → Capacitor mobile).

Run: `task e2e -- --grep @wip` (all projects) · `--project phone` / `--project iphone` for one device.

## Decisions
- Presentation only, like theme and style: no Godog scenarios. The browser feature is the spec.
- Playwright projects: `chromium` (desktop) runs everything **except** `@phone`. `phone` (Pixel 7, Chromium) and `iphone` (iPhone 15, WebKit) run **only** `@phone`. WebKit stands in for Safari until a real iPhone is available.
- WebKit on this machine needs `sudo apt-get install libavif16` (or `sudo pnpm --dir web exec playwright install-deps`). `task setup` now installs webkit too.
- The breakpoint is Tailwind's `md` (768px). Below it, the `aside` is hidden and a top bar with a "Menu" button shows instead. At `md` and up, today's sidebar is unchanged.
- The drawer is a native `<dialog>` opened with `showModal()`, so the browser provides the focus trap, Escape and inertness of the page behind it. React state `open` drives it, and the dialog's `close` event syncs the state back. Escape itself is browser behaviour, so only e2e covers it. Unit tests cover the close-event → state sync.
- "Tap outside" means a click whose target is the `<dialog>` element itself, i.e. the backdrop around the drawer panel. The panel must be an inner element that fills the drawer.
- The drawer and the sidebar render the same `SidebarContent` (logo link, nav, Appearance fieldset). There's no second copy of the nav markup. The home-page link in the drawer is the logo ("Starter"), so the close-on-navigation scenario doesn't depend on the sample Items page.
- CSS-only behaviour (hiding the aside, the top bar only on phones, tighter padding, no overflow) can't be seen in jsdom. The failing e2e scenarios demand it, not unit tests.
- Colours and shape come only from palette/style tokens (`bg-sidebar`, `border-sidebar-line`, `rounded-panel`, …), and every style × theme is checked by the "Pages fit" outline.

## Tasks
1. [x] `Layout` shows a "Menu" button with `aria-expanded="false"`, and no dialog is open at first.
   - Red: `Layout.test.tsx` › "offers a closed Menu button for phone screens" fails because no "Menu" button exists.
   - Green: a bare `<button aria-expanded={false} className="md:hidden">Menu</button>` at the top of `<main>`. `md:hidden` keeps the wide-screen e2e scenario green.
   - e2e after green (`phone` + `chromium`): 3 pass, 10 fail. "Open the menu" and "Change the style from the menu" pass **falsely** for now, because the desktop `aside` nav is still visible on phones. They become meaningful once task 6 hides the aside. "Pages fit" overflows by 139–239px, which is roughly the `w-60` aside. `iphone` can't launch WebKit until `libavif16` is installed.
   - Refactor: no change, on purpose. A `MobileNav` component holding only one static button would be an abstraction with nothing behind it yet. A top bar with the logo and token styling is new layout, which belongs to task 6 (the e2e scenarios demand it), not to a no-behaviour-change step. Carried forward:
     - extract `MobileNav` (button + `<dialog>` + `open` state) when task 2 adds the state;
     - put the button in a styled phone top bar in task 6.
2. [x] Pressing "Menu" opens a dialog that contains the navigation and the Appearance group, and sets `aria-expanded="true"`. Checked: jsdom hides a closed `<dialog>` (`display: none`, so `queryByRole` skips it) but has **no `showModal`/`close`**. Stub them on `HTMLDialogElement.prototype` in `src/test/setup.ts` (toggling `open`, and firing `close` from `close()`).
   - Red: `Layout.test.tsx` › "opens the navigation and Appearance controls in a dialog from the Menu button" fails because `aria-expanded` stays `"false"`. The `showModal`/`close` stand-ins are in `src/test/setup.ts`, defined only where missing (`??=`).
   - Green: `menuOpen` state + `menuRef`. The button's onClick sets the state and calls `showModal()`. A bare `<dialog>` holds a **copy** of the nav (an unstyled Items `NavLink`) and the Appearance fieldset with a second `StylePicker`/`ThemeToggle`, sharing the same hooks. Neither uses fixed ids, so the duplicate controls don't clash.
   - e2e after green (`phone` + `chromium`): 1 passes (wide screen), 12 fail. With the drawer open, the phone now shows **two** navs and two Style pickers (the aside's and the drawer's), so Playwright's strict mode trips. That's honest now: the earlier false passes are gone. Task 6 (hiding the aside below `md`) fixes it.
   - Refactor, all in `web/src/routes/Layout.tsx`:
     - `SidebarContent` (logo link, styled nav, Appearance fieldset) is the one copy of that markup. `Layout` builds it once and renders it in both the `aside` and the drawer.
     - `useTheme`/`useStyle` stay in `Layout` and are passed down as props, so both copies share one state. Two hook instances would drift apart.
     - `MobileNav` owns `open`, `dialogRef` and an `openMenu` handler, and takes the content as `children`.
     - Side effects of sharing the markup:
       - The drawer now also contains the logo link, which task 3's scenario uses.
       - The drawer's "Appearance" legend is now screen-reader-only, like the sidebar's. The accessible name is unchanged.
     - Deliberately left:
       - The `aside`'s container classes stay in `Layout`, where the desktop layout lives.
       - `MobileNav` stays a local component in `Layout.tsx`, matching `ListIcon`/`Home`. Move it into its own file if it grows past tasks 3–5.
       - The drawer and button are still unstyled (task 6).
3. [x] Choosing a link in the dialog closes it.
   - Red: `Layout.test.tsx` › "closes the menu when a link in it is chosen" clicks the drawer's "Starter" logo link, so it doesn't depend on the sample. It fails because the dialog is still open. The test also checks that `aria-expanded` goes back to `"false"`.
   - Green: a delegated `onClick` on the `<dialog>` in `MobileNav`. If the click target is inside an `<a>`, it calls `setOpen(false)` and `dialogRef.current?.close()`. e2e not re-run: "Close the menu › go to the home page" still can't pass while the aside's nav is visible on phones (strict mode, see task 2), so that waits for task 6.
   - Refactor: `MobileNav` now has a `closeMenu()` (state + `dialog.close()`) mirroring `openMenu()`, and a named `closeOnLinkClick` handler on the dialog with a one-line comment on why it's delegated. Deliberately left: making `open` follow the dialog's native `close` event. That changes how state is driven, so it belongs to task 4's red/green, which will specify it. `closeMenu` is then the one place to adjust.
4. [x] The dialog's `close` event (what Escape triggers natively) closes the menu in React state too.
   - Red: `Layout.test.tsx` › "marks the menu closed when the browser closes the dialog, as Escape does" opens the menu, then calls `dialog.close()` directly inside `act`. The setup stand-in fires `close`, as a browser does after Escape. It fails because `aria-expanded` stays `"true"`: the dialog shuts but React state doesn't follow.
   - Green: `onClose={() => setOpen(false)}` on the `<dialog>`. e2e not re-run: "Close the menu › press Escape" is still blocked by the duplicate nav on phones until task 6.
   - Refactor: the dialog's `close` event is now the **single** path that syncs state. `closeMenu()` only calls `dialog.close()`, and a named `markClosed()` (the `onClose` handler) sets `open` to false. A comment above them says why. Behaviour note: real browsers fire `close` as a queued task after `close()`, so `aria-expanded` flips a tick later than before. The jsdom stand-in fires it synchronously, and e2e assertions auto-wait, so no test is affected. Deliberately left: nothing else in `MobileNav`. Task 5 will reuse `closeMenu`.
5. [ ] A click on the dialog's backdrop (target is the dialog itself) closes it. A click inside the panel doesn't.
6. [ ] CSS wiring, driven by e2e: aside `hidden md:flex`, top bar `md:hidden`, phone padding on `main`. Make "The menu starts closed", "Open the menu", "Close the menu" and "Change the style" green on `phone`.
7. [ ] Fix any sideways overflow the "Pages fit" outline finds (e.g. the items table) in each style × theme.
8. [ ] Run on `iphone` (WebKit) once `libavif16` is installed; fix WebKit-only differences.
9. [ ] Screenshot 3 styles × 2 themes at 390px for review.
10. [ ] Remove the `@wip` tags, add a phone-width line to CLAUDE.md "Web styles", and delete this file.

## Changes to the plan
- 2026-10-30: fixed `phone.steps.ts` to pass `page.evaluate` a string expression. `tsc -b` in `web:build` type-checks `e2e/` without DOM types, so `document` didn't compile. The scenario's meaning is unchanged.
