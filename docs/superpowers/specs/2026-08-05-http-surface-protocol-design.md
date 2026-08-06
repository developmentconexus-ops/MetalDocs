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

**Its cost is small.** `routeRules` has three `VisibilitySessionRequired` rows
(`permissions.go:90-92`: `GET /auth/me`, `POST /auth/change-password`, `POST /auth/logout`), so the
tier that has to be stated explicitly is by far the smallest, and "a mandatory reason string will
decay into boilerplate" does not apply at n=3. A *missing* capability on a guarded operation is
overwhelmingly likelier to be an oversight than an intent, which is the argument for stating rather
than deriving.

**This paragraph previously carried a full rule-count distribution and the claim "nothing falls
through". Both are struck.** The claim is false — `PATCH /api/v1/iam/users/roles` reaches the
fallback (§10.2) — and it was the third surviving copy of a fabrication the rewrite removed
elsewhere, which is exactly how a hand-synced number outlives its own correction. No count replaces
it: under this design completeness is enforced, not counted.

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
   be. Rules 1–5 are per-document; rule 6 is the only cross-document one, and it is what forces
   the generator to take **both** documents in a single invocation (see Input below). A rule that
   compares two documents cannot be checked by a tool that has only ever seen one.

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

**Input:** **both** spec documents in one invocation — `api/openapi/v1/openapi.yaml` and (from
§11 step 4 onward) `api/openapi/internal-e2e.yaml` — plus the IAM capability registry for
validation. An earlier draft ran the generator once per document, which made validation rule 6
unenforceable: no single run could see a cross-document collision, so the only place it could
surface was `mergedSurface`'s boot panic, which is the outcome rule 6 exists to prevent. Before
step 4 the second path is simply absent and the run is single-document; rule 6 is then vacuous
rather than skipped.

**Key formula, per document.** A key is `"<METHOD> <serverBase><path>"`, and `serverBase` is read
**from the document**, not hard-coded. `openapi.yaml` declares `/api/v1`, giving
`"GET /api/v1/metrics"`; `internal-e2e.yaml` declares an **empty** base, giving
`"POST /internal/test/seed"` — its routes are mounted at the server root, not under `/api/v1`.
A generator with `/api/v1` baked in would emit unmountable keys for every e2e route and fail §5
check 3 at boot on the `useE2E` path. Validation rule: a document whose declared base does not
prefix every one of its own paths is a build failure.

**Output:** one file per input document — `apps/api/cmd/metaldocs-api/httpsurface_gen.go` for the
public spec and `httpsurface_e2e_gen.go` for the internal one — emitted by that single run, with
distinct top-level symbols (`httpSurface`/`specTags` and `httpSurfaceE2E`/`specTagsE2E`) so the
two files coexist in one package. Same validation rules, same emitted shape. Both are covered by
the widened drift pathspec (§11 step 2).

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
surface, expectedTags := httpSurface, specTags
if useE2E {
	publishers = append(publishers, e2e)
	surface = mergedSurface(httpSurface, httpSurfaceE2E)
	expectedTags = union(specTags, specTagsE2E)
}
// Both the surface and the expected tag set come from the GENERATED tables.
// Never derive expectedTags from `publishers` — see §5 check 1.
assertSurface(mounted, surface, expectedTags, publishers)
```

`useE2E` is the *only* thing either side reads, so the four cells of the (build tag × env flag)
matrix are all total: production with the flag set is `e2e == nil` ⇒ nothing declared, nothing
mounted, the variable inert. The descriptors stay untagged — `httpSurfaceE2E` is simply an unused
package-level map in builds where `e2ePublisher()` returns nil, which Go permits — so the generator
emits one shape and the drift gate covers both files.

Three details that are not incidental:

- **The e2e publisher's `Mount` is total too — §4's rule is not waived for it.** Today
  `RegisterE2EHandlers` mounts conditionally twice: it returns early on `!E2EEnabled()`
  (`internal/test/e2e_seed.go:104-106`) and it registers its fifth route only when the scheduler
  callback is non-nil (`:113-115`). Both must go. The env check is already `useE2E`'s job and
  re-reading it inside `Mount` re-creates the two-predicate split this fix exists to close; the
  fifth route mounts **unconditionally** and answers `501` from the handler when
  `runSchedulerTick == nil`, exactly as `presence` does. A publisher whose declared surface and
  mounted surface differ by one route fails check 3 at boot — and a publisher that is allowed to
  mount conditionally is a second exception mechanism, which locked ruling C forbids.
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
// surface AND expectedTags are BOTH derived from useE2E (§3) — never from the
// publisher list, which is the thing being audited, and never httpSurface bare.
expectedTags := specTags
if useE2E {
	expectedTags = union(specTags, specTagsE2E) // generated, not observed
}
if err := assertSurface(mounted, surface, expectedTags, publishers); err != nil {
	log.Fatalf("http surface: %v", err)   // fail closed at boot
}
```

