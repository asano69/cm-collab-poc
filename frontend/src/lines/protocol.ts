// What the server tells about one line.
export interface LineInfo {
  id: string; // a UUID
  updated: number; // Unix time in seconds
  admin: string;
  key: string; // public key of the client that made the edit; empty without one
  signature: string; // by the main server, base64
}

// Server -> client: the info of every line, in line order.
export interface LinesMessage {
  type: "lines";
  lines: LineInfo[];
}
