# HTTP Surface Protocol — design

- **Date:** 2026-08-05
- **Status:** draft (in adversarial review)
- **System-impact gate:** `docs/superpowers/analysis/2026-08-05-http-surface-protocol-system-impact.md` (Yellow; AS-2 cleared by operator ruling A)
- **Operator rulings carried in:**
  - **A** (scope) — one program over the whole surface. No two-regime state ships. CLAUDE.md §2 outcome (a) *restructure now*; outcome (b) *transitional label* is not available.
  - **A** (declaration site) — capability reaches tier-1 via an OpenAPI vendor extension plus a generator emitting a Go descriptor.
  - **C** (descriptor location) — the module publishes `Mount(Muxer)`; the capability table is central.
  - **C** (`/healthz`) — deleted, not excepted. The protocol has no exception mechanism.
  - **1** (approach) — pattern-keyed PDP generated from the spec.

---

## 1. Problem

Five unsynchronized enumerations state the same fact — *which route exists, and what may reach it*.

| # | Enumeration | Size | Location |
|---|---|---|---|
| 1 | OpenAPI operations | 147 operationIds / 125 paths | `api/openapi/v1/openapi.yaml` |
| 2 | Mount operations | 17, across 5 distinct Go call shapes | `apps/api/cmd/metaldocs-api/router.go:95-127` |
| 3 | Tier-1 rule table | 120 hand-typed rows | `apps/api/cmd/metaldocs-api/permissions.go:82-330` |
| 4 | Public-path fallback | 4 cases | `internal/modules/auth/delivery/http/middleware.go:118-129` |
| 5 | Password-change allowlist | 3 paths | `internal/modules/auth/delivery/http/middleware.go:131-135` |

None derives from another. #4 carries its synchronization instruction in prose —
`// Keep this in sync with the composition root's authoritative list` — and has already
drifted (`wiki/backend/http-kernel.md:279`).

Three defects are consequences, not coincidences:

- **HEAD bypass.** `routeRule.matches` (`permissions.go:35`) demands exact method equality and
  no row carries `HEAD`. Go's mux routes `HEAD /x` to the `GET /x` pattern, so a
  capability-guarded GET is reachable by HEAD with only a session. Status codes and headers
  leak.
- **`conditionalRouteFamilies` fails open** (`router.go:85`) — a hand-typed exemption set that
  can silently suppress the omitted-family defect the guard exists to catch.
- **`main.go`'s keyed struct literal** (`main.go:817-837`) is invisible to that guard: a field
  left unset is a nil handler, not a compile error.

Every previous fix built a better *detector*. This design removes the duplication instead.

**Goal:** one declaration per operation, in the contract. Both the mux registration and the
tier-1 PDP derive from it. A second statement of route truth becomes unrepresentable.

---

## 2. The single declaration

The spec already carries `x-authz-area` on 16 operations (`openapi.yaml:1108` and 15 others) and
`x-authz-area-none` with a prose reason on one (`:1705`), documenting where tier-2 reads its area
from. Tier-1 policy joins that established `<fact>` / `<fact>-none` convention rather than
introducing a nested object or a parallel enum.

```yaml
  /metrics:
    get:
      operationId: getMetrics
      tags: [observability]
      x-authz-capability: metrics.view
```

**Visibility is derived, never restated.** The spec already expresses two of the three tiers
natively: a root default `security: - sessionCookie: []` (`openapi.yaml:12-13`) and a per-operation
`security: []` override on the 4 genuinely public operations. Exactly one datum is missing — the
capability. So the three tiers come from three mutually exclusive markers, and no marker restates
another:

| Marker on the operation | Tier |
|---|---|
| `security: []` | `VisibilityPublic` |
| `x-authz-capability: <registry string>` | `VisibilityPermissionGuarded` |
| `x-authz-capability-none: '<reason>'` | `VisibilitySessionRequired` |

An `x-authz-visibility` field was in an earlier draft and is **rejected**: with the table above it
carries no information the other three markers do not, and a redundant field needs validation
rules to prove the two statements agree — which is the defect class this design exists to delete,
recreated at the declaration site. The analysis left this open as "derive or state, pick one
deliberately" (`…-system-impact.md:130`); this is the deliberate pick.

`x-authz-capability-none` is not a decoration. It is what stops "session-required" from being a
silent default that an author reaches by forgetting a field — the same reason `x-authz-area-none`
carries a reason string today.

**Its cost is three operations.** `routeRules` has exactly three `VisibilitySessionRequired` rows
(`permissions.go:90-92`: `GET /auth/me`, `POST /auth/change-password`, `POST /auth/logout`), four
public rows, and 113 capability rows; nothing falls through. So the tier that has to be stated
explicitly is the smallest of the three, and "a mandatory reason string will decay into
boilerplate" does not apply at n=3. The distribution is also the argument for stating rather than
deriving: with 140-odd operations guarded, a *missing* capability is overwhelmingly likelier to be
an oversight than an intent.

**Other fields**

| Field | Type | Required |
|---|---|---|
| `x-authz-password-change-allowed` | bool, default false | no |

**Validation — every failure is a generation failure, never a runtime default.**

1. Operation carrying none of the three markers → generator exits non-zero naming the operationId.
2. Operation carrying more than one → error (two statements of one fact).
3. `x-authz-capability` string not in the IAM registry
   (`internal/modules/iam/domain/catalog.go`) → error.
4. `x-authz-capability-none` with an empty or whitespace reason → error.
5. Operation carrying more than one OpenAPI **tag** → error. One operation, one owner: §5 check 4
   asserts the mounting publisher's tag equals the operation's declared tag, and that assertion is
   only well-formed if the tag is single-valued. No operation carries two today (verified across
   all 147), so this rule costs nothing now and makes the ambiguity unrepresentable later.
6. An operation in `internal-e2e.yaml` whose method+path collides with one in `openapi.yaml` →
   error. The two documents produce two maps that are merged at boot when e2e is active (§3), and
   a collision must be a **build** failure rather than the `mergedSurface` panic it would otherwise
   be. All six rules run on both input documents.

`Capability` is already a string type (`iam/domain/model.go:73`) whose values are registry
strings (`CapMetricsView Capability = "metrics.view"`, `model.go:129`), so the spec carries the
wire string and the generator resolves it to the Go constant. A typo is a build failure, not a
fail-closed 403 discovered in production.

`x-authz-password-change-allowed` replaces `isPasswordChangeAllowedPath`
(`auth/.../middleware.go:131`). That predicate is exact-method today —
`method == http.MethodGet && path == "/api/v1/auth/me"` and two `MethodPost` clauses — so moving
it onto a per-operation boolean keyed by the *matched pattern* carries one live authorization
delta, named here rather than discovered:

| Operation | Delta |
|---|---|
| `getCurrentUser` (`GET /api/v1/auth/me`) | `HEAD /api/v1/auth/me` matches the `GET` pattern (`$GOROOT/src/net/http/server.go:2484-2486`) and so inherits `allowed = true`. Today HEAD fails the exact-`GET` clause and a must-change-password principal gets **403**; afterwards it is admitted and answers 200 with no body. **Accepted** — HEAD carries no body, so this exposes status and headers only. Pinned by a regression test in §10. |
| `logout`, `changePassword` | none. Both clauses are `MethodPost`, and a `POST` pattern does not absorb HEAD — only `GET` does. Structurally immune. |

No test in `internal/modules/auth/delivery/http` exercises HEAD on any of these paths today
(`handler_method_not_allowed_test.go:33-55` and `middleware_test.go:89-111` use GET/POST only),
so this delta is currently unguarded in both directions — which is why §10 adds the test rather
than assuming one exists.

**Why the spec and not Go.** Operator ruling A. The structural reason: the spec is the only
enumeration that *cannot* drift from the wire, because it already generates the DTOs and the
`ServerInterface` that every handler must implement to compile. Hanging policy on it hangs
policy on the one artifact the compiler already forces to be correct.

---

## 3. The generator

`cmd/gen-http-surface` — a Go program, run by `go generate`, gated in CI by a
regenerate-and-diff check.

**The existing drift gate does not cover this file, and saying it does was wrong.** An earlier
draft claimed the check followed "the pattern `make openapi-verify` already uses"; no such target
exists — the `Makefile` has five targets (`up down logs test test-watch`, `Makefile:3`) and none
touches OpenAPI. The real backend gate is
`.github/workflows/api-contract.yml:37-38`: `go generate ./...` followed by
`git diff --exit-code -- '**/api.gen.go'`. That pathspec matches only files named exactly
`api.gen.go`, so a new generated file escapes it silently while still being regenerated. Step 2
of §11 therefore widens the pathspec to cover the emitted surface file explicitly; leaving the
generated policy table outside the drift gate would reproduce the class this program exists to
delete.

