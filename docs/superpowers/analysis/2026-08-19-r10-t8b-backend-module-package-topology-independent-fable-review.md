# R10-T8B — Backend Module & Package Topology — Independent Fable Review

```text
INDEPENDENT REVIEW
EVIDENCE ONLY
NOT TARGET AUTHORITY
```

> **Status:** INDEPENDENT COLD ADVERSARIAL REVIEW / NON-AUTHORITATIVE
> **Date:** 2026-08-19
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Reviewed remote HEAD:** `2153cbdb67276b7a79091c348563bb4415df4566`
> **Method:** DevelopmentConexus Engineering Method v1.0.0 (mirror `docs/engineering/standards/root-cause-global-maximum-method.md`)
> **Stage under review:** T8-B — ACTIVE
> **Implementation:** BLOCKED

This document is reviewer evidence. It is not target authority, not T8-B promotion, and not implementation authorization. The Lead adjudicates; the operator ratifies.

---

## 0. Review provenance and honesty statement

### 0.1 HEAD verification

The review request asserted an expected remote HEAD. That assertion was not trusted; it was revalidated:

```text
gh pr view 131            → headRefOid 2153cbdb67276b7a79091c348563bb4415df4566
                            state OPEN / base main / head docs/a8-authz-approval-redesign-ledger
git fetch origin docs/a8-authz-approval-redesign-ledger
git rev-parse origin/...   → 2153cbdb67276b7a79091c348563bb4415df4566
```

HEAD did not move. **Reviewed HEAD = `2153cbdb67276b7a79091c348563bb4415df4566`.** Review executed in a detached worktree pinned at that commit.

### 0.2 Candidate provenance — traceability gap

The T8-B candidate under review (the `A`→`I` sections and the `T8B-D01..D24` decision set) **does not exist in the repository at the reviewed HEAD**. A repository-wide search for `T8B-D01` and for the candidate's own non-authority banner returns nothing. The only T8-B artifact in the tree is the bootstrap:

```text
docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-bootstrap.md
```

Consequences, recorded rather than silently absorbed:

1. the reviewed candidate is **prompt-supplied**, not repository-reproducible;
2. a later reader of this review cannot diff the reviewed candidate against a committed source;
3. §11 of this review therefore restates the corrected candidate in full so that the adjudication has a durable, auditable subject.

This is a process finding (`L-01`), not an architecture finding. It does not weaken the candidate's technical content.

### 0.3 Authority chain actually read at the reviewed HEAD

```text
AGENTS.md
docs/engineering/standards/root-cause-global-maximum-method.md
wiki/references/current-agent-handoff.md
wiki/architecture/r10-technical-architecture.md
wiki/architecture/launch-v1-product-contract.md            (structure + §4–§8 read)
wiki/architecture/whole-product-alignment-review.md
wiki/architecture/launch-v1-ownership-topology.md
wiki/architecture/r10-t1-semantic-state-invariants.md
wiki/architecture/r10-t2-governance-effectivity-transactions.md
wiki/architecture/r10-t3-authorization-audit-enforcement.md
wiki/architecture/r10-t3-d4-responsible-owner-eligibility-amendment.md
wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
wiki/architecture/r10-t5-durable-async-search-external-effects.md
wiki/architecture/r10-t6-canonical-api-frontend-journeys.md   (§1–§5, §18–§29, §31 read)
wiki/architecture/r10-t7-historical-migration-truth-semantic-mapping.md
wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md
wiki/architecture/rebaseline-decision-registry-t8a-amendment.md
wiki/architecture/r10-post-t6-implementation-readiness-program.md
wiki/architecture/r10-technical-realization-reconciliation-baseline.md
docs/superpowers/analysis/2026-08-19-r10-t8b-backend-module-package-topology-bootstrap.md
```

