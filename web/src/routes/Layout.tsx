import { Link, Outlet } from "react-router";

export function Layout() {
  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <nav className="mx-auto flex max-w-3xl items-center gap-6 px-4 py-3">
          <Link to="/" className="font-semibold text-violet-700">
            Starter
          </Link>
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
