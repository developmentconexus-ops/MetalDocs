# Feature F5.6 — authz `effective_to` predicate (re-audit Major #1)

> **Milestone:** 5 — HS-5 remediation · **Feature:** `f5.6-authz-effective-to`
> **Status:** **DECIDED 2026-06-19 — Option A (refute & document) approved by operator, conditioned on
> "behaviour matches industry standards" (affirmed). No predicate change. Documentation + code anchors only.**

## What the milestone asked

Fix `internal/modules/iam/authz/authz.go:124` — change `upa.effective_to IS NULL` to
`(upa.effective_to IS NULL OR upa.effective_to > now())` "to allow time-bounded active memberships",
mirroring the M0/F0.1 `effective_from` fix. Re-audit **Major #1** (`architecture-re-audit-2026-06-16.md`
line 79): *"authz.Require denies time-bounded active memberships (effective_to predicate too strict)."*
Operator (2026-06-19) extended scope to the whole class: authz.go:124 + 5 siblings
(`controlleddocuments/infrastructure/repository.go:152,492`, `user_area_repository.go:155,190`,
`search/.../v2documents/reader.go:95`).

## Investigation — the data model contradicts the finding's premise

The finding assumes `effective_to` is a **future-dated validity-end** (a bitemporal *valid-time*
interval). The schema and the entire write path say otherwise: `effective_to` is a
**soft-delete tombstone** set to the *revoke moment*. "Active" is, by the system's own authoritative
definition, `effective_to IS NULL`.

**Evidence:**

| # | Evidence | Source | Implication |
|---|----------|--------|-------------|
| E1 | Three indexes define active as `WHERE effective_to IS NULL` — incl. **UNIQUE** `ux_user_process_areas_single_active (tenant_id,user_id,area_code,role) WHERE effective_to IS NULL` and `ux_user_process_areas_one_active`, plus `ix_user_process_areas_active`. | `db/baseline/0001_current_schema.sql:3618,3667,3674` | The DB enforces **exactly one active row per (user,area,role)** keyed on `effective_to IS NULL`. A future-dated "scheduled" row could not even coexist with the current row under this unique index. |
| E2 | Revoke sets `effective_to = time.Now().UTC()` (the revoke moment, always ≤ now). | `area_membership_service.go:78,245` → `CloseActiveTx` | No row ever carries a **future** `effective_to`. So `effective_to > now()` matches **zero** rows. |
| E3 | `Grant(ctx, userID, tenantID, areaCode, role, grantedBy)` takes **no** `effective_to` argument. The grant SQL inserts `effective_to = NULL` unconditionally. | `area_membership_service.go:130`, `0143/grant_area_membership`, `routes_memberships.go:219` | There is **no API** to create a time-bounded / scheduled membership. The "time-bounded active membership" the finding worries about cannot be created. |
| E4 | CHECK `effective_interval_valid: effective_to IS NULL OR effective_to > effective_from`; revoke also sets `revoked_by` (CHECK ties the two). | `0001_current_schema.sql:1643-1644`, `0136` | `effective_to` is a revoke marker paired with `revoked_by`, not a scheduled expiry. |
| E5 | The DTO field `EffectiveTo *time.Time` is **read-only output** — it serializes the revoke timestamp of a closed row; it is never an input on grant. | `api.gen.go:300`, `routes_memberships.go:62,75` | The presence of the field misled the audit; it is not a "grant until" input. |

**Why the sibling sites that DO use `(effective_to IS NULL OR effective_to > $n)` are not evidence of a bug.**
`user_area_repository.go:32,68,101,178,591` use the interval form with a **parameter** because they are
**as-of / point-in-time history reads** ("who was active *at time T*") — for a historical T, a row
revoked at R was genuinely active for any T < R, so the interval predicate is correct *there*. The
**active-now** authz checks (authz.go:124 et al.) ask a different question ("active *now*"), whose
correct and index-aligned answer is `effective_to IS NULL`. Two different questions, two correct
predicates — this is intentional, not drift.

## Conclusion — Major #1 is a false positive

Applying the proposed predicate to the active-now sites would:

1. **Change no access** — zero rows have `effective_to > now()` (E2/E3), so the `OR` clause is dead.
2. **Regress the authz hot path** — the active-now reads are served by the partial indexes
   `WHERE effective_to IS NULL` (E1). The `OR effective_to > now()` form is **not sargable against a
   partial `IS NULL` index**, pushing `authz.Require` (called on every capability check) toward seq
   scans.
