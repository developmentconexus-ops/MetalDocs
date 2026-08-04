import { describe, expect, it } from "vitest";
import { resolveErrorMessage } from "../errors";

describe("resolveErrorMessage", () => {
  it("returns the mapped Portuguese message for a known code", () => {
    expect(resolveErrorMessage("permission.sod_submitter_cannot_sign", "fallback")).toContain("submeteu este documento");
  });

  it("returns the unresolved comments message for state.approval_blocked_unresolved_comments", () => {
    expect(resolveErrorMessage("state.approval_blocked_unresolved_comments", "fallback")).toContain("comentários pendentes");
  });

  it("returns the backend message for an unknown code", () => {
    expect(resolveErrorMessage("unmatched.code", "Mensagem do backend")).toBe("Mensagem do backend");
  });

  it("returns the backend message for an undefined code", () => {
    expect(resolveErrorMessage(undefined, "Mensagem do backend")).toBe("Mensagem do backend");
  });
});
