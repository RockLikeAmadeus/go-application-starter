import { onlineManager } from "@tanstack/react-query";
import { useEffect } from "react";
import { useConnectionRequired } from "./useConnectionRequired";

/**
 * An onlineManager listener that ignores the browser's online and offline
 * events: local data is reachable whatever the network is doing.
 */
function alwaysOnline(setOnline: (online: boolean) => void) {
  setOnline(true);
  return () => {};
}

/**
 * Keeps queries and mutations running when the browser goes offline but the
 * data needs no connection (the desktop app's local-only mode).
 */
export function useLocalDataOnline(): void {
  const connectionRequired = useConnectionRequired();
  useEffect(() => {
    if (!connectionRequired) onlineManager.setEventListener(alwaysOnline);
  }, [connectionRequired]);
}
