import type { Theme } from "./theme";

type Props = {
  theme: Theme;
  onChange: (theme: Theme) => void;
};

export function ThemeToggle({ theme, onChange }: Props) {
  return (
    <label className="flex cursor-pointer items-center justify-between gap-2 text-sm text-sidebar-muted">
      Dark mode
      <input
        type="checkbox"
        role="switch"
        checked={theme === "dark"}
        onChange={(e) => onChange(e.target.checked ? "dark" : "light")}
        className="relative h-5 w-9 cursor-pointer appearance-none rounded-full bg-sidebar-muted/50 transition-colors before:absolute before:top-0.5 before:left-0.5 before:size-4 before:rounded-full before:bg-white before:transition-transform checked:bg-sidebar-accent checked:before:translate-x-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus"
      />
    </label>
  );
}
