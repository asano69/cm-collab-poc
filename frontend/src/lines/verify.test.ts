// @vitest-environment node
import { expect, test } from "vitest";
import type { LineInfo } from "./protocol";
import { checkLine, parseAuthorityKey } from "./verify";

const { subtle } = crypto;
const encoder = new TextEncoder();

// SHA-256 of "hello", as the Go authority computes it.
const helloHash =
  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824";

function toBase64(bytes: Uint8Array): string {
  return btoa(String.fromCharCode(...bytes));
}

// Signs the line "hello" the way the Go authority does. Returns the key in
// authorized_keys form's parsed shape and the line info.
async function signedHello() {
  const pair = await subtle.generateKey("Ed25519", true, ["sign", "verify"]);
  const raw = new Uint8Array(await subtle.exportKey("raw", pair.publicKey));

  // The OpenSSH wire format: type string, then the key, each with its length.
  const type = encoder.encode("ssh-ed25519");
  const blob = new Uint8Array(8 + type.length + raw.length);
  const view = new DataView(blob.buffer);
  view.setUint32(0, type.length);
  blob.set(type, 4);
  view.setUint32(4 + type.length, raw.length);
  blob.set(raw, 8 + type.length);
  const key = await parseAuthorityKey(`ssh-ed25519 ${toBase64(blob)} host`);

  const message = encoder.encode(`myapp-line-v1\nx\n${helloHash}\n7\nBob\n`);
  const signature = toBase64(
    new Uint8Array(await subtle.sign("Ed25519", pair.privateKey, message)),
  );
  const info: LineInfo = { id: "x", updated: 7, admin: "Bob", key: "", signature };
  return { key, info };
}

test("a line that was signed verifies", async () => {
  const { key, info } = await signedHello();
  expect(await checkLine(key, info, "hello")).toBe(true);
});

test("a changed line text fails", async () => {
  const { key, info } = await signedHello();
  expect(await checkLine(key, info, "hellO")).toBe(false);
});

test("a changed editor name fails", async () => {
  const { key, info } = await signedHello();
  expect(await checkLine(key, { ...info, admin: "Alice" }, "hello")).toBe(false);
});

test("an unsigned line fails", async () => {
  const { key, info } = await signedHello();
  expect(await checkLine(key, { ...info, signature: "" }, "hello")).toBe(false);
});
