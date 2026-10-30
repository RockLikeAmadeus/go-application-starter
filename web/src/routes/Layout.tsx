import { Link, Outlet } from "react-router";
import { StylePicker } from "../features/style/StylePicker";
import { useStyle } from "../features/style/useStyle";
import { ThemeToggle } from "../features/theme/ThemeToggle";
import { useTheme } from "../features/theme/useTheme";

export function Layout() {
  const [theme, setTheme] = useTheme();
  const [style, setStyle] = useStyle();
  return (
    <div className="flex min-h-screen">
      <aside className="flex w-56 flex-col border-r border-slate-200 bg-white px-4 py-3 dark:border-slate-800 dark:bg-slate-900">
        <Link to="/" className="font-semibold text-violet-700 dark:text-violet-400">
          Starter
        </Link>
        <nav className="mt-6">
          <Link to="/" className="text-sm text-slate-700 dark:text-slate-300">
            Items
          </Link>
        </nav>
        <fieldset className="mt-auto flex flex-col gap-3">
          <legend className="sr-only">Appearance</legend>
          <StylePicker style={style} onChange={setStyle} />
          <ThemeToggle theme={theme} onChange={setTheme} />
        </fieldset>
      </aside>
      <main className="mx-auto max-w-3xl flex-1 px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}

export function NotFound() {
  return <p>Page not found.</p>;
}
