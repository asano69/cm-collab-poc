import type { LineInfo } from "./protocol";

// Where the server publishes the public key that signs the lines.
const KEY_URL = "/authority.pub";

// An ssh-ed25519 key ends with the 32 raw bytes of the public key.
const RAW_KEY_SIZE = 32;

const encoder = new TextEncoder();

function fromBase64(text: string) {
  return Uint8Array.from(atob(text), (c) => c.charCodeAt(0));
}

async function sha256Hex(text: string): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", encoder.encode(text));
  return Array.from(new Uint8Array(digest), (b) =>
    b.toString(16).padStart(2, "0"),
  ).join("");
}

// Imports a key in authorized_keys form ("ssh-ed25519 AAAA... comment").
export async function parseAuthorityKey(line: string): Promise<CryptoKey> {
  const blob = fromBase64(line.trim().split(/\s+/)[1] ?? "");
  if (blob.length < RAW_KEY_SIZE) {
    throw new Error("the authority key is malformed");
  }
  return crypto.subtle.importKey(
    "raw",
    blob.slice(-RAW_KEY_SIZE),
    "Ed25519",
    false,
    ["verify"],
  );
}

// Reports whether the server signed `text` as the line described by `info`.
// The hash is computed here, so a line that differs from the signed one fails.
// The message must match `message` in internal/authority.
export async function checkLine(
  key: CryptoKey,
  info: LineInfo,
  text: string,
): Promise<boolean> {
  const hash = await sha256Hex(text);
  const message = encoder.encode(
    `myapp-line-v1\n${info.id}\n${hash}\n${info.updated}\n${info.admin}\n${info.key}`,
  );
  return crypto.subtle.verify(
    "Ed25519",
    key,
    fromBase64(info.signature),
    message,
  );
}

let authorityKey: Promise<CryptoKey> | undefined;

async function fetchAuthorityKey(): Promise<CryptoKey> {
  const res = await fetch(KEY_URL);
  if (!res.ok) throw new Error(`cannot load the authority key (${res.status})`);
  return parseAuthorityKey(await res.text());
}

// The key is fetched once; a failed fetch is retried on the next call.
function authority(): Promise<CryptoKey> {
  authorityKey ??= fetchAuthorityKey().catch((err) => {
    authorityKey = undefined;
    throw err;
  });
  return authorityKey;
}

// Checks a line against the authority's public key. Rejects when the check
// cannot be made, with a reason for the user.
export async function verifyLine(
  info: LineInfo,
  text: string,
): Promise<boolean> {
  if (!crypto.subtle) {
    throw new Error("verification needs a secure context (https or localhost)");
  }
  return checkLine(await authority(), info, text);
}
