import { useEffect, useState } from "react";
import { applyTheme, readStoredTheme, resolveTheme, writeStoredTheme, type Theme } from "./theme";

/** The theme shown on the page, and a way for the user to choose another. */
export function useTheme(): [Theme, (theme: Theme) => void] {
  const [theme, setTheme] = useState(() => {
    const system = matchMedia("(prefers-color-scheme: light)").matches ? "light" : null;
    return resolveTheme(readStoredTheme(localStorage), system);
  });

  useEffect(() => {
    applyTheme(theme, document.documentElement);
  }, [theme]);

  const choose = (chosen: Theme) => {
    writeStoredTheme(chosen, localStorage);
    setTheme(chosen);
  };

  return [theme, choose];
}
