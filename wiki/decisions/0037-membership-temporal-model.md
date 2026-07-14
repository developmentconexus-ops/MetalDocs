# ADR 0037 — Membership temporal model: soft-delete current-marker (active ⟺ `effective_to IS NULL`)

> **Status:** Accepted
> **Date:** 2026-06-19
> **Deciders:** leandrotca.work (operator), MetalDocs backend
> **Context window:** Grade-A completion program · Milestone 5 (HS-5 remediation) · Feature F5.6 (re-audit Major #1)
> **Supersedes:** none
> **Related ADRs:** [0022 — Authz capability coherence](./0022-authz-capability-coherence.md)
> **Related code (Last verified 2026-06-19):**
> - `internal/modules/iam/authz/authz.go:124` — `Require` active-membership predicate
> - `internal/modules/iam/infrastructure/postgres/user_area_repository.go:155,190` — directory / `has_managed_areas`
> - `internal/modules/controlleddocuments/infrastructure/repository.go:152,492` — CD area-grant visibility
> - `internal/modules/search/infrastructure/v2documents/reader.go:95` — search visibility
> - `internal/modules/iam/application/area_membership_service.go:78,130,245` — Grant / Revoke write path
> - `db/baseline/0001_current_schema.sql:3618,3667,3674` — partial indexes defining "active"
> - `db/baseline/0001_current_schema.sql:1643-1644` — interval + revoked_by CHECK constraints

---

## Context

The 2026-06-16 terminal re-audit (`wiki/backend/_artifacts/architecture-re-audit-2026-06-16.md`,
Major #1) flagged `authz.Require` for "denying time-bounded active memberships": it claimed
`upa.effective_to IS NULL` is "too strict" and should be
`(effective_to IS NULL OR effective_to > now())`. F5.6 was opened to apply that change to
`authz.go:124` and, on operator extension, to five sibling active-membership reads.

Investigation (F5.6 spec.md) found the finding rests on a **false premise about the table's
temporal semantics**. `user_process_areas` does not model a future-capable validity interval; it
models soft-deletion with a current-marker. This ADR records the model as the durable decision so
the finding is not re-raised, and so future temporal work starts from an explicit baseline.

There are two industry-standard designs for temporal membership; they are mutually exclusive:

- **Model A — soft-delete + current marker.** Active ⟺ `effective_to IS NULL`. Revoke stamps
  `effective_to = now()` (a past tombstone) and retains the row for history. No future-dated end is
  ever written. (Same family as SCD-2 "current row", `discarded_at IS NULL` soft-delete, partial
  unique index on the active subset.)
- **Model B — bitemporal valid-time.** `effective_to` is a possibly-future scheduled end; active-now
  is `effective_from <= now() AND (effective_to IS NULL OR effective_to > now())`. Requires range
  types / exclusion constraints, a grant-until API, and the interval predicate at every read.

The proposed predicate is the **Model B** read. MetalDocs implements **Model A**, end to end.

## Decision

**D1. MetalDocs membership is Model A. The canonical "active now" predicate is `effective_to IS NULL`.**

This is the authoritative definition, enforced by the database, not a convention. Evidence:

- **Schema (authoritative).** Partial **unique** index
  `ux_user_process_areas_single_active (tenant_id,user_id,area_code,role) WHERE effective_to IS NULL`
  declares "at most one active row per (user,area,role), where active = `effective_to IS NULL`".
  Two further partial indexes (`ux_user_process_areas_one_active`, `ix_user_process_areas_active`)
  repeat the same `WHERE effective_to IS NULL`. A Model-B future-dated row could not coexist with a
  current row under this unique index — the schema *cannot express* Model B as written.
- **Write path.** `Grant` takes no end-date argument and inserts `effective_to = NULL`. `Revoke`
  sets `effective_to = time.Now().UTC()` paired with `revoked_by` (CHECK
  `revoked_by_required_when_revoked`). No code path writes a future `effective_to`; the set of rows
  with `effective_to > now()` is empty in production.
- **API.** The `effective_to` field on the membership DTO is read-only output (the revoke timestamp
  of a closed row), never a grant input.

**D2. Re-audit Major #1 is refuted with evidence; the active-now predicate is NOT changed.**

Applying `(effective_to IS NULL OR effective_to > now())` to the active-now reads would: (1) change
no access (the `OR` clause matches zero rows — D1); (2) regress the authz hot path, which `Require`
runs on every capability check, off the `WHERE effective_to IS NULL` partial indexes (the OR-form is
not sargable against a partial `IS NULL` index); (3) contradict the unique index that *is* the
system's definition of "active" — introducing exactly the read/write split-truth this Grade-A
program exists to eliminate. It is therefore rejected as a symptom-patch against an architecture
contradiction (HS-2; CLAUDE.md runtime-truth-beats-docs; ADR 0022 never-symptom-patch-authz).

**D3. As-of / history reads correctly use the interval form — that is not drift.**

`user_area_repository.go:32,68,101,178,591` use `(effective_to IS NULL OR effective_to > $n)` with a
**parameter** because they answer a different question — "who was active *at past time T*". For a
historical T, a row revoked at R was active for any T < R, so the interval predicate is correct
there. Active-now and as-of are two different questions with two different correct predicates. The
distinction is intentional.

**D4. Adopting Model B is a separate, scoped program — not a predicate flip.**

If time-bounded / scheduled memberships become a product requirement, the work is: redesign the
partial unique indexes (current-row uniqueness can no longer mean `effective_to IS NULL`), add a
grant-until API + validation, switch **all** active-now reads to the interval form together, and
define revoke-vs-expire semantics. That is its own milestone/mission, gated by this ADR, not an
in-place edit.

## Consequences

### Positive
- One definition of "active", enforced once by the DB and matched by every active-now read.
- Authz hot path stays on its partial indexes (index probe, not scan).
- Full revoke history retained (`effective_to` + `revoked_by` on closed rows) for compliance.
- The model is now named and discoverable — future audits read the ADR + the code anchor comments
  instead of re-flagging the predicate.

### Negative
- The `effective_to` output field on the DTO can still mislead a reader into expecting Model B; the
  anchor comments and this ADR are the mitigation.

### Neutral
- No behavior change, no migration, no API change. F5.6 lands documentation + code comments only.

## Verification

A change is ADR-0037-compliant when:
- Active-now membership reads use `effective_to IS NULL` (optionally `effective_from <= <at>` for the
  lower bound), matching the partial indexes — **not** the interval form.
- As-of / history reads that take a point-in-time parameter use
  `(effective_to IS NULL OR effective_to > <param>)`.
- Any move toward future-dated memberships is gated by a successor ADR that redesigns the unique
  indexes and the write path together (D4).
