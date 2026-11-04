import { noticeClass } from "../../components/notice";
import { useConnectionRequired } from "./useConnectionRequired";
import { useOnline } from "./useOnline";

/**
 * Says the app is offline, on every page, unless its data needs no connection
 * (the desktop app's local-only mode). The status region stays in the page
 * even when empty (and takes no space then), so screen readers announce the
 * message when it appears.
 */
export function OfflineNotice() {
  const online = useOnline();
  const connectionRequired = useConnectionRequired();
  return (
    <div role="status">
      {connectionRequired && !online && (
        <p className={noticeClass}>
          You're offline. Data will load when the connection comes back.
        </p>
      )}
    </div>
  );
}
