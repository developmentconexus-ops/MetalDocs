# Backend Architecture Re-Audit Report

**Project:** MetalDocs
**Code HEAD:** `58dea742` (post-M8 tip; M8 diff `a00dd78a..HEAD`, F8.1–F8.6)
**Date:** 2026-06-20
**Audit type:** Post-M8 terminal Grade-A re-audit (mission `grade-a-completion`, §8 terminal acceptance)
**Method:** 10-dimension independent multi-agent audit (one sonnet reader per dimension) with a refute-by-default
adversarial skeptic re-reading every cited `file:line` for each Critical/Major. Only skeptic-confirmed findings
count toward the §8 bar. H-D/H-G class counts come from 2 dedicated class-counters running the §5b widened greps
and the `tools/cilint` `noresponsemap` analyzer at HEAD. Build + full unit suite + cilint run at HEAD.
Workflow `wf_0aa2dd43-6a6` (18 agents).

> **Operator note (HS-5):** Contract / API misses the A− bar for a **fifth** consecutive re-audit
> (B+ → B− → B → B+ → **B+**). Per the mission HS-5 5th-miss hard-stop this audit does **not** open M9 or any
> remediation milestone. See §9.

---

## 1. Method

Each dimension was graded A–F from the actual code at HEAD `58dea742`, citing `file:line` (no grading by
assertion, no trusting M0–M8 evidence as proof of the bar). Every Critical/Major was independently re-read by an
adversarial skeptic (refute-by-default) before a binding verdict. The four §8 acceptance checks were then
evaluated. H-D was measured at the **M8-widened scope** (`wiki/architecture/api-contract.md` §5b — the full
public-route surface, not the old path-scoped grep) plus the mechanical `noresponsemap` analyzer; H-G was
measured at the full cross-module surface, not only the two IAM tables. Results embedded in §6.

Baseline columns: **post-M6** = 2026-06-19 re-audit (`5650b328`). **post-M7** = 2026-06-20 re-audit (`dadb8275`).

---

## 2. Scorecard

| # | Dimension | post-M6 (5650b328) | post-M7 (dadb8275) | This re-audit (58dea742) | Δ vs post-M7 |
|---|-----------|:---:|:---:|:---:|:---:|
| 1 | Authz / capability model | A− | A | **A** | = |
| 2 | Security / tenant isolation | B+ | A− | **A−** | = |
| 3 | Sessions / auth lifecycle | A− | B+ | **A−** | +1 |
| 4 | Middleware / HTTP kernel | A− | A− | **A−** | = |
| 5 | Persistence / transactions | A− | A− | **A−** | = |
| 6 | Code quality / Go idioms | B+ | A− | **A−** | = |
| 7 | Legacy / dead-code | A− | A− | **A−** | = |
| 8 | Module boundaries / DDD | A− | A− | **A−** | = |
| 9 | Contract / API layer | B | B+ | **B+** | = |
| 10 | Composition / observability | A− | B+ | **A−** | +1 |

**M8 net effect:** Sessions (F8.4 CWE-613 close) and Composition (F8.2 typed metrics envelope) both recovered
to A−; the post-M7 regressions are closed. Security (2) and Code quality (6) hold A− — the auditor-proposed
Majors in those two dimensions were **downgraded by the skeptic** (see §4), leaving zero confirmed Majors in
each. The single dimension that does not clear A− is **Contract / API (B+)**, held below the bar by 3
skeptic-confirmed Majors in `documents/delivery/http/handler.go` (§4).

---

## 3. §8 Pass-Bar Verdict (all four required)

### Check 1 — module-boundaries, contract-api, composition all ≥ A−?

- **Module boundaries / DDD: A− — PASS.** F8.3 closed the sole skeptic-confirmed Major that blocked this
  dimension in every prior re-audit (search → taxonomy raw `document_profiles` SQL). `search/.../reader.go:19`
  now injects `taxonomydomain.FamilyCodeResolver` (ADR 0038) and resolves via `ProfileCodesForFamily` /
  `ResolveFamilyCodes`; the raw cross-module SQL is gone. Remaining cross-table reads are pre-existing,
  accepted-at-A− authz-predicate / snapshot reads (Minors §5). Zero live Majors in the H-G class.
- **Contract / API layer: B+ — FAIL.** F8.1 (presence typed) and F8.2 (metrics typed envelope) genuinely
  closed the two presence/metrics Majors and the full surface is now under the `noresponsemap` analyzer. But
  **3 skeptic-confirmed Majors survive in `documents/delivery/http/handler.go`** — `duplicateDocument` (201
  `map[string]string` vs named `DocumentCreateResult`), `signedRevisionURL` (200 + JSON body vs spec-declared
  302/no-body), and the comment endpoints (local `commentResponse` vs generated `DocumentCommentResponse`).
  Three live confirmed Majors → below A−.
