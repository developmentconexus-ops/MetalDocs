# PASS 9 — Platform / Composition Boundary Audit (REQ-TOP-2)

**Date:** 2026-08-09
**Baseline:** `main@418070bf38a9f358f9131bcc36b7a6bcbc069273`
**Status:** reproduced-current (fresh local measurement in an isolated worktree)

Evidence-class labels: `reproduced-current` (measured at this baseline),
`historical` (true at filing time, superseded), `stale` (contradicted by
current runtime truth).

## 0. Scope and method

Rule under audit — **REQ-TOP-2**: `internal/platform/*` packages must never
import `internal/modules/*`; `internal/composition/*` packages may import
both. Scope: the 37 top-level directories under `internal/platform/` (per
PASS 1 §3) plus the one root under `internal/composition/` —
`internal/composition/tenantdata/registry`. 38 rows total.

Method: `go list -json` module-graph edges from PASS 2
(`module-edge-evidence.txt`), cross-checked with a fresh
`grep -rl "internal/modules/" internal/platform` for existence, then each
hit's importing file read in full for semantic classification (what the
package DOES with the module import, not just that the edge exists).

## 1. Headline finding

Exactly **4 of 37** platform packages import a module package, confirming
PASS 2's count with no drift:

| Platform package | Imports | Files |
|---|---|---|
| `authn` | `auth/application`, `iam/domain` | `config.go`, `context.go` |
| `bootstrap` | `audit/{domain,infrastructure/postgres}`, `auth/{domain,infrastructure/postgres}`, `iam/{domain,infrastructure/postgres}` | `api.go`, `worker.go`, `jobs.go` |
| `docgenv2` | `documents/application`, `documents/domain` | `templates_reader.go`, `templates_snapshot_reader.go` |
| `tripwire` | `iam/domain` | `arms.go` |

The other 33 platform packages are confirmed domain-free by grep (zero
`internal/modules/` imports across the whole subtree).

## 2. The 4 offending packages, in detail

### 2.1 `internal/platform/bootstrap` — composition root mis-filed under platform

**What it does:** `BuildAPIDependencies()` (`api.go`), `BuildWorkerDependencies()`
(`worker.go`), `BuildJobsDependencies()` (`jobs.go`) each construct the full
dependency graph for one binary — Postgres repos for `audit`, `auth`, `iam`
(domain + infrastructure/postgres), MinIO clients, Gotenberg health checks,
outbox publisher — and return a single `*Dependencies` struct consumed by
`apps/{api,worker,jobs}/cmd/.../main.go`.

**Why it imports modules:** because it *is* a composition root. Every line
of `api.go`/`worker.go`/`jobs.go` is wiring: `New...Repository(db)` calls
into 3 modules' infrastructure packages, assembled into one struct, handed
to `main.go`. There is no platform behavior here — no shared primitive any
module depends on downward. It is pure "many modules in, one dependency
bag out," which is the textbook composition-root shape.

**Verdict: this is a composition root, mis-filed under `internal/platform`
by name only.** It should physically be `internal/composition/bootstrap`
(or one composition root per binary — `internal/composition/api`,
`internal/composition/worker`, `internal/composition/jobs` — mirroring how
`tenantdata/registry` is scoped to one concern rather than one giant
grab-bag). The current name (`platform/bootstrap`) reads as "shared
low-level infra" but the directory holds none — it holds only wiring code
for 3 (not all 15) modules, an accident of which modules happened to need
constructing from `main.go` at each binary's entry point rather than a
deliberate platform primitive.

**Disposition: move-to-composition.**

### 2.2 `internal/platform/docgenv2` — cross-module business capability, not platform