**Input:** one spec document, plus the IAM capability registry for validation. It runs twice —
over `api/openapi/v1/openapi.yaml`, and (from §11 step 4 onward) over
`api/openapi/internal-e2e.yaml`.

**Output:** exactly one file per input document — `apps/api/cmd/metaldocs-api/httpsurface_gen.go`
for the public spec, and `httpsurface_e2e_gen.go` for the second. Same generator, same validation
rules, same emitted shape; only the input and the output filename differ. Both are covered by the
widened drift pathspec (§11 step 2).

**Neither descriptor carries a build tag, but the *publisher* does — and getting that split wrong
is a boot-loop.** Two earlier drafts were wrong here in opposite directions, and the second one is
worth stating because it is the sharper lesson.

Draft 1 said the e2e file was "build-tagged with the e2e gate". There is no such tag.
`METALDOCS_E2E` is a **runtime** env check: `E2EEnabled()` (`internal/test/e2e_enabled.go:14`)
reads the variable and nothing else, and `mountE2EHandlersIfEnabled` (`main.go:159-166`) branches
on it at boot.

Draft 2 concluded from that "so both descriptors are untagged and one runtime `if` selects between
them — the assertion and the mounts are selected by the same predicate." **That is false, and it
would refuse to boot production.** The mounted side is gated by **two** predicates, not one. The
only e2e build tags in the repo are `integration && !production` on `internal/test/e2e_seed.go:1`
with a `!integration && !production` stub at `e2e_seed_stub.go:1` — and production builds carry
**no tags at all** (`deploy/docker/api.Dockerfile:6` runs a plain `go build`; the repo contains no
`-tags production` anywhere), so the shipped binary compiles the **stub**, a no-op
(`e2e_seed_stub.go:11`). Set `METALDOCS_E2E=1` on that binary and the env-only predicate says yes,
the merged table declares five e2e patterns, the stub mounts zero, check 3 (declared ⊆ mounted)
fails, and `log.Fatalf` turns a leaked environment variable into an unrecoverable boot loop.

**The fix is one boolean governing both sides**, and it is reached by making the *publisher* the
build-tag-selected artifact rather than the descriptor:

```go
// httpsurface_e2e_publisher.go       //go:build integration && !production
func e2ePublisher() httprouter.SurfacePublisher { return newE2EPublisher() }

// httpsurface_e2e_publisher_stub.go  //go:build !integration || production
func e2ePublisher() httprouter.SurfacePublisher { return nil }
```

```go
e2e := e2ePublisher()                        // nil in every build without the handlers
useE2E := e2e != nil && e2eHandlersEnabled() // ONE value decides both sides
surface := httpSurface
if useE2E {
	publishers = append(publishers, e2e)
	surface = mergedSurface(httpSurface, httpSurfaceE2E)
}
assertSurface(mounted, surface, activeTags(publishers), publishers)
```

`useE2E` is the *only* thing either side reads, so the four cells of the (build tag × env flag)
matrix are all total: production with the flag set is `e2e == nil` ⇒ nothing declared, nothing
mounted, the variable inert. The descriptors stay untagged — `httpSurfaceE2E` is simply an unused
package-level map in builds where `e2ePublisher()` returns nil, which Go permits — so the generator
emits one shape and the drift gate covers both files.

Two details that are not incidental:

- **The stub's tag is `!integration || production`, the exact complement of the real one.** The
  existing `internal/test` pair uses `!integration && !production`, which leaves a hole: a literal
  `-tags production` build satisfies neither file and would not compile. No build in the repo uses
  that tag today, so the hole is latent and pre-existing — but the new pair must not inherit it,
  and a total complement costs nothing.
- **Generator validation rule 6** rejects a method+path collision between the two spec documents.
  Draft 2's only guard was a `mergedSurface` panic at boot; a collision authored into
  `internal-e2e.yaml` should fail the **build**, in the same pass that already rejects the other
  five malformed-declaration classes.

```go
// Code generated by cmd/gen-http-surface. DO NOT EDIT.
package main

// httpSurface maps a mux pattern to its tier-1 rule. The key is the exact
// string the mux registers, so a lookup is a map hit, not a second matcher.
type surfaceRule struct {
	visibility                  iamdelivery.Visibility
	capability                  iamdomain.Capability
	tag                         string // the owning publisher's tag — §5 check 4
	allowedDuringPasswordChange bool   // from x-authz-password-change-allowed
}

var httpSurface = map[string]surfaceRule{
	"GET /api/v1/metrics": {
		visibility: iamdelivery.VisibilityPermissionGuarded,
		capability: iamdomain.CapMetricsView,
		tag:        "observability",
	},
	// … one entry per operationId
}

// specTags is every tag in the spec. Boot asserts each has a mounted publisher.
var specTags = []string{"approval", "audit", "auth" /* … */}
```

**The load-bearing detail.** oapi-codegen's generated `HandlerWithOptions` registers
`http.MethodGet + " " + options.BaseURL + path` — a literal concatenation, spot-checked on
simple paths (`internal/modules/security/api/api.gen.go:384-407`) and on two-parameter paths
(`templates/api/api.gen.go:1629-1638`, `iam/api/api.gen.go:1699-1700`). The generator builds its
key by the same formula, from the same spec, with the same `BaseURL` constant. Spot checks are
not the guarantee: §10's generator-level test compares every emitted pattern against all 147
normalized spec operations, and §5's boot assertion proves the bytes match on every real boot.

**HEAD.** No `HEAD` entry is emitted and none is needed. `mux.Handler(r)` returns
`"GET /api/v1/documents/{id}"` for a HEAD request (probe, §9). Keying on the pattern gives HEAD
the GET rule's capability as a consequence of the structure. The defect is not fixed; it stops
being expressible.

---

## 4. Module exposure

Operator ruling C: the module owns mounting, the composition root owns authorization.

```go
// internal/platform/httprouter
type SurfacePublisher interface {
	Name() string      // stable identity, used in boot assertion messages
	Tag() string       // the OpenAPI tag this module owns
	Mount(Muxer)       // registers every route the publisher serves
}
```

The role is deliberately **not** called `Module`. This repo reserves that word for the 15
bounded-context modules (`CLAUDE.md:35`), and three implementers here — health, observability,
configuration — are `internal/platform/*` packages that are emphatically not bounded contexts.
`SurfacePublisher` names an HTTP-mounting role, which is what it is.

`Muxer` already exists and is unchanged (`internal/platform/httprouter`).

**Tag inventory** — 16 tags cover all 147 operations, and every tag maps to exactly one owner:

| Tag | Ops | Tag | Ops |
|---|---|---|---|
| documents | 29 | audit | 4 |
| iam | 26 | security | 3 |
| approval | 20 | distribution | 3 |
| templates | 19 | health | 2 |
| taxonomy | 16 | search | 1 |
| controlled-documents | 9 | observability | 1 |
| tokens | 5 | configuration | 1 |
| notifications | 4 | auth | 4 |

`health`, `observability`, and `configuration` are today served by `internal/platform/*`
packages. They become `httprouter.SurfacePublisher` implementations in place; being platform
packages does not exempt them, and a tag with no publisher is a boot failure by §5 check 1.

No operation carries more than one tag (verified across all 147), so `Tag() string` is sound.
Should a future operation need two, the generator rejects it — §2 validation rule 5, one
operation, one owner.

Tag ownership cuts **across** path prefixes, and that is the point: `POST /api/v1/documents/{id}/obsolete`
is tagged `approval` (`openapi.yaml:3739-3743`) and is mounted by the approval module, not the
documents one. Path prefix is not ownership; the tag is.

### Mount is total — conditional mounting is unrepresentable

**A publisher's `Mount` registers every route it owns, unconditionally. Availability is a
handler-level `501`, never a routing-level absence.** This is a protocol rule, not a per-site
fix, and it is load-bearing for §5 check 3 (declared ⊆ mounted): a mount that only happens on
some boot path makes the surface env-dependent, which is the exception mechanism ruling C
deleted.

The repo already contains both strategies, and the protocol picks the one that is already the
majority:

