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
1. [ ] `Layout` shows a "Menu" button with `aria-expanded="false"`, and no dialog is open at first.
2. [ ] Pressing "Menu" opens a dialog that contains the navigation and the Appearance group, and sets `aria-expanded="true"`. (Check jsdom's `showModal` support first; stub it in the test if it's missing.)
3. [ ] Choosing a link in the dialog closes it.
4. [ ] The dialog's `close` event (what Escape triggers natively) closes the menu in React state too.
5. [ ] A click on the dialog's backdrop (target is the dialog itself) closes it. A click inside the panel doesn't.
6. [ ] CSS wiring, driven by e2e: aside `hidden md:flex`, top bar `md:hidden`, phone padding on `main`. Make "The menu starts closed", "Open the menu", "Close the menu" and "Change the style" green on `phone`.
7. [ ] Fix any sideways overflow the "Pages fit" outline finds (e.g. the items table) in each style × theme.
8. [ ] Run on `iphone` (WebKit) once `libavif16` is installed; fix WebKit-only differences.
9. [ ] Screenshot 3 styles × 2 themes at 390px for review.
10. [ ] Remove the `@wip` tags, add a phone-width line to CLAUDE.md "Web styles", and delete this file.

## Changes to the plan
_(none yet)_
