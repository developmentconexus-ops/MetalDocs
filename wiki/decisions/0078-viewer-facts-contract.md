# ADR 0078 — Viewer-Facts Contract (Server-Derived Eligibility Truth for Clients)

> **Status:** Accepted
> **Date:** 2026-07-08
> **Scope:** The approval instance view DTO exposes a server-derived `viewer` block; the frontend is
> forbidden from deriving stage eligibility. Display truth, not an enforcement path.
> **Milestone:** `docs/superpowers/milestones/approval-remediation/milestone-2d-workflow-coherence-fe/f1-viewer-contract/`
> **Governing spec:** `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` §2 (D1), §4 (A1), §8.1
> **Related:** ADR 0022 (capabilities), ADR 0077 (delegation), ADR 0074 (route versioning), ADR 0035 (generated DTOs)
> **Key files:**
> - `internal/modules/documents/approval/domain/viewer.go` — `ViewerFacts` + pure `ViewerEligibility`
> - `internal/modules/documents/approval/domain/eligibility.go` — reused `ResolveEligibleIdentity`
> - `internal/modules/documents/approval/domain/sod.go` — reused `CheckSoD`
> - `internal/modules/documents/approval/application/read_service.go` — `LoadInstanceByDocumentForViewWithViewer`
> - `internal/modules/documents/approval/http/get_instance_handler.go` — off-tx delegator display name
> - `api/openapi/v1/openapi.yaml` — `ApprovalInstanceByDocumentResponse.viewer`

## Context

M2c shipped a live `412 precondition.content_hash_mismatch`: the frontend offered the signature panel
on a review-kind stage because it derived "what can the viewer do" from document status alone
(`signoffOffered`, `useDocumentApprovalArtifact.ts:205`). Four parallel client-side derivations of
eligibility existed (governing spec §1.2-A). No client derivation can ever be correct: delegation is
resolved server-side in-tx at signoff (`decision_service.go`), and delegates are not in the DTO's
`actors[]` snapshot — the client cannot see its own eligibility (§1.2-E).

## Decision

The server computes the viewer's eligibility in the instance view-read path and emits it as a `viewer`
block on `ApprovalInstanceByDocumentResponse`:

- `is_author` — viewer is the instance's `submitted_by`.
- `eligible_for_active_stage` — the viewer may act on the active stage NOW: composed **exclusively** on
  the write-path primitives `ResolveEligibleIdentity` (snapshot pool ∪ active delegation) then `CheckSoD`
  (author-exclusion + cross-stage double-sign). No second membership or SoD rule is introduced — that
  would recreate the split-brain this ADR closes. No active stage ⇒ `false`.
- `via_delegation_from` — `{user_id, display_name}` of the delegator when (and only when) eligibility is
  satisfied purely via delegation; `null` when directly in the pool. Display name resolved **off-tx**
  via the `displayNameReader` port (H-PRE-1).
- `has_signed_active_stage` — the viewer already recorded a signoff on the active stage.

The block is **always present** whenever the instance is returned (terminal instances ⇒ all-false/null),
so the frontend selector `deriveWorkspaceMode` reads one stable shape.

**These are display facts, not an authorization path.** Enforcement stays exactly where ADR 0022 puts
it: tier-2 `authz.Require` in-tx plus the per-instance `CheckEligibility`/`CheckSoD` predicates at the
write call sites. The DTO never gates the server; a client that ignores `viewer` and POSTs anyway is
still refused by the write path. The `viewer` block only decides what the UI offers.

## Consequences

- The frontend deletes all client-side eligibility derivation; `deriveWorkspaceMode(doc, instance, viewer)`
  renders server facts (F2d.3). The M2c 412 class becomes structurally impossible: a review-kind stage
  never carries an eligible signature affordance because the server never reports approval-stage
  eligibility on a review stage.
- Contract-first: the block is defined in OpenAPI and consumed only via generated types (ADR 0035); no
  hand-written `body.data.viewer` reader.
- Eligibility logic has exactly one implementation (`ResolveEligibleIdentity` + `CheckSoD`), now serving
  both the write decision and the read projection.

## Alternatives rejected

- **Keep client-side derivation, add one more guard** — symptom patch; leaves four derivations and the
  delegation blind spot. Explicitly a validator FAIL condition for M2d (milestone §4).
- **Expose the raw delegation rows to the client and let it compute** — leaks authz internals into the
  DTO and duplicates the SoD rule on the client; violates single-source-of-truth (ADR 0022).

## Amendment 2026-07-08 — Visibility gate converges on the same primitive

Building the `viewer` block surfaced a second defect of the same class as the M2c 412: the instance
**visibility** gate (`requireInstanceVisible`, ADR 0075/F8) was written before delegation (ADR 0077/F9)
and never taught it. Live-DB proof: a delegate could **sign** a stage (`decision_service.go`,
delegation-aware) yet received `ErrInstanceNotVisible` → 404 **loading** the instance
(`TestViewerBlock_Delegate`, real Postgres 2026-07-08). Can-act-but-cannot-see — a divergent membership
rule, exactly what the viewer block exists to kill on the other surfaces.

`requireInstanceVisible` now composes on the single eligibility primitive: author fast-path →
`domain.CheckEligibility` (direct snapshot-pool membership, no delegation load) → on miss,
`repo.LoadActiveDelegationsFor` + `domain.ResolveEligibleIdentity` (delegation fallback), with the
`CapApprovalOversee` / `CapDocumentEdit`-in-area capability fallbacks unchanged. The hand-rolled
membership loop over `eligible_actor_ids` is deleted.

- **Policy is unchanged.** ADR 0075's visibility *set* (author ∪ pool ∪ oversee ∪ edit-in-area) and
  ADR 0077's delegation semantics are made mutually consistent; no new rule, no new capability, no
  write-path change. Delegation grants no capability (ADR 0077 §2).
- **Monotonic widening only.** The set gains exactly the active delegates of current-or-past pool
  members — precisely the identities that can already ACT on the instance. Every prior deny case is
  unchanged (proven: the 10 tenant-grade `TestLoadInstance_*` / `TestLoadActiveInstanceByDocument_*`
  deny+grant tests stay green alongside the 7 `TestViewerBlock_*` scenarios). Establishes the closure
  **eligible-to-act ⟹ able-to-see**.
- **H-PRE-1 safe.** The delegation SELECT is a plain non-recording read inside the plain view tx; no
  advisory lock is held on this path.
- **One projection still unpinned → M3.** The worklist/inbox SQL remains a fourth, hand-written
  re-expression of pool membership. It is not delegation-aware and is not covered here. M3 (approval
  kernel extraction) must either derive it from the kernel or pin it with a Go≡SQL parity test
  (precedent: the SoD app-predicate ↔ DB-trigger mirror). Until then it is the next most likely
  drift-defect site.

System-impact analysis of record:
`docs/superpowers/analysis/2026-07-08-delegation-aware-visibility-system-impact.md` (🟡 Yellow).
Added key file: `internal/modules/documents/approval/application/read_service.go` —
`requireInstanceVisible` (visibility gate composing the shared primitive).
