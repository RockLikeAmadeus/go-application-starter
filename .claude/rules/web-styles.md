---
paths:
  - "web/src/**"
  - "web/index.html"
  - "web/design-inspiration/**"
---

# Web styles

- A **style** (SPEC.md §4.3) is a named look for the SPA: layout, fonts, shape and density. It is separate from the dark/light **theme**, and every style must work in both.
- Styles and the shared colour palette live in `web/src/styles/`. The active style is the `data-style` attribute on `<html>`.
- Every page must work at phone width (SPEC.md §4.4), with no sideways scrolling in any style or theme. Below `md`, the sidebar becomes a Menu drawer (`MobileNav` in `web/src/routes/Layout.tsx`). In a flex row, give inputs `min-w-0` so they can shrink: WebKit's default input width is wider than Chromium's.
- Tag phone browser scenarios `@phone` (see `.claude/rules/web-e2e.md`).
- Colours: use only the semantic utilities from `palette.css` (`bg-panel`, `text-muted`, `border-line`, `bg-sidebar`, …). Tailwind's default colours are switched off, so `slate-500` and friends won't compile to anything.
- Keyboard focus: a base `:focus-visible` rule in `web/src/index.css` gives every control and link the palette's `focus` ring. Don't add per-control focus classes, and never remove the outline (`outline-none`) without putting an equally visible focus style in its place. Use real `<button>`, `<a>` and form elements, so they can take focus and get the ring.
- Style tokens: `font-sans`, `rounded-panel`, `rounded-control`, `py-row` and the `type-label` utility (headings, buttons, table headers) change per style. Reach for the `crisp:`, `terminal:` and `soft:` variants only where markup really must differ.
- Entry transitions (`starting:`): write the start value plainly, e.g. `starting:open:[translate:-100%_0]`, not with Tailwind's `translate-*`, `scale-*` or `rotate-*` utilities. Those read registered `--tw-*` variables, and WebKitGTK (the desktop app's engine) drops such a declaration inside `@starting-style`, so nothing animates in. Playwright's WebKit doesn't show this.
- To add a style: add it to `styles` in `web/src/features/style/style.ts`, create `web/src/styles/<id>.css` setting the `--style-*` variables under `:root[data-style="<id>"]` (self-host fonts with `@fontsource`), and import it in `web/src/index.css`.
- `web/design-inspiration/` holds the reference palette (`.hex`) and mockups for the current styles. A new project replaces it with its own. Colours come from the palette, not the mockups.
