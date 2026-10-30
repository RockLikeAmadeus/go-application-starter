import { styles, type Style } from "./style";

type Props = {
  style: Style;
  onChange: (style: Style) => void;
};

export function StylePicker({ style, onChange }: Props) {
  return (
    <label className="flex items-center justify-between gap-2 text-sm text-sidebar-muted">
      Style
      <select
        value={style}
        onChange={(event) => onChange(event.target.value as Style)}
        className="cursor-pointer rounded-control border border-sidebar-line bg-transparent px-2 py-1 text-sm text-sidebar-fg"
      >
        {styles.map((option) => (
          <option key={option.id} value={option.id} className="bg-panel text-fg">
            {option.name}
          </option>
        ))}
      </select>
    </label>
  );
}
