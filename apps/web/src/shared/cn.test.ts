import { describe, expect, it } from "vitest";
import { cn } from "./cn";

describe("cn", () => {
  it("merges class names and ignores falsy values", () => {
    expect(cn("px-2", undefined, null, "", "text-sm")).toBe("px-2 text-sm");
  });
});
