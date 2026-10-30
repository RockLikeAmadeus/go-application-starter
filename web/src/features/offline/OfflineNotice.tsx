import { useOnline } from "./useOnline";

export function OfflineNotice() {
  const online = useOnline();
  return (
    <p role="status">
      {!online && "You're offline. Data will load when the connection comes back."}
    </p>
  );
}
