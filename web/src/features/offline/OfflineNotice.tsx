import { useSyncExternalStore } from "react";

function subscribe(onChange: () => void) {
  window.addEventListener("offline", onChange);
  return () => window.removeEventListener("offline", onChange);
}

export function OfflineNotice() {
  const online = useSyncExternalStore(subscribe, () => navigator.onLine);
  return (
    <p role="status">
      {!online && "You're offline. Data will load when the connection comes back."}
    </p>
  );
}
