# Backend Architecture Re-Audit Report

**Project:** MetalDocs
**Code HEAD:** `dadb8275` (audited at working HEAD `9da740d6`, a docs-only status-flip over `dadb8275`; source identical)
**Date:** 2026-06-20
**Audit type:** Post-M7 terminal Grade-A re-audit (mission `grade-a-completion`, §8 terminal acceptance)
**Method:** 10-dimension independent multi-agent audit (one reader per dimension) with a refute-by-default adversarial skeptic re-reading every cited `file:line` for each Critical/Major. Only skeptic-confirmed findings count toward the §8 bar. H-D/H-G class counts come from greps run at HEAD (reproduced verbatim in §6). Build and full test suite run at HEAD.

> **Operator note (HS-5):** Contract / API misses the A− bar for a **fourth** consecutive re-audit (B+ → B− → B → **B+**). Per the mission HS-5 hard-stop this audit does **not** open M8 or any remediation milestone. See §9.

---

## 1. Method

Each dimension was graded A–F from the actual code at HEAD `dadb8275`, citing `file:line` (no grading by assertion, no trusting M0–M7 evidence as proof of the bar). Every Critical/Major was independently re-read by an adversarial skeptic (refute-by-default) before a binding verdict. The four §8 acceptance checks were then evaluated. The H-D honest two-part gate (`wiki/architecture/api-contract.md` §5b) and the H-G greps were re-run at HEAD; results are embedded verbatim in §6 and, critically, the audit also tests the **intent** of each class (zero response literals; zero cross-module reach-without-a-port), not only the path-scoped grep.

Baseline columns: **2026-06-13** = the original architecture audit (first 7 dims = B, last 3 = C). **post-M6** = the 2026-06-19 re-audit at HEAD `5650b328`.

---

## 2. Scorecard

| # | Dimension | 2026-06-13 | post-M6 (5650b328) | This re-audit (dadb8275) | Δ vs post-M6 |
|---|-----------|:---:|:---:|:---:|:---:|
| 1 | Authz / capability model | B | A− | **A** | +1 |
| 2 | Security / tenant isolation | B | B+ | **A−** | +1 |
| 3 | Sessions / auth lifecycle | B | A− | **B+** | −1 |
| 4 | Middleware / HTTP kernel | B | A− | **A−** | = |
| 5 | Persistence / transactions | B | A− | **A−** | = |
| 6 | Code quality / Go idioms | B | B+ | **A−** | +1 |
| 7 | Legacy / dead-code | B | A− | **A−** | = |
| 8 | Module boundaries / DDD | C | A− | **A−** | = |
| 9 | Contract / API layer | C | B | **B+** | +1 |
| 10 | Composition / observability | C | A− | **B+** | −1 |

Two regressions vs post-M6 (sessions, composition) are not new code defects introduced by M7 — they are findings the prior re-audit did not surface, exposed here by independent reads with a wider lens than the path-scoped H-D grep. M7 (typed-body parity) did not touch the regressed paths.

---

## 3. §8 Pass-Bar Verdict (all four required)

### Check 1 — the 3 formerly-C dimensions (module-boundaries, contract-api, composition) all ≥ A−?

- **Module boundaries / DDD: A− — PASS.** Baseline-C concern (raw SQL on IAM-owned `iam_users`/`iam_user_roles`) fully resolved: every production touch is inside `internal/modules/iam/`; consumers resolve through documented IAM ports (`auth/infrastructure/postgres/repository.go:23-24`, `security/infrastructure/postgres/repository.go:23-24`, `documents/module.go:50`). One confirmed Major remains (search → taxonomy `document_profiles` raw SQL, no port) but it is read-only, single-table, non-security; reader capped the dimension at A− (meets the bar). **However see Check 2/Check 4: that same finding is a confirmed Major and an honest-H-G-class violation.**
- **Contract / API layer: B+ — FAIL.** M7's in-scope work is genuinely strong (prior audit-export-status Major closed; all 4 documents 200 schemas tri-source aligned; Part A grep = 0; every Part B survivor correctly allowlisted). But the dimension's thesis — *every public delivery route emits a typed struct* — is falsified at HEAD by a live, spec-declared route emitting a `map[string]any` response literal with its generated model sitting unused (`iam/presence/handler.go:83`). One live confirmed Major → below A−.
- **Composition / observability: B+ — FAIL.** Composition-root wiring is A-grade and the prior WebSocket-501 + audit-export Major are closed, but `/api/v1/metrics` emits a fully untyped `map[string]any` body with no generated `MetricsResponse` Go type, and the declared schema omits the emitted `scheduler`/`db_pool` keys (contract-vs-runtime divergence). One confirmed Major → below A−.

