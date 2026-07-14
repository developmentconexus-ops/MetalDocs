# Feature Flag Lifecycle Standard

> **Last verified:** 2026-06-13 (Wave Z, Z-30)
> **Scope:** Naming convention, required metadata at introduction, ramp/cleanup rules, and current flag inventory for all MetalDocs server-controlled feature flags.

---

## 1. Naming Convention

All feature flag environment variables and JSON keys follow:

```
ff_<area>_<purpose>
```

- `<area>` — lowercase domain area: `mddm`, `iam`, `docs`, `render`, etc.
- `<purpose>` — lowercase snake-case description of the capability being gated.

The JSON key exposed by `GET /api/v1/feature-flags` uses `UPPER_SNAKE_CASE` matching the env-var stem (without the `METALDOCS_` prefix). Example: env var `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT` → JSON key `MDDM_NATIVE_EXPORT_ROLLOUT_PCT`.

---

## 2. Required Metadata at Introduction

Every flag introduced in a PR **must** declare the following as a code comment in `internal/platform/config/feature_flags.go` and in the PR description:

| Field | Description |
|-------|-------------|
| **Owner** | Team or individual responsible for the flag and its cleanup |
| **Ramp plan** | How the flag will be graduated (e.g., 0 → 10 → 50 → 100 → remove) |
| **Cleanup date** | The sprint or calendar date by which the flag must be removed |

Example comment block:
```go
// MDDMNativeExportRolloutPercent gates the client-side MDDM DOCX export path.
// Owner: documents-team
// Ramp plan: 0 (default) → 25 → 100 → remove after stable for one release cycle
// Cleanup date: 2026-Q3 (or whenever MDDM export is declared stable)
// Env: METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT (default 0)
MDDMNativeExportRolloutPercent int
```

---

## 3. Ramp Plan

- Default value must be **0** (flag off) at introduction.
- Ramp increments are deployed via environment variable change; no code change required.
- When the flag reaches 100 and has been stable for one full release cycle, the cleanup gate is triggered.

---

## 4. Dead Flag Removal Policy

> **On next touch:** if you edit `feature_flags.go`, `handler.go`, or any caller and a flag's cleanup date has passed, remove it in the same PR.

Steps to remove a dead flag:
1. Delete the struct field from `config.FeatureFlagsConfig`.
2. Delete the corresponding env-var parse block in `LoadFeatureFlagsConfig`.
3. Remove the JSON key from `featureFlagsResponse` in `platform/featureflags/handler.go`.
4. Remove all frontend consumers of the JSON key.
5. Remove the env-var from `.env.example` and `docker-compose.yml`.

---

## 5. Current Flag Inventory

As of 2026-06-13. Source: `internal/platform/config/feature_flags.go` and `internal/platform/featureflags/handler.go`.

| Flag | Env var | JSON key | Type | Default | Owner | Cleanup date |
|------|---------|----------|------|---------|-------|--------------|
| MDDM native export rollout | `METALDOCS_MDDM_NATIVE_EXPORT_ROLLOUT_PCT` | `MDDM_NATIVE_EXPORT_ROLLOUT_PCT` | `int` (0–100 %) | `0` | documents-team | TBD — requires owner declaration |

**Note:** The current flag (`MDDM_NATIVE_EXPORT_ROLLOUT_PCT`) predates this standard. The owner and cleanup date must be declared by the documents team before the next ramp increment.

---

## 6. Adding a New Flag — Checklist

- [ ] Name follows `ff_<area>_<purpose>` (env var: `METALDOCS_FF_<AREA>_<PURPOSE>`)
- [ ] Struct field added to `config.FeatureFlagsConfig` with owner/ramp/cleanup comment
- [ ] Parse block added to `LoadFeatureFlagsConfig`
- [ ] JSON key added to `featureFlagsResponse` in `platform/featureflags/handler.go`
- [ ] Default is `0` / `false`
- [ ] `.env.example` updated
- [ ] Cleanup date added to this inventory table
