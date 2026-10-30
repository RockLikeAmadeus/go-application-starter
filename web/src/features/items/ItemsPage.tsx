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
    <section className="space-y-6">
      <h1 className="text-2xl font-semibold">Items</h1>

      <form onSubmit={onSubmit} className="flex gap-2">
        <label htmlFor="new-item" className="sr-only">
          New item name
        </label>
        <input
          id="new-item"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="New item name"
          className="flex-1 rounded-md border border-slate-300 px-3 py-2 focus:outline-2 focus:outline-violet-500 dark:border-slate-700 dark:bg-slate-900"
        />
        <button
          type="submit"
          disabled={createItem.isPending}
          className="rounded-md bg-violet-600 px-4 py-2 font-medium text-white hover:bg-violet-700 disabled:opacity-50"
        >
          Add item
        </button>
      </form>
      {createItem.error && <ErrorMessage error={createItem.error} />}

      {items.isPending && <p className="text-slate-500 dark:text-slate-400">Loading…</p>}
      {items.error && <ErrorMessage error={items.error} />}
      {items.data?.length === 0 && (
        <p className="text-slate-500 dark:text-slate-400">No items yet.</p>
      )}
      {items.data && items.data.length > 0 && (
        <table
          aria-label="Items"
          className="w-full rounded-md border border-slate-200 bg-white text-left dark:border-slate-800 dark:bg-slate-900"
        >
          <thead>
            <tr className="border-b border-slate-200 text-sm text-slate-500 dark:border-slate-800 dark:text-slate-400">
              <th className="px-4 py-2 font-medium">Name</th>
              <th className="px-4 py-2">
                <span className="sr-only">Actions</span>
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200 dark:divide-slate-800">
            {items.data.map((item) => (
              <ItemRow key={item.id} item={item} />
            ))}
          </tbody>
        </table>
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
        <td colSpan={2} className="px-4 py-3">
          <RenameForm item={item} onDone={() => setEditing(false)} />
        </td>
      </tr>
    );
  }

  return (
    <tr>
      <td className="px-4 py-3">{item.name}</td>
      <td className="px-4 py-3">
        <div className="flex justify-end gap-3">
          <button
            type="button"
            onClick={() => setEditing(true)}
            aria-label={`Rename ${item.name}`}
            className="text-sm text-slate-600 hover:text-violet-700 dark:text-slate-300 dark:hover:text-violet-400"
          >
            Rename
          </button>
          <button
            type="button"
            onClick={() => deleteItem.mutate(item.id)}
            aria-label={`Delete ${item.name}`}
            className="text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
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
          className="flex-1 rounded-md border border-slate-300 px-2 py-1 dark:border-slate-700 dark:bg-slate-800"
          autoFocus
        />
        <button type="submit" className="text-sm font-medium text-violet-700 dark:text-violet-400">
          Save
        </button>
        <button
          type="button"
          onClick={onDone}
          className="text-sm text-slate-500 dark:text-slate-400"
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
    <p role="alert" className="text-sm text-red-600 dark:text-red-400">
      {error.message}
    </p>
  );
}
