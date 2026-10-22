// What the server tells about one line.
export interface LineInfo {
  id: string; // a UUID
  updated: number; // Unix time in seconds
  admin: string;
}

// Server -> client: the info of every line, in line order.
export interface LinesMessage {
  type: "lines";
  lines: LineInfo[];
}
