import type { Extension } from "@codemirror/state";
import type { Channel } from "cm-websocket";
import { lineNumberGutter } from "./gutter";
import { lineSync } from "./plugin";
import { linesField } from "./state";

// Shows line numbers left of the text; clicking one shows the info of the line.
export function linesExtension(channel: Channel): Extension {
  return [linesField, lineSync(channel), lineNumberGutter];
}
