import type { Theme } from "./theme";

type Props = {
  theme: Theme;
  onChange: (theme: Theme) => void;
};

export function ThemeToggle({ theme, onChange }: Props) {
  return (
    <input
      type="checkbox"
      role="switch"
      aria-label="Dark mode"
      checked={theme === "dark"}
      onChange={(e) => onChange(e.target.checked ? "dark" : "light")}
    />
  );
}
