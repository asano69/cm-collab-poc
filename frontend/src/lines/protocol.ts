// Server -> client: the ID (a UUID) of every line, in line order.
export interface LinesMessage {
  type: "lines";
  ids: string[];
}
