# Concept: ISO Segregation of Duties

> **Last verified:** 2026-05-04
> **Status:** Stub. Add full enforcement points + edge cases when audited.
> **Scope:** Why and how the platform enforces that submitters cannot approve their own work.
> **Out of scope:** Error messages shown to users when SoD is violated (see `concepts/error-ux.md`).
> **Key files:**
> - `frontend/apps/web/src/features/approval/components/SignoffDialog.tsx:17` — `error_sod_submitter` / `error_sod_duplicate` dialog states
> - `frontend/apps/web/src/features/approval/api/mutationClient.ts:9` — `ApprovalError` thrown on 403 with SoD code

## Why

ISO 9001 (and 14001 / 45001 / 27001) require segregation of duties for controlled documents:

- The person who **wrote/submitted** a document cannot also **approve** it.
- This prevents single-actor compromise of the controlled-document chain.
- Auditors verify this during certification audits.

## What we enforce

### Documents

- A user who submits a document version (`draft → under_review`) is **blocked** from recording any signoff on that same version.
- Enforced at the API layer in the approval module's decision service. UI also hides the Aprovar button for the submitter, but UI is just a nicety — the API is the source of truth.

### Templates

- Same rule for template versions — author of a version cannot approve it.

### Multi-stage approval

If a route has multiple stages with the same approver, the approver can sign off in each stage they're listed in **as long as they were not the submitter**. Segregation is submitter-vs-approver, not approver-vs-approver.

## What we do NOT enforce (yet)

- "Author cannot also be reviewer" beyond the strict submitter rule (verify).
- "Author cannot edit the template their document is based on" — these are separate workflows.

## Edge cases to watch

- Resubmission after rejection — the submitter on the new submission is the one who clicked submit, not the original author. Verify behavior matches policy.
- Approval delegation — TBD, currently no delegation mechanism.

## See also

- [modules/approval.md](../modules/approval.md)
- [modules/iam.md](../modules/iam.md)
- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 8 (D2 in smoke routine validates this)
- [concepts/error-ux.md](error-ux.md) — Portuguese error messages for `sod.submitter_cannot_sign` / `sod.cross_stage_duplicate`; E2 SoD dialog states in `SignoffDialog`
