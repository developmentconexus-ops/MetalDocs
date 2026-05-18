import { describe, expect, it } from "vitest";
import { resolveErrorMessage } from "../errors";

describe("resolveErrorMessage", () => {
  it("returns the mapped Portuguese message for a known code", () => {
    expect(resolveErrorMessage("sod.submitter_cannot_sign", "fallback")).toContain("submeteu este documento");
  });

  it("returns the unresolved comments approval message for approval.unresolved_comments", () => {
    expect(resolveErrorMessage("approval.unresolved_comments", "fallback")).toContain("comentários pendentes");
  });

  it("returns the backend message for an unknown code", () => {
    expect(resolveErrorMessage("unknown.code", "Mensagem do backend")).toBe("Mensagem do backend");
  });

  it("returns the backend message for an undefined code", () => {
    expect(resolveErrorMessage(undefined, "Mensagem do backend")).toBe("Mensagem do backend");
  });
});
