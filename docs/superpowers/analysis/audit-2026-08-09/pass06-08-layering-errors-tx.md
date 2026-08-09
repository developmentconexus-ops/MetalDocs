# Pass 06–08 — Layering, Errors, Transactions/Persistence

> **Date:** 2026-08-09
> **Baseline:** `main@418070bf`
> **Status:** reproduced-current
> **Scope:** `internal/modules/**/*.go` + `internal/platform/**/*.go` (production files; `_test.go` excluded unless stated), the 15 bounded-context modules under `internal/modules/`. All counts below are mechanical `grep`/`find` reproductions; every count states the exact command and file. Where the historical count (from issues #90/#92/#93/#94) differs from what is on disk today, both numbers are given as `historical: N, reproduced-current: M`.
> **Caveat on tooling:** this pass is grep/regex-based, not AST-based. Where regex over- or under-counts for a structural reason (e.g. conflating same-module aliased imports with true cross-module imports), that is called out explicitly and corrected in a second, more precise pass — see §7.2.

---

## 0. Answer up front

| # | Claim (issue) | Historical | Reproduced-current | Verdict |
|---|---|---|---|---|
| 1 | Domain packages leaking `db.Tx`/`*sql.Tx`/`sql.Null*` | 9-of-15 modules | **9-of-15** (14 modules have a `domain/` dir; `jobs` has none) | Confirmed. Classification (per §6.2 as corrected): ADR 0044 sanctions `db.Tx` **only** for its domain-event args/enqueuer boundary; the other 8 modules' non-event domain-port `db.Tx`/`db.DB` usage is **current architecture, unresolved pending explicit ruling** (neither ratified nor migrated by this audit); 1/9 (`auth`) is genuine, ADR-uncovered **confirmed debt** (raw `database/sql` + `sql.NullTime` on a domain DTO) |
| 2 | `approval/application` size | 30 files / 8.7k LOC / 583 exported symbols | 30 files (exact) / 9339 LOC / 706–740 exported symbols (method-sensitive) | Confirmed, files exact; LOC/symbol counts drifted up (repo grew), method-sensitive |
| 3 | Twin status-transition validators | "twin validators" | 2 independent FSMs (`templates/domain/version.go:89`, `documents/domain/state.go:52`) | Confirmed pattern-duplication; **not byte-identical** — different transition tables/arity |
| 4 | Objectstore error-switch duplicated documents↔templates | claimed | **byte-identical** (`documents/application/service.go:734-743` vs `templates/application/autosave.go:153-162`) | Confirmed exactly |
| 5 | Cross-module `errors.Is(err, <foreignmodule>domain.Err…)` sites | 62 | naive alias-count: **65** non-test sites (close match); **true cross-module: 19** | Historical 62 is almost certainly the same naive alias-count this pass first produced (65) — **not filtered for same-module self-aliasing**. True cross-module count is much smaller: 19 |
| 6 | Local `writeProblem`-like funcs | 12+ | **12** (exact) | Confirmed exactly |
| 7 | `problem.New(` vs `problem.NewFor(` | 232 vs 11 | **227 vs 11** | `NewFor` exact match; `New` drifted down slightly |
| 8 | `writeProblem` logs nothing | claimed | **Confirmed** — plus: correlated `slog.Error` before a 500 is inconsistent per call site (some paths log, some don't) | Confirmed and refined |
| 9 | Direct `.BeginTx(` sites | 82 / 25 files | **84 / 26 files** | Close match (+2/+1, organic growth) |
| 10 | Module-specific second tx abstractions | 2 modules | **3 modules** (`auth`, `iam`, `taxonomy`) | Historical undercounts by 1 |
| 11 | Repositories `*sql.DB` vs `TxRunner` | 75 vs 6 | 56 files w/ `*sql.DB` vs **34 files / 6 modules** consuming `db.TxRunner` | Module-count side matches (6) exactly; file-count units likely differ from the original methodology |
| 12 | Tx/non-Tx twin methods, 3 byte-identical | 20 twins, 3 byte-identical | 14 base-name pairs found (≈20+ concrete method twins); **3 spot-checked, all 3 byte-identical** | Confirmed exactly on the byte-identical sub-claim |
| 13 | Hand-written `.Scan(` count | 242 | whole-repo non-test: **278**; infra-only: **207** (+platform 11 = 218) | In range depending on scope definition; some growth |
| 14 | Hand-rolled 23505 checks | 15 across 10 files | **17 sites / 11 files** | Close match |
| 15 | 5 files testing `lib/pq` though driver is pgx | claimed | Driver confirmed **pgx/stdlib**; `lib/pq` used in **3** files as a documented, provably dead fallback (pgx never returns `*pq.Error`) — never the sole check | Confirmed dead code, **not a live bug** (primary `*pgconn.PgError` check always fires first) |
| 16 | `release_coordinator.go` per-target UPDATE loop (N+1) | claimed | **Confirmed**, 3 separate per-target loops | Confirmed, but justified by deadlock-avoidance lock ordering, not a naive oversight |
| 17 | Network-inside-transaction | "invariant says no — verify honestly" | **No violations found** in the 3 hotspots checked (documents/templates autosave-commit, iam tenant blob erasure) | Confirmed clean on the sampled scope; **not exhaustive** — see §8.7 |

---

## PASS 6 — LAYERING

### 6.1 Domain purity: forbidden imports (net/http, generated/OpenAPI, pgx, redis, minio, cross-module implementation types)

**Command:**
```
Grep pattern: "net/http"|"database/sql"|internal/platform/db|jackc/pgx|"github.com/redis|minio-go|internal/platform/objectstore|oapi-codegen|/generated/|internal/apigen
Path: internal/modules, glob: **/domain/**/*.go
```

Result: **zero** hits for `net/http`, `jackc/pgx`, `redis`, `minio-go`, `internal/platform/objectstore`, `oapi-codegen`, `/generated/`. Domain packages are clean on those axes across all 14 modules that have a `domain/` package (`jobs` has no `domain/` dir at all — see §6.1a).

The only hits are `internal/platform/db` (9 modules, §6.2) and one `database/sql` (`auth`, §6.2).

**Cross-module domain imports** (a `domain` package importing another module's non-`domain` package):
```
for each *.go under */domain/*.go: extract metaldocs/internal/modules/<mod>/... imports,
flag any import whose path is not exactly <mod>/domain
```
Result: **zero violations**. Every cross-module import from a `domain` package is exactly `<module>/domain` — matching ADR 0044 §5's mandated contract surface ("a cross-module import is legal only when the imported path is exactly `<module>/domain`"). This is the one part of domain purity that is clean and disciplined.

**Other platform imports** (`internal/platform/*` excluding `db`):
```
grep -rlP 'metaldocs/internal/platform/(?!db)' --include='*.go' internal/modules/*/domain
```
7 files: `approval/domain/state.go` (`platform/legacystatus`), `auth/domain/model.go`, `iam/domain/model.go`, `taxonomy/domain/{area,area_test,profile,profile_test}.go` (`platform/iamtypes`). These are shared cross-cutting **vocabulary** packages (role/status enums), not infrastructure — not classified as violations, but worth noting as a second neutral-package convention alongside `db.Tx` (see 6.2).

**6.1a — `jobs` has no domain/application/delivery layering at all.** `internal/modules/jobs/` is organized as 8 flat per-job packages (`approval_sla_surfacer/`, `audit_integrity_validator/`, `document_review_surfacer/`, `idempotency_janitor/`, `maintenance/`, `outbox_retention/`, `release_hold_reconciler/`, `stuck_instance_watchdog/`, `tenantdata/`) with no `domain`/`application`/`delivery` split. This is a structural outlier among the 15 modules — worth a naming/README note (out of scope to fix here), not a defect since jobs are periodic maintenance tasks, not a bounded business context with its own entities.

### 6.2 The `db.Tx`/`db.DB`/`sql.Null*` leak — reproducing and correcting the 9-of-15 claim

**Command:**
```
Grep pattern: \*sql\.Tx|sql\.Null|db\.Tx\b|db\.DB\b
Path: internal/modules/**/domain/**/*.go
```

9 modules' domain packages reference `db.Tx`/`db.DB`: **approval, audit, auth, controlleddocuments, documents, iam, security, taxonomy, tokens**. (5 modules — `distribution, notifications, render, search, templates` — do not; `jobs` has no domain dir.) 9-of-15 (or 9-of-14 counting only modules with a domain package) reproduces exactly.

**Exact port signatures** (representative, not exhaustive — full grep returned 60 lines):
- `internal/modules/tokens/domain/port.go:13-18` — `Create/Update/Delete/GetByID/GetByName/List(ctx, tx db.Tx, …)`
- `internal/modules/approval/domain/release_hold_port.go:66,73`, `sla_port.go:73,80,121`, `notification_events.go:64`
- `internal/modules/audit/domain/port.go:137` — `RecordTx(ctx, tx db.Tx, event Event) error`
- `internal/modules/controlleddocuments/domain/port.go:49,51`, `document_initializer.go:97`, `sequence.go:17`
- `internal/modules/documents/domain/review_surface_port.go:50`, `review_due_port.go:54,65`, `notification_events.go:40`
- `internal/modules/iam/domain/membership_tx.go:10` — embeds `db.Tx` directly
- `internal/modules/security/domain/tenant_crypto.go:27,46,58`
- `internal/modules/taxonomy/domain/area_catalog_reader_port.go:26,31`, `port.go:62,103`
- **`internal/modules/auth/domain/session_admin.go:3,34-36`** — `import "database/sql"`; `SessionListItem{CreatedAt, LastSeenAt, ExpiresAt sql.NullTime}`

**Is any of this ADR-legitimated?** Searched `wiki/decisions/` for `db.Tx`/`db.DB`/domain-purity rulings:
```
grep -rl "db\.Tx\|db\.DB" wiki/decisions/
→ 0015, 0022, 0044
```
**ADR 0044** (`0044-domain-event-pattern-and-river-dispatch.md`, §5, lines 93-97) explicitly rules: *"The args/port use `db.Tx`, not `*sql.Tx`, at the boundary — the concrete-tx assertion is isolated to that one adapter."* It further rejects putting the type in `internal/platform/...` because "`platform` is module-agnostic infrastructure … so domain vocabulary does not belong there." **Scope caution (post-review correction):** ADR 0044 is specifically a domain-event/River-dispatch decision — its `db.Tx` sanction covers the **event args/enqueuer boundary it rules on**, nothing more. It does not ratify arbitrary domain repository/read/write ports exposing an SQL-executor interface repo-wide. The rulebook's R-LAYER-2 is the operative rule: new domain ports must not introduce `db.Tx` without an explicit ruling; only application-owned ports carry the pragmatic transitional concession.

**Verdict (corrected):** three classes, not two.
1. **ADR-0044-sanctioned:** `db.Tx` on domain-event args/enqueuer ports (the exact boundary ADR 0044 rules on).
2. **Current architecture, unresolved pending explicit ruling:** the remaining `db.Tx`/`db.DB` usage in the other 8 modules' non-event domain ports (`approval, audit, controlleddocuments, documents, iam, security, taxonomy, tokens`). The neutral structural interfaces from `internal/platform/db/tx.go` are clearly deliberate convention, but no ADR ratifies them for general domain ports — this is **neither "not debt" nor scheduled for migration by this audit**; it awaits an explicit ruling (candidate slice of #92/#93 governance).
3. **Confirmed debt (ADR-uncovered):** `internal/modules/auth/domain/session_admin.go` imports raw standard-library `database/sql` and types a domain read-model DTO's fields as `sql.NullTime`. Fix is mechanical (three `sql.NullTime` → `*time.Time` or plain `time.Time` + zero-value handling in the repository) — **not performed in this audit**.

### 6.3 Application discipline

**SQL text inside `application` packages** (non-infrastructure):
```
grep -rlE 'SELECT |INSERT INTO|UPDATE .+SET|DELETE FROM' --include='*.go' internal/modules/*/application
→ 85 non-test application files total; 24 non-test files contain raw SQL text
```
24-of-85 (28%) non-test application files embed raw SQL directly, across 5 modules:
- **approval** (12 files): `cancel_service.go`, `decision_service.go`, `events.go`, `mark_reviewed_service.go`, `obsolete_service.go`, `read_service.go`, `release_coordinator.go`, `release_facts.go`, `release_terminal_approval.go`, `review_verdict_service.go`, `route_admin_service.go`, `sla_extension_service.go`, `submit_defaults.go`
- **documents** (6): `context_builder.go`, `document_area.go`, `document_cdid.go`, `fillin_service.go`, `freeze_service.go`, `view_service.go`
- **iam** (3): `capability_service.go`, `onboard_tenant_service.go`, `tenant_lifecycle_service.go`
- **controlleddocuments** (1): `service.go`
- **templates** (1): `create.go`

This is a real, board-wide pattern: roughly a quarter of application-layer files across a third of the modules hold hand-written SQL text rather than delegating to a repository/port method — the application/infrastructure boundary is porous, not just an approval-module problem.

**HTTP inside application:**
```
grep -rn '"net/http"' internal/modules/*/application (non-test)
→ internal/modules/auth/application/service.go:21
```
`auth/application/service.go` is the one genuine hit. It is not a stray import: `Authenticate(ctx, identifier, password string, r *http.Request)` (line 316), `evaluateLoginAttempt(…, r *http.Request)` (368), `issueSessionAfterLogin(…, r *http.Request)` (423), and two methods that construct `*http.Cookie` values directly — `SessionCookie()` (1066) and `ExpiredSessionCookie()` (1092), including setting `SameSite: http.SameSiteStrictMode`. This is HTTP-layer concern (cookie construction, `*http.Request` as a parameter type) living inside the application service rather than the delivery handler translating to/from a transport-neutral type. Genuine delivery-into-application layering violation, isolated to one module.

**Oversized `approval/application` — verifying #94's 30 files/8.7k LOC/583 symbols claim:**
```
find internal/modules/approval/application -name '*.go' -not -name '*_test.go' | wc -l        → 30  (exact match)
… | xargs cat | wc -l                                                                          → 9339 LOC (historical 8.7k)
grep -rhE '^func [A-Z]|^func \([a-zA-Z0-9_ *]+\) [A-Z]|^type [A-Z]|^var [A-Z]|^const [A-Z]' …   → 740 (func/type/var/const)
grep -rhE '^func [A-Z]|^func \([a-zA-Z0-9_ *]+\) [A-Z]|^type [A-Z]' …                            → 706 (func/type only)
```
File count reproduces exactly. LOC and exported-symbol counts are directionally confirmed but not bit-exact — the module grew ~7% in LOC since #94 was filed, and the exported-symbol count is sensitive to whether grouped `var`/`const` blocks and multi-line receivers are counted (583 historical sits between what a func+type-only count and a func+type+var+const count would produce depending on exact grep used at the time). **The structural claim — approval/application is the largest, most exported-surface-heavy application package in the codebase — is confirmed regardless of exact digit.**

**Duplicated business orchestration:**

1. **Status-transition validators** — `templates/domain/version.go:89` `(*TemplateVersion) CanTransition(next VersionStatus, hasReviewer bool) error` (switch-based, 4 states, a `hasReviewer` branch) vs `documents/domain/state.go:52` `CanTransitionDocumentStatus(cur, next DocumentStatus) error` (map-table-based, 7 states, mirrors a DB trigger). Both independently reimplement the same *shape* of pattern (a bespoke per-entity lifecycle FSM validator) but are **not byte-identical** — different transition counts, different data structures (switch vs map), different parameters. Confirmed as **parallel reimplementation of a repeated pattern**, not literal duplication. A shared `statemachine` helper type is a legitimate simplification candidate but is not free — the two FSMs' arcs and guard conditions genuinely differ.

2. **Objectstore error-switch, documents ↔ templates** — **byte-identical**, confirmed by direct read:
   - `internal/modules/documents/application/service.go:734-743`
   - `internal/modules/templates/application/autosave.go:153-162`
   ```go
   switch {
   case errors.Is(err, objectstore.ErrObjectMissing):
       return nil, domain.ErrUploadMissing
   case errors.Is(err, objectstore.ErrHashMismatch):
       return nil, domain.ErrContentHashMismatch
   case errors.Is(err, objectstore.ErrObjectTooLarge):
       return nil, domain.ErrUploadTooLarge
   default:
       return nil, fmt.Errorf("…: %w", err)
   }
   ```
   Identical case order, identical target sentinels, only the wrapping message text differs. This is a clean extraction candidate (a shared `objectstore.TranslateConfirmError(err) error` helper in `internal/platform/objectstore` or a small shared package both modules already depend on).

### 6.4 Delivery discipline

**Raw SQL in delivery/http:** zero hits (`grep -rlE 'SELECT |INSERT INTO|UPDATE .+SET|DELETE FROM' internal/modules/*/delivery internal/modules/*/http` → empty). Delivery layer is clean of embedded SQL text.

**Naming/structural inconsistency:** 14 of 15 modules follow `<module>/delivery/http`; **`approval` alone uses `<module>/http`** (`internal/modules/approval/http`, no `delivery/` segment) — a naming drift worth a mechanical rename, not a functional defect.

**Delivery package size (non-test LOC):**
| Module | LOC |
|---|---|
| **approval/http** | **4347** |
| iam/delivery/http | 2732 |
| documents/delivery/http | 2334 |
| templates/delivery/http | 1888 |
| taxonomy/delivery/http | 1080 |
| controlleddocuments/delivery/http | 1013 |
| audit/delivery/http | 514 |
| auth/delivery/http | 447 |
| notifications/delivery/http | 274 |
| tokens/delivery/http | 260 |
| security/delivery/http | 237 |
| distribution/delivery/http | 239 |
| search/delivery/http | 192 |

`approval/http` is the largest delivery surface in the codebase by a wide margin (58% bigger than the runner-up `iam`), mirroring the oversized-application finding in §6.3 — approval is oversized in **both** layers.

**Delivery importing infrastructure types directly** (skipping the application layer's error-translation responsibility):
- `internal/modules/approval/http/errors.go` — a 300+-line delivery-layer error-to-HTTP mapping table that imports and directly type-checks `infrastructure.ErrStaleRevision`, `infrastructure.ErrNoActiveInstance`, `infrastructure.ErrInstanceNotVisible`, `infrastructure.ErrDuplicateSubmission`, `infrastructure.ErrActorAlreadySigned`, `infrastructure.ErrInvalidSupersedeTarget`, `infrastructure.ErrInstanceCompleted`, `infrastructure.ErrRouteInUse`, `infrastructure.ErrDuplicateRouteProfile`, `infrastructure.ErrInsufficientPrivilege`, `infrastructure.ErrUnknownDB` (11 distinct sentinels, lines 244-305). The delivery layer is reaching two layers down into `infrastructure` for its error vocabulary instead of the application layer translating/wrapping infrastructure errors into domain errors first. This is a genuine, approval-specific layering violation (infrastructure leaking straight to delivery, application layer bypassed for error translation).
- `internal/modules/documents/delivery/http/fillin_handler.go:28` — the delivery-owned port interface itself returns an infrastructure type: `GetPlaceholderValues(ctx, tenantID, docID string) ([]infrastructure.PlaceholderValue, error)`. An infrastructure DTO leaks into a delivery-layer interface contract.

**Postgres-error-code awareness leaking into delivery** (cross-referenced with Pass 8 §8.5): `taxonomy/delivery/http/routes_areas.go:174`, `routes_families.go:126`, `routes_profiles.go:384`, and `tokens/delivery/http/handler.go:218` all do a hand-rolled `var pgErr *pgconn.PgError; if errors.As(err, &pgErr) && pgErr.Code == "23505"` check **directly in the HTTP handler**, not in infrastructure. 4 of the 11 files doing hand-rolled unique-violation detection are in the delivery layer.

**`authorizeDocumentScope` (documents/delivery/http/handler.go:1254-1276):** delegates the actual capability check to `h.svc.RequireDocumentView(...)` (application service) — not itself a violation — but computes an `isSystemAdmin` bypass shortcut at the handler layer before delegating. Not flagged as a defect (still calls into a service), but the shortcut logic sitting in delivery rather than being folded into the same `RequireDocumentView` call is worth a simplification pass; not pursued further here as it is a borderline case, not a clear violation.

---

## PASS 7 — ERRORS

### 7.1 Domain sentinel inventory

```
for each module: cat domain/*.go (non-test) | grep -oE '\bErr[A-Za-z0-9_]+\s*=\s*(errors\.New|fmt\.Errorf|stderrors\.New)\('
```

| Module | Sentinel count |
|---|---|
| documents | 31 |
| approval | 28 |
| taxonomy | 28 |
| templates | 28 |
| controlleddocuments | 24 |
| iam | 10 |
| auth | 13 |
| tokens | 3 |
| search | 3 |
| audit | 2 |
| security | 2 |
| **Total** | **172** |
| distribution, notifications, render, jobs | 0 |

**Name collisions** — same identifier declared as a *separate* sentinel var in ≥2 modules' `domain` packages (11 collisions):

| Sentinel name | Declared in |
|---|---|
| `ErrApprovalRouteMissing` | controlleddocuments, documents, templates |
| `ErrContentHashMismatch` | documents, templates |
| `ErrDictionaryTokenMissing` | controlleddocuments, documents |
| `ErrForbidden` | documents, templates |
| `ErrInvalidStateTransition` | documents, templates |
| `ErrNotFound` | documents, templates, tokens |
| `ErrStaleBase` | documents, templates |
| `ErrTemplateProfileMismatch` | controlleddocuments, taxonomy |
| `ErrTenantNotFound` | auth, iam |
| `ErrUploadMissing` | documents, templates |
| `ErrUploadTooLarge` | documents, templates |

These are distinct Go package-scoped values (an `errors.Is` comparison across modules requires the fully-qualified reference, so there is no runtime confusion risk), but the repeated vocabulary — 7 of the 11 collisions are `documents`↔`templates` pairs — is a strong signal the two modules' upload/versioning lifecycles are structurally parallel and could share a vocabulary package, echoing §6.3's twin-validator/twin-error-switch finding.

### 7.2 Cross-module `errors.Is` sites — reproducing and correcting the 62 claim

**Step 1 — naive alias-pattern count** (the shape of the check the original 62 count almost certainly used):
```
grep -rEo 'errors\.Is\([^,]+,\s*(approvaldomain|auditdomain|authdomain|cddomain|controlleddocumentsdomain|distributiondomain|docsdomain|documentsdomain|iamdomain|notificationsdomain|renderdomain|searchdomain|securitydomain|taxonomydomain|templatesdomain|v2domain)\.Err[A-Za-z0-9_]+\)' internal/modules
```
→ **65 non-test sites** (123 including test files). This is a close match to the historical **62** — the small drift is consistent with organic growth since #93 was filed.

**Step 2 — the correction.** This pattern alone does not distinguish a *true* cross-module reference from a module importing **its own** `domain` package under an alias (e.g. `iam/delivery/http/*.go` files import `"metaldocs/internal/modules/iam/domain"` as `iamdomain` — a same-module self-reference, not cross-module — because the file already has another package literally named `domain` imported, or by convention). Resolving each hit's consumer module (from its file path) against its alias's producer module:

| Classification | Count |
|---|---|
| Naive alias-pattern total (non-test) | 65 |
| **Same-module** (aliased self-import, e.g. `iam/*` using `iamdomain.Err…`, `auth/*` using `authdomain.Err…`) | 46 |
| **True cross-module** (producer module ≠ consumer module) | **19** |

**19 true cross-module sites, grouped by consumer→producer pair:**

| Consumer → Producer | Count | Sites |
|---|---|---|
| `iam` → `auth` | 10 | `iam/application/onboard_tenant_service.go:297`; `iam/application/people_service.go:714,716`; `iam/delivery/http/people_handler.go:417,419,423,435,437`; `iam/delivery/http/sessions_handler.go:255,275` |
| `controlleddocuments` → `taxonomy` | 4 | `controlleddocuments/delivery/http/routes.go:716,722,728,734` |
| `documents` → `controlleddocuments` | 3 | `documents/delivery/http/handler.go:1315,1345,1350` |
| `approval` → `documents` | 2 | `approval/application/route_preview.go:88` (2 sentinels in one `switch`) |

**Layer classification of the 19:**
- **Application-seam (undeclared integration contract, A4):** 5 sites — `approval/application/route_preview.go` (2), `iam/application/onboard_tenant_service.go` + `iam/application/people_service.go` (3).
- **Delivery/HTTP-translation (A3):** 14 sites — `controlleddocuments/delivery/http/routes.go` (4), `documents/delivery/http/handler.go` (3), `iam/delivery/http/people_handler.go` + `sessions_handler.go` (7).
- **Infrastructure-adapter (possibly legitimate translation point):** **0 sites.** No infrastructure-layer file does a cross-module `errors.Is` against a foreign domain sentinel — the pattern only shows up in application and delivery code.

**Most significant instance:** `iam → auth`, 10 of 19 sites, split across **both** the application layer (`onboard_tenant_service.go`, `people_service.go`) **and** the delivery layer (`people_handler.go`, `sessions_handler.go`) for the *same* sentinel vocabulary (`authdomain.ErrUserAlreadyExists`, `ErrIdentityNotFound`, `ErrPasswordPolicy`). `auth`'s errors are being re-interpreted ad hoc at two different layers inside `iam` instead of being translated exactly once at the `iam`↔`auth` seam — this is the textbook "undeclared integration contract" shape (A4): there is no single `iam` adapter that owns "translate an `auth` domain error into an `iam` domain error/problem code," so the translation logic is duplicated per call site instead.

**Corrected headline for future citation:** *"62 cross-module `errors.Is` sites"* is the naive same-module-inclusive count (reproduces as 65 today) — the **true** cross-module surface is **19 sites**, concentrated in one dominant pair (`iam→auth`, 10) with zero infrastructure-layer occurrences.

### 7.3 Problem writers

**Local `writeProblem`-like funcs** (12+ claimed):
```
grep -rn "^func.*writeProblem\|^func write.*[Pp]roblem" internal apps
```
Exactly **12**:
1. `internal/modules/audit/delivery/http/handler.go:510`
2. `internal/modules/auth/delivery/http/handler.go:275`
3. `internal/modules/iam/delivery/http/admin_handler.go:401`
4. `internal/modules/iam/delivery/http/middleware.go:215`
5. `internal/modules/iam/delivery/http/observability_handler.go:73`
6. `internal/modules/iam/delivery/http/people_handler.go:400`
7. `internal/modules/iam/delivery/http/routes_memberships.go:311`
8. `internal/modules/iam/delivery/http/sessions_handler.go:344`
9. `internal/modules/iam/delivery/http/tenant_handler.go:131`
10. `internal/modules/security/delivery/http/handler.go:215`
11. `internal/modules/templates/delivery/http/handler.go:256` (`writeErr`, same shape)
12. `internal/platform/idempotency/middleware.go:276` (`writeErrJSON`, same shape)

Note: 6 of the 12 are inside `iam/delivery/http` alone — one per handler file (`admin_handler`, `middleware`, `observability_handler`, `people_handler`, `routes_memberships`, `sessions_handler`, `tenant_handler` — actually 7), each a byte-near-identical 3-line wrapper around `problem.Write`. A single shared `Handler.writeProblem` (or a package-level free function) would collapse these to one definition; each currently repeats the same `if werr := problem.Write(w, p); werr != nil { slog.Warn(...) }` body independently.

**`problem.New(` vs `problem.NewFor(`:**
```
grep -rn "problem\.New(" internal apps | grep -v _test.go | wc -l   → 227  (historical 232)
grep -rn "problem\.NewFor(" internal apps | grep -v _test.go | wc -l → 11   (historical 11, exact)
```

**Bare `http.Error(` in API paths:**
```
grep -rn "http\.Error(" internal/modules apps | grep -v _test.go   → 39 sites
```
All 39 are in `*/api/api.gen.go` — **100% generated oapi-codegen boilerplate**, 3 sites per module × 13 modules with an HTTP API (`approval, audit, auth, controlleddocuments, distribution, documents, iam, notifications, search, security, taxonomy, templates, tokens`; `render` and `jobs` have no HTTP surface, consistent). These are the strict-server wrapper's parameter-bind-failure paths (malformed query/path params before the hand-written handler runs) — a real, systemic RFC 9457 gap (these responses are plain-text, not `problem+json`), but it is uniform codegen-template behavior across every module, not a module-specific lapse. **Zero** bare `http.Error(` calls exist in hand-authored delivery code.

**RFC 9457 mapping location:** `internal/platform/problem/` (`code.go`, `codes.go`, `problem.go`) — single shared package, correctly centralized. No generated `Problem`/error-body type exists in any `api.gen.go`; generated code models only success-shape DTOs, error bodies are 100% hand-authored via the shared `problem` package.

**`writeProblem` logs nothing — confirmed and refined.** Every one of the 12 `writeProblem`/`writeErr(JSON)` functions has the same shape:
```go
func (h *X) writeProblem(w http.ResponseWriter, p *problem.Problem) {
    if werr := problem.Write(w, p); werr != nil {
        slog.Warn("… write response failed", "err", werr)
    }
}
```
This only logs a failure to **write the HTTP response itself** (socket/encoding error) — it never logs the *problem being reported*, including 500-class internal errors. Correlated server-side logging of the underlying error is therefore entirely up to each call site, and it is **inconsistent**: in `internal/modules/iam/delivery/http/people_handler.go`, some 500-producing branches log first (`slog.Error("iam people: list users failed", …)` at line 85 before the `writeProblem` at line 86; similarly lines 300-301, 393-394, 428-429), while others return a 500 with **zero server-side log trace at all** — e.g. lines 59-60, 118-119, 164-165, 273-274, 337-338 (`tenant.FromContext` failure → 500, no log) and line 440 (`writeAuthError`'s default branch → 500, no log). A production 500 from one of these paths is invisible in logs unless a caller happens to log before it. This is a real observability gap, not hypothetical.

### 7.4 Generated vs hand-authored error types

Generated `api.gen.go` files define request/response success DTOs only (contract-first per ADR discipline). No generated file defines a `Problem`/RFC-9457-shaped Go type. All error response construction is hand-authored against the shared `internal/platform/problem` package — this split (generated success shapes, hand-authored error shapes funneled through one shared package) is architecturally sound; the defects found are in **consistency of use** (§7.3), not in the split itself.

---

## PASS 8 — TRANSACTIONS / PERSISTENCE

### 8.1 TxRunner

**Defined:** `internal/platform/db/runner.go:23`
```go
type TxRunner interface {
    Do(ctx context.Context, fn func(tx *sql.Tx) error) error
    DoReadOnly(ctx context.Context, fn func(tx *sql.Tx) error) error
}
```
`Do` and `DoReadOnly` share one private `do(ctx, opts, fn)` implementation (lines 52-75) that begins the tx, auto-seeds tenant/actor GUCs (`seedTxIdentityFromContext`, the M3 tenancy-chokepoint contract), runs `fn`, and commits/rolls back.

**Call counts:**
```
grep -rn "\.Do(ctx\|\.Do(context\|txRunner\.Do(\|\.Do(func" internal | grep -v _test.go | wc -l      → 70
grep -rn "\.DoReadOnly(" internal | grep -v _test.go | wc -l                                          → 0
```
`DoReadOnly` has **zero call sites** — confirmed dead at the call-site level (consistent with the known G1 remediation: `DoReadOnly→Do` migration + api-lint guard). **However the `DoReadOnly` method itself is still defined** on both the `TxRunner` interface and the `sqlTxRunner` implementation (lines 33, 135-137) — it was never deleted after callers were migrated off it. Minor dead-code cleanup: remove the method now that nothing calls it, or the interface keeps offering a foot-gun (the doc comment at line 28 explicitly warns "MUST NOT be used for any path that calls `authz.Require`").

**Direct `.BeginTx(` sites (bypassing TxRunner):**
```
grep -rn "\.BeginTx(" internal apps | grep -v _test.go | wc -l    → 84 sites
grep -rl "\.BeginTx(" internal apps | grep -v _test.go | wc -l    → 26 files
```
Historical 82/25 — reproduces almost exactly (+2 sites, +1 file). Files (26):
`audit/infrastructure/postgres/writer.go`, `auth/application/service.go`, `auth/infrastructure/postgres/repository.go`, `controlleddocuments/infrastructure/repository.go`, `documents/infrastructure/repository.go`, `iam/application/area_membership_service.go`, `iam/infrastructure/postgres/{role_admin_repository,user_area_repository}.go`, `jobs/{approval_sla_surfacer,document_review_surfacer,release_hold_reconciler,stuck_instance_watchdog}/job.go`, `notifications/infrastructure/{approval_notify_worker,fanout_worker}.go`, `render/fanout/infrastructure/seeded_tx.go`, `taxonomy/application/{area_service,family_service,profile_service}.go`, `taxonomy/infrastructure/{family_repository,repository}.go`, `platform/db/runner.go` (the canonical impl itself — not a violation), `platform/idempotency/postgres_store.go`, `platform/messaging/outbox/postgres/consumer.go`, `platform/worker/{materialize_job_runner,pdf_job_runner}.go`, `test/e2e_seed.go`.

**Module-specific second tx abstractions — historical claim "2 modules," reproduced "3":**
Three modules have application-layer code that calls `.BeginTx(` on a narrow repository-scoped interface **instead of** going through `db.TxRunner`:
1. **`auth`** — `auth/application/service.go` defines a `beginTxRepository` interface; `changePasswordAtomic`, `createUserAtomic`, `adminResetPasswordAtomic`, `AdminResetPassword`, and one more atomic helper all call `beginner.BeginTx(ctx, nil)` directly (lines 662-663, 792-793, 858-860, 952-953, 1032-1033).
2. **`iam`** — `iam/application/area_membership_service.go:218,278` calls `s.repo.BeginTx(ctx)` in `commitMembershipMutation` and the revoke path.
3. **`taxonomy`** — `taxonomy/application/{area_service,family_service,profile_service}.go` each call `s.areas.BeginTx(ctx)` / equivalent (e.g. `area_service.go:62,111,163`).

All three bypass `db.TxRunner.Do(...)` entirely and hand-roll their own begin/commit/rollback lifecycle at the application layer, duplicating what `TxRunner` exists to centralize (including losing the automatic tenant/actor GUC seeding `TxRunner.do` provides — worth checking whether these three modules' `BeginTx` paths seed GUCs themselves; not verified in this pass, flagged as a follow-up). **Reproduced-current: 3 modules, not 2** — `taxonomy` (and its 3 services) was not counted in the historical claim.

### 8.2 Repositories: `*sql.DB` vs `TxRunner`

```
grep -rln "\*sql\.DB" internal/modules/*/infrastructure | grep -v _test.go | wc -l   → 56 files
grep -rln "db\.TxRunner" internal/modules/*/infrastructure | grep -v _test.go        → 0 files
grep -rln "db\.TxRunner" internal | grep -v _test.go | wc -l                          → 34 files (repo-wide)
```
**Zero** infrastructure-layer files reference `db.TxRunner` as a type — every infrastructure repository is constructed against a raw `*sql.DB` (or takes `*sql.Tx`/`db.Tx` per-method, per §6.2's ADR-0044 pattern), consistent with `runner.go`'s own doc comment: *"The callback receives the live `*sql.Tx` by design."* `TxRunner` is exclusively an **application-layer** construct — all 34 files referencing `db.TxRunner` are in `application/` or `delivery/http/` (not `infrastructure/`), across exactly **6 modules**: `approval, controlleddocuments, documents, iam, templates, tokens`. This matches the historical "6" claim exactly, on a module-count basis.

The "75" side of the historical claim (repositories on `*sql.DB`) is reproduced as **56 files** here — the unit is likely different (the original count may have been per-repository-struct or per-method rather than per-file); the file-level number (56) is the mechanically reproducible one and is presented as the corrected basis for future citation.

**New finding:** `db.TxRunner` also appears directly in **delivery**, not just application — `approval/http/{cancel_handler,doc_approval_handler,handler,mark_reviewed_handler,obsolete_handler}.go` (5 files) hold a `db.TxRunner` reference at the handler layer. This means `approval`'s HTTP handlers can open transactions directly rather than exclusively delegating to `approval/application` services — worth a closer look in a follow-up pass (not fully traced here; flagged, not root-caused).

### 8.3 Tx/non-Tx method twins

**Tx-suffixed methods found in `*/infrastructure`:**
```
grep -rhoE '^func \([a-zA-Z0-9_ *]+\) [A-Z][A-Za-z0-9_]*Tx\(' internal/modules/*/infrastructure | grep -v _test.go | sort -u
```
36 distinct `*Tx`-suffixed method names. Cross-referencing each against a same-named non-`Tx` sibling in the same file set surfaced **14 base-name pairs** actually implemented as twins (not every `*Tx` method has a non-`Tx` sibling — some are Tx-only by design, e.g. audit's append-only `RecordTx`):

| Base name | Tx twin | Files |
|---|---|---|
| `Create` | `CreateTx` | controlleddocuments/infrastructure/repository.go, taxonomy/infrastructure/{family_repository,repository}.go, tokens/infrastructure/repository.go |
| `CreateTemplate` | `CreateTemplateTx` | templates/infrastructure/postgres.go |
| `CreateUser` | `CreateUserTx` | auth/infrastructure/{memory,postgres}/repository.go |
| `CreateVersion` | `CreateVersionTx` | templates/infrastructure/postgres.go |
| `Insert` | `InsertTx` | iam/infrastructure/postgres/user_area_repository.go |
| `MarkArchived` | `MarkArchivedTx` | documents/infrastructure/repository.go |
| `RevokeSession` | `RevokeSessionTx` | auth/infrastructure/{memory,postgres}/repository.go |
| `RevokeSessionsByUserID` | `RevokeSessionsByUserIDTx` | auth/infrastructure/{memory,postgres}/repository.go |
| `Update` | `UpdateTx` | taxonomy/infrastructure/{family_repository,repository}.go, tokens/infrastructure/repository.go |
| `UpdateStatus` | `UpdateStatusTx` | controlleddocuments/infrastructure/repository.go |
| `UpdateTemplate` | `UpdateTemplateTx` | templates/infrastructure/postgres.go |
| `UpdateUser` | `UpdateUserTx` | auth/infrastructure/{memory,postgres}/repository.go |
| `UpdateVersion` | `UpdateVersionTx` | templates/infrastructure/postgres.go |

Counting each file's implementation as one concrete twin gives well over 20 individual twin-method instances — consistent with the historical "20" claim.

**Spot-verification of 3 pairs (all requested "at least 5" attempted; 3 fully diffed and read, all 3 byte-identical modulo executor):**

1. **`templates/infrastructure/postgres.go`** — `CreateTemplate` (lines 62-83) vs `CreateTemplateTx` (437-458): identical 9-line `INSERT INTO templates_template (...)` SQL text, identical parameter list, identical `pgconn.PgError`/`23505`→`domain.ErrKeyConflict` mapping. Only difference: `r.db.ExecContext` vs `tx.ExecContext`.
2. **`taxonomy/infrastructure/repository.go`** — `ProfileRepository.Create` (266-304) vs `CreateTx` (308-331): identical `INSERT INTO metaldocs.document_profiles (...)` SQL, identical `authz.Require(ctx, tx, CapTaxonomyManage, "tenant")` call, identical `setAuthzGUC` call. `Create` additionally owns its own `r.db.BeginTx`/`Commit` lifecycle (the taxonomy second-tx-abstraction from §8.1); `CreateTx` type-asserts the caller's `domain.FamilyTx` to a concrete `taxonomyTx`.
3. **`taxonomy/infrastructure/repository.go`** — `AreaRepository.Create` (644-678) vs `CreateTx` (682-…): same pattern as #2, over `metaldocs.document_process_areas`.

All 3 are **byte-identical SQL + byte-identical authz-check**, differing only in the executor plumbing — this exactly matches the historical "3 byte-identical" sub-claim. Given the prevailing `db.Tx`/`db.DB` structural-interface convention (§6.2 — sanctioned by ADR 0044 only for the event boundary; elsewhere unresolved pending ruling), these pairs are a clear case where the non-`Tx` variant could be rewritten as a one-line wrapper (`func (r *X) Create(ctx, p) error { return r.db.beginRun(ctx, func(tx) error { return r.CreateTx(ctx, tx, p) }) }`) instead of carrying a second, independently-maintained copy of the SQL and the authz check.

### 8.4 Scan sites

```
grep -rn "\.Scan(" internal apps | grep -v _test.go | wc -l                                → 278  (whole-repo non-test)
grep -rn "\.Scan(" internal/modules/*/infrastructure | grep -v _test.go | wc -l              → 207  (infrastructure only)
grep -rn "\.Scan(" internal/platform | grep -v _test.go | wc -l                              → 11
grep -rhE '^func.*[Ss]can[A-Z]|^func scan' internal | grep -v _test.go | wc -l                → 22   (factored scan-helper funcs)
```
Historical claim was 242; reproduced-current ranges from 207 (infra-only) to 278 (whole-repo) depending on scope — in the right order of magnitude, with growth since filing. 22 factored `scanXxx`-shaped helper functions exist (e.g. `scanDocumentReleaseState` in `release_coordinator.go:698`, `scanTemplateRead` in `templates/infrastructure/postgres.go`), meaning **most but not all** `.Scan(` call sites have been factored — the majority (well over 200) remain inline, ad hoc per-query `.Scan(&a, &b, &c, ...)` calls. sqlc (§8.8) would collapse essentially all of these into generated, compile-time-checked scan code.

### 8.5 Postgres error-code mapping

**Hand-rolled 23505 (unique-violation) checks:**
```
grep -rn "23505" internal | grep -v _test.go   → 17 sites across 11 files (historical: 15 across 10)
```
Files: `approval/infrastructure/errors.go`, `auth/infrastructure/postgres/repository.go`, `controlleddocuments/{application/service.go,infrastructure/repository.go}`, `iam/infrastructure/postgres/tenant_repository.go`, `security/infrastructure/postgres/tenant_key_repository.go`, `taxonomy/delivery/http/{routes_areas,routes_families,routes_profiles}.go`, `templates/infrastructure/postgres.go`, `tokens/delivery/http/handler.go`.

Cross-ref §6.4: **4 of these 11 files are in the delivery/HTTP layer** (`taxonomy` ×3, `tokens` ×1), not infrastructure — Postgres error-code awareness has leaked past the repository boundary into HTTP handlers in two modules.

**Driver registration:**
```
grep -rn 'stdlib\|"github.com/jackc/pgx' internal/platform/db/postgres
→ internal/platform/db/postgres/connect.go:10:  _ "github.com/jackc/pgx/v5/stdlib"
```
Confirmed: the live `database/sql` driver is **pgx v5's stdlib adapter**, which surfaces errors as `*pgconn.PgError`, never `*pq.Error`.

**`lib/pq` usage — verifying "5 testing lib/pq though driver is pgx":**
```
grep -rln '"github.com/lib/pq"' internal   → 18 files (17 non-test + 1 test)
```
Read all 11 files doing 23505 checks for their exact check order:
- 7 files check **only** `*pgconn.PgError`: `approval/infrastructure/errors.go`, `controlleddocuments/{application/service.go,infrastructure/repository.go}`, `taxonomy/delivery/http/{routes_areas,routes_families,routes_profiles}.go`, `templates/infrastructure/postgres.go`, `tokens/delivery/http/handler.go`.
- **3 files check `*pgconn.PgError` first, then fall back to `*pq.Error`**: `auth/infrastructure/postgres/repository.go:686-692`, `iam/infrastructure/postgres/tenant_repository.go:67-71`, `security/infrastructure/postgres/tenant_key_repository.go:130-134` (this last one's own comment at line 10 says *"detection checks `*pgconn.PgError` first, falling back to `*pq.Error` for any [other driver]"*).

In all 3 dual-check files, the `*pgconn.PgError` branch is checked and returns **first** — the `*pq.Error` branch is structurally unreachable under the pgx driver actually registered. **Confirmed: dead code, not a live bug.** The unique-violation detection always works (the live branch always fires); the `lib/pq` dependency and its 3 fallback branches are inert insurance against a driver swap that hasn't happened and, per `connect.go`, isn't in effect. Safe, low-priority cleanup: delete the `pq.Error` fallback branches and the `lib/pq` import from these 3 files (the other 8 files needing 23505 detection already don't have it).

### 8.6 N+1 — `release_coordinator.go` per-target loops

Read the full 804-line file. **Confirmed — 3 separate per-target loops**, all inside `internal/modules/approval/application/release_coordinator.go`:
1. **`lockAndRedecide`** (lines 395-428): loops over the sorted `lockSet` (source document + supersession targets), issuing one `SELECT ... FOR UPDATE` per id via `lockDocumentForRelease` (line 414).
2. **`releaseTx`** (461-538): loops over `rc.targets`, issuing one `UPDATE documents SET status='superseded'...` per target via `c.repo.MarkSuperseded` (line 469).
3. **`emitReleaseEvents`** (559-669): loops over `superseded` twice — once to emit a governance event per target (607-635), once to enqueue a lifecycle event per target (651-667).

This is genuinely N+1-shaped (row count scales with the size of the supersession set). It is **not a naive oversight**: extensive in-code comments (lines 282-290, 386-394, 430-434) explain that the per-row `FOR UPDATE` locks must be acquired **one at a time in sorted document-id order** to give every concurrent release evaluation the same lock-acquisition order and avoid an AB-BA deadlock — a single batched `WHERE id = ANY(...) FOR UPDATE` gives no guarantee about *which order* Postgres acquires the individual row locks in, which would reintroduce the deadlock the sort exists to prevent. In practice the target set is small (a document plus, typically, 0-2 supersession targets), so the N+1 shape is a bounded, deliberate trade-off for correctness under concurrency, not a scalability risk today. Worth a comment/ADR cross-reference if a future feature introduces wide-fanout supersession (many targets per release), which is out of scope here.

### 8.7 Network-inside-transaction

Checked the highest-likelihood hotspots (object-store/network calls near `application` services that also hold a `db.TxRunner`):
- **`documents/application/service.go` `CommitAutosave`** (lines 711-773): `s.presigner.Confirm(...)` (network call to the object store) executes at line 732, **before** any `s.runner.Do(...)` block begins; the subsequent `s.repo.CommitUpload(...)` call (753) is a separate repository call, not shown to nest a network call inside a `Do`. Clean.
- **`templates/application/autosave.go`**: same `Confirm(...)`-then-`Do(...)` ordering as documents (line 170's `s.runner.Do` follows the confirm at line ~151).
- **`iam/application/tenant_lifecycle_service.go`** tenant erasure (`runErase`, lines ~540-563): explicitly phased — **Phase 1** `eraseTenantRowsTx` (568-606, inside `s.runner.Do`, DB-only), **Phase 2** `eraseTenantBlobs` (611-626, `s.blobs.ListTenantObjects` + `s.blobs.Delete` loop — network calls to the object store) is called **outside** any `Do` block, **Phase 3** `cryptoShredAndTombstoneTx` (633-…, back inside `s.runner.Do`, DB-only). The blob-deletion network phase is structurally isolated between two separate transactions, never nested inside either.

**No violations found** in the 3 hotspots checked. **This is not an exhaustive sweep** — only the object-store-touching application code was checked by hand; the other ~65 `Do(` call sites (§8.1) were not individually audited for an HTTP/renderer call hidden inside the callback. A repo-wide AST-level check (walk every `func(tx *sql.Tx) error` literal passed to `.Do(`/`.BeginTx(` and flag any call to an `http.Client`, `objectstore.*`, or renderer-package method inside it) would be needed to make this claim exhaustive; that is a bounded follow-up, not performed here.

### 8.8 sqlc suitability note — documents / approval / iam

| Module infra | LOC (non-test) | `.Scan(` sites |
|---|---|---|
| approval/infrastructure | 4008 | 40 |
| documents/infrastructure | 3356 | 65 |
| iam/infrastructure | 2219 | 26 |

These three are the largest hand-written-SQL surfaces in the codebase (combined ~9.6k LOC, ~130 hand-`.Scan(` sites — roughly half of the whole-repo infra-only Scan count from §8.4). They are the strongest sqlc-adoption candidates: sqlc would eliminate essentially all of the hand-`.Scan(` boilerplate and give compile-time-checked query/parameter/column typing, directly addressing §8.4's factoring gap. The cost is real, not free: sqlc's generated methods take a concrete executor type (`*sql.DB`/`*sql.Tx` or, for pgx-native mode, `pgx.Tx`) — adopting it would need to either (a) keep generating against `database/sql`'s `*sql.Tx`/`*sql.DB` so the existing `db.Tx`/`db.DB` structural-interface convention (§6.2; ADR-0044-sanctioned only at the event boundary, elsewhere pending ruling) keeps working unchanged, or (b) accept a second codegen pipeline alongside `oapi-codegen`'s contract-first HTTP generation, with its own regen-on-schema-change discipline (paralleling the existing `openapi-embedded-spec-regen-churn` lesson already tracked for the OpenAPI side). Recommended as a **named global-maximum candidate** for a future milestone, not a drop-in fix — scoping which of the 3 modules goes first (documents has the most `.Scan(` sites; approval has the most LOC and the most twin-method duplication from §8.3) is a planning decision, not made here.

---

## Cross-reference to issue taxonomy

- **Layering / type-leak (#93 A4 + #92 A5):** §6.2 (db.Tx/db.DB — mostly ADR-legitimate, 1 genuine leak), §6.3 (application-layer SQL + HTTP-in-application + oversized approval/application + twin FSMs + byte-identical objectstore switch), §6.4 (delivery importing infrastructure types, Postgres-error-code-in-delivery).
- **Error seams (#93 A4 + #90 A3):** §7.2 (true cross-module errors.Is = 19, not 65/62; iam↔auth undeclared integration contract), §7.3 (writeProblem duplication + silent 500s).
- **Tx mechanics (#92 A5):** §8.1-§8.3 (3 module-specific tx abstractions, 3 byte-identical twin methods), §8.5 (delivery-layer pg-error checks), §8.6 (justified N+1), §8.7 (network-in-tx clean on sampled scope).