3. **Contradict the schema's own definition of "active"** — the UNIQUE partial index (E1) *is* the
   system's authoritative "one active membership" rule. Encoding a different active-predicate in the
   read path while the write/index path keeps `IS NULL` is precisely the kind of split-truth this
   program exists to eliminate.

This is an **architecture contradiction**, not a patchable defect → **HS-2**: stop, report the
boundary, do not symptom-patch. Per CLAUDE.md ("runtime/contract truth beats docs; classify the
mismatch and stop") and ADR 0022 ("authz is the boundary; never symptom-patch authz").

## How industry-grade systems model this (for the operator decision)

Two legitimate, mutually-exclusive designs for temporal membership:

- **(A) Soft-delete + "current" flag/marker** — one canonical active predicate (`effective_to IS NULL`
  here), a partial unique index enforcing one current row, history kept as closed rows. Simple, fast,
  index-friendly. **This is the model MetalDocs already implements, consistently.** (Same shape as a
  SCD-2 "current row" view, or Rails `discarded_at IS NULL`.)
- **(B) Bitemporal valid-time** (e.g. Postgres `tstzrange` + exclusion constraint, or SQL:2011
  `PERIOD FOR`) — memberships carry a future-capable `[valid_from, valid_to)` interval; "active now"
  is `now() <@ validity`; scheduled grants/expiries are first-class. Powerful but a **feature**: it
  needs the grant-until API, a redesign of the unique constraints (current-row uniqueness no longer
  means `effective_to IS NULL`), and *every* active-now read switched to the interval form together.

The re-audit's recommendation is a half-step from (A) toward (B) applied to **one** read — which is
the worst of both: it doesn't deliver scheduling (no write path, no index support) and it breaks
model (A)'s invariant. Either stay fully in (A) (recommended) or commit to (B) as its own scoped
mission.

## Recommended decision

**Option A (recommended): Refute & document — no predicate change.** Record Major #1 as an audit
false-positive with the evidence above; make the model legible so it is not re-flagged — add a
one-line canonical comment (or a thin `metaldocs.user_process_area_is_active(effective_to)` STABLE
SQL helper / `active_user_process_areas` view) documenting that active-now ⟺ `effective_to IS NULL`
and pointing at the partial indexes. This closes the *real* root cause (no named/legible canonical
predicate) without changing behavior or regressing indexes. M5's §8 "0 confirmed Majors" is then met
by **refutation with evidence**, the same disposition the workflow already allows ("do not
re-litigate refuted findings").

**Option B: Build time-bounded memberships for real (model B).** Out of M5 surgical appetite — a new
milestone/mission (index redesign + grant-until API + consistent interval reads + revoke-vs-expire
semantics). HS-2/HS-6.

**Option C: Apply the defensive `OR` anyway** at the 6 active-now sites. Advised against — dead clause,
index regression, split-truth vs the unique index.

## Non-goals (Option A)

- No predicate change at any of the 6 active-now sites (would regress indexes + contradict the
  unique index — see Conclusion).
- No DB migration, no API change, no behavior change.
- No view/function added (a Go anchor comment + ADR is sufficient legibility and adds no DB surface);
  a canonical view is recorded as the heavier alternative in ADR 0037 D1 if ever needed.

## Validation Gate (Option A)

1. **No SQL change** — the 6 active-now query strings are byte-identical to HEAD (only Go `//`
   comments added above the query literals + `var granted bool`). Proof: sqlmock exact-match tests
   (`role_admin_repository_test.go`) pass unchanged — they would fail on any in-string edit.
2. **ADR recorded** — `wiki/decisions/0037-membership-temporal-model.md` (Accepted) states the model,
   refutes Major #1 with evidence, and gates any future Model-B work.
3. **Code anchors** — each of the 6 active-now sites carries a Go comment pointing at ADR 0037 and
   stating `effective_to IS NULL` is canonical (not an interval bug), so a future auditor/dev reads
   the rule instead of re-flagging it.
4. **Build** — `go build ./...` clean.
5. **Tests** — `go test ./internal/modules/iam/... ./internal/modules/controlleddocuments/...
   ./internal/modules/search/...` all green.
6. **Wiki + ADR index updated** — `wiki/database/tables/user_process_areas.md` documents the active
   predicate; ADR index lists 0037.
