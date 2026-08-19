# R10-T8C — Internal Communication Contracts — Independent Fable Review

```text
INDEPENDENT ADVERSARIAL REVIEW
REVIEWER EVIDENCE ONLY
NOT TARGET AUTHORITY
NOT IMPLEMENTATION AUTHORIZATION
NOT A CORRECTION MANDATE
```

> **Date:** 2026-08-19
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Review class:** material independent Fable review
> **Stage:** T8-C ACTIVE — implementation BLOCKED

Reviewer findings are evidence until Lead adjudication and operator ratification. Nothing in this artifact promotes T8-C, opens T8-D or authorizes code.

---

## 1. Reviewed remote HEAD and provenance

Remote authority was re-established independently before any file was read. The HEAD asserted in the dispatch prompt was **not** trusted as input.

```text
git fetch --all --prune
  From https://github.com/developmentconexus-ops/MetalDocs
    c79b9cac..3a882c4d  docs/a8-authz-approval-redesign-ledger -> origin/...

git rev-parse origin/docs/a8-authz-approval-redesign-ledger
  3a882c4daa3707cedbe6ed1019c2001eabebca34

gh pr view 131 --json headRefOid,headRefName,baseRefName,state
  headRefOid   3a882c4daa3707cedbe6ed1019c2001eabebca34
  headRefName  docs/a8-authz-approval-redesign-ledger
  baseRefName  main
  state        OPEN
```

```text
REVIEWED HEAD   3a882c4daa3707cedbe6ed1019c2001eabebca34
PROVENANCE      origin fetch + PR #131 headRefOid agreement
PROMPT CLAIM    matched, but independently re-derived, not assumed
```

Review executed in a detached worktree pinned to that commit (`git status --porcelain` clean). The candidate header cites revalidated baseline `6bc5e229`; the reviewed candidate content is carried unchanged at `3a882c4d`.

---

## 2. Reconstructed authority chain

Reconstructed in AGENTS.md order, from the repository only. Conversation memory, prior agent reasoning and the candidate's own authority claims were excluded as authority.

```text
1  AGENTS.md
2  docs/engineering/standards/root-cause-global-maximum-method.md   (Method v1.0.0, ACCEPTED)
3  wiki/references/current-agent-handoff.md
4  wiki/architecture/r10-technical-architecture.md                  (sole stage/status router)
5  T1  r10-t1-semantic-state-invariants.md
   T2  r10-t2-governance-effectivity-transactions.md
   T3  r10-t3-authorization-audit-enforcement.md
   D4  r10-t3-d4-responsible-owner-eligibility-amendment.md
   T4  r10-t4-exact-content-storage-integrity-restore.md
   T5  r10-t5-durable-async-search-external-effects.md
   T6  r10-t6-canonical-api-frontend-journeys.md
   T7  r10-t7-historical-migration-truth-semantic-mapping.md
   T8-A r10-t8a-technical-authority-legacy-disposition.md
   T8-B r10-t8b-backend-module-package-topology.md                  (upstream binding)
6  rebaseline-decision-registry.md + d4/t6/post-t6/t7/t8a/t8b amendments
7  docs/.../2026-08-19-r10-t8c-...-bootstrap.md
8  docs/.../2026-08-19-r10-t8c-...-global-maximum-candidate.md      (review target)
9  current repository evidence, used only for concrete claims
```

Router status independently confirmed: `T1→T8-B CLOSED / OPERATOR-RATIFIED`, `T8-C ACTIVE`, `T8-D→T12 NOT OPEN`, `implementation BLOCKED`.

Repository evidence actually opened (not remembered):

```text
go.mod                                                   River v0.37.1, riverdatabasesql v0.37.1, pgx v5.9.2, uuid v1.6.0
vendor/.../river/client.go:1765                          InsertTx signature
vendor/.../riverdriver/riverdatabasesql/...:94            UnwrapExecutor(tx *sql.Tx)
internal/platform/db/tx.go:27-31                          current narrow Tx interface
internal/modules/approval/jobs/lifecycle_event_enqueuer.go:43-46   current db.Tx -> *sql.Tx downcast
```

External primary references consulted (evidence only, never requirement-creating): River current documentation via Context7 (`/riverqueue/river`), plus the repository-vendored River source at the pinned version, which is stronger primary evidence than documentation for this repository.

---

## 3. Primary verdict

```text
APPROVE T8-C GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
```

The **selected model class is confirmed**. Every alternative contract-placement class I constructed and attacked produced equal or greater total complexity, or violated ratified authority. The candidate is not legacy-shape inheritance and does not require a T8-B reopen.

The material fixes are **instantiation gaps and one seam shape**, not model failure:

```text
the model is right
several contracts it promises to freeze are not actually frozen
one frozen contract cannot serve a mechanism the same candidate mandates
```

Under Method §2, each surviving finding touches an invariant, an authority boundary, or a correctness property, and each would otherwise reach a Writer as an undecided material architecture decision — the exact condition the binding execution law forbids.

---

## 4. Finding counts

```text
BLOCKER   5
MAJOR     6
LOW       5
```

BLOCKER = must be corrected before T8-C promotion; a Writer would otherwise invent ratified-authority-bearing contract.
MAJOR = material correctness/ownership defect that must be corrected or explicitly deferred with a recorded trigger.
LOW = precision/consistency defect; no invariant at risk.

---

## 5. BLOCKER findings

### B1 — `txscope.Scope` cannot serve the River-backed durable-intent sink the same candidate mandates

**Implicates:** T8C-D04, D05, D17.

Verified from primary repository evidence and current River documentation:

