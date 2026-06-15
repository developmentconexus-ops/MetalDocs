# Milestone 4 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-15  ·  **HEAD:** `30503533` (F4.6)  ·  **Verdict:** see C7.
> The validator judges and writes this file; the **main session flips status only on a PASS**. The
> validator never edits code, fixes findings, or flips status.

## Scope note

M4 thesis: close the **H-G class** = cross-module *reach-without-a-port* (a consumer module issuing
raw SQL against another module's owned table — `metaldocs.iam_users`) and *hardcoded-domain-state*.
Bar = **class-level zero**, proven by grep + live build/test, not one instance patched. Census was
corrected twice (validator-FAIL + HS-6 undercount → operator Option-2 full close); features executed:
F4.1, F4.1a (testdb rehab), F4.2, F4.3 (ADRs), F4.4 (auth), F4.5 (iam membership port), F4.6
(security). All M4 source is committed at HEAD; the only uncommitted paths are this `qa/` folder and
unrelated `.claude/skills/**` deletions (not M4 work).

---

## C1 — Spec & plan conformance (per feature)

Each feature's evidence acceptance matches its `spec.md` Validation Gate; consumer contract honored
(producer matches consumer, read from the consumers — not invented); non-goals respected.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F4.1 user-display-name-reader | ✅ port shape read from the 3 consumers (approval signoff, get-instance batch, documents create); iam-owned, single+batch | ✅ (gate #5 see C2 caveat) | ✅ no snapshot/freeze; no OpenAPI/route change; off-tx preserved | `f4.1-*/evidence.md` |
| F4.1a documents-testdb-rehab | ✅ consumer = documents integration tests; test-harness-only | ⚠️ see C2 — env-coupled | ✅ no production/schema/migration change | `f4.1a-*/evidence.md` |
| F4.2 template-version-state-reader | ✅ extends existing templates-owned `TemplateVersionPort`; CD `TemplateVersionChecker` `(status, doc_type_code)` shape preserved | ✅ | ✅ `IsPublished` untouched; no OpenAPI/route; no snapshot | `f4.2-*/evidence.md` |
| F4.3 port-adrs | ✅ docs-only | ✅ ADR 0029/0030 (+0031 via F4.5) present, headers, indexed, cross-linked | ✅ no code | `f4.3-*/evidence.md` |
| F4.4 auth-session-display-name-port | ✅ auth narrowed to auth-owned rows; iam consumer enriches via port; `DisplayName` removed from `SessionListItem` | ✅ | ✅ no OpenAPI/route; display_name key+value preserved consumer-side | `f4.4-*/evidence.md` |
| F4.5 iam-tenant-membership-port | ✅ contract read from F4.6's 3 coupled queries (membership id-set, no `deactivated_at` filter) | ✅ | ✅ producer-only; no consumer wiring (F4.6's job) | `f4.5-*/evidence.md` |
| F4.6 security-display-name-port | ✅ `securitydomain.Repository` + structs + Service + handler + OpenAPI unchanged (service_test mock = regression guard); rows/names byte-identical | ✅ | ✅ `MfaCoverage`/`ListOffHoursAdminActions` deferred accurately | `f4.6-*/evidence.md` |

**C1 result: PASS** — every feature's acceptance maps to its spec gate; consumer contracts were read
from the consumers (verified: F4.6 leaves the `securitydomain.Repository` interface and domain structs
untouched, F4.2 leaves CD's `TemplateVersionChecker` signature untouched, F4.4 removed `DisplayName`
from the auth domain struct so auth no longer claims ownership). Non-goals respected. The one caveat
(F4.1a) is a C2 re-run issue, not a C1 conformance gap.

## C2 — Gates re-run, isolated

Re-run by the validator from a clean state (not trusted from the transcript). DB:
dev Postgres `127.0.0.1:5433/metaldocs`, `-tags integration -count=1`.

| Feature / gate | Command re-run (validator) | Real output | Pass? |
|----------------|-----------------------------|-------------|-------|
| Build | `go build ./...` | exit 0 | ✅ |
| Vet (plain) | `go vet ./...` | exit 0 | ✅ |
| Vet (integration) | `go vet -tags integration ./internal/modules/security/... ./internal/modules/iam/... ./internal/modules/auth/...` | exit 0 | ✅ |
| Unit (3 modules) | `go test ./internal/modules/security/... ./internal/modules/iam/... ./internal/modules/auth/...` | all `ok` | ✅ |
| Unit (whole repo) | `go test ./...` | 87 `ok`, 0 FAIL | ✅ |
| F4.1 iam port | `go test -tags integration -run 'TestUserDisplayNameRepository_DisplayName(s)?_Live' ./internal/modules/iam/infrastructure/postgres/` | `--- PASS` (present/absent/tenant-scoped + batch); `ok 3.326s` | ✅ **real (live PG)** |
| F4.1 approval off-tx (H-PRE-1) | `go test -tags integration -run TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema ./internal/modules/documents/approval/repository/` | `--- PASS 1.03s`; log `AC5 … = "Alice Approver"`, `AC3 empty-on-missing = "" (nil err)` | ✅ **real (live PG)** |
| F4.1a Gate #5 snapshot value | `go test -tags integration -p 2 -run TestCreateDocumentTx_PopulatesAllSnapshotColumns ./internal/modules/documents/application/` — **with operator DSN `…&search_path=metaldocs,public`** | `--- FAIL 137.21s`: `column "tenant_id" of relation "documents" does not exist (SQLSTATE 42703)` | ❌ **env-coupled FAIL** |
| F4.1a Gate #5 snapshot value | same test — **with a DSN lacking the `search_path` param** | `--- PASS 179.87s` (`== "Snapshot Author"`) | ✅ (only this DSN form) |
| F4.1a repository_create suite | `go test -tags integration -p 2 -run 'StorageKeyInvariant|RevisionNumber|RejectsEmptyName|GetDocument_ReturnsSnapshotMetadata' ./internal/modules/documents/repository/` — DSN without `search_path` | `ok 182.995s` | ✅ (only DSN without search_path) |
| F4.2 templates port | `go test -tags integration -run TestTemplateVersionReader_GetTemplateVersionState_Live ./internal/modules/templates/infrastructure/` | `--- PASS` 4/4 (published/obsolete/absent/other-tenant); `ok 198.92s` | ✅ **real (live PG)** |
| F4.4 sessions no-join | `go test -tags integration -run TestListActiveSessions_NoIamUsersJoin ./internal/modules/auth/infrastructure/postgres/` — note: reads `TEST_DATABASE_URL`, not `METALDOCS_DATABASE_URL` | `--- PASS 0.84s` | ✅ **real (live PG)** |
| F4.5 tenant-membership | `go test -tags integration -run TestTenantUserRepository_TenantUserIDs_Live ./internal/modules/iam/infrastructure/postgres/` | `--- PASS` (all members incl. deactivated; other-tenant excluded; unknown→empty); `ok 1.048s` | ✅ **real (live PG)** |
| F4.6 security no-iam-join | `go test -tags integration -run TestSecurityRepository_NoIamUsersJoin_Live ./internal/modules/security/infrastructure/postgres/` | `--- PASS` all 4 subtests (ListLockouts/CountRecentFailedLoginsByUser/CountRecentLockouts/ListNewDeviceLogins); `ok 0.896s` | ✅ **real (live PG)** |

**C2 result: FAIL (single gate).** Every H-G port's own live gate (F4.1, F4.2, F4.4, F4.5, F4.6)
re-ran green from clean state under the operator-provided DSN. **F4.1a Gate #5 fails on isolated
re-run with the operator-provided DSN** (`search_path=metaldocs,public`) and only passes with a DSN
that has **no** `search_path` param. Root cause (validator-diagnosed): the F4.1a fix sets the test
DB's default search_path via `ALTER DATABASE … SET search_path TO public, metaldocs`, but
`tests/integration/testdb/db.go:openDBWithDatabase` runs `pgx.ParseConfig(dsn)`, which lifts the
DSN's `search_path` param into a **per-connection** `options` setting that **overrides** the database
default. With `metaldocs` first, bare `documents` resolves to the dead legacy `metaldocs.documents`
(no `tenant_id`/snapshot columns) → SQLSTATE 42703. The fix is therefore **environment-coupled**: its
green result is contingent on the absence of a `search_path` connection param, which the F4.1a
evidence reports as an unconditional `ok … 161.952s`. Per the binding C2 rule, "flaky or
environment-coupled is not green."

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff (`2e7e2009..HEAD`, 9 commits) reviewed as one unit.

- **Single owning adapter (no split-brain):** verified by grep that *all* `iam_users` SQL outside
  `iam/` is gone — every read now lives inside the `iam/` module (`user_display_name_repository.go`,
  `tenant_user_repository.go`, observability, role_provider, presence — all intra-module). The only
  cross-module `iam_users` reference is `security/…/repository.go:67` (`MfaCoverage`, an aggregate
  over `mfa_enabled`/roles, **no** display_name) and `internal/test/e2e_seed.go` (a seed DELETE, not
  a read). There is exactly **one** display-name adapter — no second source of truth.
- **No dead code:** F4.2 *deleted* the superseded `PostgresTemplateVersionChecker` (grep → 0
  references); F4.4 *removed* `DisplayName` from auth's `SessionListItem`. No orphaned superseded
  approach left behind.
- **No feature broke another:** whole-repo `go test ./...` = 87 ok / 0 fail; the three touched
  modules' integration gates all green.
- **Minor repetition (retrospective, not a finding):** the `missing→user_id` presentation fallback is
  re-implemented per consumer (`security.resolveNames`, `iam handler.resolveDisplayNames`). This is
  the deliberate ADR-0029 design (port returns raw display_name omitting empty; each consumer applies
  its own presentation fallback) — not a duplicated *fact*, but the idiom could be a shared helper.
  Recorded as a C5 retrospective input.

- Findings: none blocking.
- Staff-engineer bar met? ✅ (the production diff is senior-grade; the failing item is a test-harness
  robustness gap in F4.1a, see C2.)

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| backend-api-qa-checklist | pass | Internal module-port swaps; no route/OpenAPI/generated-surface/authz change (verified: F4.6 leaves OpenAPI + `securitydomain.Repository` untouched; F4.2 leaves CD interface untouched; F4.4 no route change). Shared consumers preserved + green. |
| workflow-async-qa-checklist (F4.2 CD-create lock-bearing) | pass | F4.2 status read on the pool conn, non-authz, call site unmoved → H-PRE-1 not in play, not regressed (structural proof in F4.2 evidence; corroborated by F4.1a's deadlock diagnosis: the off-tx read is what *requires* a multi-conn pool). |
| Regression — whole-repo unit | all pass | `go test ./...` → 87 ok, 0 FAIL |
| Regression — M1 full-HTTP `seed→finalize→signoff` E2E | **SKIPPED** (not re-proven here) | `TestE2E_HappyPath_HTTP` requires a running server (`METALDOCS_E2E_URL`), which the validator did not stand up. It is a SKIP, **not** a real-provider PASS. Mitigation: the signoff display-name code path F4.1 touches is independently re-proven by `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema` (live PG, PASS). The full-HTTP E2E was discharged as the M1 HS-1 condition on 2026-06-14. |
| Regression — prior-milestone gates (M0–M3) | no break observed | Build/vet/unit clean tree-wide; no prior gate failed under M4's changes. |

**C4 result: PASS-with-noted-SKIP.** No prior milestone regressed. The M1 full-HTTP E2E is recorded
as a SKIP (server not stood up), explicitly not counted as a pass; its risk surface (signoff
display-name) is covered by a live unit-of-integration test.

## C5 — Quality-bar re-measure + retrospective

Bar (milestone.md §Objective): module-boundaries/DDD dimension reaches ≥ A− by eliminating the H-G
class at the **class** level — `0 reach-without-a-port` + `0 hardcoded-domain-state`.

| Bar / class | Before | After (validator re-measured) | Root-cause-fixed evidence |
|-------------|--------|-------------------------------|---------------------------|
| H-G reach-without-a-port (`iam_users.display_name` outside `iam/`) | 4 reaches (1 auth + 3 security) after the corrected census | **0** | grep (DSN-independent): zero `SELECT … display_name … FROM metaldocs.iam_users` outside `iam/`; auth `sessions_admin` query = `auth_sessions` only; security 4 methods scope via `TenantUserReader`/`auth_sessions.tenant_id` + port enrichment. Reads stay **live** (no snapshot). |
| H-G reach (`templates_*` under CD) | CD `PostgresTemplateVersionChecker` reached `templates_template(_version)` | **0** | grep: 0 `templates_template` SQL under `controlleddocuments/`; checker deleted (0 refs). |
| H-G hardcoded-domain-state | `status := "published"` in wiring | **0** | grep: 0 `status := "published"` in `apps/api/internal/wiring/`; adapter reads real status via the port. |
| Remaining cross-module `iam_users` read | — | **1, accurately characterized** | `security.MfaCoverage` — aggregate over `iam_users.mfa_enabled` + `iam_user_roles` (validator inspected lines 63–119: COUNT/FILTER, **no** display_name). Genuine bounded defer with written trigger (M5 re-audit / next structural touch; owner backend). |
| ADRs present for both/all ports | — | ✅ | ADR 0029 (display-name), 0030 (template-version), 0031 (tenant-membership) — all `Accepted 2026-06-15`, indexed, cross-linked. |
| Dimension re-measured ≥ A− | C → | the *class* is gone via owning-module ports (root cause), not instance patches | grep + live tests confirm class-level closure. Production wiring injects **real** adapters (not Noop) in both `apps/api/cmd/metaldocs-api/main.go` and `apps/jobs/cmd/metaldocs-jobs/main.go` (validator verified lines 250/258–261/410/425/685 and jobs 38–39). |

- **Root cause vs symptom:** PASS — the class is closed by *owning-module ports*, not scattered
  per-call-site SQL tweaks. Reads stay live (D4/Approach-3); H-PRE-1 preserved (off-tx).
- **Could it be built better?** (a) Factor the `missing→user_id` fallback into one shared helper
  (currently re-implemented in `security` and the iam sessions handler) — defer/next-milestone input.
  (b) **F4.1a's `ALTER DATABASE` search_path strategy is fragile** — it is silently defeated by any
  DSN that carries a `search_path` connection param; the harness should normalize/strip the
  connection-level `search_path` (or set it per-session on every pooled connection) so the fix is
  DSN-independent. This is the direct cause of the C2 failure.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean
      (each H-G port has its own per-method live gate, validator re-ran them).
- [ ] Fixture/mock passed off as real-provider proof — clean (fixture vs real labeled honestly in
      every evidence file and re-confirmed by the validator).
- [ ] Consumer contract guessed rather than read from the consumer — clean (F4.5 contract read from
      F4.6's queries; F4.6 leaves the consumer interface untouched; F4.2 preserves CD's interface).
- [ ] Split-brain (one fact, two sources of truth) — clean (single owning `iam_users` display-name
      adapter; grep-verified).
- [ ] Self-judged close / validator edited or fixed code — clean (validator wrote only this file).
- [ ] Scope drift — clean (F4.4/F4.5/F4.6 added under recorded operator Option-2 + HS-6 trail; no
      unplanned work beyond F4.1–F4.6 + the F4.1a testdb rehab fix-feature).
- [x] **Symptom-patch / environment-coupled "green"** — **HIT.** F4.1a Gate #5 is reported in evidence
      as an unconditional `ok … 161.952s` with a real-reader value assertion, but on isolated re-run
      with the operator-provided DSN it **fails** (SQLSTATE 42703). Its green result is contingent on
      a specific DSN form (no `search_path` param). An environment-coupled result presented as green
      is a forbidden-list hit.

## C7 — Verdict

- **VERDICT: FAIL**
- **Failed checks:** **C2** (F4.1a Gate #5 environment-coupled — fails on isolated re-run with the
  operator-provided `search_path=metaldocs,public` DSN) and the corresponding **C6** hit
  (environment-coupled result presented as unconditional green).
- **What still holds (for the main session's context):** the **core M4 thesis is met** — the H-G
  display-name and template-state classes are at **class-level zero** (grep-proven, DSN-independent),
  and every H-G *port's own* live gate (F4.1, F4.2, F4.4, F4.5, F4.6) re-ran green from clean state
  under the operator DSN; build/vet/unit are clean tree-wide; ADRs present; production wiring injects
  real adapters. The failure is isolated to **one corroborating test** (the documents
  created_by-snapshot end-to-end value proof) whose harness is DSN-fragile.
- **Minimum fix feature to open:** **`f4.1b-testdb-search-path-robustness`** — make the
  documents-integration testdb harness DSN-independent so the `documents` table resolves to the real
  `public.documents` regardless of any `search_path` connection param in the DSN. Concretely: in
  `tests/integration/testdb/db.go` (`openDBWithDatabase` / `Open`), strip or override the
  connection-level `search_path` from the parsed pgx config (or set the desired search_path on every
  pooled connection via `AfterConnect`/`SET`), so the `ALTER DATABASE … SET search_path TO public,
  metaldocs` default is honored. **Re-run acceptance:** `TestCreateDocumentTx_PopulatesAllSnapshotColumns`
  (asserts `created_by_display_name_snapshot == "Snapshot Author"` under the real
  `iampg.NewUserDisplayNameRepository`) and the `repository_create` suite pass **with the
  operator-provided DSN** (`…&search_path=metaldocs,public`), and the F4.1a evidence is corrected to
  state the result is DSN-independent. Milestone stays **active**; the main session does **not**
  advance.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — not reached (verdict is FAIL).
> - Status flipped in `README.md`: no (only on PASS). M4 remains `in-progress`; open the named fix
>   feature, re-run its lifecycle, re-dispatch the validator (HS-4).
