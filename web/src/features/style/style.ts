/** The styles the app offers, in picker order. Projects replace this list with their own. */
export const styles = [
  { id: "crisp", name: "Crisp" },
  { id: "terminal", name: "Terminal" },
  { id: "soft", name: "Soft" },
] as const satisfies readonly { id: string; name: string }[];

export type Style = (typeof styles)[number]["id"];

/** The style shown when the user hasn't chosen one: the first in the list. */
export const defaultStyle: Style = styles[0].id;

/** Where the user's style choice is kept in browser storage. */
const storageKey = "style";

/** Whether a value names one of the styles the app offers. */
function isStyle(value: string | null): value is Style {
  return styles.some((style) => style.id === value);
}

/** Picks the style to show from the user's stored choice, falling back to the default when it isn't offered. */
export function resolveStyle(stored: string | null): Style {
  return isStyle(stored) ? stored : defaultStyle;
}

/** Reads the user's saved style choice, or null when there is none or it is no longer offered. */
export function readStoredStyle(storage: Pick<Storage, "getItem">): Style | null {
  try {
    const value = storage.getItem(storageKey);
    return isStyle(value) ? value : null;
  } catch {
    return null;
  }
}

/** Saves the user's style choice so readStoredStyle returns it later. */
export function writeStoredStyle(style: Style, storage: Pick<Storage, "setItem">): void {
  try {
    storage.setItem(storageKey, style);
  } catch {
    // Storage is unavailable; the choice lasts only for this page view.
  }
}

/** Shows the given style on the page: the stylesheets in src/styles key off the root's data-style. */
export function applyStyle(style: Style, root: HTMLElement): void {
  root.dataset.style = style;
}
