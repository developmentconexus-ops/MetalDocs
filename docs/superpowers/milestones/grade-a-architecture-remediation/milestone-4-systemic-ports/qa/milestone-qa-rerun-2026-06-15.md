# Milestone 4 — Validation Verdict (C1–C7) — RE-RUN 2026-06-15

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-15 (re-dispatch)  ·  **HEAD:** `bfdee5040de7` ·  **Verdict:** see C7.
> The validator judges and writes this file; the **main session flips status only on a PASS**. It never
> edits code, fixes findings, or flips status.

## Why this is a re-run

The original M4 verdict (`milestone-qa.md`, HEAD `30503533`) returned **FAIL** on a single isolated
gate: **F4.1a Gate #5** (`TestCreateDocumentTx_PopulatesAllSnapshotColumns`) was environment-coupled —
it passed only when the operator DSN omitted `search_path`, and FAILED with the operator DSN
`…&search_path=metaldocs,public` (`column "tenant_id" of relation "documents" does not exist`, SQLSTATE
42703). Root cause: a dead legacy `metaldocs.documents` duplicate shadowed the real `public.documents`
under any metaldocs-first search_path. **M4b migration 0240 (`071931c9`) dropped that cluster.** This
re-run re-judges the milestone from clean state with that root cause now fixed.

**Pre-flight root-cause verification (live DB, dev Postgres `127.0.0.1:5433/metaldocs`, role `metaldocs_app`):**

| Check | Command | Result |
|-------|---------|--------|
| Legacy `metaldocs.documents` dropped | `SELECT count(*) FROM information_schema.tables WHERE table_schema='metaldocs' AND table_name='documents'` | **0** (absent) ✅ |
| Real `public.documents` present | same, `table_schema='public'` | **1** ✅ |
| `metaldocs.template_audit_log` dropped | same, `table_name='template_audit_log'` | **0** ✅ |
| Bare `documents` under `search_path=metaldocs,public` resolves to | `SET search_path TO metaldocs,public; SELECT n.nspname … WHERE relname='documents' AND pg_table_is_visible` | **public** ✅ |
| That visible table has `tenant_id` | `information_schema.columns … table_name='documents' AND column_name='tenant_id' AND table_schema='public'` | **1** ✅ |
| Migration 0240 applied | `SELECT version FROM public.schema_migrations WHERE version LIKE '%0240%'` | **`0240`** present ✅ |
| `db.go` UNMODIFIED (proves schema-fix, not harness-patch) | `git log -1 -- tests/integration/testdb/db.go` → `1b04c11f` (pre-M4, unrelated); `git status --short tests/integration/testdb/db.go` → clean | **unchanged at HEAD** ✅ |

All real (live PG + git). The root cause is fixed at the schema level; no harness change.

---

## C1 — Spec & plan conformance (per feature)

Inputs loaded: `milestone.md`; each feature `spec.md`/`evidence.md` (f4.1, f4.1a, f4.2, f4.3, f4.4,
f4.5, f4.6); program `README.md`; governing spec
`docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`; aggregate diff. All
present and readable.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4.1 user-display-name-reader | ✅ port shape read from the 3 consumers (approval signoff, get-instance batch, documents create); iam-owned, single+batch | ✅ (gate #5 see C2 — now re-greens) | ✅ no snapshot/freeze; no OpenAPI/route change; off-tx preserved | `f4.1-*/evidence.md` |
| F4.1a documents-testdb-rehab | ✅ consumer = documents integration tests; test-harness-only | ✅ (root cause now fixed by M4b 0240; re-verify in C2 with operator DSN) | ✅ no production/schema/migration change | `f4.1a-*/evidence.md` |
| F4.2 template-version-state-reader | ✅ extends existing templates-owned `TemplateVersionPort`; CD `TemplateVersionChecker` `(status, doc_type_code)` shape preserved | ✅ | ✅ `IsPublished` untouched; no OpenAPI/route; no snapshot | `f4.2-*/evidence.md` |
| F4.3 port-adrs | ✅ docs-only | ✅ ADR 0029/0030 (+0031 via F4.5) present, headers, indexed, cross-linked | ✅ no code | `f4.3-*/evidence.md` |
| F4.4 auth-session-display-name-port | ✅ auth narrowed to auth-owned rows; iam consumer enriches via port; `DisplayName` removed from `SessionListItem` | ✅ | ✅ no OpenAPI/route; display_name key+value preserved consumer-side | `f4.4-*/evidence.md` |
| F4.5 iam-tenant-membership-port | ✅ contract read from F4.6's 3 coupled queries (membership id-set, no `deactivated_at` filter) | ✅ | ✅ producer-only; no consumer wiring (F4.6's job) | `f4.5-*/evidence.md` |
| F4.6 security-display-name-port | ✅ `securitydomain.Repository` + structs + Service + handler + OpenAPI unchanged; rows/names byte-identical | ✅ | ✅ `MfaCoverage`/`CountRecentLockouts` deferred accurately | `f4.6-*/evidence.md` |

