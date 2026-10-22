import { EditorView, GutterMarker, gutter } from "@codemirror/view";
import { lineIDsField } from "./state";

class LineIDMarker extends GutterMarker {
  private readonly id: number;

  constructor(id: number) {
    super();
    this.id = id;
  }

  eq(other: LineIDMarker) {
    return other.id === this.id;
  }

  toDOM() {
    return document.createTextNode(String(this.id));
  }
}

const theme = EditorView.baseTheme({
  ".cm-line-id-gutter": {
    minWidth: "5ch",
  },
  ".cm-line-id-gutter .cm-gutterElement": {
    padding: "0 8px",
    textAlign: "right",
    color: "#888",
  },
});

// A column left of the text that shows the ID of each line. Lines without a
// known ID (the text is ahead of the server) stay blank.
const idGutter = gutter({
  class: "cm-line-id-gutter",
  lineMarker(view, line) {
    const number = view.state.doc.lineAt(line.from).number;
    const id = view.state.field(lineIDsField)[number - 1];
    return id === undefined ? null : new LineIDMarker(id);
  },
  // The gutter is redrawn on document changes by itself; this covers the IDs
  // changing without one.
  lineMarkerChange: (update) =>
    update.startState.field(lineIDsField) !== update.state.field(lineIDsField),
});

export const lineIDGutter = [theme, idGutter];
