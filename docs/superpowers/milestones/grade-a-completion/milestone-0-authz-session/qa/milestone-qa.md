# Milestone 0 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-15  ·  **Verdict:** see C7 — **PASS**.
> Run only after every feature is closed (each has a complete `evidence.md`). The validator judges and
> writes this file; the **main session flips status only on a PASS**. The validator never edits code,
> fixes findings, or flips status.

**Inputs loaded:**
- Milestone spec — `../milestone.md` (M0 — Auth / authz / session correctness, 4 features F0.1–F0.4)
- Per-feature artifacts — `spec.md` / `plan.md` / `evidence.md` for `f0.1-authz-effective-from`,
  `f0.2-manual-code-create-identity`, `f0.3-tenant-grade-view`, `f0.4-changepassword-cookie`
- Program README — `../../README.md` (M0 status `in-progress`, no prior milestone)
- Governing mission — `../../mission.md` (§5 B1–B4, §7 M0, D3 sequencing locked)
- Aggregate diff — `git diff 8e017293..HEAD` (commits `6d58e61a`, `07fc5b97`, `3a092d76`, `9413361b`);
  4 production files touched (`iam/authz/authz.go`, `controlleddocuments/application/service.go`,
  `documents/approval/application/read_service.go`, `auth/delivery/http/handler.go`) + 4 new
  integration/unit test files + 4 sqlmock-harness expectation alignments.
- Working tree clean (`git status` empty) prior to test re-runs.

## C1 — Spec & plan conformance (per feature)

Each feature's evidence acceptance matches its `spec.md` Validation Gate; the **consumer contract was
honored** (producer matches consumer, not reverse); non-goals respected.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F0.1 authz-effective-from | yes — `Require` predicate conformed to canonical `ResolveEligibleActors` (`postgres_approval_repository.go:1140`), one-line `AND upa.effective_from <= now()` added at the shared layer (`iam/authz/authz.go:123`); no per-caller patch | yes — `TestRequire_FutureDatedMembershipDenied` + `TestRequire_CurrentMembershipGranted` (real-DB integration) | yes — `effective_to` clause untouched (Q2 explicit non-goal); no system_admin bypass / cap-cache / asserted-caps change; no schema | `f0.1.../evidence.md` acceptance map all `yes` |
| F0.2 manual-code-create-identity | yes — manual-code CD-create now byte-for-byte symmetric with auto branch: `runner.Do` → `SeedTxIdentity` → `Require(CapControlledDocumentCreate, ProcessAreaCode)` → `CreateTx`; ADR 0022 Phase 7 area-scoped tier-2 enforced; "DELIBERATELY-PRESERVED asymmetry" comment removed | yes — three real-DB tests cover non-admin-with-cap success, non-admin-without-cap denial, system_admin bypass | yes — auto branch untouched; OFF-tx preflight (`CodeExists`, `ensureTemplateArtifact`, `GetTemplateVersionState`, `Resolve`) stays OFF-tx (H-PRE-1 honored, verified by reading service.go:178-267); no repo-port signature change; no schema | `f0.2.../evidence.md` acceptance map all `yes` |
| F0.3 tenant-grade-view | yes — both `LoadInstance` and `LoadActiveInstanceByDocument` now pass `"tenant"` sentinel to `authz.Require(CapDocumentView, …)` byte-for-byte aligned with canonical `documents/application/view_service.go:71`; declared scope `ScopeTenant` (`capability_scope.go:51`) honored at both call sites (root-cause family fix, not symptom-patch at the single cited line) | yes — six real-DB integration tests: granted-cross-area × 2 methods, denied-with-`AreaCode:"tenant"` × 2 methods, system_admin bypass × 2 methods | yes — `authz.Require` body untouched; `capability_scope.go` untouched; area-grade approval sites (`decision_service.go`, etc.) untouched; `loadInstanceAreaCode` helper unchanged; no schema | `f0.3.../evidence.md` acceptance map all `yes` |
| F0.4 changepassword-cookie | yes — `handleChangePassword` mirrors `handleLogout`: single added line `http.SetCookie(w, h.service.ExpiredSessionCookie())` on success; service-side revocation (CWE-613) untouched | yes — `TestPasswordChangeEmitsExpiredCookie` asserts MaxAge<0 + empty value; existing `TestPasswordChangeRevokesSessionAndClearsMustChangePassword` regression intact | yes — no `auth.Service` change; only success path emits the cookie; no new audit event; no FE change | `f0.4.../evidence.md` acceptance map all `yes` |

**C1 verdict:** PASS — every feature's evidence acceptance row maps to its `spec.md` Validation Gate;
all four consumer contracts honored (producer matches consumer); non-goals explicitly respected and
auditable in the diff.

## C2 — Gates re-run, isolated

