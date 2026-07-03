// Feature flag registry. Flags are read once at module load time from a
// window-level config object injected by the backend via the HTML shell.
// When the window injection is absent, initFeatureFlags() fetches from
// GET /api/v1/feature-flags and patches the exported featureFlags object.

import { apiFetch } from "../lib/api/client";
import { isInRolloutBucket } from "./feature-flags/rollout";

type FeatureFlags = {
  /** Percentage (0..100) of users for whom the new client-side MDDM DOCX path is active. */
  MDDM_NATIVE_EXPORT_ROLLOUT_PCT: number;
  /** Always false at module level — use isMddmNativeExportEnabled(userId) for per-user check. */
  MDDM_NATIVE_EXPORT: boolean;
};

function readWindowFlags():
  | Partial<{ MDDM_NATIVE_EXPORT_ROLLOUT_PCT: number }>
  | undefined {
  if (typeof window === "undefined") return undefined;
  return (
    window as unknown as {
      __METALDOCS_FEATURE_FLAGS?: Partial<{
        MDDM_NATIVE_EXPORT_ROLLOUT_PCT: number;
      }>;
    }
  ).__METALDOCS_FEATURE_FLAGS;
}

function clampPct(raw: unknown): number {
  const n = Number(raw);
  return Number.isFinite(n) ? Math.max(0, Math.min(100, n)) : 0;
}

function strictBool(raw: unknown): boolean {
  return raw === true;
}

function readFlags(): FeatureFlags {
  const injected = readWindowFlags();
  return {
    MDDM_NATIVE_EXPORT_ROLLOUT_PCT: clampPct(injected?.MDDM_NATIVE_EXPORT_ROLLOUT_PCT),
    MDDM_NATIVE_EXPORT: false,
  };
}

export const featureFlags: FeatureFlags = readFlags();

/**
 * Fetches feature flags from GET /api/v1/feature-flags and patches the
 * exported featureFlags object. Call once at app init (e.g. in main.tsx)
 * before render. Safe to call even when window injection is present — the
 * endpoint value overrides it.
 *
 * Routed through the shared apiFetch transport (not raw fetch) so this call
 * gets credentials + problem+json handling like every other API call. Errors
 * (network failure, non-2xx, malformed body) are swallowed on purpose — a
 * flags fetch failure must never block app boot; defaults stand.
 */
export async function initFeatureFlags(): Promise<void> {
  try {
    const data = await apiFetch<Partial<{ MDDM_NATIVE_EXPORT_ROLLOUT_PCT: number }>>(
      "/api/v1/feature-flags",
    );
    featureFlags.MDDM_NATIVE_EXPORT_ROLLOUT_PCT = clampPct(data.MDDM_NATIVE_EXPORT_ROLLOUT_PCT);
  } catch {
    // Network error, non-2xx (ApiError), or non-JSON body — keep defaults
  }
}

/** Returns true when the given userId is inside the canary rollout bucket. */
export function isMddmNativeExportEnabled(userId: string): boolean {
  return isInRolloutBucket(userId, featureFlags.MDDM_NATIVE_EXPORT_ROLLOUT_PCT);
}