- **`iamdelivery.Router` mounts unconditionally and answers 501 when a dependency is nil** — the
  file has **nine** `writeIAMNotImplemented` call sites, of which **eight are nil-dependency
  guards** and one is unconditional. The eight: `sessions` 2 (`router.go:130`, `:140`),
  `observability` 2 (`:209`, `:219`), `presence`'s own *snapshot* 1 (`:232`), `tenants` 3
  (`:300`, `:310`, `:320`). The ninth, `createManagedUser` (`:248`), is a **deprecated operation
  that answers 501 unconditionally** — not an availability guard, and counted separately because
  it is a different mechanism that happens to share a status code.
- **One site skips the mount call entirely** — `router.go:104-106` guards
  `presence.RegisterRoutes` behind `h.presence != nil`, and `startPresence` returns nil whenever
  `deps.SQLDB == nil` (`main.go:1173-1175`). `conditionalRouteFamilies = map[string]bool{"presence": true}`
  (`router.go:85`) records that this is the *only* such site.

So on the SQLDB-less boot path `GET /api/v1/iam/presence/stream` — a declared spec operation
(`openapi.yaml:637-647`) — is not mounted at all, and check 3 would fail the boot. Under this
rule the stream folds into the IAM publisher alongside its snapshot sibling and answers 501 when
presence is unavailable, exactly as `GetPresenceSnapshot` already does.

Three consequences worth stating plainly:

1. `conditionalRouteFamilies` and its fail-open lookup (`router.go:85`) are **deleted** — §7. The
   held review finding against that construct dissolves rather than being patched.
2. One accepted behavior change: on the SQLDB-less path, `/api/v1/iam/presence/stream` answers
   **501 instead of 404**, matching its snapshot sibling. Risk row in §12, test row in §10.
3. `streamPresence` stays hand-mounted (it is a WebSocket upgrade, excluded from *server codegen*
   via `iam/api/cfg.yaml:11-12`) but is still a spec operation, so the generator emits its rule
   and all four checks cover it. Excluded-from-codegen is not excluded-from-surface.

**One tag, one publisher — a bijection, and check 1 enforces it strictly.** An earlier draft
claimed a tag may have more than one publisher, citing `iam`'s two mount sites (the generated
router and the hand-mounted stream). That clause directly contradicted check 1 (§5), which rejects
two publishers claiming one tag — the boot assertion would have refused the design's own topology.
The clause is deleted, not the check: `iam`'s two mount sites are an artifact of the conditional
presence mount, and **§11 step 5 folds the stream into the single IAM publisher**, which is what
makes one-publisher-per-tag true before anything asserts it. Ownership is a bijection because
check 4 (mounted pattern's tag == mounter's `Tag()`) is only meaningful if it is.

The composition root holds a **list**, not a struct:

```go
publishers := []httprouter.SurfacePublisher{
	authmod.New(...), auditmod.New(...), /* … */
}
```

This closes the `main.go` keyed-literal hole structurally: a forgotten publisher is a missing
list element, and §5's tag-coverage assertion fires. There is no field to leave unset.

No module imports `iamdomain` to declare its own capability — capability is IAM's published
concept, and a module redeclaring it is catalog §10. Publishers declare *routes*; the root
declares *policy*. This also settles locked constraint 9 (`…-system-impact.md:208`): **no new
module→iam edges are created**, because the capability binding never enters a module.

---

## 5. Boot assertion

Mounting runs through a recording `Muxer` in **production**, not only in a test:

```go
mounted := map[string][]string{}          // publisher name → patterns it registered
for _, p := range publishers {
	rec := httprouter.NewRecorder(mux)    // one recorder per publisher
	p.Mount(rec)
	mounted[p.Name()] = rec.Patterns()
}
// surface and publishers are BOTH derived from useE2E (§3) — never httpSurface
// bare, or the e2e boot path asserts a table its own mounts are absent from.
if err := assertSurface(mounted, surface, activeTags(publishers), publishers); err != nil {
	log.Fatalf("http surface: %v", err)   // fail closed at boot
}
```

`assertSurface` checks four things and refuses to start on any failure:

1. **Tag coverage** — every entry in `specTags` is claimed by exactly one publisher's `Tag()`.
   Catches a publisher constructed but not listed, and two publishers claiming one tag.
2. **Mounted ⊆ declared** — every recorded pattern has a `httpSurface` key. Catches a route
   mounted with no policy.
3. **Declared ⊆ mounted** — every `httpSurface` key was recorded. Catches a spec operation
   whose handler was never wired.
4. **Ownership** — for every pattern a publisher recorded, `httpSurface[pattern].tag` equals
   that publisher's `Tag()`.

**Check 4 is not redundant with check 1.** Recording one global pattern set would let every
publisher claim a distinct tag while mounting each other's operations, and checks 1–3 would all
pass: the tag set is covered, the pattern set matches, and nothing is missing. Only per-publisher
recording makes "documents mounts a templates route" a boot failure. That is why the recorder is
per publisher and why `surfaceRule` carries `tag`.

This is not a test hook welded into production (catalog Class 34). The recorder wraps the
*real* mount, the assertion runs on the *real* boot, and the check is the invariant itself
rather than a mirror of it. `TestRouteCoverage` is deleted because the property it approximated
is now enforced by the thing it was approximating.

**Check 3 is total, and that is load-bearing.** All 147 spec operations have a mounted handler
today — 134 generated `ServerInterface` methods, the 12 hand-registered legacy operations §8
enumerates by family (auth 4, security 3, health 2, search 1, configuration 1, observability 1),
and `streamPresence`, which `cfg.yaml` excludes from server codegen because it is a WebSocket
upgrade but which is nonetheless mounted method-qualified (`iam/presence/handler.go:73`). So
check 3 has no false positive on day one, and no "declared but not implemented" escape is
needed. Adding one would be the exception mechanism ruling C deleted.

**E2E scaffolding becomes a publisher with its own generated declaration.** `/internal/test/*`
is env-gated (`METALDOCS_E2E`) and today is mounted onto the API mux *after* `buildRouter`
(`main.go:839-848`), deliberately outside route truth (`router.go:33-38`). Two earlier drafts
were wrong about it:

- *A rule overlay merged into `httpSurface` when the flag is on* — rejected. Conditional
  membership in the governed set is precisely the exception mechanism ruling C deleted.
- *A separate listener* — rejected on cost, once the actual consumers were read. Five e2e
  endpoints (`internal/test/e2e_seed.go:109-114`) are called from at least eight Playwright
  sites as **relative** paths against the page's own `baseURL`
  (`e2e/utils/seed.ts:27,46`, `fixtures/isolation.ts:50,72`, `flows/happy_path.spec.ts:39,146`,
  `flows/sod_violation.spec.ts:87,279,309`, `flows/reject_flow.spec.ts:49`,
  `flows/quorum_m_of_n.spec.ts:61`). A second port makes every one of them absolute.

What ships instead: the scaffolding is an ordinary `SurfacePublisher` with tag `internal-e2e`,
and its declaration is **generated by the same generator from a separate spec document**,
`api/openapi/internal-e2e.yaml` — excluded from the public bundle and from frontend codegen, so
test scaffolding never enters the customer contract or the FE types. Under `METALDOCS_E2E` the
publisher is in the list and its tag is in `specTags`; otherwise neither is. Both sets move
together, so all four checks stay total on both boot paths and nothing is exempted from
checking.

The distinction that matters: an exception is *mounted but not checked*. This is a second
complete, generated, asserted surface — which honors locked constraint 1 (generated, never
hand-written) as well.

**Excluding it cleanly is verified, not assumed — but exclusion is exactly the hazard.** Every
current consumer names `v1/openapi.yaml` explicitly, not a glob: backend codegen
(`internal/modules/iam/api/gen.go:3` and the other 14 `gen.go` directives), frontend codegen
(`frontend/apps/web/package.json:14`), the Redocly lint job
(`.github/workflows/api-contract.yml:62`) and the repo's own `api-lint` guards (`:100`). So a
second document under `api/openapi/` is swept into nothing — it lands *outside every gate*, which
is the failure mode, not the goal. Three explicit additions close it, each landing in the step
that creates what it gates:

1. `api-lint` and the Redocly lint job each run over `internal-e2e.yaml` as a second target with
   the same ruleset — **§11 step 4**, the step that authors the document.
2. The generator's validation tests treat it as a first-class input; a malformed `x-authz-*` in
   it fails the same way it would in the public spec — **§11 step 4**, same reason.