Each feature's named tests + proof commands **re-run by the validator from clean working tree**
(`git status` empty; `git log` confirms HEAD `9413361b`) — not trusted from the evidence transcripts.
Integration tests executed against the locally-running Postgres 18 on `127.0.0.1:5433` using
`METALDOCS_DATABASE_URL=postgres://metaldocs_app:…@127.0.0.1:5433/metaldocs?sslmode=disable`.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| Build | `go build ./...` | clean (no output) | yes |
| F0.1 unit + sqlmock | `go test ./internal/modules/iam/authz/ ./internal/modules/iam/infrastructure/postgres/ ./internal/modules/auth/application/ -count=1` | `ok metaldocs/internal/modules/iam/authz 2.463s`; `ok metaldocs/internal/modules/iam/infrastructure/postgres 3.712s`; `ok metaldocs/internal/modules/auth/application 13.718s` | yes |
| F0.1 integration | `go test -tags integration -run 'TestRequire_FutureDatedMembershipDenied\|TestRequire_CurrentMembershipGranted' ./internal/modules/iam/authz/ -count=1 -v` | `--- PASS: TestRequire_FutureDatedMembershipDenied (1.29s)`; `--- PASS: TestRequire_CurrentMembershipGranted (0.12s)`; `ok metaldocs/internal/modules/iam/authz 1.783s` | yes |
| F0.2 unit | `go test ./internal/modules/controlleddocuments/... -count=1` | application/delivery/http/domain/infrastructure all `ok` | yes |
| F0.2 integration | `go test -tags integration -run 'TestCreate_ManualCode_(NonAdmin\|SystemAdmin)' ./internal/modules/controlleddocuments/application/ -count=1 -v` | `--- PASS: TestCreate_ManualCode_NonAdminWithCap_Succeeds (1.30s)`; `--- PASS: TestCreate_ManualCode_NonAdminWithoutCap_Denied (0.11s)`; `--- PASS: TestCreate_ManualCode_SystemAdmin_Succeeds (0.13s)`; `ok metaldocs/internal/modules/controlleddocuments/application 1.749s` | yes |
| F0.3 unit | `go test ./internal/modules/documents/approval/application/ -count=1` | `ok metaldocs/internal/modules/documents/approval/application 3.842s` (includes retargeted `TestLoadActiveInstanceByDocument_RequiresDocumentViewBeforeRepoLoad` with `denied.AreaCode == "tenant"` assertion) | yes |
| F0.3 integration | `go test -tags integration -run 'TestLoad(Instance\|ActiveInstanceByDocument)_(TenantGradeViewer_DocWithAreaCode_Granted\|NoViewGrant_Denied\|SystemAdmin_Granted)' ./internal/modules/documents/approval/application/ -count=1 -v` | All 6 tests `--- PASS`; `ok metaldocs/internal/modules/documents/approval/application 2.352s` | yes |
| F0.4 unit | `go test ./tests/unit -run TestPasswordChange -count=1 -v` | `--- PASS: TestPasswordChangeRevokesSessionAndClearsMustChangePassword (1.12s)`; `--- PASS: TestPasswordChangeEmitsExpiredCookie (0.75s)`; `ok metaldocs/tests/unit 2.154s` | yes |

**C2 verdict:** PASS — every named test re-runs from a clean tree and is green; integration proofs
exercised real Postgres (not sqlmock) for F0.1/F0.2/F0.3 — the only places fixture proof would be
insufficient (`Require` SQL predicate, area-scoped tx authz path, tenant-grade sentinel).

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff (`git diff 8e017293..HEAD`) reviewed as one unit. 1,415 insertions, 34
deletions across 26 files; only **4 production-Go files** changed
(`iam/authz/authz.go` +1 line; `controlleddocuments/application/service.go` net +16/-5;
`documents/approval/application/read_service.go` net +12/-20; `auth/delivery/http/handler.go` +3 lines).

**Aggregate findings (looking for cross-feature class issues):**

- **Duplicated logic across features?** No. Each feature touches a disjoint call site at the
  declared `file:line`.
- **Contract defined two ways / split-brain?** No. The sqlmock-harness expectation mirrors of the
  `Require` SQL at `auth/application/service_test.go:980` and
  `iam/infrastructure/postgres/role_admin_repository_test.go:73,159` are an **existing harness cost**
  of the project's sqlmock pattern, not a new split-brain introduced by M0; F0.1 evidence calls this
  out explicitly. The canonical SQL still lives at one site (`iam/authz/authz.go:115-127`); the test
  mirrors track it for fixture coverage of unrelated callers.
