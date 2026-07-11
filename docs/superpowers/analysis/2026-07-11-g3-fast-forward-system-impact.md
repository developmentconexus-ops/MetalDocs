# System-impact analysis — G3 fast-forward "Aprovar já"

**Date:** 2026-07-11
**Intent (one line):** One gesture records review verdict + approval signature as two audit records in one tx; eligibility surfaced in verdict response (spec R5, ROADMAP unit 2.3).
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟢 Green

---

## 1. Classify & own

- **Work type:** feature (extends existing services inside the `documents/approval` nested kernel).
- **Owning module(s):** `documents` (nested `documents/approval` per ADR 0072) — verdict + signoff services, quorum, freeze all live there.
- **Explicitly NOT owning:** `iam` (only consumed via `authz.Require`), `audit` (governance_events written via approval's own EventEmitter, existing pattern), `controlleddocuments` (consumed via existing `CDFieldReader` port).
- **Cross-module edges (with direction):** `documents/approval → iam/authz` (Require, SeedTxIdentity — published), `documents/approval → platform/db` (TxRunner). No new edges; fast-forward composes two flows already inside the same module.
- **Ambiguity?** None. AS-3 not triggered.

## 2. Foundation verdict

- **Base:** `ReviewVerdictService.RecordVerdict` (review_verdict_service.go) + `DecisionService.RecordSignoff` (decision_service.go), shared domain quorum (`EvaluateQuorum`), freeze (`executeFreeze`), idempotent repo inserts, governance_events emitter.
- **Sound or patch?** Sound. G1 (per-profile policy) and G2 (request_changes on approval stage) landed on this base; roadmap ordering rationale is explicit and locked: "G1–G3 land INSIDE the nested kernel BEFORE extraction — extraction then moves complete, ratified-model-conformant code once." Kernel extraction is the later unit 3.x.
- **GM fork record:** A = compose fast-forward from tx-scoped cores extracted (behavior-preserving) out of the two existing service methods, inside the current kernel. B = full approval-kernel extraction first. B is a named ROADMAP unit ordered AFTER G3 by operator lock — choosing A here is the ratified route, not a silent local max. No-deepening: extraction of tx-cores is refactoring the base into the exact seams the future kernel extraction needs, not growing it. AS-2 not triggered.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Yes | Fast-forward tx requires BOTH `approval.review` (verdict leg) and the signoff capability (signature leg) via tier-2 `authz.Require` in-tx, same as the two existing flows; no new capability | `authz.Require`, `authz.SeedTxIdentity`, `authz.WithCapCache` |
| Contract-first | Yes | `openapi.yaml` edit: `fast_forward_eligible` (+ next-stage hint) on ReviewVerdictResponse; new fast-forward route + DTOs; then regen. Never hand-edit api.gen.go | `oapi-codegen` via module `gen.go` |
| Multi-tenant pooled | Yes | Same-tx GUCs via existing service pattern; all repo calls tenant-predicated already; cross-tenant → existing 404 mapping | `tenant.FromContext`, existing repo |
| Async = transactional outbox | Yes (unchanged) | Final-approval PDF dispatch + pin already outbox-shaped inside signoff core; fast-forward reuses that core verbatim | existing `pdfDispatch.EnqueuePDFTx`, `pinInvoker` |
| DB enforces invariants | Yes (unchanged) | SoD trigger, idempotent insert constraints, OCC revision check all pre-exist and fire for both legs; no new table | existing constraints |
| Cross-module via published interface | Yes | No new cross-module reach; all inside documents/approval | — |

AS-1 not triggered. H-PRE-1: follow existing shape — display-name/authz-recording reads stay off-tx as in both current services; `authz.Require` runs in the writable business tx (the norm).

## 4. Capability wiring

**N/A** — no new/changed capability. Reuses `approval.review` (verdict leg) and the existing signoff capability (signature leg); both routes' tier-1 arms follow existing wiring for the new endpoint (route→cap map entry for the new path uses existing caps; tripwire arms are generated from the Go registry — no registry size change).

## 5. Module wiring

**N/A** — no new module; feature inside `documents/approval`.

## 6. Frameworks to reuse, not reinvent

`TxRunner.Do` (single tx for both ledger entries), `tenant.FromContext`, `authz.SeedTxIdentity`/`Require`/`WithCapCache`, `problem.Write` via module error mapper, module `contracts` strict decode, existing `PostgresSignoffIdempStore` (new route-template constant for fast-forward, same store), governance EventEmitter, `testdb` factory for integration tests. Nothing hand-rolled.

## 7. Contract & data

- **OpenAPI-first:** (1) `ReviewVerdictResponse` += `fast_forward_eligible: boolean` (required, default false semantics) + `next_stage_id` (optional); (2) new `POST /approval/instances/{instance_id}/stages/{stage_id}/fast-forward` with Idempotency-Key + If-Match, body = verdict fields + signature ceremony fields (`password_token`, `content_hash`, comment), responses incl. RFC 9457 for not-eligible. Regen `internal/modules/documents/approval/api`.
- **Migration:** none — no new tables; two existing ledger tables (`approval_review_verdicts`, `approval_signoffs`) + `governance_events` get their rows from existing insert paths.
- **Destructive change?** No — additive response field + new route.

## 8. Test & QA plan

- **Canonical framework:** `tests/integration/approval/` testdb factory (`//go:build integration`), pattern of `review_verdict_integration_test.go`; unit tests beside services.
- **QA gates applying (feature subset):** contract parity, authz (both caps enforced; negative test), multi-tenant isolation (cross-tenant 404), idempotency (replay of fast-forward), DB-invariant (SoD, two rows one tx — rollback test proves atomicity), freeze boundary regression (`ready` still forbidden on approval stage, freeze fires before signature).
- **Evidence:** `go build ./...`, api-lint strict, module boundaries, `go test ./...`, `.\scripts\test-integration.ps1` (bar: no NEW failures vs 9 accepted RED) → evidence.md with dispatch ledger.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/documents.md` approval section (fast-forward flow) + refresh `Last verified`.
- **REQ IDs cited:** REQ-AUTHZ (tier-2 in-tx), REQ-CONTRACT (spec-first), reviewer to cite exact IDs from backend-target-architecture.md in slice reviews.
- **ADR required?** No — in-bounds feature implementing ratified spec R5; no MUST-deviation, no policy change. Freeze boundary and G2 invariant unchanged.

## 10. Verdict & locked constraints

- **Verdict:** 🟢 Green.
- **Open hard-stops:** none.
- **Locked constraints handed to design:**
  1. Two ledger entries NEVER collapse — separate `approval_review_verdicts` + `approval_signoffs` rows + their governance events, one tx.
  2. Server re-checks R5 (a)∧(b) authoritatively inside the tx; not-eligible ⇒ fail whole tx with problem+json (fail closed, no partial verdict write on the fast-forward route).
  3. Freeze ordering inside tx: verdict → quorum completes stage → AdvanceStage → executeFreeze (existing call site) → signature leg. Signature never precedes freeze.
  4. `ready` on approval-kind stage stays forbidden (G2 invariant untouched).
  5. Contract-first only; behavior-preserving extraction of tx-cores (existing tests must stay green before new behavior lands).
  6. Idempotency for the new route via existing signoff idemp store with a distinct route-template constant.
