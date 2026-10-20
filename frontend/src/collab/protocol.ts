import { ChangeSet } from "@codemirror/state";
import type { Update } from "@codemirror/collab";

// Wire format of one update. `changes` is ChangeSet JSON.
export interface WireUpdate {
  clientID: string;
  changes: unknown;
}

// Client -> server: unconfirmed updates based on `version`.
export interface PushMessage {
  version: number;
  updates: WireUpdate[];
}

// Server -> client: accepted updates starting at index `from`.
export interface UpdatesMessage {
  from: number;
  updates: WireUpdate[];
}

export function toWire(updates: readonly Update[]): WireUpdate[] {
  return updates.map((u) => ({
    clientID: u.clientID,
    changes: u.changes.toJSON(),
  }));
}

export function fromWire(updates: WireUpdate[]): Update[] {
  return updates.map((u) => ({
    clientID: u.clientID,
    changes: ChangeSet.fromJSON(u.changes),
  }));
}