**C1 result: PASS** — every feature's acceptance maps to its spec gate; consumer contracts were read
from the consumers. The original validator's only C1 caveat (F4.1a) was a C2 re-run issue whose root
cause is now fixed (M4b 0240); re-verified below.

<!-- C2 appended below as completed -->

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (not trusted from transcripts). Live dev Postgres
`127.0.0.1:5433/metaldocs` (role `metaldocs_app`), `-tags integration -count=1`. **The critical
re-check (F4.1a Gate #5) was run with the operator DSN `…?sslmode=disable&search_path=metaldocs,public`
— the exact form that FAILED in the original verdict — with `db.go` UNMODIFIED.**

| Feature / gate | Command re-run (validator) | Real output | Pass? |
|----------------|-----------------------------|-------------|-------|
| Build | `go build ./...` | exit 0 | ✅ |
| Vet | `go vet ./...` | exit 0 | ✅ |
| **F4.1a Gate #5 (operator DSN, `search_path=metaldocs,public`)** | `go test -tags integration -count=1 -p 2 -run TestCreateDocumentTx_PopulatesAllSnapshotColumns ./internal/modules/documents/application/` | **`ok … 2.186s`, exit 0** — asserts `created_by_display_name_snapshot == "Snapshot Author"` under real `iampg.NewUserDisplayNameRepository`. **Was `FAIL` (SQLSTATE 42703) in the original run with this same DSN.** Now PASS with **NO db.go change** → schema-level fix (M4b 0240), not symptom-patch. | ✅ **real (live PG)** |
| F4.1 iam port | `go test -tags integration -count=1 -run 'TestUserDisplayNameRepository_DisplayName' ./internal/modules/iam/infrastructure/postgres/` | `ok 0.575s` | ✅ **real (live PG)** |
| F4.1 approval off-tx (H-PRE-1) | `go test -tags integration -count=1 -run TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema ./internal/modules/documents/approval/repository/` | `ok 0.299s` | ✅ **real (live PG)** |
| F4.2 templates port | `go test -tags integration -count=1 -run TestTemplateVersionReader_GetTemplateVersionState_Live ./internal/modules/templates/infrastructure/` | `ok 1.538s` | ✅ **real (live PG)** |
| F4.4 sessions no-join | `go test -tags integration -count=1 -run TestListActiveSessions_NoIamUsersJoin ./internal/modules/auth/infrastructure/postgres/` | `ok 2.733s` | ✅ **real (live PG)** |
| F4.5 tenant-membership | `go test -tags integration -count=1 -run TestTenantUserRepository_TenantUserIDs_Live ./internal/modules/iam/infrastructure/postgres/` | `ok 0.282s` | ✅ **real (live PG)** |
| F4.6 security no-iam-join | `go test -tags integration -count=1 -run TestSecurityRepository_NoIamUsersJoin_Live ./internal/modules/security/infrastructure/postgres/` | `ok 0.317s` | ✅ **real (live PG)** |
| Whole-repo unit | `go test ./...` | all `ok` (61 packages), 0 FAIL | ✅ |

**C2 result: PASS.** Every M4 per-feature live gate re-ran green from clean state under the operator
DSN — **including F4.1a Gate #5, the single gate that FAILED the original verdict**. It now passes with
the operator DSN unchanged and `db.go` unmodified, confirming the original FAIL's root cause (legacy
`metaldocs.documents` shadow) is fixed at the schema layer by M4b migration 0240, not patched in the
harness. No environment-coupling remains.