**Check 1: FAIL** (Contract B+ and Composition B+, both < A−).

### Check 2 — 0 skeptic-confirmed Critical/Major?

**5 skeptic-confirmed Majors** (§4). **Check 2: FAIL.**

### Check 3 — H-D class = 0 (zero response literals)?

Documented two-part gate (path-scoped to `internal/modules/*/delivery/http/`): Part A = 0, Part B = all survivors allowlisted → **gate reports 0** (§6). But the gate's path scope is itself a blindspot: **2 live response literals on public, spec-declared routes survive outside the gated path** — `iam/presence/handler.go:83` and `internal/platform/observability/http.go:183-194`. Measured by intent (the §8 definition: "handler emits ⊆ declared OpenAPI ⊆ FE codegen"), **honest H-D = 2, not 0.**

**Check 3: FAIL** (honest H-D = 2; the path-scoped gate masks both sites).

### Check 4 — H-G class = 0?

Documented greps return 0 (no cross-module `iam_users`/`iam_user_roles` SQL; the 7 `"published"` hits are doc-comments) (§6). But the §8 H-G definition is "no raw SQL against another module's owned table." Search issues raw SQL against taxonomy-owned `metaldocs.document_profiles` with no port (`search/infrastructure/v2documents/reader.go:40,65`) — a confirmed Major and a cross-module-reach the narrow grep does not test. **Honest H-G = 1, not 0.**

**Check 4: FAIL** (honest H-G = 1; the narrow grep only tests the IAM tables).

### OVERALL PASS-BAR: **FAIL** (0 of 4 checks pass)

---

## 4. Skeptic-Confirmed Critical/Major Findings

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Major | Sessions / auth lifecycle | Account deactivation not enforced on live sessions (CWE-613) | `internal/modules/auth/application/service.go:368-400, 845-882` |
| 2 | Major | Middleware / HTTP kernel | Method-routed 405s return Go stdlib `text/plain`, not problem+json (D-03 violation) | `internal/modules/documents/delivery/http/handler.go:177-178` (+ iam, approval) |
| 3 | Major | Module boundaries / DDD | Search raw SQL against taxonomy-owned `document_profiles`, no port (H-G class) | `internal/modules/search/infrastructure/v2documents/reader.go:40, 65` |
| 4 | Major | Contract / API layer | Live spec-declared presence-snapshot route emits `map[string]any` literal; generated `PresenceSnapshotResponse` unused (H-D class, gate-evading) | `internal/modules/iam/presence/handler.go:83` |
| 5 | Major | Composition / observability | `/api/v1/metrics` emits untyped `map[string]any`; no `MetricsResponse` type; undeclared `scheduler`/`db_pool` keys (H-D class, gate-evading) | `internal/platform/observability/http.go:175-196` |

**Skeptic verdicts (all `confirmed`):**

