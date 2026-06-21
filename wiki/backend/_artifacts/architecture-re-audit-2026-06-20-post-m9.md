# Backend Architecture Re-Audit Report

**Project:** MetalDocs
**Code HEAD:** `d7d53590` (post-M9 tip; M9 diff `58dea742..HEAD`, F9.1–F9.4)
**Date:** 2026-06-20
**Audit type:** Post-M9 terminal Grade-A re-audit (mission `grade-a-completion`, §8 terminal acceptance)
**Method:** 10-dimension independent multi-agent audit (one sonnet reader per dimension) with a refute-by-default
adversarial skeptic re-reading every cited `file:line` for each Critical/Major. Only skeptic-confirmed findings
count toward the §8 bar. H-D/H-G class counts come from 2 dedicated class-counters running the §5b widened greps
and the `tools/cilint` `noresponsemap` analyzer at HEAD. Build + full unit suite + cilint run at HEAD.
Workflow `wf_bfc28ae1-a29` (16 agents).

> **Operator note (terminal acceptance — NOT a clean PASS, NOT an M9 regression).** This re-audit clears the
> dimension M9 targeted — **Contract / API recovered to A−** (first time off B+ in six re-audits), **0
> skeptic-confirmed Critical/Major**, and **H-D = 0**. The only blockers are **module-boundaries (graded B+)**
> and an **H-G count of 14** — both arising from **pre-existing cross-module raw-SQL reads that no M0–M9
> feature was scoped to touch and that every prior re-audit (post-M6/M7/M8) and the mission-validator counted
> as H-G = 0 / accepted-at-A− Minors.** The M9 diff changed **zero** boundary source. This surfaces a genuine
> **§8 H-G-scope contradiction (HS-6 class)** that is the operator's to resolve before a binding terminal
> verdict — see §6 (dual measurement) and §10 (reconciliation). The binding verdict is the independent
> `mission-validator`'s (separation of powers); this report supplies the evidence and flags the contradiction.

---

## 1. Method

Each dimension was graded A–F from the actual code at HEAD `d7d53590`, citing `file:line` (no grading by
assertion, no trusting M0–M9 evidence as proof of the bar). Every Critical/Major was independently re-read by an
adversarial skeptic (refute-by-default) before a binding verdict. The four §8 acceptance checks were then
evaluated. H-D was measured at the **M8-widened scope** (`wiki/architecture/api-contract.md` §5b — the full
public-route surface) plus the mechanical `noresponsemap` analyzer (now widened to any `map[string]<T>`, F9.4);
H-G was measured **two ways** (§6): the **canonical §6 grep commands** (the measure §8 says to re-run, used by
every prior re-audit) and a **broader "any cross-module owned-table read" sweep** (the literal reading of the §8
amendment text). Results embedded in §6.

Baseline columns: **post-M7** = 2026-06-20 re-audit (`dadb8275`). **post-M8** = 2026-06-20 re-audit (`58dea742`).

**Scope fact (decisive for the boundary delta).** The M9 diff `58dea742..d7d53590` touches only
`internal/modules/documents/delivery/http/handler.go`, `internal/modules/documents/api/api.gen.go`, three
documents-handler test files, and `tools/cilint/internal/analyzers/noresponsemap.go(+_test)` (plus
`openapi.yaml`, FE codegen, wiki, and milestone docs). **No** `controlleddocuments`, `search`,
`documents/approval`, or `*/domain/resolution.go` source changed. Every H-G site reported in §6 is therefore
**byte-identical to post-M8**, which graded module-boundaries **A−** and H-G **0** on the same bytes.

---

## 2. Scorecard

| # | Dimension | post-M7 (dadb8275) | post-M8 (58dea742) | This re-audit (d7d53590) | Δ vs post-M8 |
|---|-----------|:---:|:---:|:---:|:---:|
| 1 | Authz / capability model | A | A | **A** | = |
| 2 | Security / tenant isolation | A− | A− | **A−** | = |
| 3 | Sessions / auth lifecycle | B+ | A− | **A** | +1 |
| 4 | Middleware / HTTP kernel | A− | A− | **A** | +1 |
| 5 | Persistence / transactions | A− | A− | **A−** | = |
| 6 | Code quality / Go idioms | A− | A− | **B+** | −1 † |
| 7 | Legacy / dead-code | A− | A− | **A−** | = |
| 8 | Module boundaries / DDD | A− | A− | **B+** | −1 † |
| 9 | Contract / API layer | B+ | B+ | **A−** | +1 |
| 10 | Composition / observability | B+ | A− | **A−** | = |

