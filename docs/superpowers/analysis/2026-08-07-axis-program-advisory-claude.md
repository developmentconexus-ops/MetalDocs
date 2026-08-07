# Axis-program advisory — Claude arm (dual-gate), 2026-08-07

Independent read-only review of the proposed pivot: shelve the authz grant-unification spec, run a
systemic axis-discovery program plus a "software factory" method adoption first. Evidence gathered by
direct reads and four repo sweeps; every load-bearing claim cites file:line. Where I could not verify,
I say so.

---

## Q1 — Is the pivot right, or is it scope panic?

Half right, and the wrong half is the expensive one.

**The pattern diagnosis is right.** The codebase confirms the operator's shape of complaint, in code:
the error kernel's own design is inverted in practice (232 call sites use `problem.New` with a
hand-restated status vs 11 using the ADR-0089-canonical `problem.NewFor` —
`internal/platform/problem/problem.go:29-56` declares `NewFor` "the DEFAULT constructor"); `writeProblem`
is copy-pasted 10× across iam/auth/audit/security handlers; tenant extraction is re-implemented 9×
under 4 names; `documents` and `templates` carry seven identically-named duplicated sentinel errors
and twin status-transition functions (`documents/domain/state.go:21` vs `templates/domain/version.go:89`);
`templates/domain/errors.go:66-70` literally documents itself as "Mirrors controlleddocuments'
identically-named sentinel." Hand-synced-enumeration is the repo's confessed meta-defect
(final-architecture-review; ME-01's role vocabulary on **7** surfaces).

**But the pivot's first move — shelving the authz spec — is scope panic.** The axis-discovery program
the operator wants *already exists and already ran*: it is
`docs/engineering/mechanical-enforcement-register.md`. ME-01..ME-12 are exactly "axes found
systematically, filed as issues (#74–#85), with a firing-mechanism doctrine (levels 1–5, register
§doctrine) and owners." And the register assigns ME-01, ME-02, ME-03, ME-09, ME-11, ME-12 — half the
register — to **the authz grant-unification program as owner**. The spec
(`docs/superpowers/specs/2026-08-07-authz-grant-unification-design.md`, v2, post-audit, residue
enumerated in §14) is not a competitor to the axis program; it is the axis program's first executable
axis, and it builds the generation infrastructure (`db/reference-data/role_catalog.yaml` → generated
Go sets + OpenAPI enums + seed SQL, spec §3.1) that every subsequent single-upstream axis will copy.
Shelving a finished, twice-audited design to go find "all the axes" means re-deriving a list that is
already 60% written down, while the one axis with a ready design goes stale against a moving codebase.

The honest reading of "a second-pass audit still found 15 defects": that is the *method working*, not
the method failing. The wrong-noun defects inside wrong-noun guards (ME-02, ME-09, ME-12) were found
because the register's discipline forced a second pass. The lesson is "push enforcement from level 3
(lint) to level 1 (generated/unrepresentable)," which the v2 spec already internalizes — not "stop and
survey everything."

What makes a real axis program different from "rewrite everything": (1) each axis terminates in a
**mechanism** (a generator + byte-parity golden, or a ratchet lint with a shrink-only allowlist),
not in cleaned-up code; (2) axes are sequenced by which mechanism the next axis consumes; (3) nothing
is refactored that no mechanism will then hold in place. The repo already has the exemplar:
tripwire arms generated from `internal/platform/tripwire/arms.go`, rendered by `render.go`, held by
`TRIPWIRE-ARM-PARITY` byte-equality in `scripts/api-lint/tripwire_arm_rules.go:71` plus a Go test.
That pipeline is the program's template. If a proposed axis cannot name its terminal mechanism, it is
a rewrite wearing an axis costume.

---

## Q2 — Method: how real orgs make quality mechanical, free/OSS, and the AI-author inversion

First, calibrate: this repo is **not** behind industry practice on mechanism — it is ahead of most
mid-size companies. Twenty GitHub workflows; a custom blocking lint suite (`scripts/api-lint/` — ~20
rules, "EVERY rule this linter emits is BLOCKING", `main.go:10-16`); oasdiff breaking-change gate
(`openapi-breaking.yml`); codegen-drift gates both directions; migration-gapless + no-historical-edit
check (`invariants.yml`); gitleaks; gosec+govulncheck; an ADR-status CI gate (`check-adr-status.sh`);
and shrink-only allowlists with per-entry prose reasons (`check-test-discipline.sh`) — the "quality
ratchet" pattern (Betterer-style) already implemented by hand. The operator's "muito manual, sem
eficiência" is true of the *code*, not of the *gate infrastructure*. The gap is narrower and more
specific than he thinks.

**Named tools, mapped to defect class, with real limits:**

- **Architecture / import boundaries.** `go-arch-lint` or `arch-go` (both OSS, YAML component rules,
  arch-go has cycle detection) — but both are package-granularity and cannot distinguish "port
  interface" from "concrete aggregate," which is exactly the distinction this repo needs (Q5). The
  right move is not adoption but extending the in-house checker into a graph-aware Go tool inside
  `tools/cilint` (which already exists and runs blocking in `invariants.yml`). For TypeScript:
  **dependency-cruiser** (rules-as-code, cycles, orphans; limit: JS/TS only) — the frontend currently
  has nothing equivalent.
- **Semantic lint / rules-as-code.** For Go, **go-ruleguard** (quasilyte, runs inside golangci-lint
  via gocritic): type-aware DSL, ideal for wrong-noun rules ("no `iamtypes.RoleSystemAdmin` compared
  outside the evaluator" — ME-12's level-3 target); limit: mostly single-package analysis scope.
  **Semgrep OSS** for cross-language pattern rules; hard limit: the free engine is per-file only —
  no cross-file dataflow/taint (that is Pro) — fine for noun/shape rules, useless for flow rules.
  **ast-grep**: very fast structural matching, but syntactic only, no type info. For this codebase,
  ruleguard > semgrep for Go because the defect class is typed-noun misuse, not string patterns.
- **Contract & schema drift.** Already solved: oapi-codegen drift jobs + oasdiff + problem-codes
  freshness (`api-contract.yml`). Nothing to buy. **sqlc** would catch wrong-column SQL at compile
  time but does not fit the existing hand-built query-builder style; adopting it is a rewrite of the
  data layer — do not.
- **DB invariant testing.** **pgTAP** is the industry OSS answer (assert triggers, constraints, and
  RLS behavior *per role* as SQL tests). Real value here: the repo learned that RLS was inert in dev
  because the app role was superuser (M6 memory) and that a non-owner CI role kills false-greens (M7);
  pgTAP tests running as `metaldocs_ci` would make that class permanent. Limit: another test runner
  and harness to maintain; the existing Go integration suite + tripwire goldens already cover most of
  it, so adopt pgTAP narrowly for RLS/trigger semantics or skip.
- **Mutation testing.** **gremlins** (Go) and **StrykerJS** (TS). Truth: mutation testing repo-wide on
  a 180k-LOC monolith is a furnace — slow, noisy, and the score is unactionable. Legitimate scoped
  use: run gremlins *only* on guard packages (`internal/platform/tripwire`, `scripts/api-lint`,
  `internal/modules/iam/authz`) to prove the guards' own tests bite. Annotate-only, never blocking.
- **Dead code / duplication.** `golang.org/x/tools/cmd/deadcode` (whole-program, catches what
  staticcheck U1000 misses across packages); **knip** for TS dead exports; **jscpd** or PMD CPD for
  duplication reports (both handle Go). Duplication detection is annotate-only by nature — a CPD
  report would have surfaced the byte-identical objectstore switch
  (`documents/application/service.go:655-666` vs `templates/application/autosave.go:152-163`) and the
  duplicate Problem converters (`distribution/delivery/http/handler.go:197-215` ≡
  `notifications/delivery/http/handler.go:232-250`) mechanically.
- **ADR/decision governance.** Already ahead: ADR-status CI gate + REQ-traceability gate
  (`req-traceability.yml`, MUST-classified REQs need resolvable evidence). log4brains adds a website,
  not enforcement. Skip.
- **CI gate topology.** The correct free-tier topology, which the repo mostly has: **blocking** =
  deterministic, zero-false-positive checks (drift, parity, boundaries, contract, secrets, build,
  tests); **annotate-only** = statistical or judgment checks (coverage deltas, duplication, mutation
  score); **ratchet** = anything with legacy debt (allowlist may only shrink; addition requires an
  ADR row — `check-module-boundaries.ps1:82-84` already implements this with an empty debt list).
  Concrete gaps found: `golangci-lint.yml` runs `only-new-issues: true` (pre-existing findings in
  untouched files are invisible — fine as ratchet, but undeclared); ~6 check scripts
  (`check-db-bootstrap.ps1`, `check-baseline-equivalence.ps1`, `wiki-tally-check.ps1`, …) are
  referenced by **no workflow** — they run on human memory, i.e. they are level-5 mechanisms
  pretending to be level 3; and there is **no aggregate local gate at all** — no Makefile
  verify target, no pre-commit hook, `.git/hooks` has only samples.

**The AI-author inversion — what transfers, what inverts, what is theatre.**

The software-factory scene the operator described (push PR → review bot → auto-opened issues →
conformance evaluators) assumes the bottleneck is *human carelessness caught by a machine*. Here the
author is a machine whose failure mode is the opposite of carelessness: **confident, internally
consistent wrongness** — the wrong noun used fluently everywhere, including inside the guard written
to catch it (ME-02: the REQ-AUTHZ-5 guard covers capabilities, not roles, and "reads as complete in
every audit"). Three consequences:

1. **What transfers:** every deterministic gate. Drift, parity, byte-golden, boundary, contract
   checks work identically regardless of who authored the diff. These are the repo's strength.
2. **What inverts:** the *position* of the gate. For human authors, post-hoc PR review is where
   learning happens. An AI agent does not learn from last week's PR comment; it learns from **red
   feedback inside the current loop**. So the gate must move from "CI tells you after push" to
   (a) **level 1 — unrepresentable**: the artifact is generated, the agent can only edit the upstream
   (`role_catalog.yaml`, `tripwire/arms.go`), and hand-editing the derived file goes red by byte-parity.
   The AI *cannot* ship the wrong-noun defect because the noun exists in exactly one place. This is
   the register's own doctrine (levels 1–5) and it is the correct answer to the operator's question —
   the "review bot" he watched is a level-4/5 mechanism; this repo should be spending on level 1–2.
   (b) **In-loop hooks**: the repo currently has zero Claude-Code hooks and no `make verify`. Adding
   a settings hook that runs the scoped lints (api-lint code rules, boundary check, ruleguard pack)
   after edits — plus one aggregate local target mirroring CI — puts the red light in the agent's own
   turn, before commit. That is the single cheapest "software factory" upgrade available, and it
   costs nothing.
3. **What is theatre:** an AI review bot commenting on AI-authored PRs with no human in the loop.
   reviewdog/Danger map linter output to PR comments — comments nobody reads are dead weight; every
   check should either block or write a file, never converse. SonarQube CE self-hosted: a server to
   feed and patch, whose generic rules are strictly weaker than the in-house api-lint suite. The
   auto-issue-opening machinery: the register + GitHub issues #74–#85 already does this with higher
   signal than any bot would. The one *review* that genuinely works here is the one already practiced:
   independent-model adversarial review of **designs** (it found the 15 v2 defects). Keep it for
   specs and validation contracts; do not build it into a per-PR ritual.

---

## Q3 — Axis decomposition

**(a) Error model — partially unified; small, highly mechanical axis. KEEP, demoted in scope.**
The kernel is real and universally adopted at the wire level: `internal/platform/problem/` (RFC 9457,
closed `Code` type, family→status map, `code.go:17-40` records the 147-code/3-convention cleanup);
zero hand-written `http.Error` (all 48 hits are oapi-codegen boilerplate). Partial in three ways:
the status-binding design is inverted (232 `problem.New` vs 11 `problem.NewFor` — the exact drift ADR
0089 was written to close, reopened by convention); one 350-line hand-rolled `errors.Is` mapper
survives (`internal/modules/approval/http/errors.go:208-553`, ~70 sentinels, 52 code registrations);
`writeProblem` duplicated 10×, and 15 generated per-module `Problem` structs need hand converters
(two byte-identical). Terminal mechanism: a ruleguard rule banning `problem.New` outside the kernel
package + one shared `httpresponse.WriteProblem` + (optionally) a generated sentinel→code mapper.
Size: days, not weeks.

**(b) Authn/session — the operator's "repetição" claim is half-wrong. RESHAPE.**
The scary version is false: there is exactly **one** validation path — cookie session only, one
`ResolveSession` implementation with one caller (`auth/application/service.go:408` ←
`auth/delivery/http/middleware.go:102`), zero bearer/JWT anywhere, and the `tokens` module is the
template-placeholder catalog, not auth. The true version is a **boundary** defect, not a path defect:
`auth/application/service.go` (1,164 LOC) is auth *plus* a full user-CRUD service (`CreateUser:613`,
`UpdateUser:683`, `ListUsers:573`, `AdminResetPassword:743`…) overlapping iam's PeopleHandler; iam
binds the **concrete** `authapp.Service` (`iam/application/people_service.go:34`) and serializes
auth's domain structs as its own DTOs; auth's *domain* is typed in iam's vocabulary
(`auth/domain/model.go:8`); a hand-synced in-memory repo (626 LOC, 28 methods) mirrors the postgres
one; 4 Tx/non-Tx verbatim wrapper pairs; "Authentication required" emitted from 18 sites via 3
mechanisms. Plus the register's two open category questions: ME-07 (hand-rolled IdP, build-vs-adopt
unscheduled) and ME-08 (MFA coverage dashboard with no MFA — "a dashboard, not a control"). This axis
is really **"iam↔auth boundary + identity ownership"**, and it merges with Q5 pair 1.

**(c) documents vs controlleddocuments vs templates — the split is substantially DOMAIN-TRUE. KILL
the merge candidate; replace it with a different, real axis.** This is the answer the operator cannot
see from memory, so here is the code:

- **Table ownership is fully disjoint.** Each module declares its tables in a `TenantDataPort` and a
  SQL-literal sweep confirms zero overlap: documents owns `documents/document_revisions/…`
  (`documents/infrastructure/tenant_data_port.go:37-45`), CD owns
  `controlled_documents/cd_sequence_counters/…` (`controlleddocuments/infrastructure/tenant_data_port.go:40-47`),
  templates owns `templates_template(_version)` (`templates/infrastructure/tenant_data_port.go:42-47`).
- **controlleddocuments is not a lifecycle overlay — it is a distinct identity/numbering aggregate.**
  Its own doc-comment says so: the `ControlledDocument` "carries no content of its own — Code is the
  stable, immutable, audit-traceable identity; the chain of documents rows holds the actual content"
  (`controlleddocuments/domain/controlled_document.go:30-34`); it allocates `{PROFILE}-{AREA}-{NNN}`
  codes (`:141-146`) from its own sequence table. The review/expiry lifecycle the name suggests lives
  entirely in `documents` (columns on `public.documents`, `db/baseline/0001_current_schema.sql:1878-1911`;
  zero production hits for review/expiry in CD). There is no "controlled copy" concept anywhere.
  Structurally: CD is the **factory** for documents — there is no `POST /documents` in the API at all
  (`api/openapi/v1/openapi.yaml:2654`, GET-only); creation flows exclusively through
  `POST /controlled-documents(/…/revisions)` via the `DocumentInitializer` port that documents
  implements (`controlleddocuments/domain/document_initializer.go:91-116` ←
  `documents/application/cd_initializer.go:16-93`).
- **templates is not "a document with a different status."** Different revision physics: template
  versions are draft-mutable, overwritten in place at a deterministic object key
  (`templates/application/keys.go:17-23`); document revisions are append-only and content-hash-
  addressed under an editor-session lock (`documents/application/keys.go:11-13`). Templates is a
  **leaf** — it imports neither documents nor controlleddocuments (zero Go imports; only two prose
  comments reference them).
- **A ratified decision already exists**: `docs/superpowers/specs/2026-06-30-template-document-parity-design.md:29-30`
  (Approved): "No backend module merge… Merging would violate the module-boundary invariant. Dedup is
  frontend + lifecycle-semantic only." The status-enum overlap is a *deliberate alignment* from that
  spec (D2), not drift.

So the split survives scrutiny. **What the code actually indicts is different**, and it is a real axis:

1. **Duplicated lifecycle machinery** between documents and templates: 7 identically-named sentinels
   defined twice (`ErrInvalidStateTransition` at `documents/domain/model.go:174` AND
   `templates/domain/version.go:120`, plus 6 more); twin transition validators; a byte-for-byte
   4-arm objectstore error-translation switch; twin autosave presign/commit flows; parallel
   creation-gate rules and tests; handlers that carry comments pointing at *each other* as the
   pattern being mirrored (`templates/delivery/http/handler.go:174`,
   `documents/delivery/http/handler.go:154`). The approval module already proved the correct shape —
   one subject-generic kernel (`approval/domain/subject.go:16-17`, `subject_kind ∈ {document,template}`,
   CHECK-constrained). The missing structure is a **content-lifecycle kernel** (status machine,
   autosave/objectstore translation, key building, sentinel vocabulary) consumed by both, exactly as
   approval is.
2. **Ownership inversion**: documents' status writes live in *approval* — 9 `UPDATE documents` sites
   across approval's services (`release_coordinator.go:333,475`, `document_terminal_approval.go:129`, …),
   while templates keeps publish in-module (`templates/application/lifecycle.go:52-157`) with approval
   writing back through a narrow adapter. Two subjects, two opposite ownership models, one kernel.
   Pick one (the templates shape — subject owns its writes, approval commands through a port — is the
   cleaner one).
3. **Small boundary leaks**: documents executes CD's domain policy in-process
   (`controlleddocumentsdomain.Resolve` at `documents/application/service.go:12`); CD matches
   documents' constraint name `ux_documents_cd_active` by string
   (`controlleddocuments/application/service.go:950-956`); the doc↔CD wiring cycle is broken by a
   post-construction setter (`apps/api/cmd/metaldocs-api/main.go:811`); and
   `documents/infrastructure/active_instance_reader.go:13-20` openly admits the module split does not
   match table ownership for approval-instance reads.

**(d) authz grants — real, designed, and the register's center of gravity.** Owns ME-01/02/03/09/11/12.
Nothing new to add; the v2 spec's residue list (§14) is honest. This axis goes first (Q4).

**(e) observability — NOT SUPPORTED as an axis.** I found no evidence of a systemic observability
defect in this pass (unverified — I did not probe deeply). What the register *did* find is the inverse
defect: observability **of an unimplemented control** (ME-08, MFA coverage endpoint with no MFA
enrollment/challenge anywhere in `Authenticate`, `auth/application/service.go:265`). That is an
authn-axis item, not an observability axis. Kill (e) unless a dedicated probe produces evidence.

**Axes the operator did not name, that the code names:**

- **HTTP delivery boilerplate** (couples to (a)): tenant extraction ×9 under 4 names, `writeProblem`
  ×10, "Authentication required" from 18 sites via 3 mechanisms with case-inconsistent titles
  (`controlleddocuments/delivery/http/routes.go:612` lowercase). Terminal mechanism: one shared
  handler kit + a lint banning module-local re-implementations.
- **Hand-synced enumerations** (the confessed meta-defect): role vocabulary on 7 surfaces (ME-01),
  Go↔OpenAPI mirror comments (`internal/platform/iamtypes/role.go:55-56`), Go↔SQL CHECK mirrors
  (`role.go:70-77`), the lint-and-test pair that mirror each other by prose comment (ME-11). Largely
  subsumed by (d) + the register; the residue is a standing rule: every new enum names its single
  upstream or fails review.
- **Dual hand-synced repo implementations** (memory vs postgres in auth, 28 vs 24 methods) — small,
  folds into (b).

---

## Q4 — Dependency order

**Axis 0 (parallel, immediately): boundary-linter upgrade + local gate.** Pure tooling, no code churn,
makes every later axis safe to execute: graph-aware boundary lint (Q5), ruleguard pack, `make verify`,
Claude-Code hooks. Nothing depends on it semantically; everything depends on it operationally.

**Axis 1: authz grant unification — the already-written spec. First, and unshelved.** Three reasons it
is the correct first domino: (i) it deletes the noun-drift that every other axis's guards would
otherwise have to encode against *two* grant vocabularies (any lint written before it must be
rewritten after it — that is the redo-cost the question asks about); (ii) it builds the
single-upstream generation pipeline (`role_catalog.yaml` → Go/OpenAPI/SQL) that axes (a), the enum
axis, and the delivery kit will instantiate again; (iii) six of twelve register entries are its
children — executing it closes half the register in one program.

**Axis 2: error/delivery kernel completion** ((a)+(HTTP boilerplate)). Cheap, mechanical, and it
shrinks the handler surface *before* the big lifecycle axis churns those same files. Doing it after
Axis 3 means touching every handler twice.

**Axis 3: content-lifecycle kernel + ownership repair** (the reshaped (c)). The biggest axis. Must
follow Axis 1 because every publish/signoff/transition site carries capability asserts and tripwire
arms that Axis 1 re-points (spec §6.2); doing lifecycle first means re-visiting all 9
approval-owned `UPDATE documents` sites twice.

**Axis 4: iam↔auth boundary + identity ownership** (the reshaped (b), incl. ME-07 build-vs-adopt study
and the ME-08 delete-or-implement ruling). Follows Axis 1 because grant unification already rewrites
iam's internals; moving user-CRUD across the boundary mid-flight would collide.

Wrong-order hazard, concretely: lifecycle kernel before authz = double-touch of every status-write
site; delivery kit after lifecycle = double-touch of every handler; any lint pack before the linter
upgrade = rules asserted at the wrong granularity that must be re-expressed.

---

## Q5 — The import graph and the 7 mutual pairs

Ground truth first: `module-boundaries.yml` is a 21-line workflow shelling to
`scripts/check-module-boundaries.ps1`, which is a per-file, per-import **string match on the target
layer name** (`$allowedLayers = domain|application|api`, `:50`; published packages `authz`, `fanout`,
`fanout/dispatchjobs`, `resolvers`, `:56-61`). It builds no graph, so **cycles are structurally
invisible**; it never constrains the *source* layer; and it cannot distinguish a port interface from
a concrete aggregate, free function, or sentinel. All 14 directions of the 7 cycles pass it, because
every import targets `domain`, `application`, or `iam/authz`.

Pair verdicts:

1. **iam↔auth — real violation, worst of the seven.** auth's *domain* is typed in iam's domain
   vocabulary (`auth/domain/model.go:8`, `port.go:12` — a domain-layer edge, unbreakable by
   inversion); iam binds the **concrete** `authapp.Service` (`iam/application/people_service.go:34`)
   and iam's HTTP layer serializes auth domain structs directly. Missing abstraction: extend the
   `iamtypes` precedent (`internal/platform/iamtypes/role.go:1-15` already broke one such cycle) —
   move `Capability` to platform types; then either publish a narrow auth application interface or,
   better, move user-CRUD into iam and shrink auth to sessions/credentials (Axis 4).
2. **iam↔taxonomy — legitimate.** taxonomy consumes the sanctioned `iam/authz` tool package +
   capability constants; iam consumes one read-port with a null-object
   (`iam/infrastructure/postgres/area_catalog_reader.go:7`, cites ADR-0039 D3(b)). A policy cycle,
   not a data cycle. Keep.
3. **iam↔security — legitimate.** One site each way, ports only
   (`security/infrastructure/postgres/repository.go:16`; `iam/application/tenant_lifecycle_service.go:46`).
   Keep.
4. **documents↔approval — real violation, heaviest.** approval→documents drives documents' state
   machine with raw status constants + `CanTransitionDocumentStatus`, and imports documents'
   application-layer **free functions** (`LoadDocumentAreaCode`, `LoadDocumentControlledDocumentID`,
   `ApproverContext` — `approval/application/{review_verdict_service.go:13,obsolete_service.go:12,decision_service.go:21}`)
   with no declared port, plus the 9 `UPDATE documents` sites. Missing abstraction: a
   documents-published lifecycle command port (documents owns its writes; approval requests
   transitions) — i.e. the templates-side pattern
   (`templates/infrastructure/approval_completion_writer.go:37`) generalized. Axis 3.
5. **documents↔controlleddocuments — half legitimate.** CD→documents is a textbook port
   (`ActiveInstanceReader` + Noop). documents→CD is mixed: `CDFieldReader` is a proper port, but
   documents executes CD's `Resolve` domain policy and consumes its aggregates
   (`documents/application/service.go:12`, `cd_initializer.go:8`), plus the setter-broken wiring
   cycle (`main.go:808-811`) and the constraint-name leak. Missing abstraction: a published
   template-resolution port on CD; creation orchestration declared to live in CD alone.
6. **controlleddocuments↔approval — legitimate.** Exactly one named port each way
   (`CDFieldReader` ↔ `RouteReadinessReader`) plus a subject-kind enum value. Cleanest of the
   documents-family cycles; not worth breaking.
7. **taxonomy↔approval — intentional but note the domain edge.** approval's *domain* embeds
   taxonomy's `RoutePolicy` (`approval/domain/route.go:6`); taxonomy deliberately publishes it as
   "the narrow, approval-facing consequence of a GovernanceClass"
   (`taxonomy/domain/governance_class.go:35`). Shared-vocabulary coupling, deliberate. Acceptable;
   if it ever bothers, `RoutePolicy` moves to a platform types package like `iamtypes`.

**What the boundary linter must assert that it does not today:**

1. **A checked-in allowed-edge manifest** — explicit (source-module, target-module, direction) pairs,
   ratchet semantics (may only shrink; additions need an ADR row, same as `$debtAllowList`).
2. **Module-level cycle detection** against that manifest — new mutual pairs go red.
3. **Port-vs-internals distinction**: cross-module `domain` imports restricted to declared port files
   (`*_port.go`), published enums/sentinels via a per-module export manifest; cross-module
   **application**-layer imports banned outright (today they are whitelisted — that is what legalized
   `authapp.Service` and approval's free-function reach).
4. **Source-layer constraint**: another module may not be imported from `domain` except from the
   shared-vocabulary allowlist (would have flagged `auth/domain/model.go:8` and
   `approval/domain/route.go:6` at introduction).
5. **No new post-construction setter cycles** in wiring (the `main.go:811` pattern), assertable by a
   ruleguard rule on `With*` setters called after module construction in `cmd/`.

---

## Q6 — What NOT to do

1. **Do not merge documents/controlleddocuments/templates.** The evidence (Q3c) is against it, a
   ratified Approved design explicitly forbids it, and a merge would be the definitional
   rewrite-dressed-as-refactor: months of churn to erase boundaries that mostly encode real domain
   distinctions, while the actual defect (duplicated lifecycle machinery) is fixable by extraction.
2. **Do not shelve the authz spec.** That is the program's pilot axis and its generation-template
   factory. Shelving it converts finished work into stale work.
3. **Do not build the PR-review-bot factory.** No SonarQube CE server, no reviewdog/Danger comment
   plumbing, no auto-issue bot. Comments to no reader are theatre; the register + issues + blocking
   lints already do the real version. The AI-facing gate belongs in-loop (hooks + `make verify` +
   level-1 generation), not post-push.
4. **Do not run mutation testing repo-wide** or chase a coverage/mutation score. Scoped gremlins on
   guard packages, annotate-only, or skip.
5. **Do not big-bang the import graph.** 3 of 7 cycles are legitimate published-port pairs and one
   more is intentional shared vocabulary; only iam↔auth and documents↔approval merit structural
   repair, and each belongs to an already-sequenced axis. A standalone "fix all cycles" program would
   redo Axis 3/4 work out of order.
6. **Do not adopt sqlc/pgTAP/squawk wholesale.** sqlc fights the existing query-builder style;
   squawk gates migrations that are currently an empty folder post-fold; pgTAP only if scoped to
   RLS/trigger semantics under the non-owner CI role.
7. **Do not let the method-research phase become the deliverable.** The org-practice research
   compresses to one sentence this repo has already written down for itself
   (register doctrine, levels 1–5): *prefer unrepresentable over guarded, guarded over reviewed* —
   and the marginal free-tool adoptions above are a week of work, not a program.

---

## Verdict

**Proceed-with-changes.** Keep the operator's systemic ambition but invert his first move: unshelve
the authz spec and execute it as Axis 1 (it owns half the register and mints the generation template),
while Axis 0 (graph-aware boundary linter, ruleguard pack, `make verify`, in-loop agent hooks — all
free, all small) runs in parallel; sequence error/delivery-kernel, then content-lifecycle-kernel, then
iam/auth-boundary behind it; kill the module-merge candidate and the observability axis for lack of
evidence; and replace the software-factory fantasy with the two mechanisms that actually fit an
AI-authored solo repo — level-1 generation with byte-parity goldens, and red feedback inside the
agent's own loop. The ONE thing that would most change this answer: a demonstrated *behavioral*
divergence between documents' and templates' duplicated lifecycle code (an invariant enforced on one
subject and silently absent on the other, in live QA) — that would promote the lifecycle kernel from
Axis 3 to Axis 1 ahead of authz, because it would mean the duplication is already shipping defects
rather than merely accruing maintenance risk.