`assertSurface` checks four things and refuses to start on any failure:

1. **Tag coverage** — every entry in `expectedTags` is claimed by exactly one publisher's `Tag()`,
   and no publisher claims a tag outside it. Catches a publisher constructed but not listed, and
   two publishers claiming one tag. **`expectedTags` must come from the generated tables**, never
   from `publishers` itself: an expected set derived from the list under audit makes the check
   vacuously true, which is how a missing publisher would pass a check written to catch exactly
   that.
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
| `routeRules` — a second copy of a decision that now has one home (§10.1) | `permissions.go:82-330` | generated `httpSurface` |
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
| Generator validation rules (§2, **six** failures) | table test per rule, each asserting non-zero exit and the offending operationId in the message. Rule 6 is cross-document, so its case feeds the generator both documents in one run with a deliberate method+path collision |
| `/api/v1/iam/presence/stream` on a SQLDB-less boot answers **501**, not 404 | boot the router with `presence == nil` and assert status 501 with a `problem+json` body — the Mount-is-total rule (§4) is only real if the degraded path is exercised |
| Generated output is current | CI regenerate-and-diff; drift fails the build |
| **Every operation enforces its declared capability** (§10.2 property 2) | generated from the spec, one integration case per operation, run through the real chain over a **sentinel** terminal handler: with the capability → sentinel invoked; without it → denied **and sentinel never invoked**; `security: []` → sentinel invoked unauthenticated. Positive conformance against the single authority, not a differential against a peer table |
| **Every operation has a policy at all** (§10.2 property 1) | not a test and not a count — §2 rule 2 fails the **build** on a missing `x-authz-*`. A property the build cannot violate needs no assertion |
| **The seven behavior changes** (§10.3) | one regression test each; rows 1–6 derived from the design, row 7 asserts the new table cannot be steered by parameter content |
| Generator key == emitted pattern, exhaustive | compare the generator's key for all 147 operations against the patterns oapi-codegen actually emits, rather than reasoning from three sampled paths |
| Ownership (§5 check 4) | a publisher mounting a pattern whose declared tag is another publisher's fails `assertSurface` |
| `VisibilityUnresolved` → 500, unknown value → 500 | two middleware tests, one per switch arm |
| Health non-GET now 405 | table test asserting the accepted delta from §8, so it cannot regress silently |
| Boot assertion fires on each of its **four** checks | four unit tests over `assertSurface` with hand-built inputs — the assertion is pure, so no boot needed. One per check: (1) a tag with no publisher / two publishers, (2) a mounted pattern with no declaration, (3) a declaration with no mount, (4) a mounted pattern whose tag ≠ its mounter's `Tag()` |
| HEAD inherits GET's capability | integration test: HEAD a capability-guarded route without the capability, expect 403 |
| Public routes stay public | integration test over every `visibility: public` operation, no session, expect non-401 |
| `%2F`, trailing slash, `//`, `..` | table test against the real mux — this is §10.3 delta 7's regression test, and it asserts the **new** behavior is correct, not that it matches the old one |

### 10.1 The deletion license — there is no differential, because there is no second authority

**This section replaces seven rounds of work, and the reason it does is worth stating before the
design.** Rounds 2 through 7 all landed here. Each found a real defect. Each fix was applied. The
count never fell and the altitude never dropped, and `adversarial-review` §8 says that means the
structure is wrong, not that the reviewers were thorough.

**The root cause is not in any of the seven findings. It is in the question they were answering.**

The pre-design analysis said, at `…-system-impact.md:140`: *"it may only be dropped after parity is
proven — assert byte-equal resolution against the old one for every (method, path) the mux
registers."* That sentence was authored before a single design decision existed, and it is **not**
one of the four operator rulings. Every round after it measured the design against it. The
reviewers were never wrong; they answered the question they were asked, and the question was wrong.

**`routeRules` is not an oracle.** It is one *attempt to express* which capability each route
requires. The spec annotations are a second attempt. Neither is the authority — the authority is a
product and security **decision**, and a decision has a home, not a proof. Comparing two
non-authoritative artifacts yields differences forever and truth never. That is the exact shape of
rounds 2–7, and it explains why every fix made the comparison more sophisticated without making it
more meaningful:

| Round | What it found | What the fix did |
|---|---|---|
| 2 | `notSuffix: "/roles"` disqualifies a path a pattern lookup grants | added a path dimension |
| 3 | HEAD aliases onto GET patterns | added a method dimension |
| 4 | the method dimension was sampled, not enumerated | enumerated it |
| 5 | `%2F` makes the two sides read *different strings* | replaced the differential with a derivation proof |
| 6 | a param value can spell a routing token | made the classifier three-valued |
| 7 | first-match ordering; the labels are not a lattice | *(not applied — the ratchet fired)* |

