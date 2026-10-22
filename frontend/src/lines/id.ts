// Shortens a UUID for display to the first 7 hex digits of its last block,
// e.g. "01a162fa-894e-7053-93ca-ee106b9da0ed" becomes "ee106b9".
export function shortID(id: string): string {
  return id.split("-")[4].slice(0, 7);
}
