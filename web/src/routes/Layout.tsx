// sample:start
import { NavLink } from "react-router";
// sample:end
import { usePrefetchQuery } from "@tanstack/react-query";
import { type MouseEvent, type ReactNode, type SyntheticEvent, useRef, useState } from "react";
import { Link, Outlet } from "react-router";
import { afterTransition } from "../lib/afterTransition";
import { entitlementsQuery } from "../features/entitlements/useFeature";
import { OfflineNotice } from "../features/offline/OfflineNotice";
import { useLocalDataOnline } from "../features/offline/useLocalDataOnline";
import type { Style } from "../features/style/style";
import { StylePicker } from "../features/style/StylePicker";
import { useStyle } from "../features/style/useStyle";
import type { Theme } from "../features/theme/theme";
import { ThemeToggle } from "../features/theme/ThemeToggle";
import { useTheme } from "../features/theme/useTheme";
import { WarningsNotice } from "../features/warnings/WarningsNotice";

export function Layout() {
  useLocalDataOnline();
  usePrefetchQuery(entitlementsQuery);
  const [theme, setTheme] = useTheme();
  const [style, setStyle] = useStyle();
  const content = (
    <SidebarContent theme={theme} onThemeChange={setTheme} style={style} onStyleChange={setStyle} />
  );
  return (
    <div className="flex min-h-screen">
      <aside className="sticky top-0 hidden h-screen w-60 shrink-0 border-r border-sidebar-line bg-sidebar text-sidebar-fg md:block soft:m-3 soft:h-[calc(100vh-1.5rem)] soft:rounded-panel soft:border-r-0">
        {content}
      </aside>
      <main className="min-w-0 flex-1 px-8 py-8 terminal:px-5 terminal:py-5">
        <MobileNav>{content}</MobileNav>
        <OfflineNotice />
        <WarningsNotice />
        <div className="mx-auto max-w-4xl">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

// The drawer slides in from the left while its backdrop fades in (`starting:`
// gives the entry transition its start). Closing is driven from script: the
// dialog gets `data-closing`, slides out, and only then really closes. That
// works in every browser, unlike CSS exit transitions on the top layer.
// The off-screen position is a plain `translate` value, not Tailwind's
// `-translate-x-full`: that reads `--tw-translate-y`, which is registered and
// unset, and WebKitGTK 2.52 (the desktop app's engine) then drops the whole
// declaration inside `@starting-style`, so the drawer appeared without sliding in.
const drawerMotionClasses = [
  "transition-[translate] duration-200 ease-out",
  "starting:open:[translate:-100%_0] data-closing:[translate:-100%_0]",
  "backdrop:bg-black/50 backdrop:transition-[background-color] backdrop:duration-200",
  "starting:open:backdrop:bg-transparent data-closing:backdrop:bg-transparent",
  "motion-reduce:duration-0 motion-reduce:backdrop:duration-0",
].join(" ");

/** The phone-width menu: a button that opens the sidebar's content in a modal drawer. */
function MobileNav({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const dialogRef = useRef<HTMLDialogElement>(null);

  function openMenu() {
    setIsOpen(true);
    dialogRef.current?.showModal();
  }

  // Every way of closing (a link, Escape, the backdrop) comes here: slide out,
  // then close the dialog. The browser then fires `close`, and markClosed keeps
  // `isOpen` in step.
  function closeMenu() {
    const dialog = dialogRef.current;
    if (!dialog) return;
    dialog.dataset.closing = "";
    afterTransition(dialog, () => {
      delete dialog.dataset.closing;
      dialog.close();
    });
  }

  function markClosed() {
    setIsOpen(false);
  }

  // Going somewhere from the menu, or tapping outside it, closes it. One handler
  // on the dialog covers every link the content holds. A click on the backdrop
  // targets the <dialog> itself rather than anything inside it.
  function closeOnLinkOrBackdropClick(event: MouseEvent) {
    const isBackdrop = event.target === event.currentTarget;
    const isLink = event.target instanceof Element && event.target.closest("a") !== null;
    if (isBackdrop || isLink) {
      closeMenu();
    }
  }

  // Escape fires `cancel` before the browser closes the dialog at once. Close
  // it our way instead, so it slides out.
  function closeOnEscape(event: SyntheticEvent) {
    event.preventDefault();
    closeMenu();
  }

  return (
    <>
      <button
        type="button"
        aria-expanded={isOpen}
        className="type-label mb-6 rounded-control border border-line px-3 py-1.5 text-sm text-muted hover:text-fg md:hidden terminal:mb-4 terminal:text-xs"
        onClick={openMenu}
      >
        Menu
      </button>
      <dialog
        ref={dialogRef}
        onClick={closeOnLinkOrBackdropClick}
        onClose={markClosed}
        onCancel={closeOnEscape}
        className={`m-0 h-dvh max-h-none w-72 max-w-[85vw] border-r border-sidebar-line bg-sidebar p-0 text-sidebar-fg soft:rounded-r-panel ${drawerMotionClasses}`}
      >
        {children}
      </dialog>
    </>
  );
}

type SidebarContentProps = {
  theme: Theme;
  onThemeChange: (theme: Theme) => void;
  style: Style;
  onStyleChange: (style: Style) => void;
};

/**
 * Logo, navigation, and Appearance controls, shared by the sidebar and the phone
 * menu. It fills its container, so in the menu only a tap outside it reaches
 * the dialog itself (which counts as the backdrop).
 */
function SidebarContent({ theme, onThemeChange, style, onStyleChange }: SidebarContentProps) {
  return (
    <div className="flex h-full flex-col px-3 py-4 soft:px-4 soft:py-5">
      <Link to="/" className="flex items-center gap-3 px-2 text-lg font-semibold">
        <span
          aria-hidden="true"
          className="size-7 rounded-control bg-sidebar-accent soft:rounded-xl"
        />
        <span className="type-label">Starter</span>
      </Link>
      <nav className="mt-8 flex flex-col gap-1">
        {/* Add a NavLink here for each capability's page, with an icon and a short label. */}
        {/* sample:start */}
        <NavLink
          to="/"
          end
          className="flex items-center gap-3 rounded-control px-3 py-2 text-sidebar-muted hover:text-sidebar-fg aria-[current=page]:bg-sidebar-active aria-[current=page]:text-sidebar-active-fg terminal:py-1.5 soft:aria-[current=page]:bg-sidebar-accent soft:aria-[current=page]:text-ink"
        >
          <ListIcon />
          Items
        </NavLink>
        {/* sample:end */}
      </nav>
      <fieldset className="mt-auto flex flex-col gap-3 border-t border-sidebar-line px-2 pt-4 soft:border-t-0">
        <legend className="sr-only">Appearance</legend>
        <StylePicker style={style} onChange={onStyleChange} />
        <ThemeToggle theme={theme} onChange={onThemeChange} />
      </fieldset>
    </div>
  );
}
// sample:start

function ListIcon() {
  return (
    <svg
      aria-hidden="true"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      className="size-4"
    >
      <path d="M7 5h10M7 10h10M7 15h10M3 5h.01M3 10h.01M3 15h.01" />
    </svg>
  );
}
// sample:end

/** The first page, until the app has capabilities of its own. */
export function Home() {
  return (
    <section className="space-y-6 terminal:space-y-4">
      <h1 className="text-2xl font-semibold tracking-tight soft:text-3xl soft:font-bold">Home</h1>
      <p className="text-muted">Nothing here yet.</p>
    </section>
  );
}

export function NotFound() {
  return <p>Page not found.</p>;
}
