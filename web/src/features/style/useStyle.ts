import { useEffect, useState } from "react";
import { applyStyle, readStoredStyle, resolveStyle, writeStoredStyle, type Style } from "./style";

/** The style shown on the page, and a way for the user to choose another. */
export function useStyle(): [Style, (style: Style) => void] {
  const [style, setStyle] = useState(() => resolveStyle(readStoredStyle(localStorage)));

  useEffect(() => {
    applyStyle(style, document.documentElement);
  }, [style]);

  const choose = (chosen: Style) => {
    writeStoredStyle(chosen, localStorage);
    setStyle(chosen);
  };

  return [style, choose];
}
