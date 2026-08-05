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

`Capability` is already a string type (`iam/domain/model.go:73`) whose values are registry
strings (`CapMetricsView Capability = "metrics.view"`, `model.go:129`), so the spec carries the
wire string and the generator resolves it to the Go constant. A typo is a build failure, not a
fail-closed 403 discovered in production.

`x-authz-password-change-allowed` replaces `isPasswordChangeAllowedPath`
(`auth/.../middleware.go:131`).

**Why the spec and not Go.** Operator ruling A. The structural reason: the spec is the only
enumeration that *cannot* drift from the wire, because it already generates the DTOs and the
`ServerInterface` that every handler must implement to compile. Hanging policy on it hangs
policy on the one artifact the compiler already forces to be correct.

---

## 3. The generator

`cmd/gen-http-surface` — a Go program, run by `go generate`, gated in CI by a
regenerate-and-diff check (the pattern `make openapi-verify` already uses).

**Input:** `api/openapi/v1/openapi.yaml`, plus the IAM capability registry for validation.

**Output:** exactly one file, `apps/api/cmd/metaldocs-api/httpsurface_gen.go`:

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
Should a future operation need two, the generator must reject it — one operation, one owner.

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
today — 137 generated `ServerInterface` methods, 9 hand-registered legacy patterns, and
`streamPresence`, which `cfg.yaml` excludes from server codegen because it is a WebSocket
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
| featureflags | 1 | `platform/featureflags/handler.go:26` |
| observability (`/api/v1/metrics`) | 1 | `router.go:123-125` — bare `mux.Handle` in the composition root |

**12 bare patterns, 12 spec operations across those six tags** — auth 4, security 3, health 2,
search 1, configuration 1, observability 1, matching §4's tag inventory exactly. (An earlier draft
said 15; that was arithmetic, not measurement.) `presence` is *not* on this list:
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
| Generator validation rules (§2, four failures) | table test per rule, each asserting non-zero exit and the operationId in the message |
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

**One path per pattern is not a parity gate — it is a sample, and it would license a deletion it
has not proven.** `routeRule.matches` (`permissions.go:35-55`) inspects the path through
`pathExact`, `HasPrefix`, `HasSuffix`, `Contains`, and a disqualifying `notSuffix` — 80 rows carry
a prefix and 49 carry a suffix/contains/notSuffix. So which row wins is **value-dependent**, and
substituting one sentinel per `{param}` explores exactly one point of that space. The concrete
miss: `PATCH /api/v1/iam/users/{user_id}` resolves to `user.manage` for almost any id, but row
`permissions.go:120` carries `notSuffix: "/roles"`, so a real request to
`/api/v1/iam/users/roles` is disqualified from that row, matches nothing else, and falls through
to session-required today. The new lookup gives it `user.manage`. That is a genuine tier-1 delta
— safe in direction, since it tightens — and a one-sample gate reports parity and never mentions
it.

**The candidate set is derived from the rule table, which makes it exhaustive rather than
lucky.** `matches` can only distinguish two paths via a literal that some rule carries. So for
each pattern, the parity test instantiates every `{param}` position with, in turn:

1. a sentinel that appears in no rule literal (`__p__`), and
2. every distinct literal token appearing in any rule's `pathExact` tail, `pathSuffix`,
   `contains`, or `notSuffix`.

Any two paths that agree on all of those agree on every rule in the table, so the set covers the
matcher's entire decision space for that pattern. The test asserts (1)'s no-collision property
against `routeRules` rather than trusting it, and it enumerates the candidate tokens from
`routeRules` itself — hand-listing them would be a hand-synced enumeration inside the test that
exists to kill hand-synced enumerations.

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
2. `cmd/gen-http-surface` + its validation tests. Emits the file; nothing reads it.
3. Migrate the six legacy families to codegen (§8), each with full regeneration — **and, in the
   same step, delete `/healthz` and repoint its four live consumers**, because migrating health
   to its generated GET routes is what removes the extra bare mount at
   `observability/health.go:20`. Leaving `/healthz` mounted would fail step 4's check 2; deleting
   it later would break the four consumers in the interim. Also delete the dead
   `playwright.approval.config.ts` here rather than repointing it (§7).
