import type { LineInfo } from "./protocol";

// Shows the info of a line in a box at the given viewport position. The box
// goes away on the next mouse press.
export function showLinePopup(info: LineInfo, x: number, y: number) {
  const box = document.createElement("div");
  box.className = "cm-line-popup";
  box.style.left = `${x}px`;
  box.style.top = `${y}px`;

  const rows = [
    ["ID", info.id],
    ["Updated", new Date(info.updated * 1000).toLocaleString()],
    ["Editor", info.admin],
  ];
  for (const [label, value] of rows) {
    const row = document.createElement("div");
    row.textContent = `${label}: ${value}`;
    box.append(row);
  }

  document.body.append(box);
  document.addEventListener("mousedown", () => box.remove(), { once: true });
}