Six escalating constructions for one comparison. The seventh would have been a regular-language
engine with negation and ordered difference — 400–700 lines of unverified code, built to serve as a
correctness oracle, and deleted at step 8.

**And it is this program's own defect class.** MetalDocs had three hand-synced enumerations of route
truth, none authoritative; this design exists to collapse them to one. The acceptance gate I wrote
was a **fourth artifact comparing two of them**. A gate that reproduces the defect it gates is not a
gate.

**The correction, stated positively.** Operator ruling A already put the authority in one place: the
capability is declared in the OpenAPI spec. Once there is exactly one home for the decision,
`routeRules` is not a peer awaiting a parity proof — it is a **second copy of a decision that now
has one home**, and the house rule disposes of it directly: *tudo fallback legacy é extermínio*.
Deleting it does not require its own permission. **Asking `routeRules` to license the deletion of
`routeRules` is the circularity that generated every round above.**

What `routeRules` retains is one honest role: it is the best surviving **record** of decisions
people actually made. So it is the *input to authoring* the 147 annotations, read row by row by a
human — a **review**, which is what a decision deserves, not a gate, which is what a proof is.

**Amendment to the analysis.** `…-system-impact.md:140` is amended by this section, and the
amendment is recorded there. This changes a constraint the analysis called locked; it was authored
in this program and corrected in this program, and the four operator rulings are untouched — indeed
ruling A is what makes the correction available.

### 10.2 What is actually proven — positive conformance against the one authority

Three properties, none of which mentions `routeRules`.

**1. Completeness is structural, not measured — in two halves.** The generator fails the build if
any **declared operation** lacks `x-authz-capability` or `x-authz-capability-none` (§2 rule 2).
That is the *declared → policy* half, and it is a condition the build cannot violate. It says
nothing about a route that is **mounted and in no document**, because the generator never sees one:
that is the *mounted → declared* half, and it is property 3's check 2, enforced at **boot**, not at
build. Stating this as one property would have been an overclaim — the two halves have different
enforcement points and different failure times, and "every route has a policy" is true only as
their conjunction. Both halves are closed today: `/api/v1/metrics` (`router.go:124`) enters the
public spec in §11 step 3, the `/internal/test/*` routes enter `internal-e2e.yaml` in step 4, and
`/internal/test/panic` (`main.go:845`) is deleted there. Every coverage count
this document asserted was fabricated, twice: 47 `pathPrefix` rows when there are 80, and "the
fallback is reached zero times", which is false — `PATCH /api/v1/iam/users/roles` routes to the
`{user_id}` pattern while `notSuffix: "/roles"` at `permissions.go:120` disqualifies its only
explicit PATCH rule, reaching `permissions.go:336`. **A property enforced by the build needs no
count. A count in a document is the defect class this program deletes.**

**2. Live conformance, one case per operation.** For each of the 147 operations, an integration test
asserts the declared capability is *required*. This is positive evidence against the authority — it
proves the table the server actually enforces is the table the spec declares. A differential against
a peer table could never prove this, because a peer table is not what the server should obey.

**A bare 403 is not evidence of tier-1.** MetalDocs runs a two-tier PDP and **both tiers deny with
403 by deliberate, test-pinned design** (`controlleddocuments/.../routes_contract_test.go:466-471`:
*"so both PDP tiers map to the same client-visible code"*). A suite asserting only the status would
go green on a route whose tier-1 rule is **missing or wrong** whenever tier-2 happens to deny — a
false green on the exact property being proven.

**The problem code is not evidence of tier-1 either, and an earlier revision of this section wrongly
said it was.** That revision separated the tiers by code — tier-1 writes `permission.denied`
(`problem/codes.go:120`, from `iam/delivery/http/middleware.go:143`), tier-2 writes
`permission.capability_denied` (`codes.go:116`, from `authz.ErrCapDenied`) — and asserted the
negative case on `permission.denied` as something that *"can only have come from the middleware"*.
**The code disproves that in both directions.** `iam/delivery/http/routes_memberships.go:312-318`
maps a **tier-2** `authz.ErrCapDenied` to `CodePermissionDenied`, the tier-1 code, with an
ADR-0022-citing comment. `documents/delivery/http/handler.go:1300-1325` emits the same tier-1 code
from inside the handler for `domain.ErrForbidden`/`ErrDocumentNotOwner`, and deliberately routes
*both* tiers' capability errors to `capability_denied`. The collapse is already catalogued
independently of this program (`docs/superpowers/analysis/2026-08-04-problem-code-registry-mapping.md`
rows 8, 33, 77, 116 and consolidation **C-3**), so it is neither new nor this program's to close.

