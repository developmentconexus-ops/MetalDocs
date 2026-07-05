# M2 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M2 (authz enforcement generation & cap coherence)
> **Authored:** 2026-07-03, **before any implementation** (mission D4). Committed before the first code change.
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7). The `TripwireArms` table below and the drift-check RED/GREEN
> definitions are the load-bearing clauses.

---

## 0. Runtime-truth basis (the census this contract is built on)

All arm claims below are traced to source, not to the prior migration comments. Authoritative current
trigger: `db/migrations/0270_templates_template_tripwire_archive_cap.sql` (19 CASE branches; supersedes
0269 → 0259). Capability registry: `internal/modules/iam/domain/model.go` (34 caps, pinned by
`TestCapabilityRegistrySize`). Tenancy/identity: every gated write seeds identity via
`authz.SeedTxIdentity` then calls `authz.Require(ctx, tx, cap, area)` before the mutating SQL; the
trigger reads GUC `metaldocs.asserted_caps` (JSON array of `{cap,area}`) and matches **cap-only**,
match-one, fail-closed (`P0001`) on the ELSE.

### Census — actual cap-union reaching each gated write vs. the live 0270 arm

| Gated `(table, op)` | Actual cap-union that reaches a write (source) | Live 0270 arm | Delta |
|---|---|---|---|
| approval_instances (INSERT) | `document.submit` (+`document.edit` co-asserted, submit_service.go:98/101) | `{document.submit}` | match (edit is a co-asserted superset, not a gap) |
| approval_signoffs (INSERT) | `document.signoff` (decision_service.go:230) | `{document.signoff}` | match |
| documents (INSERT) | `document.create` (repository.go:145→181) | `{document.create}` | match |
| **documents (UPDATE)** | **`document.edit` ∪ `membership.manage` ∪ `document.obsolete`** (see §1.1) | `{document.edit}` | **GAP ×2: `membership.manage`, `document.obsolete`** |
| controlled_documents (INSERT) | `controlled_documents.create` (service.go→repository.go:412) | `{controlled_documents.create}` | match |
| controlled_documents (UPDATE) | `controlled_documents.obsolete`, `controlled_documents.supersede` (service.go:761→repository.go:500) | `{obsolete, supersede}` | match |
| cd_sequence_counters (any) | `controlled_documents.create` (repository.go:616/671, in CD-create tx) | `{controlled_documents.create}` | match |
| document_profiles (any) | `taxonomy.manage` | `{taxonomy.manage}` | match |
| document_process_areas (any) | `taxonomy.manage` | `{taxonomy.manage}` | match |
| document_families (any) | `taxonomy.manage` | `{taxonomy.manage}` | match |
| templates_template (any) | `template.{create,edit,approve,publish,archive}` (submit never writes this table) | `{create,edit,submit,approve,publish,archive}` | match — `submit` is a **harmless retained superset** (see §1.3) |
| templates_template_version (any) | `template.{create,edit,submit,review,approve,publish}` | `{create,edit,submit,review,approve,publish}` | match |
| iam_user_roles (any) | `user.manage` | `{user.manage}` | match |
| user_process_areas (any) | `membership.manage` | `{membership.manage}` | match |
| iam_users (INS/DEL + UPD of scoped cols) | `user.manage` (login-telemetry UPDATEs are column-scoped OUT of the trigger, by 0259 design) | `{user.manage}` | match |
| iam_groups (any) | `user.manage` (no live writer; dormant, pre-emptive) | `{user.manage}` | match |
| iam_group_members (any) | `user.manage` (no live writer; dormant) | `{user.manage}` | match |
| iam_group_roles (any) | `user.manage` (no live writer; dormant; `tenant_id` NULL — scoped via group_id) | `{user.manage}` | match |

**Sanctioned no-cap paths (NOT gaps):** scheduler/janitor writes set `metaldocs.bypass_authz='scheduler'`
(`authz.BypassSystem`) and the trigger exempts them (0270:125–153), logging a `governance_events` row.
Covered: `scheduler_service.go:126`, `MarkSuperseded` (postgres_approval_repository.go:546),
`ExpireStaleSessions` (repository.go:903), `cancel_service.go:143` system branch. These assert no cap
by design and must stay exempt.

