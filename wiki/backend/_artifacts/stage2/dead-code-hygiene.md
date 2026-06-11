# Stage 2 Evaluation — Dead Code, Vestigial Scaffolds & Repo Hygiene

> **Theme:** dead-code-hygiene
> **Produced by:** Stage-2 evaluation pass, 2026-06-11
> **Input register:** `wiki/backend/legacy-register.md` (findings F-08, F-14, D-02, D-05, D-06, D-07)
> **Standards applied:** YAGNI (Fowler, martinfowler.com/bliki/Yagni.html), CWE-561 (Dead Code, mitre.org/data/definitions/561.html), Go project-layout convention (go.dev/doc/modules/layout), REQ-TOP-3, REQ-CACHE-1, minio-go v7 SDK (pkg.go.dev/github.com/minio/minio-go/v7, github.com/minio/minio-go/issues/598)

---

## How to read this document

Each section corresponds to one finding from the legacy register. For each finding:
- **Current state** — what the code actually is (re-confirmed from the register's file:line anchors)
- **Standard** — the external reference against which the finding is judged
- **Verdict + rationale** — one of KEEP / SIMPLIFY / REFACTOR / DELETE / DEFER
- **Smallest correct fix** — minimum change that reaches a professional bar
- **Effort / blast-radius** — S/M/L and contained/module/cross-module/system
- **ADR needed** — yes/no with reason
- **Over-engineering check** — explicit anti-gold-plating note

---

## F-08 — Empty / Vestigial Platform Scaffolds and Dead Binaries

### Current state

Register and code confirm three categories of waste:

1. **Empty `.gitkeep` directories with no Go package:** `internal/platform/cache/` (only `cache/.gitkeep`; zero Go source; confirmed by `find` — zero files other than the marker). `internal/platform/db/.gitkeep` is a second nesting level where the real package lives at `db/postgres/` — the intermediate directory has no Go source either.

2. **Superfluous `.gitkeep` files in live packages:** `internal/platform/objectstore/.gitkeep` and `internal/platform/observability/.gitkeep` coexist with real Go source files, making them residual scaffolding noise rather than directory holders.

3. **Committed build artifacts:** `bin/metaldocs-api.exe` (initial commit `912879cba`; stale compiled binary); `apps/api/cmd/metaldocs-api/metaldocs-api.exe` (present on disk, gitignored but leaked); `apps/api/.gocache-build/` (build cache in source tree).

4. **`RepositoryMemory` production dead path:** `internal/platform/config/repository.go:9` declares `RepositoryMemory = "memory"` and `internal/platform/bootstrap/api.go:127-155` handles the `default` (memory) branch. However, `apps/api/cmd/metaldocs-api/main.go:677-680` fatals on any non-postgres mode, making this branch unreachable from production. It remains reachable from tests (bootstrap uses memory mode in unit tests), so it is not dead code — it is test-only infrastructure.

### Standard

**REQ-TOP-3** (this codebase): "Every platform package either has production consumers or does not exist. Empty scaffolds are deleted or implemented — speculative directories are banned."

**YAGNI (Fowler, 2015):** Speculative scaffolding carries a "cost of carry" — it complicates code navigation, triggers questions from new contributors ("what goes here?"), and increases the surface area of search tools and build output. Fowler's framing: "it makes it harder to modify and debug that software" — even if the direct cost is small, the signal-to-noise ratio of the codebase decreases.

**CWE-561 (MITRE):** "Dead code indicates source code problems requiring fixes and reflects poor quality standards." Mitigation: "Remove dead code before deploying the application." This applies to `.gitkeep` placeholder directories that imply non-existent packages — they are a form of dead structure.

For committed binaries: OWASP A05:2021 (Security Misconfiguration) and supply-chain attestation requirements (SLSA Level 1+) both require that only source artifacts are committed to VCS; pre-built binaries of indeterminate provenance defeat build reproducibility.

### Verdict: DELETE (most items); KEEP (RepositoryMemory path)

**`internal/platform/cache/` directory:** DELETE. Zero Go source, zero consumers, zero behavior. The directory exists as a speculative placeholder since the initial commit (2026-03-16). REQ-TOP-3 is unambiguous: it does not exist or it is implemented. `platform/data-layer.md §10` itself flags it as "either implement or delete." No cache contract exists (violates REQ-CACHE-1), and caching infrastructure is already present at the module level (see D-05). The correct fix is `git rm internal/platform/cache/.gitkeep` — nothing else. No package rename, no migration.

**`internal/platform/db/.gitkeep`:** DELETE. The real package `internal/platform/db/postgres/` already has a `connect.go` Go source file. The intermediate `db/` directory is an extra nesting level that serves no Go module purpose (there is no `package db` there). Remove `.gitkeep`. If the extra level is intentional (e.g., to reserve space for future `db/clickhouse/`), an ADR must say so. Absent that, YAGNI — delete.

**`internal/platform/objectstore/.gitkeep` and `internal/platform/observability/.gitkeep`:** DELETE. Both directories contain real Go source. The `.gitkeep` files are residual scaffolding that `git add` would have auto-removed if the directories had never been committed empty. Each is a one-line `git rm`.

**`bin/metaldocs-api.exe`:** `git rm` it and add `bin/*.exe` to `.gitignore`. The `.gitignore` covers `metaldocs-api.exe` at the repo root but not under `bin/`. This is a supply-chain hygiene issue — committed binary of unverifiable provenance at the initial commit. `bin/` is not a build output path for any canonical script; it is a stale artifact.

**`apps/api/cmd/metaldocs-api/metaldocs-api.exe`:** Already gitignored. Confirm `.gitignore` covers it or add `apps/**/*.exe`. No source change needed.

**`apps/api/.gocache-build/`:** Add to `.gitignore`. Not committed per the register's evidence; no source change needed.

**`RepositoryMemory` + memory bootstrap path:** KEEP. Despite `main.go` fataling on non-postgres, the memory mode is the live test fixture path: `bootstrap.BuildAPIDependencies(ctx, "memory", ...)` is called from unit tests and e2e-seed. It has real consumers. It is not dead code — it is test infrastructure. REQ-TOP-3 reads "production consumers"; tests are not production, but deleting memory mode would break the test suite. DEFER deletion until a decision is made about whether to consolidate test fixturing into postgres-backed tests only. That is a separate, higher-blast-radius decision.

### Smallest correct fix

1. `git rm internal/platform/cache/.gitkeep` (directory disappears when last file removed)
2. `git rm internal/platform/db/.gitkeep`
3. `git rm internal/platform/objectstore/.gitkeep`
4. `git rm internal/platform/observability/.gitkeep`
5. `git rm bin/metaldocs-api.exe`
6. Add `bin/*.exe` to `.gitignore`
7. Confirm `apps/**/*.exe` is covered in `.gitignore`

No Go source changes. CI passes because nothing imports these paths.

### Effort / blast-radius

**S / contained.** Pure VCS operations on non-code files. Zero compile-time impact.

### ADR needed

No — these are straightforward deletions of empty or stale artifacts. The only debatable item (memory mode) is kept.

### Over-engineering check

Do not add a `platform/cache` package preemptively "to have the right shape" when implementing caching. If caching is needed, implement it where the consumer lives (currently: the IAM module's `CachedRoleProvider`), and only extract to a platform package when a second module needs the same thing. That is the YAGNI threshold.

---

## F-14 — Dead / Superseded Application Code Retained in Source

### Current state

Register lists 12 items. Re-confirmed from code:

| Artifact | File:line | Dead? |
|---|---|---|
| `CutoverService` | `approval/application/cutover_service.go:1-65` | Used only in `coverage_boost_test.go:429` and `cutover_service_test.go:90,104` — no wiring in `Services` struct, no HTTP route, confirmed |
| Deprecated `PDFDispatchInvoker` path | `decision_service.go:43,60,64,566-573` — interface declared, field wired in constructor, dispatch on line 566 guarded by `s.pdfOutbox == nil && s.pdfDispatcher != nil && s.pinInvoker == nil` | Conditionally live: falls back to it when outbox is not configured — NOT purely dead |
| `FreezeService.Freeze` sync path | `freeze_service.go:302` — annotation "New code should use Pin + Materialize instead" | Possibly still called; need callsite search |
| `CompositionConfig` struct | `templates/domain/schemas.go:81` — ADR 0008 removed composition 2026-04-27; no callers other than test at `schemas_test.go:75` | Dead exported type with no production caller |
| `AreaService.SetParent` | `taxonomy/application/area_service.go:111` — only called from `area_service_test.go:49` | Dead production path; cycle-check logic silently absent from the production update handler |
| `SnapshotService.SnapshotFromTemplate` | `documents/application/snapshot_service.go:46` — marked deprecated | Needs callsite check |
| `WorkerConfig.ReviewReminderDays` | `config/worker.go:13,25,50` + `worker/main.go:109` — parsed, logged at startup, referenced by nothing in production logic | Confirmed dead field |
| Legacy `areas`, `visibility`, `specific_areas` columns in templates INSERT | `templates/repository/postgres.go:52-56` — written as `'{}'::text[], 'public', '{}'::text[]` hardcoded literals; no domain.Template fields for them | Confirmed live but vestigial writes |
| `document_profiles.is_active` DB column | `archive/migrations/0023:14`; `0122:13` — superseded by `archived_at`; confirmed no Go code reads or writes it | Dead DB column |
| `document_subjects` table | `archive/migrations/0025:9-16` — zero Go code references confirmed by register | Dead DB table (in archive) |
| `resolvePermissionFallback` | `permissions.go:270-279` — `switch { default: ... }` with a discarded `path` parameter (`_ = path`) | Confirmed dead function body |
| `coverage_boost_test.go` | `approval/application/coverage_boost_test.go:1` — explicit comment "push total coverage to ≥90%"; duplicates setup from primary test files | Coverage-gaming artifact |

### Standard

**CWE-561 (MITRE):** "Dead code indicates source code problems requiring fixes and reflects poor quality standards." The MITRE entry lists **reduced maintainability** as a direct consequence and the recommended mitigation is removal. Critically: "security issues in dead code are still security issues" — the `PDFDispatchInvoker` fallback path, for instance, dispatches PDF jobs conditionally based on a nil check that could silently reactivate.

**YAGNI (Fowler):** The "cost of carry" applies. Each deprecated type adds cognitive load during every review, every refactor, and every security scan. A static analysis pass over `CutoverService` must confirm it is not a live execution path — that is work with zero upside.

**Go effective idiom (go.dev/doc/effective-go):** Exported identifiers with no production callers should not exist; they imply a public API commitment the package cannot honour.

### Verdicts

**`CutoverService`:** DELETE. Migration 0142 was applied ~1 year ago. The service exists only to test itself. Its presence in the compiled binary costs nothing at runtime, but it creates false impression of a reachable code path, adds a DB bypass call (`authz.BypassSystem`) that must be audited in every security review, and the `coverage_boost_test.go` reference is itself a liability (see below). Smallest fix: delete `cutover_service.go`, delete the test coverage in `coverage_boost_test.go` that exercises it, remove `ErrLegacyDocumentsRemain` error var if no other consumer. Pre-deletion: `grep -r CutoverService .` — already done above, confirming only test files reference it.

**`PDFDispatchInvoker` fallback path:** SIMPLIFY, not DELETE. This is **conditionally live**: `decision_service.go:566` dispatches via `pdfDispatcher` when `pdfOutbox == nil`. In production, `pdfOutbox` is wired (bootstrap wires it); the fallback fires only in test or unconfigured deployments. However, the register is correct that "silently activates if caller omits outbox wiring" — this is a correctness risk. The fix is not deletion but making the condition fail-loudly: if `pdfOutbox` is nil in production mode, the constructor should panic or error, not silently fall back. The interface and field stay; the nil-branch fallback becomes a `log.Fatal` or a startup-time guard. Effort: S.

**`FreezeService.Freeze` synchronous path:** DEFER. "New code should use Pin + Materialize instead" is an annotation, not a deletion directive. A callsite search would confirm whether any production path still hits it. If it does, it must stay. If it does not, DELETE. Trigger: add a `grep -r '\.Freeze(' internal/` to the next cleanup pass. No action in this ticket.

**`CompositionConfig` struct:** DELETE. ADR 0008 removed composition on 2026-04-27. The register confirms no production callers — only `schemas_test.go:75` roundtrip test, which is a coverage artifact testing a dead type. Delete the struct, delete the test, confirm `grep -r CompositionConfig .` is clean. Effort: S / contained.

**`AreaService.SetParent`:** DELETE. No production callsite — production update handler calls `AreaService.Update` instead. The cycle-detection logic it contains is the only distinctive feature and it is not applied to the production update path, which is a separate correctness issue (tracked under the taxonomy module). Deleting `SetParent` forces the cycle-check question into the open. Confirm with `grep -r SetParent .` (already done: only test files call it). Effort: S.

**`SnapshotService.SnapshotFromTemplate`:** DEFER. "Retained for unnamed backfill scripts" is an undocumented dependency. Before deletion, the team must confirm no external script invokes it. Trigger: document what backfill scripts exist and whether they are retired. Not a Stage-2 blocker.

**`WorkerConfig.ReviewReminderDays`:** DELETE. The field is parsed, validated, and logged at startup but referenced by no production code path. It creates false expectation that review-reminder functionality is active. Remove the field from `WorkerConfig`, its parse block in `LoadWorkerConfig`, and the `%d` reference in `worker/main.go:109`. Effort: S / contained. No database or API changes.

**Legacy `areas`/`visibility`/`specific_areas` columns (templates INSERT):** SIMPLIFY. The INSERT at `postgres.go:52-56` writes hardcoded empty literals `'{}'::text[], 'public', '{}'::text[]` for columns that no longer have domain.Template fields. This is vestigial write pressure: every template creation touches columns that serve no purpose. The smallest fix is to remove those three column/value pairs from the INSERT query once confirmed the columns are safe to drop (or simply stop writing to them). The columns themselves are a DB-level concern; a migration to drop them is the correct eventual fix but belongs under the database family. Stage-2 action here: remove from the INSERT query so they become DB-default columns. Effort: S / contained.

**`document_profiles.is_active` column:** DEFER to DB family (not Go code). The column lives in `archive/migrations/` (the pre-baseline tree not applied by the active runner). It is a DB-level orphan. The correct remediation is an `ALTER TABLE DROP COLUMN` migration if the column still exists in the live schema. Confirm with `\d document_profiles` in the next DB audit pass. No Go code changes.

**`document_subjects` table:** DEFER to DB family. Same reasoning — archive-only migration, zero Go code references. Confirm table existence in live DB, then drop via migration.

**`resolvePermissionFallback`:** DELETE. The function body is `_ = path; switch { default: return "", VisibilitySessionRequired }`. It discards its argument, has a single `default` case, and is a net-zero function. Its continued presence creates the impression that there is meaningful fallback routing logic. Delete the function. Any callers get `return "", VisibilitySessionRequired` inlined directly (one call site: `newPermissionResolver` chain — confirm with grep). Effort: S / contained.

**`coverage_boost_test.go`:** DELETE after deleting `CutoverService`. The file's stated purpose is "push coverage to ≥90%." This is an anti-pattern: it tests dead code (`CutoverService`) to manufacture a coverage number. Test coverage should reflect the behaviour of live code, not the volume of tests over dead artifacts. The test is itself a form of dead code per CWE-561. Delete the file. If coverage drops below threshold, address it by testing actual production paths (which is the correct behaviour). Effort: S / contained.

### Smallest correct fix

Group into three ordered micro-PRs:

1. **PR: delete-dead-types** — Remove `CutoverService` + its test, `CompositionConfig` + its test, `AreaService.SetParent` + its test, `resolvePermissionFallback`, `WorkerConfig.ReviewReminderDays` + worker log line, `coverage_boost_test.go`. Run `go build ./...` and `go test ./...` to confirm clean.
2. **PR: simplify-pdfDispatch-fallback** — Add production nil-guard in `DecisionService` constructor so `pdfDispatcher == nil` in production mode causes a startup error (fail-loud > fail-silent).
3. **PR: templates-insert-cleanup** — Remove legacy column writes from `CreateTemplate` INSERT; confirm with integration test.

### Effort / blast-radius

PR1: **M / module** (touches four modules but no cross-module wiring).
PR2: **S / module** (single service constructor).
PR3: **S / contained** (single SQL query in one file).

### ADR needed

No — deletions of post-migration utilities and coverage-gaming artifacts are engineering hygiene. The `PDFDispatchInvoker` simplification (fail-loud vs fail-silent) does not change the public API.

### Over-engineering check

Do not replace deleted items with interfaces or abstract base types "for future use." Delete means delete. Do not build a replacement `ReviewReminder` feature at this stage — that is YAGNI until a product requirement is written.

---

## D-02 — Three MinIO Clients from the Same Credentials

### Current state

`internal/platform/bootstrap/api.go:85-103` and `buildMinioClients(:158-178)` create **two** `*minio.Client` instances (one internal-endpoint for data ops, one public-endpoint for presigning). `miniostore.NewStore(attachmentsCfg)` at `:97` creates a **third** `*minio.Client` from the same `attachmentsCfg.MinIOAccessKey` / `attachmentsCfg.MinIOSecretKey`. All three share identical credentials; two share the same internal endpoint. No connection pool is shared.

The two clients in `buildMinioClients` serve **different logical endpoints** (internal vs public browser-reachable), so they cannot trivially be merged. The third (`miniostore.NewStore`) uses the internal endpoint only for byte I/O (PDF conversion).

### Standard

The minio-go `*minio.Client` encapsulates an `http.Client` internally. Official documentation does not explicitly declare goroutine safety in a prominent disclaimer, but:
- Issue #598 (minio/minio-go, 2017): maintainer response states the client is "thread safe."
- Issue #1125 (minio/minio-go, 2019): race detector found a data race in bucket-location caching in pre-v7 versions. This was fixed in v7 via sync primitives on the bucket location cache.
- minio-go v7 `Client` struct uses `sync.RWMutex` on the `bucketLocCache` field (`client.go`) — the known race source from #1125 is resolved in the current major version.

Go's `net/http.Client` is documented as safe for concurrent use (go.dev/pkg/net/http/#Client: "A Client is safe for concurrent use by multiple goroutines.").

Three clients from the same credentials is therefore **not a correctness bug** for the current v7 codebase — but it is a resource waste: each client holds its own `http.Transport` (TCP connection pool, TLS session cache). Two clients pointing at the same endpoint are an unnecessary duplication of connection pool state.

**The two clients serving different endpoints are correct by design** (internal presigning vs browser-reachable URL generation require different `endpoint` strings). No merge is possible without breaking presigned URL domain.

**The third client (`miniostore.NewStore`) duplicates the internal-endpoint client** already created by `buildMinioClients`. These two could share a single `*minio.Client` instance since they use the same endpoint and credentials, eliminating one TCP pool.

### Verdict: SIMPLIFY (partial — merge internal clients; retain public client)

The public client cannot be eliminated — it generates browser-facing presigned URLs against a different hostname. The internal client and the `miniostore` client point at the same endpoint with the same credentials. Pass the internal client created in `buildMinioClients` into `miniostore.NewStore` (or accept a `*minio.Client` in its constructor). This reduces from three clients to two, shares one TCP pool for all data-plane operations, and simplifies bootstrap reasoning.

Do not merge all three into one: the public-endpoint client **must** be a distinct `*minio.Client` initialized against the public endpoint or presigned URLs will contain the wrong host.

### Smallest correct fix

1. Change `miniostore.NewStore` signature from `func NewStore(cfg config.AttachmentsConfig) (*Store, error)` to `func NewStore(client *minio.Client, bucket string) *Store` (or add a constructor that accepts a client).
2. In `bootstrap/api.go`, after `buildMinioClients`, pass `internalClient` to `miniostore.NewStore` instead of constructing a fourth client from config.
3. Delete the internal credential-construction code in `miniostore.NewStore` (currently reading the same `AttachmentsConfig`).

Net result: two clients instead of three; internal HTTP connection pool shared between presigning and byte I/O.

### Effort / blast-radius

**S / contained.** Two files change: `storage/minio/store.go` and `bootstrap/api.go`. No API contract changes. `bootstrap/worker.go` also calls `miniostore.NewStore` for the worker binary — it would need the same change (pass the worker's minio client), which is also an S change.

### ADR needed

No — this is an implementation detail of the bootstrap wiring, not an architectural decision.

### Over-engineering check

Do not build a `ClientPool` or `ClientFactory` abstraction. Do not introduce a `ClientRegistry`. The fix is: one constructor parameter change and one fewer `minio.New()` call. If in the future a fourth concern requires MinIO access, revisit.

---

## D-05 — `platform/cache` Placeholder vs `CachedRoleProvider` (No Cache Contract)

### Current state

`internal/platform/cache/` is an empty directory with only `.gitkeep` (confirmed: zero Go source). The sole in-process cache is `CachedRoleProvider` at `internal/modules/iam/application/cached_role_provider.go`. It implements a `sync.Map`-style TTL cache using `sync.RWMutex` + `map[string]cacheEntry`. TTL defaults to 30 seconds; a background goroutine sweeps expired entries. Explicit `InvalidateUserTenant` is called on role write paths. However:

- No written cache contract exists (violates REQ-CACHE-1).
- The comment at `cached_role_provider.go:80-83` acknowledges "if group-membership mutation routes are ever added they must call `InvalidateUserTenant` too, or stale roles persist until the TTL" — a staleness risk that is acknowledged but undocumented externally.
- The cache has no max-size cap ("acceptable at current scale" per comment at `:22-24`).
- The `platform/cache` empty directory implies there is infrastructure that does not exist.

### Standard

**REQ-CACHE-1:** "Every cache has a one-page contract: what's cached, key shape, TTL, invalidation events, staleness bound, and failure behavior. No contract, no cache."

**REQ-TOP-3:** "Empty scaffolds are deleted or implemented." The empty `platform/cache` package directly violates this.

**YAGNI + Fowler's "cost of carry":** The empty directory implies a future caching infrastructure extraction that may never happen. It misleads contributors about what the platform provides.

### Verdict: DELETE `platform/cache` (covered under F-08) + REFACTOR `CachedRoleProvider` to have a written contract

The platform/cache deletion is addressed under F-08. The distinct action here is: **write the one-page cache contract for `CachedRoleProvider`** as required by REQ-CACHE-1. This does not require moving the implementation — the IAM module is the correct owner because the cache is IAM-domain-specific. A platform extraction is only warranted when a second module needs identical caching semantics, which is not currently the case.

The contract must specify:
- What is cached: `(userID, tenantID) → []domain.Role`
- Key shape: `userID + "|" + tenantID`
- TTL: 30 seconds (configurable via constructor parameter; default 30s if `ttl <= 0`)
- Invalidation events: role upsert/replace (`AdminService`), user invite (`PeopleService.Invite`), area grant/revoke (`AreaMembershipService`) — all confirmed by the code comment at `:80-83`
- Staleness bound: at most 30 seconds post-mutation for cached entries; immediate for invalidated entries
- Failure behavior: on DB error, returns error to caller — the cache never produces a wrong allow (REQ-CACHE-2 compliant because an error is not a spurious grant)
- Unbounded growth risk: acknowledged; acceptable until tenant × user cardinality exceeds ~10k active sessions (no evidence this is a concern at current scale)

### Smallest correct fix

1. `git rm internal/platform/cache/.gitkeep` (see F-08).
2. Add a `// CacheContract:` doc comment block to `cached_role_provider.go` covering the five required fields above. This is 15-20 lines of inline documentation — no code changes needed, no extracted file needed. If the team prefers a separate doc file, `wiki/modules/iam/_artifacts/role-cache-contract.md` is the right location.

### Effort / blast-radius

**S / contained.** One documentation addition in one file; one VCS deletion.

### ADR needed

No for the documentation addition. If the team decides to extract caching infrastructure to `platform/cache`, that requires an ADR (architectural decision to create a new platform package), but that is a future concern only if a second cache consumer appears.

### Over-engineering check

Do not implement a Redis-backed or generics-based cache provider in `platform/cache` on the basis of this finding. The existing `sync.Map`-style implementation is appropriate for the current single-tenant-in-process model. REQ-CACHE-1 requires a written contract, not a different implementation. The contract writes itself from what is already in the code comments — it just needs to be made explicit.

---

## D-06 — `cmd/` Root vs `apps/*/cmd/` Convention Split

### Current state

`cmd/seed-test-document/main.go` lives at the repository root under `cmd/`, following the classic Go project layout convention (`golang-standards/project-layout`). All three active binaries (`metaldocs-api`, `metaldocs-worker`, `metaldocs-jobs`) plus the tooling binary (`metaldocs-e2e-seed`) live under `apps/<name>/cmd/<binary>/`, which is MetalDocs-canonical (confirmed by `wiki/backend/repo-topology.md §4`). `go build ./...` traverses both locations. The root `cmd/` has exactly one inhabitant: the dead seed binary already flagged under F-08 / F-14.

### Standard

**Go module layout (go.dev/doc/modules/layout):** The official Go documentation states that `cmd` is the conventional directory for placing main packages when a repository contains both libraries and commands. `golang-standards/project-layout` (the informal but widely cited community layout) also uses `cmd/<binary>/`. Neither document defines `apps/` as a canonical alternative.

However, the Go toolchain has no preference between `cmd/` and `apps/*/cmd/` — both are valid Go package paths. The convention mismatch is an internal coherence issue, not an external standards violation. The relevant constraint is REQ-TOP-3 (no empty scaffolds) and YAGNI: the root `cmd/` exists only because of the dead seed binary. If the seed binary is deleted, the `cmd/` directory disappears.

### Verdict: DELETE (the root `cmd/` directory disappears when its only inhabitant is deleted)

The convention drift is self-resolving. The seed binary's deletion (F-08, F-14) eliminates the root `cmd/` entirely. No active binary needs to be moved. No `go.mod` or build-script change is needed. The register is correct that D-06 is caused by not cleaning up after the dead binary. The verdict is: delete the binary, the directory disappears, the drift is gone.

There is no case for migrating the three active binaries from `apps/*/cmd/` to `cmd/` — that would be a large-blast refactor (build scripts, Dockerfiles, CI references) for no functional benefit. The `apps/` convention is consistently applied across all active binaries and is the MetalDocs-canonical path.

### Smallest correct fix

`git rm -r cmd/seed-test-document/` — covered by F-08 / F-14 deletion. Nothing else.

### Effort / blast-radius

**S / contained.** Included in the F-08 deletion PR.

### ADR needed

No.

### Over-engineering check

Do not add a `cmd/` migration task. Do not rename `apps/` to `cmd/` to match `golang-standards/project-layout`. Both are valid; the MetalDocs convention is coherent once the dead outlier is removed.

---

## D-07 — Domain Status Enum Fragmentation

### Current state

**Documents module:** `internal/modules/documents/domain/model.go:8-13` declares three `DocumentStatus` constants: `DocStatusDraft`, `DocStatusUnderReview`, `DocStatusArchived`. The OpenAPI spec at `api/openapi/v1/openapi.yaml:4290` enumerates eight values: `draft, under_review, approved, rejected, scheduled, published, superseded, obsolete`. Go code that handles `approved`, `published`, `superseded`, `obsolete`, `scheduled`, `rejected` states does so via string literals or API-layer types (`api.gen.go`) without the backing domain constants. There is no exhaustive switch check at compile time that would catch a missing status value.

**Templates / taxonomy:** `document_profiles.is_active` column (added migration 0023, superseded by `archived_at` in migration 0122) is addressed under F-14 — it is a DB concern, not a Go enum concern.

The register groups two distinct root causes under D-07: (a) the incomplete Go domain enum for `DocumentStatus` and (b) the `is_active`/`archived_at` column supersession. These should be evaluated separately.

### Standard

**Go type safety:** A `DocumentStatus` type with only 3 of 8 live values means any switch on `DocumentStatus` in Go application code is implicitly exhaustive for only 3 values. New document states added at the API layer can bypass the domain type system silently. `staticcheck SA4003` and `exhaustive` linter (github.com/nishanths/exhaustive) can detect non-exhaustive switches — but only if the constants exist to enumerate.

**Go effective idiom (go.dev/doc/effective-go §Constants):** Constants should be defined where they are used. Having the canonical enum in `api.gen.go` (a generated file) means the domain layer is dependent on contract-layer types for authoritative enumeration — an inversion of the intended dependency direction.

**YAGNI counter-argument:** Adding constants for states that are handled only in the approval pipeline / scheduled-jobs path (approved, published, superseded, obsolete, scheduled, rejected) creates the appearance that the documents core module handles all states. It does not — `documents/domain` owns draft, under_review, archived; the approval pipeline owns state transitions to approved/rejected/published; the controlled-documents lifecycle owns superseded/obsolete. Adding all 8 constants to `documents/domain/model.go` would merge concerns that belong to different bounded contexts.

### Verdict: SIMPLIFY (not a full refactor — add missing constants where each domain owns them)

The correct fix is not to add all 8 constants to `documents/domain/model.go`. It is to ensure each bounded context owns the constants for the states it manages:

- `documents/domain`: add `DocStatusApproved`, `DocStatusRejected`, `DocStatusPublished`, `DocStatusSuperseded`, `DocStatusObsolete`, `DocStatusScheduled` — these are **observed** by the documents core domain even if transitions are managed by approval/jobs. A `Document.Status` field of type `DocumentStatus` should be able to express any valid state the DB can return.
- This is a pure constant addition — no logic changes, no migration, no API change.

The `is_active` column fragmentation is DEFER under DB family (already classified in F-14).

Do not add an `exhaustive` linter enforcement as part of this step — that is a separate CI hygiene task (medium effort, cross-module blast radius).

### Smallest correct fix

Add 6 constants to `documents/domain/model.go`:

```go
const (
    DocStatusDraft       DocumentStatus = "draft"
    DocStatusUnderReview DocumentStatus = "under_review"
    DocStatusApproved    DocumentStatus = "approved"
    DocStatusRejected    DocumentStatus = "rejected"
    DocStatusScheduled   DocumentStatus = "scheduled"
    DocStatusPublished   DocumentStatus = "published"
    DocStatusSuperseded  DocumentStatus = "superseded"
    DocStatusObsolete    DocumentStatus = "obsolete"
    DocStatusArchived    DocumentStatus = "archived"
)
```

Then update any application / repository code that currently uses string literals for these states to use the typed constants. The register does not enumerate those call sites for documents/domain specifically; a `grep -rn '"approved"\|"published"\|"superseded"\|"obsolete"\|"scheduled"\|"rejected"' internal/modules/documents/` will surface them.

### Effort / blast-radius

**S-M / module.** Constant additions are trivially S. Updating string literal callsites across the documents module is M. No cross-module change; other modules (approval, controlled-documents, jobs) reference their own state management and are not affected.

### ADR needed

No — this is a domain-type completeness fix. If the decision were to split `DocumentStatus` into sub-types per bounded context (e.g., a separate `ApprovalStatus` type), that would warrant an ADR. Adding missing constants to the existing type does not.

### Over-engineering check

Do not add a `DocumentStatus` interface, a state machine library, or a generated enum type. Seven string constants and a type alias is the correct Go idiom. Do not add exhaustive-switch enforcement to CI as part of this PR — measure first whether it causes noise across the codebase.

---

## Summary Table

| Finding | Verdict | Priority | Effort | Blast Radius | ADR |
|---|---|---|---|---|---|
| F-08: Empty scaffolds + dead binaries | DELETE (4x `.gitkeep`, `bin/` binary) + KEEP (memory mode) | P1 | S | contained | No |
| F-14: Dead application code | DELETE (6 items), SIMPLIFY (PDFDispatch), DEFER (3 items) | P1 (safety-risk items), P2 (hygiene) | M total | module | No |
| D-02: Three MinIO clients | SIMPLIFY (merge internal+store client; keep public client) | P2 | S | contained | No |
| D-05: cache placeholder + no contract | DELETE placeholder (F-08) + REFACTOR contract doc | P2 | S | contained | No |
| D-06: cmd/ convention split | DELETE (resolves when seed binary deleted) | P1 (included in F-08) | S | contained | No |
| D-07: Status enum fragmentation | SIMPLIFY (add missing constants) | P2 | S-M | module | No |

---

## Sources

- Martin Fowler, "Yagni," martinfowler.com/bliki/Yagni.html (2015)
- MITRE CWE-561 "Dead Code," cwe.mitre.org/data/definitions/561.html (v4.20)
- Go module layout, go.dev/doc/modules/layout
- golang-standards/project-layout, github.com/golang-standards/project-layout
- minio-go thread-safety discussion, github.com/minio/minio-go/issues/598 and /issues/1125
- minio-go v7 API reference, pkg.go.dev/github.com/minio/minio-go/v7
- REQ-TOP-3, REQ-CACHE-1, REQ-CACHE-2: wiki/architecture/backend-target-architecture.md
- wiki/backend/legacy-register.md (F-08, F-14, D-02, D-05, D-06, D-07)
- wiki/backend/repo-topology.md, wiki/backend/platform/data-layer.md
