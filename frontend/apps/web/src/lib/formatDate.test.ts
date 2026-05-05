import { describe, expect, it } from "vitest";
import { formatISODate, formatDateTime } from "./formatDate";

describe("formatISODate", () => {
  it("extracts YYYY-MM-DD from ISO string", () => {
    expect(formatISODate("2026-05-05T12:34:56Z")).toBe("2026-05-05");
  });

  it("returns empty string for null", () => {
    expect(formatISODate(null)).toBe("");
  });

  it("returns empty string for undefined", () => {
    expect(formatISODate(undefined)).toBe("");
  });

  it("returns empty string for empty string", () => {
    expect(formatISODate("")).toBe("");
  });

  it("returns empty string for invalid date", () => {
    expect(formatISODate("not-a-date")).toBe("");
  });

  it("accepts Date object", () => {
    expect(formatISODate(new Date("2026-01-15T00:00:00Z"))).toBe("2026-01-15");
  });
});

describe("formatDateTime", () => {
  it("returns YYYY-MM-DD HH:MM from ISO string", () => {
    expect(formatDateTime("2026-05-05T12:34:56Z")).toBe("2026-05-05 12:34");
  });

  it("returns empty string for null", () => {
    expect(formatDateTime(null)).toBe("");
  });

  it("returns empty string for invalid date", () => {
    expect(formatDateTime("bad")).toBe("");
  });
});
