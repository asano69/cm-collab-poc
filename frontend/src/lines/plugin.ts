import { ViewPlugin, type EditorView } from "@codemirror/view";
import type { LinesMessage } from "./protocol";
import { setLines } from "./state";

// Stores the line info that the server sends over `socket`.
export function lineSync(socket: WebSocket) {
  return ViewPlugin.fromClass(
    class {
      private readonly view: EditorView;

      constructor(view: EditorView) {
        this.view = view;
        socket.addEventListener("message", this.onMessage);
      }

      destroy() {
        socket.removeEventListener("message", this.onMessage);
      }

      private onMessage = (event: MessageEvent<string>) => {
        const data = JSON.parse(event.data);
        // Other plugins share this socket.
        if (data.type !== "lines") return;
        const msg: LinesMessage = data;
        this.view.dispatch({ effects: setLines.of(msg.lines) });
      };
    },
  );
}
