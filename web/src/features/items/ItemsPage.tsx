import { useState, type FormEvent } from "react";
import type { Item } from "../../api/client";
import { useCreateItem, useDeleteItem, useItems, useRenameItem } from "./queries";

export function ItemsPage() {
  const items = useItems();
  const createItem = useCreateItem();
  const [name, setName] = useState("");

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    createItem.mutate(name, { onSuccess: () => setName("") });
  };

  return (
    <section className="space-y-6 terminal:space-y-4">
      <h1 className="text-2xl font-semibold tracking-tight soft:text-3xl soft:font-bold">Items</h1>

      <form onSubmit={onSubmit} className="flex gap-2">
        <label htmlFor="new-item" className="sr-only">
          New item name
        </label>
        <input
          id="new-item"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="New item name"
          className="min-w-0 flex-1 rounded-control border border-line bg-panel px-3 py-2 placeholder:text-muted soft:px-5"
        />
        <button
          type="submit"
          disabled={createItem.isPending}
          className="type-label whitespace-nowrap rounded-control bg-accent px-4 py-2 font-medium text-accent-fg hover:opacity-90 disabled:opacity-50 terminal:text-sm soft:px-5"
        >
          Add item
        </button>
      </form>
      {createItem.error && <ErrorMessage error={createItem.error} />}

      {items.isPending && <p className="text-muted">Loading…</p>}
      {items.error && <ErrorMessage error={items.error} />}
      {items.data?.length === 0 && <p className="text-muted">No items yet.</p>}
      {items.data && items.data.length > 0 && (
        <div className="overflow-hidden rounded-panel border border-line bg-panel soft:border-transparent">
          <table aria-label="Items" className="w-full text-left">
            <thead>
              <tr className="border-b border-line bg-sunken text-sm text-muted terminal:text-xs soft:bg-transparent">
                <th className="type-label px-4 py-2.5 font-medium soft:px-6 soft:pt-5">Name</th>
                <th className="px-4 py-2.5">
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              {items.data.map((item) => (
                <ItemRow key={item.id} item={item} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function ItemRow({ item }: { item: Item }) {
  const deleteItem = useDeleteItem();
  const [editing, setEditing] = useState(false);

  if (editing) {
    return (
      <tr>
        <td colSpan={2} className="px-4 py-row soft:px-6">
          <RenameForm item={item} onDone={() => setEditing(false)} />
        </td>
      </tr>
    );
  }

  return (
    <tr className="hover:bg-sunken/60">
      <td className="px-4 py-row font-medium terminal:border-r terminal:border-line soft:px-6">
        {item.name}
      </td>
      <td className="px-4 py-row soft:px-6">
        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => setEditing(true)}
            aria-label={`Rename ${item.name}`}
            className="type-label text-sm text-muted hover:text-fg terminal:text-xs"
          >
            Rename
          </button>
          <button
            type="button"
            onClick={() => deleteItem.mutate(item.id)}
            aria-label={`Delete ${item.name}`}
            className="type-label text-sm text-danger hover:underline terminal:text-xs"
          >
            Delete
          </button>
        </div>
      </td>
    </tr>
  );
}

/** Edits an item's name in place; the draft starts from the current name each time it opens. */
function RenameForm({ item, onDone }: { item: Item; onDone: () => void }) {
  const renameItem = useRenameItem();
  const [draft, setDraft] = useState(item.name);

  const save = (e: FormEvent) => {
    e.preventDefault();
    renameItem.mutate({ id: item.id, name: draft }, { onSuccess: onDone });
  };

  return (
    <>
      <form onSubmit={save} className="flex gap-2">
        <input
          aria-label={`New name for ${item.name}`}
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          className="flex-1 rounded-control border border-line bg-panel px-2 py-1 soft:px-4"
          autoFocus
        />
        <button type="submit" className="type-label text-sm font-medium text-fg terminal:text-xs">
          Save
        </button>
        <button
          type="button"
          onClick={onDone}
          className="type-label text-sm text-muted terminal:text-xs"
        >
          Cancel
        </button>
      </form>
      {renameItem.error && <ErrorMessage error={renameItem.error} />}
    </>
  );
}

function ErrorMessage({ error }: { error: Error }) {
  return (
    <p role="alert" className="text-sm text-danger">
      {error.message}
    </p>
  );
}
