import { EditorState } from "@codemirror/state";
import { expect, test } from "vitest";
import type { LineInfo } from "./protocol";
import { linesField, setLines } from "./state";

const info: LineInfo = { id: "x", updated: 7, admin: "Bob" };

test("setLines replaces the known line info", () => {
  const state = EditorState.create({ extensions: [linesField] });
  expect(state.field(linesField)).toEqual([]);

  const next = state.update({ effects: setLines.of([info]) }).state;
  expect(next.field(linesField)).toEqual([info]);
});

test("line info is kept when the document changes", () => {
  const state = EditorState.create({ extensions: [linesField] });
  const withInfo = state.update({ effects: setLines.of([info]) }).state;

  const edited = withInfo.update({ changes: { from: 0, insert: "a" } }).state;
  expect(edited.field(linesField)).toEqual([info]);
});
