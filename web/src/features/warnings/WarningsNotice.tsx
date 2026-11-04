import { useQuery } from "@tanstack/react-query";
import { api, unwrap } from "../../api/client";

const warningsKey = ["warnings"] as const;

/**
 * Shows the app's warnings (e.g. that the data folder looks synced), each
 * message as the server gives it. Nothing shows until the server answers.
 */
export function WarningsNotice() {
  const { data } = useQuery({
    queryKey: warningsKey,
    queryFn: async () => unwrap(await api.GET("/api/v1/warnings")),
  });
  if (!data) return null;
  return (
    <section aria-label="Warnings">
      {data.warnings.map(({ message }) => (
        <p key={message}>{message}</p>
      ))}
    </section>
  );
}
