export type Theme = "dark" | "light";

/** Picks the theme to show from the user's stored choice and the system preference. */
export function resolveTheme(stored: Theme | null, system: Theme | null): Theme {
  return stored ?? system ?? "dark";
}

/** Reads the user's saved theme choice, or null when there is none. */
export function readStoredTheme(storage: Pick<Storage, "getItem">): Theme | null {
  try {
    return storage.getItem("theme") as Theme | null;
  } catch {
    return null;
  }
}
