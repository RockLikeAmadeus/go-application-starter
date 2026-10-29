export type Theme = "dark" | "light";

/** Where the user's theme choice is kept in browser storage. */
const storageKey = "theme";

/** Picks the theme to show from the user's stored choice and the system preference. */
export function resolveTheme(stored: Theme | null, system: Theme | null): Theme {
  return stored ?? system ?? "dark";
}

/** Reads the user's saved theme choice, or null when there is none. */
export function readStoredTheme(storage: Pick<Storage, "getItem">): Theme | null {
  try {
    const value = storage.getItem(storageKey);
    return value === "dark" || value === "light" ? value : null;
  } catch {
    return null;
  }
}

/** Saves the user's theme choice so readStoredTheme returns it later. */
export function writeStoredTheme(theme: Theme, storage: Pick<Storage, "setItem">): void {
  try {
    storage.setItem(storageKey, theme);
  } catch {
    // Storage is unavailable, so the choice lasts only for this page view.
  }
}

/** Shows the given theme on the page by styling the root element. */
export function applyTheme(theme: Theme, root: HTMLElement): void {
  root.style.colorScheme = theme;
  root.classList.toggle("dark", theme === "dark");
}
