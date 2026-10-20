import { sendableUpdates } from "@codemirror/collab";
import { ViewPlugin, type EditorView, type ViewUpdate } from "@codemirror/view";
import type { CursorMessage } from "./protocol";
import { setCursor } from "./state";

export interface User {
  name: string;
}

// Stable color per client, derived from its ID.
function colorOf(id: string): string {
  let hash = 0;
  for (const ch of id) hash = (hash * 31 + ch.charCodeAt(0)) >>> 0;
  return `hsl(${hash % 360} 70% 40%)`;
}

// Sends the local cursor over `socket` and applies the ones received.
export function cursorSync(socket: WebSocket, user: User) {
  const clientID = Math.random().toString(36).slice(2);
  const color = colorOf(clientID);

  return ViewPlugin.fromClass(
    class {
      private readonly view: EditorView;

      constructor(view: EditorView) {
        this.view = view;
        socket.addEventListener("open", this.send);
        socket.addEventListener("message", this.onMessage);
      }

      update(update: ViewUpdate) {
        if (update.selectionSet || update.docChanged) this.send();
      }

      destroy() {
        socket.removeEventListener("open", this.send);
        socket.removeEventListener("message", this.onMessage);
      }

      // The cursor is only sent while no local update is unconfirmed. Others
      // then already have the text the position refers to, because the server
      // relays an accepted update before the cursor that follows it.
      private send = () => {
        const state = this.view.state;
        if (socket.readyState !== WebSocket.OPEN) return;
        if (sendableUpdates(state).length > 0) return;

        const msg: CursorMessage = {
          type: "cursor",
          clientID,
          name: user.name,
          color,
          pos: state.selection.main.head,
        };
        socket.send(JSON.stringify(msg));
      };

      private onMessage = (event: MessageEvent<string>) => {
        const msg = JSON.parse(event.data);
        switch (msg.type) {
          case "updates":
            // This listener runs after the collab plugin's, so our updates
            // may just have been confirmed: retry sending the cursor.
            this.send();
            break;
          case "cursor":
            this.view.dispatch({
              effects: setCursor.of({
                id: msg.clientID,
                cursor: { name: msg.name, color: msg.color, pos: msg.pos },
              }),
            });
            break;
          case "leave":
            this.view.dispatch({
              effects: setCursor.of({ id: msg.clientID, cursor: null }),
            });
            break;
        }
      };
    },
  );
}
