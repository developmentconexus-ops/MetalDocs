import { describe, expect, it } from "vitest";
import { applyEvent } from "../usePresenceStream";

const alice = { user_id: "u1", username: "alice", display_name: "Alice", last_seen_at: "2026-06-03T12:00:00Z", status: "online" as const };
const bob = { user_id: "u2", username: "bob", display_name: "Bob", last_seen_at: "2026-06-03T12:00:00Z", status: "online" as const };

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
    const out = applyEvent([alice, bob], { type: "leave", user_id: "u1" });
    expect(out.map((p) => p.user_id)).toEqual(["u2"]);
  });

  it("upserts on join with new user_id", () => {
    const out = applyEvent([alice], {
      type: "join",
      user_id: "u3",
      username: "carol",
      display_name: "Carol",
      last_seen_at: "2026-06-03T12:05:00Z",
      status: "online",
    });
    expect(out.map((p) => p.user_id).sort()).toEqual(["u1", "u3"]);
    const joined = out.find((p) => p.user_id === "u3");
    expect(joined?.username).toBe("carol");
  });

  it("upserts on idle event for existing user, updates status", () => {
    const out = applyEvent([alice, bob], {
      type: "idle",
      user_id: "u1",
      last_seen_at: "2026-06-03T12:10:00Z",
      status: "idle",
    });
    const updated = out.find((p) => p.user_id === "u1");
    expect(updated?.status).toBe("idle");
    expect(updated?.display_name).toBe("Alice");
    expect(updated?.username).toBe("alice");
  });
});
