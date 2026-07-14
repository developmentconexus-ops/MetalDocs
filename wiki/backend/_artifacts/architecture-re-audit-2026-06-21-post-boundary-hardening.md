# Backend Architecture Re-Audit Report — Post Module-Boundary Hardening

**Project:** MetalDocs
**Code HEAD:** `44b83071` (mission `backend-module-boundary-hardening` tip; M0–M4 complete, local, not merged/pushed)
**Date:** 2026-06-21
**Audit type:** Terminal acceptance re-audit (mission `backend-module-boundary-hardening`, §8 definition-of-done).
The parent `grade-a-completion` Grade-A is HELD pending this PASS.
**Method:** F5.1 10-dimension independent multi-agent audit (one sonnet reader per dimension) reading actual
source at HEAD `44b83071`, with a refute-by-default adversarial skeptic re-reading every cited `file:line`
for each proposed Critical/Major. Only skeptic-confirmed findings count toward the bar. Class counts (H-G both
readings, build, tests, cilint) measured by the main session at HEAD. Re-audit workflow `wf_6fd45429-eb4`
(10 dimension agents + skeptic stage; 0 Critical/Major proposed, so 0 required skeptic refutation rounds).

> **Verdict signal (evidence-only; the binding verdict is the independent `mission-validator`'s).** Every §8
> criterion is met at HEAD `44b83071`: **module-boundaries / DDD = A** (up from post-M9 B+; the mission
> target), **H-G = 0 under BOTH readings** (canonical §6 greps AND the broad cross-module owned-**base**-table
> sweep), **cilint H-G guard exit 0**, **0 skeptic-confirmed new Critical/Major** across all ten dimensions,
> and **no dimension below its post-M9 floor** (two recovered: code-quality B+→A−, module-boundaries B+→A).

---

## 1. Method

Each dimension was graded A–F from the actual code at HEAD `44b83071`, citing `file:line` — no grading by
assertion, no trusting M0–M4 milestone evidence as proof of the bar. Every proposed Critical/Major was wired
to an independent refute-by-default skeptic before counting; **zero** Critical/Major were proposed by any
dimension reader, so the verified count is 0/0. H-G was measured **two ways** (§6): the **canonical §6 grep
commands** (the measure §8 says to re-run, used by every prior re-audit) and the **broad "any cross-module
owned-base-table read" sweep** (the literal §8-amendment reading, reconciled by ADR-0039's
published-view/read-port exemption). Build, unit suite, integration suite (testdb-framework packages), and the
`tools/cilint` H-G analyzer were run at HEAD.

Baseline column: **post-M9** = `wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m9.md` (`d7d53590`),
the parent mission's terminal re-audit that surfaced the boundary debt this mission remediated.

**Scope fact.** The mission diff `d7d53590..44b83071` (M0–M4) touches: ADR-0039; the cilint `hgcrossmodule`
analyzer + tests; `controlleddocuments/domain/resolution.go` (M1 typed constants); owner-published read-ports
+ consumers (M2); the `metaldocs.v_active_user_areas` membership view + CD/approval consumption (M3); the CD
visibility views `v_cd_search_facts` + `v_cd_grantee` + documents `v_document_search_facts` (migrations
0243/0244) + `search/infrastructure/v2documents/reader.go` consumption (M4); plus parity tests, wiki, and
milestone docs. **No runtime-behavior change** — seam-only, parity-proven per site (ADR-0039 D6).

---

## 2. Scorecard

| # | Dimension | post-M9 (`d7d53590`) | This re-audit (`44b83071`) | Δ |
|---|-----------|:---:|:---:|:---:|
| 1 | Authz / capability model | A | **A** | = |
| 2 | Security / tenant isolation | A− | **A−** | = |
| 3 | Sessions / auth lifecycle | A | **A** | = |
| 4 | Middleware / HTTP kernel | A | **A** | = |
| 5 | Persistence / transactions | A− | **A−** | = |
| 6 | Code quality / Go idioms | B+ | **A−** | +1 |
| 7 | Legacy / dead-code | A− | **A−** | = |
| 8 | **Module boundaries / DDD** | B+ | **A** | **+1 ★ target** |
| 9 | Contract / API layer | A− | **A−** | = |
| 10 | Composition / observability | A− | **A−** | = |