## C3 — Senior review of the aggregate milestone diff

Whole-milestone production diff (`2e7e2009..30503533`, the M4 feature commits; M4b/M4c follow as
separate milestones with their own gates) reviewed as one unit — 45 files, +1836/−247.

- **Single owning adapter (no split-brain):** grep-verified zero `iam_users.display_name` reads outside
  `iam/`. The only remaining cross-module `iam_users` SQL in production is
  `security/infrastructure/postgres/repository.go:67,80` — `MfaCoverage` `COUNT(*) FILTER`/by-role
  aggregate (validator read lines 63–117: aggregates over `mfa_enabled`/`deactivated_at`, **no**
  `display_name`) — the accurately-characterized bounded defer, not a display-name reach. Inside `iam/`
  the display-name port impl `user_display_name_repository.go` is the single owning adapter; the other
  iam files reading `iam_users` (observability, role_provider, tenant_user_repository, presence) are
  distinct intra-module concerns, not a second source of truth for the cross-module display-name fact.
- **No dead code from superseded approaches:** F4.2 deleted `PostgresTemplateVersionChecker` (grep → 0
  references; the lone `:698` hit is a doc comment). F4.4 removed `DisplayName` from auth's
  `SessionListItem` struct (verified: struct now has only auth-owned fields). No orphaned superseded
  code.
- **No feature broke another:** whole-repo `go test ./...` all `ok`; all three touched modules' (auth,
  iam, security) live integration gates green; templates + documents gates green.
- **GitNexus index note:** the GitNexus advisory flagged `PostgresTemplateVersionChecker` /
  `NewPostgresTemplateVersionChecker` as live — this is a **stale index** (predates the F4.2 deletion).
  Authoritative live grep returns 0 references. Recorded so the next agent does not trust the stale
  call-graph.
- **Minor repetition (retrospective, not blocking):** the `missing→user_id` presentation fallback is
  re-implemented per consumer (`security.resolveNames`, iam `sessions_handler.resolveDisplayNames`).
  This is the deliberate ADR-0029 design (port returns raw display_name; each consumer owns its
  presentation fallback) — not a duplicated *fact*. C5 retrospective input.

