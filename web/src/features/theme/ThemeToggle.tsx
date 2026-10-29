import type { Theme } from "./theme";

type Props = {
  theme: Theme;
  onChange: (theme: Theme) => void;
};

export function ThemeToggle({ theme, onChange }: Props) {
  return (
    <label className="flex cursor-pointer items-center gap-2 text-sm text-slate-600 dark:text-slate-300">
      <input
        type="checkbox"
        role="switch"
        checked={theme === "dark"}
        onChange={(e) => onChange(e.target.checked ? "dark" : "light")}
        className="relative h-5 w-9 cursor-pointer appearance-none rounded-full bg-slate-300 transition-colors before:absolute before:top-0.5 before:left-0.5 before:size-4 before:rounded-full before:bg-white before:transition-transform checked:bg-violet-600 checked:before:translate-x-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-violet-500 dark:bg-slate-600 dark:checked:bg-violet-500"
      />
      Dark mode
    </label>
  );
}
