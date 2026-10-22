import { expect, test } from "vitest";
import { shortID } from "./id";

test("shortID takes the first 7 digits of the last block", () => {
  expect(shortID("01a162fa-894e-7053-93ca-ee106b9da0ed")).toBe("ee106b9");
});
