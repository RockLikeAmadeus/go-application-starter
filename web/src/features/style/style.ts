/** The styles the app offers, in picker order. Projects replace this list with their own. */
export const styles = [
  { id: "crisp", name: "Crisp" },
  { id: "terminal", name: "Terminal" },
  { id: "soft", name: "Soft" },
] as const satisfies readonly { id: string; name: string }[];

export type Style = (typeof styles)[number]["id"];

/** The style shown when the user hasn't chosen one: the first in the list. */
export const defaultStyle: Style = styles[0].id;
