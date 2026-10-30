import { useSyncExternalStore } from "react";

const connectionEvents = ["online", "offline"] as const;

function subscribe(onChange: () => void) {
  for (const event of connectionEvents) window.addEventListener(event, onChange);
  return () => {
    for (const event of connectionEvents) window.removeEventListener(event, onChange);
  };
}

/** Whether the browser has a connection, kept up to date as it drops and returns. */
export function useOnline(): boolean {
  return useSyncExternalStore(subscribe, () => navigator.onLine);
}
