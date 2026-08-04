# ADR 0088 — Template version content is always materialized

> **Status:** Accepted 2026-08-04 (operator ruling: blank template is submittable
> without an edit, and the same rule governs every version-creation path —
> "fechar o buraco na origem", no patch).
> **Supersedes:** the "leave `ContentHash` empty so the publish gate still forces
> a real edit" stance in `spawnNextDraft`
> (`internal/modules/templates/application/lifecycle.go:166-189`) and the
> emptiness-based gate at `lifecycle.go:72-74`.
> **Relates to:** [ADR 0034](0034-integration-test-fixture-framework.md) (test
> framework), the 2026-07-29 baseline fold (baseline frozen; forward migrations
> only).
> **Scope:** the `templates` module and its object-store port. System-impact
> gate: Yellow
> (`docs/superpowers/analysis/2026-08-04-blank-template-materialization-system-impact.md`).

## Context

A template version row could exist without the object it points at, and without
a `content_hash`. Two paths produced that state deliberately:

- **Blank creation.** The wizard's "start blank" option
  (`frontend/apps/web/src/features/templates/pages/TemplateWizardPage.tsx:139-161`)
  created only the metadata row and navigated to the editor. No bytes were ever
  written. The editor treated the resulting 404 as "blank canvas"
  (`useTemplateDraft.ts:12-14`), deferring materialization to the autosave
  debounce on first edit. A user who created a blank template and submitted it
  before touching it got `409 UPLOAD_MISSING` — "DOCX file not yet uploaded" —
  for a template they never intended to upload.
- **Next revision.** `spawnNextDraft` copies the source object PRE-TX at the new
  draft's canonical key, but deliberately left `ContentHash` empty so the publish
  gate would force an edit.

The root cause is not the missing copy — it is that **`content_hash` carried two
incompatible meanings**: the verified hash of the object a version points at
(`autosave.go:151-165`, CHECK `chk_template_version_content_hash_non_draft`), and
"the user has actually edited this" (`lifecycle.go:72-74`, `spawnNextDraft`).
Expressing the second meaning by *absence of the first* is what made "version
without content" a reachable state by construction. No amount of extra copying
closes that while the overload stands.

## Decision

1. **`content_hash` has exactly one meaning:** the verified hash of the object
   this version points at. It is **always present** on every template version
   row, from the moment the row exists. Nothing infers user intent from it.

2. **Store-then-reference, uniformly.** Every path that creates a template
   version materializes its object first, PRE-TX, at the version's own canonical
   key (`templateDocxKey(tenantID, templateID, n)`), then commits the row that
   references it — the pattern `spawnNextDraft` already established. The only
   crash outcome remains a safe orphan object. No network call inside the
   transaction; no outbox job after commit (that would re-open the very window
   this ADR closes).

3. **Blank creation copies the system asset.** A from-scratch template
   materializes by `Presigner.Copy` from `system/templates/blank.docx` — the
   deterministic asset the stack already seeds (`deploy/assets/system-blank.docx`
   via compose `minio-init`) — then `Presigner.Confirm` yields the real hash that
   the row is born with. The `docx-renderer` is **not** used to synthesize an
   empty document: that would add a `templates → render` edge and a second
   generator competing with the asset the repo already ships.

4. **No "must edit before submit/publish" gate anywhere.** A blank template is a
   valid artifact and may be submitted for review untouched; a spawned revision
   may likewise be submitted without a change. Whether an empty or unchanged
   document deserves approval is a judgment for the reviewer, not a shape the
   system fabricates out of a null column. The emptiness check at
   `lifecycle.go:72-74` is deleted rather than reworded — with (1) in force it is
   unreachable.

5. **DB is the authoritative line — forward migration 0317** (baseline frozen
   since the 2026-07-29 fold): `chk_template_version_content_hash_non_draft` is
   replaced by an unconditional `length(content_hash) = 64`, so a content-less
   version row cannot exist in any status. Existing draft rows are backfilled
   from their object where one exists; rows with no object at all (only
   reachable through the superseded blank path) are materialized from the system
   blank asset or removed — the migration states which, and no live tenant data
   depends on the outcome (the product has not shipped).

## Consequences

- The 409 `UPLOAD_MISSING` on a freshly created blank template becomes
  impossible; the sentinel remains only for a genuinely interrupted upload.
- The editor never opens on a 404 — every draft has bytes behind it, so the
  "treat 404 as blank canvas" branch in `useTemplateDraft` is deleted.
- Autosave keeps its job (committing edits and their hash) and loses its
  accidental second job (being the only thing that ever made a version valid).
- Reviewers may now receive an untouched blank template. That is intended: the
  approval route is where "is this worth approving" is decided.
- One less overloaded field: reasoning about a template version no longer
  requires knowing which of two meanings a null hash carried.

## Rejected alternatives

- **Disable the submit button while `content_hash` is null** — hides the symptom
  and preserves the unreachable-state-by-construction defect. This is the patch
  the ruling explicitly rejected.
- **Generate an empty DOCX in `docx-renderer` on create** — new cross-module
  edge, a network call on the creation path, and a second source of truth for
  "what an empty template is" alongside the shipped asset.
- **Materialize via an outbox job after commit** — leaves a window in which the
  version exists without content, which is exactly the state being eliminated.
- **Keep the "forces a real edit" gate, expressed as hash ≠ source hash** —
  honest, but it re-introduces a second concept the system must carry forever to
  serve a judgment that belongs to the reviewer.
