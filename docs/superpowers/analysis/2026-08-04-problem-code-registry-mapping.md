# Problem-code registry: semantic family taxonomy + complete rename mapping

Date: 2026-08-04
Scope: closed `Code` type + registry in `internal/platform/problem`; hard rename of all 147
emitted problem codes; downstream generation (OpenAPI `Problem.code` enum, FE `errorMessages`,
wiki table); CI freshness gate.
Status: ANALYSIS — §1 families and §3 status rulings require operator ratification before any
code moves.

## 0. Facts verified against the tree (HEAD 71692e0b + uncommitted ADR 0088)

| Fact | Verified |
|---|---|
| `go run ./scripts/dump-error-codes.go` emits **147** codes | yes — ran it; file on disk had 144, i.e. **stale by exactly 3** |
| The 3 stale codes | `validation.route_stages_not_permitted`, `validation.approval_stage_required`, `validation.route_stage_required` (ADR 0087, commit 1d3f8db5) |
| `internal/platform/problem/codes.go` constant count | **47** declarations = 43 `Problem.code` + 4 `FieldError.code` (`FieldCode*`) |
| Dead/unreferenced constants | **11** — 7 Problem-level + 4 field-level (enumerated in §2.1) |
| `type Code string` at `codes.go:7` | confirmed; the doc comment "prevents arbitrary strings from being used as codes" is **false** — every raw literal in `controlleddocuments/delivery/http/routes.go` and `documents/delivery/http/fillin_handler.go` converts implicitly |
| Module-local code const blocks | `approval/http/errors.go:29-113` (**68** consts), `taxonomy/delivery/http/routes_profiles.go:26-33` (**8**), `tokens/delivery/http/handler.go:33-36` (**4**) |
| Pure-alias const block (no new strings) | `templates/delivery/http/errors.go:14-42` — 22 aliases onto catalog constants |
| Raw string-literal emit sites | `controlleddocuments/delivery/http/routes.go` (43 sites), `documents/delivery/http/fillin_handler.go` (11 sites) |
| Generator misses field-level codes | yes — `extract()` drops idents prefixed `FieldCode`; `tokens` emits `WithFieldError("name", CodeTokenImmutableField, …)` so field codes DO reach the wire but are outside the 147 |

**ADR 0088 working-tree collision check.** The 17 modified/untracked paths were cross-checked
against every emit site in §2. **Zero §2 rows live in a currently-modified file.**
`internal/modules/templates/delivery/http/errors.go` — the templates mapper — is *not* modified.
The only ADR-0088 error-surface changes are:

- `templates/domain/errors.go` **+`ErrBlankSourceUnavailable`** — deliberately unmapped, lands on
  `MapErr`'s `default:` 500 arm. No new code. Nothing to rename.
- `templates/application/create.go` now wraps `domain.ErrUploadMissing` for a blank-source with a
  short content hash. This **increases** the traffic through `codeTplUploadMissing` and therefore
  raises the stakes on ruling **R-1** (§3).

Sequencing consequence: the sweep does **not** need to wait on 0088 for file-level conflicts, but
ruling R-1 (`UPLOAD_MISSING` status) should be decided with 0088's author in the room. Rows whose
*semantics* 0088 touches are marked `⚠0088` in §2.

---

## 1. Semantic family taxonomy

### 1.1 Dotted prefixes currently in use (inventory)

Extracted from all 147 codes. `n` = distinct codes carrying the prefix.

| Prefix | n | Nature | Verdict |
|---|---|---|---|
| `validation.` | 30 | mixed — 400 syntax/param failures AND 422 business-rule rejections under one prefix | SPLIT (see 1.2) |
| `state.` | 10 | subject lifecycle | KEEP |
| `not_found.` | 6 | 404 | RENAME → `notfound.` |
| `internal.` | 5 | 500 | KEEP |
| `conflict.` | 3 | races | KEEP |
| `precondition.` | 2 | If-Match / content-hash | KEEP |
| `idempotency.` | 2 | 400 required + 409 conflict — two different families under one prefix | SPLIT |
| `route.` | 2 | **module-named** (approval route entity) | RE-HOME |
| `sod.` | 2 | **module-named** (segregation-of-duties) | RE-HOME |
| `signoff.` | 2 | **module-named** | RE-HOME |
| `template.` | 2 | **module-named** | RE-HOME |
| `authn.` | 2 | 401 + 429 under one prefix | SPLIT |
| `authz.` | 1 | 403 | RENAME → `permission.` |
| `approval.` | 1 | **module-named** | RE-HOME |
| `submit.` | 1 | **module-named** (approval submit verb) | RE-HOME |
| `freeze.` | 1 | **module-named** (documents freeze) | RE-HOME |
| `creation_context.` | 1 | **module-named** (CD creation context) | RE-HOME |
| bare lower_snake | 5 | `timeout`, `immutable_field`, `reserved_name`, `not_a_choice_placeholder`, `template_invalid` | RE-HOME |
| SCREAMING_SNAKE | 67 | no family at all | RE-HOME |

### 1.2 The closed set (10 families)

The operator's starting axis is adopted with **two refinements**, both justified below.

| Family | Meaning (one line) | Default status |
|---|---|---|
| `request.` | The HTTP request itself is malformed, unparseable, or protocol-unacceptable — fix the *syntax/shape* and retry. | **400** |
| `validation.` | The request parses fine; a *value* or *business rule* rejects it, independent of stored state. | **422** |
| `auth.` | The caller's identity is missing, unproven, or blocked. | **401** |
| `permission.` | The caller is identified but lacks the required capability in the required area. | **403** |
| `notfound.` | The addressed subject does not exist, or is outside this caller's visibility boundary. | **404** |
| `state.` | The subject exists but is in the wrong lifecycle state for this operation; retrying is futile until the state changes. | **409** |
| `conflict.` | A race with another writer, a uniqueness collision, or an idempotency-key collision; retrying against fresh state may succeed. | **409** |
| `precondition.` | A precondition the **caller supplied** (If-Match, expected content hash, lock version) no longer holds. | **412** |
| `ratelimit.` | The caller is being throttled. | **429** |
| `internal.` | Server fault: bug, deployment misconfiguration, unimplemented surface, or upstream/dependency failure. | **500** |

**Refinement 1 — `request.` split out of `validation.` (RECOMMENDED, needs ratification).**
Today 30 `validation.*` codes span 400, 413, 415, and 422 with no way for a client to tell "your
JSON is broken" from "your input is business-rejected". RFC 9110 §15.5.1 reserves 400 for requests
the server *cannot process due to client error in the request itself*; 422 (RFC 9110 §15.5.21) is
for a well-formed request whose *content* is semantically wrong. Those are different client
remedies — regenerate the request vs. change the data — and a wire contract that erases the
distinction forces the FE to keep a hand-maintained side-table. `request.` also gives 405, 410
(expired cursor), 413, and 415 a coherent home instead of scattering them.
Trade-off: ~20 rows in §2 move to `request.` instead of `validation.`, and the boundary needs the
decision rule in §1.4 to stay stable.

**Refinement 2 — no separate `config.` family (RECOMMENDED).** `creation_context.unconfigured`,
`template.artifact_invariant_unconfigured`, and `internal.signature_misconfigured` are all
"deployment is broken". They are indistinguishable to a client (all 500, all "not your fault"), so
a fourth 5xx family buys nothing on the wire. They fold into `internal.` with a
`misconfigured_*` / `_unconfigured` name segment that keeps them greppable server-side.

**Explicitly rejected:** module-named families (`approval.`, `template.`, `documents.`). Ratified
by the operator; the ADR 0082 extraction of `approval` into a top-level module is the proof — a
wire code that names a Go module becomes a lie the moment the module boundary moves.

### 1.3 Re-homing every module-named / bare prefix

| Old prefix or bare code | New family | Reason |
|---|---|---|
| `approval.unresolved_comments` | `state.` | The document is in a state (open comments) that blocks the write. |
| `signoff.duplicate` | `conflict.` | Uniqueness collision on (actor, stage). |
| `signoff.not_eligible` | `permission.` | Actor is not in the eligible pool → capability-shaped denial, 403. |
| `sod.submitter_cannot_sign` | `permission.` | 403 denial derived from separation-of-duties policy. |
| `sod.cross_stage_duplicate` | `permission.` | Same — a 403 policy denial, not a race. |
| `route.in_use` | `state.` | The route entity is in a state (referenced) that blocks deactivation. |
| `route.duplicate_profile` | `conflict.` | Uniqueness collision on (tenant, profile_code). |
| `submit.invalid_supersede_target` | `validation.` | The supplied target id is an unacceptable value. |
| `freeze.effective_date_missing` | `validation.` | A required field is absent at submit time. |
| `template.artifact_missing` | `state.` | The template version has no committed artifact yet. |
| `template.artifact_invariant_unconfigured` | `internal.` | Deployment misconfiguration. |
| `creation_context.unconfigured` | `internal.` | Deployment misconfiguration. |
| `authn.signature_invalid` | `auth.` | Credential proof failed. |
| `authn.rate_limited` | `ratelimit.` | Throttling, not identity. |
| `authz.capability_denied` | `permission.` | Direct rename of the family. |
| `not_found.*` | `notfound.` | Family-name normalisation (no underscore inside the family token). |
| `idempotency.key_required` | `request.` | A required header is absent — request shape. |
| `idempotency.key_conflict` | `conflict.` | Key reused with a different fingerprint. |
| `timeout` | `internal.` | Upstream/dependency deadline; 504. |
| `immutable_field` | `validation.` | The supplied value violates an immutability rule. |
| `reserved_name` | `validation.` | The supplied value collides with a reserved namespace. |
| `not_a_choice_placeholder` | `validation.` | The addressed placeholder is the wrong kind for this payload. |
| `template_invalid` | `validation.` | Template↔profile mismatch on supplied input (see collision C-9). |
| all 67 SCREAMING_SNAKE | per-row, §2 | — |

