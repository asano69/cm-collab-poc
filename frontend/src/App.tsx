import { Show, createSignal, onCleanup, onMount } from "solid-js";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { EditorState } from "@codemirror/state";
import { EditorView, keymap } from "@codemirror/view";
import { collabExtension, connect, cursorsExtension } from "cm-websocket";
import { linesExtension } from "@/lines";

// A throwaway display name; there is no login in this proof of concept.
const userName = `User ${Math.floor(Math.random() * 900) + 100}`;

function collabURL(): string {
  const scheme = location.protocol === "https:" ? "wss" : "ws";
  return `${scheme}://${location.host}/connection/collab`;
}

function App() {
  const [error, setError] = createSignal("");
  let parent!: HTMLDivElement;

  onMount(() => {
    const channel = connect(collabURL());
    channel.socket.addEventListener("close", (event) => {
      const reason = event.reason ? `: ${event.reason}` : "";
      setError(`Connection lost${reason}. Reload the page to reconnect.`);
    });

    const view = new EditorView({
      parent,
      state: EditorState.create({
        extensions: [
          history(),
          keymap.of([...defaultKeymap, ...historyKeymap]),
          collabExtension(channel),
          cursorsExtension(channel, { name: userName }),
          linesExtension(channel),
        ],
      }),
    });

    onCleanup(() => {
      channel.socket.close();
      view.destroy();
    });
  });

  return (
    <div class="app">
      <Show when={error()}>
        <p class="error">{error()}</p>
      </Show>
      <h1>Collaborative real-time editor</h1>
      <div class="editor" ref={parent} />
    </div>
  );
}

export default App;
