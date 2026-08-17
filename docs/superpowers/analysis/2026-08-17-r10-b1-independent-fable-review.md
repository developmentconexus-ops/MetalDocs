# R10-B1 Relational Substrate, Tenancy & Reference Law — Independent Fable Review

> **Status:** INDEPENDENT REVIEW OF RECORD — evidence, **NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Reviewer:** Claude Fable 5 — cold session, repository-only bootstrap per `AGENTS.md`
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Design branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Packet under review:** `docs/superpowers/analysis/2026-08-17-r10-b1-relational-substrate-fable-review-request.md` @ `a3bb4ac8`
> **Candidate baseline:** `05a87fa4841ea71128c0538fe86f583075cb4643`
> **Method:** DevelopmentConexus Engineering Method v1.0.0 (`docs/engineering/standards/root-cause-global-maximum-method.md`)
> **Implementation gate:** CLOSED — this review authorizes nothing.

Findings below are evidence for operator adjudication, not requirement authority.

---

# 1. Verdict

```text
VERDICT: APPROVE R10-B1 WITH MATERIAL FIXES
```

No BLOCKER. Two MAJOR law-text defects leave a structurally reachable defect
class inside the substrate's own target invariants; both are bounded,
one-cluster corrections that do not reopen any candidate structural decision.
Four LOW findings are completeness/successor-naming corrections.

Every candidate structural decision — one canonical `metaldocs` schema,
`(tenant_id, id)` composite identity, same-tenant composite FK law,
authority-neutral cross-owner FK, cross-owner cascade prohibition, fail-closed
RLS, non-owner/NOSUPERUSER/NOBYPASSRLS runtime role, tenant-by-tenant system
execution, one-local-transaction cross-owner atomicity, same-commit
Audit/durable-intent — survives independent adversarial attack. None requires a
better alternative substrate.

---

# 2. Review basis

Read order executed: `AGENTS.md` → Method mirror →
`wiki/references/current-agent-handoff.md` →
`wiki/architecture/cohesive-platform-redesign.md` → frozen R3–R9.5 ledger
(`docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`) →
`wiki/architecture/r10-technical-architecture.md` → the packet → current-state
evidence anchors.

Current-state claims in packet §2 were verified against source, not accepted:

- dual-schema product-state split — CONFIRMED
  (`wiki/database/dictionary-index.md`: ~40 `metaldocs.*` + ~25 `public.*`
  product tables);
- `public.document_revisions` is technical/autosave, not the frozen
  `DocumentRevision` — CONFIRMED (`wiki/database/tables/document_revisions.md:8`);
- legacy Approval binds `document_id` + TEXT `content_hash_at_submit` and
  carries a `rejected` status the frozen model no longer has — CONFIRMED
  (`wiki/database/tables/approval_instances.md:41`, ledger §2 "No normal
  terminal reject V1");
- Audit append-only enforced today by privilege posture (INSERT/SELECT-only
  grant, explicit REVOKE) — CONFIRMED (`wiki/database/tables/audit_events.md:55`);
  this is direct evidence that the §6.14 privilege-based immutability slot is
  realizable under a single runtime role;
- RLS false-green history under owner/superuser connections — CONFIRMED
  (`wiki/decisions/0027-rls-adoption-sequencing.md`, M7 F7.4 amendment: every
  pre-fix proof ran as `metaldocs_app` = SUPERUSER + BYPASSRLS + owner);
- current RLS policy is deliberately NULL-permissive (unset GUC → all rows)
  and load-bearing for janitors/claim scans — CONFIRMED (ADR 0027 M3 amendment
  §1, §4). This is the single most important current-state fact for this
  review; see F1.

---

# 3. Findings

## F1 — MAJOR — the tenant-isolation law's scope predicate is undefined, leaving the durable-intent/claim surface classifiable both ways

**Claim.** Packet §6.9 applies fail-closed FORCE RLS to "tenant-owned product
tables"; §6.13 requires durable async intent inserted in the same commit as the
tenant business mutation; §7.2 requires system work to proceed tenant-by-tenant
after discovering eligible Tenants. Nothing binds the scope term "tenant-owned
product table" to the §6.14 classification taxonomy, and no law states which
class durable-intent/queue rows fall into.

**Why it is material.** The two readings diverge on a known real failure class:

- if intent rows are tenant-owned product data (they are created inside a
  tenant-scoped commit, carry tenant attribution, and reference tenant
  aggregates), fail-closed FORCE RLS makes the R10-D claim/dispatch scan
  structurally impossible — a claimer cannot seed a Tenant context it has not
  yet discovered from the row it cannot see;
- if intent rows are EPHEMERAL MECHANISM state, the claim scan is legal, but
  the packet never says so, and a B2–B6/R10-D designer can honestly derive
  either answer.

Current-state evidence proves this is not hypothetical: the entire reason
today's RLS policy is NULL-permissive is exactly this surface (ADR 0027 M3
amendment §1 — outbox claim steps, janitor sweeps and cross-tenant maintenance
scans "depend on this permissive behavior to function at all"; ADR 0054 rule 1).
The candidate rightly kills the NULL-permissive escape hatch but does not state
where its sanctioned residue goes. Note also that today's `outbox_events` has
**no `tenant_id` column at all** (`wiki/database/tables/outbox_events.md`) —
i.e., the current system already implicitly treats intent as mechanism state;
the target law must make that classification explicit and deliberate instead of
inherited.

