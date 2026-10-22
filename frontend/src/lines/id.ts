// Shortens a UUIDv7 for display to 7 digits of its random part. The 32 hex
// digits are 12 of time, the version, 3 random, the variant, and 15 random;
// the version and the variant carry (almost) no randomness and are skipped.
export function shortID(id: string): string {
  const hex = id.replaceAll("-", "");
  return hex.slice(13, 16) + hex.slice(17, 21);
}