3. The drift gate covers the files the generator emits — **§11 step 2**, which is where the
   generator itself lands (see the pathspec widening in §11).

The ordering matters and is not cosmetic: items 1 and 2 cannot be in step 2, because
`internal-e2e.yaml` does not exist until step 4. A gate authored before its subject is a gate
that passes vacuously.

The one thing it stays excluded from is the **public bundle and the FE types** — that exclusion
is the point of the second document and is asserted by the bundle target continuing to name only
`v1/openapi.yaml`.

**The panic probe is deleted, not moved.** `GET /internal/test/panic` (`main.go:845-847`) has
exactly one occurrence in the repository — its own registration. No workflow, spec, or
Playwright flow ever requests it, so it has never verified REQ-MW-1; it is a guarded branch no
environment executes (catalog §22). It goes, and REQ-MW-1 is verified where it can actually
fail: a Go test that drives the composed API handler and asserts a panicking handler yields a
500 problem+json.

---

## 6. Tier-1 enforcement

The resolver stops matching and starts looking up.

```go
// iam/delivery/http — signature change
type PermissionResolver func(*http.Request) (iamdomain.Capability, Visibility)
```

Composition-root implementation:

```go
func newPermissionResolver(mux *http.ServeMux) iamdelivery.PermissionResolver {
	return func(r *http.Request) (iamdomain.Capability, iamdelivery.Visibility) {
		_, pattern := mux.Handler(r)
		if pattern == "" {
			// The mux matched nothing: this request is a 404 or a 405, not a
			// route with a policy. Demand a session — identical to today's
			// fall-through (permissions.go:336-338) — and let the mux emit the
			// status, which method_not_allowed rewrites into problem+json.
			return "", iamdelivery.VisibilitySessionRequired
		}
		rule, ok := httpSurface[pattern]
		if !ok {
			// A pattern was matched but carries no rule. §5 makes this
			// unreachable at boot; if it happens it is a wiring bug, not a
			// tier to guess at.
			return "", iamdelivery.VisibilityUnresolved
		}
		return rule.capability, rule.visibility
	}
}
```

**The two unresolved cases are not the same case**, and conflating them was a defect in an
earlier draft. `Middleware.Wrap` switches on visibility and 500s anything it does not recognize
(`iam/delivery/http/middleware.go:90-96`). Returning a single "unresolved" for both would turn
every authenticated 404 and 405 into a 500 the moment the resolver flips — an outage manufactured
by a design meant to change nothing observable. Splitting them means:

- **no pattern** → `VisibilitySessionRequired`, byte-identical to today's behavior for a
  nonexistent path;
- **pattern without rule** → a new `VisibilityUnresolved` enum value, nonzero, appended after
  `VisibilityPublic` so `VisibilityPermissionGuarded` keeps `iota` = 0 as the fail-closed zero
  value.

`VisibilityUnresolved` gets its **own explicit case** in `Middleware.Wrap`'s switch, returning the
same 500 + `slog.Warn` the `default:` arm returns today. Leaving it to fall through `default:`
would work and would be wrong: a named, expected value silently relying on an
unknown-value branch is a guard a future author closes by "completing" the switch, at which point
an unwired route becomes a pass-through. The `default:` arm **stays** for genuinely unknown
values, and §10 tests both arms.

**The chain is not touched** (locked constraint 6, `…-system-impact.md:205`; `chain.go:25`). An
earlier draft added a `route_resolve` link that resolved once and passed the rule through the
request context. Three things died with it: a violated lock, a context key whose trust boundary
had to be argued ("who else can set this?"), and a new lifecycle stage. Each of the three
consumers instead resolves on demand:

| Consumer | Needs | How |
|---|---|---|
| `pre_auth_login_rate_limit` | is this login | unchanged — compares against `authdelivery.PathLogin` (`chain.go:64`) and never needed the surface |
| `authn` | public? password-change-allowed? | two `func(*http.Request) bool` closures over the same lookup |
| `iam_authz` | capability + visibility | `PermissionResolver` above |

**Accepted cost:** two extra `mux.Handler(r)` calls per request (the mux does a third when it
serves). That is a pointer-tree walk against a per-request DB round trip, and it buys not
touching a locked structure.

**The auth-side consumers collapse into the same table.** `newPublicPathChecker`
(`permissions.go:352`) already derives from the resolver; its signature changes from
`(method, path)` to `*http.Request` with it. `defaultPublicPaths` is deleted and
`WithPublicPathChecker` becomes mandatory (nil → boot failure, not a silent fallback).
`isPasswordChangeAllowedPath` becomes `rule.allowedDuringPasswordChange`.

No new module→module edge: the rule struct stays in `package main`, and each consumer receives
a narrow function. The composition root is the only place holding the whole rule.

---

## 7. Deletions

Ruling A forbids a two-regime ship, so these go in the same program.

| Deleted | Location | Replaced by |
|---|---|---|
| `routeRules` (120 rows) | `permissions.go:82-330` | generated `httpSurface` |
| `routeRule` + `matches()` | `permissions.go:23-55` | `mux.Handler(r)` |
| `resolveRoutePermission` | `permissions.go:342` | map lookup |
| `defaultPublicPaths` | `auth/.../middleware.go:118` | `security: []` in the spec |
| `isPasswordChangeAllowedPath` | `auth/.../middleware.go:131` | `allowedDuringPasswordChange` |
| `routeHandlers` struct | `router.go:39-59` | `[]httprouter.SurfacePublisher` |
| `routeFamilies` + `routeFamily` | `router.go:61-127` | `SurfacePublisher.Mount` |
| `conditionalRouteFamilies` | `router.go:85` | — (fail-open exemption, no successor) |
| `buildRouter` | `router.go:134` | the mount loop in §5 |
| `TestRouteCoverage` + `routeHandlerFields` | `permissions_test.go:351-481` | §5 boot assertion |
| `/healthz` mount + rule | `observability/health.go:20`, `permissions.go:85` | `/api/v1/health/live` |
| `if r.Method != X { WriteMethodNotAllowed }` in every migrated legacy handler | e.g. `auth/.../handler.go:98`, `:146` | the method-qualified pattern; the mux rejects |
| `x-rate-limit` | `openapi.yaml:3213` | — dead: no consumer; the real config is `ratelimit/config.go:35` |
| `x-websocket-message` | `openapi.yaml` (1 op) | — dead: no consumer |
| `GET /internal/test/panic` | `main.go:845-847` | a Go test over the composed handler (§5) |
| `frontend/apps/web/e2e/playwright.approval.config.ts` | whole file | — dead: no workflow or package script selects it |
| **15 orphaned `RegisterRoutes` methods** left behind by past codegen migrations — including `pdf_webhook_handler.go:41`, whose *whole file and test* go too | §7.1 table | the generated `HandlerWithOptions` mount that already supersedes each (except the PDF webhook, which has no successor — see §7.1) |

The last five rows are the extermination ruling applied to this surface. `x-rate-limit` and
`x-websocket-message` look like policy and are read by nothing; a declaration nobody consumes is
indistinguishable from one that is enforced, which is worse than no declaration. `x-authz-area`
stays — it is live, consumed by a blocking test (`permissions_authz_scope_test.go:76`), as does
`x-pagination-exempt`, consumed by `scripts/api-lint/spec_rules.go`.

The Playwright config is the sharper case, because an earlier draft listed it as a `/healthz`
consumer **to be repointed**. It is referenced by nothing; its `START_BACKEND` knob
(`:11`) is set by nothing and its branch shells out to `./cmd/api` (`:42`), a path that does not
exist. Repointing a dead file would have been the pure form of the defect this program exists to
remove: maintaining a copy of route truth inside an artifact nobody runs.

### 7.1 The orphaned mounting methods — the defect class this program exists to delete

An independent sweep of every delivery seam found **15** exported `RegisterRoutes` methods that no
production code path reaches. (An earlier draft said 13. The number was low because the
verification prompt handed the reviewer a *fixed list to check* instead of asking it to sweep —
a prompt that can only confirm, never extend. Asked to sweep, it returned two more. The prompt
shape was the defect, not the reviewer.) They are not a scattering of small oversights; they are one defect
with one cause, repeated: **each past codegen migration added the generated
`HandlerWithOptions` mount and left the hand-written mounting method in place.** The generated
router dispatches straight to the sub-handler's *business* method, so the old `RegisterRoutes`
became a second, silent copy of route truth that nothing executes.

The codebase says so about itself. `iam/delivery/http/router.go:102-113` names the exact five
`RegisterRoutes` call sites its `RegisterGenerated` replaces — and all five are still there.

