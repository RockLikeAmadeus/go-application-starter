export function OfflineNotice() {
  return (
    <p role="status">
      {!navigator.onLine && "You're offline. Data will load when the connection comes back."}
    </p>
  );
}