- Findings: none blocking.
- Staff-engineer bar met? ✅ — owning-module ports, deleted dead code, preserved consumer contracts,
  reads live + off-tx; the original FAIL's root cause is fixed at the schema layer (M4b 0240) with the
  harness untouched.

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| backend-api-qa-checklist | pass | Internal module-port swaps only; no route/OpenAPI/generated-surface/authz change (F4.6 leaves OpenAPI + `securitydomain.Repository` untouched; F4.2 leaves CD's `TemplateVersionChecker` interface untouched; F4.4 no route change, struct field removed). Shared consumers preserved + green. |
| workflow-async-qa-checklist (F4.2 CD-create lock-bearing; F4.1 signoff lock-bearing) | pass | H-PRE-1 intact: F4.1 approval display-name read proven **off-tx** by `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema` (live PG, PASS); F4.2 template-status read is on the pool conn, non-authz, call site unmoved → not inside the CD-create lock. No authz-recording read inside a lock-holding tx. |
| Regression — whole-repo unit | all pass | `go test ./...` → all `ok` (61 packages), 0 FAIL |
| Regression — M0 (docs-destaling) | no break | doc-surface milestone; tree-wide build/vet clean, no doc structure touched by M4 source. |
| Regression — M2 (contract-tail / FE regen) | no break | No OpenAPI/route/generated-type change in M4 (verified C1/C3); `go build`+`go vet` clean tree-wide. |
| Regression — M3 (mechanical-quality: orphan deletes, tx-hazard hoist) | no break | Build clean (no reintroduced orphans); H-PRE-1 off-tx hoist preserved (live off-tx gate PASS). |
| Regression — M4c (test-fixture framework) | no break | M4c PASS on disk; the factory-migrated tests this validator re-ran (Gate #5, repository_create path consumers) are green under the operator DSN. |
| Regression — M1 full-HTTP `seed→finalize→signoff` E2E | **SKIPPED (not re-proven)** | `TestE2E_HappyPath_HTTP` needs a running server (`METALDOCS_E2E_URL`) the validator did not stand up. Recorded as a **SKIP, not a pass** (fail-closed). Mitigation: the signoff display-name code path F4.1 touches is independently re-proven by the live off-tx integration test; the full-HTTP E2E was discharged as the M1 HS-1 condition on 2026-06-14 (program README line 34) and M4 introduces no route/contract change to that path. |

**C4 result: PASS-with-noted-SKIP.** No prior milestone regressed. The M1 full-HTTP E2E is explicitly
recorded as a SKIP (server not stood up), not counted as a pass; its risk surface (signoff display-name)
is covered by a live integration test that passed.

## C5 — Quality-bar re-measure + retrospective

Bar (milestone.md §Objective): module-boundaries/DDD dimension reaches ≥ A− by eliminating the H-G
class at the **class** level — `0 reach-without-a-port` + `0 hardcoded-domain-state`. Re-measured by the
validator with grep (DSN-independent) + live tests.

| Bar / class | Before | After (validator re-measured) | Root-cause-fixed evidence |
|-------------|--------|-------------------------------|---------------------------|
| H-G reach (`iam_users.display_name` outside `iam/`) | 4 reaches (1 auth + 3 security) after corrected census | **0** | `grep -rn "iam_users" --include=*.go internal/ apps/ \| grep -i display_name \| grep -v /iam/` → only comments + test seeds + e2e_seed; **zero production display-name reads**. auth `sessions_admin` query is `auth_sessions`-only; security's 3 methods scope via `TenantUserReader`/`auth_sessions.tenant_id` + port enrichment. Reads stay **live** (no snapshot). |
| H-G reach (`templates_*` under CD) | CD `PostgresTemplateVersionChecker` reached `templates_template(_version)` | **0** | `grep -rn templates_template controlleddocuments/` → exit 1 (0 matches); checker deleted (0 refs). |
| H-G hardcoded-domain-state | `status := "published"` in wiring | **0** | `grep -rn 'status := "published"' wiring/ internal/` → exit 1 (0 matches); adapter reads real status via the port. |
| Remaining cross-module `iam_users` read | — | **1, accurately characterized** | `security.MfaCoverage` (`repository.go:67,80`) — `COUNT(*) FILTER`/by-role aggregate over `mfa_enabled`/roles, **no** display_name (validator read lines 63–117). Genuine bounded defer with written trigger (M5 re-audit / next structural touch; owner backend). |
| ADRs present for all ports | — | ✅ | 0029 (display-name), 0030 (template-version), 0031 (tenant-membership) — all `Accepted 2026-06-15`, indexed, cross-linked. Plus 0032 (M4b legacy-cluster drop = the root-cause fix). |
| Dimension re-measured ≥ A− | C → | the *class* is gone via owning-module ports (root cause), not instance patches | grep + live tests confirm class-level closure. Production wiring injects **real** adapters (not Noop): `apps/api/cmd/metaldocs-api/main.go:250/260/261/410` + `apps/jobs/cmd/metaldocs-jobs/main.go:38` use `iampg.NewUserDisplayNameRepository`/`NewTenantUserRepository`. |

- **Root cause vs symptom:** PASS — and this is the crux of the re-run. The original FAIL was an
  *environment-coupled* test, and the program correctly **rejected the harness-patch** (F4.1b) and
  instead fixed the root cause at the schema layer (M4b migration 0240 dropped the dead
  `metaldocs.documents` shadow + `template_audit_log`). Proof it is root-cause-not-symptom: F4.1a Gate #5
  now passes under the operator DSN with `db.go` **unmodified** (C2). The H-G class itself is closed by
  *owning-module ports*, reads stay live (D4/Approach-3), H-PRE-1 preserved (off-tx).
- **Could it be built better?** (a) Factor the `missing→user_id` fallback into one shared helper
  (re-implemented in `security` + iam sessions handler) — next-milestone/defer input, not unsound.
  (b) The original `ALTER DATABASE search_path` harness strategy was fragile; M4c's unified test-fixture
  framework + CI grep-guards is the durable replacement (M4c PASS). No remaining unsound construction.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — **clean**: each
      H-G port has its own per-method live gate, all re-run by the validator from clean state (C2).
- [ ] Fixture/mock passed off as real-provider proof — **clean**: every C2 gate labeled real (live PG);
      fixture-vs-real distinguished honestly.
- [ ] Consumer contract guessed rather than read from the consumer — **clean**: F4.5 contract read from
      F4.6's queries; F4.6 leaves the consumer interface untouched; F4.2 preserves CD's interface;
      F4.4 removes the field auth doesn't own.
- [ ] Split-brain (one fact, two sources of truth) — **clean**: single owning `iam_users` display-name
      adapter; grep-verified zero cross-module display-name reads.
- [ ] Self-judged close / validator edited or fixed code — **clean**: the validator wrote only this
      verdict file (`milestone-qa-rerun-2026-06-15.md`); it did not edit source, run migrations, or flip
      status. The root-cause fix (M4b 0240) was made by the implementation session in a prior milestone,
      not by this validator.
- [ ] Scope drift — **clean**: F4.4/F4.5/F4.6 added under recorded operator Option-2 + the HS-6 trail;
      the schema fix delivered as its own milestone (M4b) under operator authorization. No unplanned work.
- [ ] Symptom-patch (bar "moved" by masking, root cause intact) — **CLEARED** (was the original FAIL's
      C6 hit). The environment-coupled "green" is gone: F4.1a Gate #5 now passes under the operator DSN
      (`search_path=metaldocs,public`) with `db.go` **unmodified**, because the dead `metaldocs.documents`
      shadow was dropped at the schema layer (M4b 0240) — the harness-patch (F4.1b) was explicitly
      rejected as symptom-patching. Root cause is fixed, not masked.

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- **What this re-run proves:** the single check that FAILED the original verdict — **C2 / F4.1a Gate #5**
  and its paired **C6 symptom-patch hit** — is now cleared at the root cause. The gate passes under the
  operator DSN with `search_path=metaldocs,public` (the exact form that failed) with **no harness
  change** (`db.go` unmodified at HEAD), because M4b migration 0240 dropped the dead `metaldocs.documents`
  cluster that shadowed the real `public.documents`. This is a schema-level fix, not a symptom-patch.
- **Both dimensions pass.** Code-wise: owning-module ports, dead code deleted, consumer contracts
  honored, no split-brain, reads live + off-tx (H-PRE-1). Function-wise: every M4 per-feature live gate
  green from clean state under the operator DSN; whole-repo unit green; the H-G class is at
  **class-level zero** (grep-proven: 0 `iam_users.display_name` outside `iam/`, 0 `templates_*` under
  CD, 0 `status := "published"` in wiring); ADRs present; production wiring injects real adapters.
- **Bounded defer (accurate, not a gap):** security's `MfaCoverage` aggregate `iam_users` JOIN (no
  display_name) — trigger: M5 re-audit / next structural touch; owner: backend.
- **C4 noted SKIP:** M1 full-HTTP `seed→finalize→signoff` E2E not re-stood-up here (recorded as SKIP,
  not a pass); risk surface covered by the live off-tx signoff display-name test and unchanged by M4.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — present this PASS to the operator.
> - Status flipped in `README.md`: not by the validator — the main session flips M4 → `passed` only on
>   this PASS, then proceeds to the M5 open gate per the program plan.

VERDICT: PASS