- **Dead code left by superseded approach?** F0.2 leaves the non-tx `Repository.Create` method on
  the `ControlledDocumentRepository` port; it is now unused by `service.go` but the port signature
  removal is an out-of-spec change (non-goal: "**Not** changing `s.docs.Create` or `s.docs.CreateTx`
  repo signatures or bodies"). F0.2 evidence records this as a bounded defer with a written trigger
  ("future repo-port cleanup feature or M1 contract-tightening"). Not a hidden defer.
- **One feature breaking another?** No. F0.1 (predicate tightening) ↔ F0.3 (tenant-sentinel routing):
  the tenant sentinel path in `authz.go:126` (`$2 = 'tenant'`) is unaffected by the new
  `effective_from` predicate — the sentinel only short-circuits area filtering, not the membership
  validity check; the F0.3 integration tests for `TenantGradeViewer_DocWithAreaCode_Granted` PASS,
  proving the tenant grant still lands. F0.1 ↔ F0.2: F0.2 now goes through the tightened predicate;
  F0.2 integration tests use a current-dated membership and succeed, confirming the predicates
  compose. F0.4 is hermetic to authz.
- **H-PRE-1 hazard?** Honored. F0.2's new tx contains only `SeedTxIdentity` + `Require` + `CreateTx`;
  all preflight reads (`CodeExists`, `GetTemplateVersionState`, `ensureTemplateArtifact`, `Resolve`)
  remain off-tx, verified by reading `service.go:178-267`. F0.3 strictly **removes** SQL from the
  in-tx prologue of `LoadActiveInstanceByDocument` — nothing added.
- **Symmetry with canonical sibling?** F0.2 manual branch is byte-for-byte the auto branch shape;
  F0.3 byte-for-byte the `view_service.go:71` idiom; F0.4 byte-for-byte the `handleLogout:115`
  pattern. Three of four fixes anchor on an in-repo canonical sibling — exactly the
  contract-first / shared-predicate discipline mission §3 D3 prescribes.
- **Helper / import discipline:** F0.3 removed the now-unused `docapp` import — clean. The retained
  `loadInstanceAreaCode` helper still serves its `found` existence-probe contract; bounded defer for
  the cosmetic rename is recorded.

**Findings:** none staff-engineer-blocking. The three bounded defers (recorded in evidence files):
non-tx `Repository.Create` removal, `loadInstanceAreaCode` cosmetic rename, F0.1 `effective_to`
half — all carry written triggers and owners.

- Staff-engineer bar met? **yes**

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (`wiki/quality/backend-api-qa-checklist.md` with authz-correctness lens) | pass | Deny-by-default preserved (`Require` still returns `ErrCapDenied` on any failed predicate; deny envelope correctly carries `AreaCode:"tenant"` for tenant-grade caps post-F0.3). No privilege widening: F0.1 strictly **denies** future-dated grants (security tightening); F0.2 enforces tier-2 area-scoped on a path that previously bypassed (closes B2 silently-bypassing-admin path); F0.3 aligns to declared `ScopeTenant` (does **not** widen — it correctly matches the in-repo canonical idiom); F0.4 forces client-side cookie expiry on top of already-correct server revocation. Symptom-patch lens: F0.1 fixed at the shared `authz.Require` predicate, not per-caller (ADR 0022 / authz-root-cause memory). F0.3 fixed at both sibling call sites in the family (root-cause discipline), not only the cited line. |
| Regression vs prior milestones | n/a — M0 is the first milestone in this program | Existing authz/session test corpus stayed green: full `go test ./... -count=1` returned zero `FAIL` lines across 100+ packages (iam/authz, iam/infrastructure/postgres, auth/application, controlleddocuments/{application,delivery/http,domain,infrastructure}, documents/approval/{application,domain,http,repository,…}, tests/unit, tests/unit/iam_memberships, tests/unit/iam_people, all `ok`). |

**C4 verdict:** PASS.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Mission §5 B1 — `authz.Require` ignores `effective_from` (premature-access security bug) | Future-dated `user_process_areas` rows granted access today | Denied with `ErrCapDenied` at the shared predicate | `iam/authz/authz.go:123` adds `AND upa.effective_from <= now()` **once** at the shared predicate (not per-caller); milestone bar criterion ("F0.1 lands at the shared authz predicate, not per-caller / ADR 0022") satisfied; `TestRequire_FutureDatedMembershipDenied` re-run PASS on real Postgres |
| Mission §5 B2 — manual-code CD-create bypassed PEP/PDP | Non-system-admin caller failed-closed on missing actor_id GUC (`MustActorID` NULL scan error) | Non-admin-with-cap succeeds; non-admin-without-cap → `ErrCapDenied`; system_admin still bypasses | `controlleddocuments/application/service.go:269-283` wraps persist in `runner.Do` with `SeedTxIdentity` + `Require(CapControlledDocumentCreate, ProcessAreaCode)` + `CreateTx`, byte-for-byte symmetric with auto branch (the canonical sibling). Fixed at the named site without widening any grant — the new authz call is the ADR 0022 Phase 7 area-scoped tier-2 the auto branch already enforces |
| Mission §5 B3 — `CapDocumentView` narrowed to area-grade in approval reads | Tenant-role-only viewer denied for documents carrying a real area code, on both `LoadInstance` and `LoadActiveInstanceByDocument` | Tenant-grade viewer granted across areas on both methods; deny envelope correctly `AreaCode:"tenant"` | `documents/approval/application/read_service.go:64-67, 99-102` aligns both call sites with declared `ScopeTenant` (`iam/domain/capability_scope.go:51`) — byte-for-byte the canonical `view_service.go:71` idiom. Fixed at both sibling sites in the family (root-cause discipline), not only the cited line — operator-approved per F0.3 spec interview Q1 |
| Mission §5 B4 — self-service ChangePassword leaves dead session cookie client-side | Browser kept revoked cookie until natural expiry; 401 on next request | `Set-Cookie metaldocs_session=; Max-Age=-1` on success; mirrors `handleLogout` | `auth/delivery/http/handler.go:158-160` adds the single mirror line at the named site. Server-side revocation (CWE-613 fix at `service.go:494`) untouched — F0.4 closes the client-side half of the contract |

**Retrospective:** F0.3 evidence records the candidate program-wide improvement — adding a
`Require`-side static assert that `ScopeOf(cap) == ScopeTenant ⇒ areaCode == "tenant"` would catch
this class of bug at the shared layer. Recorded as a bounded defer with a written trigger (M1 or a
dedicated authz-hardening micro-milestone). Does **not** by itself FAIL M0 (current construction is
sound; both runtime call sites now obey the declared scope). F0.1 has a parallel observation about
`effective_to` half-alignment, also deliberately deferred with rationale (changing it would widen
access — security-direction change requiring its own decision, not an F5.1 finding).

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — clean; C1
      maps each feature's evidence row to its `spec.md` Validation Gate explicitly.
- [ ] Fixture/mock passed off as real-provider proof — clean; F0.1/F0.2/F0.3 acceptances each carry a
      real-Postgres `//go:build integration` test (sqlmock unit tests are kept for regression but
      explicitly marked "fixture" in evidence). F0.4 is a handler test — appropriate for the HTTP
      cookie-emission contract, no provider needed.
- [ ] Consumer contract guessed rather than read from the consumer — clean; F0.1 reads the canonical
      `ResolveEligibleActors` predicate; F0.2 reads the auto-code branch; F0.3 reads
      `capability_scope.go:51` + `view_service.go:71`; F0.4 reads `handleLogout:115`.
- [ ] Split-brain (one fact, two sources of truth) — clean; the sqlmock-harness predicate mirrors are
      a pre-existing testing-pattern cost, not a new fact-level split-brain.
- [ ] Self-judged close / validator edited or fixed code — clean; this verdict is written by the
      validator subagent in a fresh thread; no source / spec / evidence edits made; no status flipped.
- [ ] Scope drift (work beyond the spec, no rationale) — clean; F0.3 covered both sibling sites (the
      one cited by mission §5 B3 *and* the identical-bug sibling at `read_service.go:114`) — the
      operator pre-approved the family-fix scope in F0.3 spec interview Q1 with explicit rationale
      ("root-cause family fix, matches F0.1 'shared predicate, not per-caller' discipline").
- [ ] Symptom-patch (bar "moved" by masking, root cause intact) — clean; F0.1 at the shared
      predicate; F0.2 at the named bypass site closing it instead of patching downstream; F0.3 at the
      declared-scope property (both sites in the family); F0.4 at the literal missing client-side
      cookie clear — root causes addressed, not symptoms.

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- All four features deliver the milestone spec's "what to validate"; the milestone's stated bar is
  measurably moved (B1 — `Require` now denies future-dated memberships at the shared predicate per
  ADR 0022; B2/B3/B4 fixed at their named sites without widening any grant); whole-repo `go test
  ./...` is green; named integration tests re-run from a clean tree are all PASS; the existing
  authz/session test corpus stays green.
- Bounded defers carried forward (each with a written trigger in evidence): F0.1 `effective_to`
  half-alignment; F0.2 non-tx `Repository.Create` port-method removal; F0.3 `Require`-side
  `ScopeTenant` assert + `loadInstanceAreaCode` cosmetic rename; F0.2-flagged pre-existing
  `TestSequenceAllocatorNextAndIncrement_Concurrent` flake (reproduced on pre-M0 HEAD, unrelated to
  M0 scope, did not trigger in the validator's runs) — all out of M0 scope by design.
- On **PASS** — handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending operator approval
> - Status flipped in `README.md`: pending (only on PASS — this verdict is PASS)