| Location | Reached by | Live mount instead |
|---|---|---|
| `documents/.../export_handler.go:41` | test only (`export_handler_test.go:16`) | `routes_generated.go:115-131` → `h.export.exportDocxURL` / `exportPDF` |
| `documents/.../export_handler.go:46` (`RegisterRoutesWithRateLimit`) | **nothing at all** | same |
| `documents/.../fillin_handler.go:39` | test only (`fillin_handler_test.go:41`) | `routes_generated.go:133-168` |
| `documents/.../view_handler.go:28` | **nothing at all** | `routes_generated.go:211-218` → `h.view.HandleView` |
| `documents/.../reconstruct_handler.go:28` | **nothing at all** | `routes_generated.go:171-178` |
| `documents/.../placeholder_options_handler.go:36` | **nothing at all** | `routes_generated.go:142-150` |
| `documents/module.go:157` (`RegisterRoutes`, no rate limit) | test only (`module_wrapper_test.go:138`, `:179`) | production calls `RegisterRoutesWithRateLimit` exclusively (`router.go:113`) |
| `documents/.../handler.go:144` (`Handler.RegisterRoutes`) | **transitively dead** — 4 tests (`handler_test.go:231`, `:421`, `handler_comments_test.go:124`, `handler_pagination_test.go:96`) plus the already-dead `module.go:158` | `RegisterRoutesWithRateLimit` (`handler.go:146`), the only production entry (`router.go:113`) |
| `iam/.../admin_handler.go:132` | **nothing at all** | `Router.RegisterGenerated` (`iam/.../router.go:114`) |
| `iam/.../people_handler.go:54` | **nothing at all** | same |
| `iam/.../observability_handler.go:31` | **nothing at all** | `Router.GetKpi` / `Router.GetUsage` (`iam/.../router.go:205-223`), which delegate to the same two unexported methods and answer 501 when unwired |
| `iam/.../routes_memberships.go:102` | test only (`routes_memberships_contract_test.go:127`, `:156`) | same |
| `iam/.../routes_roles_caps.go:51` | test only (`routes_roles_caps_test.go:201`) | same |
| `iam/.../sessions_handler.go:96` | test only (9 call sites in `sessions_handler_test.go`) | same |
| `documents/.../pdf_webhook_handler.go:41` | **nothing at all** | **nothing** — see below |

`observability_handler.go:31` is the one that most deserved to be found by a sweep rather than a
checklist: it is the *last* pre-codegen IAM mounting method, still typed `*http.ServeMux` where
every sibling now takes `httprouter.Muxer`, which is precisely the signal that the migration
walked past it.

`handler.go:144` is the transitive case — a method whose only non-test caller is itself dead
(`module.go:158`). Deleting it is safe for rate limiting specifically: the limiter lives in
`registerRoutes`'s `rl`/`userFn` parameters and the `rateLimitedRoutes` map
(`handler.go:138-141`), both of which belong to `RegisterRoutesWithRateLimit` and are untouched.
The nil-nil variant existed only so a test could skip constructing a limiter.

All 15 are deleted. **Their tests are re-pointed by assertion class, not moved wholesale** — an
earlier draft said they "move to invoking the business method directly", which would have silently
dropped coverage:

| What the test asserts | Where it goes |
|---|---|
| Business behavior — the service is called, the DTO is shaped, the error maps | the business method directly. This is what the majority of sibling tests already do; `view_handler_test.go`, `rbac_test.go` and `reconstruct_handler_test.go` never used `RegisterRoutes` in the first place |
| **Routing** — typed path validation, method dispatch, 404/405, the value of `r.Pattern` | **stays on a mux**, re-pointed to the generated `HandlerWithOptions` mount that supersedes the deleted method. Same request, same assertions, different registrar |
| **Wrapper forwarding** — that `documents.Module` delegates to its inner handler (`module_wrapper_test.go:176`, `:184`) | stays, re-pointed to `RegisterRoutesWithRateLimit` (`handler.go:169`), the only production entry |

A direct call cannot assert `r.Pattern`, cannot exercise oapi-codegen's typed path-parameter
validation, and cannot produce a 405 — so for those three classes the mux is not a detail of the
test, it is the subject. Deleting the *registrar* is the goal; deleting the *coverage* is not, and
the two were conflated. Each deletion commit states which of the three rows each moved test landed
in, so the reviewer can check the count rather than trust the sentence.

`pdf_webhook_handler.go` is the sharpest instance and is deleted outright, tests included: it is a
complete HMAC-authenticated endpoint (`POST /api/v1/documents/{id}/pdf-complete`) that is absent
from the OpenAPI spec, mounted by nothing, and armed by no env var or config knob. Its own header
comment (`:28`) concedes *"its RegisterRoutes is not called anywhere — it is currently UNWIRED."*
Contract-first makes an off-spec route illegal; the extermination ruling makes "documented as
dead" not a resting state. Keeping it is strictly worse than deleting it — it is a security
surface that looks maintained.

**Why this belongs to this program and not a follow-up.** Under §5, mounting is the only way a
route enters the surface, and `assertSurface` reads what publishers actually mounted. A method
that mounts nothing therefore *cannot* be reached by any check the protocol adds — it would
survive the restructure invisibly, which is precisely how it survived the last three migrations.
The program that makes mounting authoritative is the one that must delete the mounts that lie.

**One comment becomes false in the same commit and is rewritten in it.**
`internal/platform/httprouter/muxer.go:11-14` documents `Muxer` by naming its consumer:
*"a recording wrapper used by TestRouteCoverage (`apps/api/cmd/metaldocs-api/permissions_test.go`)
… letting the test observe every pattern the real server mounts instead of relying on a
hand-maintained fixture list."* §11 step 6 deletes `TestRouteCoverage`, so that sentence would
point at nothing. It is not a cosmetic edit: the comment states *why the type exists*, and after
this program the reason is different and stronger — the recorder is no longer a test fixture, it
is the boot-time surface recorder `assertSurface` consumes, and the wrapper runs in production
startup, not in a test. The rewrite lands in the same commit as the deletion, per the fail-closed
rule that a guard and the text describing it never diverge across a commit boundary. This is the
same defect class as the four known-false comments already on the list; all five are corrected or
deleted with the code they describe:

| Comment | What is false |
|---|---|
| `router.go:72` | names `buildPresence`, which does not exist — the real name is `startPresence` (`main.go:1168`) |
| `permissions_test.go:415-417` | says five bare-pattern families where there are six |
| `router.go:78-80` | claims a loud boot failure that does not happen |
| `iam/.../router.go:102-113` | says `RegisterGenerated` "replaces the **six** `RegisterRoutes` call sites this router supersedes" and then names **five**, omitting `ObservabilityHandler` (§7.1). Worse, the same comment asserts the swap "changes no tier-1 behavior" because tier-1 "keys off `r.Method`/`r.URL.Path`, not off mux dispatch mechanics" — which is exactly the premise this program deletes. It is rewritten, not recounted |

**One thing the sweep cleared, worth recording so it is not re-litigated:**
`mountE2EHandlersIfEnabled` / `METALDOCS_E2E` is **not** dead — `e2e_gate_test.go` exercises both
branches and CI sets the variable. It becomes a publisher (§5); it is not deleted.

So `/healthz` has **four** live consumers, all repointed to `/api/v1/health/live`:
`.github/workflows/e2e-coverage-gate.yml:161`, `perf.yml:34`, `perf.yml:65`,
`frontend/apps/web/playwright.config.ts:70`.

---

## 8. Legacy family migration

Six families register bare, method-less patterns. A bare pattern returns one string for every
method (§9 probe), so it cannot carry a per-method rule — every one of them must at minimum
become method-qualified.

| Family | Bare patterns | Anchor |
|---|---|---|
| auth | 4 | `auth/delivery/http/handler.go:84-87` |
| security | 3 | `security/delivery/http/handler.go:73-75` |
| health | 2 (+`/healthz`, deleted) | `platform/observability/health.go:18-20` |
| search | 1 | `search/delivery/http/handler.go:63` |
| featureflags — spec tag **`configuration`** | 1 | `platform/featureflags/handler.go:26` |
| observability (`/api/v1/metrics`) | 1 | `router.go:123-125` — bare `mux.Handle` in the composition root |

