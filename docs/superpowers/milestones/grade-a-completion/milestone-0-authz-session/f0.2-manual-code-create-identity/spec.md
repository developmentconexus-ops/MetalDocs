# Feature F0.2 — Spec

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Folder:** `f0.2-manual-code-create-identity`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / leandrotca.work (Path A — symmetric refactor; closes the documented "do NOT fix silently" bypass) — *implementation may begin.*

> This is the feature's **contract**, written and approved **before any code**.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Mission §7 F0.2 says "seeds the tx identity" — but the manual-code branch has no tx today (explicit code comment `service.go:174-178`: "no SeedTxIdentity, no authz.Require, and no transaction. Flagged here for operator review — do NOT 'fix' silently"). What is the fix shape? | **Path A — symmetric refactor (operator-approved).** Wrap the manual-code branch in `s.runner.Do`, switch `s.docs.Create` → `s.docs.CreateTx`, call `authz.SeedTxIdentity` + `authz.Require(controlled_documents.create, cmd.ProcessAreaCode)` inside that tx — exactly mirroring the auto branch. The code comment was waiting for an operator-approved fix; the mission is that approval. |
| 2 | Why not "identity seed only, keep the bypass"? | A simple identity seed without `authz.Require` does not grant non-admins anything — tier-2 caps are granted by `Require` returning nil. Without it, the audit acceptance "**non-system-admin** manual-code create succeeds" cannot be met. Identity-only is therefore not viable. |
| 3 | H-PRE-1 advisory-lock hazard — anything authz-recording moved inside the tx? | No. All OFF-tx preflight reads (`CodeExists`, `ensureTemplateArtifact`, `GetTemplateVersionState`, `Resolve`) stay OFF-tx. Inside the new tx: only `SeedTxIdentity`, `authz.Require`, `s.docs.CreateTx`. This matches the auto branch's discipline (per the existing comment at `service.go:277-287`). |
| 4 | Race on `CodeExists` (OFF-tx) vs `CreateTx` (IN-tx)? | Pre-existing race, not introduced by this change. The `controlled_documents.code` unique constraint catches it as `ErrCDCodeTaken`. Out of F0.2 scope. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** every caller of `ControlledDocumentService.Create` with a non-nil `cmd.ManualCode` —
  primarily the HTTP `AtomicCreateControlledDocument` route
  (`controlleddocuments/delivery/http/routes.go:69`) and any future API/CLI consumer. The HTTP layer
  has already authenticated the actor (tier-1) and passes `cmd.ActorUserID`; the service is expected
  to honor that identity through tier-2 (area-scoped) authz.
- **Contract:** when a non-system-admin actor who **holds**
  `controlled_documents.create` in `cmd.ProcessAreaCode` (via curated `role_capabilities`) calls
  `Create` with a `ManualCode`, the call **succeeds** (returns a `*CreateResult` with no error). A
  system-admin actor always succeeds (existing bypass, unchanged). An actor without the cap (and
  not a system-admin) is **denied** with the standard `authz.ErrCapDenied`. The behavior is symmetric
  to the auto-code branch.
- **Source of truth for the contract:** the auto-code branch at `service.go:292-405` (the existing
  symmetric pattern), ADR 0022 Phase 7 (area-scoped tier-2 for CD create), and the
  `s.docs.CreateTx` repo signature (`infrastructure/repository.go:353-366`).

## What this feature implements

In `internal/modules/controlleddocuments/application/service.go`, the manual-code branch
(`if cmd.ManualCode != nil`, lines 173-276):

1. Keep all OFF-tx preflight unchanged: `isReasonValid`, `s.docs.CodeExists`, `s.tplCheck.GetTemplateVersionState`, `controlleddocumentsdomain.Resolve`, `s.ensureTemplateArtifact`, governance-event marshalling, `controlleddocumentsdomain.NewVisibility`, `controlleddocumentsdomain.NewControlledDocument`.
2. Replace the bare `s.docs.Create(ctx, doc)` call with an `s.runner.Do(ctx, func(tx *sql.Tx) error { … })` block that:
   - calls `authz.SeedTxIdentity(ctx, tx, cmd.TenantID, cmd.ActorUserID)`,
   - calls `authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), cmd.ProcessAreaCode)`,
   - calls `s.docs.CreateTx(ctx, tx, doc)`.
3. Remove the "DELIBERATELY-PRESERVED asymmetry" comment block (lines 174-178); the asymmetry is now closed.

No other branch, helper, or caller is changed. No schema change.

## Non-goals (mandatory)

- **Not** changing the auto-code branch in any way (it is already correct; it is the reference).
- **Not** changing `s.docs.Create` or `s.docs.CreateTx` repo signatures or bodies.
- **Not** moving any preflight read inside the tx (H-PRE-1 — keep `CodeExists`, `ensureTemplateArtifact`, `GetTemplateVersionState` OFF-tx).
- **Not** fixing the pre-existing OFF-tx CodeExists race (unique constraint already catches it).
- **No** schema/migration change.
- **No** new repo method, no new service method.
- **Not** touching the HTTP handler.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Non-system-admin actor holding `controlled_documents.create` in the area → manual-code `Create` returns nil error and a `*CreateResult` with the CD persisted | `TestCreate_ManualCode_NonAdminWithCap_Succeeds` (new, `//go:build integration`, real Postgres) | **real** |
| Non-system-admin actor **without** the cap → manual-code `Create` returns `authz.ErrCapDenied` | `TestCreate_ManualCode_NonAdminWithoutCap_Denied` (new, same integration file) | **real** |
| System-admin actor → manual-code `Create` succeeds (bypass path still works) | `TestCreate_ManualCode_SystemAdmin_Succeeds` (new, same integration file) | **real** |
| No regression in the auto-code branch | `go test -tags integration ./internal/modules/controlleddocuments/...` and existing controlleddocuments unit tests green | mixed |
| Whole-repo green | `go build ./...` clean; `go test ./...` no `FAIL` | mixed |

> TDD: write the three integration tests first; **non-admin-with-cap MUST fail on current code**
> (proves B2 reproduces — `ErrActorContextMissing`); then apply the symmetric refactor; all three
> integration tests must pass.

## ADR needed?

- [x] No durable decision — skip. The mission + ADR 0022 Phase 7 already authorize area-scoped tier-2
  on CD create; this conforms the manual branch to that existing decision. The change is recorded
  in this feature's spec/evidence and in the F0.1-style commit message.
