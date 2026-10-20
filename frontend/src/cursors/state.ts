import { StateEffect, StateField } from "@codemirror/state";

export interface RemoteCursor {
  name: string;
  color: string;
  pos: number;
}

// Sets the cursor of one remote client, or removes it when `cursor` is null.
export const setCursor = StateEffect.define<{
  id: string;
  cursor: RemoteCursor | null;
}>();

// Remote cursors by clientID. Positions follow document changes.
export const cursorsField = StateField.define<ReadonlyMap<string, RemoteCursor>>(
  {
    create: () => new Map(),
    update(cursors, tr) {
      let next = new Map(cursors);
      if (tr.docChanged) {
        for (const [id, c] of next) {
          next.set(id, { ...c, pos: tr.changes.mapPos(c.pos) });
        }
      }
      for (const effect of tr.effects) {
        if (!effect.is(setCursor)) continue;
        const { id, cursor } = effect.value;
        // The sender's document may be ahead of ours; clamp to stay valid.
        if (cursor) {
          next.set(id, { ...cursor, pos: Math.min(cursor.pos, tr.newDoc.length) });
        } else {
          next.delete(id);
        }
      }
      return next;
    },
  },
);
