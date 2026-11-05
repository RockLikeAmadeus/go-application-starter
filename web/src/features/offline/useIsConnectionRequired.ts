import { useQuery } from "@tanstack/react-query";
import { api, unwrap } from "../../api/client";

const connectionKey = ["connection"] as const;

/**
 * Whether the app needs a network connection to reach its data. False when
 * the data is on this device (the desktop app's local-only mode).
 */
export function useIsConnectionRequired(): boolean {
  const { data } = useQuery({
    queryKey: connectionKey,
    // Ask even when the browser says it's offline: local data is still
    // reachable then, and that's exactly when the answer matters.
    networkMode: "always",
    queryFn: async () => unwrap(await api.GET("/api/v1/connection")),
  });
  return data?.isRequired ?? true;
}
