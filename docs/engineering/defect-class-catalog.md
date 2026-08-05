# Defect-Class Catalog

> **Purpose.** Not a bug list — a list of **defect *classes*** observed in a large,
> multi-module Go + TypeScript codebase, each paired with the mechanism that makes
> the class *unreachable by construction* in a future project.
> **Audience.** The software factory: whoever sets up repo scaffolding, CI, and
> review doctrine on day 0 of the next system.
> **Rule of the document.** Every class carries real evidence from this repo
> (`file:line`, ADR, commit). No hypothetical defects. If a class has no evidence
> here, it does not belong in this file yet.
> **Last verified:** 2026-08-04

---

## 0. The Prevention Ladder

The single most important idea in this document. When you decide how to prevent a
defect class, you are choosing a rung. **Always climb as high as the problem allows.**

| Rung | Mechanism | Fails when | Cost |
|---|---|---|---|
| 1 | **Impossible to express** — the type system rejects it | never | design effort up front |
| 2 | **Generated** — one source of truth, everything else emitted from it | generator not run → rung 3 catches it | build wiring |
| 3 | **CI gate** — blocking lint/check on every push | someone disables it | pipeline time |
| 4 | **Test** — a failing assertion | test written at wrong altitude (→ §8) | maintenance |
| 5 | **Doc / convention** — written down | always, eventually | ~0 |
| 6 | **Review vigilance** — a human notices | reviewer tired, new, or absent | reviewer time |

**The central finding of this repo's audit:** most of its real defects are conventions
sitting on rungs 5–6 that everyone *believed* were on rungs 1–3. A convention without
enforcement is not a standard. It is a wish with good intentions.

**Factory rule:** a convention that matters enough to write down is a convention that
must be enforced at rung ≤3. If you cannot enforce it, either it does not matter, or
your design is wrong.

---

## Class 1 — The Fake Type Guard

**Symptom.** A type exists to constrain values, its doc comment claims it does, and it
does not.

**Evidence.** `internal/platform/problem/codes.go:7`:
```go
// Using a distinct type prevents arbitrary strings from being used as codes.
type Code string
```
False. `Code` is a *defined string type*, so any untyped string constant converts
implicitly. `problem.New(409, "whatever_i_want", …)` compiles. Result: 147 distinct
error codes in 3 competing conventions, 26 of them raw literals in a single function
(`internal/modules/controlleddocuments/delivery/http/routes.go:501-577`).

**Root cause.** Go's untyped-constant conversion. `type X string` constrains *variables*
of other named string types, never literals. The author reasoned by analogy to a real
newtype and never tested the negative case.

**Prevention (rung 1).** If a type must reject literals, it needs an unexported field:
```go
type Code struct{ s string }        // literal cannot construct it
func Register(module, code string, status int) Code { … }
```
Keeps comparability, map-key usability, and `MarshalJSON` to the same wire string.

**Detection in an existing repo.** For every "this type prevents X" comment, write the
test that *asserts X is prevented*. If you cannot write a compile-failure test, the
guarantee is not real. `go vet` will not tell you.

**Generalization.** Any comment asserting a guarantee is a test that was never written.
Grep for `prevents`, `ensures`, `guarantees`, `cannot` in comments and treat each as a
missing test.

---

## Class 2 — Hand-Synced Enumerations

*This repo's own named meta-defect (final architecture review, 2026-07-03).*

**Symptom.** Two or more lists in different languages/files must agree. Nothing enforces
that they do. They diverge silently.

**Evidence.**
- Backend error codes ↔ `frontend/apps/web/src/lib/api/errorMessages.ts`. Bridged by
  `scripts/dump-error-codes.go`, which is **wired to nothing** — zero references in
  `.github/workflows/`, `Makefile`, `scripts/*.ps1`. Currently 3 codes stale, added by
  commit `1d3f8db5` (2026-08-04), snapshot last touched `f3b5dc60` (2026-07-29). Users
  hitting those conditions see a raw code instead of a message.
- OpenAPI `Problem.code` is `type: string` with description *"Machine-readable code from
  canonical taxonomy"* (`api/openapi/v1/openapi.yaml:7190`) — the spec names a taxonomy
  it does not encode. No enum. **Closed 2026-08-04:** the field now carries a `pattern`
  for the vocabulary's shape, kept in sync with the registry by a test.
- `wiki/architecture/api-design-system.md:88` documents `IDEMPOTENCY_KEY_CONFLICT`, which
  exists nowhere. Real constant is `IDEMPOTENCY_KEY_REUSED` (`codes.go:26`).
- Historic instances: capability registry size, authz tripwire arms (closed in GMR M2 by
  generating the arms from the Go registry + 2 blocking drift/parity lints — the correct
  fix, and proof the pattern generalizes).

**Status in this repo: CLOSED for error codes (ADR 0089, 2026-08-04).** The evidence
above is preserved as observed. `dump-error-codes.go` is deleted; `cmd/problem-codes-dump`
reads the runtime registry and generates both the FE snapshot and the wiki table, with a
CI job (`problem-codes-freshness`) that regenerates and byte-compares. The spec now
carries a `pattern` — deliberately not the enum this catalog originally called for, since
an enum makes adding an error code a breaking change; see the ADR's Implementation record.
The wiki table is no longer hand-written at all, which is why its 15 stale rows could
happen and now cannot.

**Root cause.** Duplication across a language boundary, where no compiler spans both sides.

**Prevention (rung 2 + 3).** One source of truth; every other representation *generated*
from it; a CI job that regenerates and fails on diff. The OpenAPI enum, the FE message
keys, and the wiki table are all **outputs**, never inputs.

**Detection.** Ask of every list: "what happens if someone adds an entry to only one
side?" If the answer is not "CI fails", it is this class.

**Factory rule.** Never let two enumerations that must agree be maintained by humans.
Generation is not an optimization here; it is the only correct implementation.

---

## Class 3 — The Allowlist Guard That Ratifies Drift

**Symptom.** A guard exists, looks reassuring, and is *opt-in* — so the worst offenders
are simply not on the list. Worse: the exclusion is documented as intentional, which
converts an unpaid debt into apparent policy.

**Evidence.** `internal/platform/problem/codes_catalog_guard_test.go`. `guardedPackages`
covers 7 packages. Not guarded: **all of `approval/http`** (67 codes), the
`controlleddocuments` 501-line switch with 26 raw literals, `taxonomy`, `tokens`,
`documents/.../fillin_handler.go`. The exclusion comment reads:

> the dotted-taxonomy packages are intentionally excluded: they own a separate,
> documented code taxonomy

The guard does not merely miss the drift. It **legitimizes** it. Every later reader sees
a sanctioned second standard.