- **Composition / observability: A− — PASS.** F8.2 closed the prior blocking Major: `MetricsResponse` is a
  typed struct (`observability/http.go:183`), key-set-locked by `http_typed_test.go`; the three optional
  providers are wired in strict startup order; OTel middleware is correctly positioned. All findings Minor.

**Check 1: FAIL** (Contract/API B+ < A−).

### Check 2 — 0 skeptic-confirmed Critical/Major?

**3 skeptic-confirmed Majors** (§4), all in Contract / API. **Check 2: FAIL.**

### Check 3 — H-D class = 0 (zero response literals) at the M8-widened scope?

The widened §5b two-part grep + the `noresponsemap` analyzer report **0** `map[string]any` response literals
(cilint exits 0; all Part B survivors are allowlisted or named exemptions; §6). **But the gate's type scope is
a blindspot:** the `noresponsemap` analyzer (`tools/cilint`) flags only `map[string]any` / `map[string]interface{}`
literals — its `isMapStringAnyLiteral` returns false for `map[string]string`. The 3 confirmed contract Majors
emit `map[string]string` response literals on public routes and so **pass the mechanical gate silently**.
Measured by the §8 *intent* ("every public route emits its declared typed struct"), H-D is **not** 0 — there are
3 untyped response literals live on spec-declared routes.

**Check 3: FAIL** (mechanical gate reports 0, but the `map[string]any`-only scope masks 3 `map[string]string`
response literals — the same gate-evasion class as the prior path-scope blindspot, one type wider).

### Check 4 — H-G class = 0 at the full cross-module scope?

Class-counter sweep of all production Go (excluding `_test.go` and owning modules): **0** genuine cross-module
raw reads. `document_profiles` survives only as an FK-constraint-name string constant (`approval/.../route_admin_service.go:23`)
and test seed; `iam_users`/`iam_user_roles` outside `iam/` survive only as port-declaration comments in
auth/security. F8.3 removed the last `search → taxonomy` reach. **Check 4: PASS** (honest H-G = 0).

### OVERALL PASS-BAR: **FAIL** (2 of 4 checks pass: H-G ✔, module-boundaries/composition ✔ within Check 1;
Contract/API ✘, confirmed-Majors ✘, H-D-by-intent ✘).

---

## 4. Skeptic-Confirmed Critical/Major Findings

| # | Sev | Dimension | Title | File:Line | Skeptic |
|---|-----|-----------|-------|-----------|---------|
| 1 | Major | Contract / API | `duplicateDocument` emits `map[string]string` vs named spec schema `DocumentCreateResult` | `documents/delivery/http/handler.go:674` | **confirmed** |
| 2 | Major | Contract / API | `signedRevisionURL` emits 200 + JSON body vs spec-declared 302/no-body (double violation) | `documents/delivery/http/handler.go:1105` | **confirmed** |
| 3 | Major | Contract / API | Comment list/create/update emit local `commentResponse` vs generated `DocumentCommentResponse` | `documents/delivery/http/handler.go:1122,1159,1193` | **confirmed** |

**Skeptic verdicts (confirmed):**

1. **duplicateDocument** — `handler.go:674` emits `WriteJSON(201, map[string]string{document_id,initial_revision_id,session_id})`;
   `openapi.yaml:2576` declares `$ref: DocumentCreateResult`; generated `documentsapi.DocumentCreateResult`
   (`api.gen.go:208`) exists and is unused. `map[string]string` escapes the `noresponsemap` analyzer
   (`map[string]any`-only). Confirmed Major.
2. **signedRevisionURL** — `handler.go:1105` emits `WriteJSON(200, map[string]string{"url":url})`; `openapi.yaml:2904`
   declares only `302` with no body. Wrong status code **and** unspecced untyped body. `wiki/modules/documents.md:258`
   marking this "Aligned" is stale. Confirmed Major.
3. **Comment endpoints** — `handler.go:1122/1159/1193` call `toCommentResponse()` → package-local `commentResponse`
   (`:1217`); `openapi.yaml:2930/2956/2988` declare `$ref: DocumentCommentResponse`; generated type (`api.gen.go:188`)
   unused. Material field divergence: `content` local `json.RawMessage` vs generated `[]DocumentCommentContentNode`;
   `id` `string` vs `openapi_types.UUID`. Confirmed Major.

**Downgraded by skeptic (do NOT count toward the bar):**

