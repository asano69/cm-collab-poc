import { lineNumbers } from "@codemirror/view";
import { showLinePopup } from "./popup";
import { linesField } from "./state";

// The usual line numbers; clicking one shows the info of that line. Lines
// without known info (the text is ahead of the server) do nothing.
export const lineNumberGutter = lineNumbers({
  domEventHandlers: {
    click(view, line, event) {
      const number = view.state.doc.lineAt(line.from).number;
      const info = view.state.field(linesField)[number - 1];
      if (info === undefined) return false;
      const { clientX, clientY } = event as MouseEvent;
      showLinePopup(info, clientX, clientY);
      return true;
    },
  },
});