**Status in this repo: CLOSED (ADR 0089, 2026-08-04)** — and the way it closed is the
lesson. The guard was not widened; it was **deleted**, because `problem.Code` became a
type no other package can construct, which moves the check from rung 3 to rung 1 and
leaves no list to be absent from. Two live defects were found in exactly the unguarded
packages: wrong and unregistered codes in `platform/idempotency`, and the approval
module's `dot.notation` — which the exclusion comment called a deviation and which turned
out to be the *better* convention, later adopted repo-wide. Read that twice: the allowlist
was not just hiding drift, it was blessing the wrong standard as the default.

The same trap was then re-encountered while *building* the replacement: a first draft of
the frontend vocabulary guard flagged profile codes and role codes as false positives, and
the obvious remedy was an allow-list. It was rejected for a narrower signal instead. A
check that cries wolf earns an allow-list, and an allow-list is how drift becomes
legitimate — so precision in the rule is not polish, it is what keeps the rule honest.

**Root cause.** Introducing a guard into a non-conforming codebase, and choosing
allowlist (cheap, ships today) over blocklist-with-expiry (honest).

**Prevention (rung 3).** Guards are **deny-by-default**. Migration exceptions live in an
explicit exception file where each entry carries an owner and an expiry date, and CI
fails on an expired entry. An exception is a debt with a due date, never a second
standard.

**A third failure mode, found later: the allowlist anchored by position.** During the
approval accountability loop (2026-08-05), an unrelated edit shifted lines in a file whose
`api-lint` tripwire allowlist entry was keyed to a *location* rather than to a symbol. The
entry silently stopped matching. Nothing failed, and the change passed review. An allowlist
keyed by `file:line` is a second copy of a fact (Class 2) whose divergence is *invisible by
construction*, because the guard's own success path is "entry not matched → nothing to
report". **Rule: an exception must be keyed to something the compiler or the parser can
resolve — a symbol, a fully-qualified name, an AST node — never to a line number, and never
to a substring of a line. And every allowlist needs a staleness check: an entry that matches
nothing is a failure, not a pass.**

**Detection.** Every guard with an inclusion list is suspect. Invert it and count the
failures — that number is the true debt, and it will be much larger than expected. Then
check the inverse: for each allowlist entry, assert it still matches something.

---

## Class 4 — Semantic Overload of One Field

**Symptom.** One column/field carries two meanings. Reasoning about it requires knowing
which one is active. Bad states become reachable *by construction*.

**Evidence.** ADR 0088. `template_versions.content_hash` meant simultaneously:
1. the verified hash of the object this version points at
   (`autosave.go:151-165`, CHECK `chk_template_version_content_hash_non_draft`);