```text
vendor/.../river/client.go:1765
  func (c *Client[TTx]) InsertTx(ctx, tx TTx, args JobArgs, opts *InsertOpts)

vendor/.../riverdriver/riverdatabasesql/river_database_sql_driver.go:94
  func (d *Driver) UnwrapExecutor(tx *sql.Tx) riverdriver.ExecutorTx
```

`TTx` is the driver's **concrete** transaction type. For the driver this repository pins, that is `*sql.Tx`.

The candidate freezes (§4.1):

```go
type Scope interface { ExecContext(...); QueryContext(...); QueryRowContext(...) }
```

and freezes (§15.2):

```go
EnqueueOfficialRenditionIn(ctx context.Context, scope txscope.Scope, submissionID uuid.UUID, requiredFormat string) error
```

There is **no contract by which the River implementation obtains a concrete transaction from `Scope`.** `*sql.Tx` satisfies `Scope`; `Scope` does not yield `*sql.Tx`.

Current code shows precisely what a Writer invents when this seam is unfrozen:

```go
// internal/modules/approval/jobs/lifecycle_event_enqueuer.go:43
sqlTx, ok := tx.(*sql.Tx)
if !ok {
    return fmt.Errorf("lifecycle_event_enqueuer: river requires *sql.Tx, got %T", tx)
}
```

A runtime downcast with a runtime error path, annotated in-tree as "the one allowed coupling point with River infra."

Why this is BLOCKER, not stylistic:

```text
§15.3 asserts   Submission-requiring-rendition commits <=> intent row commits
that atomicity  rests on a mechanism the frozen contract cannot express
Method §Enforcement  requires the strongest reasonable mechanism covering all
                     paths the boundary structurally admits, preferring earlier
                     feedback among equally sufficient mechanisms
result          a compile-time-provable binding is downgraded to an unproven
                runtime type assertion, and T8-C did not decide it
```

Secondary defect, opposite direction: the executor shape **silently forecloses a T8-D option**. A pgx-native realization (`pgx.Tx`/`pgxpool`) cannot produce `*sql.Rows`/`*sql.Row`. T8-D would be restricted to `database/sql` (including `pgx/stdlib`). §4.4 says only "the candidate does not require pgx-native transaction types" — permissive language for what is in fact a foreclosure. A constraint imposed on a later stage must be **declared**, not implied.

**Correction direction (contract-level, not realization):** freeze an explicit, typed transaction-binding contract in `platform/txscope` usable by `platform` mechanisms only — for example a declared unwrap/binding accessor whose only legal callers are platform mechanism packages, enforced by the same closed-world `cilint` catalog T8-B §10 already requires — and state the `database/sql`-family restriction on T8-D as a ratified consequence of D04.

### B2 — No Audit read contract exists

**Implicates:** T8C-D06, D22, and the operation census.

The candidate's entire public Audit surface is `AppendIn` (§5.3). §18.3 names a lens (`AuditEventView`, "Audit facts + current Authorization audit.read scope") but freezes **no method, no filter vocabulary, no scope-parameter shape, no pagination**.

Ratified authority requires the operation:

```text
T6 §29     GET /api/v1/audit/events
T6 §26     AuditEventView is a purpose-built read model; lists are cursor-paginated
T3 §14     audit.read @ Company -> Company events + all Area-attributed events
           audit.read @ Area X  -> only events historically attributed to Area X
           visibility snapshotted at action time; current relocation never rewrites it
```

This is not a mechanical omission. Audit **cannot** import Authorization (T8-B §9.2 owner→owner forbidden), so the only legal shape is application-routed: `Authorization` decides `audit.read` and its scope, application maps a **bounded resolved scope fact** into an Audit-owned query, and Audit applies it against snapshot attribution. That is exactly the candidate's own selected pattern — never instantiated here.

Left unfrozen, a Writer chooses between three options, two of which are defects: application post-filters Audit rows (creates a second visibility evaluator over T3 §14 semantics), or Audit re-derives scope itself (absorbs Authorization meaning).

### B3 — No Authorization scope-enumeration contract; T6 §8 and cursor pagination have no legal path

**Implicates:** T8C-D07, D22.

T6 requires answers to *"where is this subject permitted?"*, not only *"is this target permitted?"*:

```text
T6 §8   Company-scoped document.create -> current ACTIVE Areas eligible for creation
        Area-scoped document.create    -> only matching current ACTIVE Areas
        Responsible-owner candidates returned only when caller has document.owner.manage
        in the selected scope
T6 §26  potentially unbounded lists: ?cursor=&limit=, default 20, max 100,
        next_cursor / has_more; cursor binds filter/order semantics
```

The candidate's Authorization surface offers only `Decide` / `DecideIn` / `DecideMany` / `DecideManyIn` — per-target evaluation. Neither answers the scope question.

The two remaining routes are both closed:

```text
enumerate all Areas -> DecideMany over each
  = the unbounded candidate-set anti-pattern; cost scales with directory size,
    and for Library it forces filter-after-paginate, so a page of 20 cannot be
    filled deterministically and has_more / next_cursor become incoherent

derive scopes in application from Authorization.RoleAssignments + role bundles
  = a second evaluator of canonical Authorization semantics
    forbidden by T3 §2, T8-B §8.3/§9.2, and by the candidate's own D07
```

So the frozen contract set cannot satisfy a ratified T6 read requirement by any legal path. The minimal correction is additive and stays inside the selected model: an Authorization-owned bounded scope-enumeration query (conceptually `AuthorizedScopes(ctx, subject, permission) -> company-wide | []AreaID`), evaluated by the same canonical evaluator. It is not a policy language, it introduces no owner, and it simultaneously removes the N+1 and the pagination-coherence break.

