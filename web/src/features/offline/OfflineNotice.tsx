import { useOnline } from "./useOnline";

/**
 * Says the app is offline, on every page. The status region stays in the page
 * even when empty (and takes no space then), so screen readers announce the
 * message when it appears.
 */
export function OfflineNotice() {
  const online = useOnline();
  return (
    <div role="status">
      {!online && (
        <p className="mb-6 rounded-control border border-line bg-sunken px-4 py-2 text-sm text-muted terminal:mb-4 terminal:text-xs">
          You're offline. Data will load when the connection comes back.
        </p>
      )}
    </div>
  );
}