**Root cause.** §6.9's strongest control is scoped by a vocabulary term
("product") that the packet never defines, while §6.14 introduces the taxonomy
that could define it. One law uses a predicate the other law owns.

**Target property.** Substrate invariants 9 and 13 simultaneously satisfiable:
missing tenant context fails closed on tenant-owned product data **and** the
durable intent inserted same-commit remains claimable by R10-D execution
without a global content bypass.

**What must change (one bounded correction).**

1. Bind the tenant-isolation law's scope to the §6.14 classification:
   MUTABLE AUTHORITY / APPEND-ONLY / TERMINAL tenant-owned tables get the full
   §6.9 stack; EPHEMERAL MECHANISM and PROJECTION / ATTRIBUTED SUPPORT tables
   get an explicit per-class isolation posture decided by their owning block.
2. Classify durable async intent explicitly: platform mechanism state carrying
   tenant attribution, written inside the tenant-scoped business commit,
   claimable through a platform-owned dispatch surface that exposes routing
   attribution (tenant id, kind, due) but no tenant business content; all
   post-claim content access re-enters an ordinary tenant-scoped context.
3. Constrain the successor: R10-D may design claim/retry/DLQ mechanics but may
   not satisfy them by granting any consumer implicit all-tenant *content*
   visibility (invariant 14 restated at the seam where it will be tested).

**Adjudication recommendation:** APPLY as candidate text amendment before
promotion. This is a clarification of the candidate's own evident intent, not
a new decision; without it, proof obligation §12.8 cannot be evaluated.

## F2 — MAJOR — cross-owner FK action law leaves `ON DELETE SET NULL` / `SET DEFAULT` reachable

**Claim.** Packet §6.7 forbids cross-owner `ON DELETE CASCADE` and
`ON UPDATE CASCADE` and states "normal FK action = RESTRICT / NO ACTION", and
proof obligation §12.2 tests only that no cross-owner FK can "cascade-delete"
another owner's history. `ON DELETE SET NULL` and `ON DELETE SET DEFAULT` are
neither forbidden by name nor excluded by the proof obligation.

**Why it is material.** `SET NULL`/`SET DEFAULT` are the same defect class the
law exists to make unrepresentable: one owner's delete silently **mutates**
another owner's durable row through a referential side effect. Against an
APPEND-ONLY/immutable fact family (Approval evidence referencing a Submission,
snapshots referencing sources) a nulled reference is exactly the "silently
rewritten history" of substrate invariants 6–7 — arguably worse than a blocked
delete, because it succeeds silently. A later block author reading "cascade is
forbidden" can legally write `SET NULL` and pass §12.2.

**Root cause.** The law was written as a prohibition list (cascade) plus an
informal default, instead of a closed allowed set. Prohibition lists fail open
on unenumerated members.

**Target property.** Cross-owner deletion can never mutate or delete another
owner's durable state through any referential action; the only cross-owner FK
outcomes are "delete blocked" or "explicit coordinated behavior".

**What must change.** Restate §6.7 as a closed positive law: across semantic
owners the only permitted FK actions are `RESTRICT` / `NO ACTION`; every other
referential action (`CASCADE`, `SET NULL`, `SET DEFAULT`) is forbidden.
Extend §12.2 to "no cross-owner FK action can delete **or mutate**" the
referencing/referenced owner's durable state.