### B4 — ManagedContent port omits operations T4 lists as required, and the T5-J GC family has no contract path

**Implicates:** T8C-D15, D14.

T4-D states the **minimum** contract and its conceptual operations:

```text
T4 §6   minimum contract includes: cleanup of reclaimable mechanism content
        PresignCreate / Stat / OpenExact / OpenRange / CopyToNewHandle / DeleteReclaimable
```

Candidate port (§12.1): `Allocate`, `PresignCreate`, `Stat`, `OpenExact`, `Copy`.

```text
OpenRange         correctly omitted — T4 marks range read optional; candidate's
                  reasoning and its reopen trigger are sound  (credit)
DeleteReclaimable MISSING — T4 lists it in the minimum contract
admission claim   MISSING — T4-F requires an opaque unforgeable binding/claim that
lifecycle         protects a READY handle from GC "until the claim is consumed,
                  explicitly released or reaches a bounded mechanism expiry"
```

Downstream, an entire activated Launch mechanism family has no contract:

```text
T5-J  claim bounded GC_PENDING set
      -> re-prove no current WorkingContent reference
      -> re-prove no immutable governed/imported reference
      -> re-prove no live admission claim/binding
      -> re-prove no backup exclusion/pin
      -> provider DeleteReclaimable outside semantic tx
      -> finalize/remove mechanism state
```

The candidate mentions "reclaimable bytes remain mechanism state" (§11.4, §14.2) but freezes no port operation, no ControlledDocs "no live semantic reference for these handles" query, and no claim consume/release. `GC_PENDING` appears nowhere in the candidate. T5-J runs `transport/jobs → application`, so its contracts are T8-C, not T8-D. Naming the required operations is contract work; their schema remains T8-D.

### B5 — Idempotency `BeginIn` concurrent-loser path is unspecified and can poison the caller's Scope

**Implicates:** T8C-D19.

T6 §18 ratifies the behavior:

```text
insert/lock scoped Idempotency-Key
Concurrent same-key requests serialize on the key; after winner commit, the
other may replay. No baseline public/durable IN_PROGRESS replay state exists.
```

The candidate's D19 correctly derives that no `FailReplay` is needed for the **crash/rollback** path: an uncommitted claim disappears with its transaction. That reasoning is sound and I confirm it.

It does not address the **concurrent-loser** path, which is the path T6 explicitly ratifies. With acquisition inside the caller's transaction:

```text
loser's insert blocks on the unique key
winner commits
loser's insert returns a unique violation
PostgreSQL marks the loser's transaction ABORTED
no further statement in that transaction is valid
```

`BeginIn` therefore cannot return `replay` to the caller as §16.2 promises, because by that moment the caller's `Scope` is unusable. Resolving it requires either an internal savepoint/subtransaction inside the mechanism, or an application-visible bounded retry. `Runner.Within` (§4.2) defines neither; §16.6's success ordering shows a single linear path with no conflict branch.

This is the load-bearing consequence of the otherwise-correct "no FailReplay" decision, and it is a crash-consistency law, so it must be frozen rather than left to realization. Note the direction: the choice between savepoint-internal and caller-retry is a **contract** choice (it changes `BeginIn`/`Within` semantics), even though the SQL is T8-D.

---

## 6. MAJOR findings

### M1 — T3 §11 eligibility serialization is not expressible in the frozen contracts

T3 §11 requires *serialization*, not merely same-transaction participation:

```text
Governed/security-changing operations whose correctness depends on an enabled
User must serialize with offboarding on current User eligibility.
Applies at least to: new Session issuance, GroupMembership add, new direct User
RoleAssignment, governance ACCEPT / RETURN_FOR_CHANGES, Submission / withdraw /
cancel / obsolescence user mutations.
```

T2 §3 fixes the posture as `READ COMMITTED + narrow explicit lifecycle serialization + OCC/CAS`. Under READ COMMITTED a plain read does not serialize against a concurrent update.

The candidate offers one contract, `SecuritySubjectIn`, used both for ordinary actor-subject reads (§17.4, §17.5, §17.6) and for eligibility-critical reads, with the justification (§6.2) that it "participates in the same local transaction." Participation is not serialization.

Consequence: T8-D cannot know where to attach row locking without re-deriving T3 §11's operation list by inspecting every application flow — duplicating T3 semantics into the persistence stage. If T8-D instead locks unconditionally, every governed command takes a row lock on the actor, adding contention and lock-ordering risk that no stage decided.

**Important scoping:** T2 §3's *Document* serialization root is **not** a finding. It is intra-owner (all listed operations are ControlledDocs mutations), so ControlledDocs-private persistence can acquire it without any cross-owner contract. Correct deferral. The eligibility root differs precisely because the row is Organization-owned while the serializing operations belong to ControlledDocs, Authorization and Authentication — which makes it a T8-C contract question.

No proof in §25 or §29 covers eligibility serialization.

### M2 — If-Match / expected-version routing is unfrozen

T6 §24 requires strong ETag + `If-Match` + `412 precondition.resource_changed` + zero mutation across at least eleven whole-replacement resources, and names three singleton APIs explicitly; T6 §4 repeats it for provider-binding replacement; D4 repeats it for responsible-owner replacement.

Only DRAFT carries a precondition in the candidate (§17.9, "expected generation"). `UpdateCompanyIn`, `UpdateUserProfileIn`, `SetUserEligibilityIn`, `ReplaceProviderBindingIn`, `UpdateAreaIn`, `SetAreaLifecycleIn`, `UpdateGroupIn`, `UpdateDocumentTypeIn`, `ReplaceDocumentGovernanceIn`, `ReplaceEligibleTemplatesIn`, `ChangeResponsibleOwnerIn`, `SetTemplateRoleIn` carry none, and the corresponding reads are not stated to return a version token.

