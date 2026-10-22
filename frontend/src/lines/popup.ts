import type { LineInfo } from "./protocol";

function makeRow(label: string, value: string): HTMLElement {
  const row = document.createElement("div");
  row.textContent = `${label}: ${value}`;
  return row;
}

// Shows the info of a line in a box at the given viewport position. The editor
// gets a lock when `verified` resolves to true and a cross when it resolves to
// false; if it rejects, the reason is shown instead. The box goes away on the
// next mouse press.
export function showLinePopup(
  info: LineInfo,
  verified: Promise<boolean>,
  x: number,
  y: number,
) {
  const box = document.createElement("div");
  box.className = "cm-line-popup";
  box.style.left = `${x}px`;
  box.style.top = `${y}px`;

  const editor = makeRow("Editor", info.admin);
  box.append(
    makeRow("ID", info.id),
    makeRow("Updated", new Date(info.updated * 1000).toLocaleString()),
    editor,
  );
  verified.then(
    (ok) => {
      editor.textContent += ok ? " 🔒" : " ❌";
    },
    (err: Error) => box.append(makeRow("Verification", err.message)),
  );

  document.body.append(box);
  document.addEventListener("mousedown", () => box.remove(), { once: true });
}
