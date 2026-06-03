import { describe, expect, it } from "vitest";
import { resolveDateWindow } from "../AuditTab";

const FIXED_NOW = Date.UTC(2026, 5, 3, 12, 0, 0); // 2026-06-03T12:00:00Z

describe("resolveDateWindow", () => {
  it("returns empty window when preset is undefined", () => {
    expect(resolveDateWindow(undefined, undefined, undefined, FIXED_NOW)).toEqual(
      {},
    );
  });

  it("computes 24h rolling window from `now`", () => {
    const result = resolveDateWindow("24h", undefined, undefined, FIXED_NOW);
    expect(result.occurredAfter).toBe(
      new Date(FIXED_NOW - 24 * 60 * 60 * 1000).toISOString(),
    );
    expect(result.occurredBefore).toBeUndefined();
  });

  it("computes 7d rolling window from `now`", () => {
    const result = resolveDateWindow("7d", undefined, undefined, FIXED_NOW);
    expect(result.occurredAfter).toBe(
      new Date(FIXED_NOW - 7 * 24 * 60 * 60 * 1000).toISOString(),
    );
  });

  it("passes custom occurredAfter/occurredBefore through unchanged", () => {
    const after = "2026-05-01T00:00:00.000Z";
    const before = "2026-05-31T23:59:59.000Z";
    expect(
      resolveDateWindow("custom", after, before, FIXED_NOW),
    ).toEqual({ occurredAfter: after, occurredBefore: before });
  });
});