**12 bare patterns, 12 spec operations across those six tags** — auth 4, security 3, health 2,
search 1, configuration 1, observability 1, matching §4's tag inventory exactly. The Go package is
`platform/featureflags`; its spec tag is `configuration`. They are the same slice under two names,
and the tag is the one that governs — §5 check 4 compares tags, never package names. (An earlier
draft said 15 operations; that was arithmetic, not measurement.) `presence` is *not* on this list:
its one hand-mounted route is already method-qualified (`iam/presence/handler.go:73`) and is
excluded from server codegen deliberately, because it is a WebSocket upgrade
(`iam/api/cfg.yaml: exclude-operation-ids: [streamPresence]`). It stays hand-mounted and stays
conformant — §5 asserts *patterns*, not how they were registered, so "generated" is not a
requirement the protocol imposes and `streamPresence` is not an exception to it.

**Method-qualifying alone would satisfy the assertion. It is still not what ships.** Every one of
these legacy handlers is 1:1 with a single method (`handleMe` rejects non-GET at
`auth/.../handler.go:146`), so registering each under its method prefix would be sound and far
cheaper than codegen migration. An earlier draft argued migration was *forced*; that argument was
wrong and is withdrawn. The conclusion stands on a different and stronger basis: stopping at
method-qualification leaves two ways to mount a route — generated for ten tags, hand-written for
six — which is the two-regime state ruling A forbids, kept alive purely because it is cheaper.
Cost is not a reason to keep a second regime.

So each family gains an oapi-codegen `ServerInterface` and mounts via `HandlerWithOptions`,
matching the ten already-migrated tags, and the internal `if r.Method != X` guards are deleted
(§7) rather than left as dead branches. `security` already has a generated
`internal/modules/security/api/api.gen.go` that nothing wires up — that dead artifact becomes
live rather than being written from scratch.

**One accepted behavior change, in health.** `handleLive` and `handleReady`
(`platform/observability/health.go:23,32`) check no method at all: today
`POST /api/v1/health/live` reaches the handler. The tier-1 row is GET-only
(`permissions.go:84`), so such a request is session-gated and then answered 200. After
method-qualification the mux answers 405. This is a deliberate tightening, listed here and in
§10's delta table rather than discovered — every other family already rejects non-matching
methods itself, so health is the only one where mounting changes what a client sees.

---

## 9. Edge semantics

Measured against Go 1.25 `net/http` (`go.mod:3`), not assumed:

| Request | `mux.Handler(r)` pattern | Served |
|---|---|---|
| `GET /api/v1/documents/abc` | `GET /api/v1/documents/{id}` | 200 |
| `HEAD /api/v1/documents/abc` | `GET /api/v1/documents/{id}` | 200 |
| `POST` on a GET-only path | `""` | 405 |
| `OPTIONS` on a GET-only path | `""` | 405 |
| trailing slash `…/abc/` | `""` | 404 |
| `//` or `/../` in path | real pattern | 307 to the cleaned path |
| `%2F` in a path segment | real pattern | 200 |
| bare pattern, any method | the bare pattern | handler decides |

Two consequences worth stating:

- **No redirect bypass.** The `//` and `/../` cases resolve to the *cleaned* path's pattern, so
  the 307 target re-authorizes against the same rule. An attacker cannot use path noise to get
  one pattern authorized and another served.
- **OPTIONS.** Preflight resolves to `""`. `cors` runs before `iam_authz` in the chain
  (`chain.go:29` vs `:33`) and must terminate preflight before the resolver sees it. This is an
  implementation check, not an assumption.

---

## 10. Testing

| Property | How |
|---|---|
| Generator validation rules (§2, five failures) | table test per rule, each asserting non-zero exit and the operationId in the message |
| `/api/v1/iam/presence/stream` on a SQLDB-less boot answers **501**, not 404 | boot the router with `presence == nil` and assert status 501 with a `problem+json` body — the Mount-is-total rule (§4) is only real if the degraded path is exercised |
| Generated output is current | CI regenerate-and-diff; drift fails the build |
| **Tier-1 derivation proof — total by construction over all 120 rules** (§10.1–10.3) | for each of the 120 `routeRules` rows, compute the generated-table entries it governs by reading pattern **skeletons**, and assert the verdicts agree or the pair is a listed delta. Three completeness assertions: every rule reached, every pattern covered, every governed pair agrees. No request enumeration, so no dimension to omit. Locked constraint 5 / `…-system-impact.md:140`. |
| **The delta list is complete** (§10.3, eight rows) | each delta has its own regression test, and the derivation proof fails on any *unlisted* disagreement — completeness is the gate's output, not a claim beside it |
| Generator key == emitted pattern, exhaustive | compare the generator's key for all 147 operations against the patterns oapi-codegen actually emits, rather than reasoning from three sampled paths |
| Ownership (§5 check 4) | a publisher mounting a pattern whose declared tag is another publisher's fails `assertSurface` |
| `VisibilityUnresolved` → 500, unknown value → 500 | two middleware tests, one per switch arm |
| Health non-GET now 405 | table test asserting the accepted delta from §8, so it cannot regress silently |
| Boot assertion fires on each of its **four** checks | four unit tests over `assertSurface` with hand-built inputs — the assertion is pure, so no boot needed. One per check: (1) a tag with no publisher / two publishers, (2) a mounted pattern with no declaration, (3) a declaration with no mount, (4) a mounted pattern whose tag ≠ its mounter's `Tag()` |
| HEAD inherits GET's capability | integration test: HEAD a capability-guarded route without the capability, expect 403 |
| Public routes stay public | integration test over every `visibility: public` operation, no session, expect non-401 |
| `%2F`, trailing slash, `//`, `..` | table test against the real mux — this is §10.3 delta 7's regression test, and it asserts the **new** behavior is correct, not that it matches the old one |

### 10.1 The deletion license — why it is a derivation proof and not a differential test

**This section was rewritten under an operator §2 ruling.** Four consecutive review rounds landed
on this gate, three of them at the same altitude, each naming a dimension of the input space the
enumeration did not cover:

| Round | Dimension found missing |
|---|---|
| 2 | **path** — `permissions.go:120`'s `notSuffix: "/roles"` disqualifies `PATCH /api/v1/iam/users/roles` from the `user.manage` row today, while a pattern lookup grants it |
| 3 | **method** — HEAD aliases onto GET patterns; a method mismatch returns pattern `""`; methodless rows are only reachable by varying the method |
| 4 | **method, again** — sampled at one wrong value rather than enumerated; different wrong methods reach different rows |
| 5 | **encoding** — `GET /api/v1/documents/foo%2Fapproval-preview` |

`adversarial-review` §1's two-patch ratchet was exceeded and §8's same-altitude-recurrence rule
fired. A third extension of the basis was **not** applied. The §2 question was put to the operator
and answered: **outcome (a), restructure.**

**Root cause — deeper than any of the four findings.** The differential tried to prove behavioral
equality between two functions that **read different inputs**. Go's `ServeMux` matches on
`r.URL.EscapedPath()` (`$GOROOT/src/net/http/server.go:2659-2680`), where `%2F` stays inside one
segment; the old resolver reads `r.URL.Path`, which is **decoded**, where the same bytes are a real
separator. So `foo%2Fapproval-preview` is one segment to the router (→ `/documents/{id}` →
`document.view`) and two to `permissions.go:163`'s suffix rule (→ `document.submit`).

That is not a hole in the sample. It is the discovery that **where the two disagree, the old
resolver is the wrong one** — it classifies strings the router never routed. Chasing equality over
that input space is chasing a bug into its corners, and the space is infinite: after encoding come
trailing slashes, `..` segments, `;`-parameters, duplicate slashes, and unicode normalization.
**A proof by enumeration over an infinite input space cannot be closed by adding dimensions.**

**§2 verdict.**

| Question | Answer |
|---|---|
| Global-maximum structure | **A row-level derivation proof plus an enumerated delta list.** Compare the 120 `routeRules` rows to the generated table row by row — finite, total, and complete *by construction*, with no input-space enumeration anywhere in it. Every row that does not map identically becomes a named, approved delta with a regression test. |
| What a proven system does | The standard policy-table migration: prove the new artifact is **derived** from the source of truth and enumerate the intentional differences. It is what every codegen migration in this repo already does (`oapi-codegen` + the drift gate), and what §5 checks 2 and 3 already are. Fuzzing two implementations against each other is what you do when there is *no* source of truth — and this program's whole premise is that the spec **is** one. |
| Cost of the global maximum | Rewriting this section. Steps 1–3 and 5 of §11 are unaffected. |
| Cost of the local maximum later | A BLOCKER every round, on a new dimension each time, and a deletion licensed by a gate that has never once been total. |
| Chosen by | **The operator**, 2026-08-05, on the §2 escalation recorded in the review ledger. Not the author, and not the reviewer. |