**What it does:** `TemplatesTemplateReader` and `TemplatesSnapshotReader`
implement `documents/application.TemplateReader` /
`.SnapshotTemplateReader` (interfaces *owned by the `documents` module`)
by running raw SQL directly against `templates_template_version` and
`templates_template` — tables owned and normally only touched by the
`templates` module's own infrastructure package. Confirmed via grep: these
two types are constructed in exactly one place,
`apps/api/cmd/metaldocs-api/main.go:411,419`.

**Why it imports modules:** it is the adapter that lets the `documents`
module's docgen pipeline read `templates` module data, satisfying an
interface `documents/application` defines. That is a legitimate need — but
implementing it via **raw SQL against another module's tables** is exactly
the invariant-6 violation ("cross-module access goes through a module's
application service or published Go interface — never another module's
repository, SQL, or domain internals," CLAUDE.md) that `tenantdata.Port`
and `tripwire` were explicitly designed to avoid (see their own doc
comments, §3 and §4.1 below). This one adapter reaches straight into
`templates`'s schema instead of going through a `templates`-published port.

**Verdict: this is a misplaced business capability, and the mechanism it
uses (raw cross-module SQL) is itself an invariant-6 violation, not just a
directory-naming issue.** The clean shape: `templates` module publishes a
`TemplateReader`/`SnapshotTemplateReader`-shaped port (mirroring how every
other module publishes a `tenantdata.Port`), `templates/infrastructure`
implements it against its own tables, and `documents/application`'s
adapter simply calls that port instead of hand-rolling SQL. That adapter
call site can then live in composition (wiring two published ports
together) with zero raw SQL crossing a module boundary.

**Disposition: move-to-module** (the reader implementations belong inside
`templates/infrastructure`, exposed as a `templates`-published port;
`documents/application` keeps only the interface it already owns). This is
a two-module change (new port in `templates`, adapter update in
`documents`), out of scope for this read-only audit pass — flagged as the
A4 finding, not executed here.

### 2.3 `internal/platform/authn` — split identity: real platform half + composition half

**What it does — two distinct things in one package:**
- `context.go`: defines the fail-closed `UserIDFromContext(ctx) (string, bool)`
  helper — a genuine platform primitive (pure `context.Context` read, no
  module dependency at all in this file).
- `config.go`: `LoadRuntimeConfig()` builds an `authapp.Config` (a struct
  *owned by* `internal/modules/auth/application`) from environment
  variables, and `DevRoleMap()` parses dev-mode role strings into
  `iamdomain.Role`. Both of these exist only to hand a fully-built
  module-shaped config value to `main.go` at startup.

**Why it imports modules:** `config.go` is env-var-to-module-config
translation — composition-root wiring, same genre as `bootstrap`, just
scoped to one concern (auth runtime config) instead of the whole
dependency graph. `context.go`'s fail-closed helper needs no module import
at all — it is the one part of this package that is legitimately platform.

**Verdict: this package is two packages wearing one name.** `context.go`
is truly-domain-free-platform and should stay. `config.go` is
composition-root wiring mis-filed as platform, structurally identical to
`bootstrap`'s problem (same "build a module Config from env" shape) — it
should move next to (or into) `bootstrap`'s composition-root home.

**Disposition: split** — `authn/context.go` stays in platform (rename
package if desired, e.g. `platform/actorctx`, to stop the file-level split
being invisible from the import path); `authn/config.go` moves-to-composition
alongside `bootstrap`.

### 2.4 `internal/platform/tripwire` — deliberate, self-documented, narrow one-way exception

**What it does:** `arms.go` is the Go-side source of truth for capability
literals that must match the `enforce_capability_asserted()` Postgres
trigger (the DB tripwire, tier-3 of the two-tier-PDP-plus-DB-backstop
model in ADR 0022). It imports `iamdomain` for exactly one thing: the
`Capability` string-const type, so the Go arm list and the DB trigger's
literal list can be generated/diffed from one enum rather than two
hand-synced copies.

**Why it imports modules:** to avoid a worse alternative — duplicating
`iamdomain.Capability`'s literal values in platform, which is exactly the
"hand-synced enumeration" meta-defect flagged in the 2026-07-03 final
architecture review (`docs/superpowers/analysis/...` per user memory).
The package's own header comment documents this as a considered "one-way
import" exception, not an oversight.

**Verdict: this is the one case of the 4 that is arguably fine as
platform**, not a composition root or a misplaced capability — it imports
one module's *type*, never a service or repository, purely to keep one
enum canonical. It is structurally identical in shape to `iamtypes` (§3)
except that `iamtypes` avoided the module import entirely by extracting a
copy of the vocabulary into platform, while `tripwire` chose to import the
type directly instead of extracting. Both are legitimate resolutions of
the same "who owns the shared vocabulary" question; `tripwire`'s choice is
the one that technically breaks REQ-TOP-2's letter while satisfying its
spirit (still zero risk of an import cycle, since `iam` never imports
platform's `tripwire`).

**Disposition: stay, with a documented REQ-TOP-2 exception** (the package
already carries the justification in its header; recommend cross-linking
that comment to REQ-TOP-2 explicitly in `wiki/architecture/backend-target-architecture.md`
so future readers don't have to rediscover the reasoning cold), OR,
if REQ-TOP-2 is to be enforced with zero exceptions, extract a
`platform/iamtypes`-style `Capability` const list the same way `Role` was
extracted — trading one exception for one more shared-vocabulary package.
Both are legitimate; this audit does not adjudicate between them (product
decision, not an architecture defect).

## 3. `platform/iamtypes` — healthy shared vocabulary, not a smell

`iamtypes.Role` (`internal/platform/iamtypes/role.go`) is a types-only
package: one string-const `Role` type plus its literal values, no logic,
no imports beyond the stdlib. Consumed by exactly 3 modules — confirmed by
grep: `auth`, `iam`, `taxonomy`.

Its own doc comment explains the origin precisely: `Role` used to live in
`iam/domain`, but `auth` needed it too, and `auth`↔`iam` already had a
dependency in the other direction for unrelated reasons — leaving `Role`
in `iam/domain` would have created a bidirectional module edge (forbidden;
modules only reach each other through published application-service
interfaces, never both ways on the same concept). Extracting the pure
vocabulary to platform breaks the cycle at zero cost, because a
const-only type has no business logic to misplace.

**This is the target pattern**, not an exception to it — it is the same
shape `tenantdata.Port` and `tripwire`'s `iamdomain.Capability` question
are both reaching for: when N modules need to agree on one small piece of
vocabulary and putting it in any one module creates a cycle, a
dependency-free platform package is the correct home, provided (as here)
it stays pure types/consts with no behavior. The risk pattern to watch for
is scope creep — if `iamtypes` ever grows a function that encodes a
business rule (e.g., "which roles can approve"), that function belongs in
a module, not here.

**Disposition: stay.**

## 4. `platform/tenantdata` vs `composition/tenantdata/registry` — the split explained

Two packages, deliberately separated, and the separation is load-bearing
(not incidental):

- **`internal/platform/tenantdata`** (`port.go` + `export_helper.go`):
  defines the `Port` interface (`Module()`, `Tables()`,
  `ExportTenantData()`, `EraseTenantData()`) and generic SQL helper
  functions (`ExportTable`, `EraseTable`, ...). Zero module imports. 12
  modules (`approval, audit, auth, controlleddocuments, documents, iam,
  jobs, notifications, render, taxonomy, templates, tokens` — confirmed by
  grep) import *this* package downward, each to declare
  `var _ tenantdata.Port = (*XxxTenantDataPort)(nil)` against their own
  infrastructure implementation.

- **`internal/composition/tenantdata/registry`** (`registry.go`): imports
  `platform/tenantdata` plus all 12 modules' infrastructure packages by
  name, and exposes one function, `AllTenantDataPorts(db) []tenantdata.Port`,
  that constructs and returns the full slice. Consumed by both
  `apps/api` (export/erase HTTP handlers) and `apps/jobs` (the fan-out
  worker) so there is exactly one wiring list, not two.

**Why the split is necessary, not stylistic** (this is spelled out
verbatim in `registry.go`'s own header comment, confirmed by reading it):
if the registry function lived in the same package as the `Port`
interface, then every implementing module (`render/fanout`,
`documents/infrastructure`, etc.) would import that package to satisfy the
interface — and the registry in that same package would import those same
modules back, a straight import cycle. Splitting "the contract" (platform,
zero deps) from "the wiring that knows every implementation" (composition,
depends on everything) is the only way to let 12 modules depend downward
on one shared interface while one composition root depends upward on all
12 — precisely REQ-TOP-2's own boundary rule, applied correctly.

**Verdict: this is the right shape, and it is the model the other 3
offending platform packages (§2.1–2.3) should be refactored to match.**
`bootstrap` and `authn/config.go` are doing exactly the "wiring that knows
every implementation" job registry.go does — they just weren't given a
composition-root home when they were written. `docgenv2` is different in
kind (it's an adapter reaching into raw SQL, not a constructor list) and
needs the port-extraction fix in §2.2, not a directory move alone.

**Disposition:** `platform/tenantdata` stays; `composition/tenantdata/registry`
stays; no change recommended to this pair — it is cited as the reference
pattern for §2.1/§2.3's move-to-composition recommendation.

## 5. Full 38-row classification table

Categories: **domain-free-platform** (truly shared, zero module knowledge)
· **module-specific-misplaced** (belongs inside one module) ·
**composition-root** (wiring many modules together for one binary/concern)
· **adapter** (implements one module's interface against another's data) ·
**utility** (generic, stateless helper, domain-free) · **misplaced-business-capability**
(encodes a business rule that belongs in a module).

| # | Package | Category | Disposition | Note |
|---|---|---|---|---|
| 1 | `apibase` | utility | stay | single `BaseURL` const, guarded by test; zero deps |
| 2 | `authn` | **split: domain-free-platform (context.go) + composition-root (config.go)** | split | see §2.3 |
| 3 | `bootstrap` | composition-root | move-to-composition | see §2.1 |
| 4 | `config` | utility | stay | attachment/env config primitives, no module import found |
| 5 | `crypto` | domain-free-platform | stay | self-documented "pure, DB-free... no knowledge of tenants" |
| 6 | `db` | domain-free-platform | stay | tx runner / connection primitives |
| 7 | `docgenv2` | adapter (raw-SQL, invariant-6 violation) | move-to-module | see §2.2 |
| 8 | `featureflags` | utility | stay | serves one generated surface tag; no module import |
| 9 | `formval` | utility | stay | JSON-schema validation wrapper |
| 10 | `httpclient` | utility | stay | generic HTTP client |
| 11 | `httpresponse` | utility | stay | response-writing helpers |
| 12 | `httprouter` | domain-free-platform | stay | minimal `ServeMux` surface + publisher contract |
| 13 | `iamtypes` | domain-free-platform (shared vocabulary) | stay | see §3 |
| 14 | `idempotency` | domain-free-platform | stay | canonical `Require()` middleware + `Store` (PASS 10 §7) |
| 15 | `jobs` (+`jobs/river`) | utility/adapter | stay | River queue adapter primitives, no module import |
| 16 | `legacystatus` | utility | stay | one retired-literal const, self-documented cilint-guard escape hatch |
| 17 | `messaging` | domain-free-platform | stay | outbox consumer/publish primitives |
| 18 | `middleware` | domain-free-platform | stay | chain-link primitives (`method_not_allowed`, etc.) |
| 19 | `migrate` | domain-free-platform | stay | SQL-file migration runner |
| 20 | `objectstore` | domain-free-platform | stay | blob-store error/interface primitives |
| 21 | `observability` | domain-free-platform | stay | health/metrics publishers, generated-surface-scoped |
| 22 | `pagination` | domain-free-platform | stay | canonical keyset-cursor primitive (PASS 10 §6) |
| 23 | `passwordhash` | domain-free-platform | stay | self-documented "pure, DB-free" Argon2id primitive |
| 24 | `problem` | domain-free-platform | stay | RFC 9457 problem+json writer/registry |
| 25 | `ratelimit` | domain-free-platform | stay | rate-limit config/algorithm |
| 26 | `render` (+`render/gotenberg`) | domain-free-platform | stay | generic PDF-render client |
| 27 | `requesttrace` | domain-free-platform | stay | trace-id context helpers |
| 28 | `security` | domain-free-platform | stay | signed URLs, CORS, origin protection, trusted-proxy CIDR |
| 29 | `servicebus` | domain-free-platform | stay | generic Gotenberg client types |
| 30 | `sqlescape` | utility | stay | Postgres LIKE-pattern escaping |
| 31 | `storage` (+`storage/minio`) | adapter | stay | MinIO object-store adapter, generic |
| 32 | `strictjson` | utility | stay | strict-decode helper, explicitly promoted out of a module for reuse |
| 33 | `tenant` | domain-free-platform | stay | tenant-ID context helpers (`ActorFromContext` — PASS 10 §4) |
| 34 | `tenantdata` | domain-free-platform | stay | see §4 |
| 35 | `tripwire` | domain-free-platform (documented exception) | stay-with-exception | see §2.4 |
| 36 | `useragent` | utility | stay | UA-string device-label heuristics |
| 37 | `worker` | utility | stay | tenant-scoped object-key builder |
| 38 | `composition/tenantdata/registry` | composition-root | stay (reference pattern) | see §4 |

## 6. Summary disposition counts

| Disposition | Count | Packages |
|---|---|---|
| stay | 32 | all domain-free-platform + utility + adapter rows above |
| stay-with-exception | 1 | `tripwire` |
| split | 1 | `authn` (context.go stays, config.go moves) |
| move-to-composition | 2 | `bootstrap`, `authn/config.go` |
| move-to-module | 1 | `docgenv2` |
| stay (reference pattern) | 1 | `composition/tenantdata/registry` |

**Net REQ-TOP-2 violation count after the recommended moves: 1**
(`tripwire`, and only if the product decision in §2.4 is "keep the
exception" rather than "extract a `Capability` vocabulary package").

## 7. Root cause

All 3 packages recommended for move (`bootstrap`, `authn/config.go`,
`docgenv2`) share one root cause: **composition-root code and adapter
code were written inline at the point a binary's `main.go` needed them,
before `internal/composition/` existed as a recognized second-class
citizen next to `internal/platform/`.** `composition/tenantdata/registry`
(added later, for M7 F7.3) is the only place this was done correctly —
new module was born already knowing the platform/composition split
mattered. This is the #93/A4 root cause already named in PASS 2: platform
domain-free-ness erodes by accretion at binary-wiring time, not by any
single bad decision — each of the 3 offending packages was a reasonable
local choice ("just wire this module here, in this file") that
collectively broke the global rule. The fix is structural (give every
binary's composition root a home under `internal/composition/`, matching
`tenantdata/registry`'s shape) rather than a one-off patch to any single
package.