### 1.4 The rule that picks a family for a NEW code

Evaluate in order; **first match wins**. Write this into the registry's package doc and into
`wiki/architecture/api-contract.md`.

1. Could the caller have avoided it by sending a **syntactically or structurally different**
   request (bad JSON, missing/duplicate param, wrong content-type, oversized body, wrong method,
   missing required header)? → **`request.`**
2. Does the request parse, but a **supplied value or a business rule over supplied values** rejects
   it, with no reference to stored subject state? → **`validation.`**
3. Is the problem **who the caller is** — no session, bad credentials, blocked/inactive account,
   missing tenant claim? → **`auth.`**
4. Is the caller identified but **lacks a capability** (tier-1, tier-2, eligibility pool, SoD, ISO
   segregation, origin policy)? → **`permission.`**
5. Does the **addressed subject not exist**, or is it invisible to this caller? → **`notfound.`**
6. Did the **caller supply a precondition** (If-Match / ETag / expected content hash / lock
   version) that no longer holds? → **`precondition.`**
7. Is it a **race or a uniqueness/idempotency collision** where retrying against refreshed state
   could succeed? → **`conflict.`**
8. Is the **subject in the wrong lifecycle state** for this operation, such that retrying is futile
   until something else changes the state? → **`state.`**
9. Is the caller being **throttled**? → **`ratelimit.`**
10. Otherwise it is a **server fault** — bug, misconfiguration, unimplemented surface, upstream
    failure. → **`internal.`**

Tie-breaker between 7 and 8 (the two 409 families): if the failure would disappear on a *retry of
the same request* once a concurrent writer finishes, it is `conflict.`; if it requires a *different
operation* to change the subject first, it is `state.`.

Name segment after the family: `snake_case`, subject-first, no module name, no verb tense
(`state.document_not_published`, not `state.cannot_publish_document`).

---

## 2. Complete mapping table

### How to read this table (execution notes for the sweep agent)

- **Edit class** column tells you the mechanical shape:
  - `CONST` — the string lives in exactly one `const` declaration. Change the value there; all call
    sites follow automatically. Cite line is the const line.
  - `LITERAL` — the string is a raw literal at each listed site. Every listed `file:line` must be
    edited individually.
  - `DELETE` — remove the declaration; there are no emit sites.
- **NEW default status** is what the code gets in `Register(...)`. A call site whose current status
  differs must either be changed to the default, or keep its status as an explicit documented
  override — the `note` column says which. Where the answer is a ruling, it points at §3.
- `⚠0088` marks rows whose *semantics* the concurrent ADR 0088 work touches. No row's *file* is
  currently modified.
- `COLLAPSE→X` in the note means this old code and the referenced other code(s) MUST become the
  same new code. Collisions are indexed C-1…C-14 and repeated in §2.9.

---

### 2.1 `internal/platform/problem` — canonical catalog (`codes.go`)

All rows are `CONST` (single declaration in `codes.go`) unless noted. Line numbers are `codes.go`
declaration lines. Call-site counts are non-test references across `internal/` + `apps/`.