The root cause of that mistake is more general than the mistake and is the reason this section now
reads differently: **the suite had chosen a discriminator it does not own.** The problem code on a
denied request is decided by fifteen modules' error mappers, each free to change it for its own
reasons. A proof of *this table's* correctness cannot rest on it, and a fix that made the suite
depend on it would have made every future error-mapper edit a silent hole in the proof.

**The discriminator the suite owns is the terminal handler.** Tier-1 denies *in the middleware*,
before `next.ServeHTTP` (`iam/delivery/http/middleware.go:99-143`). So the suite mounts the real
chain (`chain.go:25`) over a **sentinel** terminal handler that records invocation and returns 200,
registered under the generated pattern for each of the 147 operations — the same key both `Mount`
and the capability table derive from (§3):

- *Negative* — the request is denied **and the sentinel was never invoked** ⇒ the denial happened in
  the middleware. Unambiguous, and independent of every module's error mapper, of `problem` code
  vocabulary, and of any handler.
- *Positive* — **the sentinel was invoked** ⇒ the middleware admitted the request, which is the
  whole of the tier-1 claim.

This is strictly stronger than the code assertion it replaces, and it **retires the weak positive
case** an earlier draft accepted as unavoidable. With a sentinel there is no 404 (cross-tenant), 400
(validation), 405 or 501 to reason around and no valid body to construct for 147 operations, because
no real handler runs. The suite proves exactly *this route, this principal, admitted or refused by
the middleware* — and nothing about tier-2 or the handler, which is correct, because neither is what
this table decides.

The suite depends on one property it does not itself establish: that the pattern it registers is the
pattern the real publisher mounts. That is property 3's check, at boot, and the dependency is named
here rather than assumed.

**Mechanization, stated rather than assumed** (`tests/integration/testdb/factory.go:286-330`).
`testdb` seeds principals by **role**, and capabilities come from `role_capabilities` — there is no
exact-one-capability builder, and curated roles carry overlapping sets. The suite therefore does not
need one:
- *Negative* — `NewUser` with **no** `WithRole` yields a principal holding no capability at all: the
  `iam_user_roles` insert is guarded by `if s.Role != ""` (`factory.go:318`), so a role-less user has
  no row there and therefore no `role_capabilities` grant. One fixture serves every guarded
  operation, and it is available today with no change to `testdb`.
- *Positive* — the generator emits, per operation, any role whose `role_capabilities` set contains
  the declared capability, resolved by query at suite-build time. A capability granted by **no**
  seeded role is itself a finding, and the suite fails rather than skipping.

Neither is a new fixture framework; both are constraints on the generator, which is where the rest
of this design already puts its enumerations.

The suite is generated from the spec, so it cannot drift from the operation set, and it replaces
`TestRouteCoverage`'s hand-maintained fixture list (`permissions_test.go:369-481`) — the defect that
motivated this whole program.

**3. The boot assertion, unchanged (§5).** Declared ⊆ mounted ⊆ declared, one publisher per tag,
mounter owns its tag. That is what makes the *mounted* surface and the *declared* surface the same
object, and it is a runtime invariant rather than a migration artifact — it survives step 8, which
nothing in §10.1's discarded machinery would have.

**What is deliberately not proven:** that the new table decides identically to the old one. If a
`routeRules` row encoded a capability decision the annotation review transcribed wrongly, **the
row-by-row human review in §11 step 7 is the sole mitigation.** Property 2 cannot catch it and it
must not be credited with doing so: the suite derives its *expected* capability from the same
annotation the implementation derives its *enforced* capability from, so a transcription error
copied consistently into both is invisible to it. Property 2 proves the server obeys the spec; it
cannot prove the spec says the right thing. An earlier draft of this paragraph listed property 2
alongside the review, which was an overclaim of exactly the kind §10.1 exists to stop. This is a real and stated cost of
the §2 correction — it trades an equivalence proof for an authority. The trade is worth taking
because the equivalence proof was against an artifact that decides on **decoded substrings of the
whole path**, and can therefore be steered by user-supplied parameter content.

### 10.3 Behavior changes — derived from the design, not discovered by a differ

Seven rows, and their provenance matters: rows 1–6 were each derived by reading the design against
the middleware chain and the mux. **None came from the differential.** The differential produced
exactly two findings of its own, and they are the same finding — the old table could be steered by
param content — which row 7 states once.