The token's **meaning** is owner truth; T8-E owns only its wire syntax. Leaving it unfrozen invites the inversion where T8-E becomes the de facto owner of concurrency control — the exact direction T8-B forbids.

**Credit where due:** T6 §24's "exact already-current repeat may be no-op and must not fabricate duplicate semantic Audit" **is** structurally satisfied, because §11.3 makes a mutation result carry *zero*-or-more evidence.

### M3 — `ProviderClient` leaves `sub` inside `json.RawMessage`, pushing protocol parsing into a semantic owner

T8-B §7.3 assigns platform "discovery, token exchange, JWKS/protocol verification and raw provider protocol handling"; T6 §3 requires the callback to "verify state/code/provider → resolve issuer + subject"; OIDC Core makes issuer+subject the protocol identity coordinates — a point the candidate itself records in §22.

The candidate types `issuer` but returns everything else as `json.RawMessage`, leaving Authentication to parse provider JSON to obtain `sub`. That is asymmetric: one protocol coordinate is extracted by the protocol layer, its twin is not. It makes Authentication a parser of provider-specific protocol structure, and it reduces the "provider roles/groups/claims never escape" law from a structural property to Authentication's parsing discipline.

`SearchSubjects` is the sharper case: T6 §4 already ratifies the result shape — "opaque `provider_subject_ref` + bounded display hints" — so returning raw provider directory JSON forces Authentication to parse a provider-specific user representation that T6 already reduced.

Smallest correct anti-corruption split: the port returns **verified typed protocol coordinates** (issuer, subject) plus bounded typed references for directory search, with any residue opaque. Platform still owns no meaning; Authentication still owns binding, session and assurance.

### M4 — `MalwareInspector` carries no byte-identity correlation

T4-G is a production-safety invariant over *those exact immutable bytes*. The port is:

```go
Inspect(ctx context.Context, content io.Reader) (clean bool, err error)
```

The verdict binds to nothing. Admission derives SHA-256 in a *separate* read via `OpenExact` (§12.3). Create-once immutability means the bytes cannot change between reads — a real mitigation, and the candidate deserves credit for making the store create-once — but nothing structurally prevents a clean verdict for one handle from being consumed by an admission of another.

Minimal, near-zero-cost strengthening: the inspector returns the digest of exactly what it consumed; the owner admits only when that digest equals the descriptor it derived. Converts caller discipline into a checkable correlation, with no coupling between the two mechanism ports.

### M5 — The replay PII / erasure decision T8-B explicitly assigned to T8-C is not made

T8-B §7.6 is explicit:

```text
replay persistence must be erasure-safe and not become a PII retention root
Exact replay contract/schema/retention and PII-free-vs-purge realization remain
T8-C/T8-D.
```

T6 §19 repeats the constraint. The candidate states only that a snapshot contains "no more PII than the original replayable result requires" (§16.4) and defers retention to T8-D. The actual assigned decision — **PII-free snapshot with re-projection at replay, versus stored payload with purge-on-erase** — is never taken, and the two horns genuinely conflict:

```text
PII-free + re-project   replay would return current state, breaking "exact
                        original status/body", and cannot reconstruct after erasure
stored payload          POST /users replay retains profile PII after
                        DELETE /users/{id}/profile, for the retention window
```

`POST /users` is in T6 §18's required-Idempotency-Key list and `DELETE /users/{user_id}/profile` is a ratified erasure operation, so the collision is concrete, not hypothetical.

### M6 — `GET /api/v1/official-renditions/{rendition_id}/content` has no named read contract

T6 §20 distinguishes four semantic byte resources and treats the OfficialRendition PDF as distinct from Release *source*:

```text
GET /revisions/{revision_id}/draft/source
GET /submissions/{submission_id}/source
GET /releases/{release_id}/source
GET /official-renditions/{rendition_id}/content
```

The candidate names `OpenSemanticSource` (§11.4), which plausibly covers the three source resources, and `RenditionWorkCandidate` (worker-side). No contract covers reading rendition content for presentation. By the census test, a Writer would invent it — including deciding which `AccessFacts` action governs it (`READ_EFFECTIVE` for current, `READ_HISTORY` for historical is the evident mapping, but it is undecided).

---

## 7. LOW findings

```text
L1  §17.5 RoleAssignment flow omits idempotency although POST /role-assignments
    is in T6 §18's required-Idempotency-Key list. §16's general law governs, so
    no invariant is at risk; the named flow is simply inconsistent with it.

L2  DecideMany / DecideManyIn return []Decision with no stated index
    correspondence to the input []Check. A length or ordering mismatch would be
    silent. State the correspondence law.

L3  GET /documents/{id}/responsible-owner and GET /documents/{id}/template-role
    have no named read family; both are plausibly inside DocumentOfficial but
    this is not stated, and both are If-Match singletons per T6 §24 (see M2).

L4  Obsolescence completion has no named mutation. It is plausibly inside
    DecideGovernanceStepIn (final ACCEPT) and inside InitiateObsolescenceIn for
    NoHumanApproval, mirroring the explicit Release treatment in §11.2, but T3
    §15 requires same-commit Audit for "obsolescence completed", so the hosting
    mutation should be stated as Release already is.

L5  "ProviderSubjectBinding accepted / disabled / replaced" (T3 §15) exposes no
    distinct disable operation; it is plausibly subsumed by
    ReplaceProviderBindingIn. Worth one explicit sentence.
```

---