**Net effect.** The two post-M9 variance dips on pre-existing code (dimensions 6 and 8, graded B+ at post-M9
because of the `resolution.go` status literals and the cross-module raw-SQL reads) are **resolved** by this
mission: M1 typed the constants (→ dim 6 A−) and M0–M4 eliminated every in-scope cross-module base-table read
(→ dim 8 A). No dimension regressed; every other dimension is byte-stable or unchanged in grade.

---

## 3. §8 Pass-Bar Verdict (evidence)

**Pass bar:** module-boundaries / DDD = A, H-G = 0 under both readings, 0 skeptic-confirmed new Critical/Major,
no other-dimension regression. **All met.**

| § Criterion | Result | Evidence |
|---|---|---|
| 1. ADR-0039 exists, Accepted, unambiguous definition + exemption list | **PASS** | `wiki/decisions/0039-cross-module-base-table-read-boundary.md` — Status **Accepted 2026-06-20** (amended for D3(d)–(f) exemptions). Raw read of another module's base table = violation; published view / read-port = compliant. |
| 2. cilint/CI H-G grep-guard green on full tree | **PASS** | `go run ./tools/cilint ./...` → **exit 0**. `hgPendingRemediation` slice **EMPTY** (`tools/cilint/internal/analyzers/hgcrossmodule.go`); `hgExempt` retains only the permanent D3(d)–(f) carve-outs. |
| 3. Every §5 inventory item (rows 3–16) ported or ADR-0039-exempt | **PASS** | M1: resolution.go literals → typed constants (0 bare `"published"`/`"obsolete"`). M2: 9 point-reads → owner read-ports. M3: `user_process_areas` reads → `v_active_user_areas`. M4: search → `v_*` views. Broad sweep (§6) finds 0 remaining cross-module base-table reads among the §5 sites. |
| 4. Re-audit grades module-boundaries = A and H-G = 0 both readings | **PASS** | Dim 8 = **A** (§2). H-G canonical = 0, broad = 0 (§6). |
| 5. `go build ./...` + `go test ./...` green; integration green where docker PG available, else not-run | **PASS (with bounded not-run)** | `go build ./...` exit 0; `go test ./...` (unit) exit 0; testdb-framework integration packages green (search v2documents, iam/infrastructure/postgres, documents/repository + approval/repository, taxonomy/infrastructure). Raw-base-DSN `_Live` tests not-run — see §7. |
| 6. 0 new Critical/Major, each skeptic-confirmed | **PASS** | 0 Critical/Major proposed by any of the 10 dimension readers; verified count 0/0. |

---

## 4. Per-dimension findings

**0 Critical and 0 Major across all ten dimensions.** Every reader returned an empty findings array; the
refute-by-default skeptic stage therefore had nothing to refute. Minor observations recorded by readers (not
counting toward the bar, no action required for this gate):

- **Dim 1 (Authz, A):** stale NOTE comment `iam/authz/capability_scope.go:31-35` claims doc.create call sites
  pass `"tenant"`; refuted against source — all `CapDocumentCreate`/`CapControlledDocumentCreate` calls pass
  real area codes (`repository.go:145`, `controlleddocuments/.../service.go:306,341,513,735`). Comment is
  misleading only; no enforcement gap.
- **Dim 2 (Security, A−):** `tenantIDFromContext` idempotency-scope helper falls back to DevTenantID; used only
  for idempotency-key namespacing (best-effort), not the mutation/authz path — no cross-tenant breach.
- **Dim 5 (Persistence, A−):** `fillin_authz.go:22` uses `runner.Do` for a read-only gate (TODO, operator
  deferred); idempotency `BeginReplay` self-retry bounded in practice; outbox is at-most-once (documented).
- **Dim 9 (Contract, A−):** search accepts 3 undeclared query params (ignored by reader); CD routes use a
  documented separate problem-code taxonomy. Pre-existing, documented.
- **Dims 3/4/6/7/8/10:** minor cosmetic/comment items only (full text in workflow `wf_6fd45429-eb4`).

---

## 6. Class Re-Measurement (verbatim at HEAD `44b83071`)

### H-G — cross-module reach (TWO readings; both 0)

**(a) Canonical §6 greps:**
```
grep -rEn 'FROM[[:space:]]+(metaldocs\.)?document_profiles' internal/ --include='*.go' \
  | grep -v _test.go | grep -v 'internal/modules/taxonomy/'        → 0 SQL reads
grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_users|iam_user_roles' internal/modules/ --include='*.go' \
  | grep -v '/iam/' | grep -v _test.go                             → 0 SQL (auth/security hits = port-decl COMMENTS)
```
**Canonical H-G = 0 — PASS.**

