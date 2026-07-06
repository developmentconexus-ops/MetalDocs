# ADR 0052 — Template revisions are created manually, never auto-spawned on approve/publish

- **Status:** Accepted
- **Last verified:** 2026-06-30
- **Date:** 2026-06-30
- **Scope:** Removes the automatic creation of a next-version draft when a template version is approved or published. After this decision, `CreateNextVersion` (`POST /api/v1/templates/{id}/versions`) is the **only** path that starts a new template revision. Also records the lifecycle status-vocabulary alignment (`in_review` → `under_review`) shipped alongside it.
- **Supersedes:** The `feat/templates-approve-next-draft-response` behaviour (2026-05-31) that had `Approve` (accept path) and `PublishTemplateVersion` auto-spawn the next draft and return it as `next_draft` / `next_draft_id` / `next_draft_version_num`. That was a feature-branch decision (documented in the `wiki/modules/templates.md` changelog), not a prior ADR.
- **Amends:** ADR [`0013`](0013-template-revision-labels.md) — this decision amends 0013's *version-creation trigger* (removes the implicit auto-spawn on approve/publish; `CreateNextVersion` is the sole trigger going forward). It does **not** touch 0013's `revision_number` persisted-column mechanics or `REV{nn}` labeling — those stay exactly as 0013 specified (the counter allocation in `internal/modules/templates/repository/postgres.go` is unchanged by this ADR).

---

## Context

A controlled document and a controlled template are both **governed artifacts**: a published revision is frozen, and starting a *new* revision is a deliberate, audited act. Documents already model this correctly — publishing a document transitions status only; the next revision is created on explicit user action (`createRevision`). Templates had diverged: on the **accept** path of `Approve` and on `PublishTemplateVersion`, the service automatically byte-copied the just-published version into a fresh `draft` (`spawnNextDraft`) and returned it on the response envelope.

This asymmetry was the bug, not a feature:

1. **It manufactures governance state nobody asked for.** Every approval silently created a new editable draft version row. In a QMS, a revision is opened when the process owner decides to revise — not as a side effect of closing the previous one. The auto-draft invites accidental edits to a version that was never meant to exist yet.
2. **It breaks parity with documents.** The shared controlled-artifact view layer (ADR 0053) renders both kinds through one shell; divergent lifecycle semantics would leak `kind` branching into that shell or surface a phantom "draft v+1" for templates only.
3. **It coupled a state-write to a response-shape obligation.** `next_draft*` fields forced every approve/publish caller to reason about a second version it did not request.

The machinery to create a next version already existed and is correct (`nextVersionNumber` + `spawnNextDraft`); the defect was *who triggers it*. The fix is to stop calling it implicitly, not to delete it.

---

## Decision

1. **`Approve` (accept path) no longer spawns a next draft.** `internal/modules/templates/application/lifecycle.go` — the accept branch transitions the version to `published`, updates the template head pointers, appends the audit event, and returns `ApproveResult{Version}` only. `ApproveResult.NextDraft` is removed.

2. **`PublishTemplateVersion` no longer spawns a next draft.** Same file — publish transitions status only; `PublishTemplateVersionResult.NextDraft` (and the `next_draft_id` / `next_draft_version_num` it carried) is removed.

3. **`CreateNextVersion` (`POST /api/v1/templates/{id}/versions`) is the sole revision path.** It is the only caller of `spawnNextDraft` now. `nextVersionNumber` + `spawnNextDraft` are retained unchanged — they were always the correct primitive; only the implicit invocation from approve/publish was wrong.

4. **Contract shrink (forward, additive-removal).** `api/openapi/v1/openapi.yaml` drops `next_draft_id` / `next_draft_version_num` from the `PublishTemplateVersion` 200 response and `next_draft` from `ApproveTemplateVersionResponse`. Both `templates` `api.gen.go` and the frontend `lib/api-types` were regenerated; `oapi-codegen` shows **no drift**. Expand/contract ordering was honoured: the frontend stopped consuming `next_draft*` and stopped auto-navigating to it (commit `ca467cb9`) **before** the backend removed the fields (commit `0d7bfe55`, contract `8240b84f`).

5. **Lifecycle status vocabulary aligned (related, shipped together as M1·T3).** The template-version status `in_review` was renamed to `under_review` to match the documents vocabulary the shared shell depends on: a forward-only expand/contract DB migration (add `under_review` to the CHECK, backfill `in_review` → `under_review`, drop `in_review`), `domain.VersionStatusInReview` value → `"under_review"`, OpenAPI enum + regen, and the frontend literals (`features/templates/lib/canActOnVersion.ts`, gates, badges).

---

## Consequences

### Positive
- **Parity restored.** Templates and documents now share one lifecycle semantic: publish transitions state; a new revision is an explicit, audited act. This is the precondition for the shared controlled-artifact view layer (ADR 0053) to render both kinds with **zero `kind` branching**.
- **No phantom drafts.** Approving/publishing a template produces exactly one published version and no extra rows. Verified by integration e2e (publish ⇒ 1 version; manual create ⇒ draft v2) and the flipped lifecycle unit tests.
- **Leaner contract.** Approve/publish responses no longer carry a version the caller did not ask about.

### Negative / trade-offs
- **One extra user action to revise.** Opening a new revision is now a deliberate "Criar nova versão" click rather than an automatic by-product of publishing. This is the intended QMS behaviour, not a regression — but it is a workflow change for anyone who relied on the auto-draft.
- **The 2026-05-31 auto-next-draft feature is reverted.** Any client code written against `next_draft*` must migrate to calling `CreateNextVersion`. The only such client was the template editor frontend, migrated in this branch.

---

## References

- ADR 0022 — AuthZ = capabilities, never roles (the approve/publish in-tx `authz.Require` checks are unchanged by this decision).
- ADR 0030 — template version state port.
- ADR 0013 — template revision labels (REV{nn}); the published revision counter surfaced on approve is unchanged.
- ADR 0053 — shared controlled-artifact frontend view layer (the parity consumer of this decision).
- REQ-API-3 — breaking changes require contract discipline; the `next_draft*` removal is a contract-first regen with no drift.
- `internal/modules/templates/application/lifecycle.go` — `Approve` (~:198), `PublishTemplateVersion` (~:332), `nextVersionNumber` (~:435), `spawnNextDraft` (~:450).
- Implementation commits: `0d7bfe55` (remove auto next-draft spawn), `8240b84f` (drop `next_draft` from contract + regen), `e11dcf6f` (drop dead `NextDraft` result field), `ca467cb9` (FE stops consuming `next_draft`), `654a55b5` + `b8087dc7` (status `in_review` → `under_review` migration + contract).