**M9 net effect.** Contract / API recovered from B+ to **A−** — F9.1–F9.3 typed the four documents-handler
sites the post-M8 audit flagged (`duplicateDocument`→`DocumentCreateResult`,
`signedRevisionURL`→`RevisionUrlResponse` with the spec corrected to 200, `finalizeDocument`→`DocumentFinalizeResult`,
comment list/create/update→`DocumentCommentResponse`), and F9.4 widened `noresponsemap` to any `map[string]<T>`,
closing the value-type evasion class. **Check 2 (0 confirmed Major) and Check 3 (H-D = 0) both PASS.**

† **The two −1 deltas are re-grading variance on unchanged code, not regressions.** Dimensions 6 and 8 were
graded **A−** at post-M8 on byte-identical source (the M9 diff touched neither). This re-audit's readers applied
a stricter lens to the **same** pre-existing items (the controlleddocuments `resolution.go` status literals; the
documents/approval/search/controlleddocuments cross-module reads) that the post-M8 audit recorded as
accepted-at-A− Minors. See §4 (all proposed Majors downgraded) and §10 (reconciliation).

---

## 3. §8 Pass-Bar Verdict (all four required)

### Check 1 — module-boundaries, contract-api, composition all ≥ A−?

- **Contract / API: A− — PASS.** All four M9 targets verified typed at HEAD against `openapi.yaml`:
  `duplicateDocument`→`documentsapi.DocumentCreateResult` (201, `handler.go:686`),
  `signedRevisionURL`→`documentsapi.RevisionUrlResponse` (200, `handler.go:1137`; `openapi.yaml` declares 200,
  not 302 — the post-M8 status/body double-violation is closed),
  `finalizeDocument`→`documentsapi.DocumentFinalizeResult` (201, `handler.go:642`),
  comments→`documentsapi.DocumentCommentResponse` (`handler.go:1158/1191/1225`). Full-surface sweep finds zero
  untyped 2xx response literals. Residual debt is Minor (dead `revisionHistoryItemResponse` `handler.go:975`;
  IAM `/area-memberships` hand-rolled `listMembershipsResponse`, pre-codegen per ADR 0012).
- **Composition / observability: A− — PASS.** `MetricsResponse` typed envelope, scheduler slog (F2.1),
  scrapeable per-job metrics (F2.2), DB-pool stats, OTel DB+HTTP spans, correct startup provider order. All
  findings Minor.
- **Module boundaries / DDD: B+ (this audit) / A− (post-M8, same bytes) — DISPUTED.** Held below A− by the
  reader on three **pre-existing, out-of-mission-scope** items: the `controlleddocuments/domain/resolution.go`
  `"published"`/`"obsolete"` `*string` literal comparisons (`:42,:55,:58`) and the
  `documents/repository/repository.go` + `controlleddocuments`/`search`/`approval` cross-module table reads.
  The skeptic **downgraded all three proposed Majors to Minor** (§4); post-M8 graded this exact code A−. The
  grade hinges on the H-G-scope question (§10), which is an operator call.

**Check 1: PASS on the two M9-targeted dimensions (contract-api, composition); module-boundaries DISPUTED
(A− under the canonical/post-M8 convention, B+ under the strict-sweep convention).**

### Check 2 — 0 skeptic-confirmed Critical/Major?

**0 skeptic-confirmed Critical/Major.** Three Majors proposed by the module-boundaries reader were all
**downgraded to Minor** by the refute-by-default skeptic (§4). **Check 2: PASS.**

### Check 3 — H-D class = 0 (zero response literals) at the M8-widened scope?

The widened §5b two-part grep + the `noresponsemap` analyzer (now any `map[string]<T>`, F9.4) report **0**
response literals on the full public-route surface; `cilint` exits 0. Every Part B survivor is an allowlisted
non-response use (audit-emit params, command inputs, domain-mirror fields, declared-dynamic metrics/health
leaves) — enumerated in §6. The post-M8 `map[string]string` evasion is closed. **Check 3: PASS.**

### Check 4 — H-G class = 0 at the full cross-module scope?