**Active-authority contradictions found: none.** The router, handoff, bootstrap, T8-A page and T8-A registry amendment agree on stage state, on the T8-A Global Maximum, on the 4+1 ownership baseline and on the T8-B/T8-C boundary. One phrasing tension is noted in `M-03` (bootstrap §4 and handoff say a semantic owner does **not** imply exactly one Go package; the candidate's D03 makes one package the default) — this is a candidate-vs-authority tension, not an authority-vs-authority contradiction.

### 0.4 Current-implementation evidence remeasured at this HEAD

Only load-bearing evidence was remeasured. Everything below is `CURRENT-PROVEN at 2153cbdb`.

| Fact | Value | Command / anchor |
|---|---|---|
| Go module path | `module metaldocs` | `go.mod:1` |
| Runtime shells | `apps/{api,worker,jobs,dbprovision}/cmd/…` | `find apps -maxdepth 3 -type d -name cmd` |
| Root `cmd/` content | `gen-http-surface`, `gen-tripwire`, `problem-codes-dump` (generators, **not** runtime shells) | `ls cmd/` |
| Non-Go runtime | `apps/docx-renderer` is Node/TS (`package.json`, `tsconfig.json`) | `ls apps/docx-renderer` |
| Other non-Go workspaces | `frontend/`, `packages/{docx-editor,editor-ui,eigenpal-adapter,form-ui,shared-tokens,shared-types}` | `ls packages/` |
| `internal/` roots | `composition`, `modules`, `platform`, `test` | `ls internal/` |
| Legacy modules | 15 under `internal/modules/` | `ls internal/modules/` |
| Platform packages | 38 under `internal/platform/` | `ls internal/platform/` |
| `internal/` package dirs | 180 directories | `find internal -type d \| wc -l` |
| `internal/composition` content | one file (`tenantdata/registry/registry.go`) — real wiring lives elsewhere | `find internal/composition -name '*.go'` |
| Nested owner-private `internal/` under `internal/` | **none** | `find internal -mindepth 1 -type d -name internal` |
| Migrations / DB assets | `db/{baseline,dev-seeds,grants,migrations,prerequisites,reference-data}` | `ls db/` |
| Test roots | `tests/{unit,integration,e2e,docx_v2}` | `ls tests/` |
| Verification registry | `tools/verify` (check registry / SSOT orchestrator) | `tools/verify/registry.go:1-9` |
| Architecture analyzers | `tools/cilint/internal/analyzers/*` — **not** in `tools/verify` | `ls tools/cilint/internal/analyzers/` |

Legacy size, `CURRENT-STATE ONLY`, non-test Go LOC (no target entitlement whatsoever — used only as an order-of-magnitude input to `M-03`):

```text
approval             102 files   27,729 LOC
documents             52 files   15,913 LOC
iam                   69 files   16,078 LOC
templates             41 files   10,332 LOC
taxonomy              28 files    7,970 LOC
controlleddocuments   19 files    6,402 LOC
auth                  14 files    4,630 LOC
audit                 11 files    3,710 LOC
```

Proven platform→module dependency leaks (the failure class `platform/` must not repeat):

```text
internal/platform/authn/config.go:12-13    → modules/auth/application, modules/iam/domain
internal/platform/authn/context.go:8       → modules/iam/domain
internal/platform/docgenv2/…reader.go:9-10 → modules/documents/application, modules/documents/domain
internal/platform/bootstrap/api.go:13-18   → modules/{audit,auth,iam}/{domain,infrastructure/postgres}
```

The guard that freezes those exceptions is `tools/cilint/internal/analyzers/platformboundary.go:11,19`. Its allowlist still contains `internal/platform/objectstore` — but that package imports **zero** module packages at this HEAD. **A frozen exception has outlived its violation.** That is direct evidence for `M-05`.

---

## 1. Verdict

```text
APPROVE R10 T8-B BACKEND MODULE & PACKAGE TOPOLOGY
WITH MATERIAL FIXES
```

Meaning, stated without hedging:

- the candidate's **architecture classes and dependency semantics are correct** and are materially better than Alternative A and Alternative C;
- **no materially better fourth topology exists** on current evidence — the strongest alternative found is the candidate *corrected*, not the candidate replaced;
- the candidate is **not promotable as written**: four BLOCKER-class defects would freeze an incomplete dependency graph and would delegate material architecture decisions to a Writer, which the post-T6 program invariant prohibits;
- all four BLOCKERs are **omissions or internal contradictions**, each closable inside T8-B without reopening any upstream stage and without trespassing on T8-C.

### Finding counts

```text
BLOCKER   4
MAJOR     8
LOW       6
```

### Sub-verdicts

| Question | Verdict |
|---|---|
| Method / Global Maximum | **Global Maximum class CONFIRMED**; realization INCOMPLETE |
| One module vs multi-module | **ONE Go module — CONFIRMED** (D01 ACCEPT, wording fix) |
| Owner granularity | **CORRECT DIRECTION, WRONG UNIT** — freeze public surfaces, not package count |
| Controlled Documents granularity | **Do not split into peer owners — CONFIRMED**; do not gate owner-private splitting |
| Application orchestration | **CORRECT ROLE, UNDER-CONSTRAINED** — needs plurality + anti-God rules |
| Allowed/forbidden dependencies | **INCOMPLETE + INTERNALLY CONTRADICTORY** (B-02, B-03, B-04) |
| Platform / shared mechanism | **CORRECT INTENT, WRONG LAW STATEMENT** (M-01, M-02) |
| Composition / DI | **CONFIRMED** — and it is a proven correction of a real current defect |
| Mechanical enforcement | **CONCEPT RIGHT, TOO WEAK** — rejection-only, no completeness, wrong host named |
| Future evolution | **CONFIRMED** — Launch+/Future attach additively; no dismantling required |
| Upstream reopen required | **NONE** |
| Another broad review | **NO** |
| Bounded Round-2 delta review | **YES — sufficient** |
| Promotable after fixes + adjudication + ratification | **YES** |

---

## 2. Structural Inversion

> If MetalDocs had never had a 15-module `internal/modules` topology, which backend/package boundaries would the ratified R10 semantic architecture still require?

Answer, derived only from T1→T8-A:

```text
STILL REQUIRED

1. exactly five semantic homes           Authentication / Organization / Authorization /
                                         Controlled Documents / Audit
                                         (ownership-topology §1; T6 §1)

2. one inbound adaptation edge           HTTP + durable-job invocation are transports;
                                         T6 §2 makes HTTP a surface class, not an owner

3. one construction locus                T2 §2 single local ACID transaction + N runtime
                                         shells ⇒ wiring must be assemblable once and
                                         reusable by tests and every shell

4. a non-semantic mechanism region       T4 (ManagedContentStore), T5 (durable jobs),
                                         T6 (idempotency store), plus DB/observability/config
                                         are explicitly mechanism, and T5 §1 forbids
                                         inferring business truth from them

5. a cross-owner composition locus       T3 §10 offboarding writes Organization +
                                         Authentication + Authorization + Audit in ONE
                                         local transaction; T3 §9 composes Authorization
                                         grants with Controlled Documents predicates.
                                         Something must compose without owning.

6. an enforceable "mechanism ≠ authority" boundary
                                         Method §"Authority and mechanism"; bootstrap §7

STILL NOT REQUIRED

- any package named approval / templates / artifact / taxonomy / distribution /
  notifications / tokens / interchange / workflow / records (GCR A1, A3–A7, A10; T6 §31)
- a search owner (T5 §7: discovery mechanism; baseline is a canonical query/view)
- a generic domain-event bus (T5 §12)
- a tenant/RLS universal ontology (T6 §31)
- a second governance family (T2 §10)
```

**Inversion result:** the candidate's five architecture classes — `owners`, `application`, `transport`, `platform`, `composition` — are each independently re-derivable from ratified semantics without reference to the legacy tree. The candidate is **not** the legacy topology renamed. That is the single most important thing this review confirms.

The inversion also produces the review's central criticism: items **5** and **6** above are precisely where the candidate stops short. It names the composition locus but never says *what flows across it* (the transaction handle, the evidence sink, the authorization decision input) — and those three seams are exactly what make the dependency graph decidable.

---

## 3. BLOCKER findings

### B-01 — Same-local-commit Audit is structurally unenforceable under `owner → audit FORBIDDEN`

**Evidence.**
T3 §12 is unambiguous:

```text
BEGIN
business/security semantic mutation
required owning-domain evidence
required AuditEvent(s)
COMMIT
```

and "If required Audit append fails, the governed/security mutation fails with the same local transaction." T3 §15 then enumerates a census of required same-commit events that are **owner-intrinsic**: Release completed, Submission created, ACCEPT, RETURN_FOR_CHANGES, Revision cancelled, RoleAssignment granted/revoked, GroupMembership added/removed, User offboarded. T2 §9 permits Release to occur **inside** the same transaction that satisfied the last gate — i.e. deep inside a Controlled Documents state transition, not at an application entry point.

The current system enforces this lexically at the mutation site: `tools/cilint/internal/analyzers/postcommitaudit.go:15-21` flags any non-`…Tx` audit sink called after `Commit()` **inside `internal/modules/**`**.

The candidate says both of the following:

```text
E.  owner → audit   FORBIDDEN by default
D05 application ... may coordinate required Audit evidence
```

**Known / Inferred / Unknown.**
Known: audit append must be inside the owner's transaction, at sites the owner controls. Known: the current firing control's subject is the owner package tree. Inferred: under the candidate, the only legal appender is `application`. Unknown: nothing material — the contradiction is structural, not empirical.

**Root cause.** The candidate states its dependency law in terms of **package imports** but the invariant it must protect is about **call-site locality inside a transaction**. Those are different things, and collapsing them relocates a same-commit obligation away from the code that performs the commit-worthy mutation.

**Target invariant.**

> Every mutation requiring T3 §15 evidence appends that evidence inside the same local transaction, at a site the owning package controls, and no owner package imports the `audit` package.

**Why this is a BLOCKER and not a MAJOR.** Three compounding effects:

1. an owner-internal cascading transition (T2 §9 same-transaction Release; T3 §10 multi-event offboarding teardown) can commit with **no** audit append and no mechanical objection, because `application` cannot see inside the owner;
2. the existing proven control loses its subject — it scans mutation packages, and the appends move out of them;
3. the alternative reading (owners *may* append, contradicting E) leaves a Writer to choose, which the post-T6 program invariant (§2) explicitly prohibits.

**Fix (inside T8-B authority; registry amendment §8 authorizes naming seams to prove dependency direction).**
Replace the blanket prohibition with a **consumer-owned evidence port**:

```text
each semantic owner DECLARES, in its own package, the minimal evidence-append
interface it needs  (e.g. type AuditAppender interface{ Append(ctx, tx, Event) error })

internal/audit IMPLEMENTS that interface

composition INJECTS the audit implementation into each owner

ALLOWED   owner → its own declared evidence port          (no import of internal/audit)
FORBIDDEN owner → internal/audit                          (unchanged)
FORBIDDEN evidence append outside the mutating transaction
FORBIDDEN application appending evidence on an owner's behalf for an
          owner-intrinsic T3 §15 event
```

`application` retains audit coordination only for evidence about the **orchestration itself** when a cross-owner use case produces an event no single owner owns.

**Enforcement.** Keep the post-commit-audit analyzer, retarget its file scope from `internal/modules/**` to `internal/{authentication,organization,authorization,controlleddocs}/**`, and add a negative fixture proving a mutation with no in-tx append is rejected.

**Adversarial challenge to my own fix.** Does the port make Audit a lifecycle authority? No — the port is *consumer-owned* and carries only `AuditEvent` shape, which T3 §13 already bounds to ids/codes/outcomes and explicitly forbids resource-lifecycle ownership. Does it recreate owner→owner coupling? No — no owner imports `audit`; the direction is owner → own abstraction, and composition binds. This is the same dependency-inversion pattern that already worked in this repository for `platform/observability` (documented at `platformboundary.go:34-37`).

**Decision.** `D05` and `D06` require correction. `D06` is otherwise sound.

**Reopen trigger.** A concrete assurance requirement that reintroduces a deployment-wide chained/anchored Audit (explicitly deferred by GCR A9 and ownership-topology §6) would change the port's shape and must reopen this seam.

---

### B-02 — The transaction-scope carrier is undefined, so the allowed dependency graph is undecidable for every mutating operation

**Evidence.**
Four ratified authorities force one shared local transaction across package boundaries:

```text
T2 §2   one native business transition = one local ACID product-state transaction
T3 §10  offboarding = ONE transaction over Organization + Authentication +
        Authorization + Audit
T5 §5   required durable intent inserted in the SAME local transaction
T6 §18  idempotency key insert/lock and completed replay result committed with the
        semantic fact ("semantic fact commits ⇔ completed replay result commits")
```

The current system already proves the carrier is a real type crossing the domain boundary: `tools/cilint/internal/analyzers/nosqltxindomain.go:14-27,80` requires domain ports to name `metaldocs/internal/platform/db.Tx` and forbids `database/sql.{Tx,DB,Rows,Row,Result}`.

The candidate says `application` "may coordinate local transaction composition" and lists as FORBIDDEN: `owner → concrete provider/adaptor by convenience`. It never states which package owns the transaction type, whether owner public methods accept it, or whether that edge is allowed.

**Root cause.** The candidate freezes a dependency graph while leaving unstated the one value that must appear in the signature of essentially every mutating public method in the system.

**Target invariant.**

> One transaction scope is created by exactly one layer, is passed explicitly to every participant of a single business transition, is expressed as a mechanism abstraction that carries no persistence-provider identity into owner surfaces, and cannot be silently created inside an owner.

**Why BLOCKER.** T8-B's deliverable is the allowed/forbidden dependency graph. Without the carrier decision, the graph cannot be drawn for mutations, cannot be verified, and a Writer must invent it — which is exactly the class of hidden decision the post-T6 program was restructured to eliminate. This is *not* T8-D: T8-D owns schemas, constraints and lock mapping; the *package that owns the transaction abstraction* is layout and dependency direction, i.e. T8-B.

**Fix.** T8-B must freeze, at minimum:

```text
transaction scope abstraction lives in a mechanism package        platform/<tx-seam>
    (name refinable; class is not)
it is a narrow handle type, NOT a *sql.Tx / *pgx.Tx / pool type
application OPENS and COMMITS/ROLLS BACK the scope
owner public write surfaces ACCEPT the scope as an explicit parameter
owners MUST NOT open, commit or roll back a scope
transport MUST NOT open a scope
ALLOWED   owner → platform tx-scope abstraction      (explicitly listed as allowed)
FORBIDDEN owner → connection pool / driver / concrete provider type
```

Deferred to T8-D without ambiguity: isolation level realization, lock/serialization primitive selection, and the mapping from T2's "Document serialization root" to a concrete mechanism.

**Adversarial challenge.** Doesn't `owner → platform` violate the spirit of owner independence? No. The candidate's own directional law already allows `owner → consumer-owned technical abstractions`, and T5/T6/T2 make a shared scope non-optional. The real risk is the opposite one — an *implicit* ambient transaction (context-carried, invisible in signatures), which would make the same-commit invariants unverifiable. Explicit-parameter is therefore the safer and more provable choice, and it matches the property already enforced at `nosqltxindomain.go:80`.

**Decision.** New decision required (proposed `T8B-D25`). `D05` and `D09` need corresponding text.

---

### B-03 — `transport → owner` is internally contradictory: E forbids it by omission, D07 permits it by implication

**Evidence.** The candidate states, in section E:

```text
transport
  → application
```

and in the decision set:

```text
T8B-D07  transport depends on application, not owner-private implementations
```

The forbidden list contains `transport → owner-private implementation` and `transport → SQL`, but **not** `transport → owner public surface`.

These are two different laws. `D07` reads as "transport may reach an owner's public surface, just not its internals." Section E reads as "application is transport's only downstream."

**Materiality.** T6 §29 freezes 41 `GET` operations and 14 purpose-built read models (`DocumentOfficialView`, `DocumentWorkView`, `GovernanceCaseView`, `DocumentHistoryView`, `WorkAuthoringItem`, `AuditEventView`, …, T6 §26). The ambiguity decides whether every one of those reads passes through `application` or whether transport may call owners directly. That is a first-order topology question, not a detail:

- reading E strictly ⇒ `application` grows ~40 pass-through query methods (interface ceremony, attack surface 16);
- reading D07 ⇒ two inbound doors into owners, and the door that skips `application` also skips the place where T3 authorization composition happens (see `B-04`).

**Root cause.** The candidate expresses the inbound rule twice, once as a graph and once as a prohibition, and the two are not equivalent.

**Target invariant.**

> There is exactly one inbound door into semantic authority per operation class, and the door is the same place where authorization is composed.

**Fix — recommended resolution (and my reasoning is stated so it can be attacked).**
Make `application` the sole inbound door for **both** commands and queries, and remove the pass-through objection by construction rather than by exception:

```text
ALLOWED    transport → application
FORBIDDEN  transport → any semantic owner package (public or private)
FORBIDDEN  transport → platform persistence
```

Justification that this is not ceremony: T3 §9 makes **reads authorization-composed too** — `document.read_effective` requires a grant *and* a scope match *and* the Controlled Documents predicate "target Revision = current EFFECTIVE"; `document.read_working` requires the responsible-owner relationship. A "simple read" that bypasses `application` therefore bypasses the composition site. T6 §26 additionally requires `allowed_actions` to derive from "the same canonical T3 permission/scope components plus Controlled Documents predicates used by command authorization, or a provably shared equivalent" — a second read path makes "provably shared" strictly harder to prove.

The residual pass-through cost is real but is the *smaller* accidental complexity, and `M-04`'s lens-scoped application removes most of it by making each read model belong to exactly one small package rather than to one God interface.

**Rejected alternative (recorded because it is credible).** Allow `transport → owner public read surface` only for single-owner, non-authorization-composed reads. Rejected because the qualifying set is approximately empty under T3 §9, and a rule whose exception set is empty is worse than no exception.

**Decision.** `D07` requires rewording; section E requires the explicit prohibition.

---

### B-04 — Zero owner→owner imports is realizable **only** if the Authorization decision input contract is named; otherwise the candidate is self-contradictory

**Evidence.** T3 §2 canonical equation:

```text
enabled authenticated User
+ (direct User RoleAssignments ∪ GroupMemberships → Group RoleAssignments)
+ static Role → Permission bundle
+ scope match
+ Controlled Documents relationship/state/governance predicates
= ALLOW      (otherwise default DENY)
```

Ownership topology §4: "Authorization owns grants. The owning business domain owns case/resource relationship meaning and lifecycle predicates."

The candidate forbids:

```text
authorization → controlleddocs   (implied by D06)
controlleddocs → authorization   (explicit)
```

and forbids `application` from "reimplement[ing] semantic rules" (D05).

**The contradiction.** The `+` in T3 §2 is itself a ratified semantic rule — the ALLOW/DENY composition, including *which* predicate is required for *which* permission (T3 §9). Under the candidate as written:

- if `application` performs the composition, it encodes T3 §9 ⇒ violates D05 and creates a **second authorization authority**;
- if `authorization` performs it, it needs Controlled Documents predicates ⇒ violates D06;
- if `controlleddocs` performs it, it needs grants ⇒ violates D06.

All three doors are closed. The topology as written cannot realize the ratified authorization equation.

**Root cause.** The candidate treats "no owner→owner import" as sufficient to define cross-owner composition. It is not: forbidding the edge without naming the substitute leaves the composition homeless.

**Target invariant.**

> The ALLOW/DENY composition rule lives in Authorization. Authorization receives domain predicate *facts* as inputs. It never reaches into another owner to obtain them, and no other package re-implements the composition.

**Fix.** Name the seam class and direction (exact contract stays T8-C, per registry amendment §8):

```text
Authorization exposes a decision surface that accepts:
    actor identity + eligibility
    requested operation/permission
    requested scope
    a bounded, owner-supplied set of domain predicate FACTS

application GATHERS the facts from the owning owner's public surface and PASSES them

ALLOWED    application → authorization decision surface
ALLOWED    application → owner public predicate/fact surface
FORBIDDEN  authorization → any other semantic owner
FORBIDDEN  application computing ALLOW/DENY itself
FORBIDDEN  any second evaluation of the T3 §2 equation anywhere in the tree
```

This preserves D06 exactly, keeps the decision rule in its ratified home, and keeps `application` as pure choreography (gather → ask → act).

**Second-order requirement (do not lose this).** T6 §26 "Every command rechecks current canonical truth" plus T3 §11's eligibility serialization mean the decision surface must be callable **inside** the transaction scope from `B-02`. The decision surface therefore takes the transaction scope as a parameter. `B-02` and `B-04` must be fixed together or neither is fixed.

**Adversarial challenge.** Is this over-abstraction? No: it adds one interface that already had to exist somewhere, and it removes a duplicate-authority defect class rather than moving it. Would direct `authorization → controlleddocs` be simpler? Locally yes; globally no — that is Alternative A, and it reintroduces reciprocal owner edges (Controlled Documents needs Authorization on every command; Authorization would need Controlled Documents on every predicate), i.e. a guaranteed 2-cycle between the two largest owners.

**Decision.** `D06` ACCEPT WITH FIX. New decision required (proposed `T8B-D26`).

---

## 4. MAJOR findings

### M-01 — The `platform/` law is stated in terms of SQL; the proven leak vector is semantic types

**Evidence.** The candidate's section F forbids `platform/postgres` from containing Organization/Authorization/Controlled Documents/Audit SQL. But at this HEAD the actual platform→owner leaks contain **no SQL at all** — they are type and application-service imports:

```text
internal/platform/authn/config.go:12   modules/auth/application
internal/platform/authn/config.go:13   modules/iam/domain
internal/platform/authn/context.go:8   modules/iam/domain
internal/platform/docgenv2/…:9-10      modules/documents/{application,domain}
```

**Root cause.** A prohibition aimed at the previous system's most visible symptom (foreign SQL) rather than at the structural condition (a mechanism package naming a semantic type).

**Fix.** State the platform law over **surface**, not storage:

```text
no package under platform/ may import ANY semantic owner package
no platform signature/struct/field may name a semantic owner type or a
    semantic identifier type
platform speaks primitives, mechanism types, and consumer-owned port types only
platform owns no semantic SQL and no cross-owner query
```

**Enforcement.** This is a strict, cheap, complete import-direction rule — `platformboundary.go:11,60` already implements exactly this shape and needs only a retarget plus an **empty** allowlist (see `M-05`).

---

### M-02 — `platform/identityprovider` collides with Authentication's ratified ownership of the IdP anti-corruption boundary

**Evidence.** Ownership topology §2 assigns to **Authentication**:

```text
provider-subject binding
application Session
Session lifecycle / revocation
authentication-assurance / fresh-auth evidence when required
IdP anti-corruption boundary          ← owner-owned, explicitly
```

and states "Provider roles/groups/organizations/permissions never become canonical MetalDocs Authorization." T6 §2.2 places `/auth/login` and `/auth/callback` outside the `/api/v1` product census — they are transport integration routes.

The candidate places `identityprovider` under `platform/`. Current evidence shows this exact seam already leaking in the opposite direction: `internal/platform/authn` holds config and request-context code that imports two owners' types.

**Root cause.** "Talks to an external provider" was treated as sufficient to classify the whole concern as mechanism. Anti-corruption *translation* is semantic; the protocol client is not.

**Fix.**

```text
platform/identityprovider   = protocol/transport client mechanism only
                              (discovery, token exchange, JWKS, raw claims)
                              MUST NOT map claims to MetalDocs identity
authentication (owner)      = anti-corruption translation, provider-subject binding,
                              Session issuance/revocation, assurance evidence
transport/http              = /auth/login, /auth/callback browser routes
```

Add explicitly to the forbidden list: *provider claims/roles/groups may not appear in any package outside `authentication`.*

---

### M-03 — "One package per owner" freezes the wrong unit; freeze **one public surface** per owner instead

**Evidence.**

- Bootstrap §4: "A semantic owner may realize through multiple cohesive packages. Package count follows isolation and clarity, not owner count or legacy module count."
- Handoff: "A semantic owner does not imply exactly one Go package."
- Ownership topology §5: "`Controlled Documents` is one **semantic authority**, not one giant aggregate, file, package or transaction."
- Post-T6 program §2 corollary 3: "A Writer **may** choose local tactics inside accepted realization."
- T1 §1 enumerates ~14 Controlled Documents state families; T2 defines distinct transaction laws for create, submit, governance, withdraw, cancel, release, obsolescence, configuration.
- Legacy scope, `CURRENT-STATE ONLY`: the packages that map to Controlled Documents concerns total ≈ 68k non-test LOC across five legacy modules.

The candidate's D03/D04 make one package the default and require **evidence** ("real dependency boundary, material test/build isolation benefit, unavoidable cyclic pressure or comprehension/change-isolation failure") before an owner may split.

**Root cause.** The candidate is defending against a real defect — premature *peer semantic owners* — but it defends by constraining package **count**, which is an accidental property, instead of constraining **public surface count**, which is the essential one. Go's `internal` mechanism makes owner-private package count invisible to the dependency graph: splitting or merging owner-private packages cannot violate any T8-B invariant.

**Consequences of leaving D03/D04 as written.**

1. an architecture-gate is imposed on a decision the program already delegates to Writers ⇒ accidental process complexity;
2. the "evidence" burden will be discharged immediately and repeatedly for Controlled Documents, so the default is a default in name only;
3. worse, it creates pressure to keep one enormous package to avoid the gate — the God-package outcome the candidate is trying to prevent.

**Fix.**

```text
D03'  each semantic owner exposes EXACTLY ONE importable public package path
D04'  owner-private structure is unconstrained: files or nested
      internal/<owner>/internal/<responsibility> packages, at the owner's discretion,
      with NO evidence gate
D04'' adding a SECOND public package path for an owner is an architecture decision
      and requires a named cross-owner/transport consumer that cannot be served
      by the single surface
```

What this preserves: the anti-fragmentation invariant (no peer semantic owners; §G unchanged) and the enforcement story (Go `internal` + import policy operate on public paths). What it removes: a gate that protects nothing.

**Note on the God-package attack (mandatory surface 3).** With `D03'`, "does `controlleddocs` become a God package?" is correctly reframed: it becomes one *authority* with an internal structure, which is exactly what ownership topology §5 ratified. The God-package risk that remains is **surface** God-ness — one public package exporting 200 methods — and that is what `M-04`'s lens scoping and `B-03`'s single inbound door actually bound.

---

### M-04 — A single `application/` package is an unguarded God-orchestrator seat

**Evidence.** T6 §1 and §5 ratify the lens vocabulary (`Library`, `My Work`, `Document Official`, `Document Work`, `Governance Case`, `Document History`, `Audit`, `Administration`); T6 §28 already binds the *frontend* feature vocabulary to those same lenses. T6 §29 freezes 76 operations, all of which would land in one package under the candidate as written.

The candidate's D05 prohibitions are all **negative** ("may not own state", "may not reimplement rules", "may not become a generic workflow/event bus") and none are mechanically checkable as stated.

**Root cause.** The candidate defines `application` by what it must not become, with no structural property that makes becoming it hard.

**Fix.** Give the layer structure from ratified authority rather than from invention:

```text
internal/application/<lens>          one package per ratified T6 lens, e.g.
    library/ documentwork/ governancecase/ history/ audit/ administration/ session/

FORBIDDEN  application/<a> → application/<b>      (no intra-layer layering)
FORBIDDEN  any application package holding persistent or process-global state
FORBIDDEN  any semantic owner importing any application package
```

`no application → application` is trivially machine-checkable and is what actually prevents a "core orchestration" package from accreting. This is not invention: T6 already ratified the lens set, so T8-B is consuming upstream authority, not creating it.

**Adversarial challenge.** Does lens-scoping recreate the "semantic lens becomes an owner" defect T6 §1 warns about ("Frontend/API surfaces are semantic lenses over those owners, never new semantic owners")? Only if a lens package holds state — which the second rule forbids and which is checkable.

---

### M-05 — The enforcement concept is rejection-only: no completeness direction, no stale-exception decay

**Evidence — this repository has already paid for both defects.**

1. Completeness: `tools/cilint/internal/analyzers/table_ownership.go:23-30` records that ownership parity "walks catalog → docs, and that direction cannot see an omission: 19 of the 56 live base tables were simply absent, and because `hgViolation` treats an unknown table as a non-violation, every one of them was unguarded. **An omission is not a classification.**"
2. Default-in vs default-out: `tools/verify/registry.go:41-45` — "a NEW check is in `release` by default. Opting out has to be a deliberate line in this map, which is reviewable; forgetting to opt IN is silent, and **silence is how coverage rots**."
3. Stale exceptions: `platformboundary.go:19` still exempts `internal/platform/objectstore`, which imports **zero** module packages at this HEAD.

The candidate's proof concept is six `reject …` clauses. Every one is a rejection of a *known* edge shape; none establishes that every package is classified.

**Root cause.** Rejection lists are open-world. A new, unclassified package is unguarded by construction.

**Fix.** Add three obligations to `D20`:

```text
COMPLETENESS   every Go package in the target tree maps to exactly ONE architecture
               class (owner / owner-private / application / transport / platform /
               composition / tooling); an unmapped package FAILS the check
BIDIRECTIONAL  the class catalog is proven against the live package universe in BOTH
               directions (catalog → tree and tree → catalog)
DECAY          every exception carries a removal trigger and FAILS when its
               violation no longer exists (the repository already runs expiry files
               for eslint and dead-code baselines; the architecture allowlist has none)
```

Plus the negative-fixture rule already learned here: each rejection clause needs a fixture that is proven RED by restoring the pre-fix shape, not merely added to an already-failing tree.

---

### M-06 — Enforcement host is misnamed: analyzers live in `tools/cilint`, `tools/verify` is the registry

**Evidence.** `tools/verify/registry.go:1-9` describes itself as the check registry — "a check exists here or it does not exist". The architecture analyzers (`hgcrossmodule`, `platformboundary`, `nosqltxindomain`, `postcommitaudit`, `table_ownership`, …) live under `tools/cilint/internal/analyzers/`. The candidate's tree contains `tools/verify` and no analyzer host; its proof concept attributes the import-graph policy to `tools/verify`.

Additionally, the T8-A registry amendment §4 lists "legacy target-specific architecture guards" as **not inherited**, so any reuse must pass the five-part selective-reuse gate.

**Fix.** T8-B should freeze the **property** and the **placement class**, and either decide the host or defer it with a named owner:

```text
PRESERVED PROPERTY   one verification registry is the SSOT for what "verified" means
                     (T8-A PRESERVE list: "verification registry / local-CI SSOT model")
TARGET               architecture policy is expressed as static analysis over the Go
                     import graph, registered in the verification registry, with
                     negative fixtures
PLACEMENT            tooling lives outside internal/ and imports no product package
HOST SELECTION       one analyzer host; whether it remains a separate tools/ program
                     or folds into the registry is a bounded T8-B choice, not a
                     Writer choice
```

**Selective-reuse assessment (mandatory attack surface 17), applied honestly to the analyzer mechanism:**

| Gate | `tools/cilint` analyzer harness | Verdict |
|---|---|---|
| named current R10 consumer | yes — D20 requires exactly this class of control | PASS |
| public contract free of legacy semantic authority | the *harness* is generic (`[]string → []Finding`); individual analyzers encode legacy REQ-IDs and the `internal/modules` prefix | PASS for harness, FAIL for policy content |
| dependency direction fits target | tooling → nothing product-side | PASS |
| proof asserts target property | current fixtures assert legacy topology | FAIL — rewrite fixtures |
| reuse smaller than rewrite | writing a `go/ast` import-policy harness from scratch is small, but the fixture/registry/allowlist/expiry plumbing is not | PASS |

**Conclusion:** the analyzer *harness and registry integration* pass the gate; the *policy content and fixtures* do not and must be rewritten. This is the concrete answer to "does any current mechanism actually pass the T8-A five-part gate" — yes, exactly one class does, at harness level only.

---

### M-07 — The candidate tree omits repository homes that T8-B is chartered to freeze

**Evidence.** T8-B explicitly freezes "target repository/package layout". The candidate tree shows `go.mod`, `api/openapi/v1/`, `cmd/`, `internal/`, `tools/verify/` and nothing else. At this HEAD the following exist and have no home in the candidate:

| Current | Candidate home | Problem |
|---|---|---|
| `cmd/gen-http-surface`, `cmd/gen-tripwire`, `cmd/problem-codes-dump` | none | candidate *repurposes* `cmd/` for runtime shells; generators are evicted with no destination |
| `apps/{api,worker,jobs,dbprovision}/cmd/…` | `cmd/<runtime-shells>` | relocation is fine, but must be stated as a decision, not implied |
| `db/{migrations,baseline,grants,prerequisites,reference-data,dev-seeds}` | none | migrations/bootstrap assets are layout; T8-D owns their *content*, not their location |
| `tests/{unit,integration,e2e,docx_v2}` | none | test-root topology interacts with the import policy's package universe (`M-05`) |
| `frontend/`, `packages/*`, `apps/docx-renderer` (Node/TS) | none | not backend, but their continued existence must be explicit or the tree reads as exhaustive |

**Root cause.** The tree is a backend-relevant projection presented in repository-root form.

**Fix.** Either (a) show the complete repository root with non-backend roots marked `OUT OF T8-B SCOPE — see T8-F / T8-G`, or (b) label the tree explicitly "backend projection; roots not shown are unchanged by T8-B". Then add explicit homes for code generators, DB assets and test roots. Contract-first generators are load-bearing: T8-A PRESERVE retains "contract-first OpenAPI + generated Go/TypeScript boundaries", so their host is not incidental.

---

### M-08 — `platform/idempotency` stores authorization-sensitive, privacy-sensitive payloads and needs an explicit law

**Evidence.** T6 §18 requires the idempotency key insert/lock and the "completed replay result sufficient for exact status/body replay" to commit with the business transaction. T6 §19 requires that every replay still pass a **live** T3 permission/scope + resource-visibility check before disclosure, and warns: "Replay storage must not become an unintended retention root for erasable UserProfile PII."

The candidate lists `platform/idempotency` with no further law.

**Root cause.** A mechanism package is being handed durable custody of response bodies whose disclosure is authorization-governed and whose contents are erasure-governed (T3 §15 `UserProfile erased`, T4 §16 restore erasure barrier).

**Fix.** State three rules:

```text
platform/idempotency stores and returns opaque payloads; it NEVER authorizes
replay disclosure authorization happens in application (same composition path as B-04)
the replay store exposes a purge/redact operation reachable from the
    UserProfile-erasure use case, so erasure is not defeated by a mechanism cache
```

Deferred correctly to T8-D: the replay record's schema and retention realization.

---

## 5. LOW findings

**L-01 — Candidate not in the repository.** See §0.2. Recommend committing the adjudicated candidate as T8-B staging so the promotion diff is auditable.

**L-02 — "One Go module" should read "one Go module for backend Go code."** The repository is already polyglot and multi-workspace: `frontend/`, six `packages/*`, and `apps/docx-renderer` (Node/TS, `package.json` + `tsconfig.json`). D01 is correct but its phrasing invites a later reader to think the repository has one build unit. Also worth recording as a constraint datum: `vendor/` is present, which raises multi-module friction and further supports D01.

**L-03 — `platform/rendering` risks conflating two mechanisms that T5/T6 deliberately separate.** T5 §2 and T6 §21: the interactive DOCX editor/viewer adapter and the server-side OfficialRendition renderer are distinct mechanisms that "may use another product". One package name suggests one mechanism. Name two seams or state that the directory hosts two unrelated adapters.

**L-04 — `platform/managedcontent` should carry T4's explicit non-semantic law.** T4 §8: "The mechanism does not parse or own Document/Revision/Submission semantics. No generic `owner_type/owner_id` registry is introduced." T4 §9's `MalwareInspector` seam is also unlisted. Cheap to state; expensive to rediscover.

**L-05 — Where the Library/Search read model lives is unstated.** D15 correctly refuses a Search owner; T5 §7's baseline is "a canonical PostgreSQL query/view over the current facts", and those facts are Controlled Documents-owned. Without a positive statement, a Writer may create `platform/search`. Under `M-04` the correct home is the `library` application lens over the Controlled Documents public read surface. Say so.

**L-06 — D10's negative form hides a positive decision.** Saying `platform/postgres` may not own owner SQL implicitly decides that persistence adapters are **owner-private**. That is correct and should be stated affirmatively, because it forecloses a T8-D option (a separate shared persistence layer) and a foreclosure should be visible.

---

## 6. Subtractive pass

Each element was attacked with: *delete it — which real invariant weakens?*

| Element | Delete? | Reasoning |
|---|---|---|
| `application/` | **NO** | T3 §10 offboarding is one transaction across 4 owners; T3 §2/§9 composition needs a non-owner site. Deleting it forces owner→owner edges = Alternative A. |
| `transport/` | **NO** | T6 §2 defines three HTTP surface classes and T5 requires job invocation; deleting it puts wire/codec concerns inside owners or application. |
| `platform/` | **NO** | T4/T5/T6 name mechanisms explicitly (ManagedContentStore, durable jobs, idempotency store). Deleting it scatters them into owners and guarantees duplication. |
| `composition/` | **NO — but its justification is not process count** | Justified by: N runtime shells (count deferred to T8-G) **and** integration tests needing full wiring without HTTP. Current evidence strengthens this: wiring today lives in `internal/platform/bootstrap`, which is a *proven* platform→module violator (`bootstrap/api.go:13-18`). Separating composition from platform is a proven correction, not symmetry. |
| separate owner packages | **NO** | Five ratified authorities; one package would erase the boundary the whole program exists to establish. |
| nested owner-private seam | **YES — delete the gate, keep the option** | See `M-03`. The *mechanism* is valuable (Go `internal` makes foreign access impossible); the *evidence gate* protects nothing. |
| import-graph verifier | **NO** | Method §Enforcement: "A control that cannot be shown to fire is not proven." Without it, every dependency law in the candidate is documentation. |
| one-Go-module assumption | **NO** | See §7 Alternative C. |
| `platform/postgres` | NO | T8-A PRESERVE: PostgreSQL substrate. |
| `platform/managedcontent` | NO | T4-D named contract. |
| `platform/identityprovider` | **PARTIAL** | Keep the protocol client; move anti-corruption to `authentication` (`M-02`). |
| `platform/rendering` | NO | T5 §2 conditional renderer has a named consumer (`RequireOfficialRendition`). |
| `platform/river` | NO | T5 §3 selects one durable-job mechanism with a named consumer. |
| `platform/idempotency` | NO | T6 §18 requires durable replay state. |
| `platform/observability` | NO | T5 §16 requires operational visibility of activated async work. |
| `platform/config` | NO | Configuration/bootstrap is named mechanism (ownership topology §7). |

**Subtractive verdict:** the candidate carries **no** deletable structure except the D04 evidence gate. It is not bloated. This is the strongest thing the review can say in the candidate's favour, and it is said on evidence.

---

## 7. Alternatives

### Alternative A — owner packages with direct owner→owner imports

**Rejected, with a specific mechanism.** Under T3 §9 every Controlled Documents command needs Authorization; under T3 §2 every Authorization decision needs a Controlled Documents predicate. That is a guaranteed 2-cycle between the two largest owners on day one, before Organization (grants need memberships) and Audit (every owner) are added. The candidate's suspicion is correct and the review confirms it independently: the cycle is derivable from ratified semantics alone, without appeal to legacy SCC counts.

### Alternative B — one-module owner-first modular monolith + thin orchestration + adapters

**The candidate.** Confirmed as the correct class; incomplete as written (§3).

### Alternative C — separate Go modules per owner

**Rejected.** T8-A §7's five-part gate applied to the *proposal*: no independent consumer, no independent release lifecycle, no repository/trust boundary, no deployment need (T8-G has not even opened). Costs are concrete: five `go.mod` files, cross-module version pinning for a shared transaction-scope type, `replace` directives or synchronized tagging on every change, and interaction with the present `vendor/` tree. And the decisive point: **module boundaries do not add the enforcement the candidate needs** — Go `internal` already gives owner-private inaccessibility inside one module, and cycle prohibition still requires the import-graph verifier either way. Multi-module buys nothing that `internal` + the verifier does not already buy, at materially higher cost.

*What would falsify this:* a real external consumer of one owner, a separate release train, or a trust boundary requiring independent build provenance. None exists at this HEAD.

### Alternative D — **the strongest alternative found: B corrected** (recommended)

The mandated search for a materially better fourth alternative produced **no different topology class**. It produced a materially better *realization* of B:

```text
D = B
  + one PUBLIC SURFACE per owner (not one package)          [M-03]
  + application scoped by ratified T6 lens, no intra-layer imports, stateless  [M-04]
  + three named seams that make the dependency graph decidable:
        transaction scope       (platform-owned, explicit parameter)   [B-02]
        evidence sink           (consumer-owned port per owner)        [B-01]
        authorization decision  (fact-carrying, Authorization-owned)   [B-04]
  + single inbound door: transport → application only                  [B-03]
  + platform law over TYPES, not SQL                                   [M-01]
  + enforcement with completeness + decay, not rejection-only          [M-05]
```

**Honesty note required by the Method (record what would have falsified the conclusion).** Two further alternatives were constructed and tested, and both were rejected on evidence rather than on preference:

- **D-alt-1 — use cases hosted inside their primary owner**, with a thin cross-owner layer only for genuinely multi-owner writes. Attractive because Launch has few multi-owner *writes* (chiefly T3 §10 offboarding). Rejected: it produces two inbound doors and a "sometimes owner, sometimes application" rule that is not mechanically checkable, and it puts orchestration inside the authority it orchestrates. The saved hop does not pay for an unverifiable rule.
- **D-alt-2 — a dedicated `policy` package hosting the T3 §2 composition**, separate from `authorization`. Rejected: it creates a sixth semantic home for a rule T3 already assigns to Authorization — a textbook duplicate-authority defect under Method §Authority and mechanism.

Had either survived, this review would have reported a different topology. Neither did.

---

## 8. Authority-duplication pass

| Pair | Verdict | Note |
|---|---|---|
| `application` vs owners | **AT RISK → fixed by B-04 + M-04** | Unfixed, `application` acquires the T3 §2 rule and becomes a second authorization authority. |
| Authorization vs Controlled Documents predicates | **CLEAN once B-04 lands** | Ownership topology §4 already partitions: grants vs relationship/lifecycle meaning. |
| Audit vs domain history | **CLEAN** | T6 §27 and T3 §12 keep History (Controlled Documents facts) separate from Audit (action evidence); the candidate does not blur them. |
| `platform` vs owners | **AT RISK → fixed by M-01 + M-02** | Proven current leak vector is types, not SQL. |
| transport/read models vs canonical state | **CLEAN** | T6 §26: read models are not mutation/current-state authority; the candidate does not create a read-model owner. |
| `composition` vs `application` | **CLEAN, with one guard needed** | D08 is correct. Add: composition may contain no conditional business logic and no request-scoped behaviour — otherwise "wiring" becomes a startup-time orchestrator. |
| `allowed_actions` vs command authorization | **AT RISK, T6-mandated** | T6 §26 requires a provably shared derivation; the B-04 decision surface is the natural single source. Recommend T8-B state that `allowed_actions` derives from the same decision surface. |

---

## 9. Failure-class pass

| Failure class | Under candidate as written | Under corrected candidate |
|---|---|---|
| owner import cycles | prevented by D06 | prevented + verified (SCC check) |
| hidden cross-owner write authority | **reachable** — no tx carrier named (B-02) | prevented: scope is an explicit parameter |
| God application layer | **reachable** — one package, negative rules only | bounded by lens packages + no intra-layer imports |
| God platform layer | **reachable via types** (M-01) | prevented by type-level import law |
| foreign SQL reintroduction | prevented (F) | prevented + verified |
| transport bypass | **ambiguous** (B-03) | prevented by single inbound door |
| semantic package fragmentation | prevented (G) | prevented (G unchanged) |
| semantic package over-concentration | **weakly addressed** (M-03) | addressed by surface-level scoping |
| interface explosion | **at risk** — a per-owner port per mechanism can multiply | bounded: ports are consumer-declared and minimal; only 3 seam classes are architectural |
| adapter leakage | **reachable** (M-02) | prevented by owner-owned anti-corruption |
| unprovable dependency policy | **present** — rejection-only (M-05) | closed by completeness + decay + fixtures |
| future retrofit requiring authority rewrite | not reachable | not reachable (see §10) |
| Writer discretion left by incomplete topology | **present** — B-01..B-04, M-07 | closed |
| missing same-commit audit | **reachable** (B-01) | prevented by in-tx evidence port + retargeted analyzer |

---

## 10. Future-evolution verification (mandatory surface 19)

Tested against the named horizon in ownership topology §9 and T6 §30. Question asked for each: *does attaching this require dismantling the T8-B topology, or only adding to it?*

| Capability | Attachment under corrected candidate | Dismantling required? |
|---|---|---|
| Distribution / Read & Acknowledge | new owner package + new application lens; consumes Release facts via Controlled Documents public surface | **No** |
| Periodic Review | new owner or CD-internal responsibility + lens; reads current EFFECTIVE Revision | **No** |
| Evidence / quality records | new owner package; reuses `platform/managedcontent` and the exact-content descriptor | **No** |
| Records / Retention / Legal Hold | new owner + policy attachment to stable identities; storage stays mechanism | **No** |
| Governed Export | new application lens reading owner public surfaces + descriptors | **No** |
| Change Control | new owner orchestrating stable Document/Revision identities via application | **No** |
| pooled multi-customer tenancy | **partially — flagged** | see below |
| CRDT / realtime coauthoring | replaces the DRAFT concurrency mechanism inside Controlled Documents' private structure; Submission identity untouched | **No** — and `M-03`'s owner-private freedom makes this *cheaper* |

**Pooled tenancy — the one honest caveat.** T1/T3/T6 keep tenancy reopenable around a stable Company identity, and T6 §31 removes the "Tenant/RLS universal product ontology" from Launch. A pooled-tenancy retrofit would require a tenant scope to reach **every** owner and every mechanism. Under the corrected candidate that is additive in exactly one place — the transaction scope from `B-02` is the natural carrier for a tenant scope. This is a further argument for naming the scope now rather than deferring it: the cheapest future attachment point for tenancy is a seam T8-B is currently declining to name. Recorded as evidence for `B-02`, **not** as a reason to build tenancy machinery now (that would violate ownership topology §10.5).

---

## 11. Mandatory attack-surface disposition

| # | Surface | Disposition |
|---|---|---|
| 1 | one Go module genuinely Global Maximum? | **YES** — Alternative C rejected on the five-part gate; `internal` + verifier already provide what modules would (§7) |
| 2 | one package per owner too coarse / appropriately minimal? | **WRONG UNIT** — freeze public surfaces (`M-03`) |
| 3 | does `controlleddocs` become a God package? | **Not as an authority** (ratified §5); God-**surface** risk is real and is bounded by `B-03` + `M-04` |
| 4 | would splitting CD resurrect false ownership? | **YES if split into peers** — G is correct and must not be relaxed; owner-*private* splitting carries no such risk |
| 5 | does `application` become a hidden sixth domain? | **AT RISK as written** — `M-04` |
| 6 | zero owner→owner justified or dogmatic? | **JUSTIFIED**, but unrealizable without `B-04` |
| 7 | cross-owner local ACID realizable cleanly? | **NOT DECIDABLE as written** — `B-02` |
| 8 | same-commit Audit fits without Audit becoming authority? | **NO as written** — `B-01`; fixed by consumer-owned in-tx port |
| 9 | does Authorization composition demand a different direction? | **NO** — fact-passing preserves direction (`B-04`); direct import would be Alternative A |
| 10 | does `platform/` become the old God `common`? | **AT RISK via types** — `M-01`, `M-02`; proven leaks at `platform/authn`, `platform/docgenv2`, `platform/bootstrap` |
| 11 | consumer- vs adapter-owned interfaces on the right side? | **UNSTATED** — `B-01`/`B-02` require consumer-owned ports; state the rule |
| 12 | `transport → application` sufficient for T6 without leaking authority? | **YES**, once the ambiguity is removed (`B-03`) |
| 13 | job invocation under transport or elsewhere? | **CONFIRMED as transport** — T5 §1 makes jobs execution mechanism; a job handler is an inbound adapter, symmetric with HTTP. Add: job handlers may not embed semantic rules and reload canonical state (T5 §14) |
| 14 | does `composition` stay wiring-only? | **YES with one guard** — forbid conditional business logic / request-scoped behaviour in composition (§8) |
| 15 | overfit to Go conventions vs product invariants? | **NO** — every class re-derives from T1→T8-A (§2). Go `internal` is used as *enforcement*, not as the organizing principle |
| 16 | unnecessary abstraction/interface ceremony? | **BOUNDED** — three architectural seam classes only; `M-03` removes the one gratuitous gate |
| 17 | does any current mechanism pass the T8-A five-part gate? | **EXACTLY ONE CLASS** — the analyzer harness + verification registry integration, at harness level; policy content and fixtures FAIL and must be rewritten (`M-06`) |
| 18 | does the candidate delete a legacy package with a live R10 consumer? | **NO deletion is authorized by T8-B** (T10 owns sequencing). Checked the highest-risk cases: `internal/platform/bootstrap` (composition wiring → survives as concept, contents rewritten), `tools/cilint` (harness survives, policy rewritten), `api/openapi/v1` (T8-A PRESERVE contract-first → survives, content is T8-E), `cmd/gen-*` generators (**no home in the candidate tree — `M-07`**), `db/` (**no home — `M-07`**) |
| 19 | future capabilities attach additively? | **YES** for all named capabilities; pooled tenancy strengthens `B-02` (§10) |
| 20 | can the enforcement actually fire, and is it strong enough? | **CONCEPT YES, STRENGTH NO** — rejection-only, no completeness, no decay, host misnamed (`M-05`, `M-06`) |
| 21 | can the rules be simpler and still enforce every invariant? | **YES, marginally** — `M-03` deletes one gate; otherwise the rule set is already near-minimal |
| 22 | does concrete toolchain evidence favour multi-module or another source topology? | **NO** — `vendor/` present, one `go.mod`, polyglot workspaces already separate non-Go code (§7, `L-02`) |
| 23 | has T8-B accidentally decided T8-C→T8-G / T10? | **NO material trespass.** D21–D24 hold the line; naming the three seam *classes* is explicitly authorized by the T8-A registry amendment §8. Two near-misses to keep bounded: (a) `B-02` must not choose isolation levels or locks (T8-D); (b) `M-04`'s lens names must be read as ratified-T6 vocabulary, not as frontend feature realization (T8-F) |
| 24 | any material T8-B decision still missing? | **YES — five:** transaction-scope carrier (`B-02`), evidence-append seam (`B-01`), authorization decision-input seam (`B-04`), inbound-door law (`B-03`), repository homes for generators/DB assets/test roots (`M-07`) |

---

## 12. Candidate decision disposition

| ID | Decision | Disposition | Basis |
|---|---|---|---|
| **T8B-D01** | one Go application module | **ACCEPT** (wording: "for backend Go code" — `L-02`) | §7 Alt C; `go.mod:1`; vendor/polyglot evidence |
| **T8B-D02** | owner roots = authentication / organization / authorization / controlleddocs / audit | **ACCEPT** | ownership topology §1; T6 §1; Structural Inversion §2 |
| **T8B-D03** | one package per owner by default | **ACCEPT WITH FIX** → one **public surface** per owner | `M-03`; bootstrap §4; post-T6 §2 corollary 3 |
| **T8B-D04** | owner-private nested packages only on demonstrated need | **ACCEPT WITH FIX** → keep the mechanism, **delete the evidence gate**; gate only *additional public* surfaces | `M-03` |
| **T8B-D05** | application = non-semantic cross-owner orchestration | **ACCEPT WITH FIX** → lens-scoped, stateless, no intra-layer imports; must not append owner-intrinsic audit; must not compute ALLOW/DENY | `M-04`, `B-01`, `B-04` |
| **T8B-D06** | direct owner→owner imports forbidden by default | **ACCEPT WITH FIX** → keep the prohibition; add the substitute seams (evidence port, decision surface) that make it realizable | `B-01`, `B-04` |
| **T8B-D07** | transport depends on application, not owner-private | **ACCEPT WITH FIX** → transport depends on application **only**; owner *public* surfaces are also closed to transport | `B-03` |
| **T8B-D08** | composition owns construction/wiring only | **ACCEPT** (add guard: no conditional business logic, no request-scoped behaviour) | §8; `bootstrap/api.go:13-18` proves the failure mode |
| **T8B-D09** | platform contains technical mechanisms only | **ACCEPT WITH FIX** → law stated over types/imports, not SQL | `M-01`, `M-02` |
| **T8B-D10** | `platform/postgres` cannot own semantic SQL | **ACCEPT** (state the positive corollary: persistence adapters are owner-private — `L-06`) | T8-A §4; `M-01` |
| **T8B-D11** | no generic common/shared dumping ground | **ACCEPT** | Method §Authority and mechanism; GCR outcome |
| **T8B-D12** | no Launch peer package for Approval | **ACCEPT** | GCR A3/A4; ownership topology §5; T6 §31 |
| **T8B-D13** | no Launch peer package for Templates | **ACCEPT** | T1 §2 "Template = ordinary governed Document role"; T6 §31 |
| **T8B-D14** | no Launch peer package for Artifact | **ACCEPT** | GCR A1; T4 §2; T1 §5 |
| **T8B-D15** | no Launch semantic owner for Search | **ACCEPT** (add where the Library read model lives — `L-05`) | T5 §7; ownership topology §7 |
| **T8B-D16** | no dormant Launch+/Future packages | **ACCEPT** | ownership topology §10.5; T6 §30 |
| **T8B-D17** | legacy topology gets no inheritance entitlement | **ACCEPT** | T8-A §1/§4; registry amendment §2 |
| **T8B-D18** | selective reuse subject to the five-part gate | **ACCEPT** (applied in `M-06`: exactly one mechanism class passes, at harness level) | T8-A §7 |
| **T8B-D19** | Go `internal` participates in owner-private enforcement | **ACCEPT** — mechanism verified correct: `internal/<owner>/internal/x` is importable only from within `internal/<owner>/…`, so application, transport, composition and foreign owners are all structurally excluded | Go import rules; no such nesting exists today (`find internal -mindepth 1 -type d -name internal` → none) |
| **T8B-D20** | `tools/verify` import-graph + negative fixtures required | **ACCEPT WITH FIX** → add completeness (both directions), exception decay, RED-proven fixtures; correct the host attribution | `M-05`, `M-06`; `table_ownership.go:23-30`; `registry.go:41-45`; `platformboundary.go:19` |
| **T8B-D21** | exact inter-owner contracts deferred to T8-C | **ACCEPT** — with the explicit note that naming seam *classes/directions* is T8-B work per registry amendment §8, and is not a T8-C trespass | registry amendment §8 |
| **T8B-D22** | exact persistence deferred to T8-D | **ACCEPT** | post-T6 §6 |
| **T8B-D23** | exact runtime/process topology deferred to T8-G | **ACCEPT** — `cmd/<runtime-shells>` with deferred count is the correct shape | post-T6 §6 |
| **T8B-D24** | transition/deletion sequencing deferred to T10 | **ACCEPT** | T7 §4; post-T6 §8 |
| **T8B-D25** | *(new — required)* transaction-scope carrier: platform-owned abstraction, opened by application, explicit parameter on owner write surfaces | **REQUIRED ADDITION** | `B-02` |
| **T8B-D26** | *(new — required)* authorization decision surface accepts owner-supplied predicate facts; no other package evaluates the T3 §2 equation | **REQUIRED ADDITION** | `B-04` |
| **T8B-D27** | *(new — required)* evidence-append is a consumer-owned port implemented by `audit`, invoked in-transaction at the mutation site | **REQUIRED ADDITION** | `B-01` |
| **T8B-D28** | *(new — required)* repository homes for code generators, DB assets and test roots; tree scope stated | **REQUIRED ADDITION** | `M-07` |

No decision was invented to fill the matrix: D25–D28 each close a specific BLOCKER/MAJOR that would otherwise reach a Writer.

---

## 13. Corrected T8-B candidate

This is the exact shape this reviewer would approve. Deltas from the reviewed candidate are marked `[FIX]` / `[NEW]`.

### 13.1 Repository/module posture

```text
module metaldocs          one Go module for all backend Go code            [FIX: scope]

Non-Go workspaces (frontend/, packages/*, apps/docx-renderer) are NOT part of this
module and are NOT decided by T8-B.                                        [FIX]
Multiple Go modules require a proven independent consumer, release lifecycle,
repository/trust boundary or deployment need. None exists.
```

### 13.2 Target tree (backend projection; unlisted repository roots are unchanged by T8-B) `[FIX]`

```text
MetalDocs/
├── go.mod
│
├── api/
│   └── openapi/v1/openapi.yaml            contract SSOT (content = T8-E)
│
├── cmd/
│   ├── <runtime-shells>                   count/topology = T8-G
│   └── <contract/codegen tools>           explicit home for generators        [NEW]
│
├── db/                                    migrations + bootstrap assets       [NEW]
│                                          (content/schema = T8-D)
│
├── internal/
│   ├── authentication/                    semantic owner — ONE public surface [FIX]
│   ├── organization/                      semantic owner — ONE public surface
│   ├── authorization/                     semantic owner — ONE public surface
│   ├── controlleddocs/                    semantic owner — ONE public surface
│   ├── audit/                             supporting semantic owner
│   │       each MAY use internal/<owner>/internal/<responsibility> freely,
│   │       with no evidence gate                                              [FIX]
│   │
│   ├── application/                       cross-owner orchestration, lens-scoped [FIX]
│   │   ├── library/  documentwork/  governancecase/  history/
│   │   └── audit/    administration/  session/
│   │
│   ├── transport/
│   │   ├── http/                          incl. /auth/login, /auth/callback   [FIX]
│   │   └── jobs/                          durable-job inbound adapters
│   │
│   ├── platform/                          non-semantic mechanisms only
│   │   ├── postgres/                      pool, tx mechanics, bootstrap primitives
│   │   ├── txscope/                       transaction-scope abstraction       [NEW]
│   │   ├── managedcontent/                incl. malware-inspector seam        [FIX]
│   │   ├── identityprovider/              PROTOCOL CLIENT ONLY                [FIX]
│   │   ├── rendering/                     two distinct adapters, no shared semantics [FIX]
│   │   ├── river/
│   │   ├── idempotency/                   opaque payloads; never authorizes   [FIX]
│   │   ├── observability/
│   │   └── config/
│   │
│   └── composition/                       construction/wiring only
│
├── tests/                                 test roots                          [NEW]
│
└── tools/
    ├── verify/                            verification registry / SSOT
    └── <architecture policy analyzers>    host selection is a bounded T8-B choice [FIX]
```

Names below mechanism level remain refinable. Architecture **classes** and **directions** are frozen.

### 13.3 Owner granularity law `[FIX]`

```text
each semantic owner exposes EXACTLY ONE importable public package path
owner-private structure is unconstrained and requires no justification
a SECOND public package path for an owner is an architecture decision requiring a
    named consumer that the single surface cannot serve
entity names never justify a package
peer semantic owners are never created for approval / templates / artifact / search /
    distribution / periodicreview / taxonomy / notifications / interchange /
    workflow / records
```

### 13.4 Application law `[FIX]`

```text
MAY   coordinate one application/business operation
      open, commit and roll back the local transaction scope
      invoke public surfaces of semantic owners
      gather domain predicate facts and pass them to the Authorization decision surface
      coordinate justified idempotency/durable intent
      return use-case results / read models

MUST NOT
      own User/Role/Permission/Document lifecycle state
      reimplement semantic rules
      evaluate the T3 §2 ALLOW/DENY composition itself
      append evidence for an owner-intrinsic T3 §15 event on the owner's behalf
      become current-state authority
      contain owner semantic SQL
      become a generic workflow/event-bus/domain platform
      import another application package
      hold persistent or process-global state

application = choreography      owners = authority
```

### 13.5 Dependency graph `[FIX]` `[NEW]`

```text
ALLOWED
  cmd/*            → composition
  transport        → application
  application      → owner public surfaces
  application      → authorization decision surface
  application      → platform/txscope, platform/idempotency
  owner            → its own private implementation
  owner            → its own consumer-declared ports
                        (evidence sink, durable-intent sink, content mechanism, …)
  owner            → platform/txscope abstraction                        [NEW]
  platform adapter → non-semantic mechanism seams
  composition      → everything it constructs

FORBIDDEN
  transport        → any semantic owner package (public or private)      [FIX]
  transport        → SQL / persistence
  transport        → opening a transaction scope                         [NEW]
  owner            → another owner (public or private)
  owner            → internal/audit
  owner            → composition
  owner            → application
  owner            → connection pool / driver / concrete provider type   [NEW]
  owner            → opening/committing a transaction scope              [NEW]
  application      → application
  application      → owner-private implementation
  platform         → any semantic owner package or type                  [FIX]
  provider claims/roles/groups outside `authentication`                  [NEW]
  foreign SQL as owner communication
  shared write authority hidden in platform/common code
  any second evaluation of the T3 §2 authorization equation              [NEW]
  evidence append outside the mutating transaction                       [NEW]
```

### 13.6 The three named seams `[NEW]`

```text
TRANSACTION SCOPE
  home        platform/txscope (mechanism)
  opened by   application only
  passed      explicitly, as a parameter, to every participant
  forbids     ambient/implicit scope; provider types in owner signatures
  defers      isolation level, lock primitive, serialization mapping → T8-D

EVIDENCE SINK
  declared by each owner (consumer-owned port)
  implemented by internal/audit
  injected by composition
  invoked in-transaction at the mutation site
  defers exact event schema → T8-C/T8-D

AUTHORIZATION DECISION SURFACE
  home        authorization (owner)
  accepts     actor + eligibility + operation + scope + owner-supplied predicate facts
              + transaction scope
  returns     ALLOW / DENY
  also the single source for T6 §26 `allowed_actions`
  defers      exact fact vocabulary and contract → T8-C
```

### 13.7 Enforcement `[FIX]`

```text
Go internal visibility
+ import-graph policy registered in the verification registry
+ negative fixtures proven RED against the pre-fix shape

REJECT   forbidden import edges
REJECT   semantic-owner SCCs
REJECT   foreign owner-private reachability
REJECT   transport bypass of application
REJECT   owner dependency on composition / concrete adapters
REJECT   prohibited semantic peer roots
REJECT   platform naming any semantic owner type
REJECT   application → application
REJECT   evidence append outside a transaction scope
REJECT   any unmapped Go package                                          [NEW]

COMPLETENESS   every Go package maps to exactly one architecture class,
               proven in BOTH directions against the live package universe [NEW]
DECAY          every exception carries a removal trigger and FAILS when its
               violation no longer exists                                  [NEW]
```

### 13.8 Legacy disposition (unchanged from the candidate, confirmed)

```text
legacy semantic module topology       REWRITE / REHOME
false peer semantic modules           DELETE / REHOME
composition concept                   concept survives; contents not inherited
current platform packages             selective reuse via the T8-A five-part gate
current cmd/apps roots                CURRENT-STATE ONLY; process topology = T8-G
architecture guards                   harness may be reused; POLICY AND FIXTURES REWRITTEN
tools/verify control-plane property   PRESERVE
one Go module                         independently reselected
```

---

## 14. Successor-stage obligations discovered

Recorded so they are not lost; **no successor stage is designed here.**

| Stage | Obligation surfaced by this review |
|---|---|
| **T8-C** | exact contract of the three seams (evidence event shape, decision-surface fact vocabulary, transaction-scope method set); the owner-supplied predicate fact set per T3 §9 permission |
| **T8-C** | the durable-intent enqueue contract must accept the transaction scope (T5 §5) |
| **T8-D** | isolation level, Document-serialization primitive (T2 §3), idempotency replay-record schema and its purge/redaction path (`M-08`), and whether replay records live in an owner or mechanism schema |
| **T8-E** | `allowed_actions` must be generated/derived from the single decision surface, not hand-maintained (T6 §26) |
| **T8-F** | frontend feature vocabulary already tracks the same T6 lenses (T6 §28); confirm no accidental 1:1 coupling to Go package names, which T6 §28 explicitly disclaims |
| **T8-G** | runtime shell count; whether `composition` collapses into `cmd/` if exactly one shell is proven; operations surface placement (T6 §2.3) |
| **T8-H** | verify that the three seams remain coherent across persistence + wire + runtime |
| **T9** | Golden Flow proof that a governed mutation without in-tx evidence is impossible; that no second authorization evaluation exists; that a cross-owner offboarding commits atomically |
| **T10** | generator relocation (`cmd/gen-*`), `apps/*/cmd` → `cmd/` move, `db/` retention, deletion of the 15-module tree, retirement of the legacy analyzer policy and its stale allowlist entries |

---

## 15. Upstream reopen assessment

```text
Product Contract REV001      NO REOPEN REQUIRED
Whole-Product GCR A1–A10     NO REOPEN REQUIRED
4+1 ownership topology       NO REOPEN REQUIRED
T1                           NO REOPEN REQUIRED
T2                           NO REOPEN REQUIRED
T3                           NO REOPEN REQUIRED
T3-D4 amendment              NO REOPEN REQUIRED
T4                           NO REOPEN REQUIRED
T5                           NO REOPEN REQUIRED
T6                           NO REOPEN REQUIRED
T7                           NO REOPEN REQUIRED
T8-A                         NO REOPEN REQUIRED
Decision Registry + amendments  NO REOPEN REQUIRED
```

Every BLOCKER and MAJOR in this review is a **defect against existing authority**, correctable inside T8-B. None is a proposal that creates new requirement authority. Per Method §"Adversarial challenge and findings", that classification was made deliberately before writing each finding.

---

## 16. Reopen triggers for this review's conclusions

Reopen the corrected candidate only on material evidence that:

- a semantic owner acquires an external consumer, independent release lifecycle or trust boundary (⇒ reconsider D01);
- the single inbound door demonstrably cannot serve a ratified T6 read without a pass-through that adds no authorization value (⇒ reconsider `B-03`);
- an owner's single public surface cannot express a cross-owner need without a second public path (⇒ `D04''` fires as designed, not as a reopen);
- the fact-passing authorization seam proves unable to express a ratified T3 §9 predicate without leaking owner internals (⇒ reconsider `B-04`);
- T8-D proves that an explicit transaction-scope parameter cannot express a required serialization primitive (⇒ reconsider `B-02`'s carrier shape, not its existence);
- T8-G proves exactly one runtime shell **and** integration wiring needs no shared assembly (⇒ reconsider `composition` as a separate package).

Preference, ceremony aversion, sunk cost and hypothetical futures are not triggers.

---

## 17. Adjudication guidance

```text
PRIMARY VERDICT      APPROVE R10 T8-B BACKEND MODULE & PACKAGE TOPOLOGY
                     WITH MATERIAL FIXES

BLOCKER   4          B-01 audit in-tx seam
                     B-02 transaction-scope carrier
                     B-03 transport→owner ambiguity
                     B-04 authorization decision-input seam

MAJOR     8          M-01 platform law over types
                     M-02 identityprovider vs Authentication ownership
                     M-03 public surface, not package count
                     M-04 application lens scoping / anti-God rules
                     M-05 enforcement completeness + decay
                     M-06 enforcement host attribution + reuse gate result
                     M-07 missing repository homes
                     M-08 idempotency replay authority/privacy law

LOW       6          L-01 candidate not in repository
                     L-02 "one Go module" scope wording
                     L-03 rendering conflation
                     L-04 managedcontent non-semantic law
                     L-05 Library/Search read-model home
                     L-06 D10's positive corollary

ANOTHER BROAD REVIEW        NO
BOUNDED ROUND-2 DELTA       YES — sufficient; scope = B-01..B-04, M-01..M-08 only
PROMOTABLE AFTER FIXES      YES, after Lead adjudication + operator ratification
LEAD ADJUDICATION           MAY PROCEED DIRECTLY — no prerequisite blocks it
```

Round-2 scope, if the Lead's corrections land as described in §13, is a delta review only: confirm the three seams are named without T8-C trespass, confirm the dependency graph is complete and non-contradictory, confirm the enforcement completeness/decay obligations are stated, and confirm the tree's scope statement and added homes. No re-derivation of the topology is warranted.

---

**End of independent review. Reviewer findings are evidence, not authority.**