**Adjudication recommendation:** APPLY. One-line law tightening; no candidate
decision changes.

## F3 — LOW — FORCE RLS + fail-closed binds the table owner too; the platform migration/backfill/restore posture needs an explicitly sanctioned home

`FORCE ROW LEVEL SECURITY` applies policies to the owning role. Under the
target fail-closed policy, an owner-role data migration, backfill or restore
touching many tenants' rows with no seeded Tenant context fails closed — which
is correct for runtime and wrong as the *only* posture for platform data
maintenance. Today this is masked because `metaldocs_app` is
SUPERUSER+BYPASSRLS in dev (ADR 0027 M7 amendment). The candidate separates
DDL/ownership from runtime DML (§6.10) but never states the sanctioned
mechanism for cross-tenant platform data operations: a non-serving
platform/maintenance role with explicit `BYPASSRLS`, or per-tenant iteration,
or per-operation policy suspension. Not naming it invites either a silent
BYPASSRLS grant on the wrong role later, or the conclusion that fail-closed is
unworkable. **Fix:** one sentence in §6.10 or §7 assigning this to its
successor (R10-F/implementation) with the constraint that the serving runtime
role never gains the bypass. **Adjudication:** APPLY as deferral-with-named-
successor.

## F4 — LOW — "discover eligible Tenant IDs" presumes a cross-tenant-readable due-work signal whose home is unnamed