4. **One commit:** `httprouter.SurfacePublisher` + `[]SurfacePublisher` replacing
   `routeHandlers`, *and* the recorder + `assertSurface` at boot, *and* deletion of
   `TestRouteCoverage`. These cannot be separate commits — step 4's deletion removes the types
   `routeHandlerFields` and `TestRouteCoverage` consume (`permissions_test.go:369-481`), so
   splitting them leaves a window where neither the old completeness guard nor the new
   assertion is in force.
5. The exhaustive tier-1 parity test (§10). Both tables now exist; this is the commit that
   proves they decide identically, with every delta named. Locked constraint 5 forbids deleting
   `routeRules` before this is green.
6. Flip `PermissionResolver` to the pattern lookup; delete §7's list. The parity test is deleted
   in this same commit — it exists to license the deletion, and once one side is gone it has
   nothing to compare.
7. Make the e2e scaffolding a publisher with its own generated declaration (§5); delete the dead
   panic probe.

**Dependency graph, not a topic list:**

```
1 ──┐
    ├─→ 2 ──┐
3 ──┘       ├─→ 4 ─→ 5 ─→ 6
7 ──────────┘
```

- `1 → 2` — the generator reads the extensions step 1 authors.
- `3 → 4` — a bare pattern records as `"/api/v1/auth/login"` while the generated key is
  `"POST /api/v1/auth/login"`, so mounting an unmigrated family under step 4's assertion fails
  checks 2 *and* 3 and the server does not boot.
- `2 → 4` — `assertSurface` reads the `httpSurface` step 2 emits.
- `7 → 4` — under `METALDOCS_E2E`, an e2e route mounted with no declaration fails check 2, so the
  e2e publisher and its generated table must exist before the assertion goes live. (Step 7 is
  last only in the writing order; in the graph it is a prerequisite of 4.)
- `4 → 5 → 6` — parity needs both tables to exist; deletion needs parity green.

Steps 4→6 are the only span where the old and new tier-1 coexist, and step 5 is the gate that
licenses step 6.

---

## 12. ADR and risks

**ADR required — one, not two.** Tier-1 route→capability resolution changes mechanism, and
ADR 0022 governs the two-tier PDP. The ADR amends it rather than superseding it: tier-2
(`authz.Require` in-tx) and the DB tripwire are untouched, and this design changes only how
tier-1 learns which capability a route needs. A second ADR for the middleware chain is **no
longer needed** — §6 dropped the `route_resolve` link, so the chain is untouched and locked
constraint 6 holds. The analysis also names a second ADR for the protocol-as-platform-framework
(`…-system-impact.md:150`); that one stands.

**Risks**

| Risk | Handling |
|---|---|
| Generator key and oapi-codegen key diverge | §5 check 2+3 makes divergence a boot failure, not a 403; §10 adds an exhaustive generator-level comparison |
| 147 `x-authz-*` declarations transcribed by hand | mechanical from `routeRules`, which covers everything: an exhaustive first-match run of all 147 spec operations against the 120 rows reaches the fallback **zero** times. There is no set of "routes with no current rule" to discover — an earlier draft claimed ~28 and that number was fabricated. The authoring risk is transcription error, and §10's parity test is what catches it |
| `BaseURL: "/api/v1"` is hard-coded in 10 module files (`audit/.../handler.go:89` + 9) — itself a hand-synced constant | promote to one exported constant the generator and every module read |
| A second spec document (`internal-e2e.yaml`) is a new artifact that can rot | it is generated by the same generator and asserted by the same four checks whenever `METALDOCS_E2E` is on; the e2e CI job is where that boot path runs |
| Health's non-GET responses change from 200 to 405 | accepted and listed in §8, tested in §10 — the only client-visible change in the program |
| CORS preflight reaching the resolver | §9, verified during implementation |
| Six legacy families are the bulk of the work | independent of steps 1–2, parallelizable |

**Open** — none. Both AS-2 questions from the system-impact analysis are closed by operator
rulings, and the `/healthz` exception is deleted rather than declared.
