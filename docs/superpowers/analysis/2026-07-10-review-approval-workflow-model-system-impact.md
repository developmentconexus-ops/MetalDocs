# System-impact analysis — Review/approval workflow model (G1/G2/G3 + 2 screens)

**Date:** 2026-07-10
**Intent (one line):** Implement the ratified 5-rule review/approval model — G1 per-profile signature policy, G2 request_changes on approval stages, G3 fast-forward "Aprovar já", plus route-builder-v2 + approver-execution-panel screens (and the reject-default fix).
**Work type:** feature (3 gated backend features + 2 frontend screens; NOT a new module)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10)*

Source of truth: `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (LOCKED). Memory: `review-approval-workflow-model-ratified`.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** feature. Additive parameterization on the mature approval kernel; no new bounded context.
- **Owning module(s):**
  - `documents/approval` (kernel — ADR 0072 nested exception) — owns `Route.Validate` (`internal/modules/documents/approval/domain/route.go:76`), the verdict guard (`application/review_verdict_service.go:128`), signoff recording, and the fast-forward composition. **Owns G2, G3, and the app-side of G1's route-shape rule.**
  - `documents` — owns document profile snapshot + status; `ErrProfileNotConfigured`/`ErrApprovalRouteMissing` (ADR 0073) already bind profile→route. Route-shape policy check surfaces here at submit/route-admin time.
  - `taxonomy` — owns the `document_profiles` table (ADR 0038 / FamilyCodeResolver; `db/baseline/0001_current_schema.sql:1116`). **The new signature-policy attribute's home is a taxonomy-owned column** → G1 crosses this boundary.
  - frontend `features/approval` — route builder v2 + approver execution panel.
- **Explicitly NOT owning:** `iam`/`auth` — no new capability is required (existing `approval.review` / signoff caps cover both actions; verify in design). `security`, `distribution`, `render` — untouched.
- **Cross-module edges (with direction):**
  - `documents/approval → documents` — route/verdict services already consume document area/profile via `docapp.LoadDocumentAreaCode` (`review_verdict_service.go:109`). Published-interface path, keep it.
  - `documents → taxonomy` — reading the new profile signature-policy attribute MUST go through taxonomy's published interface/read model, never raw `document_profiles` SQL from approval. `Route.Validate` stays **pure domain** — policy is passed IN as a parameter, not fetched inside the domain.
- **Ambiguity?** The signature-policy attribute's owning table (taxonomy `document_profiles` vs a documents-side policy row) is a design-time wiring choice, not a blocking boundary — owning module for the *feature* is unambiguous (`documents/approval`). No **AS-3**.

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** M2b substrate is already landed and sound — `StageKind{review,approval}` domain type + DB CHECK (migration 0286), quorum vocabulary (`any_1_of`/`all_of`/`m_of_n`), route versioning (0287), `signature_meaning IN ('approval','rejection')` on signoffs (0286), `AdvanceStage()` is kind-agnostic (a review-only route already terminates in `InstanceApproved` — R1's Simples profile is a FEATURE of existing code, not new machinery).
- **Sound, or legacy/patch?** **Sound.** This is Grade-A kernel (signed off 2026-06-21). The three gaps are additive parameterization on deliberate substrate the M2b team left as "F1 substrate only — no service wiring". We are wiring the substrate as designed, not patching legacy. **No AS-2.**
- **One foundation caveat (minor):** `strictjson` strict-decode is module-private (`.../http/contracts/strictjson.go`) — if a new G1/G3 contract needs it cross-module, promote to `internal/platform/strictjson`, don't import the private pkg. Carry as a design note, not a blocker.

## 3. Invariant alignment
*(the 6 non-negotiables)*

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes (G2, G3) | G2 keeps `authz.Require(CapApprovalReview)` on the review path; G3 requires BOTH the review cap and the signoff cap in-tx (two checks, one gesture). Reason in caps, never "approver can X". Per-profile policy is a domain rule, NOT an authz rule. | `authz.Require` / `authz.SeedTxIdentity` |
| Contract-first (OpenAPI + oapi-codegen) | Yes (G1, maybe G3) | G1 profile signature-policy surfaces in profile-admin + route-admin API → edit `api/openapi/v1/openapi.yaml` + regen; never hand-edit `api.gen.go`. G3 fast-forward likely a new field/endpoint → same. G2 reuses the existing verdict endpoint (verb unchanged). | `oapi-codegen` per-module `cfg.yaml`+`gen.go` |
| Multi-tenant pooled | Yes | New policy attribute on a tenant table (`document_profiles` has `tenant_id`); all reads carry tenant predicate + tx-local GUC. Cross-tenant → 404. | `tenant.FromContext`, `authz.SeedTxIdentity` |
| Async = transactional outbox | Yes (G3) | G3 signature emits governance/notification events → enqueue in the business tx, consumer does network idempotently. No inline network call. | outbox repo |
| DB enforces invariants | Yes (G1) | **G1:** the friendly app check in `Route.Validate` (Controlado ⇒ ≥1 approval stage) needs a DB last-line — a CHECK/trigger relating a route's stage-kinds to its profile's policy. **G2:** verdict×stage-kind block is **app-domain-only** (0286 adds NO verdict-vs-stage CHECK; signoff `signature_meaning='rejection'` already supports signed-reject) → relaxing `ErrVerdictWrongStageKind` needs NO DB change; confirm no verdict CHECK exists before closing. | migration + named CHECK/trigger |
| Cross-module via published interface only | Yes (G1) | approval/documents read profile policy through taxonomy's published read interface; `Route.Validate` receives policy as a param. Never reach into `document_profiles` from approval. | owner `domain/port.go` / consumer `application/ports.go` |

Any violation → **AS-1**: none. All six are satisfiable within the existing frameworks. No hard-stop.

## 4. Capability wiring
**N/A** — no new IAM capability. G2/G3 reuse existing review + signoff capabilities; G1 is a domain policy, not an authz capability. (Design must confirm the two existing caps compose for G3's dual-write; if a genuinely new cap emerges, re-open the 10-touchpoint walk + bump `TestCapabilityRegistrySize`.)

## 5. Module wiring
**N/A** — no module is born. `documents/approval` is the existing nested kernel (ADR 0072); all work lands in existing folders (`domain/`, `application/`, `http/contracts/`, `api/`).

## 6. Frameworks to reuse, not reinvent

- `TxRunner` (`Do`) — G2 verdict tx, G3 dual-write in one business tx. Never `*sql.DB`.
- `authz.Require` / `authz.SeedTxIdentity` — tier-2 checks for G3 (review + signoff caps).
- `audit.RecordTx` — G3's TWO ledger entries (review verdict + signature) are separate audit records in the same tx.
- Outbox repo — G3 governance/notification side effects.
- `problem.New`/`Write` — every error (new `Err... ` domain errors mapped to problem codes).
- `oapi-codegen` — G1/G3 contract surfaces.
- `testdb.Open` + factory builders — all integration tests.
- `strictjson` — new request contracts (promote to platform if used cross-module — §2 caveat).

No hand-rolled equivalents. No new cross-cutting framework needed.

## 7. Contract & data

- **OpenAPI-first:** G1 — profile signature-policy field on profile-admin (and route-admin validation error surface); G3 — fast-forward eligibility flag/endpoint on the signoff/verdict contract. Edit spec → regen. G2 — no route change (reuses verdict endpoint), only a relaxed domain rule + possibly a new 200-path for `request_changes` from approval stage.
- **Migration:**
  - G1: `document_profiles.signature_policy` (or governance_class) column + backfill (existing profiles → policy that preserves today's behavior, i.e. Controlado/required as safe default OR explicit per operator), tenant-scoped; DB CHECK/trigger enforcing route-shape ↔ profile-policy as the last line.
  - G2: **no migration expected** — verify no verdict×stage-kind CHECK before closing.
  - G3: likely no schema change (dual-write uses existing verdict + signoff tables); confirm ledger tables accept the composed write.
- **Destructive change?** G1 column add is expand-only (additive). Backfill must preserve current behavior (no route silently becomes invalid) — expand/contract, never break live routes in one step.

## 8. Test & QA plan

- **Canonical framework:** `testdb` integration factory; `//go:build integration`; R1–R4 discipline. Strict TDD — failing test first per gap. Delete legacy one-off approval tests that break; repair only contract/invariant guards.
- **QA gates that apply (feature subset):** contract (G1/G3 openapi parity), authz (G2/G3 cap checks), multi-tenant isolation (G1 profile policy per-tenant), async/idempotency (G3 dual-write idempotent), DB-invariant (G1 route-shape CHECK), docs. Distribution/render gates **N/A**.
- **Evidence shape:** `go build ./...`, `go test ./...` + `go test -tags=integration ./...` (vs docker postgres), `.\scripts\check-system-runnable.ps1`, `npm run typecheck:docx-v2` where relevant, FE vitest, **live gateway smoke on :80 with UI-driven QA** for both screens (curl-only = FAIL). Per-gap evidence.md; commit per gap, no batch merge; NEVER push without operator OK.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/documents.md` (approval kernel section) + refresh `Last verified`; note the profile signature-policy in the taxonomy/profiles doc. No new module doc.
- **REQ IDs cited:** from `wiki/architecture/backend-target-architecture.md` (approval/route + profile REQ IDs — pull exact IDs in design). ADR 0022 (authz), 0072 (nested approval), 0073 (profile→route binding), 0077 (delegation) are the governing decisions in scope.
- **ADR required? YES.** Two policy changes ratify into the record:
  - **G1** introduces a per-profile signature-policy dimension (Controlado ⇒ ≥1 approval stage). This is a new standing policy on profiles.
  - **G2** relaxes the domain guard `ErrVerdictWrongStageKind` to permit `request_changes` (never `ready`) on approval-kind stages — the R3 backend half.
  One ADR can ratify the 5-rule model → cite REQ IDs, reference the locked spec, and **explicitly record that the per-profile policy SUPERSEDES / rejects any system-global "must have signature" rule** (operator explicitly killed the global version — never reintroduce).

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — proceed to brainstorming/design. Fits the architecture cleanly on a sound foundation; the yellow is the required ADR + the DB-lockstep and cross-module wiring risks, all carried as locked constraints (not blockers).
- **Open hard-stops:** none. AS-1 no (all invariants satisfiable) · AS-2 no (Grade-A foundation, substrate deliberately staged) · AS-3 no (owning module unambiguous).
- **Locked constraints handed to brainstorming/design:**
  1. **ADR required** ratifying the 5-rule model + G1 policy + G2 relaxation; cite REQ IDs; explicitly record per-profile-NOT-global (supersede the global-invariant proposal the operator killed).
  2. **Per-profile signature policy, NEVER system-global** — hard operator lock.
  3. **G1 `Route.Validate` stays pure domain** — policy passed IN as a parameter; approval/documents read the attribute via taxonomy's published interface, never raw `document_profiles` SQL.
  4. **DB is the last line for G1** — route-shape ↔ profile-policy CHECK/trigger; app check is friendly first line. G1 column add is expand-only with behavior-preserving backfill.
  5. **G2 relax the guard in BOTH the service pre-check (`review_verdict_service.go:128`) AND `domain.NewVerdict`** — allow ONLY `request_changes`; `ready` stays hard-blocked. Verify no verdict×stage-kind DB CHECK exists (0286 shows none) before closing. Keep signed-reject capability in the kernel (`signature_meaning='rejection'`) even though it leaves the UI.
  6. **G3 preserves H-PRE-1** — any authz-recording read stays OFF the lock-holding/audit tx (mirror the existing `LoadActorDisplayName` off-tx hoist). TWO separate ledger entries (`audit.RecordTx` ×2), one UX ceremony. Freeze boundary UNCHANGED (end of last review-kind stage; approver-only route freezes at submit). NO pre-signing.
  7. **Contract-first** — G1/G3 API surfaces via `api/openapi` + `oapi-codegen`; never hand-edit `api.gen.go`.
  8. **Gated delivery** — G1 → G2 → G3 → screens, one at a time, each: analysis → plan → TDD → evidence.md → commit. No batch merge. Carry the `defaultOptionKey="reject"` fix WITH the approver-execution-panel screen (no option preselected).
  9. **Frontend** lands AFTER the backend contract for each gap; UI-driven live QA on :80 mandatory.
