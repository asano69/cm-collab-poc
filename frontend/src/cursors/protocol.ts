// Client <-> server: the position of one client's cursor.
export interface CursorMessage {
  type: "cursor";
  clientID: string;
  name: string;
  color: string;
  pos: number;
}

// Server -> client: the client disconnected.
export interface LeaveMessage {
  type: "leave";
  clientID: string;
}