| # | Change | Why | Test |
|---|---|---|---|
| 1 | Health authenticated non-GET: 200 → **405** | the method-qualified pattern replaces in-handler `r.Method` dispatch (§8) | table test |
| 2 | `GET /api/v1/health/<unknown>` unauthenticated: 404 → **401** | deleting the `GET` prefix row `permissions.go:84`; routes nowhere either way | asserts 401 |
| 3 | `GET /healthz` unauthenticated: **200 → 401** | authn precedes the mux (`chain.go:25`), so deleting route *and* row still changes the answer — rejected at `auth/.../middleware.go:78` before the 404. §7 repoints all four live consumers to `/api/v1/health/live` in the same step | asserts 401, not 404 |
| 4 | `/api/v1/iam/presence/stream` on a SQLDB-less boot: 404 → **501** | §4's Mount-is-total rule | dev-path only |
| 5 | **HEAD on every capability-guarded GET route** inherits its GET pattern's capability | mux-level HEAD/GET aliasing (`$GOROOT/src/net/http/server.go:2484-2486`). Includes `HEAD /api/v1/auth/me` for a must-change-password principal: 403 → 200, no body | HEAD a guarded route without the capability, expect 403 |
| 6 | Public routes are exactly the `security: []` operations | `newPublicPathChecker` derives public from `visibility == VisibilityPublic` (`permissions.go:351-356`), so rows 2, 3 and 6 have one source | integration test over every public operation, no session, expect non-401 |
| 7 | **The old table could be steered by path-parameter content; the new one cannot.** `user_id="roles"` reached the fallback; `pid="approval-preview"` reached `CapDocumentSubmit` instead of `CapDocumentView`; `foo%2Fapproval-preview` reached a suffix rule the router never routed | the old resolver matched decoded substrings of the whole path; the new table is keyed by the **route**. Structural, not enumerable | one test per named instance, asserting the **new** behavior is correct — plus `%2F`, trailing slash, `//`, `..` against the real mux |

**Row 7 is one row on purpose.** Under the differential its membership was an open set that grew
every round, and revision 6 tried to close it with a regenerated golden file — a list produced by
the same classifier it was meant to hold accountable. The new table **cannot** be steered by param
content, because a parameter value is never part of its key. That is a property, and a property does
not have a membership list.

Per `wiki/quality/test-discipline.md`, the conformance suite is `//go:build integration` + `testdb`;
the `assertSurface` tests are pure and need neither.

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
4. **Create `api/openapi/internal-e2e.yaml`** — it does not exist today; `api/openapi/` contains
   only `v1/` — make the e2e scaffolding a publisher with its own generated declaration (§5), and
   delete the dead panic probe. **This lands before the assertion, not after it** — an earlier draft numbered it
   last while the graph already required it earlier, and a numbered order that contradicts its own
   dependency graph is a defect, not a presentation choice. `mountE2EHandlersIfEnabled`'s direct
   mount (`main.go:839-848`) runs *after* `buildRouter` today, so a recorder installed in step 6
   would not see those routes at all: following the old numbering would have left the e2e surface
   silently outside the assertion for three commits while the assertion reported completeness.

   *What the new spec contains.* Exactly the routes `RegisterE2EHandlers` mounts
   (`internal/test/e2e_seed.go:100-116`), and nothing else. The same generator runs over it in the
   same invocation (§2 rule 6) and emits `httpsurface_e2e_gen.go`.

   *Visibility — all e2e operations are declared `security: []`, public.* This records the truth
   rather than inventing an authorization story. `POST /internal/test/seed` is the **bootstrap**:
   its *response* is what issues the session cookies (`e2e/utils/seed.ts:27` → `SeedResult.cookies`;
   `e2e/fixtures/isolation.ts:50`), so a caller cannot present a session it has not yet been issued.
   `reset` and `governance-events` are likewise driven from Playwright's unauthenticated `request`
   fixture (`e2e/utils/seed.ts:46`, `e2e/flows/happy_path.spec.ts:146`) as well as from an
   authenticated `page.request` (`e2e/flows/sod_violation.spec.ts:87`) — called both ways, so
   session-required is factually wrong for them. The guard on this surface is **not** tier-1 and
   never was: it is `E2EEnabled()`, checked at registration (`e2e_seed.go:104-106`) and re-checked
   inside every handler (`e2e_seed.go:120-123`). Declaring these session-required would add a
   second, weaker guard and break teardown paths that run after the session is gone. What the
   protocol *does* add is completeness — §5 check 2 makes an undeclared e2e mount a boot failure,
   which is more than this surface has today.

   *Mount is total here too* (§4), and the publisher has **three** conditionals of two different
   kinds. An earlier draft of this step counted two and missed the first, which is the same
   partial-sweep error §10.1 records against the analysis amendment — a guard read past because the
   two beside it were the ones being looked for.
   - `db == nil` (`e2e_seed.go:101-103`) — a **mount** conditional, and the most consequential of
     the three, because it is the one that fires on the SQLDB-less boot path where §5 check 3 is
     evaluated. Deleted. The routes mount and answer **501** when the DB is absent.
   - `runSchedulerTick != nil` (`e2e_seed.go:113-115`) — a **mount** conditional. Deleted the same
     way: mount unconditionally, answer **501** when the tick is unavailable, exactly as the
     presence stream does.
   - `!E2EEnabled()` (`e2e_seed.go:104-106`) — **not** a mount conditional under this protocol. It
     is a *composition-root* condition and it moves there: the publisher is either in the list or
     it is not, and when it is, it mounts everything it declares. `E2EEnabled()` survives as the
     per-handler check (`e2e_seed.go:120-123`), which is the real guard.

   The `mux == nil` half of the first guard is not a conditional at all — under §4 the publisher
   receives the muxer as a parameter and a nil muxer is a wiring bug that must panic, not a branch
   that silently mounts nothing.

   *Two extermination items surface here, and step 4 owns both.* They are mirror images, and
   neither is discretionary — an enumeration this program is about to **generate** must not be
   generated over a route nobody calls or a call nobody serves.
   - `POST /internal/test/advance-clock` is mounted (`e2e_seed.go:112`) and has **zero** callers
     anywhere in the Playwright suite. Delete the route and `h.advanceClock`.
   - `POST /internal/test/seed-doc` is **called** (`e2e/flows/quorum_m_of_n.spec.ts:61`) and has
     **no handler**, with the failure swallowed by
     `.catch(() => { /* endpoint may not exist; author submits later */ })`. That is a mask so a
     test passes, which is exactly what the extermination directive forbids — and the comment
     itself states the test works without it. Delete the call; do not add the endpoint.