2. "the user has actually edited this"
   (`lifecycle.go:72-74`, `spawnNextDraft` — *"ContentHash is left empty … so the publish
   gate still forces a real edit"*).

Meaning (2) was expressed as *absence of meaning (1)*. So "a template version that points
at no content" was not a bug — it was the encoding. A user who created a blank template
and submitted it got `409 UPLOAD_MISSING` for a file they never intended to upload.

**Root cause.** A second requirement arrived and reused an existing field instead of
adding one. Cheap at the time; the cost is that every future reader must disambiguate,
and the invalid state is *required to exist* for the encoding to work.

**Prevention (rung 1).** One field, one meaning, stated in the schema comment. A new
meaning gets a new field or a new table. Then the DB constraint can be unconditional
(`length(content_hash) = 64` always), which makes the bad state unrepresentable.

**Detection.** For each nullable column, ask what NULL *means*. If the answer has the word
"or" in it, this class is present.

---

## Class 5 — Absence as Configuration

*Sibling of Class 4, distinct enough to name separately.*

**Symptom.** A valid, intended system state is encoded as the **absence** of a record.
Absence is then indistinguishable from misconfiguration, from a failed migration, and
from a bug.

**Evidence.**
- ADR 0087. `governance_class='livre'` meant "no approval route may exist". Consequence:
  a livre profile could own *nothing* — documents and templates both hard-block creation
  without an active route. "Ungoverned material" degenerated into "profile that can own
  nothing". Fixed by making livre an explicitly configured **zero-stage route**: absence
  of a route is now *always* misconfiguration, for every class.
- Same ADR, review round 2: `GetApprovers` inferred auto-approval from "zero signoff
  rows". But simples review-only routes also approve with zero signoffs — so those PDFs
  would have falsely rendered *"Aprovação automática — rota livre"*. Fixed by gating on
  the instance's pinned stage-instance count == 0, a positive fact.

**Root cause.** Absence is free to write and needs no migration. It is also unfalsifiable:
you cannot distinguish "deliberately empty" from "never populated".

**Prevention (rung 1).** Configuration is explicit and present. Model the "nothing
required" case as a first-class configured object (a zero-stage route), not as a missing
row. Then absence has exactly one meaning: misconfiguration — and can be alerted on.

**Detection.** Every `IF NOT EXISTS` / `len(x) == 0` branch that produces *success*.
Ask: could this emptiness also arise from a bug or an incomplete migration? If yes,
the code cannot tell, and neither can the on-call engineer.

---

## Class 6 — Fallback That Fabricates Truth

**Symptom.** Integrity-critical read cannot find a value and substitutes a plausible one.
The caller cannot distinguish real from fabricated.

**Evidence.** `GetFinalApprovalDate` used `coalesce(…, now())` — a document whose approval
date was unknown rendered on a **regulated eQMS PDF** with today's date, indistinguishable
from a real approval date. Fixed (ADR 0087 work) to source
`COALESCE(MAX(signoff.signed_at), MAX(ai.completed_at))` and return `ErrNoApprovalDate`
when neither exists.

**Root cause.** `COALESCE`/`??`/`|| default` are ergonomic and make the null-pointer go
away. The failure moves from a crash (loud, immediate, cheap) to a wrong document (silent,
downstream, expensive, and in a regulated context, a compliance defect).

**Prevention (rung 1 + doctrine).** Integrity-critical reads **fail closed**. Standing rule
in this repo: *no-fallback principle* — explicit status-scoped queries over polymorphic
`COALESCE`. At the type level, return `(T, error)` or an option type, never a zero value
that reads as valid.

**Detection.** Grep `COALESCE`, `?? `, `|| default`, `.unwrap_or`. For each, ask: if this
default is wrong, does anything break loudly? If not, it is this class.

**Boundary.** Fallbacks are fine for cosmetics (a default avatar). They are never fine for
anything that lands in a record, a document, an audit trail, or a decision.

---

## Class 7 — Validation Duplicated Across Layers

**Symptom.** The same rule is enforced at several layers. Relaxing it in one leaves the
others enforcing the old rule — the feature is unshipped while appearing complete.

**Evidence.** ADR 0087 review round 1, P0-1. Domain validation was relaxed to permit
zero-stage routes; migration and service layer both correct; **but** the HTTP contract
still carried `minItems: 1` on `CreateRouteRequest.stages` plus a `validateStages` check.
Net effect: the entire ADR was unimplemented — livre profiles still could not own
anything — while service-level and SQL-level tests passed green.

**Root cause.** Defense-in-depth applied to *validation* rather than to *invariants*.
Copies drift because nothing ties them together.

**Prevention.** Distinguish the two:
- **Invariants** (DB constraints, triggers) — deliberately duplicated with app checks.
  The app check is the friendly error; the DB is the truth. Both must be updated
  together, and a test asserts the DB rejects what the app rejects.
- **Validation** (shape, ranges, cardinality) — single source. Generate the contract
  layer from the schema, or the schema from the contract. Never hand-write both.

**Detection.** For each rule, grep its concept across layers. More than one hand-written
site = drift waiting.

---

## Class 8 — Tests at the Wrong Altitude (False Green)

**Symptom.** Good coverage numbers, passing suite, broken feature. The tests cannot fail
when the thing users touch is broken.

**Evidence.**
- Class 7's P0-1: tests existed at the service and SQL layers, **none through the HTTP
  contract**. The feature was 100% broken end-to-end and 100% green.
- FE error-message coverage test compares `errorMessages.ts` ↔ the generated snapshot —
  **never against backend source**. A stale snapshot passes green forever. This is
  exactly how Class 2's 3-code drift stayed invisible.
- `//go:build integration` files are **not compiled** by an untagged `go test ./...`.
  After any seam signature change, integration tests can be uncompilable while the
  default suite is green. Repo rule: run `go vet -tags integration` before commit.

**Root cause.** Tests written where it is *convenient to write them* (nearest the code
just changed) rather than where the feature is *observable*. Guard tests written against
a cached copy instead of the source, because the copy is easier to read.

**Prevention.**
- At least one test per feature at the **outermost layer the user actually touches**
  (HTTP contract, or the UI).
- **Guard tests compare against the source of truth, never a snapshot of it.** A snapshot
  test detects unintended change; it can never detect staleness.
- CI compiles *every* build tag.

**Detection.** For each test, ask: "if I broke this feature completely at the boundary,
would this test fail?" If no, it measures implementation, not behavior.

---

## Class 9 — The Second Copy of a Critical Path

**Symptom.** An invariant-bearing sequence is copy-pasted. Later fixes land in one copy.

**Evidence.** Terminal approval (instance CAS → authz check → release recorder →
OCC update → lifecycle enqueue) was inlined in **both** `decision_service.go` and
`review_verdict_service.go`. ADR 0087 needed a third caller (auto-approve) — extracted to
`document_terminal_approval.go` / `template_terminal_approval.go` with an explicit
standing note: *never write a third copy*.

**Root cause.** The first duplication is always locally cheaper. Cost appears at the
*third* site, or at the first divergent bugfix.

**Prevention (doctrine + review).** Duplication of a path carrying an invariant is a
defect at **n = 2**, not n = 3. Pure-logic duplication can wait; invariant duplication
cannot, because the failure mode is a security or correctness hole in the copy nobody
updated.

**Second evidence — the duplication can be a *table*, not a code path.**
`approval_review_verdicts` (`db/baseline/0001_current_schema.sql:2062`) and
`approval_signoffs` (`:2151`) carry the same instance/stage/actor/tenant columns, the same
comment and display-name snapshot, the same `(stage_instance_id, actor_user_id)` uniqueness
constraint and the same `enforce_approval_sod` trigger (`:4024`). The verdict table is the
signoff table minus five signature columns. Two tables, two services, two endpoints, two
screens — and the SoD invariant is now enforced twice, so it can drift in one place. Extend
the rule: **an invariant enforced in two schemas is the same defect as an invariant enforced
in two functions.** Detection is a grep for triggers and unique constraints appearing on more
than one table with the same predicate.

**Detection.** Grep for authz calls, state-machine transitions, and audit-record calls.
Each set of near-identical sequences is a candidate.

---

## Class 10 — Cross-Module Redeclaration of a Shared Concept

**Symptom.** Modules independently declare the same concept. Values duplicate, then drift.

**Evidence.**
- `internal/modules/tokens/delivery/http/handler.go:33-36` declares
  `CodeTokenAlreadyExists = "ALREADY_EXISTS"` and `CodeTokenNotFound = "NOT_FOUND"` —
  **verbatim duplicates** of `problem.CodeAlreadyExists` / `problem.CodeNotFound`
  (`codes.go:21,23`), redeclared rather than imported.
- `PROFILE_NOT_FOUND` / `PROFILE_ARCHIVED` are typed constants in
  `taxonomy/delivery/http/routes_profiles.go:26,27` **and** raw literals in
  `controlleddocuments/delivery/http/routes.go:563,567`.
- `taxonomy` declares 8 catalog-style UPPER codes outside the catalog, shadowing its
  namespace.

**Root cause.** Importing across module boundaries feels like coupling, so authors
redeclare. But a **shared wire contract is not coupling** — it is the contract. The
coupling already exists; redeclaring only removes the compiler's ability to see it.

**Prevention (rung 1 + 3).** Shared concepts live in `platform/`, declared once. A lint
forbids redeclaring a platform value. The registry from Class 1 makes it structural:
you cannot register the same code twice.

**Detection.** Collect all constant *values* (not names) repo-wide and group by value.
Any value declared in two packages is this class.

---

## Class 11 — The Module-Private Platform Primitive

**Symptom.** A genuinely general utility lives inside one module. The second module that
needs it copies it. Often there is a comment admitting it should be promoted.

**Evidence.** `strictjson.go` — module-private, carrying a standing note to promote it to
`internal/platform`. The note has outlived several releases.

**Root cause.** Promotion is a refactor with no visible feature value, so it is always
deferred to a quieter week that does not arrive.

**Prevention (rung 3 + doctrine).** **Promotion is triggered by the second consumer, not
scheduled.** Make it a merge-blocking rule: if module B needs a utility from module A, the
PR that introduces the need is the PR that promotes it. A `TODO: promote` comment is
itself the defect — a rung-5 note standing in for a rung-3 rule.

**Detection.** Grep `TODO: promote`, `should live in platform`, `move this to`. Each is an
unpaid debt with a known solution.

---

## Class 12 — The Document as False Authority

**Symptom.** Documentation states something that is no longer (or was never) true.
It is worse than absent docs: absent docs prompt a code read; wrong docs stop the search.

**Evidence.**
- `wiki/architecture/api-design-system.md:88` documents `IDEMPOTENCY_KEY_CONFLICT`,
  which exists nowhere in the codebase.
- The same doc presents a 15-row code table and one acknowledged exception, while
  **79 non-catalog codes** exist unmentioned. A reader is left believing the taxonomy is
  small and enforced.
- `internal/platform/problem/codes.go:7` — the fake guarantee of Class 1.

**Root cause.** Docs are written once, at peak understanding, and never re-derived.
Nothing fails when they rot.

**Prevention (rung 2 + 3).**
- Generate every doc that enumerates code facts (code tables, capability lists, route
  tables) from source.
- Every hand-written doc carries a `Last verified: <date>` stamp, and CI warns past a
  threshold. *(This repo already does this — the practice is sound and worth carrying
  forward.)*
- `file:line` anchors in docs are checked by CI for existence.

**Factory rule.** Runtime truth beats docs. When code and doc disagree, the doc is a
defect ticket, not a decision.

---

## Class 13 — The Spec That Names a Constraint It Does Not Encode

**Symptom.** A machine-readable contract describes a constraint in prose while leaving
the machine-readable field unconstrained.

**Evidence.** `api/openapi/v1/openapi.yaml:7190`:
```yaml
code: { type: string, description: 'Machine-readable code from canonical taxonomy' }
```
No enum. Clients cannot generate an exhaustive switch; codegen produces `string`; a typo
is valid forever. Five code strings appear elsewhere in the spec (in descriptions and
examples) — and **two of those five are the same condition in two different styles**,
so the spec documents its own inconsistency.

**Root cause.** Prose is free; enums require maintenance. Without Class 2's generation,
an enum would go stale, so the author correctly avoided a worse defect and left the
underlying one open.

**Prevention (rung 2).** If the spec names a constraint, the spec **encodes** it — and
the encoding is generated, so maintenance cost is zero. Enum + generation solve each
other; either one alone is a trap.

**Status in this repo: CLOSED (ADR 0089, 2026-08-04), with a correction to the advice
above.** Encoding the constraint as an **enum of all 141 codes** was specified and then
rejected during implementation, because it makes **adding an error code a breaking
change** for strict clients. Errors are the part of a contract that grows most, and a
design whose safe move is "reuse a vaguer existing code" pushes straight back toward the
vocabulary collisions of Class 19. What shipped is a `pattern` encoding the vocabulary's
**shape** (`<family>.<name>` over ten closed families), which is also the part a client
can usefully branch on — `permission.*` is 403 whichever member it is.

Refined factory rule: **encode the constraint at the granularity that is stable.** Shape
is stable; membership grows. Encode shape in the contract, enforce membership at the
boundary that can be regenerated without a client migration (here, the generated frontend
snapshot). And when the copy is unavoidable — a spec cannot import code — pair it with a
test that fails on divergence, or the encoding becomes Class 2.

**Detection.** Grep spec descriptions for "one of", "must be", "from the", "canonical",
"valid values". Each is a constraint that should be a schema keyword.

---

## Class 14 — Optimizing Inside a Local Maximum

*The class that produces the other classes.*

**Symptom.** Effort is spent improving something whose *shape* is the actual problem. The
result works, ships, and locks the bad shape in deeper, because now there is more code
depending on it.

**Evidence.**
- ADR 0088 explicitly rejected *"disable the submit button while `content_hash` is null"*
  — it would have removed the user-visible symptom and preserved the
  reachable-invalid-state-by-construction defect (Class 4) permanently.
- ADR 0088 also rejected outbox materialization: it would fix the copy while leaving a
  window where a version exists without content — the exact state being eliminated.
- The error-code work: choosing a style and running a repo-wide rename **without** first
  closing the type hole (Class 1) and building the registry would relocate an unenforced
  convention, not enforce one, and would touch every call site twice.

**Root cause.** The patch is always visible, bounded, and estimable. The redesign is
none of those. Incentives point down the ladder.

**Prevention (process, rung 3).** A **foundation-judgment gate before improvement work**:
before optimizing/extending anything, state in writing whether the base is sound or is
itself a patch. If it is a patch, name the global-maximum structure and its trade-off,
and let the operator choose. This repo operationalizes it as the `developing-new-work`
skill, which emits a written system-impact analysis with a Green/Yellow/Red verdict where
Red hard-blocks design.

**Detection.** In any improvement proposal, ask: "if we had a free hand, is this the shape
we would build?" If no, you are on a local maximum, and every hour spent raises the exit
cost.

---

## Class 15 — Two Runtimes, One Primitive

**Symptom.** A primitive built for one execution context is reused in another where its
implicit assumptions do not hold. It compiles, and fails at runtime or fails silently.

**Evidence.**
- Repositories that resolve tenant from request context are **HTTP-shaped**. Background
  jobs have no request; they need explicit `SeedTxTenant(param)` + a system-bypass path +
  an explicit COMMIT. Reusing the HTTP-shaped repo in a job silently reads the wrong
  (or no) tenant.
- `authz.Require` needs a **writable** transaction because it records the authz decision —
  so it is structurally incompatible with a read-only transaction. Resolved by collapsing
  `DoReadOnly` into `Do` plus an api-lint guard.
- Calling an authz-recording read **inside a lock-holding atomic transaction** deadlocks.
  Standing constraint: keep those reads off-transaction.

**Root cause.** The assumption (a request exists / the tx is writable / no lock is held)
is implicit in the primitive's design and invisible at the call site.

**Prevention (rung 1, else 3).** Make the context a **parameter of the type**, so a
background caller cannot construct the HTTP-shaped variant. Where the type system cannot
express it, encode the rule as a blocking lint — which is what this repo did, correctly,
in `scripts/api-lint`.

**Detection.** For each platform primitive, list its implicit environmental assumptions.
Every unlisted assumption is a future incident.

---

## Class 16 — Compiles ≠ Works

**Symptom.** Wiring is present, types check, tests pass, and the feature does nothing
at runtime — a listener not registered, a consumer not subscribed, a handler not routed.

**Evidence.** GMR M2: live QA caught **non-functional drives** that compiled and passed
tests. The lesson was recorded verbatim as *compile ≠ work*. Related: the metrics
listener remediation (F-R1) in the same program.

**Root cause.** Static wiring in Go is mostly invisible — a registration that is never
called looks identical to one that is, at compile time.

**Prevention (rung 3 + 4).** A **live QA gate** per milestone: exercise the feature
against a running stack, not a test harness. Plus registration-parity tests (the registry
in Class 1 supports this directly: assert every declared handler/consumer/code is
actually reachable).

**Detection.** For each registration point, ask what test would fail if the registration
line were deleted. If none, only live QA can catch it.

---

## Class 17 — Shared Infrastructure Without Isolation Contracts

**Symptom.** Parallel work shares mutable infrastructure. Failures are intermittent,
environment-dependent, and blamed on flakiness rather than on the missing contract.

**Evidence.**
- Two parallel tracks with divergent schema fingerprints on the same test Postgres port
  mutually deleted each other's templates → `3D000` errors and timeouts.
- A test asserted goroutine lifecycle via `runtime.NumGoroutine` inside a **parallel**
  suite — inherently racy. Recorded rule: never do this.
- Test-DB garbage collection could not be run mid-suite without destroying peers; gated
  behind an explicit env flag.

**Root cause.** Shared infrastructure defaults to "works when one thing uses it".
Concurrency is added later, and the isolation contract is never written down.

**Prevention (rung 1 + 3).** Lease-based isolation: each track/worker gets a namespaced
resource it exclusively owns. Destructive operations gate behind an explicit flag and
refuse to run when peers are active. Never assert global process state in a parallel test.

**Detection.** For each shared resource, ask what two concurrent consumers do to each
other. "We do not run them concurrently" is a convention (rung 5), not an answer.

---

## Class 18 — Symptom-Patching at a Boundary

**Symptom.** A failure surfaces at layer N and is fixed at layer N, though it originates
at layer N−k. The boundary's guarantee is weakened to accommodate the bug.

**Evidence.** Standing repo rule, learned the hard way: **authz is never symptom-patched.**
ADR 0022 (capabilities, never roles) is the boundary; a permission failure is fixed by
correcting the capability model, never by adding a bypass, widening a check, or
special-casing a route.

**Root cause.** The symptom is where the pain is and where the reporter is looking. The
patch is smaller and ships today. Its cost is that the boundary's guarantee is now
conditional, and nobody who reads the boundary later will know.

**Prevention (doctrine).** Name your **inviolable boundaries** explicitly on day 0
(authz model, tenancy, the async/outbox rule, the contract-first rule). A fix that
weakens one of them is not a fix — it is an architecture change requiring an ADR. Say
this out loud in the contributor doc, because the pressure to patch is strongest exactly
when the boundary matters most.

**Detection.** Any special case, bypass flag, or `if isSpecialRoute` near a security or
tenancy check.

---

## Class 19 — Vocabulary Fragmentation Across Modules

*The class that this whole audit began from — the aggregate of Classes 1, 2, 3, 10, 13.*

**Symptom.** Each module invents its own dialect for a **shared wire contract**. The
envelope is standardized; the payload is not. Clients cannot reason uniformly.

**Evidence.** The error-code audit. Envelope: standardized and CI-enforced (RFC 9457,
`ENVELOPE-DRIFT` rule in `scripts/api-lint`, 15/15 modules). Codes: 3 conventions,
147 values, and the same condition emitted differently by different modules:

| Condition | Module A | Module B |
|---|---|---|
| no active approval route | `state.approval_route_missing` (409) | `APPROVAL_ROUTE_MISSING` (409) |
| upload missing | documents: **410** | templates: **409** |
| content-hash mismatch | documents 422, templates 409 | approval **412** |
| `ErrActorNotEligible` | approval: `signoff.not_eligible` | templates: `FORBIDDEN_CAPABILITY` |

A comment in `controlleddocuments/.../routes.go:519-523` claims it mirrors *"the SAME wire
contract the submit path already emits … so both surfaces are one contract for the
client"*. True for one module, false for the other. **The comment was accurate when
written and rotted silently** — Class 12 compounding Class 19.

**Root cause.** Standardizing the *container* is easy to specify and easy to lint.
Standardizing the *vocabulary* requires a registry, and nobody built one, so each module
solved it locally — correctly, in isolation, and incompatibly.

**Prevention (rung 1 + 2 + 3).** Registry-first, always, for any cross-module vocabulary
(error codes, event names, capability names, metric names, feature flags):
1. closed type — literals cannot compile;
2. `Register(module, value, metadata)` — one declaration site per value;
3. all downstream representations generated (spec enum, FE map, docs);
4. CI freshness gate on the generated artifacts.

**Factory rule — the strongest one in this document.** Any string that crosses a process
boundary and that *both sides must agree on* needs a registry from day 0. Retrofitting
one costs a repo-wide hard break; building one costs an afternoon.

**Status in this repo: CLOSED (ADR 0089, 2026-08-04).** The retrofit was paid: 155 codes
→ 141, 111 removed, 108 added, 26 status rulings, one hard break with no compatibility
layer. Prevention steps 1, 2 and 4 shipped as written; step 3 shipped with the Class 13
correction above (`pattern`, not enum). Two costs the plan did not predict, both worth
recording for the next retrofit:

- **The registry creates a new failure mode.** A registering package nobody imports never
  runs its init, so its codes vanish from every artifact with nothing looking wrong. Needs
  its own lint (`PROBLEM-DUMP-IMPORT`); budget for it.
- **The rename's damage was on the consumer side, not the producer side** — see Class 21.
  The backend was protected by the new type on day one. Ten frontend branches were dead
  and nothing failed. Retrofitting a registry is only half the job if consumers compare
  against the old vocabulary in a medium the type cannot reach.

---

## Class 20 — Irreversible Artifact With a Reversible-Feeling Workflow

**Symptom.** A workflow presents editing affordances for something that is, in
practice, permanently distributed. People act as if a mistake can be taken back. It
cannot, and the attempt to take it back produces false confidence.

**Evidence.** A secret reached git history (F-18). Git *feels* editable — `amend`,
`rebase`, `filter-branch`, `push --force`. Once pushed, it is on every clone, fork, CI
cache, and mirror, none of which the repo owner controls. The resolution chosen here was
a **clean re-baseline at v1** rather than a history purge, and that was the correct call:
a purge would have produced a repo that *looks* clean while the secret survives in copies.

Same class, other instances: a published package version (yankable, not deletable — it is
already in lockfiles); a sent email; an API contract already consumed by a client.

**Root cause.** The tool's affordances describe the *local* artifact. The artifact's real
lifetime is *distributed*, and no local operation reaches the copies. The gap between
those two is invisible at the moment of the mistake.

**Prevention.**
- **Rung 3 — stop it at the boundary.** Secret scanning in **both** pre-commit and CI (CI
  is the one that matters; pre-commit hooks are bypassable and not installed on every
  clone). `.env` and equivalents never tracked, enforced by `.gitignore` *and* a CI check
  that fails on a tracked match.
- **Rung 5 but load-bearing — name the write-once artifacts on day 0.** Git history,
  published versions, sent messages, released contracts. Write them down, because the
  affordance lies and only doctrine corrects it.
- **The remediation rule, which is the real lesson:** for a distributed artifact, the fix
  is to **invalidate the leaked thing** (rotate the secret, publish a fixed version,
  send a correction), **never to erase the trace**. Erasure is unverifiable across copies
  and buys a feeling of resolution instead of resolution. Treat "we purged it" as an
  unproven claim.

**Detection.** For each artifact your workflow produces, ask: after it leaves this
machine, can I reach every copy? If no, it is write-once, regardless of what the tool's
buttons say.

---

## Class 21 — The Rename That Leaves a Dead Branch, Not a Broken One

**Symptom.** A shared identifier is renamed across a system. Every *declaration* of it
moves; every *comparison* against it keeps compiling and keeps running, and simply stops
being true. Nothing fails. A branch that used to fire never fires again, silently and
permanently.

**Evidence.** ADR 0089 renamed 111 Problem codes. The backend was safe by construction —
`problem.Code` is a closed type, so a stale code does not compile. The frontend was not:
`if (error.code === 'AUTH_UNAUTHORIZED')` is a comparison between two strings, and after
the rename it is merely a comparison that is always false. **Ten such branches were dead**
when the guard was written, including session-expiry detection in `lib/api/client.ts` —
the app had stopped recognizing its own expired sessions. The existing coverage test could
not see any of it: it compared the *message map* against the code snapshot, and these
codes were in control flow, not in the map. Two further branches tested codes **no backend
had ever registered** — dead from the day they were written.

Same class, other instances: a renamed feature-flag key still checked by an
`if (flags['old_name'])`; a renamed event name in an analytics `switch`; a renamed CSS
class in a `classList.contains`; a renamed enum value compared as a string across a
process boundary.

**Root cause.** The rename is type-safe on **one side of a boundary only**. Wherever the
identifier crosses into a medium with no shared type — JSON on the wire, a string
comparison, a config file, a database value — the compiler stops helping, and the failure
mode inverts: instead of *breaking loudly*, the code *succeeds at doing nothing*. A
comparison that is always false is indistinguishable from a condition that never occurs.

**Prevention.**
- **Rung 3 — a lint that reads the other side's generated truth.** The consumer already
  receives a generated artifact listing the valid vocabulary (`error-codes.generated.json`).
  A test that scans consumer source for literals in identifier positions and checks them
  against that artifact turns the silent case into a build failure.
- **Make the signal narrow, deliberately.** The obvious rule — flag every literal assigned
  to a field named `code` — cried wolf immediately: this codebase also has profile codes,
  area codes, and role codes. **A check that cries wolf earns an allow-list, and an
  allow-list is how drift becomes legitimate (Class 3).** The rule that shipped matches
  unambiguous *positions* (`.code === '…'`, known constructors) plus any literal carrying
  one of the ten closed family prefixes — so it is precise by construction rather than by
  exception. Its exemption list is three named files, each with a written reason.
- **Rung 1 where the boundary allows it.** If the consumer is typed (TypeScript), generate
  a union type from the same artifact so the comparison itself fails to typecheck. This
  is strictly better than a test and should be preferred when the codegen path exists.

**Detection.** Grep the consumer for string literals compared against values that a
generator owns. For each: is there anything that would fail if that literal became
obsolete? If the honest answer is "the branch would just stop running", it is this class.
Note that **code coverage will not find it** — the line is still executed, it just always
evaluates false.

---

## Class 22 — The Guarded Branch That No Environment Executes

**Symptom.** A branch exists to handle the abnormal case. Every environment the change is
tested in is normal, so the branch is never entered. In a language that resolves names at
**execution** rather than at parse time, the branch can contain an outright invalid
reference — a table that does not exist, a column that was renamed, a function with the
wrong arity — and still ship green. It fails on the first machine whose *data* differs.

**Evidence.** Migration `0317` (ADR 0088) inventories content-less template versions and,
`IF v_doomed = 0`, does nothing. The `ELSE` arm asserts that no doomed row is referenced,
and it read `public.document_profiles` — that table is in the `metaldocs` schema. The other
three tables in the same assert genuinely are in `public`, which is what let one wrong
qualifier hide in plain sight. Every database it was verified against was already clean, took
the `v_doomed = 0` branch, and applied it successfully. The dev volume had content-less
drafts, entered the `ELSE`, and failed with `relation "public.document_profiles" does not
exist (SQLSTATE 42P01)` — putting `metaldocs-api` into a startup restart loop, because
migrations are applied on boot.

Same class, other instances: a PL/pgSQL error handler that references a dropped column and
only runs when the error fires; a rarely-hit `except` block calling a renamed helper; a
reflection-based or string-built SQL path behind a feature flag nobody enabled in staging;
a rollback path exercised only by the failure it exists to handle.

**Root cause.** Two things compound. First, the language does no build-time name resolution
for the branch (PL/pgSQL function bodies, dynamic SQL, reflection, interpreted config). The
compiler that would have caught this in Go is simply not present. Second — and this is the
part that generalizes past SQL — the branch's guard is **correlated with the data**, so the
set of environments that exercise it is exactly the set the author did not have. "It applied
cleanly everywhere I tried" is *evidence of the branch not running*, not evidence of it
working. This is Class 8 (false green) with a sharper edge: the test was not at the wrong
altitude, it was at the wrong **state**.

**Prevention.**
- **Rung 4 — a migration must be applied against a database in the state it exists to
  change, not only against a clean one.** For a data-repair migration, that means a fixture
  that plants the rows the migration targets. If the repair arm never ran, the migration was
  not tested; it was type-checked by absence.
- **Rung 3 — assert the reference, not the branch.** For PostgreSQL specifically, a
  migration can validate its own names up front: resolve every table it touches through
  `to_regclass('schema.table')` and `RAISE` on `NULL`, at the top of the block where it runs
  unconditionally. That converts a data-dependent runtime failure into a deterministic one
  on every database, clean or not.
- **Rung 5 — write the schema qualifier explicitly and never rely on `search_path`.** Half
  of this class in a multi-schema database is a name that was correct under the author's
  session `search_path` and wrong under the deploying role's.

**Detection.** For each conditional arm in a migration, dynamic query, or error handler, ask:
**what data would have to exist for this line to execute, and did any environment have it?**
If the answer is "no environment we tested", the branch is unverified regardless of what the
suite reports. Coverage tooling does not help here — most of it does not instrument SQL at
all, and where it does, the arm shows as uncovered rather than as wrong.

---

## Class 23 — The Invariant Delegated to Configuration

**Symptom.** A rule is ratified as a product invariant, implemented correctly in code, and
then made *optional at runtime* because its enforcement point is a per-tenant (or per-
install, per-environment) configuration value. Every code path is right. A valid
configuration still violates the invariant, and nothing anywhere reports it.

**Evidence.** The ratified approval model (2026-07-10) carries rule R3, *"signing includes
conversing"*: whoever can sign a stage can also return the document for correction without
signing. The code implements it — `review_verdict_service.go:177` explicitly permits
`request_changes` on an `approval`-kind stage and blocks only `ready` there. But the same
method hard-requires `authz.Require(CapApprovalReview)` (`:164`), and `approval.review`
(`internal/modules/iam/domain/model.go:94`) is a capability *granted per profile*. A tenant
may therefore configure a signer profile holding the signature capability without
`approval.review`. That signer's only available "no" is the terminal rejection that kills
the instance — precisely the outcome R3 exists to prevent. No test, constraint, or lint
detects the configuration.

**Root cause.** Capability systems are designed to make authority configurable, which is
correct. The error is applying that machinery to a rule that is not meant to vary: an
invariant expressed as a grant becomes an invariant the operator can switch off by
accident, and the switch looks like ordinary administration.

**Prevention (rung 1, then 3).**
1. Separate the two questions before designing: *is this something a tenant chooses, or
   something the product guarantees?* A guarantee must not be reachable through the
   configuration surface at all.
2. If the rule is an implication between two capabilities ("holding A implies B"), encode
   it as an **implication in the capability model**, resolved at check time — not as two
   independent grants that an administrator is trusted to keep consistent.
3. Where an implication is impossible, add a **configuration-validity gate**: a check that
   runs on every config write and in CI against the shipped seed data, asserting no
   reachable configuration violates a named invariant. Rung 3, but it fails loudly.

**Detection.** For every rule in a ratified model or spec, name the artifact that makes it
false. If that artifact is a row in a configuration table, the rule is a default, not an
invariant. A blunter version: read the seed/reference data as an adversary and try to build
a legal configuration that breaks each stated rule.

**Generalization.** This is Class 1's shape moved one layer out. Class 1 is a *type* that
claims a guarantee it does not enforce; this is a *product rule* that claims a guarantee its
enforcement mechanism structurally cannot provide. Both are answered the same way: write the
test that asserts the negative case, and if you cannot express it, the guarantee is not real.

---

## Class 24 — Two Operations, One Precondition, Divergent Terminality

**Symptom.** Two operations are available to the same actor, in the same state, under
identical preconditions — and their consequences differ irreversibly. Nothing in the
contract, the schema, or the UI states which one applies when. The actor picks, and the
system honours a choice it never defined.

**Evidence.** On a single active `approval`-kind stage, an eligible actor may:

| Operation | Record written | Consequence | Reversible? |
|---|---|---|---|
| `request_changes` | `approval_review_verdicts` | instance → `changes_requested`, document → `draft`, `revision_version + 1` (`review_verdict_service.go:365-412`) | yes — author corrects and resubmits |
| signoff `decision='reject'` | `approval_signoffs`, `signature_meaning='rejection'` | stage → `rejected_here`, instance → **`rejected`** (terminal), document rejected (`decision_service.go:638-694`) | no |

Same actor, same stage, same eligibility check, same SoD rule. One returns the work; the
other ends it. There is no rule anywhere saying which situation calls for which, so the
distinction is carried entirely by whichever screen the actor happened to open.

**Root cause.** The two operations arrived at different times, each locally sensible, and
nobody asked whether they occupy the same semantic slot. This is Class 14 (optimizing inside
a local maximum) observed from the user's side rather than the maintainer's: the shape was
never decided, so it accumulated.

**Prevention (rung 1).** Model the decision as **one operation with an explicit outcome
enum**, not as N endpoints. The outcome becomes a value the contract enumerates, the DB
constrains, and the UI renders as a genuine choice with stated consequences. If two outcomes
genuinely differ in terminality, that difference belongs in the enum's documentation and in
a confirmation affordance — not in which URL was called.

**Detection.** For each pair of write endpoints on the same resource, diff their
preconditions. **Identical preconditions plus divergent terminality is this class**,
regardless of how different the handlers look. Ask of the pair: could a competent user pick
the wrong one and be unable to undo it?

**Factory rule.** Irreversibility is a property of an *outcome*, never of an *endpoint*. If
choosing a route is how a user chooses to destroy something, the API has delegated a product
decision to routing.

---

## Class 25 — The Snapshot That Reads Like Source

**Symptom.** A generated point-in-time artifact — a schema dump, a lockfile, a vendored
bundle, a baseline — sits in the repository looking exactly like hand-authored source. It is
correct only as of its fold date; the delta chain that supersedes it lives elsewhere. Any
single read of it is silently stale, and the staleness is invisible because the file is
internally consistent, well-formatted, and committed.

**Evidence — from this audit, made by the auditor.** Investigating approval route shape, I
read `assert_route_shape` in `db/baseline/0001_current_schema.sql:184` and reported to the
operator that the `simples` and unknown-class arms were unconstrained — a fail-open. That
was the truth of the baseline text and **false at runtime**: migration
`db/migrations/0316_livre_zero_stage_route.sql:73` had already `CREATE OR REPLACE`d the
function so unknown classes fail *closed* onto the `controlado` rule. Runtime truth is
baseline + migrations; a baseline read alone is a read of the past. The repo's own doctrine
already says this. It did not stop the error, because nothing in the file says so at the
point of reading.

**Root cause.** Baseline-fold is a good practice — it keeps a growing migration chain
readable. Its cost is that the most authoritative-looking copy of the schema becomes the
least current one, and the correction lives in files a reader must know to go find.

**Prevention (rung 2, then 3).**
- **Generate a folded view.** If a baseline exists, generate `schema-current.sql` = baseline
  + all migrations applied, on every migration commit, and make *that* the file people read
  and grep. The baseline stops being a read target.
- **Mark the snapshot in-band.** A header on every generated snapshot: what it is, its fold
  date, what supersedes it, and how to get current truth. A reader who lands mid-file via
  grep should still hit the warning — repeat it as a section banner, not only at line 1.
- **CI (rung 3):** fail if any object defined in the baseline is also redefined by a
  migration and no folded view was regenerated. That is exactly the drift that bit here.

**Detection.** For every committed artifact, ask: *is this authored or emitted?* If emitted,
where is its input, and what proves the copy is current? Any emitted file that a human is
expected to read is either regenerated on every relevant commit or is a trap. A related
smell: two files defining the same database object, template, or symbol, where precedence is
established by an external application order rather than by the files themselves.

**Generalization.** Class 12 is a *document* that lies. This is an *artifact* that lies while
being syntactically valid code — harder to catch, because the usual defence ("read the code,
not the docs") points straight at it.

---

## Class 26 — The Local Gate That Is Weaker Than the CI Gate

**Symptom.** The same tool runs in the developer's verification routine and in CI, under
different flags. The local invocation passes; the strict one would not. Every defect in the
gap between them is found late — after review, after merge, sometimes after push — and the
person who introduced it had a green run in front of them the whole time.

**Evidence.** `scripts/api-lint` defaults to non-strict (`main.go:19`,
`flag.Bool("strict", false, …)`). CI runs it as
`go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .`
(`.github/workflows/api-contract.yml:100`), where every rule is blocking. The harness
verification ladder listed the lane as plain "api-lint". So the routine every task ran was a
strictly weaker check than the one that decides the build — which is how the de-anchored
allowlist in Class 3 above survived a review.

**Root cause.** Two invocations of one tool, maintained in two places (Class 2 again), with
the weaker one placed where feedback is fastest and the stronger one where feedback is
slowest. That is the incentive gradient pointing exactly the wrong way: cheap-and-lenient
early, expensive-and-strict late.

**Prevention (rung 2, then 3).**
- **One invocation, one definition.** The gate is a single script or make target
  (`make lint-api`) that both the local ladder and CI call. Neither side spells out flags.
  Divergence stops being possible rather than being discouraged.
- If flags must differ, **the local default is the strict one.** Leniency is the flag you
  opt into, never the one you inherit.
- **CI check (rung 3):** assert that the workflow's invocation string equals the one the
  local target uses. Cheap, and it catches the drift the day it appears.

**Detection.** Diff every CI step's command against the documented local ladder line for the
same tool. Any difference in flags, paths, or scope is this class. Ask specifically: *does
any tool have a "strict", "ci", or "all" mode that the local routine does not enable?*

**Generalization.** A gate is not what the tool can check — it is what the routine actually
runs. A capability nobody invokes is documentation. This is the process-level twin of Class
22 (the guarded branch no environment executes): there the code path never ran, here the
check never ran, and in both cases the suite reported success.

---

## Appendix A — Day-0 Factory Checklist

Derived directly from the classes above. Each line prevents a class already observed.

**Type-level (rung 1)**
- [ ] Every cross-boundary vocabulary is a closed type with an unexported field (§1, §19)
- [ ] No nullable field carries two meanings; NULL has one documented meaning (§4)
- [ ] "Nothing required" is a configured object, never a missing row (§5)
- [ ] A product invariant is never expressed as a grant an administrator can withhold;
      capability implications are resolved at check time (§23)
- [ ] A decision is one operation with an outcome enum, never N endpoints whose
      terminality differs (§24)
- [ ] Integrity-critical reads return `(T, error)`; no defaulting (§6)
- [ ] Execution context (request vs background) is a type parameter, not an assumption (§15)

**Generation (rung 2)**
- [ ] One registry per shared vocabulary: codes, events, capabilities, metrics, flags (§19)
- [ ] Spec enums, FE maps, and doc tables are **generated outputs** (§2, §13)
- [ ] Zero hand-maintained lists that must agree with another list (§2)

**CI (rung 3)**
- [ ] Generated artifacts have a freshness gate — regenerate, fail on diff (§2)
- [ ] Guards are **deny-by-default**; exceptions carry owner + expiry, CI fails on expiry (§3)
- [ ] Allowlist entries key on symbols the parser resolves, never on `file:line`; an entry
      matching nothing fails the build (§3)
- [ ] Each gate has ONE invocation shared by the local ladder and CI; if flags differ, the
      local default is the strict one (§26)
- [ ] All build tags compile in CI (§8)
- [ ] Doc `file:line` anchors verified; `Last verified` stamps warn past threshold (§12)
- [ ] Every emitted artifact a human reads is regenerated on every relevant commit; an
      object defined in a baseline and redefined by a delta fails the build unless the
      folded view was regenerated (§25)
- [ ] Shipped seed/reference configuration is read adversarially in CI: no legal
      configuration may violate a named product invariant (§23)
- [ ] Lint: no redeclaration of a platform value in a module (§10)
- [ ] Lint: encode known runtime-context hazards (§15)

**Test doctrine (rung 4)**
- [ ] Every feature has ≥1 test at the outermost user-touchable layer (§8)
- [ ] Guard tests compare against **source**, never a snapshot of source (§8)
- [ ] Shared test infrastructure is leased/namespaced; destructive ops gated (§17)
- [ ] Never assert global process state in a parallel test (§17)

**Process**
- [ ] Secret scanning in pre-commit **and** CI; tracked-`.env` check fails the build (§20)
- [ ] Write-once artifacts named on day 0; remediation = invalidate, never erase (§20)
- [ ] Foundation-judgment gate before any improvement work; Red blocks design (§14)
- [ ] Inviolable boundaries named on day 0; weakening one requires an ADR (§18)
- [ ] Promotion to platform triggered by the **second** consumer, in that PR (§11)
- [ ] Invariant-bearing path duplicated at n=2 is a defect (§9)
- [ ] Live QA gate per milestone — compile ≠ work (§16)
- [ ] A data-repair migration is applied against a database seeded into the state it
      repairs; a clean-database run does not count as tested (§22)
- [ ] Migrations resolve every table they touch up front (`to_regclass`) and qualify every
      name explicitly — never rely on `search_path` (§22)

---

## Appendix B — Recurring Root Causes

The 26 classes reduce to five underlying causes. Useful when classifying a *new* defect
that does not obviously match a class above.

1. **A guarantee was asserted, not enforced.** Comments, docs, and conventions standing in
   for compilers, generators, and gates — including a guarantee whose enforcement point is
   a configuration value an administrator may legally set the other way. → §1, §3, §12,
   §13, §23
2. **Two things that must agree, maintained separately.** No compiler spans the boundary —
   including an artifact and the delta chain that supersedes it, and one tool invoked two
   ways in two places. → §2, §7, §10, §11, §19, §21, §25, §26
3. **Absence used to carry meaning.** Unfalsifiable, indistinguishable from failure.
   → §4, §5, §6
4. **The local fix was cheaper than the right fix,** and the incentive gradient always
   points there — including leaving two operations in one semantic slot because deciding
   between them was the expensive part. → §9, §14, §18, §24
5. **An implicit assumption escaped its context.** Runtime, concurrency, or wiring
   conditions invisible at the call site — including the assumption that an artifact
   still belongs to you after it leaves the machine — including the assumption that the
   environments you tested in covered the states the code branches on. → §8, §15, §16,
   §17, §20, §22
