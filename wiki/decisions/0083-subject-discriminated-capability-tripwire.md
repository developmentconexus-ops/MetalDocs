# ADR 0083 — Subject-discriminated capability tripwire for shared kernel tables (extends ADR 0082)

> **Status:** Accepted 2026-07-12
> **Extends:** [ADR 0082](0082-approval-kernel-extraction.md) — completes the subject generalization
> of `approval_instances` down to the authz last line of defense.
> **Relates to:** [ADR 0022](0022-authorization-capabilities-not-roles.md) (two-tier PDP + DB
> tripwire), GMR M2 authz-enforcement-generation validation-contract (tripwire arm parity/drift).
> **Scope:** ROADMAP unit 3.1 / approval-remediation M3 — the `internal/platform/tripwire` arm model
> and the generated `enforce_capability_asserted()` trigger.

## Context

ADR 0082 promoted `approval` to a first-class module and generalized `approval_instances` from a
document-only table to a **shared kernel table** keyed by `(subject_kind, subject_key)` with
`subject_kind ∈ {document, template}`. Migrations 0296/0297 generalized the table's columns, indexes,
and CHECK constraints — but the capability tripwire that guards inserts was **not** reached:

- The DB trigger `enforce_capability_asserted()` (ADR 0022's last line of defense) hardcodes
  `approval_instances INSERT → ARRAY['document.submit']`.
- Its Go source of truth `internal/platform/tripwire.TripwireArms` arm #1 is
  `{Table: approval_instances, Op: INSERT, Caps: [CapDocumentSubmit]}`.
- `RenderMigration()` regenerates the trigger SQL byte-for-byte from that arm, and the
  `TRIPWIRE-ARM-PARITY` api-lint rule pins the two together.

When the `templates` module submits a template version for approval, it must assert a **tenant-scoped,
area-blind** capability — templates are not process-area-scoped entities, so an `ScopeArea` capability
asserted with the `"tenant"` sentinel is (correctly) rejected by the `authz-area-scope-binding` lint
(ADR 0022 Phase 7). The registry already carries the right capability: `CapTemplateSubmit`
(`template.submit`, `ScopeTenant`). But the document-only trigger arm rejects it at runtime with
`P0001 ErrCapabilityNotAsserted`. Neither capability choice satisfies both gates.

### Why the obvious widen is a security regression

The trigger matches **match-one**: any single capability listed in the arm that is present in
`metaldocs.asserted_caps` satisfies the branch. Widening arm #1 to
`[document.submit, template.submit]` would therefore let a principal holding **only** `template.submit`
authorize a **document**-subject insert (and vice versa). On a shared table, a flat match-one arm
cross-contaminates the two subjects' capability requirements. The `Arm` model
`(table, op) → caps` has no column-value discriminator and structurally cannot express
"document rows require `document.submit`; template rows require `template.submit`".

Template-only tables (`templates_template`, `templates_template_version`) do not have this problem
because they are not shared with documents — their coarse whole-lifecycle arms never authorize a
document write.

## Decision

**Extend the tripwire arm model with an optional subject discriminator**, so a single gated
`(table, op)` can carry per-`subject_kind` capability requirements, and generate a nested
`CASE NEW.subject_kind` in the trigger.

1. `tripwire.Arm` gains an optional discriminator (`WhenColumn` + `WhenValue`, e.g.
   `subject_kind`/`template`). Un-discriminated arms render exactly as before (byte-identical for
   every existing arm — no behavior change for documents or any other gated table).
2. `approval_instances INSERT` is expressed as two discriminated arms:
   - `subject_kind = 'document'` → `[document.submit]`
   - `subject_kind = 'template'` → `[template.submit]`
3. `RenderMigration()` emits, for a table carrying discriminated arms, a nested
   `CASE NEW.<WhenColumn> WHEN '<WhenValue>' THEN v_required_caps := ARRAY[...] ... END CASE`
   inside that table's outer branch. Match-one semantics are preserved **within** each subject's
   capability set; the two sets never union.
4. A new forward migration (`0284`) `CREATE OR REPLACE`s the trigger from the regenerated SQL; the
   golden-file pointer `tripwireMigrationPath` (`scripts/api-lint/tripwire_arm_rules.go`) advances to
   it. `TRIPWIRE-ARM-PARITY` and `TRIPWIRE-ARM-DRIFT` remain blocking and green.
5. `TemplateSubmitService` asserts `CapTemplateSubmit` with the `"tenant"` area-blind sentinel —
   api-lint clean (a `ScopeTenant` capability), trigger-accepted (template arm).

The **security property is strengthened, not weakened**: a document insert still requires exactly
`document.submit`; a template insert requires exactly `template.submit`. There is no cross-subject
capability path.

## Consequences

- **Amends the GMR M2 validation-contract arm set** (flagged HS-7 in `arms.go`). This ADR is the
  ratifying record for that amendment: arm #1 becomes two subject-discriminated entries; the arm count
  and the golden migration advance accordingly. Operator-ratified 2026-07-12.
- The same discriminator mechanism is the intended path for the parallel `approval_signoffs` gap
  (M3 P3.S2b-3b-iii): signoff has no direct `subject_kind` column, so its discriminator resolves via
  the parent instance — tracked as a follow-on, not closed here.
- No change to any non-`approval_instances` arm; documents, controlled_documents, iam, taxonomy, and
  tenant-lifecycle arms render byte-identically.
- Reversible: dropping the discriminated arms and restoring the flat `[document.submit]` arm reverts
  the trigger — but that would re-break template submit, so the extension stands for the life of the
  shared-kernel design.

## Alternatives rejected

- **Coarse match-one union** (`[document.submit, template.submit]` on one flat arm) — security
  regression (cross-subject capability authorization). Rejected.
- **Keep `document.submit` as the sole physical-table capability** and have templates assert
  `CapDocumentSubmit` — dead-ends on the `authz-area-scope-binding` lint (templates have no process
  area to scope an `ScopeArea` capability). Rejected.
- **Reclassify `CapDocumentSubmit` to `ScopeTenant`** — would silently weaken area-scoped enforcement
  for the document submit path across the whole system. Rejected.
