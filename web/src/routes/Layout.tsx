import { Link, NavLink, Outlet } from "react-router";
import { StylePicker } from "../features/style/StylePicker";
import { useStyle } from "../features/style/useStyle";
import { ThemeToggle } from "../features/theme/ThemeToggle";
import { useTheme } from "../features/theme/useTheme";

export function Layout() {
  const [theme, setTheme] = useTheme();
  const [style, setStyle] = useStyle();
  return (
    <div className="flex min-h-screen">
      <aside className="sticky top-0 flex h-screen w-60 shrink-0 flex-col border-r border-sidebar-line bg-sidebar px-3 py-4 text-sidebar-fg soft:m-3 soft:h-[calc(100vh-1.5rem)] soft:rounded-panel soft:border-r-0 soft:px-4 soft:py-5">
        <Link to="/" className="flex items-center gap-3 px-2 text-lg font-semibold">
          <span
            aria-hidden="true"
            className="size-7 rounded-control bg-sidebar-accent soft:rounded-xl"
          />
          <span className="type-label">Starter</span>
        </Link>
        <nav className="mt-8 flex flex-col gap-1">
          <NavLink
            to="/"
            end
            className="flex items-center gap-3 rounded-control px-3 py-2 text-sidebar-muted hover:text-sidebar-fg aria-[current=page]:bg-sidebar-active aria-[current=page]:text-sidebar-active-fg terminal:py-1.5 soft:aria-[current=page]:bg-sidebar-accent soft:aria-[current=page]:text-ink"
          >
            <ListIcon />
            Items
          </NavLink>
        </nav>
        <fieldset className="mt-auto flex flex-col gap-3 border-t border-sidebar-line px-2 pt-4 soft:border-t-0">
          <legend className="sr-only">Appearance</legend>
          <StylePicker style={style} onChange={setStyle} />
          <ThemeToggle theme={theme} onChange={setTheme} />
        </fieldset>
      </aside>
      <main className="min-w-0 flex-1 px-8 py-8 terminal:px-5 terminal:py-5">
        <div className="mx-auto max-w-4xl">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

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

export function NotFound() {
  return <p>Page not found.</p>;
}