§7.2's discovery step works trivially when eligibility is computable from root
tables (`Tenant` is not tenant-owned). But time-due work whose truth lives in
tenant-owned rows (release `not_before`, periodic-review due, retention
eligibility) has exactly two lawful discovery sources under fail-closed RLS:
per-tenant iteration (seed each tenant, query due work), or a platform
mechanism registry/intent row (F1's class) written same-commit. Both are
viable at MetalDocs scale; the packet should name that these are the only two
shapes, so no later block invents a fail-closed *exemption* on tenant product
tables as a third path. **Adjudication:** APPLY as one clarifying sentence in
§7.2 (naturally merges with F1's correction).

## F5 — LOW — supporting-owner persistent state has no named home in the B2–B6 decomposition

§1 assigns B2 (AuthN/Org/AuthZ), B3 (CI authoring/submission), B4
(Approval/Release/Distribution), B5 (DC/Records), B6 (cross-owner atomicity +
historical truth). The persistent-state models of the three supporting
semantic owners — Audit timeline shape, Artifact identity/staging/confirmation
facts, Interchange batch/plan/outcome state — and of Notifications/Search
support state are not explicitly assigned to any block. §6.14 requires
classification of "every durable table/fact family in B2–B5", which is
impossible if some families have no block. **Fix:** extend the decomposition
map (e.g., Artifact → B3 with B5 consumer closure, Audit → B6, Interchange →
B6, Notifications/Search → R10-D boundary) or state the assignment rule.
**Adjudication:** APPLY as decomposition-map completion; no new decision.

## F6 — LOW — target namespace reuses the name of a legacy-populated schema; cutover disambiguation belongs to R10-F but should be acknowledged

The canonical target schema `metaldocs` is also today's partially-populated
legacy namespace (~40 legacy tables). The choice of name is correct — target
law should carry the clean product name, and migration inconvenience is
explicitly not reopen evidence — but R10-F will need an explicit
disambiguation mechanism (staging namespace, rename choreography, or
table-level manifest) because "lives in `metaldocs`" will not distinguish
target from legacy during transition. One acknowledging sentence prevents a
future reviewer from reading namespace as provenance. **Adjudication:** APPLY
one sentence, or ACCEPT-as-known and route to R10-F. Either disposition is
sound.

---

# 4. Adversarial questions — disposition of all 20 (packet §11)

1. **Canonical `metaldocs` schema — Global Maximum?** YES. Alternative A
   (keep split) preserves the accident; C (schema-per-BC) adds
   grants/search-path/cross-schema FK machinery while every real boundary
   (deployment, transactions, references, trust) stays shared — namespace
   theater; D and E weaken referential proof. B is the smallest structure
   preserving every invariant. See F6 for the only residue.
2. **Composite `(tenant_id,id)` accidental complexity?** Net negative
   complexity. Costs: wider indexes, two-column FKs. Removes: the entire
   cross-tenant-reference class, trigger/discipline machinery, and duplicate
   global-identity indexes. A child carrying one `tenant_id` column shared by
   several composite FKs structurally forces same-tenant coherence across its
   whole parent set — stronger than any application check. E-with-extra-UNIQUE
   achieves the same FK capability only by adding the redundant index the
   candidate rejects.
3. **`UNIQUE(id)` needed?** NO real consumer. All lookups are
   tenant-resolved (session → tenant), cross-tenant URL → 404 falls out of the
   composite key naturally; Audit attribution is non-authoritative text.
4. **Cross-owner FKs improperly coupling lifecycle/migration/erasure?** None
   found. RESTRICT makes disposition/erasure an explicitly ordered coordinated
   delete — that is the frozen semantics, not a defect. Audit is correctly
   exempted from FK coupling (§6.6) — an Audit FK would either block disposal
   or cascade, both wrong. Historical Migration under composite FKs imports in
   dependency order per semantic unit — consistent with the ledger §13.
5. **Cascade ban sufficient?** NO — F2 (`SET NULL`/`SET DEFAULT`).
6. **Cross-tenant reference still reachable?** Where existence-FKs exist: no.
   Residuals are the declared no-FK surfaces: Artifact's opaque owner
   reference (bounded defer, successor B3/B5, named in §6.6/§8 — verified as
   a legal TRANSITIONAL/DEFER, not a gap), Audit attribution
   (non-authoritative), projections (rebuildable). Acceptable.
7. **Product-global facts wrongly tenant-forced?** No — §6.2 exception clause
   covers credential/product facts (e.g., `auth_identities` tenant-global by
   design, ADR 0027 §1) and defers exact design to B2. The
   "do not mechanically add `tenant_id`" sentence prevents uniformity creep.
8. **Fail-closed RLS operationally impossible?** No — it requires explicit
   tenant execution context, which §7.2 supplies, **provided** F1/F3/F4 are
   applied. Without them the claim/backfill surfaces are ambiguous.
9. **Tenant-by-tenant sufficient?** YES for release, periodic review,
   retention, erasure, reconciliation at current and foreseeable scale
   (per-tenant transactions are cheap; discovery via root/platform surfaces).
   No bounded global-content capability is justified by evidence.
10. **Does the role law guarantee RLS exercised?** Non-owner + NOBYPASSRLS +
    ENABLE suffices for the serving role; FORCE additionally covers accidental
    owner-connection regressions — exactly the historical false-green class
    (ADR 0027 M7). Residual: deployment discipline that the serving pool
    actually uses that role — implementation-stage proof (promotion condition
    P5).
11. **RLS → Authorization creep?** Law forbids Area/role/participant
    predicates in policies (§6.9). Design-level prohibition is proportionate
    at B1; a policy-shape guard is implementation-stage work.
12. **Cross-owner tx without repository inversion?** Feasible — a
    tx/UnitOfWork carrier through published application seams; current
    TxRunner chokepoint is mechanism evidence. Concrete interface correctly
    deferred.
13. **Same-commit law overconstrains multi-step operations?** No — §7.3
    explicitly scopes atomicity to frozen single-transition semantics; erasure,
    disposition, publication remain coordinated processes with verified
    completion. Consistent with the outbox premise (state-write and external
    effect never share a tx).
14. **Immutable facts unprotectable?** No — the privilege-slot is proven
    realizable today (`audit_events` INSERT/SELECT-only + durable REVOKE);
    triggers/constrained transitions remain for row-state facts. §6.14's
    falsifiable-enforcement requirement is the right B1-level control.
15. **TEXT+CHECK vs ENUM?** TEXT+CHECK correct — equal validation strength,
    no type-migration coupling for frozen-but-evolvable vocabularies; matches
    current practice evidence.
16. **BYTEA(32) SHA-256?** Correct. Exact-byte semantics, existing constraint
    idiom (`documents_*_hash_len`), hex/base64 are edge encodings. The current
    TEXT `content_hash_at_submit` in legacy Approval is counter-evidence for
    TEXT, not for BYTEA.
17. **Prejudges B2–B6?** No. It constrains them — which is its charter. The
    UUID technical-id law overrides current TEXT ids (`iam_users`,
    `audit_events`); current shape is not target entitlement. Participant
    types, table sets, state models all remain open.
18. **Essential decision deferred too late?** Only the F1 classification
    binding (must be in B1 because both §6.9 and §6.13 are B1 laws). F5 is a
    map defect, not a timing defect. UUID generation flavor, timestamp
    conventions, naming conventions are correctly below B1's altitude.
19. **Abstraction caused by abstraction?** None. No ORM, no UoW framework, no
    generic relation registry, no per-BC roles/schemas, no global
    SERIALIZABLE. The candidate is notably subtractive.
20. **Strongest realistic fragility path?** (a) per-tenant fan-out cost if
    tenant count grows large — bounded, and the platform dispatch surface
    (F1 fix) is the natural relief valve; (b) operations/support friction from
    fail-closed reads — answered by the F3 sanctioned platform posture;
    (c) erasure delete-ordering complexity under RESTRICT — inherent to the
    frozen "explicit coordinated erasure" semantics, not substrate-induced.
    None reaches operational non-viability.

---

# 5. Structural Inversion result

Applied against current shape, per law:

- **Schemas inverted** (one schema today / per-BC today): one canonical
  product schema still follows from "namespace is not authority" + shared
  deployment/tx reality. HOLDS.
- **IDs inverted** (global single-column PKs everywhere): composite tenant
  identity still follows from frozen explicit tenant ownership + the
  cross-tenant-reference class, not from the current mixed key styles. HOLDS.
- **RLS inverted** (no RLS today / RLS-everywhere today): fail-closed
  tenant-isolation defense-in-depth still follows from frozen ledger
  invariants; notably the candidate **inverts the current NULL-permissive
  policy against its own installed base** — the opposite of
  current-shape-as-entitlement. HOLDS, and is the strongest evidence the
  candidate is not schema archaeology.
- **Role inverted** (per-BC roles today): still no evidenced per-context trust
  boundary → still one runtime role. HOLDS.
- **Tx helpers inverted** (no TxRunner today): one-local-transaction law
  derives from frozen atomicity invariants. HOLDS.

Structural Inversion: **PASS**. The one law that leans on current-state
evidence rather than frozen semantics alone is the same-commit outbox premise —
and it leans correctly (mechanism evidence, shape explicitly non-surviving).

---

# 6. Global Maximum vs local maximum

The candidate is a Global Maximum at the substrate altitude:

- it removes the structure that produced the defect class (namespace accident,
  discipline-only tenancy, NULL-permissive escape hatch, owner/superuser
  false-greens) instead of patching inside it;
- it is smaller than every credible stronger-looking alternative
  (schema-per-BC, per-BC roles, global SERIALIZABLE, generic polymorphic
  registry) — each was tested and each adds machinery without eliminating a
  failure class;
- it prepares seams (dispatch surface, UoW carrier, immutability enforcement
  slots) without pulling R10-C/D/E/F mechanism design forward.

The two MAJORs are law-text incompleteness inside the chosen structure, not
evidence of a wrong structure. No finding at the structural altitude recurred
across this review — convergence is toward mechanical text fixes.

---

# 7. YAGNI / subtractive pass

Checked every candidate mechanism for removability without weakening a
distinct property:

- **FORCE RLS on top of non-owner role** — keep: protects the distinct
  owner-connection-regression path with zero added complexity (historical
  false-green class).
- **App predicate + composite keys + RLS (three layers)** — keep all three:
  result-correctness, referential tenant integrity, and leak backstop are
  distinct properties; no pair proves the third.
- **§6.14 five-class taxonomy** — keep; it becomes load-bearing under the F1
  fix (it scopes the isolation law).
- **Isolation-level law** — keep; it is a prohibition (anti-SERIALIZABLE
  creep), not machinery.
- Nothing found removable. Conversely, nothing speculative found: no unused
  extensibility, no second authority, no compatibility layer. The candidate
  already deletes the strongest temptations (UNIQUE(id), per-BC schemas/roles,
  polymorphic registry).

Subtractive pass: **CLEAN** both directions.

---

# 8. Operational viability — will MetalDocs actually run on this?

**YES, conditional on F1+F3+F4.** Assessment per surface:

- **API traffic:** unchanged shape — tenant ctx from session, every request tx
  seeded at a chokepoint (mechanism evidence: current TxRunner autoseed).
  Fail-closed converts today's silent-leak failure mode into a loud 4xx/5xx —
  correct direction.
- **Background/jobs:** tenant-by-tenant execution matches how the current
  system already processes work post-claim (per-message `SeedTxTenant`
  evidence). The delta is discovery/claim, which today rides the
  NULL-permissive hatch — this is exactly F1's seam and is solvable inside the
  candidate's own laws (platform dispatch surface or per-tenant iteration).
  Without the F1 text, R10-D inherits an ambiguity; with it, the path is
  closed and buildable.
- **Maintenance (watchdog/janitor/integrity scans):** per-tenant iteration
  from root discovery is sufficient at realistic scale; TTL sweeps on
  mechanism tables fall out of the F1 classification.
- **Migration/backfill/restore:** viable once F3 names the sanctioned
  platform posture; without it, FORCE + fail-closed + non-bypass owner is
  operationally contradictory for bulk platform data work.
- **Tenant erasure:** per-tenant by construction; RESTRICT FKs impose an
  explicit delete order — coordination complexity the frozen semantics already
  demand. Audit skeleton reads for an ERASED tenant still work (seed the
  erased tenant's id).
- **Failure visibility:** §7.4's fail-loud rule is the correct substrate
  posture; observability contracts correctly deferred.

The design permits an operational MetalDocs, not merely an elegant one. The
honest cost is fan-out latency at high tenant counts and stricter ops
discipline — both bounded, neither near non-viability at the product's scale.

---

# 9. R10-A / R9.5 reopen assessment

**R10-A reopen set: EMPTY.** The substrate respects the 8+3 topology
throughout: schema namespace explicitly carries no ownership semantics;
cross-owner FKs are authority-neutral existence proofs; composition
coordinates transactions without owning meaning; no law moves a fact family
between owners. No candidate decision requires reopening ownership, and this
review introduces none. F5 asks for block *assignment* of already-owned fact
families — routing, not ownership.

**R9.5 reopen set: EMPTY.** Every law implements a frozen invariant
(tenant ownership, immutable Submission/Decision/snapshot truth, same-commit
audit/durable intent, no fabricated history, fail-closed posture, provider
identity never business identity). No frozen semantic is weakened, and no
finding proposes weakening one. The NULL→fail-closed RLS inversion changes
current *implementation* behavior, which is not frozen authority.

No finding in this review constitutes a new product requirement, bounded
context, external service, framework or engine.

---

# 10. Exact promotion conditions

R10-B1 may be promoted into `wiki/architecture/r10-technical-architecture.md`
when all of the following hold:

- **P1 (F1):** the tenant-isolation law's scope is bound to the §6.14
  classification; durable async intent is explicitly classified (platform
  mechanism state, tenant-attributed, claimable via a platform-owned dispatch
  surface exposing routing attribution but no tenant business content); R10-D
  successor constraint recorded (no implicit all-tenant content visibility).
- **P2 (F2):** cross-owner FK actions restated as a closed allowed set
  (`RESTRICT`/`NO ACTION` only; `CASCADE`/`SET NULL`/`SET DEFAULT` forbidden),
  and proof obligation §12.2 extended from "cascade-delete" to
  "delete or mutate".
- **P3 (F3, F4):** the platform migration/backfill/restore posture and the
  two lawful due-work discovery shapes are named with their successors
  (R10-F / R10-D), each with the serving-role-never-bypasses constraint.
- **P4 (F5, F6):** the B2–B6 decomposition map assigns supporting-owner and
  support/projection persistent state to explicit blocks; the
  target-vs-legacy `metaldocs` namespace disambiguation is acknowledged and
  routed to R10-F (or explicitly accepted as known).
- **P5 (proof slots, carried to implementation stage):** the promoted text
  preserves falsifiable proof obligations for later stages — fail-closed
  negative proof under a non-owner NOBYPASSRLS role (the `metaldocs_ci`-class
  proof, not an owner connection), a cross-owner FK-action census, and a
  same-commit audit/intent rollback proof. No promotion-blocking work; these
  are the §12 obligations restated so they survive adjudication.

Operator adjudication of F1–F6 under the Method is the promotion gate; no
finding here requires returning to the alternatives comparison.

---

# 11. Convergence

```text
findings: 0 BLOCKER / 2 MAJOR / 4 LOW
altitude: law-text completeness and successor naming; zero structural rejections
recommendation: fix-and-promote, no further full adversarial round required;
adjudicated text delta may be verified by a bounded delta check
```