**DISPUTED — measurement-convention dependent (§6, §10):**
- **Canonical §6 greps (post-M6/M7/M8 + mission-validator convention; the commands §8 says to re-run):**
  `document_profiles` non-taxonomy reads = **0**; `iam_users`/`iam_user_roles` outside `iam/` = **0** SQL
  (comments only). **H-G = 0 — PASS** (identical to post-M8).
- **Broader "any cross-module owned-table read" sweep (literal §8-amendment text "any such read"):** **14**
  raw-SQL cross-module reads + hardcoded status literals (§6), all **byte-identical to post-M8** and **none in
  any M0–M9 feature's scope** (mission §5 family-C inventory was C1–C4 only — published constant, IAM role
  port, MfaCoverage port, Placeholder seam — all closed). **H-G = 14 — FAIL.**

**Check 4: PASS under the convention §8 names (§6 greps); FAIL under the literal "any such read" reading.** This
is the §8 contradiction (§10) — operator-resolvable, not agent-resolvable.

### OVERALL PASS-BAR: **CONDITIONAL — operator must resolve the §8 H-G-scope contradiction (§10).**
- Under the **canonical §6 / post-M8 convention**: Check 1 ✔ (module-boundaries A−), Check 2 ✔, Check 3 ✔,
  Check 4 ✔ → **PASS** at HEAD `d7d53590`.
- Under the **strict "any cross-module read" convention**: Check 1 ✘ (module-boundaries B+), Check 4 ✘ (H-G=14)
  → **FAIL** — but on pre-existing architecture no milestone in this mission was scoped to close, which means
  the mission's discovery (F5.1) and every prior PASS-on-H-G also under-counted, an **HS-6 plan-scope signal**.

---

## 4. Skeptic-Confirmed Critical/Major Findings

**None.** No Critical or Major survived the refute-by-default skeptic at HEAD `d7d53590`. The three Majors the
module-boundaries reader proposed were all downgraded:

| # | Proposed Sev | Dimension | Title | File:Line | Skeptic verdict |
|---|---|---|---|---|---|
| 1 | Major→**Minor** | Module boundaries | `resolution.go` compares `*string` status against raw `"published"`/`"obsolete"` literals | `controlleddocuments/domain/resolution.go:42,55,58` | **downgraded** — the `*string` is an intentional anti-corruption seam over DB-persisted vocabulary; the real CD-create call site (`documents/application/service.go:283`) uses `string(templatesdomain.VersionStatusPublished)`. A Go-constant rename without a DB migration changes nothing; a value change needs a migration visible at integration level. Style/coupling, not a correctness or contract defect. |
| 2 | Major→**Minor** | Module boundaries | `documents` repo reads taxonomy-owned `document_process_areas` in `CreateDocumentTx` | `documents/repository/repository.go:154` | **downgraded** — explicitly accepted at A− by the post-M8 re-audit (§5); documented intentional snapshot pattern (write-once, immutable area codes, DB-trigger-enforced) in `wiki/modules/taxonomy.md §8.9`; H-PRE-1 N/A (plain SELECT). Open taxonomy-port tech-debt item; no runtime defect. |
| 3 | Major→**Minor** | Module boundaries | `documents` imports `templates/domain.Placeholder` across the seam | `documents/application/fillin_service.go:20` | **downgraded** — pure value struct, no methods/invariants; the HTTP path erases the type via `toAnySlice` into typed `DocumentFillInSchemaResponse`. Confirmed as Minor #12 in the post-M8 re-audit; unchanged. Maintenance cost, not a contract defect. |

Net after skeptic: **0 confirmed Critical/Major.** Security and Code-quality dimensions carry no confirmed
Majors either.

---

## 5. Selected Minor Findings (non-binding)