1. **Sessions** — `ResolveSession`→`buildCurrentUser` never re-checks `identity.IsActive` (consulted only at login, `service.go:270`); IAM `deactivate` (`people_service.go:664-666`)→`UpdateUser` (`service.go:618-631`) issues the `is_active` UPDATE with no `RevokeSessionsByUserID`; force-logout explicitly deferred (`people_service.go:655`). Change-password/reset DO revoke in-tx (`service.go:495, 702`), so the CWE-613 pattern is known and deliberately absent from deactivate. Bounded window (idle 30m / absolute 12h) → Major, not Critical.
2. **Middleware** — Go 1.22 method-prefixed routes (`"PATCH /api/v1/documents/{id}"`) with no fallthrough/405 interceptor (`main.go:327` plain `NewServeMux`) cause `DELETE` to return stdlib `405` `Content-Type: text/plain` body, bypassing D-03 problem+json. In-body `WriteMethodNotAllowed` guards never execute (mux rejects method before dispatch). Intra-module inconsistency: iam `sessions_handler.go:84` emits problem+json 405 while `routes_memberships.go:92-96` emit text/plain. Live error-contract violation across many endpoints → Major (Allow header is correct → not Critical).
3. **Module boundaries** — `reader.go:38-45` (projection) and `63-70` ($5 family filter) embed raw SQL on `metaldocs.document_profiles`, hardcoding columns (`code`, `family_code`, `tenant_id`) and the sentinel-tenant fallback `ORDER BY`. Taxonomy owns the table exclusively and exposes a typed domain; search imports no taxonomy port. Taxonomy-side schema/fallback change breaks search silently. Canonical DDD-boundary Major → not downgradable (read-side projection, no wire drift → not Critical).
4. **Contract / API** — `handler.go:83` writes `json.NewEncoder(w).Encode(map[string]any{"items": items})` after `WriteHeader(200)` on `GET /api/v1/iam/presence/snapshot` (wired `main.go:344/758`). OpenAPI declares `200 → PresenceSnapshotResponse` (`openapi.yaml:527-538`); `api.gen.go:465` already generates the typed model, unused. Not on §5b allowlist. The §5b greps are path-scoped to `internal/modules/*/delivery/http/`; this handler lives at `internal/modules/iam/presence/` → real terminal-gate evasion. Wire bytes currently match the model → not Critical, but exact untyped-200 class this dimension exists to kill, live at the Grade-A gate.
5. **Composition** — `MetricsHandler` (`http.go:175-196`) builds `payload := map[string]any{"items": items}` augmented with runtime/scheduler/db_pool (each `map[string]any`), json-encoded to the wire on public route `/api/v1/metrics` (`main.go:553`). `grep MetricsResponse` over `internal`+`apps` (non-test) returns nothing — no Go type binds the declared schema. `scheduler`/`db_pool` keys are not declared in `openapi.yaml`. Lives in `internal/platform/observability/` → invisible to the §5b gate by construction. Confirmed Major (runtime correct → not Critical).

No findings were refuted or downgraded to Minor at the Critical/Major level.

---

## 5. Selected Minor Findings (non-binding)

- Authz: `requireDocEditDraft` uses a RW tx for a read-only authz probe (`documents/application/fillin_authz.go:21-23`) — explicitly deferred, non-security.
- Composition: scheduler job invocations carry no OTel span (`jobs/scheduler/scheduler.go`), though the span pattern exists in two app services.
- Contract: two off-spec health routes emit untyped literals (`internal/platform/observability/health.go:24,33`).

(Full per-dimension Minor lists live in the workflow transcript; the prior post-M6 artifact §7 enumerates the carried Minors, none of which gate.)

---

## 6. Class Re-Measurement (verbatim at HEAD)

### H-D — two-part honest gate (`wiki/architecture/api-contract.md` §5b)

**Part A — `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | grep -v _test.go`:**
```
(no output)   → 0
```

**Part B — `grep -rEn 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | grep -v _test.go`:**
```
internal/modules/audit/delivery/http/handler.go:404:    payload := map[string]any{}                         # decode buffer → AuditEventItem.Payload (domain-mirror) — allowlist
internal/modules/auth/delivery/http/handler.go:98,109,127,204                                               # recordAudit emit params — allowlist
internal/modules/controlleddocuments/delivery/http/routes.go:89,224: formData := map[string]any(nil)       # command input — allowlist
internal/modules/documents/delivery/http/handler.go:615: ContentFormData: map[string]any{...}              # command input — allowlist
internal/modules/iam/delivery/http/people_handler.go:321,454                                                # recordAudit emit params — allowlist
internal/modules/security/delivery/http/handler.go:54: Evidence map[string]any                             # domain-mirror struct field — allowlist
```
Every Part B survivor is on the §5b non-response allowlist → the **path-scoped gate reports H-D = 0.**

**Gate-evasion sweep (outside the gated path — the honest measure):**
```
internal/modules/iam/presence/handler.go:83          _ = json.NewEncoder(w).Encode(map[string]any{"items": items})   # response literal, PresenceSnapshotResponse unused
internal/platform/observability/http.go:183-194      payload := map[string]any{...}; json.NewEncoder(w).Encode(payload)  # response literal, no MetricsResponse type
```
**Honest H-D class = 2** (live response literals on public spec-declared routes; both confirmed Major).