## 8. Disposition of T8C-D01 → T8C-D25

| ID | Disposition | Technical basis |
|---|---|---|
| **D01** authority-aligned hybrid ownership | **ACCEPT** | Survived every alternative class I constructed; aligns Method §Authority ("mechanism ≠ authority", one authority per meaning) with T8-B §9. No cheaper class found. |
| **D02** concrete owner Services, no default owner interface | **ACCEPT** | One product implementation at Launch; Go permits additive producer evolution; an implementor-side interface would be mock-driven ceremony. Matches Method §Complexity law question 1 (no current consumer). |
| **D03** consumer-owned interfaces only for real consumers | **ACCEPT** | Correct inversion placement. Every retained interface has a named call site. See D12/D15 for the *content* corrections, which do not disturb the placement rule. |
| **D04** `Scope` = narrow `database/sql` executor; `Runner.Within` app-owned lifecycle | **ACCEPT WITH CORRECTION** | Lifecycle ownership and the non-escape/rollback/panic law are correct and match T8-B §8.1. Two corrections: freeze a typed platform-only transaction-binding accessor (**B1**); declare the `database/sql`-family restriction it imposes on T8-D instead of implying it. |
| **D05** no `*sql.Tx`/`pgx.Tx` on owner public signatures | **ACCEPT** | Preserves provider-neutrality at the semantic boundary. Independent of B1, which concerns the *platform* side, not owner signatures. |
| **D06** owner-local evidence → mechanical app mapping → `AppendIn` | **ACCEPT WITH CORRECTION** | Direction, ownership and the commit-only-after-append law match T3 §12 and T8-B §8.2, and "zero or more" makes the no-op-repeat case representable. Correction: the Audit **read** contract is entirely absent (**B2**). |
| **D07** `Subject`/`Check`/`Decision` + `Decide`/`DecideIn`/`DecideMany`/`DecideManyIn` | **ACCEPT WITH CORRECTION** | Sole-authority and default-DENY laws are correctly preserved; `DecideMany` is genuinely the same evaluator in bulk, not a second ruleset. Correction: no scope-enumeration query exists, so T6 §8 and cursor pagination have no legal path (**B3**); state index correspondence (**L2**). |
| **D08** `SecuritySubject`/`SecuritySubjectIn` as canonical grant-subject source | **ACCEPT WITH CORRECTION** | The apparent duplication between `organization.SecuritySubject` and `authorization.Subject` is correct, not a defect: Organization owns the *fact*, Authorization owns what it *means*; a shared type would create the sixth home D01 rejects. Correction: the contract cannot distinguish a plain subject read from an eligibility-**serializing** read required by T3 §11 (**M1**). |
| **D09** closed ControlledDocs access-fact vocabulary | **ACCEPT** | `AccessFacts{PredicateKnown, PredicateOK}` preserves domain meaning without becoming a policy language; unknown → `Provided=false` → DENY is fail-closed and matches T8-B §8.3. Application cannot weaken requiredness because Authorization owns the static requiredness rule — verified as non-duplicative of T3, since T3 §9 states the predicates while Authorization states *which permissions demand one*. |
| **D10** request-scoped `EnabledGroupMembersResolver` | **ACCEPT** | Correctly placed: ControlledDocs owns *when*, Organization owns *who*, application owns neither. It is not a service locator — single named capability, passed per invocation, no registry, no lookup-by-type. Type coupling is `uuid.UUID` + `txscope.Scope` only, so zero owner-import topology holds. |
| **D11** empty GROUP snapshot stays empty; no fallback/reassign | **ACCEPT** | Independently verified against upstream, as instructed. T2 §6: "No baseline reassign/overseer engine"; bounded recovery is "withdraw → DRAFT → fix route → resubmit", with the dedicated withdrawal rule for obsolescence. The candidate's conclusion is exactly supported; **no reopen required**. |
| **D12** Authentication-owned `ProviderClient` using raw/primitive data | **ACCEPT WITH CORRECTION** | Ownership split is right. Correction: type the verified protocol coordinates (issuer **and** subject) and return bounded typed directory refs per T6 §4, keeping the residue opaque (**M3**). |
| **D13** Organization role-target / owner-eligibility / deletion-dependency facts | **ACCEPT** | Verified fail-closed and non-deciding. `ResponsibleOwnerEligibilityIn` returns exactly D4 (existing User + same Company + ENABLED) and nothing more; Group deletion stays Organization-owned with explicit `Resolved` markers for both foreign sources, matching T3 §8's four live dependencies; RoleAssignment stays Authorization-owned. Application maps, never decides. |
| **D14** opaque preflight proof/candidate values | **ACCEPT WITH CORRECTION** | Unexported-field construction control correctly models T4-F's unforgeable binding. Correction: the claim's **liveness** (consume / release / bounded expiry) and its interaction with GC are unfrozen (**B4**). |
| **D15** consumer-owned `ManagedContent` + `MalwareInspector` | **ACCEPT WITH CORRECTION** | Provider neutrality holds — no bucket/key/version/ETag, server-derived SHA-256/size/ContentFormat, client hash demoted to hint per T4-E. Omitting `OpenRange` is correct (T4 marks it optional). Corrections: `DeleteReclaimable` and claim lifecycle are required by T4-D/T4-F (**B4**); add byte-identity correlation to inspection (**M4**). |
| **D16** consumer-owned `OfficialRenditionRenderer` | **ACCEPT** | Correct consumer (the application leaf owning rendition work), correct minimalism (`io.Reader → io.ReadCloser`), no renderer/provider id becomes semantic identity, matching T5 §2 and T8-B §7.4. The `requiredFormat string` is adequately constrained because T5-B fixes the single Launch value; a closed owner-local enum would be marginally stronger but is not material. |
| **D17** one named `OfficialRenditionIntentSink`, River hidden | **ACCEPT WITH CORRECTION** | Naming and hiding are right, and the transaction-coupled property is real — confirmed from vendored River source and current River documentation. Correction: the sink cannot obtain River's required concrete transaction from `Scope` (**B1**). |
| **D18** no generic EventBus/outbox | **ACCEPT** | Verified against T5-B/T5-K: exactly one activated durable-effect family; notifications, search refresh and IdP disable are all OFF. Generic infrastructure has no current consumer — Method §YAGNI applies cleanly. |
| **D19** `BeginIn`/`CompleteIn` in one Scope; no target `FailReplay` | **ACCEPT WITH CORRECTION** | The no-`FailReplay` derivation is sound for the crash/rollback path and I confirm it. Correction: the concurrent-loser path T6 §18 ratifies is unspecified and can abort the caller's transaction (**B5**). |
| **D20** operation-local versioned application `ReplaySnapshot`, opaque platform storage | **ACCEPT WITH CORRECTION** | Direction is right and prevents wire authority leaking inward; explicit versioning is the correct answer to deployment evolution. On whether T6 demands byte-for-byte replay: T6 §18 requires a result "sufficient for exact status/body replay", which a deterministic versioned mapping satisfies — so the snapshot approach is adequate, and this is **not** a hidden wire model, because it holds semantic result fields, not encoded responses. Correction: the assigned PII-free-vs-purge decision is not made (**M5**). |
| **D21** replay rechecks live disclosure authority, never re-executes | **ACCEPT** | Matches T6 §19 exactly, including the harder half — not rejecting a historically successful command merely because original preconditions no longer hold. Reusing each operation's own disclosure path instead of a generic replay ACL is the correct avoidance of a second evaluator. |
| **D22** owner facts + application composition + `DecideMany` read law | **ACCEPT WITH CORRECTION** | No foreign SQL, no persistent duplicate truth, no hidden Search owner — correctly consistent with T5-F/T6 §6 keeping materialization OFF, and Search stays a canonical PostgreSQL query/view privately owned by ControlledDocs. Corrections: scope enumeration and pagination coherence (**B3**); Audit read (**B2**); rendition content read (**M6**). |
| **D23** producer-owned errors; no common errors package | **ACCEPT** | Categories listed are sufficient for T8-E's RFC 9457 mapping without owners importing wire errors. Correctly avoids a shared package that would become a sixth home. |
| **D24** external/provider execution outside semantic Scope | **ACCEPT** | Inside/outside classification (§19) is complete and matches T2 §12, T4-E, T5-C. "External call required for success inside Scope is forbidden" is the right absolute form. |
| **D25** reuse only where the T8-A gate passes; tests grant no entitlement | **ACCEPT WITH CORRECTION** | Disposition table is honest and mostly well-argued. Correction: the `db.Tx` row claims "PRESERVE PROPERTY", but the property actually carried forward includes the unprovable downcast this shape forces (**B1**) — the reuse gate was applied to the shape's ergonomics, not to its enforcement weakness. |