| Sev (proposed) | Dimension | Title | Verdict | Why |
|---|---|---|---|---|
| Major→Minor | Security | `CreateDocumentTx` pointer UPDATE missing `tenant_id` predicate (`repository.go:197`) | **downgraded** | `docID` is `RETURNING id` from the same-tx INSERT bound with `d.TenantID`; not caller-supplied. No live cross-tenant path; defense-in-depth/style only. |
| Major→Minor | Code quality | Dead if/else in presence `writeJSON` — both arms `return err` (`presence/handler.go:191`) | **downgraded** | Inert `errors.Is` guard; observable contract unchanged. Readability nit, not Major. |
| Major→Minor | Code quality | `problemInterceptor` missing `Unwrap() http.ResponseWriter` (`method_not_allowed.go:42`) | **downgraded** | Explicit `Flush`/`Hijack` satisfy direct interface assertion (Go 1.20 `ResponseController` tries assertion before Unwrap); innermost wrapper only needs pass-through. Latent style gap, not Major. |

Net after skeptic: **3 confirmed Majors (all Contract/API); Security and Code quality clear A− with 0 confirmed Majors.**

---

## 5. Selected Minor Findings (non-binding)

- Contract: `finalizeDocument` (`handler.go:624`) emits `map[string]string` for an inline-only spec schema (no
  generated type) — wire shape preserved, off the typed-struct posture. **cilint scope gap:** `noresponsemap`
  targets `map[string]any` only, so `map[string]string` response literals are unguarded class-wide.
- Authz: stale Phase-2 comment in `capability_scope.go:31-35` describing a gap Phase 7 closed; `requireDocEditDraft`
  RW tx for a read-only authz probe (`fillin_authz.go:21`, flagged-for-operator defer).
- Security: `BumpLastSeen` updates all tenant rows for a user (`presence/repository.go:52`); cross-tenant session
  revoke / `active_session_id` clears lack `tenant_id` (safe via UUID uniqueness, inconsistent pattern).
- Sessions: deactivation path emits no audit event (`service.go:656`); latent CWE-613 if a future caller passes
  `newPassword` without `IsActive=false` through `UpdateUser` (`service.go:637`).
- Persistence: invite area-grants + audit emit are post-commit best-effort (documented bounded defers);
  `tx_ownership_test.go:17` walks a non-existent `documents_v2` path (passes vacuously).
- Module boundaries: pre-existing authz-predicate / snapshot cross-table reads (`user_process_areas` EXISTS in
  search/controlleddocuments; `document_process_areas` snapshot in documents; `audit_events` in templates) —
  all accepted at the post-M7 A−, none re-flagged Major.
- Composition: duplicate private `writeJSON` in `health.go:40`; unsynchronized metrics setter fields (startup
  single-threaded); narrow app-layer OTel span coverage; `normalizeRoute` collapse table missing M8-era routes.

---

## 6. Class Re-Measurement (verbatim at HEAD `58dea742`)

### H-D — M8-widened two-part gate (`api-contract.md` §5b) + mechanical analyzer

```
ROUTE_PATHS='internal/modules/*/delivery/http/ internal/modules/documents/approval/http/ \
             internal/modules/iam/presence/ internal/platform/observability/'

Part A  grep -rEn 'write(JSON|FillInJSON)|WriteJSON' $ROUTE_PATHS --include='*.go' | grep -v _test.go | grep 'map\[string\]any'
  → 2 hits: observability/health.go:24, :33   (liveness/readiness probe fallbacks — RECORDED EXEMPTION §5b)

Part B  grep -rEn 'map\[string\]any' $ROUTE_PATHS --include='*.go' | grep -v _test.go
  → 18 further hits, every one allowlisted or exempt:
     recordAudit emit-params (auth:98,109,127; iam:321) + signatures (auth:204; iam:454)
     command-input maps (controlleddocuments:89,224; documents:615; approval doc/signoff/submit handlers)
     domain-mirror field (security/handler.go:54 Evidence)
     comments (approval/contracts/route.go:237; iam/presence/handler.go:90)
     health.go probes (recorded exemption)
     MetricsResponse.{Runtime,Scheduler,DBPool} declared-dynamic leaves (recorded exemption, F8.2)

GOFLAGS=-mod=mod go run ./tools/cilint ./...   → exit 0 (no findings)
```
**Mechanical/path gate reports H-D = 0.** **BUT** the analyzer's type scope is `map[string]any`-only
(`noresponsemap.go` `isMapStringAnyLiteral` → false for `map[string]string`). The 3 confirmed Majors (§4) emit
`map[string]string` response literals on public spec-declared routes and pass the gate silently. **H-D measured
by §8 intent = 3** (untyped response literals on declared routes).

### H-G — cross-module reach (full scope, not just IAM tables)