| # | old code | decl | emit sites | current status(es) | NEW code | NEW default | note |
|---|---|---|---|---|---|---|---|
| 1 | `VALIDATION_ERROR` | :12 | 101 refs across audit, auth, controlleddocuments, documents, iam, search, taxonomy, tokens, idempotency, templates(alias) | 400 (95 sites), 422 (6 sites) | `request.invalid` | 400 | **C-1.** The 6 sites currently at 422 are semantic rejections and must move to a `validation.*` code, not carry an override. Per-site list in §3 R-6. |
| 2 | `UNKNOWN_FIELD` | :13 | none | — | **DELETE** | — | dead |
| 3 | `UNKNOWN_FILTER` | :14 | none | — | **DELETE** | — | dead |
| 4 | `INVALID_SORT_FIELD` | :15 | none | — | **DELETE** | — | dead |
| 5 | `INVALID_CURSOR` | :16 | `distribution/delivery/http/handler.go:104`, `notifications/delivery/http/handler.go:80` | 400 | `request.cursor_invalid` | 400 | |
| 6 | `INCLUDE_NOT_SUPPORTED` | :17 | none | — | **DELETE** | — | dead |
| 7 | `UNAUTHENTICATED` | :18 | `documents/delivery/http/pdf_webhook_handler.go:74`; **LITERAL** `controlleddocuments/delivery/http/routes.go:551` | 401 | `auth.unauthenticated` | 401 | **C-2.** COLLAPSE→ with `AUTH_UNAUTHORIZED` (#26). CD site is a raw literal — edit in place. |
| 8 | `FORBIDDEN_CAPABILITY` | :19 | `controlleddocuments/…/routes.go:510`, `documents/…/handler.go:1322,1324`, `iam/…/tenant_handler.go:124,201`, `templates/…/errors.go:74,139,141` | 403 | `permission.capability_denied` | 403 | **C-3.** COLLAPSE→ with `authz.capability_denied` (#125). |
| 9 | `FORBIDDEN_AREA` | :20 | none | — | **DELETE** | — | dead |
| 10 | `FORBIDDEN_ORIGIN` | :21 | `platform/security/cors.go:64`, `platform/security/origin_protection.go:155` | 403 | `permission.origin_forbidden` | 403 | |
| 11 | `NOT_FOUND` | :22 | 27 refs: audit, distribution, documents, iam, notifications, tokens, `platform/middleware/method_not_allowed.go:34` | 404 | `notfound.resource` | 404 | **C-4.** `tokens` redeclares this string (#146) — delete the redeclaration, use the catalog code. |
| 12 | `METHOD_NOT_ALLOWED` | :23 | `audit/…/handler.go:137,192,284`, `platform/httpresponse/response.go:24`, `platform/middleware/method_not_allowed.go:32` | 405 | `request.method_not_allowed` | 405 | |
| 13 | `ALREADY_EXISTS` | :24 | `templates/…/errors.go:16` (alias `codeTplKeyConflict`) | 409 | `conflict.already_exists` | 409 | **C-4.** `tokens` redeclares this string (#145). |
| 14 | `STATE_TRANSITION_INVALID` | :25 | `documents/…/handler.go:1343`, `templates/…/errors.go:17,26` (aliases `codeTplInvalidStateTransition`, `codeTplArchived`) | 409 | `state.transition_invalid` | 409 | |
| 15 | `CONCURRENT_MODIFICATION` | :26 | `documents/…/handler.go:1340`, `templates/…/errors.go:18` (alias `codeTplStaleLockVersion`) | 409 (documents), **412** (templates) | `conflict.concurrent_modification` | 409 | **R-7.** The templates 412 use is a *lock-version precondition*, not a race — split it out to `precondition.lock_version_stale` (new code, #150). |
| 16 | `IDEMPOTENCY_KEY_REUSED` | :27 | `platform/idempotency/middleware.go:146` | **422** | `conflict.idempotency_key_reused` | 409 | **R-8.** 422 is wrong for a key/fingerprint collision. |
| 17 | `IDEMPOTENCY_KEY_INVALID` | :28 | `platform/idempotency/middleware.go:126`, `approval/http/errors.go` (idempotency.ErrKeyInvalid arm) | 400 | `request.idempotency_key_invalid` | 400 | |
| 18 | `IDEMPOTENCY_REPLAY` | :29 | none | — | **DELETE** | — | dead |
| 19 | `REQUEST_BODY_TOO_LARGE` | :30 | `documents/…/handler.go:1330`, `templates/…/errors.go:20` (alias), `platform/idempotency/middleware.go:137` | 413 | `request.body_too_large` | 413 | **C-5.** COLLAPSE→ with `validation.body_too_large` (#93). |
| 20 | `RATE_LIMITED` | :31 | `platform/ratelimit/middleware.go:219` | 429 | `ratelimit.exceeded` | 429 | **C-6.** COLLAPSE→ with `authn.rate_limited` (#56). |
| 21 | `INTERNAL_ERROR` | :32 | 141 refs across audit, auth, controlleddocuments, distribution, documents, iam, notifications, search, security, taxonomy, tokens, `platform/middleware/recovery.go:46`, `templates/…/errors.go:29` (alias) | 500; **501** at ~25 iam/security sites; **502** at `documents/…/export_handler.go:139` | `internal.unknown` | 500 | **C-7.** COLLAPSE→ with `internal.unknown` (#79). **R-9**: the 501 sites must emit `internal.not_implemented`; the 502 site `internal.upstream_failed`. |
| 22 | `NOT_IMPLEMENTED` | :33 | `audit/…/handler.go:200,253,265,288`, `iam/…/people_handler.go:308`, `iam/…/router.go:326` | 501 | `internal.not_implemented` | 501 | absorbs the R-9 sites |
| 23 | `CURSOR_EXPIRED` | :34 | `iam/…/people_handler.go:97` | 410 | `request.cursor_expired` | 410 | |
| 24 | `CONFLICT_ERROR` | :35 | `documents/…/export_handler.go:137`, `documents/…/handler.go:1338`, `iam/…/people_handler.go:433`, `iam/…/tenant_handler.go:120,122,197`, `templates/…/errors.go:19,41` (aliases `codeTplContentHashMismatch`, `codeTplApprovalConflict`) | 409; **422** via `codeTplApprovalConflict` at `templates/…/errors.go:128,133,137` | `conflict.generic` | 409 | Catch-all. **R-10**: the four templates 422 uses are distinct conditions and must move to specific codes (see #106,#116,#117,#118). |
| 25 | `IDEMPOTENCY_KEY_REQUIRED` | :36 | `controlleddocuments/…/handler.go:150`, `templates/…/errors.go:124` | 400 | `request.idempotency_key_required` | 400 | **C-8.** COLLAPSE→ with `idempotency.key_required` (#53). |
| 26 | `AUTH_UNAUTHORIZED` | :38 | 17 refs: audit, auth, iam, search, security | 401 | `auth.unauthenticated` | 401 | **C-2.** COLLAPSE→ #7. |
| 27 | `AUTH_INVALID_CREDENTIALS` | :39 | `auth/…/handler.go:196,198` | 401 | `auth.invalid_credentials` | 401 | |
| 28 | `AUTH_ACCOUNT_LOCKED` | :40 | `auth/…/handler.go:200` | 403 | `auth.account_locked` | 403 | family default 401 overridden — identity is proven, account is blocked (RFC 9110 §15.5.4) |
| 29 | `AUTH_ACCOUNT_INACTIVE` | :41 | `auth/…/handler.go:204` | 403 | `auth.account_inactive` | 403 | as above |
| 30 | `AUTH_TENANT_FORBIDDEN` | :42 | `auth/…/handler.go:206` | 403 | `auth.tenant_forbidden` | 403 | as above |
| 31 | `AUTH_TENANT_REQUIRED` | :43 | `auth/…/handler.go:208` | 403 | `auth.tenant_required` | 403 | as above |
| 32 | `AUTH_PASSWORD_CHANGE_REQUIRED` | :44 | `auth/…/middleware.go:95` | 403 | `auth.password_change_required` | 403 | as above |
| 33 | `AUTH_FORBIDDEN` | :45 | 11 refs: `audit/…/handler.go:154,259,496`, `documents/…/handler.go:1305`, `iam/…/middleware.go:142,170`, `iam/…/routes_memberships.go:139,152,216,317`, `templates/…/errors.go:24` (alias) | 403 | `permission.denied` | 403 | mis-prefixed today: it is a capability/ownership denial, not an identity failure |
| 34 | `APPROVAL_ROUTE_MISSING` | :50 | `templates/…/errors.go:40` (alias `codeTplApprovalRouteMissing`) | 409 | `state.approval_route_missing` | 409 | **C-9.** COLLAPSE→ with `state.approval_route_missing` (#86) and the CD literal (#132). |
| 35 | `UPLOAD_EXPIRED` | :51 | `documents/…/handler.go:1326` | 410 | `state.upload_expired` | 410 | |
| 36 | `UPLOAD_MISSING` | :52 | `documents/…/handler.go:1328`; `templates/…/errors.go:21` (alias) used at `errors.go:66` and the `ErrTemplateVersionNoContent` arm | **410** (documents) / **409** (templates) | `state.upload_missing` | **409** (R-1) | ⚠0088 — `templates/application/create.go` now routes blank-source hash failures here |
| 37 | `STALE_BASE` | :53 | `templates/…/errors.go:17` (alias `codeTplStaleBase`) | 409 | `conflict.stale_base` | 409 | **R-11**: documents maps the same `ErrStaleBase` sentinel to `CONCURRENT_MODIFICATION` (`handler.go:1339-1340`). |
| 38 | `ISO_SEGREGATION_VIOLATION` | :54 | `templates/…/errors.go:23` (alias) | 403 | `permission.iso_segregation_violation` | 403 | |
| 39 | `SYSTEM_TEMPLATE_IMMUTABLE` | :55 | `templates/…/errors.go:25` (alias) | 409 | `state.system_template_immutable` | 409 | |
| 40 | `PRECONDITION_REQUIRED` | :61 | none (handler removed by ADR 0073) | — | **DELETE** | — | dead. Its live counterpart `precondition.if_match_required` (#51) survives. |
| 41 | `MEMBERSHIP_EXISTS` | :68 | `iam/…/routes_memberships.go:319` | 409 | `conflict.membership_exists` | 409 | |
| 42 | `MEMBERSHIP_NOT_FOUND` | :69 | `iam/…/routes_memberships.go:321` | 404 | `notfound.membership` | 404 | |
| 43 | `UNKNOWN_ROLE` | :70 | `iam/…/people_handler.go:431`, `iam/…/routes_memberships.go:323` | 400 | `validation.role_unknown` | **422** (R-12) | supplied value rejected by a business rule, not a syntax error |
| 44 | `REQUIRED` (FieldCode) | :74 | none | — | **DELETE** | — | dead field-level |
| 45 | `INVALID_FORMAT` (FieldCode) | :75 | none | — | **DELETE** | — | dead field-level |
| 46 | `OUT_OF_RANGE` (FieldCode) | :76 | none | — | **DELETE** | — | dead field-level |
| 47 | `INVALID_ENUM` (FieldCode) | :77 | none | — | **DELETE** | — | dead field-level |

**Dead-constant tally (11): #2, #3, #4, #6, #9, #18, #40, #44, #45, #46, #47.**

> Field-level codes are a **separate namespace** from `Problem.code` — `FieldError.code` describes a
> field, not the response. The registry must keep them apart (see §4.2). The only live field code
> is `immutable_field` / `reserved_name` via `tokens` (#147, #148), which today reuse the
> `Problem.code` type. Recommendation: register them in a `field.` namespace with no HTTP status.

---

### 2.2 `internal/modules/approval` — 68 module-local codes

All rows are `CONST` in `internal/modules/approval/http/errors.go`; the value is declared once and
every emit is via `MapErrorToResponse` (`errors.go:145-455`). Rename the const **value** only; the
Go identifier can stay. Statuses shown are the ones assigned in the switch.

| # | old code | decl line | status | NEW code | NEW default | note |
|---|---|---|---|---|---|---|
| 48 | `internal.unknown` | :29 | 500 | `internal.unknown` | 500 | **C-7.** unchanged string; becomes the single registered 500 code |
| 49 | `conflict.stale_revision` | :30 | 409 | `conflict.stale_revision` | 409 | unchanged |
| 50 | `not_found.instance` | :31 | 404 | `notfound.approval_instance` | 404 | |
| 51 | `not_found.instance_not_visible` | :39 | 404 | `notfound.approval_instance_not_visible` | 404 | keep distinct — deliberate log/monitoring split |
| 52 | `conflict.duplicate_submission` | :40 | 409 | `conflict.duplicate_submission` | 409 | unchanged. **C-10** COLLAPSE→ templates `codeTplApprovalConflict` on `ErrDuplicateSubmission` (`templates/…/errors.go:113`) |
| 53 | `signoff.duplicate` | :41 | 409 | `conflict.signoff_duplicate` | 409 | re-home |
| 54 | `submit.invalid_supersede_target` | :42 | 409 | `validation.supersede_target_invalid` | **422** (R-13) | re-home + status ruling |
| 55 | `state.instance_completed` | :43 | 409 | `state.approval_instance_completed` | 409 | also the target of `domain.ErrNoActiveStage` (`errors.go:277`) — see R-14 |
| 56 | `route.in_use` | :44 | 409 | `state.approval_route_in_use` | 409 | re-home |
| 57 | `route.duplicate_profile` | :45 | 409 | `conflict.approval_route_duplicate_profile` | 409 | re-home |
| 58 | `signoff.not_eligible` | :46 | 403 | `permission.signoff_actor_not_eligible` | 403 | re-home. **C-11** COLLAPSE→ templates `FORBIDDEN_CAPABILITY` on `ErrActorNotEligible` |
| 59 | `sod.submitter_cannot_sign` | :47 | 403 | `permission.sod_submitter_cannot_sign` | 403 | re-home. **C-12** COLLAPSE→ templates `FORBIDDEN_CAPABILITY` on `ErrAuthorCannotSign` |
| 60 | `sod.cross_stage_duplicate` | :48 | 403 | `permission.sod_cross_stage_duplicate` | 403 | re-home |
| 61 | `freeze.effective_date_missing` | :49 | 422 | `validation.effective_date_required` | 422 | re-home |
| 62 | `precondition.if_match_required` | :50 | 428 | `precondition.if_match_required` | **428** | unchanged string; explicit non-default status registered |
| 63 | `validation.if_match_malformed` | :51 | 400 | `request.if_match_malformed` | 400 | malformed header = request shape |
| 64 | `idempotency.key_required` | :52 | 400 | `request.idempotency_key_required` | 400 | **C-8.** COLLAPSE→ #25 |
| 65 | `idempotency.key_conflict` | :53 | 409 | `conflict.idempotency_key_reused` | 409 | **C-13.** COLLAPSE→ #16 |
| 66 | `precondition.content_hash_mismatch` | :54 | 412 | `precondition.content_hash_mismatch` | 412 | unchanged. **C-14** COLLAPSE→ documents (#-, `handler.go:1331`) and templates (`errors.go:63,119`) — R-2 |
| 67 | `authn.signature_invalid` | :55 | 401 | `auth.signature_invalid` | 401 | re-home |
| 68 | `authn.rate_limited` | :56 | 429 | `ratelimit.exceeded` | 429 | **C-6.** COLLAPSE→ #20 |
| 69 | `internal.db_privilege_missing` | :57 | 500 | `internal.db_privilege_missing` | 500 | unchanged |
| 70 | `internal.db_unknown` | :58 | 500 | `internal.db_unknown` | 500 | unchanged |
| 71 | `internal.signature_misconfigured` | :59 | 500 | `internal.signature_misconfigured` | 500 | unchanged |
| 72 | `validation.param_format` | :60 | 400 | `request.param_format` | 400 | re-home to `request.` |
| 73 | `validation.param_unmarshal` | :61 | 400 | `request.param_unmarshal` | 400 | re-home |
| 74 | `validation.param_required` | :62 | 400 | `request.param_required` | 400 | re-home |
| 75 | `validation.header_required` | :63 | 400 | `request.header_required` | 400 | re-home |
| 76 | `validation.param_too_many_values` | :64 | 400 | `request.param_too_many_values` | 400 | re-home |
| 77 | `authz.capability_denied` | :65 | 403 | `permission.capability_denied` | 403 | **C-3.** COLLAPSE→ #8 and #125 |
| 78 | `approval.unresolved_comments` | :66 | 409 | `state.approval_blocked_unresolved_comments` | 409 | re-home |
| 79 | `validation.reason_required` | :67 | 400 | `validation.reason_required` | **422** (R-15) | value-level rule, not syntax |
| 80 | `not_found.route` | :68 | 404 | `notfound.approval_route` | 404 | |
| 81 | `state.route_inactive` | :69 | 409 | `state.approval_route_inactive` | 409 | |
| 82 | `timeout` | :70 | 504 | `internal.upstream_timeout` | 504 | re-home from bare |
| 83 | `validation.json_decode` | :71 | 400 | `request.json_decode` | 400 | re-home. COLLAPSE→ fillin `validation.json_decode` (#129) |
| 84 | `validation.json_type_error` | :72 | 400 | `request.json_type_error` | 400 | re-home |
| 85 | `validation.empty_body` | :73 | 400 | `request.empty_body` | 400 | re-home. COLLAPSE→ fillin `validation.empty_body` (#128) |
| 86 | `validation.content_type` | :74 | 415 | `request.content_type_unsupported` | 415 | re-home. COLLAPSE→ fillin `validation.bad_content_type` (#127) |
| 87 | `validation.body_too_large` | :75 | 413 | `request.body_too_large` | 413 | **C-5.** COLLAPSE→ #19 |
| 88 | `validation.request_invalid` | :76 | 400 | `request.invalid` | 400 | **C-1.** COLLAPSE→ #1 |
| 89 | `validation.profile_unknown` | :77 | 422 | `validation.profile_unknown` | 422 | unchanged |
| 90 | `validation.reason_for_change_required` | :78 | 422 | `validation.reason_for_change_required` | 422 | unchanged |
| 91 | `validation.reason_category_invalid` | :79 | 422 | `validation.reason_category_invalid` | 422 | unchanged |
| 92 | `validation.revision_title_required` | :80 | 422 | `validation.revision_title_required` | 422 | **R-3**: documents maps the same `approvalapp.ErrRevisionTitleRequired` to 400 `VALIDATION_ERROR` (`documents/…/handler.go:1315`) |
| 93 | `validation.document_subject_key_mismatch` | :81 | 422 | `validation.document_subject_key_mismatch` | 422 | unchanged |
| 94 | `validation.template_subject_key_mismatch` | :82 | 422 | `validation.template_subject_key_mismatch` | 422 | unchanged |
| 95 | `state.document_not_draft` | :83 | 409 | `state.document_not_draft` | 409 | unchanged |
| 96 | `validation.profile_not_configured` | :84 | 400 | `validation.profile_not_configured` | **422** (R-16) | |
| 97 | `state.approval_route_missing` | :85 | 409 | `state.approval_route_missing` | 409 | **C-9.** absorbs #34 and #132 |
| 98 | `not_found.document` | :86 | 404 | `notfound.document` | 404 | COLLAPSE→ documents `NOT_FOUND` for `ErrNotFound` remains generic; keep this specific one |
| 99 | `state.document_not_published` | :87 | 409 | `state.document_not_published` | 409 | unchanged |
| 100 | `conflict.mark_reviewed_stale_revision` | :88 | 409 | `conflict.mark_reviewed_stale_revision` | 409 | unchanged |
| 101 | `validation.review_due_before_effective` | :89 | 422 | `validation.review_due_before_effective` | 422 | unchanged |
| 102 | `validation.effective_to_not_after_effective_from` | :90 | 422 | `validation.effective_to_not_after_effective_from` | 422 | unchanged |
| 103 | `validation.empty_eligible_pool` | :91 | 422 | `validation.empty_eligible_pool` | 422 | **R-4**: templates maps the same `approvaldomain.ErrEmptyEligiblePool` to 422 `CONFLICT_ERROR` (`templates/…/errors.go:127-128`) |
| 104 | `validation.submit_choice_required` | :92 | 422 | `validation.submit_choice_required` | 422 | **R-4**: templates → 422 `CONFLICT_ERROR` (`errors.go:132-133`) |
| 105 | `validation.submit_choice_constraint_violated` | :93 | 422 | `validation.submit_choice_constraint_violated` | 422 | **R-4**: templates → 422 `CONFLICT_ERROR` (`errors.go:136-137`) |
| 106 | `validation.route_stages_not_permitted` | :98 | 422 | `validation.route_stages_not_permitted` | 422 | **stale in generated JSON** |
| 107 | `validation.approval_stage_required` | :99 | 422 | `validation.approval_stage_required` | 422 | **stale in generated JSON** |
| 108 | `validation.route_stage_required` | :100 | 422 | `validation.route_stage_required` | 422 | **stale in generated JSON** |
| 109 | `validation.self_delegation` | :103 | 422 | `validation.self_delegation` | 422 | unchanged |
| 110 | `validation.delegation_window_invalid` | :104 | 422 | `validation.delegation_window_invalid` | 422 | unchanged |
| 111 | `not_found.delegation` | :105 | 404 | `notfound.delegation` | 404 | |
| 112 | `state.verdict_ready_on_approval_stage` | :108 | 422 | `validation.verdict_ready_on_approval_stage` | 422 | **R-17**: prefix says `state.` but status is 422 and the condition is "your supplied verdict is wrong for this stage kind" → `validation.` |
| 113 | `internal.verdict_wrong_stage_kind` | :109 | 500 | `internal.verdict_wrong_stage_kind` | 500 | unchanged |
| 114 | `state.fast_forward_stage_not_completed` | :112 | 409 | `state.fast_forward_stage_not_completed` | 409 | unchanged |
| 115 | `state.fast_forward_not_eligible` | :113 | 409 | `state.fast_forward_not_eligible` | 409 | unchanged |

> Approval declares **68** consts; rows #48-#115 cover them all (68 rows).

---

### 2.3 `internal/modules/templates` — aliases only, zero own strings

`templates/delivery/http/errors.go:14-42` declares **22 aliases** onto catalog constants. They
introduce **no new wire strings**, so they are not separate table rows — but each alias must be
re-pointed at the new registry code, and several are *wrong bindings* that §3 must fix.

| alias | decl | binds to | used at | issue |
|---|---|---|---|---|
| `codeTplNotFound` | :15 | `CodeNotFound` | :54 | ok |
| `codeTplKeyConflict` | :16 | `CodeAlreadyExists` | :56 | ok |
| `codeTplInvalidStateTransition` | :17 | `CodeStateTransitionInvalid` | :58, :62 | ok |
| `codeTplStaleBase` | :18 | `CodeStaleBase` | :60 | **R-11** |
| `codeTplStaleLockVersion` | :19 | `CodeConcurrentModification` | :64 (**412**) | **R-7** — code name ≠ status class |
| `codeTplContentHashMismatch` | :20 | `CodeConflict` (409) | :66, :119 | **R-2** — approval says 412 `precondition.content_hash_mismatch` |
| `codeTplUploadMissing` | :21 | `CodeUploadMissing` | :68, :107 (409) | **R-1** ⚠0088 |
| `codeTplUploadTooLarge` | :22 | `CodeRequestBodyTooLarge` | :70 | ok |
| `codeTplISOSegregation` | :23 | `CodeISOSegregationViolation` | :72 | ok |
| `codeTplForbidden` | :24 | `CodeAuthForbidden` | :76 | → `permission.denied` |
| `codeTplSystemImmutable` | :25 | `CodeSystemTemplateImmutable` | :78 | ok |
| `codeTplArchived` | :26 | `CodeStateTransitionInvalid` | :80 | ok |
| `codeTplPlaceholderNameInvalid` | :27 | `CodeValidationError` (422) | :82, :84, :86 | **R-6** — 422 site of `VALIDATION_ERROR`; must become `validation.placeholder_name_invalid` |
| `codeTplDuplicatePlaceholder` | :28 | `CodeAlreadyExists` | :88 (422) | **R-18** — 422 carrying `ALREADY_EXISTS`; recommend `validation.placeholder_name_duplicate` 422 |
| `codeTplInternalError` | :29 | `CodeInternalError` | :143 | ok |
| `codeTplInvalidRequest` | :30 | `CodeValidationError` (422) | :92 | **R-6** → `validation.doc_type_code_required` |
| `codeTplInvalidBody` | :31 | `CodeValidationError` | unused in errors.go | dead alias — DELETE if no other referrer |
| `codeTplInvalidLimit` | :32 | `CodeValidationError` | unused in errors.go | dead alias — DELETE if no other referrer |
| `codeTplInvalidParam` | :33 | `CodeValidationError` | unused in errors.go | dead alias — DELETE if no other referrer |
| `codeTplApprovalRouteMissing` | :40 | `CodeApprovalRouteMissing` | :96, :110 | **C-9** |
| `codeTplApprovalConflict` | :41 | `CodeConflict` | :101,:113,:115,:117,:128,:133,:137,:139 | **R-10/R-4** — 4 of these are 422 |
| `codeTplApprovalNotFound` | :42 | `CodeNotFound` | :99, :114 | ok |

Additional templates bindings that are **not** aliases and must be re-homed:
`problem.CodeForbiddenCapability` at `errors.go:74, 139, 141` (the last two are **C-11/C-12**
collapses), `problem.CodeIdempotencyKeyRequired` at `errors.go:124` (**C-8**).

---

### 2.4 `internal/modules/documents` — fill-in taxonomy (`fillin_handler.go`)

All rows are **LITERAL** — each string appears inline. Emit sites are exact.

| # | old code | emit sites | status | NEW code | NEW default | note |
|---|---|---|---|---|---|---|
| 116 | `authz.capability_denied` | `fillin_handler.go:134` | 403 | `permission.capability_denied` | 403 | **C-3.** COLLAPSE→ #8/#77 |
| 117 | `not_a_choice_placeholder` | `fillin_handler.go:136` | 400 | `validation.placeholder_not_choice` | **422** (R-19) | re-home from bare |
| 118 | `not_found.revision` | `fillin_handler.go:138` | 404 | `notfound.revision` | 404 | |
| 119 | `state.placeholder_not_author_editable` | `fillin_handler.go:140` | 409 | `state.placeholder_not_author_editable` | 409 | unchanged |
| 120 | `state.revision_not_draft` | `fillin_handler.go:142` | 409 | `state.revision_not_draft` | 409 | unchanged |
| 121 | `validation.failed` | `fillin_handler.go:144` | 422 | `validation.failed` | 422 | unchanged; becomes the canonical generic 422 |
| 122 | `validation.empty_body` | `fillin_handler.go:146` | 400 | `request.empty_body` | 400 | COLLAPSE→ #85 |
| 123 | `validation.bad_content_type` | `fillin_handler.go:148` | 415 | `request.content_type_unsupported` | 415 | COLLAPSE→ #86 |
| 124 | `validation.json_decode` | `fillin_handler.go:150` | 400 | `request.json_decode` | 400 | COLLAPSE→ #83 |
| 125 | `internal.unknown` | `fillin_handler.go:152`, `fillin_handler.go:170` | 500 | `internal.unknown` | 500 | **C-7.** COLLAPSE→ #21/#48 |

Documents' **main** mapper (`handler.go:1299-1349`) and the PDF webhook
(`pdf_webhook_handler.go:67-134`) emit only catalog constants — covered by §2.1. Their per-arm
status divergences are §3 rows R-1, R-2, R-3, R-5, R-11.

---

### 2.5 `internal/modules/controlleddocuments` — 21 own strings, all LITERAL

Every row must be edited at each listed `file:line` in
`internal/modules/controlleddocuments/delivery/http/routes.go`.

| # | old code | emit sites | status | NEW code | NEW default | note |
|---|---|---|---|---|---|---|
| 126 | `NO_ACTIVE_INSTANCE` | `:349`, `:512` | 404 | `notfound.active_document_instance` | 404 | |
| 127 | `CONTROLLED_DOCUMENT_NOT_FOUND` | `:514` | 404 | `notfound.controlled_document` | 404 | **R-5a**: documents maps the same `ErrCDNotFound` to generic `NOT_FOUND` (`handler.go:1310`) |
| 128 | `CONTROLLED_DOCUMENT_NOT_ACTIVE` | `:516` | 409 | `state.controlled_document_not_active` | 409 | **R-5b**: documents maps the same `ErrCDNotActive` to `STATE_TRANSITION_INVALID` (`handler.go:1342`) |
| 129 | `ACTIVE_REVISION_ALREADY_EXISTS` | `:518` | 409 | `state.active_revision_exists` | 409 | |
| 130 | `state.approval_route_missing` | `:525` | 409 | `state.approval_route_missing` | 409 | **C-9.** already aligned with approval; templates must join |
| 131 | `CONTROLLED_DOCUMENT_CODE_TAKEN` | `:527` | 409 | `conflict.controlled_document_code_taken` | 409 | |
| 132 | `CONTROLLED_DOCUMENT_CODE_ARCHIVED` | `:529` | 409 | `conflict.controlled_document_code_archived` | 409 | |
| 133 | `MANUAL_CODE_REASON_REQUIRED` | `:531` | 400 | `validation.manual_code_reason_required` | **422** (R-20) | |
| 134 | `OVERRIDE_REASON_REQUIRED` | `:533` | 400 | `validation.override_reason_required` | **422** (R-20) | |
| 135 | `VISIBILITY_SCOPE_INVALID` | `:535` | 400 | `validation.visibility_scope_invalid` | **422** (R-20) | |
| 136 | `OVERRIDE_TEMPLATE_DELETED` | `:537` | 409 | `state.override_template_deleted` | 409 | |
| 137 | `OVERRIDE_TEMPLATE_NOT_PUBLISHED` | `:539` | 409 | `state.override_template_not_published` | 409 | |
| 138 | `DICTIONARY_TOKEN_MISSING` | `:541` | 422 | `validation.dictionary_token_missing` | 422 | |
| 139 | `template_invalid` | `:543` | 422 | `validation.template_profile_mismatch` | 422 | **C-14.** COLLAPSE→ taxonomy `TEMPLATE_PROFILE_MISMATCH` (#152) — same condition, two codes, two statuses |
| 140 | `template.artifact_missing` | `:545` | 409 | `state.template_artifact_missing` | 409 | re-home |
| 141 | `template.artifact_invariant_unconfigured` | `:547` | 500 | `internal.template_artifact_invariant_unconfigured` | 500 | re-home |
| 142 | `creation_context.unconfigured` | `:549` | 500 | `internal.creation_context_unconfigured` | 500 | re-home |
| 143 | `PROFILE_NO_DEFAULT_TEMPLATE` | `:559` | 409 | `state.profile_no_default_template` | 409 | **R-21**: documents maps the same `ErrProfileHasNoDefaultTemplate` to `CONFLICT_ERROR` (`handler.go:1337`) |
| 144 | `DEFAULT_TEMPLATE_OBSOLETE` | `:561` | 409 | `state.default_template_obsolete` | 409 | |
| 145 | `AREA_NOT_FOUND` | `:565` | 404 | `notfound.process_area` | 404 | |
| 146 | `AREA_ARCHIVED` | `:569` | 409 | `state.process_area_archived` | 409 | |

CD also emits, as raw literals, catalog strings already covered in §2.1 — these must be replaced by
registry constant references, not just renamed:

| literal | sites | maps to §2.1 row |
|---|---|---|
| `VALIDATION_ERROR` | `routes.go:36,49,87,91,196,270,274`; `handler.go:153` | #1 |
| `INTERNAL_ERROR` | `routes.go:57,133,138,294,324,357,375,383,557,575`; `handler.go:73` | #21 |
| `UNAUTHENTICATED` | `routes.go:551` | #7 |
| `PROFILE_NOT_FOUND` | `routes.go:563` | #151 (taxonomy owns the typed const) — **C-15** |
| `PROFILE_ARCHIVED` | `routes.go:567` | #153 — **C-15** |
| `FORBIDDEN_CAPABILITY` (typed) | `routes.go:510` | #8 |
| `IDEMPOTENCY_KEY_REQUIRED` (typed) | `handler.go:150` | #25 |

---

### 2.6 `internal/modules/taxonomy` — 8 own strings, all CONST

Declared `taxonomy/delivery/http/routes_profiles.go:26-33`; emitted through the mapper at
`routes_profiles.go:325-350`.

| # | old code | decl | emit site | status | NEW code | NEW default | note |
|---|---|---|---|---|---|---|---|
| 147 | `PROFILE_NOT_FOUND` | :26 | `:327` | 404 | `notfound.document_profile` | 404 | **C-15.** CD emits the same string as a raw literal (`routes.go:563`) — collapse onto this registered code |
| 148 | `PROFILE_ARCHIVED` | :27 | `:329` | 409 | `state.document_profile_archived` | 409 | **C-15.** same, CD `routes.go:567` |
| 149 | `TEMPLATE_NOT_PUBLISHED` | :28 | `:331` | 409 | `state.template_not_published` | 409 | |
| 150 | `TEMPLATE_PROFILE_MISMATCH` | :29 | `:333` | 409 | `validation.template_profile_mismatch` | **422** (R-22) | **C-14.** COLLAPSE→ CD `template_invalid` (#139, already 422) |
| 151 | `PROFILE_CODE_IMMUTABLE` | :30 | `:335` | 400 | `validation.profile_code_immutable` | **422** (R-20) | |
| 152 | `PROFILE_ALREADY_EXISTS` | :31 | `:345` | 409 | `conflict.document_profile_exists` | 409 | |
| 153 | `FAMILY_NOT_FOUND` | :32 | `:347` | **409** | `notfound.document_family` | **404** (R-23) | a 404-shaped condition emitted at 409 |
| 154 | `PROFILE_CLASS_ROUTE_CONFLICT` | :33 | `:341` | 409 | `state.profile_class_route_conflict` | 409 | |

Taxonomy also emits catalog `VALIDATION_ERROR` (`:56,92,123,173,257,261,337,343` at 400; **`:339` at
422** → R-6) and `INTERNAL_ERROR` (14 sites).

---

### 2.7 `internal/modules/tokens` — 4 consts (2 redeclarations, 2 own)

Declared `tokens/delivery/http/handler.go:33-36`.

| # | old code | decl | emit site | status | NEW code | NEW default | note |
|---|---|---|---|---|---|---|---|
| 155 | `ALREADY_EXISTS` | :33 | `:209` | 409 | `conflict.already_exists` | 409 | **C-4.** DELETE the local const; reference the registry code (#13) |
| 156 | `NOT_FOUND` | :34 | `:197` | 404 | `notfound.resource` | 404 | **C-4.** DELETE the local const; reference #11 |
| 157 | `immutable_field` | :35 | `:200` (Problem.code, 422) and `:201` (**FieldError.code**) | 422 | `validation.field_immutable` | 422 | re-home from bare; field-level use needs the `field.` namespace decision (§4.2) |
| 158 | `reserved_name` | :36 | `:204` (Problem.code, 422) and `:205` (**FieldError.code**) | 422 | `validation.name_reserved` | 422 | same |

Tokens also emits catalog `VALIDATION_ERROR` (`:68,101,147` at 400; **`:207,:211` at 422** → R-6)
and `INTERNAL_ERROR` (`:79,86,107,131,153,178,214`).

---

### 2.8 New codes introduced by the rulings

These do not exist today; they are created by §3 rulings that split a currently-overloaded code.

| # | NEW code | default | replaces | ruling |
|---|---|---|---|---|
| 159 | `precondition.lock_version_stale` | 412 | templates `codeTplStaleLockVersion` (`CONCURRENT_MODIFICATION` @412) | R-7 |
| 160 | `internal.upstream_failed` | 502 | `INTERNAL_ERROR` @502 (`documents/…/export_handler.go:139`) | R-9 |
| 161 | `validation.placeholder_name_invalid` | 422 | templates `codeTplPlaceholderNameInvalid` (`VALIDATION_ERROR` @422) | R-6 |
| 162 | `validation.placeholder_name_duplicate` | 422 | templates `codeTplDuplicatePlaceholder` (`ALREADY_EXISTS` @422) | R-18 |
| 163 | `validation.doc_type_code_required` | 422 | templates `codeTplInvalidRequest` (`VALIDATION_ERROR` @422) | R-6 |
| 164 | `state.approval_stage_not_active` | 409 | templates `codeTplApprovalConflict` on `ErrStageNotActive`/`ErrNoActiveStage` | R-14 |
| 165 | `validation.content_hash_mismatch` — **NOT created**; use `precondition.content_hash_mismatch` | 412 | documents `VALIDATION_ERROR` @422, templates `CONFLICT_ERROR` @409 | R-2 |

**Grand total after the sweep:** 147 old strings − 11 deleted − 24 collapsed + 6 new ≈ **118
registered codes**. The exact figure is fixed once §3 is ratified.

### 2.9 Collision index (must collapse to ONE new code)

| id | condition | competing old codes | new code |
|---|---|---|---|
| C-1 | generic request-shape rejection | `VALIDATION_ERROR`(400), `validation.request_invalid` | `request.invalid` |
| C-2 | no/invalid session | `UNAUTHENTICATED`, `AUTH_UNAUTHORIZED` | `auth.unauthenticated` |
| C-3 | capability denied | `FORBIDDEN_CAPABILITY`, `authz.capability_denied` (approval + fillin) | `permission.capability_denied` |
| C-4 | tokens redeclares catalog strings | `tokens.NOT_FOUND`, `tokens.ALREADY_EXISTS` | `notfound.resource`, `conflict.already_exists` |
| C-5 | body over limit | `REQUEST_BODY_TOO_LARGE`, `validation.body_too_large` | `request.body_too_large` |
| C-6 | throttled | `RATE_LIMITED`, `authn.rate_limited` | `ratelimit.exceeded` |
| C-7 | unmapped server fault | `INTERNAL_ERROR`, `internal.unknown` (approval + fillin) | `internal.unknown` |
| C-8 | Idempotency-Key header absent | `IDEMPOTENCY_KEY_REQUIRED`, `idempotency.key_required` | `request.idempotency_key_required` |
| C-9 | no active approval route | `APPROVAL_ROUTE_MISSING` (templates), `state.approval_route_missing` (approval + CD) | `state.approval_route_missing` |
| C-10 | duplicate submission | `conflict.duplicate_submission`, templates `CONFLICT_ERROR` | `conflict.duplicate_submission` |
| C-11 | actor not in eligible pool | `signoff.not_eligible`, templates `FORBIDDEN_CAPABILITY` | `permission.signoff_actor_not_eligible` |
| C-12 | author cannot sign own doc | `sod.submitter_cannot_sign`, templates `FORBIDDEN_CAPABILITY` | `permission.sod_submitter_cannot_sign` |
| C-13 | idempotency key reused w/ different payload | `IDEMPOTENCY_KEY_REUSED`, `idempotency.key_conflict` | `conflict.idempotency_key_reused` |
| C-14 | template version ≠ document profile | CD `template_invalid`(422), taxonomy `TEMPLATE_PROFILE_MISMATCH`(409) | `validation.template_profile_mismatch` |
| C-15 | profile not found / archived | taxonomy typed consts vs CD raw literals (same strings) | `notfound.document_profile`, `state.document_profile_archived` |
| C-16 | JSON decode / empty body / content-type | approval `validation.*` vs fillin `validation.*` | `request.json_decode`, `request.empty_body`, `request.content_type_unsupported` |
| C-17 | content-hash mismatch | approval(412), documents(422), templates(409) | `precondition.content_hash_mismatch` — see R-2 |
| C-18 | empty eligible pool / submit choice | approval `validation.*`(422) vs templates `CONFLICT_ERROR`(422) | approval's codes win |

---

## 3. Status unification rulings needed

Each is **RECOMMENDED**; operator ratifies before implementation. Divergences are cited at
`file:line`.

**R-1 — `UPLOAD_MISSING`: 410 vs 409.** ⚠0088
`documents/delivery/http/handler.go:1327-1328` → **410 Gone**.
`templates/delivery/http/errors.go:21` (`codeTplUploadMissing`) used at `errors.go:68` and
`errors.go:107` → **409 Conflict**.
RFC 9110 §15.5.11: 410 asserts the target resource *is permanently gone* — it is a statement about
the **request target**, not about a missing prerequisite. In both modules the target (document /
template version) exists and is addressable; what is missing is its uploaded artifact. That is a
prerequisite-not-satisfied conflict.
**RECOMMENDED: 409** for `state.upload_missing` in both. Note ADR 0088 increases traffic here
(`templates/application/create.go` wraps `ErrUploadMissing` for short/absent blank-source hashes),
so 409 is also the status 0088's own tests already assume.

**R-2 — content-hash mismatch: 422 vs 409 vs 412.**
`documents/delivery/http/handler.go:1331-1332` → **422 `VALIDATION_ERROR`**.
`templates/delivery/http/errors.go:20,66` and `:119` (`approvalapp.ErrContentHashMismatch`) →
**409 `CONFLICT_ERROR`**.
`approval/http/errors.go:54,~250` (`ErrContentHashMismatch`) → **412 `precondition.content_hash_mismatch`**.
The client *supplies* the expected hash; the server compares it to current state and refuses. That
is textbook RFC 9110 §15.5.13 **412 Precondition Failed** — the same class as If-Match. 422 is
wrong (the value is well-formed and was correct at some point); 409 is wrong (no concurrent-writer
race is asserted).
**RECOMMENDED: 412 `precondition.content_hash_mismatch` in all three.** OpenAPI response sets for
the documents and templates routes must gain 412.

**R-3 — `ErrRevisionTitleRequired`: 400 vs 422.**
`documents/delivery/http/handler.go:1315` (with `:1317`) → **400 `VALIDATION_ERROR`**.
`approval/http/errors.go:80` → **422 `validation.revision_title_required`**.
The body parses; a business rule (REV≥1 needs a title) rejects it. RFC 9110 §15.5.21.
**RECOMMENDED: 422 `validation.revision_title_required`.**

**R-4 — approval-domain sentinels mapped differently by approval and templates.**

| sentinel | approval (`http/errors.go`) | templates (`delivery/http/errors.go`) | RECOMMENDED |
|---|---|---|---|
| `domain.ErrActorNotEligible` | 403 `signoff.not_eligible` (:46) | 403 `FORBIDDEN_CAPABILITY` (:139-140) | 403 `permission.signoff_actor_not_eligible` |
| `domain.ErrAuthorCannotSign` | 403 `sod.submitter_cannot_sign` (:47) | 403 `FORBIDDEN_CAPABILITY` (:141-142) | 403 `permission.sod_submitter_cannot_sign` |
| `domain.ErrEmptyEligiblePool` | 422 `validation.empty_eligible_pool` (:91) | 422 `CONFLICT_ERROR` (:127-128) | 422 `validation.empty_eligible_pool` |
| `domain.ErrSubmitChoiceRequired` | 422 `validation.submit_choice_required` (:92) | 422 `CONFLICT_ERROR` (:132-133) | 422 `validation.submit_choice_required` |
| `domain.ErrSubmitChoiceConstraintViolated` | 422 `validation.submit_choice_constraint_violated` (:93) | 422 `CONFLICT_ERROR` (:136-137) | 422 `validation.submit_choice_constraint_violated` |

Statuses already agree in every row; only the **codes** diverge — templates flattens five distinct
conditions onto one generic `CONFLICT_ERROR`, which is precisely the defect that makes the FE
message map useless for template submit. Templates must adopt approval's codes verbatim.

**R-5 — `ErrCDNotActive` / `ErrCDNotFound` / `ErrProfileHasNoDefaultTemplate`: CD vs documents.**

| sentinel | controlleddocuments | documents | RECOMMENDED |
|---|---|---|---|
| `ErrCDNotActive` | 409 `CONTROLLED_DOCUMENT_NOT_ACTIVE` (`routes.go:516`) | 409 `STATE_TRANSITION_INVALID` (`handler.go:1341-1343`) | 409 `state.controlled_document_not_active` |
| `ErrCDNotFound` | 404 `CONTROLLED_DOCUMENT_NOT_FOUND` (`routes.go:514`) | 404 `NOT_FOUND` (`handler.go:1310-1311`) | 404 `notfound.controlled_document` |
| `ErrProfileHasNoDefaultTemplate` | 409 `PROFILE_NO_DEFAULT_TEMPLATE` (`routes.go:559`) | 409 `CONFLICT_ERROR` (`handler.go:1337-1338`) | 409 `state.profile_no_default_template` |

Statuses agree; the specific code must win over the generic one in all three.

**R-6 — `VALIDATION_ERROR` emitted at 422 (6 sites).** One code cannot default to both 400 and 422.
Sites currently at 422:
`documents/delivery/http/handler.go:1332` (content-hash → R-2, becomes 412);
`documents/delivery/http/handler.go:1345` (`form_data_invalid` prefix match);
`templates/delivery/http/errors.go:82,84,86` (`codeTplPlaceholderNameInvalid`);
`templates/delivery/http/errors.go:92` (`codeTplInvalidRequest` / `ErrDocTypeCodeRequired`);
`taxonomy/delivery/http/routes_profiles.go:339`;
`tokens/delivery/http/handler.go:207,211`.
**RECOMMENDED:** `VALIDATION_ERROR` → `request.invalid` @400. Each 422 site gets a specific
`validation.*` code: `validation.form_data_invalid`, `validation.placeholder_name_invalid` (#161),
`validation.doc_type_code_required` (#163), and for taxonomy/tokens the generic `validation.failed`
@422 (#121) unless a specific condition is identified during the sweep.

**R-7 — `CONCURRENT_MODIFICATION` at 412.** `templates/delivery/http/errors.go:19` binds
`codeTplStaleLockVersion` to `CONCURRENT_MODIFICATION` but returns **412** at `errors.go:63-64`,
while `documents/delivery/http/handler.go:1340` returns the same code at **409**.
`ErrStaleLockVersion` is an OCC precondition on a caller-supplied lock version.
**RECOMMENDED:** new `precondition.lock_version_stale` @412 for templates;
`conflict.concurrent_modification` @409 keeps the documents site.

**R-8 — `IDEMPOTENCY_KEY_REUSED` at 422.** `platform/idempotency/middleware.go:146` returns **422**.
A key reused with a different request fingerprint is a collision with prior state, not a content
defect. RFC 9110 §15.5.10.
**RECOMMENDED: 409** for `conflict.idempotency_key_reused`.

**R-9 — `INTERNAL_ERROR` emitted at 501 and 502.** ~25 sites in `iam/delivery/http/*` and
`security/delivery/http/handler.go` pair `http.StatusNotImplemented` with `problem.CodeInternalError`
(e.g. `iam/…/people_handler.go:70,129,175,229,261,284,348,395`; `iam/…/routes_memberships.go:112,189,261,336`;
`iam/…/routes_roles_caps.go:78`; `iam/…/sessions_handler.go:106,226`; `iam/…/admin_handler.go:168`;
`iam/…/observability_handler.go:41,59`; `security/…/handler.go:87,109,149`).
`documents/delivery/http/export_handler.go:139` pairs **502** with `INTERNAL_ERROR`.
**RECOMMENDED:** 501 sites → `internal.not_implemented` @501 (the code already exists, #22);
502 site → new `internal.upstream_failed` @502 (#160). A code whose registered default is 500 must
never be emitted at 501/502 — this is exactly what the registry's status binding prevents.

**R-10 — templates `codeTplApprovalConflict` emitted at both 409 and 422.**
409 at `errors.go:101,113,115,117,139`; 422 at `errors.go:128,133,137`.
Resolved by R-4 (the 422 arms adopt approval's `validation.*` codes) plus R-14.

**R-11 — `ErrStaleBase`: two codes at the same status.**
`documents/delivery/http/handler.go:1339-1340` → 409 `CONCURRENT_MODIFICATION`.
`templates/delivery/http/errors.go:18,59-60` → 409 `STALE_BASE`.
**RECOMMENDED: 409 `conflict.stale_base`** in both — `STALE_BASE` is the more precise name and is
already in the FE catalog.

**R-12 — `UNKNOWN_ROLE` at 400.** `iam/…/people_handler.go:431`, `routes_memberships.go:323`. The
body parses; the *value* names a role that does not exist.
**RECOMMENDED: 422 `validation.role_unknown`.** (Low blast radius: 2 sites.)

**R-13 — `submit.invalid_supersede_target` at 409.** `approval/http/errors.go:42`. The caller
supplied a target id that fails a business rule; no race is asserted.
**RECOMMENDED: 422 `validation.supersede_target_invalid`.**

**R-14 — `domain.ErrNoActiveStage` / `approvalapp.ErrStageNotActive` fold into
`state.instance_completed`.** `approval/http/errors.go:276-278` maps `ErrNoActiveStage` to
`approvalCodeStateInstanceCompleted`; `templates/delivery/http/errors.go:117,139` maps both
`ErrStageNotActive` and `ErrNoActiveStage` to `CONFLICT_ERROR`. "No active stage" and "instance
completed" are different conditions with different operator remedies.
**RECOMMENDED:** new `state.approval_stage_not_active` @409 (#164) for both sentinels in both
modules; `state.approval_instance_completed` keeps only `infrastructure.ErrInstanceCompleted`.

**R-15 — `validation.reason_required` at 400.** `approval/http/errors.go:67`, emitted for
`ErrReasonRequired` / `ErrRouteDeactivateReasonRequired`.
**RECOMMENDED: 422** — sibling reason codes (`validation.reason_for_change_required`,
`validation.reason_category_invalid`) are already 422; 400 vs 422 for the same field class is
incoherent.

**R-16 — `validation.profile_not_configured` at 400.** `approval/http/errors.go:84`. The comment
records that it was inherited from the removed finalize handler.
**RECOMMENDED: 422**, matching `validation.profile_unknown` (422) which is the same class.

**R-17 — `state.verdict_ready_on_approval_stage` at 422.** `approval/http/errors.go:108`. The
prefix claims lifecycle state; the status and the in-file comment both classify it as a business-rule
rejection of a supplied verdict.
**RECOMMENDED:** rename to `validation.verdict_ready_on_approval_stage`, keep 422.

**R-18 — `ALREADY_EXISTS` at 422.** `templates/delivery/http/errors.go:28,88`
(`codeTplDuplicatePlaceholder` for `ErrDuplicatePlaceholderName`). `ALREADY_EXISTS` defaults to 409
everywhere else.
**RECOMMENDED:** new `validation.placeholder_name_duplicate` @422 (#162) — the duplicate is inside
the caller's own payload, so it is a content defect, not a collision with stored state.

**R-19 — `not_a_choice_placeholder` at 400.** `documents/delivery/http/fillin_handler.go:136`. The
JSON parses; the addressed placeholder is the wrong kind.
**RECOMMENDED: 422 `validation.placeholder_not_choice`.**

**R-20 — the CD/taxonomy "reason/scope/immutable required" cluster at 400.**
`controlleddocuments/…/routes.go:531` (`MANUAL_CODE_REASON_REQUIRED`), `:533`
(`OVERRIDE_REASON_REQUIRED`), `:535` (`VISIBILITY_SCOPE_INVALID`);
`taxonomy/…/routes_profiles.go:335` (`PROFILE_CODE_IMMUTABLE`).
All four parse fine and fail a business rule.
**RECOMMENDED: 422** for all four.

**R-21 — covered by R-5 (third row).**

**R-22 — `TEMPLATE_PROFILE_MISMATCH` 409 vs `template_invalid` 422.**
`taxonomy/…/routes_profiles.go:333` → 409; `controlleddocuments/…/routes.go:543` → 422. Same
condition (template version does not belong to the document profile), two codes, two statuses.
**RECOMMENDED: 422 `validation.template_profile_mismatch`** — the caller supplied a
template/profile pair that is invalid; nothing is racing.

**R-23 — `FAMILY_NOT_FOUND` at 409.** `taxonomy/…/routes_profiles.go:347`. A missing referenced
family is a 404-class condition on a supplied reference. Two readings are defensible: 404 (the
referenced family does not exist) or 422 (the supplied `family_code` value is invalid). Because the
family is a *referenced value inside the body*, not the request target, RFC 9110 §15.5.5 argues
against 404 on the request URI.
**RECOMMENDED: 422 `validation.family_unknown`** — and if the operator prefers 404, then the code
must be `notfound.document_family` and the route's OpenAPI response set gains 404.
*(This is the one row in §3 where the recommendation changes the §2 new-name; §2 row #153 currently
carries the 404 form. Ratify before the sweep.)*

**R-24 — `RATE_LIMITED` vs `authn.rate_limited`.** `platform/ratelimit/middleware.go:219` (429) and
`approval/http/errors.go:56` (429). Same status, two codes for "you are being throttled". The
approval one is the signature-reauth limiter.
**RECOMMENDED:** single `ratelimit.exceeded` @429. If per-surface distinction matters for the FE,
use `ratelimit.exceeded` + `ratelimit.signature_reauth` as two registered codes — but do **not**
keep `authn.*` for a throttle.

---

## 4. Registry API sketch + execution order

### 4.1 Making a bare string literal fail to compile

```go
package problem

// Code is an opaque, registry-issued problem code. The zero value is invalid and
// never reaches the wire: problem.New rejects it.
//
// Code is a struct, not a string alias, precisely so that an untyped string
// constant CANNOT be implicitly converted to it. `problem.New(400, "OOPS", …)`
// is a compile error; the only way to obtain a Code is problem.Register.
type Code struct{ s string }

func (c Code) String() string  { return c.s }
func (c Code) IsZero() bool    { return c.s == "" }

// MarshalJSON keeps the wire representation a plain JSON string — byte-identical
// to today's `"code":"state.route_inactive"`.
func (c Code) MarshalJSON() ([]byte, error) { return json.Marshal(c.s) }

// UnmarshalJSON is LENIENT by design: clients and tests decode Problems produced
// by other builds. It does not require registry membership; use Lookup for that.
func (c *Code) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	c.s = s
	return nil
}
```

**Ergonomic cost (before → after).**

| aspect | before | after |
|---|---|---|
| declaration | `const codeX problem.Code = "state.route_inactive"` | `var CodeX = problem.Register("approval", "state.route_inactive", 409)` |
| call site | `problem.New(http.StatusConflict, codeX, msg)` | `problem.NewFor(CodeX, msg)` (status from registry) or `problem.NewWithStatus(http.StatusGone, CodeX, msg)` for a documented override |
| raw literal | compiles silently | **compile error** |
| comparable | yes (`==`) | yes — struct with one comparable field |
| map key | yes | yes — `map[Code]string` (used by `templates.friendlyMsg`) still compiles |
| JSON out | `"code":"x"` | `"code":"x"` — identical |
| JSON in | any string | any string (lenient); `problem.Lookup(s) (Code, bool)` for strict checks |
| `string(code)` | works | replaced by `code.String()` — ~12 call sites (`approval/http/errors.go:470`, `templates/…/handler.go:250,276`, `documents/…/handler.go:1352,1358`, logging sites) |
| `switch code { case … }` | works | works |

The only real cost is the mechanical `string(code)` → `code.String()` sweep and the loss of
`const`-ness (registrations are `var` because `Register` runs at init). Neither is load-bearing.

### 4.2 `Register`, status binding, and the field-code namespace

```go
type Registration struct {
	Code          Code
	Module        string // owning module, for the wiki table + drift lint only — NEVER on the wire
	Family        string // derived from the code prefix; validated against the closed set
	DefaultStatus int
	DeclaredAt    string // runtime.Caller file:line, for the drift report
}

var (
	regMu    sync.Mutex
	registry = map[string]Registration{}
	families = map[string]int{ // closed set → family default status
		"request": 400, "validation": 422, "auth": 401, "permission": 403,
		"notfound": 404, "state": 409, "conflict": 409, "precondition": 412,
		"ratelimit": 429, "internal": 500,
	}
)

// Register issues a Code. It PANICS at init on: unknown family, malformed name,
// duplicate registration, or a status outside [400,599]. A panic here is a build
// break in every binary and in every test — that is the point.
func Register(module, code string, defaultStatus int) Code { … }

// RegisterField issues a FIELD-level code (FieldError.code). Separate namespace,
// no HTTP status, no family prefix requirement.
func RegisterField(module, code string) FieldCode { … }
```

`problem.New` keeps its signature for the override path but gains the status-free constructor:

```go
func NewFor(c Code, title string) *Problem            // status = registry default
func New(status int, c Code, title string) *Problem   // explicit, documented override
```

An api-lint rule flags every `New(` whose status ≠ the registered default and requires a
`//problem:override <reason>` comment on the line above. That turns today's silent per-call-site
status drift (the root cause of every §3 row) into a reviewable, greppable annotation.

### 4.3 Self-registration without an import cycle

Direction is already one-way and stays that way: **modules import `platform/problem`; `platform/problem`
imports nothing from `internal/modules`.** Each module keeps its codes in one file:

```go
// internal/modules/approval/http/codes.go
package approvalhttp

import "metaldocs/internal/platform/problem"

var (
	CodeStateRouteInactive = problem.Register("approval", "state.approval_route_inactive", 409)
	CodeNotFoundRoute      = problem.Register("approval", "notfound.approval_route", 404)
	// …
)
```

Package-level `var` initialisation populates the registry when the package is **linked**. That is
the discovery problem: a generator that does not link a package never sees its codes.

**Discovery — RECOMMENDED: linked dumper + import-coverage lint.**

`cmd/problem-codes-dump/main.go` blank-imports every package that registers codes and prints the
registry as JSON:

```go
package main

import (
	_ "metaldocs/internal/modules/approval/http"
	_ "metaldocs/internal/modules/controlleddocuments/delivery/http"
	_ "metaldocs/internal/modules/documents/delivery/http"
	_ "metaldocs/internal/modules/taxonomy/delivery/http"
	_ "metaldocs/internal/modules/templates/delivery/http"
	_ "metaldocs/internal/modules/tokens/delivery/http"
	"metaldocs/internal/platform/problem"
)

func main() { problem.DumpRegistry(os.Stdout) }
```

- Cannot import `apps/api/cmd/metaldocs-api` (it is `package main`), so the import list is explicit.
- The list is drift-prone → guard it with an **api-lint** that AST-scans `internal/**` for
  `problem.Register(` calls and fails if the containing package is not blank-imported by
  `cmd/problem-codes-dump`. This is a 40-line lint and it makes the list self-maintaining.
- Rejected alternative: pure AST scan of `Register(` literals. It re-implements constant folding,
  cannot validate uniqueness across packages, and cannot see the status binding if anyone ever
  writes `Register(mod, code, statusConst)`. The linked dumper reports **runtime truth**, which is
  the repo's standing tie-breaker.

The dumper replaces `scripts/dump-error-codes.go` entirely (delete it — it is a regex scraper of the
exact kind this work exists to remove) and emits three artifacts:

1. `frontend/apps/web/src/lib/api/error-codes.generated.json` — now `{code, family, default_status}` triples.
2. An OpenAPI fragment for `components.schemas.Problem.code.enum` (currently `{type: string}` with
   no enum, `api/openapi/v1/openapi.yaml:7190`).
3. `wiki/architecture/problem-codes.md` — the code × family × status × owning-module table.

**CI freshness gate** (`make check-problem-codes`, blocking):
`go run ./cmd/problem-codes-dump > /tmp/actual` then `git diff --exit-code` against all three
artifacts, plus the import-coverage lint, plus a FE test that `Object.keys(errorMessages)` equals
the generated code set exactly (today's coverage test compares map↔snapshot, never map↔backend —
that is why the snapshot could sit 3 codes stale).

### 4.4 Ordered execution plan

Legend: **[J]** = judgment required (Opus/Sonnet + operator gate). **[M]** = mechanical, safe for a
Haiku-class agent given §2 verbatim.

| # | step | class | gate / notes |
|---|---|---|---|
| 0 | ADR 0088 commits | — | No §2 row's file is modified, so this does **not** block steps 1-3. It **does** block R-1's ratification (`templates/application/create.go` newly routes through `UPLOAD_MISSING`). |
| 1 | Operator ratifies §1 family set (incl. the `request.` split) and the §1.4 decision rule | **[J]** | hard gate — every later step depends on the family names |
| 2 | Operator ratifies §3 R-1…R-24 | **[J]** | hard gate — fixes the NEW default status column |
| 3 | Implement `Code` struct + `Register` + `RegisterField` + `NewFor` + registry + family validation in `internal/platform/problem`; keep `codes.go` constants temporarily as `Register(...)` calls with their CURRENT strings | **[J]** | one commit, compiles green, wire unchanged. This is the only step that touches type shape. |
| 4 | Sweep `internal/platform/*` + `apps/*` off raw `string(code)` onto `code.String()` | **[M]** | ~12 sites; `go build ./...` is the oracle |
| 5 | Rename catalog values per §2.1; delete the 11 dead constants | **[M]** | §2.1 is literal; wire changes here |
| 6 | Rename approval's 68 const values per §2.2 | **[M]** | single file, single-line-per-row edits |
| 7 | Replace CD's 43 raw literals with registry references per §2.5 | **[M]** | `controlleddocuments/delivery/http/routes.go` + `handler.go` only |
| 8 | Replace fill-in's 11 raw literals per §2.4 | **[M]** | `documents/delivery/http/fillin_handler.go` only |
| 9 | Rename taxonomy (8) and tokens (4) per §2.6/§2.7; delete tokens' 2 redeclarations | **[M]** | |
| 10 | Apply the collapses C-1…C-18 | **[J]** | each touches ≥2 modules' mappers and changes which sentinel maps where; not mechanical |
| 11 | Re-point templates' 22 aliases; create new codes #159-#164; apply R-2/R-4/R-5/R-7/R-10/R-14 | **[J]** | ⚠0088 — sequence after 0088 merges |
| 12 | Apply the status-only rulings (R-8, R-9, R-12, R-13, R-15, R-16, R-19, R-20, R-23) | **[M]** given §3 | each is "change this int at this line" |
| 13 | Add `//problem:override` annotations wherever a call site keeps a non-default status | **[M]** | mechanical once step 12 lands |
| 14 | Build `cmd/problem-codes-dump`; delete `scripts/dump-error-codes.go` | **[J]** | |
| 15 | Regenerate the 3 artifacts; rewrite `errorMessages.ts` keys; add the enum to `api/openapi/v1/openapi.yaml:7190`; full `oapi-codegen` regen | **[M]** | ⚠ full regen only — partial spec edits churn every module's embedded `swaggerSpec` |
| 16 | Add the api-lint import-coverage rule + `make check-problem-codes` + FE map↔backend parity test; **delete** `codes_catalog_guard_test.go` and its `guardedPackages` allowlist (the type system now enforces what the allowlist approximated) | **[J]** | |
| 17 | Update OpenAPI per-route response sets where §3 adds a status (412 on documents/templates hash routes; 404 or 422 on the taxonomy family route) | **[J]** | contract-first: spec first, then regen |
| 18 | `go build ./...`, `go test ./...`, `go vet -tags integration ./...`, `make test` | **[M]** | the integration-tag compile gap is a known trap: untagged `go test` does not compile `//go:build integration` files |

**Sequencing note for step 15.** Every module embeds the OpenAPI spec; a partial regen is forbidden
drift. Step 15 must be one commit that regenerates everything.

**Why steps 5-9 are safe for a Haiku-class agent and step 10 is not.** Steps 5-9 are
value-substitution inside a single file, with §2 giving the exact old string, exact file:line, and
exact new string — no decision is left open. Step 10 requires deleting a `case` arm in one module's
mapper and redirecting a sentinel to another module's code, which changes *which error reaches which
arm* and can silently reorder `errors.Is` precedence. That needs a reviewer who reads the whole
switch.
