import { queryOptions, useQuery } from "@tanstack/react-query";
import { api, unwrap } from "../../api/client";

const sessionQuery = queryOptions({
  queryKey: ["session"] as const,
  queryFn: async () => unwrap(await api.GET("/api/v1/session")),
});

/**
 * Whether the server signs people in, and who is signed in (`name`, missing
 * when no one is). Undefined until the server answers, or if it can't be
 * asked: the app then shows as it would with sign-in off, and the server
 * still refuses any data to someone signed out.
 */
export function useSession() {
  return useQuery(sessionQuery).data;
}