5. Fold `presence` into the IAM publisher and delete `conditionalRouteFamilies` (§4). Mounting
   becomes total before anything asserts totality.
6. **One commit:** `httprouter.SurfacePublisher` + `[]SurfacePublisher` replacing
   `routeHandlers`, *and* the recorder + `assertSurface` at boot, *and* deletion of
   `TestRouteCoverage`. These cannot be separate commits — the deletion removes the types
   `routeHandlerFields` and `TestRouteCoverage` consume (`permissions_test.go:369-481`), so
   splitting them leaves a window where neither the old completeness guard nor the new
   assertion is in force.
7. **Flip `PermissionResolver` to the generated table lookup, and land the conformance suite**
   (§10.2 property 2) plus the §10.3 regression tests — **one commit**. The flip belongs here, not
   in step 8, and the per-edge audit below is what forced the correction: a conformance suite run
   against a table that is not yet enforcing exercises the *old* resolver and proves nothing about
   the new one, so a step 8 that flipped afterwards would be flipping onto an unverified table.
   The suite is one generated integration case per operation, asserting the declared capability is
   required, driven through the real chain over a **sentinel** terminal handler so that "the
   middleware decided" is observed directly rather than inferred from a status or a problem code
   (§10.2). It is positive evidence against the spec — the single authority — and it is **not** a
   comparison with `routeRules`; §10.1 explains at length why the program spent seven review rounds
   discovering that a comparison could not be the license.
   The 147 `x-authz-*` annotations authored in step 1 are also **reviewed row by row against
   `routeRules`** here, while it still exists. That review is a human reading of the surviving
   record of past decisions, and it is the accepted mitigation for the one residual risk in §12.
8. **Deletion.** Delete §7's list — `routeRules` included, along with all 15 orphaned
   `RegisterRoutes` methods and the whole of `pdf_webhook_handler.go` (§7.1). By this point
   `routeRules` decides nothing at runtime, but it is **not yet unreferenced**, and this step owns
   the difference. Six test sites still reach it directly and each needs an explicit disposition in
   this commit, not a compile error discovered during it:
   `permissions_test.go:527` (rule-shape walk), `:567` and `:575-591`
   (`NoMethodlessWriteShadowing`, both loops), `:713` (capability-registry membership), and
   `permissions_authz_scope_test.go:115` (`resolveRoutePermission` over base-relative paths).
   Their disposition is **delete**, not convert — every one of them guards a property of
   `routeRules`' hand-written *shape* (methodless rows, prefix shadowing, untyped capability
   strings) and none of those properties can exist in a generated table: rows are emitted
   method-qualified, keys are exact, and capabilities are validated at generation (§2 rule 3).
   A converted version would be a guard against a state the generator makes unrepresentable, which
   §3 of the review doctrine deletes on sight. It is deleted because it is a **second copy of a
   decision that now has one home**, not because a proof permitted it. Everything from step 7
   **stays**: the conformance suite guards the new table permanently, and the §10.3 tests guard the
   seven approved behavior changes. Nothing built in this program is thrown away at the end — which
   is itself a check on whether the gate was the right one.

9. **Governance close-out**, and it is a step, not a follow-up: the two ADRs (§12), the three
   architecture-wiki updates, and the `BaseURL` promotion to one exported constant. A program that
   deletes three enumerations and leaves the documents describing them stale has reproduced its own
   defect class in the wiki.