**Only one table has a real gap: `documents (UPDATE)`, with two missing caps. Every other arm is
correct-or-superset.**

---

## 1. F2.1 — tripwire arms generated from a Go source of truth + CI drift check

### 1.1 The two latent incidents this feature fixes (HS-6 discovered truth)

Both are **function-local** (`authz.Require` and the mutating SQL in the same Go function) and both
currently fail-closed with `P0001` → 500 for **every** actor (the trigger checks the recorded cap set,
not the role — identical mechanism to the shipped 0269/0270 incidents):

1. **`membership.manage` on documents(UPDATE).** `documents/repository/repository.go:798`
   (`ForceReleaseSession`) and `:828` (`ForceReleaseSessionTx`) assert **only** `CapMembershipManage`
   (deliberate per ADR 0022 Phase 11 F4), then UPDATE `documents.active_session_id`. Recorded cap
   `membership.manage` ∉ `{document.edit}` → `P0001`.
2. **`document.obsolete` on documents(UPDATE).** `documents/approval/application/obsolete_service.go:88`
   asserts **only** `CapDocumentObsolete`, then `:93` runs `UPDATE documents SET status='obsolete',
   revision_version = revision_version + 1`. Recorded cap `document.obsolete` ∉ `{document.edit}` →
   `P0001`, unconditionally.

Contrast (why only these two): the other cross-layer documents writers — publish, supersede, submit,
decision-approve/reject — **defensively co-assert `document.edit`** in the same tx, so `document.edit`
in the arm already covers them. Force-release and obsolete do **not** co-assert edit, so their sole
asserted cap must be in the arm. The fix is **additive** (widen the arm), never a tightening.

### 1.2 The binding `TripwireArms` source-of-truth map (18 gated entries)

The implementation MUST produce a single Go map (name `TripwireArms`, location chosen in `plan.md`;
must be lint-visible to the drift check and reference **registry capability consts**, not raw strings)
that is **exactly** this set. The migration 0271 arm literals are **generated** from it. Divergence of
the shipped map from this table is HS-7.

| # | Gated key (`table`, `op`) | Required-cap arm (canonical, sorted) |
|---|---|---|
| 1 | approval_instances, INSERT | `document.submit` |
| 2 | approval_signoffs, INSERT | `document.signoff` |
| 3 | iam_user_roles, * | `user.manage` |
| 4 | user_process_areas, * | `membership.manage` |
| 5 | documents, INSERT | `document.create` |
| 6 | **documents, UPDATE** | **`document.edit`, `document.obsolete`, `membership.manage`** |
| 7 | controlled_documents, INSERT | `controlled_documents.create` |
| 8 | controlled_documents, UPDATE | `controlled_documents.obsolete`, `controlled_documents.supersede` |
| 9 | cd_sequence_counters, * | `controlled_documents.create` |
| 10 | document_profiles, * | `taxonomy.manage` |
| 11 | document_process_areas, * | `taxonomy.manage` |
| 12 | document_families, * | `taxonomy.manage` |
| 13 | templates_template, * | `template.create`, `template.edit`, `template.submit`, `template.approve`, `template.publish`, `template.archive` |
| 14 | templates_template_version, * | `template.create`, `template.edit`, `template.submit`, `template.review`, `template.approve`, `template.publish` |
| 15 | iam_users, * | `user.manage` |
| 16 | iam_groups, * | `user.manage` |
| 17 | iam_group_members, * | `user.manage` |
| 18 | iam_group_roles, * | `user.manage` (tenant scoped via group_id; `v_tenant_id := NULL`) |