```
grep -rEn 'document_profiles' internal/ --include='*.go' | grep -v _test.go | grep -v 'internal/modules/taxonomy/'
  → 1 hit: approval/application/route_admin_service.go:23  (FK-constraint-name string constant, not SQL) → 0 reads

grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_users|iam_user_roles' internal/modules/ | grep -v '/iam/' | grep -v _test.go
  → 0 SQL (auth/security hits are port-declaration comments, e.g. "auth does NOT own metaldocs.iam_user_roles")
```
**Honest H-G class = 0** (F8.3 removed the last `search → taxonomy` reach; all IAM/taxonomy table SQL stays in
its owning module).

### Build & tests

```
GOFLAGS=-mod=mod go build ./...        → exit 0 (clean)
GOFLAGS=-mod=mod go test ./...          → exit 0, 0 FAIL (full unit suite green)
GOFLAGS=-mod=mod go run ./tools/cilint  → exit 0
```

---

## 7. Terminal Acceptance Verdict

**VERDICT: FAIL** — Contract / API does not clear A− (3 skeptic-confirmed Majors); H-D fails by §8 intent.

| Check | Requirement | Result |
|---|---|---|
| 1 | module-boundaries, contract-api, composition all ≥ A− | **FAIL** — Contract/API B+ (module-boundaries A−, composition A− both PASS) |
| 2 | 0 skeptic-confirmed Critical/Major | **FAIL** — 3 confirmed Majors (all Contract/API, `documents` handler) |
| 3 | H-D class = 0 (widened scope) | **FAIL by intent** — mechanical gate = 0, but `map[string]any`-only analyzer masks 3 `map[string]string` response literals |
| 4 | H-G class = 0 (full scope) | **PASS** — honest H-G = 0 (F8.3 closed the last reach) |

M8's declared work is real and verified: F8.1 (presence typed), F8.2 (metrics typed envelope), F8.3
(taxonomy `FamilyCodeResolver` port — closes H-G), F8.4 (CWE-613 deactivation, lifts Sessions to A−), F8.5
(problem+json 404/405 interceptor), F8.6 (widened §5b gate + `noresponsemap`). Seven of ten dimensions are at or
above A−, H-G is genuinely 0, and the two post-M7 regressions (Sessions, Composition) are recovered. The
terminal gate nonetheless fails on a single dimension: **Contract / API (B+)**, held below A− by three untyped
`map[string]string` response literals in `documents/delivery/http/handler.go` that the new `noresponsemap`
analyzer does not catch because it is scoped to `map[string]any` only.

---

## 8. Root Cause

The repeat Contract/API miss is structural, and this round exposes the next layer of the same pattern. M8's
F8.6 closed the **path-scope** blindspot (widened the gate to the full route surface) and added a mechanical
analyzer — but the analyzer's **type scope** is `map[string]any` only. The `documents` handler emits its
untyped bodies as `map[string]string`, one type-width outside the guard. Each milestone has closed the exact
instances its gate greps for, and each subsequent independent read finds equivalent untyped-body sites just
outside the gate's *current* scope — first the path scope (post-M7: `iam/presence`, `observability`), now the
type scope (post-M8: `map[string]string` in `documents`). Until the contract is enforced positively (every
public handler routed through the generated `StrictServerInterface` typed responses, so an untyped body cannot
compile) rather than negatively (grep/analyze for the specific untyped shapes seen so far), the dimension will
keep missing on the next shape the gate does not yet name.

---

## 9. HS-5 Hard-Stop (5th consecutive Contract/API miss)

Contract / API has now missed the A− bar a **fifth** consecutive time (B+ → B− → B → B+ → B+). Per the mission
HS-5 rule and the operator's 5th-miss-closure directive, this audit **does not** open M9 and **does not** start
any remediation milestone. The decision is the operator's. Bounded options:

- **(A) Targeted M9 micro-milestone (small, well-scoped):** fix exactly the 3 confirmed `documents`-handler
  Majors — route `duplicateDocument`, the comment endpoints through their generated types; correct
  `signedRevisionURL` to the spec-declared 302 (or amend the spec to the implemented 200+body if 200 is the
  intended contract). Regen BE/FE codegen. **Also widen `noresponsemap` to flag any `map[string]<T>` response
  literal** (not just `map[string]any`) so this exact evasion cannot recur. Est. small: one handler file + the
  analyzer + codegen regen. This is the bounded close of the *known* instances + one-type-wider gate.
- **(B) Structural close (closes the class, not the instances):** route every public handler through the
  generated `StrictServerInterface` typed-response path so an untyped body is a compile error. Larger; the true
  root-cause fix per §8.
- **(C) Re-scope / accept:** if any of the 3 are intentionally off-spec, amend `openapi.yaml` + the §5b
  allowlist to declare the exemption explicitly, then re-measure. Converts a blindspot into a declared boundary.

Grade-A terminal sign-off is **not** reached at HEAD `58dea742`. Grade A is the operator's declaration, not the
agent's. Per HS-5 5th-miss: **STOP — surface to operator; do not auto-continue.**
