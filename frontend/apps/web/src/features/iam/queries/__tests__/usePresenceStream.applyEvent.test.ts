import { describe, expect, it } from "vitest";
import { applyEvent } from "../usePresenceStream";

const alice = { userId: "u1", displayName: "Alice", lastSeenAt: "2026-06-03T12:00:00Z", status: "online" as const };
const bob = { userId: "u2", displayName: "Bob", lastSeenAt: "2026-06-03T12:00:00Z", status: "online" as const };

describe("usePresenceStream applyEvent", () => {
  it("replaces list on snapshot", () => {
    const out = applyEvent([alice], { type: "snapshot", presence: [bob] });
    expect(out).toEqual([bob]);
  });

  it("noops on heartbeat", () => {
    const prev = [alice];
    const out = applyEvent(prev, { type: "heartbeat" });
    expect(out).toBe(prev);
  });

  it("removes user on leave", () => {
    const out = applyEvent([alice, bob], { type: "leave", userId: "u1" });
    expect(out.map((p) => p.userId)).toEqual(["u2"]);
  });

  it("upserts on join with new userId", () => {
    const out = applyEvent([alice], {
      type: "join",
      userId: "u3",
      displayName: "Carol",
      lastSeenAt: "2026-06-03T12:05:00Z",
      status: "online",
    });
    expect(out.map((p) => p.userId).sort()).toEqual(["u1", "u3"]);
  });

  it("upserts on idle event for existing user, updates status", () => {
    const out = applyEvent([alice, bob], {
      type: "idle",
      userId: "u1",
      lastSeenAt: "2026-06-03T12:10:00Z",
      status: "idle",
    });
    const updated = out.find((p) => p.userId === "u1");
    expect(updated?.status).toBe("idle");
    expect(updated?.displayName).toBe("Alice");
  });
});
