import {
  Decoration,
  EditorView,
  WidgetType,
} from "@codemirror/view";
import { cursorsField, type RemoteCursor } from "./state";

// A caret with a name flag above it.
class CursorWidget extends WidgetType {
  constructor(private readonly cursor: RemoteCursor) {
    super();
  }

  eq(other: CursorWidget) {
    return (
      other.cursor.name === this.cursor.name &&
      other.cursor.color === this.cursor.color
    );
  }

  toDOM() {
    const caret = document.createElement("span");
    caret.className = "cm-remote-cursor";
    caret.style.borderLeftColor = this.cursor.color;

    const flag = document.createElement("span");
    flag.className = "cm-remote-cursor-flag";
    flag.style.background = this.cursor.color;
    flag.textContent = this.cursor.name;

    caret.append(flag);
    return caret;
  }

  ignoreEvent() {
    return true;
  }
}

const theme = EditorView.baseTheme({
  ".cm-remote-cursor": {
    position: "relative",
    borderLeft: "2px solid",
    marginLeft: "-1px",
    marginRight: "-1px",
  },
  ".cm-remote-cursor-flag": {
    position: "absolute",
    bottom: "100%",
    left: "-2px",
    padding: "0 4px",
    fontSize: "11px",
    lineHeight: "16px",
    color: "white",
    whiteSpace: "nowrap",
    borderRadius: "3px 3px 3px 0",
    userSelect: "none",
  },
});

const decorations = EditorView.decorations.compute([cursorsField], (state) =>
  Decoration.set(
    [...state.field(cursorsField).values()].map((cursor) =>
      Decoration.widget({ widget: new CursorWidget(cursor), side: 1 }).range(
        cursor.pos,
      ),
    ),
    true,
  ),
);

export const cursorDecorations = [theme, decorations];
