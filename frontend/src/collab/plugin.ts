import {
  collab,
  getSyncedVersion,
  receiveUpdates,
  sendableUpdates,
} from "@codemirror/collab";
import type { Extension } from "@codemirror/state";
import { ViewPlugin, type EditorView, type ViewUpdate } from "@codemirror/view";
import {
  fromWire,
  toWire,
  type PushMessage,
  type UpdatesMessage,
} from "./protocol";

// Connects the editor to the server over `socket`. The document starts empty
// and is rebuilt by replaying the server's update history.
export function collabExtension(socket: WebSocket): Extension {
  const clientID = Math.random().toString(36).slice(2);

  const plugin = ViewPlugin.fromClass(
    class {
      private readonly view: EditorView;
      // True while a push is in flight; only one is allowed at a time.
      private pushing = false;

      constructor(view: EditorView) {
        this.view = view;
        socket.addEventListener("open", this.push);
        socket.addEventListener("message", this.onMessage);
      }

      update(update: ViewUpdate) {
        if (update.docChanged) this.push();
      }

      destroy() {
        socket.removeEventListener("open", this.push);
        socket.removeEventListener("message", this.onMessage);
      }

      private push = () => {
        const state = this.view.state;
        const updates = sendableUpdates(state);
        if (this.pushing || updates.length === 0) return;
        if (socket.readyState !== WebSocket.OPEN) return;

        this.pushing = true;
        const msg: PushMessage = {
          version: getSyncedVersion(state),
          updates: toWire(updates),
        };
        socket.send(JSON.stringify(msg));
      };

      private onMessage = (event: MessageEvent<string>) => {
        const msg: UpdatesMessage = JSON.parse(event.data);
        const state = this.view.state;

        // Skip updates that were already applied.
        const skip = getSyncedVersion(state) - msg.from;
        if (skip < 0) {
          socket.close(4000, "editor is out of sync with the server");
          return;
        }

        this.pushing = false;
        const fresh = fromWire(msg.updates).slice(skip);
        if (fresh.length > 0) this.view.dispatch(receiveUpdates(state, fresh));
        // Retry anything that was rejected or typed in the meantime.
        this.push();
      };
    },
  );

  return [collab({ clientID }), plugin];
}
