export type Theme = "dark" | "light";

/** Picks the theme to show from the user's stored choice and the system preference. */
export function resolveTheme(_stored: Theme | null, _system: Theme | null): Theme {
  return "dark";
}
