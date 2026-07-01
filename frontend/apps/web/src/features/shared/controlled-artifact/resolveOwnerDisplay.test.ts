import { describe, it, expect } from "vitest";
import { resolveOwnerDisplay } from "./resolveOwnerDisplay";

describe("resolveOwnerDisplay", () => {
  it("returns null when createdBy is null", () => {
    expect(resolveOwnerDisplay(null, { userId: "u1", displayName: "Alice" })).toBeNull();
  });

  it("returns null when createdBy is undefined", () => {
    expect(resolveOwnerDisplay(undefined, { userId: "u1", displayName: "Alice" })).toBeNull();
  });

  it("returns displayName when creator matches current user", () => {
    expect(
      resolveOwnerDisplay("u1", { userId: "u1", displayName: "Alice" }),
    ).toBe("Alice");
  });

  it("returns createdBy when creator matches current user but displayName is null", () => {
    expect(
      resolveOwnerDisplay("u1", { userId: "u1", displayName: null }),
    ).toBe("u1");
  });

  it("returns createdBy when creator does not match current user", () => {
    expect(
      resolveOwnerDisplay("u2", { userId: "u1", displayName: "Alice" }),
    ).toBe("u2");
  });

  it("returns createdBy when user is null", () => {
    expect(resolveOwnerDisplay("u2", null)).toBe("u2");
  });
});