**Every step ends green, and the evidence is named per step, not once at the end.** Locked
constraint `…-system-impact.md:141-142` makes `go build ./...`, `go vet -tags integration ./...`
and the declared test set mandatory per commit — the integration-tag vet specifically, because
several of these steps change a seam signature and untagged `go test` does not compile
`//go:build integration` files. Per step, in addition to build + vet + the touched packages'
tests: steps 1 and 3 add a full `oapi-codegen` regeneration with a clean `git diff`; step 2 adds
the six generator-validation table tests and a green widened drift gate; **step 3 lands the §10.3 regression tests for rows 1, 2 and 3** — health non-GET 405,
`/api/v1/health//<unknown>` 401, `/healthz` 401 — because migrating health
(`observability/health.go:17-20` mounts all three bare) is what makes those three changes live,
three commits before step 7; **step 5 lands row 4's** (presence-stream 501) for the same reason;
**step 6** adds the four `assertSurface` unit tests plus a real boot on both the ordinary and the
`useE2E` path — all of it, because `assertSurface` does not exist until step 6 and an earlier draft
demanded its tests from steps 4 and 5, which could not have compiled; steps 4 and 5 carry build,
vet and a real boot on both paths *without* the assertion, which is what their edge rows describe;
step 7 adds the per-operation conformance suite green over all 147 operations against the **new**
resolver, rows 5–7's regression tests, and the written record of the row-by-row annotation review;
step 8 re-runs both suites with `routeRules` deleted — identical results, since neither suite ever
referenced it, and that identity is the evidence the deletion was inert; step 9 adds the ADR-status
CI gate over the two new ADRs.

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
- `6 → 7 → 8` — step 7 flips the resolver and proves the result (§10.2), which needs the generated
  table mounted and asserted complete; step 8's deletion needs step 7 green.

Steps 3, 4 and 5 are mutually independent and parallelizable **with each other**; 4 additionally
requires 2 to have landed. Step 0 is parallelizable with everything. Steps 6→7 are the only span
where the old and new tier-1 coexist. **The program is not complete until step 9**; treating step 8
as the terminus is what leaves the governance deliverables stranded.

**What is broken on each edge, and for how long.** One row per dependency edge, not a summary of
two spans — the checklist question is *“what is broken between these two commits”*, and it has to
be answered edge by edge or the spans nobody looked at are the ones that carry the exposure.

| Edge | State while the head has landed and the tail has not | Exposure |
|---|---|---|
| `0 → 9` | Nothing. `BaseURL` promotion is a constant extraction; every call site compiles against the new constant in the same commit. | None — which is why it is drawn as step 0 and may sit open indefinitely. |
| `1 → 2` | Spec extensions exist and no code reads them; the same-commit full regeneration keeps `swaggerSpec` and the generated servers consistent. | Inert. A reader sees annotations with no consumer. |
| `2 → 4` | `httpSurface` is emitted and unread; the e2e surface is still the old direct mount. | Inert at runtime. |
| `2 → 6` | Same span: the generated table exists and decides nothing. **This is where a generator mistake is invisible** — and it is exactly why step 2's six validation table tests are mandatory *in step 2*, not deferred to step 6. | Latent, and the mitigation is in the earlier commit by design. |
| `3 → 6` | **Three live client-visible changes land here** — §10.3 rows 1, 2 and 3. `health.go:17-20` mounts `/api/v1/health/live`, `/api/v1/health/ready` and `/healthz` bare; migrating them to method-qualified patterns is what produces the non-GET 405, the `/api/v1/health/<unknown>` 401 and the `/healthz` 401. Tier-1 policy itself is unchanged — the old resolver never read the mux's patterns. | Real, and rows 1–3's regression tests therefore land **in step 3**. An earlier draft of this row said "None", which was true only of the policy half and silently omitted the routing half. |
| `4 → 6` | Under `METALDOCS_E2E` only: the e2e publisher exists, `assertSurface` does not, so a declared-but-unmounted e2e route (or the reverse) fails nothing. | Zero on any non-e2e boot; bounded by whoever sets the env var. |
| `5 → 6` | **A live client-visible behavior change lands here without the assertion that motivated it**: `/api/v1/iam/presence/stream` answers 501 instead of 404 on the SQLDB-less path, in force from step 5. | Real, and it is why its §10.3 regression test lands **in step 5** rather than with the others in step 7. This row is the reason the edge table exists. |
| `6 → 7` | Both tier-1 mechanisms exist and the old one decides. The new table is **not** wholly inert: from step 6 the boot assertion reads its **keys**, so a wrong method, path or tag is a hard boot failure. What is latent is only the **capability value**, which nothing reads until step 7. | Split, and the split is the point: key errors fail loudly and immediately; capability errors are latent until step 7's suite. |
| `7 → 8` | The new resolver is authoritative and proven. `routeRules` decides nothing **at runtime**, but it is still **referenced by six test sites** (`permissions_test.go:527`, `:567`, `:575-591`, `:713`; `permissions_authz_scope_test.go:115`), so the package still compiles against it and those tests still pass against a table the server no longer obeys. | Contained but non-zero: green tests asserting properties of a dead object. Step 8 deletes them explicitly rather than discovering them as compile errors. An earlier draft of this row claimed the old table was "unreachable, not merely unused", which was false. |
| `8 → 9` | The deletions have landed and the wiki + the two ADRs describe a topology that no longer exists. | Documentation drift — bounded, owned, and the reason step 9 is a **step** rather than a follow-up. |

