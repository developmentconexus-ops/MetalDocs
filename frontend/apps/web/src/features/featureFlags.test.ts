import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "../lib/api/errors";

const apiFetchMock = vi.hoisted(() => vi.fn());

vi.mock("../lib/api/client", () => ({
  apiFetch: apiFetchMock,
}));

describe("initFeatureFlags", () => {
  beforeEach(() => {
    apiFetchMock.mockReset();
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("routes through the shared apiFetch transport (not raw fetch)", async () => {
    const rawFetch = vi.fn();
    vi.stubGlobal("fetch", rawFetch);
    apiFetchMock.mockResolvedValueOnce({ MDDM_NATIVE_EXPORT_ROLLOUT_PCT: 25 });

    const { initFeatureFlags, featureFlags } = await import("./featureFlags");
    await initFeatureFlags();

    expect(apiFetchMock).toHaveBeenCalledOnce();
    expect(apiFetchMock).toHaveBeenCalledWith("/api/v1/feature-flags");
    expect(rawFetch).not.toHaveBeenCalled();
    expect(featureFlags.MDDM_NATIVE_EXPORT_ROLLOUT_PCT).toBe(25);
  });

  it("keeps defaults when apiFetch rejects with an ApiError (problem+json path)", async () => {
    apiFetchMock.mockRejectedValueOnce(
      new ApiError({ code: "http_500", status: 500, title: "boom", detail: "boom" }),
    );

    const { initFeatureFlags, featureFlags } = await import("./featureFlags");
    await expect(initFeatureFlags()).resolves.toBeUndefined();

    expect(featureFlags.MDDM_NATIVE_EXPORT_ROLLOUT_PCT).toBe(0);
  });

  it("clamps an out-of-range percentage from the response", async () => {
    apiFetchMock.mockResolvedValueOnce({ MDDM_NATIVE_EXPORT_ROLLOUT_PCT: 250 });

    const { initFeatureFlags, featureFlags } = await import("./featureFlags");
    await initFeatureFlags();

    expect(featureFlags.MDDM_NATIVE_EXPORT_ROLLOUT_PCT).toBe(100);
  });
});
