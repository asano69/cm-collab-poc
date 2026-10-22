import { StateEffect, StateField } from "@codemirror/state";

// Replaces the known line IDs.
export const setLineIDs = StateEffect.define<readonly number[]>();

// The line IDs of the server's latest message; the ID of line n is at n - 1.
export const lineIDsField = StateField.define<readonly number[]>({
  create: () => [],
  update(ids, tr) {
    for (const effect of tr.effects) {
      if (effect.is(setLineIDs)) return effect.value;
    }
    return ids;
  },
});
