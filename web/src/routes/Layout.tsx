import { Link, Outlet } from "react-router";
import { ThemeToggle } from "../features/theme/ThemeToggle";
import { useTheme } from "../features/theme/useTheme";

export function Layout() {
  const [theme, setTheme] = useTheme();
  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
        <nav className="mx-auto flex max-w-3xl items-center gap-6 px-4 py-3">
          <Link to="/" className="font-semibold text-violet-700 dark:text-violet-400">
            Starter
          </Link>
          <div className="ml-auto">
            <ThemeToggle theme={theme} onChange={setTheme} />
          </div>
        </nav>
      </header>
      <main className="mx-auto max-w-3xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}

export function NotFound() {
  return <p>Page not found.</p>;
}