**Only entry #6 changes vs migration 0270.** Every other generated arm literal must be byte-identical
to 0270. The `templates_template` `submit` entry (#13) is a **deliberately retained** harmless superset
(§1.3). Tenancy source per branch (`NEW.tenant_id` / `NEW.actor_tenant_id` for approval_signoffs /
`NULL` for iam_group_roles) and the scheduler-bypass block and the JSONB match loop are preserved
byte-for-byte from 0270.

> **Forward erratum — 2026-07-04 (M6 F6.2):** the `documents, UPDATE` arm (entry #6) additionally
> gains **`document.review`** — the eQMS mark-reviewed workflow asserts it before UPDATEing
> `public.documents` (sets `last_reviewed_at` + next `review_due_at`), so without the arm every
> mark-reviewed UPDATE fail-closes `P0001`, the same additive extension as 0271. Rendered into
> migration **0275** (`0275_documents_update_tripwire_review_cap.sql`), which supersedes 0271 as the
> latest `enforce_capability_asserted()` definition; the M2 golden test and `TRIPWIRE-ARM-PARITY`
> target advance from 0271→0275. This is the intended registry-driven growth of the generate-from-Go
> mechanism (governed by M6 `validation-contract.md` §3), **not drift** of this frozen table — the
> §1.2 table itself is unchanged; this note records the sanctioned M6 extension per the mission §9
> HS-7 surface-don't-hide rule.

### 1.3 Retained-superset & pruning non-goal

`templates_template`'s arm keeps `template.submit` even though no writer asserts submit while writing
that table. **Rationale:** the generation is **additive** this milestone (matches 0269/0270 convention;
zero regression risk). Pruning a cap from an arm is a **tightening** whose safety the census cannot
100% exclude for sqlmock-tested paths, and it belongs to a dedicated arm-hygiene pass. **Bounded defer
→ M9** (governance-hygiene arm-hygiene): prune arm supersets to the exact census union, proven by
integration drives. Trigger to act: M9 F9.x arm-hygiene. Owner: milestone M9.

### 1.4 Migration 0271 — generation faithfulness

- **Expected behavior:** 0271 is a forward-only `CREATE OR REPLACE FUNCTION
  public.enforce_capability_asserted()` **regenerated from 0270** with all 19 CASE branches preserved
  byte-for-byte **except** the `documents … TG_OP = 'UPDATE'` branch, whose `v_required_caps` becomes
  `ARRAY['document.edit', 'document.obsolete', 'membership.manage']`. Scheduler-bypass block, JSONB
  parse/type/match loop, ledger insert, `BEGIN/COMMIT`, `SECURITY DEFINER`, `search_path`, and trigger
  attachments are unchanged. Header documents the two fixed latent incidents (file:line) in the
  0269/0270 house style. No down migration.
- **POSITIVE proof:** a diff of the generated 0271 function body against 0270 shows **only** the
  documents(UPDATE) arm literal changed (plus header/ledger text); the migration applies cleanly on a
  fresh tripwire DB.
- **NEGATIVE proof:** hand-editing any other arm literal in the generated SQL to differ from
  `TripwireArms` makes the **parity check** (§1.5.a) RED.
- **Exit criteria:** 0271 present, forward-only, generated (not hand-typed), diff-minimal vs 0270.

### 1.5 The CI drift check (blocking, `scripts/api-lint`)

Registered in `RunCodeRules`/`RunRegistryRules` so it runs under the existing blocking
`api-design-system-lint` job (`.github/workflows/api-contract.yml`); every violation fails CI
(api-lint `main.go`: no reported-only tier). Two sub-rules, both blocking:

**(a) `TRIPWIRE-ARM-PARITY` — Go map ↔ registry ↔ generated SQL.**
- **Expected:** every capability referenced by `TripwireArms` is a real registry cap
  (`IsValidCapability` / present in `model.go`); every gated `(table,op)` in the generated migration
  0271 has a `TripwireArms` entry and vice-versa; the generated arm literals equal the map (golden).
- **POSITIVE proof:** clean tree → 0 violations.
- **NEGATIVE proof:** (i) a `TripwireArms` entry referencing a non-registry cap → RED; (ii) a
  hand-edit making 0271's SQL arm differ from the map → RED.

**(b) `TRIPWIRE-ARM-DRIFT` — asserted-cap coverage (the incident-class catcher).**
- **Expected:** AST-scan `authz.Require(ctx, tx, <cap>, <area>)` call sites (same const-resolution
  technique as the `authz-area-scope-binding` lint: build constName→value from `model.go`, unwrap
  `string(...)`, resolve the cap arg). For any **function** that both (i) calls `authz.Require(cap)`
  and (ii) runs mutating SQL (`Exec`/`ExecContext` with INSERT/UPDATE/DELETE) against a **gated table
  T** (table name parsed from the SQL string literal after `INTO` / `UPDATE` / `DELETE FROM`), require
  `cap ∈ TripwireArms[T, op]`. Fail otherwise. This is the function-local generalization of the
  existing `checkTripwirePairing`; it extends pairing-*presence* to arm-*membership*.
- **Coverage scope (documented limitation, not a gap in the check):** cross-layer edges (an
  application service asserts the cap; a repository method in a different function runs the SQL —
  e.g. publish/supersede) are **not** function-local and are **not** resolved by this AST rule; they
  are covered by the integration drives (§1.6). The rule catches exactly the incident class that has
  actually shipped (0269, 0270) and the two latent ones (force-release, obsolete) — **all four are
  function-local**. A future call-graph-resolving check is out of scope (HS-2). This limitation is a
  recorded bounded defer.
- **POSITIVE proof:** on the clean tree **post-0271**, 0 violations — because the corrected
  documents(UPDATE) arm now contains `membership.manage` and `document.obsolete`, the force-release
  and obsolete functions pass. (**Pre-0271**, this rule is RED on those two functions — that is the
  proof the rule detects the real latent bug; captured as the pre-fix negative.)
- **NEGATIVE proof (the mission's required synthetic):** add a throwaway function (or fixture) that
  asserts a **new** registry cap and writes a gated table without adding that cap to `TripwireArms` →
  rule RED, naming the file:line, the table, and the unarmed cap. Removing the fixture (or adding the
  arm) → GREEN.
- **Exit criteria:** rule registered + blocking; clean-tree GREEN post-0271; synthetic-unarmed-cap
  RED with captured output; pre-0271 RED on the two latent functions captured as the detection proof.

### 1.6 Integration drives (the only test class that can pin a `P0001` arm regression)

Application/unit tests are sqlmock and cannot exercise the trigger. Two **new** integration drives
(`//go:build integration`, testdb factory, mirroring `tripwire_caps_test.go`) pin the two fixed paths:

- `TestTripwire_ForceReleaseWritesDocumentRow` — an actor holding `membership.manage` force-releases a
  stuck session; the `documents` UPDATE must succeed under the live tripwire.
- `TestTripwire_ObsoleteWritesDocumentRow` — `MarkObsolete` on a published document asserting
  `document.obsolete` must succeed under the live tripwire.
- **POSITIVE proof:** both GREEN against 0271.
- **NEGATIVE proof:** both **RED against 0270** (`ErrCapabilityNotAsserted … none of {document.edit}
  present … on documents`) — captured, proving they pin the real fix, not a tautology.
- The existing `TestTripwire_ReviewerStageWritesVersionRow` (0269) and
  `TestTripwire_ArchiveWritesTemplateRow` (0270) stay GREEN (no regression).
- **Run scope:** targeted `go test -tags integration -run 'Tripwire' ./tests/integration/...` only —
  never the full suite (mission §10). If the box cannot run integration locally, the drives are
  authored and the run recorded as a bounded defer with the run-trigger (M1 env-risk precedent).

### 1.7 F2.1 exit criteria (all required)

`TripwireArms` == §1.2 table exactly · 0271 generated & diff-minimal vs 0270 · `TRIPWIRE-ARM-PARITY`
and `TRIPWIRE-ARM-DRIFT` registered, blocking, GREEN on clean tree post-0271 · synthetic-unarmed-cap
NEGATIVE captured RED · pre-0271 detection NEGATIVE captured RED on the two latent functions · two new
integration drives GREEN post-0271 / RED pre-0271 · 0269+0270 drives + `TestCapabilityRegistrySize`(34)
+ 5 authz lints + `go build ./...` all GREEN.

---

## 2. F2.2 — tier-1 ↔ tier-2 capability-name coherence

### 2.1 Decision (binding): both named divergences are already CODE-resolved; F2.2 = verify + pin + doc-truth

Runtime truth (source-traced) contradicts the 2026-07-03 review's "two open divergences" premise:

- **Force-release.** tier-1 `apps/api/cmd/metaldocs-api/permissions.go:157` → `CapMembershipManage`;
  tier-2 `documents/repository/repository.go:798` & `:828` assert `CapMembershipManage`. **Agree**
  (`membership.manage`). Resolved in ADR 0022 Phase 11 F4 (operator ruling a: tier-2 moved
  `document.edit`→`membership.manage`).
- **Approval-route management.** tier-1 four explicit `/approval/routes*` rows → `CapRouteManage`,
  ordered before the generic `/approval/` fallback; tier-2 route-management service asserts
  `route.manage`. **Agree** (`route.manage`). Resolved in ADR 0022 Phase 11 F4; pinned by
  `permissions_test.go`.

The remaining tier-1(coarse) / tier-2(fine) differences on other `/approval/` operations (tier-1
`document.submit` gate vs tier-2 `document.signoff` on decision, etc.) are the **deliberate coarse/fine
PDP split** the review's Dimension 2 vindicated — a **written intentional exception**, not a defect,
and **out of scope** to "align". This contract records that as the accepted answer to the mission's
"align names OR ADR-record as intentional": **intentional, already ADR-recorded (ADR 0022).**

### 2.2 What F2.2 implements

1. **Verify** (evidence, from source) the two sites' tier-1 == tier-2 cap names as above.
2. **Regression pin** — a test that binds, for exactly the two reconciled routes, the tier-1 route→cap
   resolution to the tier-2 asserted cap, so a future edit that re-diverges either reddens CI. Scope
   is **targeted to these routes** (a blanket tier-1==tier-2 assertion is wrong — coarse/fine
   legitimately differ elsewhere). Prefer extending `permissions_test.go` /
   `permissions_authz_scope_test.go`.
3. **Doc-truth restore** — back-annotate the stale ADR 0022 Phase 7/8 ⚠️-follow-up lines (≈198,
   236–237, 250) as **RESOLVED in Phase 11 F4** (cross-referencing lines 349–351), and correct any
   wiki page that still describes either site as an open divergence. No code-behavior change.

### 2.3 Proof

- **POSITIVE:** source excerpts showing tier-1==tier-2 for both sites; the regression pin GREEN on
  HEAD; all **5 authz CI lints** green; ADR 0022 contains no un-annotated "open divergence" claim for
  either site.
- **NEGATIVE:** a synthetic re-divergence (flip the force-release tier-1 row to a different cap, or the
  tier-2 assertion) makes the regression pin **RED** — captured output.
- **Exit criteria:** both sites verified aligned · regression pin RED-on-divergence / GREEN-on-HEAD ·
  5 authz lints green · ADR 0022 + affected wiki reflect the Phase-11 resolution · coarse/fine
  difference recorded as intentional.

---

## 3. Cross-feature constraints (bind both features)

- **Registry-first:** capability names originate only in `internal/modules/iam/domain/model.go`;
  `TripwireArms` and any new test reference registry consts, never raw cap strings.
- **Additive-only trigger change:** the generated 0271 only **widens** an arm; no arm loses a cap this
  milestone (pruning deferred to M9, §1.3).
- **Forward-only migration:** 0271 has no down; regenerated from 0270, diff-minimal.
- **Blocking CI:** both new lint sub-rules fail the build on any violation (no reported-only tier).
- **Targeted tests only:** no full integration suite; `-run 'Tripwire'` + the two new drives.
- **Separation of powers:** implementation via subagents (sonnet implement/review, haiku mechanical,
  never fable, ≤15 concurrent); main session orchestrates/reviews/commits; the `milestone-validator`
  judges and writes `qa/milestone-qa.md`; the main session flips status only on PASS.
- **Commit after verified work; never push; never commit `docs/release/`; plans dir gitignored.**

## 4. Bounded defers (recorded, with triggers)

| Defer | Rationale | Trigger / owner |
|---|---|---|
| Prune harmless arm supersets (e.g. `template.submit` on templates_template) to the exact census union | Tightening carries residual regression risk for sqlmock-tested paths; additive-only this milestone | M9 arm-hygiene feature; proven by integration drives |
| Cross-layer (assert-in-service / write-in-repository) drift coverage via call-graph analysis | The AST drift check is function-local by design (covers the actual incident class); call-graph resolution is a larger effort (HS-2) | Post-mission static-analysis hardening; integration drives cover cross-layer meanwhile |
| Integration-drive execution if the local box cannot run `-tags integration` | 20-min box constraint / env (mission §10) | Run on CI or a capable box before program close-out; drives are authored regardless |
