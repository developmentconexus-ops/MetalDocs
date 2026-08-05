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
for the public spec, and `httpsurface_e2e_gen.go` (build-tagged with the e2e gate) for the second.
Same generator, same validation rules, same emitted shape; only the input and the output filename
differ. Both are covered by the widened drift pathspec (§11 step 2).

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

- **`iamdelivery.Router` mounts unconditionally and answers 501 when a dependency is nil** — eight
  operations do this today (`sessions`, `observability`, `tenants`, `presence`'s own *snapshot*,
  and the deprecated `createManagedUser`: `iam/delivery/http/router.go:128-134,207-223,230-236,247-249,298-324`).
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

A tag may have more than one publisher (`iam` has two: the generated router and the hand-mounted
stream). Check 1 asserts every tag has **at least one** publisher; check 4 asserts every mounted
pattern's tag matches its mounter. Neither requires a bijection.

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
if err := assertSurface(mounted, httpSurface, specTags, publishers); err != nil {
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
| `pdf_webhook_handler.go` + its test | `documents/delivery/http/pdf_webhook_handler.go` | — dead: off-spec, unwired (see below) |
| **13 orphaned `RegisterRoutes` methods** left behind by past codegen migrations | table below | the generated `HandlerWithOptions` mount that already supersedes each |

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

An independent sweep of every delivery seam found **13** exported `RegisterRoutes` methods that no
production code path reaches. They are not a scattering of small oversights; they are one defect
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
| `documents/module.go:157` (`RegisterRoutes`, no rate limit) | test only (`module_wrapper_test.go:138`) | production calls `RegisterRoutesWithRateLimit` exclusively (`router.go:113`) |
| `iam/.../admin_handler.go:132` | test only | `Router.RegisterGenerated` (`iam/.../router.go:114`) |
| `iam/.../people_handler.go:54` | test only | same |
| `iam/.../routes_memberships.go:102` | test only | same |
| `iam/.../routes_roles_caps.go:51` | test only | same |
| `iam/.../sessions_handler.go:96` | test only | same |
| `documents/.../pdf_webhook_handler.go:41` | test only | **nothing** — see below |

All 13 are deleted. The tests that call them move to invoking the business method directly,
which is what the majority of the sibling tests already do (`view_handler_test.go`,
`rbac_test.go`, `reconstruct_handler_test.go` never used `RegisterRoutes` in the first place).

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
same defect class as the three known-false comments already on the list (`router.go:72` names
`buildPresence`, which does not exist — the real name is `startPresence`, `main.go:1168`;
`permissions_test.go:415-417` says five bare-pattern families where there are six;
`router.go:78-80` claims a loud boot failure that does not happen); all four are corrected or
deleted with the code they describe.

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
| **Tier-1 parity, exhaustive over the matcher's value-dependence** (see below) | for every recorded pattern × every path in its derived candidate set: call the *old* `resolveRoutePermission(method, path)` and the *new* lookup, assert equal `(Capability, Visibility)`. Deltas are listed in the test as data with a reason each, not tolerated by a loose assertion. Locked constraint 5 / `…-system-impact.md:140`. |
| Generator key == emitted pattern, exhaustive | compare the generator's key for all 147 operations against the patterns oapi-codegen actually emits, rather than reasoning from three sampled paths |
| Ownership (§5 check 4) | a publisher mounting a pattern whose declared tag is another publisher's fails `assertSurface` |
| `VisibilityUnresolved` → 500, unknown value → 500 | two middleware tests, one per switch arm |
| Health non-GET now 405 | table test asserting the accepted delta from §8, so it cannot regress silently |
| Boot assertion fires on each of its three checks | three unit tests over `assertSurface` with hand-built inputs — the assertion is pure, so no boot needed |
| HEAD inherits GET's capability | integration test: HEAD a capability-guarded route without the capability, expect 403 |
| Public routes stay public | integration test over every `visibility: public` operation, no session, expect non-401 |
| `%2F`, trailing slash, `//` | table test against the real mux |

### 10.1 The parity gate — root cause, and why this is a restructure and not a third patch

This gate has been the review's BLOCKER for three consecutive rounds. Two patches have already
been applied to it (one sentinel per `{param}`; then a candidate set of literals derived from
`routeRules`). `adversarial-review` §1 forbids a third patch to the same construct without a
written local-vs-global verdict, so here it is.

**Root cause — not "the sample was too small".** The gate compared *two tables*. But the new
tier-1 is not a table; it is **table + mux**. `mux.Handler(r)` is half the new function, and it
was outside the comparison. Everything round 3 named is a consequence of that single omission:

- HEAD resolves to a `GET` pattern (`$GOROOT/src/net/http/pattern.go:258`, `server.go:2485`:
  a pattern with method `GET` matches both GET and HEAD) — a mux behavior, invisible to a
  table-vs-table test.
- A method mismatch returns pattern `""` *and* the 405 handler
  (`$GOROOT/src/net/http/server.go`, `findHandler`'s `matchingMethods` branch) — a mux behavior.
- Methodless rows in `routeRules` are only reachable by varying the method — and the old side
  took `(method, path)` while the new side was being fed a pattern string.

Round 2 found the *path* dimension of the same defect (`permissions.go:120` carries
`notSuffix: "/roles"`, so `PATCH /api/v1/iam/users/roles` is disqualified from the `user.manage`
row and falls through to session-required today, while a pattern lookup gives it `user.manage`).
Round 3 found the *method* dimension. There is no reason to expect round 4 to be the last one,
because the search space of a table-vs-table differential test is infinite and the test is a
sample.

**What makes the finding impossible (§1 step 2).** Define the comparison at the level the new
tier-1 actually operates on: **an `*http.Request` against the real mux**. Not `(method, path)`,
not a pattern string.

```go
// old side — unchanged, the live function today
oldCap, oldVis := newPermissionResolver()(r.Method, r.URL.Path)
// new side — the WHOLE new tier-1, mux included
newCap, newVis := newPermissionResolver(mux)(r)
```

Every dimension of an HTTP request the mux is sensitive to is then inside the test **by
construction**, including dimensions neither review has thought of yet. That is the difference
between a gate that survives a round and a gate that cannot generate this finding class again.

**§2 verdict — outcome (a), restructure now.**

| Question | Answer |
|---|---|
| Global-maximum structure | A request-level differential over a mechanically enumerated, *finite* domain — not a sample over an infinite one. |
| What a proven system does | The standard shape for replacing a routing/dispatch table: prove equivalence over the **reachable** input set and state the reachability argument, rather than attempting equivalence over all strings. |
| Cost of the global maximum | One paragraph (the off-surface argument, §10.3) plus one test. |
| Cost of the local maximum later | A BLOCKER every round, and a deletion licensed by a gate that has never been total. |
| Chosen by | Author, under `adversarial-review` §2 outcome (a). The restriction to the observable surface is an interpretation of locked constraint 5's word *exhaustive* and is surfaced explicitly in §10.3 for the operator, not resolved silently. |

### 10.2 The domain, enumerated mechanically from both tables

For every pattern `p` the recorder captured (`"<METHOD> <path>"`), the test builds requests from:

**Paths** — instantiate every `{param}` position in `p` with, in turn:

1. a sentinel appearing in no rule literal (`__p__`), asserted no-collision against `routeRules`
   rather than assumed; and
2. every distinct literal token appearing in any rule's `pathExact` tail, `pathSuffix`,
   `contains`, or `notSuffix`.

`matches` can only distinguish two paths through a literal some rule carries, so two paths
agreeing on all of these agree on every row. The tokens are enumerated **from `routeRules`
itself** — hand-listing them would be a hand-synced enumeration inside the test that exists to
kill hand-synced enumerations.

**Methods** — for each such path:

1. `p`'s own method;
2. `HEAD`, whenever `p`'s method is `GET` (the mux-level aliasing above);
3. one method registered on that path by no pattern at all — the 405 case, which is where
   `routeRules`' methodless rows become observable.

The product is finite, derived from both tables, and total over the surface the server actually
serves. Deltas are listed in the test **as data with a reason each**, never absorbed by a loose
assertion.

### 10.3 Off-surface behavior is stated and tested, not proven equal

The domain above covers registered patterns. For a path no pattern serves, the old resolver still
returns a verdict (a prefix row can match a path that routes nowhere) while the new one returns
`VisibilitySessionRequired` for the empty pattern. These are **not** proven equal, and they do not
need to be: the mux emits 404/405 either way, so tier-1's verdict is unobservable — with exactly
one exception, which is where an old `Public` row would have let an *unauthenticated* caller
through:

| Old public row (`permissions.go:84-87`) | Off-surface delta |
|---|---|
| `GET` prefix `/api/v1/health/` | `GET /api/v1/health/<anything else>`: unauthenticated 404 today → **401**. Accepted; the path routes nowhere in both regimes. |
| `pathExact /healthz` | row and route both deleted in step 3 — no delta to have. |
| `POST /api/v1/auth/login` | `pathExact` + registered ⇒ on-surface, covered by §10.2. |
| `GET /api/v1/feature-flags` | `pathExact` + registered ⇒ on-surface, covered by §10.2. |

Note this is the **same** resolver the authn middleware consults for its public check —
`newPublicPathChecker` derives public from `visibility == VisibilityPublic`
(`permissions.go:351-356`), so there is one source, not two, and the table above is the complete
unauthenticated-exposure diff.

Per `wiki/quality/test-discipline.md`, DB-touching cases use the `testdb` factory; the
generator and `assertSurface` tests are pure and need neither.

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
7. The exhaustive tier-1 parity gate (§10). Both tables now exist; this is the commit that
   proves they decide identically over the whole enumerated domain, with every delta named.
   Locked constraint 5 forbids deleting `routeRules` before this is green.
8. Flip `PermissionResolver` to the request-level lookup; delete §7's list, including the 13
   orphaned `RegisterRoutes` methods and `pdf_webhook_handler.go` (§7.1). The parity gate is
   deleted in this same commit — it exists to license the deletion, and once one side is gone it
   has nothing to compare.

**Dependency graph, not a topic list:**

```
1 ─→ 2 ──┐
         ├─→ 6 ─→ 7 ─→ 8
3 ───────┤
4 ───────┤
5 ───────┘
```

- `1 → 2` — the generator reads the extensions step 1 authors.
- `2 → 6` — `assertSurface` reads the `httpSurface` step 2 emits.
- `3 → 6` — a bare pattern records as `"/api/v1/auth/login"` while the generated key is
  `"POST /api/v1/auth/login"`, so mounting an unmigrated family under the assertion fails checks
  2 *and* 3 and the server does not boot.
- `4 → 6` — under `METALDOCS_E2E`, an e2e route mounted with no declaration fails check 2; worse,
  mounted *outside* the recorder it fails nothing and is silently ungoverned. The publisher must
  exist first.
- `5 → 6` — check 3 (declared ⊆ mounted) is false on the SQLDB-less boot path while any mount is
  conditional. Totality of mounting is a precondition of asserting totality.
- `6 → 7 → 8` — parity needs both tables to exist; deletion needs parity green.

Steps 3, 4 and 5 are mutually independent and parallelizable. Steps 6→8 are the only span where
the old and new tier-1 coexist, and step 7 is the gate that licenses step 8.

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
| Generator key and oapi-codegen key diverge | §5 check 2+3 makes divergence a boot failure, not a 403; §10 adds an exhaustive request-level comparison |
| 147 `x-authz-*` declarations transcribed by hand | mechanical from `routeRules`, which covers everything: an exhaustive first-match run of all 147 spec operations against the 120 rows reaches the fallback **zero** times. There is no set of "routes with no current rule" to discover — an earlier draft claimed ~28 and that number was fabricated. The authoring risk is transcription error, and §10's parity gate is what catches it |
| `BaseURL: "/api/v1"` is hard-coded in 10 module files (`audit/.../handler.go:89` + 9) — itself a hand-synced constant | promote to one exported constant the generator and every module read |
| A second spec document (`internal-e2e.yaml`) is a new artifact that can rot | it is generated by the same generator and asserted by the same four checks whenever `METALDOCS_E2E` is on; the e2e CI job is where that boot path runs. §11 step 2 also widens the CI drift pathspec past `'**/api.gen.go'` so the generated table is covered by the same regenerate-and-diff gate |
| **Four client-visible behavior changes**, not one — the "only change" claim in an earlier draft was wrong | each accepted, each named, each with a regression test (§10, and §10.3's exception table for the off-surface pair): (a) health's authenticated non-GET 200 → 405 (§8); (b) `GET /api/v1/health/<unknown>` unauthenticated 404 → 401, a consequence of deleting the methodless `/healthz` row; (c) and (d) are the two rows below |
| `/api/v1/iam/presence/stream` on a SQLDB-less boot: 404 → 501 | accepted. §4's Mount-is-total rule replaces a routing-level absence with a handler-level `501 Not Implemented`, which is the honest answer and the repo's existing majority convention (8 IAM operations already do this). No production boot has `SQLDB == nil`; the delta is observable only in the degraded dev path |
| `HEAD /api/v1/auth/me` for a must-change-password principal: 403 → 200 (no body) | accepted, §2's delta table. Go's `ServeMux` matches HEAD against a `GET` pattern, so the generated per-operation `allowedDuringPasswordChange` boolean is inherited by HEAD where today's exact-`GET` clause rejects it. HEAD carries no body, so no data reaches a principal who must change their password. `logout` and `changePassword` are POST-gated and structurally immune |
| CORS preflight reaching the resolver | §9, verified during implementation |
| Six legacy families are the bulk of the work | independent of steps 1–2, parallelizable with steps 4 and 5 |

**Open** — none. Both AS-2 questions from the system-impact analysis are closed by operator
rulings, and the `/healthz` exception is deleted rather than declared.
