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