```text
ACCEPT                    14
ACCEPT WITH CORRECTION    11
REJECT                     0
```

No decision is rejected. Every correction is bounded and additive within the selected model.

---

## 9. New findings beyond the candidate's own review contract

The candidate's §29 listed fifteen attack points. Findings **B2**, **B3**, **B4** (the GC/claim half), **M1**, **M2**, **M5** and **M6** are not reachable from that list: they are omissions of *required operations and laws*, whereas §29 mostly asks whether the *present* contracts are correct. That asymmetry is itself worth recording — a self-authored review contract tends to enumerate what was built, not what was skipped.

The most consequential class:

```text
the candidate verified its contracts against its own census
it did not verify its census against the ratified T6 §29 operation list,
T6 §24's If-Match list, T4-D's minimum contract, or T5-J's activated GC family
```

---

## 10. Strongest alternative Global Maximum considered

I constructed and attacked one alternative seriously enough to try to defeat the candidate on its own strongest seam (B).

**Alternative — opaque transaction token + owner-private executor adapter.**

```text
txscope.Scope carries no SQL methods at all
owner-private persistence obtains an executor through a platform accessor
application structurally cannot execute SQL: it has no method to call
```

Attraction: it would convert the candidate's §4.3 *lint rule* into a *type-system* guarantee, which Method §Enforcement prefers.

**It does not dominate, and I reject it.** Go cannot express the required visibility. Everything under `internal/` is importable by the entire module, and Go has no friend/package-set visibility. The accessor must live where owners can reach it (`platform/txscope`, the only platform edge T8-B §9.1 grants owners) — and `application → platform/txscope` is likewise allowed, so application can call the same accessor. The prohibition remains lint-enforced either way; the alternative only relocates the capability while adding an adapter layer and an extra indirection per query.

Therefore the candidate's choice — reuse the standard-library executor shape and enforce the application prohibition with the closed-world `cilint` catalog T8-B §10 already mandates — is the **correct available maximum** for that sub-property. The residual defect is narrower and separately fixable: the missing platform-side binding contract (B1).

I also re-tested the four classes the candidate rejected (producer interfaces everywhere; consumer interfaces per owner call; shared `internal/contracts`; generic UnitOfWork/EventBus/policy language). Each reproduces a defect ratified authority forbids — implementor-side ceremony, interface explosion, a sixth semantic home, or infrastructure with no current consumer. None survives.

```text
No materially superior contract-placement model was found.
```

---

## 11. Global Maximum confirmation

