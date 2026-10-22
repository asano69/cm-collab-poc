import { StateEffect, StateField } from "@codemirror/state";
import type { LineInfo } from "./protocol";

// Replaces the known line info.
export const setLines = StateEffect.define<readonly LineInfo[]>();

// The line info of the server's latest message; the info of line n is at n - 1.
export const linesField = StateField.define<readonly LineInfo[]>({
  create: () => [],
  update(lines, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setLines)) return effect.value;
    }
    return lines;
  },
});
