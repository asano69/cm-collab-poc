import { ViewPlugin, type EditorView } from "@codemirror/view";
import type { Channel } from "cm-websocket";
import type { LinesMessage } from "./protocol";
import { setLines } from "./state";

// Stores the line info that the server sends over `channel`.
export function lineSync(channel: Channel) {
  return ViewPlugin.fromClass(
    class {
      private readonly view: EditorView;
      private readonly unsubscribe: () => void;

      constructor(view: EditorView) {
        this.view = view;
        this.unsubscribe = channel.on<LinesMessage>("lines", this.onLines);
      }

      destroy() {
        this.unsubscribe();
      }

      private onLines = (msg: LinesMessage) => {
        this.view.dispatch({ effects: setLines.of(msg.lines) });
      };
    },
  );
}