```text
SELECTED GLOBAL MAXIMUM CLASS   CONFIRMED
```

The authority-aligned hybrid model is confirmed as the Global Maximum for T8-C, subject to the corrections above. Method outcome, restated in Method §4 vocabulary:

```text
CURRENT STRUCTURE CONFIRMED  (model class + T8-B topology)
+ bounded corrections        (five BLOCKER, six MAJOR)
```

The candidate's Structural Inversion (§24) is sound for the *ownership* conclusions it lists, and I independently reach the same result: those conclusions survive substituting pgx, another provider, another IdP and no legacy IAM module. One qualification — the inversion was **not** applied to the executor *shape*. Under genuine inversion (a pgx-native current codebase), no one proposes `QueryRowContext(...) *sql.Row`. That is exactly where D04 needed the test most, and it is why B1 went unnoticed.

---

## 12. T8-B reopen determination

```text
T8-B REOPEN REQUIRED:  NO
```

No finding contradicts a T8-B ruling. Each is an unfrozen or incomplete T8-C instantiation *within* the ratified topology:

```text
B1  needs a contract inside platform/txscope        — an existing platform home
B2  needs a method on the existing Audit owner      — no new surface, no new owner
B3  needs a method on the existing Authorization    — same evaluator, same owner
B4  needs operations on an existing platform port   — mechanism stays mechanism
B5  needs semantics on an existing platform contract
M1  needs vocabulary on an existing Organization contract
```

None requires a second public package path, a new semantic owner, a direct owner→owner import, foreign SQL or hidden shared write authority. T8-B §12's reopen triggers are not met — in particular, no accepted seam has been shown unable to express required semantics without owner leakage.

---

## 13. Upstream T1 → T7 reopen determination

```text
UPSTREAM SEMANTIC REOPEN REQUIRED:  NO
```

Every finding is resolved by **conforming to** upstream authority, never by changing it. Explicitly tested and found *not* to require reopen:

```text
T2 §6   empty GROUP snapshot / no reassign engine    candidate conclusion CORRECT
D4      responsible-owner eligibility                 candidate reproduces it exactly
T3 §14  Audit read visibility                         satisfiable once B2 is frozen
T4-D    ManagedContent minimum contract               satisfiable once B4 is frozen
T5-B    single durable-intent family                  candidate conclusion CORRECT
T6 §18  no public IN_PROGRESS replay state            candidate conclusion CORRECT
```

---

## 14. T8-D trespass

```text
CLASSIC T8-D TRESPASS:  NONE
REVERSE-DIRECTION CONSTRAINT: ONE (undeclared)
```

The candidate is disciplined about persistence: §4.4 and §26 correctly defer schema, indexes, isolation options, lock clauses, lock order, serialization roots, `*sql.Tx` realization, idempotency schema/retention, Audit tables, River tables and managed-content mechanism persistence. `BeginIn`/`CompleteIn` name a *coordination contract*, not a schema, and stay within T8-B §7.6's grant of the replay contract to T8-C.

The one issue runs the other way: D04's executor shape **constrains** T8-D to the `database/sql` family without saying so (B1). A stage that narrows a later stage's option set must declare it. Two consequential clarifications belong with it:

```text
B5's resolution (savepoint-internal vs caller-retry) is a CONTRACT choice
    even though its SQL is T8-D — freeze the semantics, not the statements
M1's lock PLACEMENT is T8-D, but the contract must distinguish which reads
    require serialization, or T8-D must re-derive T3 §11 (duplicated semantics)
```

---

## 15. T8-E trespass

```text
T8-E TRESPASS:  NONE
```

The candidate holds the wire boundary well, including §16.4's explicit refusal to let T8-E force generated wire DTOs into application, and §20.1's refusal to let owners import generated HTTP errors. `ReplaySnapshot` is a semantic result model, not an encoded response, so it is not a hidden wire model.

One forward risk, not a trespass: leaving the If-Match precondition token unfrozen (M2) creates the conditions for T8-E to become the de facto owner of concurrency control. Freezing the owner-side token in T8-C forecloses that.

---

## 16. Operation-census completeness verdict

```text
OPERATION CENSUS:  INCOMPLETE
```

I walked the ratified T6 §29 census plus the T2/T3/T4/T5 flows. The large majority map cleanly, including families worth naming as genuinely covered:

```text
COVERED   login/callback (BeginBrowserLogin / VerifyBrowserCallbackPreflight,
          correctly outside /api/v1 per T6 §2.2), session resolve/revoke,
          provider-subject search, provider-binding replace + session invalidation;
          User create/profile update/profile erase/eligibility;
          offboarding as one transaction matching T3 §10 step for step;
          Areas / Groups / GroupMembership / RoleAssignments;
          DocumentType + governance + eligible-templates + numbering preview
          + template configuration read;
          blank/template create, next Revision, DRAFT read/update/upload,
          SUBMIT, feedback, ACCEPT/RETURN, withdrawal, cancellation,
          obsolescence request/withdraw, responsible-owner change;
          Library / My Work / Official / Work / History / Governance Case lenses;
          document-creation options; allowed_actions via the same evaluator;
          idempotent POST replay contract.
```

Material omissions, by the prompt's own test — a Writer would have to invent the contract:

```text
MISSING   Audit reads — GET /api/v1/audit/events has no owner contract        (B2)
MISSING   document-creation options cannot be built legally: no scope
          enumeration exists, and the two workarounds are an unbounded
          candidate set or a second evaluator                                  (B3)
MISSING   exact rendition read — GET /official-renditions/{id}/content         (M6)
MISSING   managed-content cleanup + admission-claim lifecycle + the whole
          T5-J GC reconciliation family                                        (B4)
UNFROZEN  If-Match precondition routing across 11+ resources                   (M2)
UNSTATED  obsolescence completion host; responsible-owner and template-role
          reads; provider-binding disable                                (L3/L4/L5)
```

