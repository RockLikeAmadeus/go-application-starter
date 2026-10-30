import { styles, type Style } from "./style";

type Props = {
  style: Style;
  onChange: (style: Style) => void;
};

export function StylePicker({ style, onChange }: Props) {
  return (
    <label className="flex items-center justify-between gap-2 text-sm text-slate-600 dark:text-slate-300">
      Style
      <select
        value={style}
        onChange={(event) => onChange(event.target.value as Style)}
        className="rounded border border-slate-300 bg-white px-2 py-1 text-sm text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-violet-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
      >
        {styles.map((option) => (
          <option key={option.id} value={option.id}>
            {option.name}
          </option>
        ))}
      </select>
    </label>
  );
}
