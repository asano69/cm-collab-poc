// Server -> client: the ID of every line, in line order.
export interface LinesMessage {
  type: "lines";
  ids: number[];
}