No commit in the sequence leaves a live route unrouted or unguarded. **Four rows changed the design
rather than merely describing it**, which is the argument for the table existing at all:
`7 → 8` moved the resolver flip out of step 8 and into step 7; `5 → 6` moved one regression test
out of step 7 and into step 5, and `3 → 6` moved three more into step 3; and `7 → 8` again forced
step 8 to own six surviving test references instead of meeting them as compile errors. Three of the
four were found only by answering the question **per edge** — the two-span summary this table
replaced reported "None" for the edge carrying the most behavior change in the program.

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
| Generator key and oapi-codegen key diverge | §5 check 2+3 makes divergence a **boot failure**, not a 403 — a runtime invariant that outlives the migration |
| 147 `x-authz-*` declarations transcribed by hand | **This is the one real residual risk, and it is accepted deliberately.** `routeRules` is the surviving *record* of the decisions, so it is read row by row as the authoring input — a human review, which is what a decision deserves. A transcription slip is caught by that review and by §10.2 property 2's live conformance case, not by a proof. Two counts asserted here earlier were fabricated (“~28 routes with no rule”, then “the fallback is reached zero times” — both false); no count replaces them, because completeness is enforced by the build |
| `BaseURL: "/api/v1"` is hard-coded in 10 module files (`audit/.../handler.go:89` + 9) — itself a hand-synced constant | promote to one exported constant the generator and every module read. Scheduled as **§11 step 0** — no predecessor, startable on day one |
| A second spec document (`internal-e2e.yaml`) is a new artifact that can rot | it is generated by the same generator and asserted by the same four checks whenever `METALDOCS_E2E` is on; the e2e CI job is where that boot path runs. §11 step 2 also widens the CI drift pathspec past `'**/api.gen.go'` so the generated table is covered by the same regenerate-and-diff gate |
| **Client-visible behavior changes** — §10.3, seven rows | The count was wrong five times (one, four, five, eight, nine) and the reason was structural: a *differential* gate can only ever **notice** deltas one review round at a time, so its count is a lower bound by construction, and each round's fix made the differ more sophisticated rather than the list more complete. §10.3's rows 1–6 are instead **derived from the design** — from the middleware chain and the mux — which is where they always came from; row 7 is a structural property of the new key, not an enumeration. This row deliberately no longer carries a number as its claim |
| `/api/v1/iam/presence/stream` on a SQLDB-less boot: 404 → 501 | accepted. §4's Mount-is-total rule replaces a routing-level absence with a handler-level `501 Not Implemented`, which is the honest answer and the repo's existing majority convention (8 IAM operations already do this). No production boot has `SQLDB == nil`; the delta is observable only in the degraded dev path |
| `HEAD /api/v1/auth/me` for a must-change-password principal: 403 → 200 (no body) | accepted, §2's delta table. Go's `ServeMux` matches HEAD against a `GET` pattern, so the generated per-operation `allowedDuringPasswordChange` boolean is inherited by HEAD where today's exact-`GET` clause rejects it. HEAD carries no body, so no data reaches a principal who must change their password. `logout` and `changePassword` are POST-gated and structurally immune |
| CORS preflight reaching the resolver | §9, verified during implementation |
| Six legacy families are the bulk of the work | independent of steps 1–2, parallelizable with steps 4 and 5 |
| The two ADRs and three wiki updates could be treated as follow-ups and never land | scheduled as **§11 step 9**, a step with a dependency edge (`8 → 9`), not a closing paragraph. A program that deletes three hand-synced enumerations and leaves the documents describing them stale has reproduced its own defect class in the wiki |

**Open** — none. Both AS-2 questions from the system-impact analysis are closed by operator
rulings; the `/healthz` exception is deleted rather than declared; and the §2 escalation on §10's
gate structure was put to the operator on 2026-08-05 and answered **outcome (a), restructure** —
positive conformance against the single authority ruling A establishes, recorded in the review ledger.
