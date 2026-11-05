import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, unwrap } from "../../api/client";

const itemsKey = ["items"] as const;

export function useItems() {
  return useQuery({
    queryKey: itemsKey,
    queryFn: async () => unwrap(await api.GET("/api/v1/items")).items,
  });
}

function useItemsMutation<TVars>(fn: (vars: TVars) => Promise<unknown>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: fn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: itemsKey }),
  });
}

export function useCreateItem() {
  return useItemsMutation(async (name: string) =>
    unwrap(await api.POST("/api/v1/items", { body: { name } })),
  );
}

export function useRenameItem() {
  return useItemsMutation(async ({ id, name }: { id: string; name: string }) =>
    unwrap(await api.PATCH("/api/v1/items/{id}", { params: { path: { id } }, body: { name } })),
  );
}

export function useDeleteItem() {
  return useItemsMutation(async (id: string) =>
    unwrap(await api.DELETE("/api/v1/items/{id}", { params: { path: { id } } })),
  );
}

/**
 * Downloads the items as `items.csv`. The request goes through the API client,
 * so on Android it crosses the Wails bridge like every other call; the server
 * refuses it when the plan doesn't have item export.
 */
export function useExportItems() {
  return useMutation({
    mutationFn: async () =>
      saveFile(unwrap(await api.GET("/api/v1/items/export", { parseAs: "blob" })), "items.csv"),
  });
}

/** Hands the browser a file to save, through a temporary object URL. */
function saveFile(file: Blob, name: string) {
  const link = document.createElement("a");
  link.href = URL.createObjectURL(file);
  link.download = name;
  link.click();
  URL.revokeObjectURL(link.href);
}
