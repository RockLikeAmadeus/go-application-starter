import { type ReactNode, useId } from "react";
import { useFeature } from "./useFeature";

/** What a gated control is given to spread on itself: `<button {...control}>`. */
export type GatedControlProps = { disabled: boolean; "aria-describedby"?: string };

/**
 * Shows a control by the plan's feature with this key (e.g. "item-export"):
 * as it is when on, disabled with a note when locked ("Available on the pro
 * plan", or "Not available" with no hint), and not at all when hidden. Until
 * the entitlements arrive it renders the control disabled, so the page keeps
 * its size and nothing pops in. The server still refuses a feature that's off.
 */
export function FeatureGate({
  feature,
  children,
}: {
  feature: string;
  children: (control: GatedControlProps) => ReactNode;
}) {
  const state = useFeature(feature);
  const noteId = useId();
  if (state.state === "hidden") return null;
  if (state.state === "on") return children({ disabled: false });
  if (state.state === "pending") return children({ disabled: true });
  return (
    <>
      {children({ disabled: true, "aria-describedby": noteId })}
      <p id={noteId} className="text-sm text-muted">
        {lockedNote(state.availableOn)}
      </p>
    </>
  );
}

/** "Available on the pro plan", "… the pro or team plans", or "Not available". */
function lockedNote(plans: string[]): string {
  if (plans.length === 0) return "Not available";
  if (plans.length === 1) return `Available on the ${plans[0]} plan`;
  const last = plans.length - 1;
  return `Available on the ${plans.slice(0, last).join(", ")} or ${plans[last]} plans`;
}
