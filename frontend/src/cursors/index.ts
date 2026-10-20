import type { Extension } from "@codemirror/state";
import { cursorDecorations } from "./decorations";
import { cursorSync, type User } from "./plugin";
import { cursorsField } from "./state";

// Shows the cursors of other users. Must come after collabExtension: it relies
// on the collab state and on the collab plugin handling messages first.
export function cursorsExtension(socket: WebSocket, user: User): Extension {
  return [cursorsField, cursorDecorations, cursorSync(socket, user)];
}
