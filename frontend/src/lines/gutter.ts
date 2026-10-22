import { lineNumbers } from "@codemirror/view";
import { showLinePopup } from "./popup";
import { linesField } from "./state";
import { verifyLine } from "./verify";

// The usual line numbers; clicking one shows the info of that line and whether
// its signature checks out. Lines without known info (the text is ahead of the
// server) do nothing.
export const lineNumberGutter = lineNumbers({
  domEventHandlers: {
    click(view, line, event) {
      const docLine = view.state.doc.lineAt(line.from);
      const info = view.state.field(linesField)[docLine.number - 1];
      if (info === undefined) return false;
      const { clientX, clientY } = event as MouseEvent;
      showLinePopup(info, verifyLine(info, docLine.text), clientX, clientY);
      return true;
    },
  },
});