### H-G — cross-module reach / hardcoded domain-state

**`grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_users' internal/modules/ | grep -v iam/ | grep -v _test.go`:** `(no output) → 0`
**`grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_user_roles' internal/modules/ | grep -v iam/ | grep -v _test.go`:** `(no output) → 0`
**`grep -rEn '"published"' internal/modules/ | grep -v _test.go | grep -v /domain/ | grep -v api.gen.go`:** 7 hits, all doc-comments in `documents/approval/application/{obsolete,publish,supersede}_service.go` → 0 SQL/status literals.

Narrow grep → **H-G (IAM-table scope) = 0.** But the §8 H-G definition ("no raw SQL against another module's owned table") is violated by search → taxonomy `document_profiles` (`reader.go:40,65`). **Honest H-G class = 1** (confirmed Major #3).

### Build & tests

```
GOFLAGS=-mod=mod go build ./...          → exit 0 (clean)
GOFLAGS=-mod=mod go test -count=1 ./...   → exit 0, 0 FAIL (full suite green)
```

---

## 7. Terminal Acceptance Verdict

**VERDICT: FAIL** — 0 of 4 §8 checks pass.

| Check | Requirement | Result |
|---|---|---|
| 1 | 3 formerly-C dims all ≥ A− | **FAIL** — Contract B+, Composition B+ |
| 2 | 0 skeptic-confirmed Critical/Major | **FAIL** — 5 confirmed Majors |
| 3 | H-D class = 0 (zero response literals) | **FAIL** — honest H-D = 2 (gate path-scope masks presence + metrics) |
| 4 | H-G class = 0 | **FAIL** — honest H-G = 1 (search → taxonomy `document_profiles`) |

M7's in-scope contract work is real and good (audit-export Major closed, 4 documents schemas tri-source aligned, every in-path Part B survivor correctly allowlisted, build+tests green). The terminal gate nonetheless fails because the program's H-D/H-G gates are **path-scoped to `internal/modules/*/delivery/http/` and the two IAM tables**, and that scope does not cover the whole public API surface: two response literals (`iam/presence`, `platform/observability/metrics`) and one cross-module raw-SQL reach (`search`→taxonomy) sit just outside the greps and are live at HEAD. Independent grading also placed Contract and Composition at B+ (each holds one live Major), so even by grade alone Check 1 fails.

---

## 8. Root Cause

The repeat Contract/API miss is structural, not effort: a **bounded, grep-scoped sweep** can only ever close what its scope greps. Each milestone closed the sites inside the scope, and each subsequent independent read found equivalent untyped-200 / cross-module sites just outside it (this round: `iam/presence`, `platform/observability`, `search→document_profiles`). The gate's definition of "done" (Part A = 0 + Part B allowlisted *within `delivery/http`*) is narrower than the §8 intent ("every public route typed; no cross-module reach"). Until the typed-body contract is enforced at the whole public-route surface (codegen-first `StrictServerInterface` wiring) or the gate scope is widened to match intent, the dimension will keep missing.

---

## 9. HS-5 Hard-Stop

Contract / API has now missed the A− bar a **fourth** time (B+ → B− → B → B+). Per the mission HS-5 rule this audit **does not** open M8 and **does not** start any remediation milestone. The bounded-sweep approach has repeatedly failed to close the dimension. The decision is the operator's:

- **(A) Full codegen-first rewire:** route every public handler (including `iam/presence`, `platform/observability`, and any non-`delivery/http` route) through generated `StrictServerInterface` typed responses; eliminate `map[string]any` response literals program-wide; add a port for search→taxonomy `document_profiles`. This is the structural fix that closes the class rather than the in-scope instances.
- **(B) Re-scope the §8 bar:** if presence/metrics/health are intentionally exempt (e.g., off-spec operational endpoints), amend the §8 acceptance + the §5b gate to state the exemption explicitly and widen the H-D/H-G greps to the true in-scope surface, then re-measure. This converts a hidden blindspot into a declared boundary.

Either way, the Grade-A terminal sign-off is **not** reached at HEAD `dadb8275`. Grade A is the operator's declaration, not the agent's.
