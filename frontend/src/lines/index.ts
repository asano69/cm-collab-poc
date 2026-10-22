import type { Extension } from "@codemirror/state";
import { lineIDGutter } from "./gutter";
import { lineIDSync } from "./plugin";
import { lineIDsField } from "./state";

// Shows the ID of every line in a column left of the text.
export function lineIDsExtension(socket: WebSocket): Extension {
  return [lineIDsField, lineIDSync(socket), lineIDGutter];
}