The request-level differential, the cross-product, the token basis and the sentinel are all
**deleted**. They are not kept as a supplementary check: two gates where one suffices is exactly
the shape §3 of the review doctrine tells us to collapse, and the weaker one would be the only one
anybody read when it went red.

### 10.2 The derivation proof — total by construction, 120 rows

The gate is a pure table-to-table function with no request in it.

```go
// For each of the 120 rows in routeRules, compute the set of generated-table
// entries that row governs, and compare the verdict.
func TestRouteRulesDerivation(t *testing.T) {
	generated := httpSurface // the generated map, keyed by mux pattern
	for _, rule := range routeRules {
		got := entriesGovernedBy(rule, generated) // pattern → surfaceRule
		...
	}
}
```

`entriesGovernedBy` is the only non-trivial piece, and it is **decidable**, which is the whole
point: for a given rule and a given mux pattern, "does this rule govern this pattern?" is answered
by reading the pattern's **static segments**, never by generating candidate request strings.

- `pathExact` — string equality against the pattern's path. Decidable.
- `pathPrefix` — `strings.HasPrefix` on the pattern's static head. Decidable, and **provably
  unaffected by param values**: all 47 prefix rules end at a `/` that sits before the first
  `{param}` of every pattern they cover, and Go 1.22 `{param}` segments cannot contain `/`.
- `pathSuffix` / `contains` / `notSuffix` — evaluated against the pattern's **literal** segments.
  A `{param}` position is treated as the wildcard it is: the rule governs the pattern iff the
  predicate holds for the literal skeleton. A param value that *coincidentally* satisfies the
  predicate is not a derivation question — it is a **delta**, and §10.3 is where it goes.
- `method` — equality, or "any" for the one methodless row (`/healthz`, `permissions.go:85`).

**Three completeness assertions, each finite:**

1. **Every rule is reached.** All 120 rows appear in the iteration; a row governing zero patterns
   is reported, not skipped — that is a rule for a route that no longer exists, and it must be
   accounted for before deletion.
2. **Every pattern is covered.** All 148 recorded patterns (134 generated + 14 hand-registered;
   §7 deletes `/healthz`, taking it to 147) are governed by at least one rule or explicitly listed
   as new.
3. **Every governed pair agrees, or is a listed delta.** The verdict `(Capability, Visibility)` the
   rule yields equals the one the generated entry carries.

There is no sampling anywhere in this construction, so there is no dimension to omit. The counts
above are measured, not assumed: 120 rows (GET 42, POST 53, PUT 11, PATCH 5, DELETE 8, plus **one
methodless**), 148 patterns cross-checked three ways against the spec's 147 operations. When the
test is written, all three are re-derived from the code.

### 10.3 The delta list — the gate's output, not a footnote beside it

Under a derivation proof, **completeness of the delta list is what the gate produces.** Any row
that does not map identically lands here by construction, including every class the four review
rounds found by enumeration. Each delta carries a reason and a regression test; none is absorbed by
a loose assertion.

| # | Delta | Why it exists | Disposition |
|---|---|---|---|
| 1 | Health authenticated non-GET: 200 → **405** | the method-qualified pattern replaces in-handler `r.Method` dispatch (§8) | accepted, table test |
| 2 | `GET /api/v1/health/<unknown>` unauthenticated: 404 → **401** | deleting the `GET` prefix row `permissions.go:84`; the path routes nowhere in both regimes | accepted |
| 3 | `GET /healthz` unauthenticated: **200 → 401** | authn precedes the mux (`chain.go:25`), so deleting route **and** row still changes the answer — the request is rejected at `auth/.../middleware.go:78` before reaching the 404. §7 repoints all four live consumers to `/api/v1/health/live` in the same step. | accepted, asserts 401 not 404 |
| 4 | `/api/v1/iam/presence/stream` on a SQLDB-less boot: 404 → **501** | §4's Mount-is-total rule | accepted, dev-path only |
| 5 | `HEAD /api/v1/auth/me` for a must-change-password principal: 403 → **200** (no body) | HEAD inherits the GET pattern's generated `allowedDuringPasswordChange` | accepted, §2's delta table |
| 6 | **HEAD on every capability-guarded GET route**: previously resolved by the same `(method, path)` rules; now inherits its GET pattern's capability | mux-level HEAD/GET aliasing (`$GOROOT/src/net/http/server.go:2484-2486`) — this is the *general* form of delta 5, which an earlier draft had only in its `auth/me` instance | accepted; integration test HEADs a guarded route without the capability, expects 403 |
| 7 | **Encoded-separator requests** — e.g. `GET /api/v1/documents/foo%2Fapproval-preview`: `document.submit` → `document.view` | the old resolver classified decoded `r.URL.Path`; the router routes `EscapedPath()`. Where they disagree the old verdict was **never the one the router acted on**. Whole class, not one instance. | accepted as a **correctness fix**, not a regression; table test over `%2F`, trailing slash, `//`, `..` |
| 8 | **Off-surface prefix matches** — a path no pattern serves for which a `pathPrefix` row still returns a verdict today | tier-1's verdict is unobservable behind the mux's 404/405, **except** where an old `Public` row admitted an unauthenticated caller — which is deltas 2 and 3, already listed | accepted; no residual class |

`newPublicPathChecker` derives public from `visibility == VisibilityPublic`
(`permissions.go:351-356`), so deltas 2, 3 and 8 come from one source, not two, and rows 1–8 are
the complete unauthenticated-exposure diff.

**Eight deltas, not five.** The count has been wrong three times (one, then four, then five), each
time because deltas were being *noticed* rather than *derived*. Under this gate they are derived,
and the count is an output. Per `wiki/quality/test-discipline.md`, DB-touching cases use the
`testdb` factory; the derivation proof and the `assertSurface` tests are pure and need neither.

---

## 11. Sequencing

Ruling A: no intermediate state ships. Commits may be incremental; the release is atomic.

1. `x-authz-capability` / `x-authz-capability-none` on all 147 operations, **plus full
   regeneration** — `go generate ./...` and `pnpm run gen:api`, because any spec edit churns
   every module's embedded `swaggerSpec` and partial regeneration is forbidden drift
   (`.github/workflows/api-contract.yml:29-53`; locked constraint 7). Nothing consumes the new
   extensions yet.
2. `cmd/gen-http-surface` + its validation tests, **and widen the CI drift pathspec** to cover the
   emitted file. `.github/workflows/api-contract.yml:38` diffs `'**/api.gen.go'` only, so a
   generated file with any other name is regenerated by line 37 and then never checked. The
   generated policy table must not be the one generated artifact outside the drift gate. Emits
   the file; nothing reads it yet.
3. Migrate the six legacy families to codegen (§8), each with full regeneration — **and, in the
   same step, delete `/healthz` and repoint its four live consumers**, because migrating health
   to its generated GET routes is what removes the extra bare mount at
   `observability/health.go:20`. Leaving `/healthz` mounted would fail step 6's check 2; deleting
   it later would break the four consumers in the interim. Also delete the dead
   `playwright.approval.config.ts` here rather than repointing it (§7).
4. Make the e2e scaffolding a publisher with its own generated declaration (§5); delete the dead
   panic probe. **This lands before the assertion, not after it** — an earlier draft numbered it
   last while the graph already required it earlier, and a numbered order that contradicts its own
   dependency graph is a defect, not a presentation choice. `mountE2EHandlersIfEnabled`'s direct
   mount (`main.go:839-848`) runs *after* `buildRouter` today, so a recorder installed in step 6
   would not see those routes at all: following the old numbering would have left the e2e surface
   silently outside the assertion for three commits while the assertion reported completeness.
5. Fold `presence` into the IAM publisher and delete `conditionalRouteFamilies` (§4). Mounting
   becomes total before anything asserts totality.
6. **One commit:** `httprouter.SurfacePublisher` + `[]SurfacePublisher` replacing
   `routeHandlers`, *and* the recorder + `assertSurface` at boot, *and* deletion of
   `TestRouteCoverage`. These cannot be separate commits — the deletion removes the types
   `routeHandlerFields` and `TestRouteCoverage` consume (`permissions_test.go:369-481`), so
   splitting them leaves a window where neither the old completeness guard nor the new
   assertion is in force.