Exact source reads and numbering preview are covered; consistent cross-owner pagination is **not** achievable under the current read law (B3).

---

## 17. Selective-reuse and reference-check verdict

```text
SELECTIVE REUSE:  PASS WITH ONE CORRECTION
REFERENCE CHECK:  PASS
```

The candidate maintains the required distinction between *reuse this property* and *reuse this existing code/contract*: `TxRunner` lifecycle property preserved while its exact contract is rewritten; River's transactional-insert property preserved while its client/driver/transaction types are excluded; OpenAPI contract-first property preserved with exact content deferred; IAM `authz.Require`, objectstore and the idempotency middleware correctly rewritten rather than inherited. "No current interface survives merely because tests exist" is the correct T8-A posture.

The one correction is the `db.Tx` row (D25/B1): what was preserved is not only a proven narrow shape but also the enforcement weakness that shape forces on every mechanism needing a concrete transaction.

Reference handling is sound. External sources were used as falsification evidence and never to create Product requirements; no capability absent from the Product Contract was imported. I independently re-verified the load-bearing external facts rather than accepting the candidate's summary:

```text
River v0.37.1 pinned                       go.mod:28-30, 74-75  CONFIRMED
InsertTx is transaction-coupled            vendor client.go:1765 + Context7  CONFIRMED
InsertTx requires a CONCRETE driver tx     riverdatabasesql:94 UnwrapExecutor(*sql.Tx)
                                           — material, and NOT reflected in the candidate
PostgreSQL READ COMMITTED posture          T2 §3 ratified; candidate consistent
OIDC issuer+subject as protocol coordinates  supports typing subject (M3)
S3 conditional create-if-absent             supports create-once; provider detail
                                            correctly kept out of the owner contract
Range read optional                         T4 §6; candidate's omission CORRECT
```

Context7 was used for River, the one external behavior that is load-bearing to a decision (D17), per AGENTS.md. Vendored source at the pinned version was treated as the stronger primary evidence.

---

## 18. Whether another Fable round is materially required

```text
ANOTHER FULL INDEPENDENT ROUND:  NO
BOUNDED DELTA RE-REVIEW:         CONDITIONAL
```

Every finding is stated with its authority citation, its failure mode and a correction direction, and each is independently checkable without re-deriving the model. The model class is confirmed, so the corrections do not reopen the design space.

A bounded delta re-review is warranted only if the Lead's corrections to **B1**, **B3** or **B5** change the model class — for example if B3 is answered by anything other than an Authorization-owned enumeration, or if B1 is answered by abandoning the executor shape. Corrections that stay additive within the selected model need adjudication, not another round.

---

## 19. Whether final Lead adjudication may proceed

```text
LEAD ADJUDICATION MAY PROCEED:  YES
T8-C PROMOTION:                 NOT YET
T8-D:                           REMAINS NOT OPEN
IMPLEMENTATION:                 REMAINS BLOCKED
```

Per Method §3, these findings are evidence, not requirement authority. The Lead must first classify each against current authority. On my own classification:

```text
DEFECT AGAINST EXISTING AUTHORITY (correctable in place, no new requirement)
  B1 B2 B3 B4 B5 M1 M2 M3 M4 M6 L1 L2 L3 L4 L5
     each cites a ratified T2/T3/T4/T5/T6/T8-A/T8-B rule the candidate does not
     yet satisfy, so correcting it restores conformance rather than adding scope

RETURNS TO DECISION (assigned but unmade — not a correction)
  M5 T8-B §7.6 assigns the PII-free-vs-purge replay decision to T8-C; it must be
     decided as a decision, with alternatives and a reopen trigger, not patched
```

No finding creates new authority or new Product requirement, and none should enter as a disguised correction.

---

## 20. Reviewer summary

```text
REVIEWED HEAD                3a882c4daa3707cedbe6ed1019c2001eabebca34
VERDICT                      APPROVE T8-C GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
BLOCKER / MAJOR / LOW        5 / 6 / 5
DECISIONS                    14 ACCEPT, 11 ACCEPT WITH CORRECTION, 0 REJECT
GLOBAL MAXIMUM               CONFIRMED
T8-B REOPEN                  NO
UPSTREAM T1→T7 REOPEN        NO
T8-D TRESPASS                none classic; one undeclared reverse constraint (B1)
T8-E TRESPASS                NO
OPERATION CENSUS             INCOMPLETE
SELECTIVE REUSE / REFERENCES PASS (one correction)
ANOTHER FABLE ROUND          NO (bounded delta only if a correction changes the model class)
LEAD ADJUDICATION            MAY PROCEED
```

The candidate is strong work. It selects the right model, rejects the right alternatives for the right reasons, holds the wire and persistence boundaries with discipline, and reaches several conclusions I independently confirmed against upstream authority rather than merely accepting — the empty GROUP snapshot, the single durable-intent family, the absence of a public IN_PROGRESS replay state, and the deliberate omission of range read.

Its weakness is not judgment but **coverage**: it verified its contracts against its own census instead of verifying its census against the ratified operation lists. That is what left an Audit read, a scope enumeration, a content-cleanup family and a precondition vocabulary unfrozen, and it is what allowed a Structural Inversion applied to ownership to miss the one seam where the shape itself was inherited.

```text
T8-C remains ACTIVE and NON-AUTHORITATIVE.
This artifact is reviewer evidence only.
```
