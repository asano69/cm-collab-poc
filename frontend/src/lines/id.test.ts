import { expect, test } from "vitest";
import { shortID } from "./id";

test("shortID takes the first random digits and skips version and variant", () => {
  expect(shortID("0192f3a4-b5c6-7abc-8def-0123456789ab")).toBe("abcdef0");
});