7. The **derivation proof** (§10.2) and the delta list it emits (§10.3). Both tables now exist;
   this is the commit that walks all 120 `routeRules` rows against the generated table, proves
   every row is either derived by it or a listed delta, and proves no generated pattern is
   ungoverned. Locked constraint 5 forbids deleting `routeRules` before this is green — and the
   delta list is this step's output, not a footnote written beside it.
8. Flip `PermissionResolver` to the generated table lookup; delete §7's list, including all 15
   orphaned `RegisterRoutes` methods and the whole of `pdf_webhook_handler.go` (§7.1). The
   derivation proof is deleted in this same commit — it exists to license the deletion, and once
   `routeRules` is gone it has nothing to derive from. The §10.3 delta regression tests **stay**;
   they guard the new behavior, not the migration.

9. **Governance close-out**, and it is a step, not a follow-up: the two ADRs (§12), the three
   architecture-wiki updates, and the `BaseURL` promotion to one exported constant. A program that
   deletes three enumerations and leaves the documents describing them stale has reproduced its own
   defect class in the wiki.

**Dependency graph, not a topic list:**

```
1 ─→ 2 ──┬──────→ 4 ──┐
         │            │
         └────────────┼─→ 6 ─→ 7 ─→ 8 ─→ 9
3 ────────────────────┤              ↗
5 ────────────────────┘        0 ───┘
```

- `0 → 9` — **`BaseURL` promotion has no predecessor and can start on day one.** It is drawn as
  step 0 because it is a §12 deliverable (`audit/.../handler.go:89` + 9 files hard-code
  `"/api/v1"`), it is itself a hand-synced constant of exactly the class this program deletes, and
  nothing in steps 1–8 depends on it. Landing it early shrinks the surface every later step edits.
- `8 → 9` — the ADRs record what was decided **and executed**; the wiki updates cite line anchors
  that do not stabilize until the deletions land. Writing either before step 8 guarantees a second
  editing pass.
- `1 → 2` — the generator reads the extensions step 1 authors.
- `2 → 4` — step 4 runs the **same** generator over `internal-e2e.yaml` to emit
  `httpsurface_e2e_gen.go`. The generator does not exist until step 2, so step 4 is a successor,
  not a sibling. Consequence: steps 3 and 5 are parallelizable with each other and with 4, but 4
  is **not** parallelizable with 2.
- `2 → 6` — `assertSurface` reads the `httpSurface` step 2 emits.
- `3 → 6` — a bare pattern records as `"/api/v1/auth/login"` while the generated key is
  `"POST /api/v1/auth/login"`, so mounting an unmigrated family under the assertion fails checks
  2 *and* 3 and the server does not boot.
- `4 → 6` — under `METALDOCS_E2E`, an e2e route mounted with no declaration fails check 2; worse,
  mounted *outside* the recorder it fails nothing and is silently ungoverned. The publisher must
  exist first.
- `5 → 6` — check 3 (declared ⊆ mounted) is false on the SQLDB-less boot path while any mount is
  conditional. Totality of mounting is a precondition of asserting totality.
- `6 → 7 → 8` — the derivation proof needs both tables to exist; deletion needs it green.

Steps 3, 4 and 5 are mutually independent and parallelizable **with each other**; 4 additionally
requires 2 to have landed. Step 0 is parallelizable with everything. Steps 6→8 are the only span
where the old and new tier-1 coexist, and step 7 is the gate that licenses step 8. **The program is
not complete until step 9**; treating step 8 as the terminus is what leaves the governance
deliverables stranded.

**What is broken between commits, and for how long.** Between 3 and 6 the six migrated families
are guarded by the old resolver, unchanged (`permissions.go:332-338`) — the migration changes
mounting, not policy. Between 6 and 8 both tier-1 mechanisms exist and the old one is still
authoritative; the new one is asserted for *completeness* but decides nothing. No commit in the
sequence leaves a live route unrouted or unguarded.

---

## 12. ADR and risks

**ADRs required — exactly two.** An earlier draft opened "one, not two" and then named a second
one two sentences later; that was a self-contradiction, not a subtlety. The unambiguous
deliverable:

| # | ADR | Scope | Why it is separate |
|---|---|---|---|
| 1 | **Amendment to ADR 0022** (two-tier PDP) | Tier-1 route→capability resolution changes mechanism: from a hand-typed `routeRules` table to a spec-generated per-operation table keyed by the mux pattern. Tier-2 (`authz.Require` in-tx) and the DB tripwire are untouched. | ADR 0022 owns the PDP. An amendment, not a supersession — the two-tier shape and its capability vocabulary are unchanged; only tier-1's data source moves. |
| 2 | **New ADR — the HTTP surface protocol as a platform framework** | `httprouter.SurfacePublisher`, the generated `httpSurface` table, the boot assertion's four checks, and the Mount-is-total rule (§4) as a binding contract every module's delivery seam implements. | Required by the system-impact analysis (`…-system-impact.md:150`). It governs 15 modules' delivery seams, not the PDP — different owner, different reviewers, different lifetime. |

**No third ADR for the middleware chain.** §6 dropped the `route_resolve` link, so the chain
(`chain.go:25`) is byte-identical afterwards and locked constraint 6 holds without a decision
record.

**Risks**

| Risk | Handling |
|---|---|
| Generator key and oapi-codegen key diverge | §5 check 2+3 makes divergence a boot failure, not a 403; §10's derivation proof compares the two tables row by row |
| 147 `x-authz-*` declarations transcribed by hand | mechanical from `routeRules`, which covers everything: an exhaustive first-match run of all 147 spec operations against the 120 rows reaches the fallback **zero** times. There is no set of "routes with no current rule" to discover — an earlier draft claimed ~28 and that number was fabricated. The authoring risk is transcription error, and §10's derivation proof is what catches it |
| `BaseURL: "/api/v1"` is hard-coded in 10 module files (`audit/.../handler.go:89` + 9) — itself a hand-synced constant | promote to one exported constant the generator and every module read. Scheduled as **§11 step 0** — no predecessor, startable on day one |
| A second spec document (`internal-e2e.yaml`) is a new artifact that can rot | it is generated by the same generator and asserted by the same four checks whenever `METALDOCS_E2E` is on; the e2e CI job is where that boot path runs. §11 step 2 also widens the CI drift pathspec past `'**/api.gen.go'` so the generated table is covered by the same regenerate-and-diff gate |
| **Eight client-visible behavior changes**, not one — and the count itself was the tell | §10.3 carries the full list with a reason and a regression test each. The count has now been wrong four times (one, four, five, eight), and the reason is structural: under a *differential* gate, deltas were **noticed** one review round at a time, so the count could only ever be a lower bound. Under §10's derivation proof they are **derived**, and completeness of the list is the gate's output rather than a claim standing beside it. That is the single strongest argument for the operator's §2 ruling, and it is why this row now points at §10.3 instead of restating a number |
| `/api/v1/iam/presence/stream` on a SQLDB-less boot: 404 → 501 | accepted. §4's Mount-is-total rule replaces a routing-level absence with a handler-level `501 Not Implemented`, which is the honest answer and the repo's existing majority convention (8 IAM operations already do this). No production boot has `SQLDB == nil`; the delta is observable only in the degraded dev path |
| `HEAD /api/v1/auth/me` for a must-change-password principal: 403 → 200 (no body) | accepted, §2's delta table. Go's `ServeMux` matches HEAD against a `GET` pattern, so the generated per-operation `allowedDuringPasswordChange` boolean is inherited by HEAD where today's exact-`GET` clause rejects it. HEAD carries no body, so no data reaches a principal who must change their password. `logout` and `changePassword` are POST-gated and structurally immune |
| CORS preflight reaching the resolver | §9, verified during implementation |
| Six legacy families are the bulk of the work | independent of steps 1–2, parallelizable with steps 4 and 5 |
| The two ADRs and three wiki updates could be treated as follow-ups and never land | scheduled as **§11 step 9**, a step with a dependency edge (`8 → 9`), not a closing paragraph. A program that deletes three hand-synced enumerations and leaves the documents describing them stale has reproduced its own defect class in the wiki |

**Open** — none. Both AS-2 questions from the system-impact analysis are closed by operator
rulings; the `/healthz` exception is deleted rather than declared; and the §2 escalation on §10's
gate structure was put to the operator on 2026-08-05 and answered **outcome (a), restructure** —
derivation proof plus enumerated delta list, recorded in the review ledger.
