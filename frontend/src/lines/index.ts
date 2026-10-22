import type { Extension } from "@codemirror/state";
import { lineNumberGutter } from "./gutter";
import { lineSync } from "./plugin";
import { linesField } from "./state";

// Shows line numbers left of the text; clicking one shows the info of the line.
export function linesExtension(socket: WebSocket): Extension {
  return [linesField, lineSync(socket), lineNumberGutter];
}