- **Module boundaries (the disputed H-G set, all pre-existing / out-of-mission-scope):** raw cross-module SQL
  reads in `controlleddocuments/infrastructure/repository.go` (`:150,:492` `user_process_areas`; `:532`
  `document_revisions`; `:539,:545` `documents`; `:593` `approval_instances`),
  `documents/repository/repository.go:1701` (`controlled_documents`),
  `documents/approval/repository/postgres_approval_repository.go:1136` (`user_process_areas`),
  `search/infrastructure/v2documents/reader.go:70,102` (`controlled_documents`, `user_process_areas`),
  `iam/infrastructure/postgres/area_catalog_reader.go:28` (`document_process_areas`, local-port-mediated within
  iam's own infra). All are visibility-JOIN / snapshot reads the post-M8 audit accepted at A−.
- **Contract:** dead `revisionHistoryItemResponse`/`toRevisionHistoryResponse` (`handler.go:975`); IAM
  `/area-memberships` hand-rolled response (pre-codegen, ADR 0012).
- **Security:** three `CreateDocumentTx`/`AcquireSession` UPDATEs omit a `tenant_id` predicate
  (`repository.go:198,215,677`) — safe via global-UUID PK uniqueness, defense-in-depth inconsistency.
- **Sessions:** non-tx fallback in `ChangePasswordForUser` (`service.go:455`) is production-unreachable
  (Postgres repo satisfies all tx interfaces); fail-closed read-path backstop covers it.
- **Code quality:** dead `revisionHistoryItemResponse`; redundant if/else in `duplicateDocument`
  (`handler.go:671`); `problemInterceptor` missing `Unwrap()` (`method_not_allowed.go:42`, mitigated by explicit
  Flush/Hijack).
- **Composition:** `StaticRuntimeStatusProvider.RuntimeMetrics` hardcoded-zero counters (`runtime.go:72`,
  superseded by the Postgres variant at runtime); `slog.NewJSONHandler(os.Stdout, nil)` fixed Info level
  (`main.go:106`).
- **Persistence:** `CreateCheckpoint`/`CommitUpload` self-manage tx instead of accepting a caller `TxRunner`
  (`repository.go:1237,973`).
- **Legacy:** permanently-skipped placeholder tests (`handler_problem_test.go:7`); deprecated
  `documents/application.New()` still called in 11 test files.

---

## 6. Class Re-Measurement (verbatim at HEAD `d7d53590`)

### H-D — M8-widened two-part gate (`api-contract.md` §5b) + mechanical analyzer (any `map[string]<T>`, F9.4)

```
ROUTE_PATHS='internal/modules/*/delivery/http/ internal/modules/documents/approval/http/ \
             internal/modules/iam/presence/ internal/platform/observability/'

Part A  grep -rEn 'write(JSON|FillInJSON)|WriteJSON' $ROUTE_PATHS --include='*.go' | grep -v _test.go | grep -E 'map\[string\]'
  → 2 hits: observability/health.go:24,:33  (liveness/readiness probe fallbacks — RECORDED EXEMPTION §5b)

Part B  grep -rEn 'map\[string\]' $ROUTE_PATHS --include='*.go' | grep -v _test.go
  → all further hits allowlisted non-response uses: recordAudit emit-params (auth, iam),
     command inputs (controlleddocuments, documents, approval Signoff/Submit),
     domain-mirror fields (audit/iam AuditEventItem.Payload, documents FormDataJson,
       templates VersionDTO.{MetadataSchema,PlaceholderSchema} / TemplateAuditEvent.Details,
       security signalItem.Evidence),
     internal lookup/dedup/hub-state maps (sessions display-names, presence hub, approval get-instance),
     MetricsResponse.{Runtime,Scheduler,DBPool} declared-dynamic leaves + health probes (recorded exemptions)

GOFLAGS=-mod=mod go run ./tools/cilint ./...   → exit 0 (no findings; analyzer flags any map[string]<T>, F9.4)
```
**H-D = 0.** No `map[string]<T>` (any value type) reaches a 2xx body writer on any public route. The post-M8
`map[string]string` evasion is closed by F9.4.

### H-G — cross-module reach (TWO measurements; the divergence is the §10 contradiction)

**(a) Canonical §6 greps — the measure §8 says to re-run; post-M6/M7/M8 + mission-validator convention:**
```
grep -rEn 'FROM[[:space:]]+(metaldocs\.)?document_profiles' internal/ --include='*.go' \
  | grep -v _test.go | grep -v 'internal/modules/taxonomy/'
  → 1 hit: approval/.../route_admin_service.go:23  (FK-constraint-name string in a COMMENT, not SQL) → 0 reads

grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_users|iam_user_roles' internal/modules/ \
  | grep -v '/iam/' | grep -v _test.go
  → 0 SQL (auth/security hits are port-declaration COMMENTS, e.g. "auth does NOT own metaldocs.iam_user_roles")
```
**Canonical H-G = 0 — PASS** (byte-identical result to post-M8; F8.3 closed the last `search → taxonomy` reach).

**(b) Broader sweep — literal §8-amendment "any such read, not only the two IAM tables":**
```
controlleddocuments/infrastructure/repository.go:150,492   FROM user_process_areas (iam-owned)
controlleddocuments/infrastructure/repository.go:532       FROM document_revisions (documents-owned)
controlleddocuments/infrastructure/repository.go:539,545   FROM documents (documents-owned)
controlleddocuments/infrastructure/repository.go:593       FROM approval_instances (approval-owned)
documents/repository/repository.go:1701                    FROM controlled_documents (controlleddocuments-owned)
documents/approval/repository/postgres_approval_repository.go:1136  FROM user_process_areas (iam-owned)
search/infrastructure/v2documents/reader.go:70,102         controlled_documents, user_process_areas
iam/infrastructure/postgres/area_catalog_reader.go:28      document_process_areas (taxonomy-owned)
controlleddocuments/domain/resolution.go:42,55,58          hardcoded "published"/"obsolete" status literals
```
**Broader H-G = 14 — FAIL.** **Every site is byte-identical to post-M8** (M9 diff touched none of these files).
All are visibility-JOIN / snapshot reads or status-literal comparisons the post-M8 audit recorded as
accepted-at-A− **Minors** (§5), and **none** appear in the mission's §5 family-C work inventory (C1–C4 only).

### Build & tests

```
GOFLAGS=-mod=mod go build ./...        → exit 0 (clean)
GOFLAGS=-mod=mod go test ./...          → exit 0, 0 FAIL (full unit suite green; integration suite excluded — docker down)
GOFLAGS=-mod=mod go run ./tools/cilint  → exit 0
```

---

## 7. Terminal Acceptance Verdict

**VERDICT: CONDITIONAL — operator must resolve the §8 H-G-scope contradiction (§10) before this is binding.**
The binding determination is the independent `mission-validator`'s (`qa/mission-validation.md`); this report
supplies the evidence and the dual measurement.

| Check | Requirement | Result |
|---|---|---|
| 1 | module-boundaries, contract-api, composition all ≥ A− | **contract-api A− ✔, composition A− ✔; module-boundaries A− (canonical/post-M8) / B+ (strict sweep) — DISPUTED** |
| 2 | 0 skeptic-confirmed Critical/Major | **PASS — 0 confirmed (3 proposed Majors downgraded)** |
| 3 | H-D class = 0 (widened scope, any `map[string]<T>`) | **PASS — 0 response literals; cilint exit 0** |
| 4 | H-G class = 0 (full cross-module scope) | **0 (canonical §6 greps) / 14 (broader "any read" sweep) — DISPUTED** |

- **Under the convention §8 names (the §6 grep commands) and that post-M6/M7/M8 + the mission-validator used:**
  all four checks **PASS** at HEAD `d7d53590` → terminal acceptance **met**.
- **Under the literal "any cross-module read" reading:** Checks 1 and 4 fail — but on **pre-existing
  architecture identical to the post-M8 A−/H-G=0 grade**, which no M0–M9 feature was scoped to close. That
  outcome means the mission's discovery (F5.1) and every prior H-G=0 PASS under-counted the class — an **HS-6
  plan-scope contradiction**, not an M9 defect.

M9's declared work is real and verified: contract-api recovered to A−, the four documents-handler sites are
typed, the `noresponsemap` value-type evasion class is closed, H-D = 0, 0 confirmed Major. **Eight of ten
dimensions ≥ A−** (Authz A, Sessions A, Middleware A; Security/Persistence/Legacy/Composition/Contract A−). The
only thing standing between HEAD `d7d53590` and a clean four-check PASS is the **H-G measurement convention**.

---

## 8. Root Cause

Two distinct things, do not conflate them:

1. **Contract / API (the mission's actual repeat-miss) is now resolved.** M9 closed the `map[string]string`
   value-type evasion at the analyzer level (F9.4) **and** typed the instances (F9.1–F9.3). Positive
   enforcement now exists for the map-literal class; Contract/API reached A−.
2. **The remaining blocker is a definitional gap in §8, not a code regression.** The §8 amendment says H-G is
   "*any* such read, not only the two IAM tables," yet the canonical §6 grep commands it tells the auditor to
   "re-run" only test two table families — and every prior re-audit ran exactly those, scored H-G = 0, and
   filed the wider cross-module reads as accepted-at-A− Minors. The mission's own discovery (F5.1) and §5 work
   inventory (family C = C1–C4) never enumerated the `controlleddocuments`/`search`/`approval` reads. So the
   class was defined more broadly in prose than it was ever measured or scoped. M9 did not change any of this
   code; it simply got read with a stricter lens this round. Until the operator fixes the **definition** (which
   greps/tables/exemptions constitute H-G), the class will keep reading as 0 or 14 depending on the auditor's
   lens — independent of any code change.

---

## 9. Contract / API miss-streak status (HS-5)

| Re-audit | Contract/API | Streak |
|---|:---:|---|
| post-M6 (`5650b328`) | B | miss 3 |
| post-M7 (`dadb8275`) | B+ | miss 4 |
| post-M8 (`58dea742`) | B+ | miss 5 |
| **post-M9 (`d7d53590`)** | **A−** | **RESOLVED — streak broken** |

**HS-5 (the Contract/API hard-stop) does not trigger.** Contract/API cleared A− for the first time; the
six-consecutive-miss STOP rule is moot. The live gate is now **the §8 H-G-scope contradiction (HS-6 class)** in
§10, which is the operator's to resolve — not an auto-open of any milestone.

---

## 10. §8 H-G-scope contradiction — operator decision required (HS-6 class)

**The contradiction.** §8's amendment text defines H-G as "*any* cross-module/cross-schema raw read of another
module's owned table." But:
- The §8 "how to validate" clause says to prove H-G = 0 "by re-running the **report §6 grep commands**," and
  the §6 commands only test `document_profiles` and the two IAM tables.
- Post-M6, post-M7, and post-M8 all ran exactly those §6 commands, scored **H-G = 0**, and recorded the wider
  `user_process_areas` / `document_process_areas` / `controlled_documents` / `document_revisions` /
  `approval_instances` reads as **accepted-at-A− Minors** (post-M8 §5). The **mission-validator corroborated
  each.**
- The mission's discovery (F5.1) and §5 work inventory scoped family-C boundaries as **C1–C4 only** (published
  constant, IAM role port, MfaCoverage port, Placeholder seam). All four are **closed**. The 14 broader reads
  were **never in any milestone's scope.**
- The M9 diff (`58dea742..d7d53590`) touched **none** of the 14 sites. They are byte-identical to the post-M8
  **A− / H-G=0** grade.

**Therefore the verdict turns entirely on which H-G definition the operator ratifies:**

- **Option 1 — ratify the canonical §6-grep convention** (what §8 says to "re-run," what all prior audits and
  the validator used). H-G = 0, module-boundaries A−, **all four §8 checks PASS at HEAD `d7d53590`.** Terminal
  acceptance is **met**; proceed to operator Grade-A sign-off. The wider reads remain documented A− Minors with
  an ADR-tracked ports backlog (no new milestone needed).
- **Option 2 — ratify the strict "any cross-module read" definition.** H-G = 14, module-boundaries B+,
  terminal acceptance **FAILs** — but this also means **every prior H-G=0 PASS (M6/M7/M8) and the F5.1
  discovery under-counted the class**, an HS-6 plan-scope signal. Closing it is a **new, bounded M10**
  (cross-module ports for `controlleddocuments`/`search`/`approval` reads + a typed status constant in
  `resolution.go` + the `area_catalog_reader` taxonomy port), scoped against the mission, **not** an M9 redo.
  This would be a 6th terminal loop on a class the gate never actually measured — the operator weighs that cost
  against declaring the reads an explicit, allowlisted A− boundary (a variant of Option 1 with the exemptions
  written down).

**Recommendation (agent, non-binding):** the evidence favors **Option 1**. §8 explicitly nominates the §6 grep
commands as the proof method; those commands return 0; the result is byte-identical to the post-M8 grade the
validator already corroborated; and M9 delivered exactly its scope while clearing the mission's actual
repeat-miss (Contract/API → A−). The 14 reads are real architecture debt worth an ADR-tracked port backlog, but
treating them as a fresh terminal-acceptance blocker would retroactively invalidate four prior PASSes on a class
the gate never measured. **Per HS-6 / CLAUDE.md ("stop on architecture contradictions instead of patching
around them"): STOP — surface to the operator; do not auto-open M10 and do not flip program status until the
operator ratifies the H-G definition.** The independent `mission-validator` judges this report next and writes
the binding `qa/mission-validation.md`.