**(b) Broad sweep — any cross-module owned-base-table read among the §5 debt tokens
(`user_process_areas`, `document_revisions`, `approval_instances`, `controlled_documents`,
`document_process_areas`, `templates_template_version`, `controlled_document_*_grants`):**
```
Every remaining FROM/JOIN hit is either (i) same-module — the owner reading its OWN table (ADR-0039 D4c
compliant: iam→user_process_areas in iam/authz/authz.go:124; CD→controlled_documents in
controlleddocuments/**; documents→document_revisions/approval_instances in documents/**;
taxonomy→document_process_areas in taxonomy/**; templates→templates_template_version in templates/**),
or (ii) jobs worker-layer (ADR-0039 D3f exempt: jobs/stuck_instance_watchdog/job.go:147).
The former §5 cross-module sites are GONE:
  - CD infrastructure/repository.go now reads v_active_user_areas (membership) + ActiveInstanceReader port
  - search/infrastructure/v2documents/reader.go reads v_document_search_facts + v_cd_search_facts + v_cd_grantee
  - documents/iam taxonomy reads via taxonomy read-port; CD point-reads via CDFieldReader port
  - resolution.go → 0 bare status literals
```
**Broad H-G = 0 — PASS.**

### cilint H-G guard
```
GOFLAGS=-mod=mod go run ./tools/cilint ./...   → exit 0   (hgPendingRemediation EMPTY; only D3(d)–(f) exemptions)
```

### Build & tests
```
GOFLAGS=-mod=mod go build ./...   → exit 0 (clean)
GOFLAGS=-mod=mod go test ./...    → exit 0 (full unit suite green)
Integration (testdb-framework packages, METALDOCS_DATABASE_URL :5434) → green:
  search/infrastructure/v2documents, iam/infrastructure/postgres, documents/repository,
  documents/approval/repository, taxonomy/infrastructure, templates/application
```

---

## 7. Bounded not-run — raw-base-DSN `_Live` integration tests (HS-3, not a regression)

Tests that connect **directly to the raw base DSN** (the `*_Live` integration tests) error at seed setup with
`relation "metaldocs.<table>" does not exist (SQLSTATE 42P01)`. Confirmed environmental cause: the `:5434`
test container's base `metaldocs` database is intentionally **empty** (`to_regclass('metaldocs.tenants')`,
`v_active_user_areas`, `v_cd_search_facts` all return NULL) — schema materializes only inside the
testdb-framework **per-test clones**, which the framework-based packages above use and pass.

This is the same class the M4 milestone-validator **independently reproduced at the pre-M4 commit**
(`e147f33e^` worktree) and dispositioned as pre-existing — see
`docs/superpowers/milestones/backend-module-boundary-hardening/milestone-4-search-visibility-contract/qa/milestone-qa.md`
(C4). Affected tests observed this run (all setup-stage `relation does not exist`, **not** parity mismatches):
`TestSecurityRepository_PortParity_Live` (security/infrastructure/postgres, mission F4.5 proof — passes under a
migrated DB, proven green in M4 validation), plus the known-three (`TestSequenceAllocatorNextAndIncrement_Concurrent`,
`TestPostgresLimiter_Live`, `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema`). Recorded **not-run
(HS-3)**, no false-green. The actual mission parity proofs (M2 `TestActiveInstanceReader_*`/`TestCDFieldReader_*`,
M3 `TestCanRead_ViewParityWithRaw`/`TestList_ViewParityWithRaw`, M4 F4.1 `TestCDVisibilityContract_ComposedDecisionParityWithRaw`
+ F4.3 frozen-raw parity) run under testdb clones and are green.

---

## 8. Terminal Acceptance — evidence summary

All six §8 checks PASS (one with a bounded, non-false-green HS-3 not-run on raw-base-DSN `_Live` tests).
module-boundaries = **A**; H-G = **0** under both readings; **0** skeptic-confirmed Critical/Major; no
dimension below its post-M9 floor. This report is the evidence base; the **binding** terminal verdict is the
independent `mission-validator`'s, written to
`docs/superpowers/milestones/backend-module-boundary-hardening/qa/mission-validation.md`. Grade-A sign-off on
the parent `grade-a-completion` remains the operator's, gated on that verdict.
