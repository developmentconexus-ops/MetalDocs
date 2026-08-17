# MetalDocs — Whole-System Implementation Conformance — Fable Review Request

> **Status:** REVIEW REQUEST — **EVIDENCE ONLY — NOT TARGET AUTHORITY**
> **Date:** 2026-08-17
> **Repository:** `developmentconexus-ops/MetalDocs`
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131
> **Expected review baseline HEAD before this request:** `42c61644102d249fe95892c530c94a709b5c4c31`
> **Method:** DevelopmentConexus Engineering Method v1.0.0
> **Implementation gate:** **CLOSED — design/documentation only.**
> **Purpose:** perform one independent, repository-wide conformance audit of the current implementation against **promoted authority and validated solutions**, while distinguishing expected legacy/cutover debt from active defects or unauthorized implementation drift.

---

# 0. Why this review exists

The architecture program has matured enough that local micro-reviews are no longer the right assurance level. This review must inspect the **whole implementation surface as one system** and answer:

> Is the current MetalDocs implementation, schema, API, frontend, tests, runtime and operational posture consistent with the promoted architecture and engineering standards — or, where it is intentionally still legacy because implementation is blocked, is the mismatch explicitly understood, safely bounded and routable to cutover without hidden contradiction?

This is **not** a requirement that the current legacy runtime already implements the target architecture. The implementation gate remains closed. A mismatch with a promoted target is not automatically a defect: it may be expected migration debt. The reviewer must classify it correctly.

The review must detect two opposite failure modes:

1. **False failure:** treating every legacy current-state difference as a bug even though target implementation has not started.
2. **False success:** dismissing an active security/correctness defect or unauthorized target drift as “legacy” merely because a redesign exists.

---

# 1. Mandatory read order / authority chain

Start fresh from repository truth. Do **not** use conversation history as authority.

Read in the exact order routed by `AGENTS.md`, including at minimum:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/cohesive-platform-redesign.md`
5. `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
6. `wiki/architecture/r10-technical-architecture.md`
7. review artifacts **only** when needed to understand how a promoted decision was challenged
8. current code/schema/OpenAPI/frontend/tests/runtime as current-state evidence

Promoted authority is binding. Candidate/review artifacts that have not been promoted are **evidence only**.

Especially:

- `R9.5` is frozen / GCR-refined / single-company-refined;
- `R10-A` ownership topology is closed;
- `R10-B1` relational substrate is closed / single-company-restructured;
- `R10-B2-1` Authentication binding/ApplicationSession/assurance is closed / approved;
- `R10-B2-2` is not promoted at the review-request baseline; its candidate and independent review are evidence only;
- implementation remains blocked.

Never use this review request itself as architecture authority.

---

# 2. Method — non-negotiable

Apply the DevelopmentConexus Engineering Method v1.0.0 proportionally but rigorously:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Future Cost
→ Authority / Boundary
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Use the Structural Inversion Test repeatedly:

> If the current implementation were the opposite in every relevant respect, which conclusions would still follow from the promoted authority?

Do not preserve a current mechanism because it exists. Do not delete a current control merely because the target mechanism changes.

Mechanism ≠ Authority.

---

# 3. Required finding classification

Every material finding MUST receive exactly one primary classification:

```text
CONFORMANT
  Current implementation already matches the promoted invariant/property.

ACTIVE DEFECT
  Current runtime/security/correctness behavior is wrong or unsafe NOW,
  independent of future migration.

UNAUTHORIZED DRIFT
  Implementation has started encoding an unpromoted/rejected target decision,
  violating the implementation gate or current authority.

EXPECTED LEGACY / CUTOVER DEBT
  Current implementation intentionally differs from promoted target because
  target implementation has not started; the mismatch is real but belongs to
  planned migration/cutover and is not itself a current correctness failure.

MISSING PROOF
  The implementation may be correct, but the claimed invariant lacks adequate
  tests/runtime evidence/constraint proof.

DEFERRED BY AUTHORITY
  Capability is explicitly outside V1 or routed to a later R10 stage.

OUTSIDE PROMOTED AUTHORITY
  Evidence concerns a candidate/unpromoted decision; report it but do not call
  it a target violation.
```

Severity is separate from classification:

```text
BLOCKER
  active safety/correctness/authority failure requiring stop or immediate action;
  or a hidden structural contradiction that makes the validated target impossible.

MAJOR
  material architecture/cutover/proof gap that must close before implementation
  planning/promotion/merge as applicable, but does not require emergency runtime action.

LOW
  bounded hardening, wording, proof coverage or implementation-quality issue.
```

Do **not** inflate expected legacy debt to BLOCKER merely because target differs.

---

# 4. First mandatory question — has unauthorized implementation started?

Before auditing details, establish the branch history and implementation scope.

Verify:

```text
git status --short
git rev-parse HEAD
git log --oneline --decorate --graph --max-count=<sufficient>
git diff / compare relevant promotion points
```

Determine whether any commits after promoted architecture decisions changed:

```text
product code
DB schema/migrations
OpenAPI/generated contracts
frontend
Keycloak/deployment configuration
runtime scripts
```

with intent to implement unapproved B2/B3/B4/etc decisions.

Required answer:

```text
UNAUTHORIZED TARGET IMPLEMENTATION DRIFT = YES | NO
```

If YES, enumerate exact commits/files/decisions and severity.

Documentation/review artifacts are not product implementation.

---

# 5. Promoted architecture conformance matrix — mandatory

Build a matrix with one row per promoted invariant/property and columns:

```text
Authority requirement
Current implementation evidence
Classification
Severity
What must change / survive
Owning future stage (if debt)
Proof
```

At minimum cover all sections below.

---

# 6. Single-company deployment / Tenant-root conformance

Promoted V1 target includes:

```text
one company per deployment
same codebase/build/migrations for every deployment
no customer-specific forks
exactly one Tenant root per product DB
Tenant.id immutable UUID
expected_tenant_id ↔ DB Tenant.id startup/readiness fail-closed handshake
no Tenant customer ACTIVE/SUSPENDED/ERASED lifecycle V1
no customer deletion/tombstone family V1
```

Audit current implementation for:

- all Tenant representations;
- root multiplicity constraints;
- startup/readiness identity checks;
- mutable Tenant UUID paths;
- slug/customer routing assumptions;
- customer lifecycle/deletion/tombstone runtime;
- customer-specific hardcoding;
- implicit multi-company assumptions.

Known target distinction:

```text
Tenant = semantic singleton company root / whole-company scope target
Tenant != DB partition dimension V1
```

Do not confuse semantic Tenant references with legacy partition plumbing.

Required sub-verdicts:

```text
Tenant root identity safety NOW
Tenant root target migration debt
customer-fork risk
handshake proof status
```

---

# 7. Relational substrate / DB constraint conformance

Promoted B1 target:

```text
one MetalDocs product-state PostgreSQL DB
canonical product schema = metaldocs
ordinary entity: id UUID PRIMARY KEY
no universal tenant_id/company_id/deployment_id partition column
no composite (tenant_id,id) PK/FK law
ordinary typed FK target_id → target(id)
cross-owner FK DELETE/UPDATE = RESTRICT | NO ACTION only
cross-owner CASCADE/SET NULL/SET DEFAULT forbidden
no universal polymorphic business registry
technical IDs = UUID
business instants = TIMESTAMPTZ
canonical SHA-256 = BYTEA + length=32
real unknown = NULL
READ COMMITTED default
no provider DB atomicity / XA / 2PC
```

Audit:

- current PK/FK shapes;
- all tenant-qualified composite keys;
- universal tenant/company columns;
- cross-owner FK actions;
- polymorphic registries;
- JSONB escape-hatch usage;
- timestamps/hashes;
- provider/external IDs used as PKs;
- uniqueness scopes that still encode pooled tenancy assumptions.

Important: current `tenant_id`/RLS presence is likely EXPECTED LEGACY/CUTOVER DEBT, not automatically ACTIVE DEFECT. But identify any place where it causes a current correctness/security problem.

Required mechanical searches include at least:

```text
tenant_id
company_id
organization_id
deployment_id
FORCE ROW LEVEL SECURITY
ENABLE ROW LEVEL SECURITY
current_setting(
set_config(
tenant context
BYPASSRLS
CASCADE
SET NULL
SET DEFAULT
```

Use semantic inspection, not grep counts alone.

---

# 8. Database security / serving-role conformance

Promoted target:

```text
ordinary serving DB role = NOSUPERUSER + non-owner of protected product tables
DDL/object ownership = separate
maintenance/migration/restore identity = separate non-serving trust surface
maintenance identity never request-reachable
canonical Authorization is application/domain authority
no Tenant/Area/Role/Permission RLS target policy engine
```

Audit:

- actual connection roles/configuration;
- table ownership;
- migration users;
- whether serving identity can bypass intended constraints;
- current RLS behavior and whether removing it later would silently remove unrelated protection;
- any request-reachable maintenance/elevated path;
- any `system_admin` or DB bypass conflation.

Classify current RLS carefully:

```text
current safety control worth preserving until cutover
vs
legacy target debt
```

Do not recommend removing a current safety control before replacement enforcement exists.

---

# 9. Authentication / Keycloak / B2-1 conformance

Promoted Authentication target is binding:

```text
Keycloak = V1 AuthN provider
ProviderSubjectBinding + ApplicationSession = exactly two Authentication semantic families
stable provider identity = issuer + subject
email/username/display name never technical identity
provider roles/groups/orgs/permissions/arbitrary claims never canonical AuthZ
no provider-role/claim-to-permission bridge
no local credential/password/MFA/lockout authority target
no provider DB atomicity
```

Promoted `ProviderSubjectBinding` target:

```text
id UUID PK
user_id → Organization.User
issuer
subject
created_at
disabled_at?
UNIQUE(issuer,subject)
UNIQUE(user_id) WHERE disabled_at IS NULL
mapping fields immutable
reversible acceptance
```

Promoted `ApplicationSession` properties:

```text
local opaque high-entropy bearer
raw bearer never stored
DB row disclosure not replayable
finite absolute expiry
server-side revocation
multiple Sessions/User allowed
Session → Binding → User
no tenant dimension
no duplicated persisted user_id
no AuthZ snapshot
no provider token authority
fresh-auth bounded; bare non-NULL latest_* never sufficient
forced reauth pins same issuer+subject
```

Audit current auth code, schema, tests, cookies/tokens, Keycloak integration/configuration for:

- local passwords/hash algorithms/lockout;
- current Session shape;
- token opacity/entropy/hash storage;
- direct provider JWT use;
- provider claims influencing AuthZ;
- provider role/group mapping;
- tenant/session coupling;
- current revocation properties;
- fresh-auth representation;
- provider disable vs live Session behavior;
- login/offboarding races;
- Keycloak outage behavior;
- provider provisioning/binding reconciliation.

Known current tests around opaque sessions may prove useful properties; retain good safety properties even if surrounding architecture changes.

Required sub-verdicts:

```text
opaque-session safety property NOW
credential-authority debt
provider anti-corruption status
B2-1 migration gap
active auth security defects
```

---

# 10. Organization / Authorization implementation audit

Use only **promoted** Organization/AuthZ semantics as target law. Unpromoted B2-2 evidence may be cited as evidence, never as authority.

Promoted/frozen facts include:

```text
Organization owns Tenant / Area / User / Group / GroupMembership
Groups flat V1
Area is organizational truth reused by Document ownership, scoped AuthZ and Approval actor resolution
five roles exactly:
  tenant_owner
  area_manager
  author
  approver
  viewer
43 permissions exactly in frozen R9 + R9.5 catalogs
RoleAssignment:
  subject = User | Group
  scope = TenantScope | AreaScope
additive/default-deny
no tenant_owner bypass
no generic ACL/ReBAC graph
no nested groups
domain relationship predicates remain domain-owned
PlatformOperator/SystemPrincipal outside company RBAC
```

Audit current implementation for:

- old 8-role catalog;
- `system_admin`, `qms_admin`, `editor`, `signer`, `area_admin` behavior;
- capability short circuits;
- tenant-owner/system-admin bypasses;
- user/group role tables;
- user-process-area grants;
- group nesting/dynamic groups;
- provider roles;
- area scope representation;
- current authorization evaluation;
- cached/effective permission tables;
- frontend role assumptions;
- OpenAPI role enums.

Treat target mismatch as EXPECTED LEGACY/CUTOVER DEBT unless it causes an active defect or unauthorized post-promotion drift.

Required answer:

> Does any current mechanism encode semantics that would make migration to the validated 5-role/43-permission/additive-default-deny target structurally difficult or ambiguous?

If yes, name the root cause and migration obligation.

---

# 11. Cross-owner transaction / Audit / durable-intent conformance

Promoted B1 law:

```text
single-owner → owner service boundary
cross-owner frozen atomic use case:
  composition opens one local PostgreSQL transaction
  published owner seams share it
  one COMMIT/ROLLBACK

required Audit append = same commit as authoritative mutation
required durable external-effect intent = same commit when future effect is necessary
provider/external effect = after commit
```

Audit representative critical flows across the whole product, not only auth:

- User/admin mutation + Audit;
- current Area archive + Audit;
- Document/Revision mutations;
- Submission;
- Approval decisions/reassignment/cancel;
- Release/effectivity;
- Distribution audience snapshot/acknowledgement;
- LegalHold/disposition;
- Artifact confirmation;
- external repository publish/import;
- notification intent;
- provider provisioning/disable;
- Historical Migration.

For each material mutation classify:

```text
same-commit proof exists
same-commit missing
nested hidden commit
provider call inside DB tx
best-effort audit
best-effort async intent
```

A current transaction implementation that already preserves stronger safe semantics is valuable evidence even if target owner packages change.

---

# 12. Async / worker / external-effect conformance

Promoted target retains only independently justified reliability machinery:

```text
transactional outbox / durable intent
idempotency
claim/lease
retry
DLQ
truthful external-effect state
```

Removed target assumptions:

```text
cross-customer Tenant enumeration
tenant context seeding
tenant_id customer routing
```

Audit current jobs/workers/outbox for:

- tenant enumeration;
- customer routing metadata;
- business payload duplication in claim tables;
- idempotency;
- leases/claims;
- retry/DLQ;
- truthful uncertain provider/external outcome;
- provider calls inside transactions;
- duplicate business authority in job payloads;
- search/notification workers bypassing owner boundaries.

Classify which current reliability mechanisms must survive cutover.

---

# 13. Privacy / retention / restore conformance

Promoted V1 privacy posture:

```text
customer-company deletion lifecycle = deferred
User/data-subject privacy = live requirement
User offboarding can revoke Sessions / prevent access
human-readable/user enrichment must be separable from immutable evidence
surviving immutable Audit state = PII-minimized/non-PII skeleton
restore must not silently resurrect lawfully erased PII
no generic privacy workflow implied
no mandatory Tenant DEK/KEK/crypto-shred without named immutable Target Data
```

Audit current schema/runtime/tests/backups for:

- PII embedded in immutable Audit rows;
- actor names/emails copied into immutable evidence;
- user profile vs stable identity coupling;
- local auth/session PII;
- delete/anonymize flows;
- retention/LegalHold collisions;
- restore reconciliation;
- existing DEK/KEK/crypto-erasure machinery;
- customer erasure/tombstones;
- data resurrected from backups;
- provider-subject retention.

Do not give legal advice. Reason architecturally about separability, historical validity and restore behavior.

Required answer:

```text
current PII immutable-coupling risks
current lawful-erasure capability/proof
restore non-resurrection proof status
crypto-erasure target entitlement = YES | NO
```

---

# 14. Artifact / storage / malware integrity conformance

Promoted/frozen properties include:

```text
Artifact = immutable exact-byte identity
canonical SHA-256
opaque immutable object keys
no overwrite
provider-independent identity
ManagedArtifactStore port/conformance
Local dev/test; AWS S3 reference production
production malware inspection before confirming untrusted bytes
no confirmed orphan Artifact
Tenant/company key prefix not a V1 isolation invariant
```

Audit current storage/artifact code/schema/tests for:

- hash representation and validation;
- mutable objects/overwrite;
- provider URL/version entering business identity;
- object key tenant dependence;
- MinIO assumptions/entitlement;
- staging → scan → confirm ordering;
- fail-open scanning;
- orphan confirmed artifacts;
- restore hash verification;
- provider-specific semantics leaking into domain state.

Route detailed scanner/provider ordering to R10-C when not yet promoted; distinguish missing design from current active unsafe behavior.

---

# 15. Controlled Information / Submission / Approval / Release conformance

Audit implementation against frozen R3–R9.5 semantics, including at minimum:

```text
Document stable governed identity
Document code/type/Area stable V1
Revision states/lifecycle
DRAFT mutable persisted working truth
RevisionSubmission immutable exact attempt
Approval binds exact RevisionSubmission
Rendition binds exact Submission
Release binds exact Submission/digest
return/withdraw → same REV DRAFT; resubmit creates new Submission
strict Approval SoD
NoHumanApproval still creates immutable Submission
at most one EFFECTIVE + one open Revision V1
release atomically swaps EFFECTIVE/SUPERSEDED state
release = system-owned, no publish button
OfficialRepresentationPolicy = SourceOnly | RequireRendition(ContentFormat)
```

Look for current legacy behaviors that make future migration difficult:

- mutable submitted content;
- Approval tied to Revision instead of Submission;
- renderer output treated as approved authority;
- publish button semantics;
- PDF-only assumptions;
- duplicated Template lifecycle;
- REV reuse;
- approvals that can bypass SoD;
- release partial commits.

Again classify target mismatch correctly.

---

# 16. Documentary Context / Records / Distribution / Interchange conformance

Audit frozen semantics:

```text
Dossier = bounded documentary context, not ERP/PLM platform
Evidence/EvidenceType
RetentionBinding / RetentionExtension
LegalHold blocks disposition in live scope
DispositionRecord explicit
Distribution = obligation/acknowledgement, not AuthZ
release snapshots concrete audience users
AcknowledgementRecord explicit; notification read != acknowledgement
Historical Migration preserves uncertainty
Governed Subject Export
External Repository IMPORT_COPY / PUBLISH_COPY
Tenant Portability Export deferred
```

Look for:

- context granting access;
- Group membership rewriting historical distribution denominator;
- automatic deletion on retention expiry;
- holds not blocking disposal;
- imported history fabricated as native;
- external repository silent sync;
- tenant portability surfaces still live;
- customer-deletion coupling in Records.

---

# 17. API / OpenAPI / frontend conformance

Audit generated/public contracts and UI assumptions for semantic drift:

- old role enum vs 5-role target;
- system_admin bypass UX;
- tenant selector/company switching;
- customer lifecycle/delete/export actions;
- local password/MFA management;
- provider role/group concepts surfaced as product truth;
- Area archive/retirement behavior;
- current auth/session APIs;
- publish button vs automatic release;
- PDF-only assumptions;
- recognition of explicit acknowledgements;
- permission catalog mismatches;
- API fields that would force target persistence choices.

Classify:

```text
current API contract safe legacy
cutover-breaking external contract
unauthorized target drift
active incorrect behavior
```

Do not hand-edit generated contracts.

---

# 18. Tests / verification / proof quality

Run the repository verification authority when environment permits:

```text
go run ./tools/verify --profile=pr
```

Also run targeted read-only commands needed to validate findings. Do not weaken/bypass tests.

If runtime/startup checks are applicable and the environment supports them, run the repository-authorized runnable checks. If PowerShell/infrastructure is unavailable, state the limitation rather than guessing.

Audit whether tests prove actual invariants or merely mock success.

Required proof-quality categories:

```text
DB constraint proof
concurrency/race proof
transaction atomicity proof
real provider integration proof
negative/fail-closed test
privacy/restore proof
API contract proof
runtime/startup proof
```

For every promoted material invariant mark:

```text
PROVEN
PARTIALLY PROVEN
UNPROVEN
NOT YET IMPLEMENTED / CUTOVER DEBT
```

Do not claim green without fresh command output.

---

# 19. Current safety controls that must survive migration

Produce a dedicated list of current controls that are **good properties even if their current mechanism is legacy**, for example if evidence confirms them:

```text
opaque session token entropy/hash storage
server-side revocation
same-commit audit on critical mutations
row locking / OCC
RLS protection during legacy phase
idempotency/outbox reliability
malware fail-closed behavior
hash/no-overwrite storage properties
SoD enforcement
strict DB constraints
```

For each, state:

```text
property to preserve
current mechanism
target mechanism/owner
risk of accidental deletion during cutover
```

This section is mandatory. Redesign must not regress safety while removing legacy machinery.

---

# 20. R10-F cutover inventory — mandatory

Produce an exact implementation-debt inventory of current legacy structures that the validated target says must eventually be removed/migrated.

At minimum inspect for:

```text
universal tenant_id plumbing
Tenant RLS/context/GUC/routing
customer lifecycle/tombstones/portability
local credential/password/MFA/lockout
old 8-role/capability catalog
system_admin bypass
legacy user_process_areas role model
provider-role/claim mapping if any
Tenant DEK/KEK machinery without entitlement
tenant-namespaced storage assumptions
customer-routing async metadata
current module/package ownership mismatches
```

For each item state:

```text
source tables/files/APIs
current consumers
replacement validated authority
required data migration/backfill
safe deletion preconditions
proof before deletion
```

Do not design the full R10-F implementation plan; produce the conformance/cutover inventory only.

---

# 21. Adversarial end-to-end scenarios — mandatory

Walk at least these scenarios through current implementation and target authority:

1. Deployment starts with wrong database / Tenant root mismatch.
2. Valid Keycloak subject has no MetalDocs binding.
3. User offboarded while a login/session is concurrently being issued.
4. Provider account disabled while local Session already exists.
5. Role/grant revoked while Session remains active.
6. User profile PII erased while Approval/Audit/Document history remains.
7. Backup restore contains PII erased after the backup was taken.
8. Area becomes retired while Documents/grants/policies already reference it.
9. Group membership changes after Distribution release audience snapshot.
10. Submitted Revision returns for changes and resubmits.
11. Approval/Rendition/Release disagree on exact Submission/digest.
12. Release races or retries and risks two EFFECTIVE revisions.
13. LegalHold intersects retention/disposition.
14. Historical Migration has unknown actor/time/history facts.
15. External repository publish returns an uncertain timeout.
16. Malware scanner unavailable for an untrusted upload in production.
17. Object-store database row is leaked — can bearer/storage secret be replayed?
18. Serving DB role is compromised — what authority/DB protections remain?
19. Worker retries after process crash between DB commit and provider effect.
20. Legacy tenant/RLS machinery is removed — what unrelated safety property would be lost if cutover were naive?

For each scenario provide:

```text
current behavior evidence
target invariant
classification
failure mode
proof needed
future owner/stage
```

---

# 22. Structural inversion / subtractive pass

Answer both:

### Structural inversion

> If MetalDocs had been built from day one with the promoted single-company, Keycloak, B1 and B2-1 architecture, which current implementation mechanisms would still be necessary properties, and which exist only because of legacy topology?

### Subtractive

> What current machinery can eventually be deleted without weakening a validated property, and what apparently-legacy machinery actually hides an essential safety property that must be re-expressed before deletion?

This is a whole-system deletion test, not a file cleanup exercise.

---

# 23. Required overall verdict

Choose exactly one:

```text
A. IMPLEMENTATION CONFORMANCE BASELINE ACCEPTABLE — SAFE TO CONTINUE DESIGN

B. SAFE TO CONTINUE DESIGN WITH MATERIAL IMPLEMENTATION / CUTOVER FINDINGS

C. STOP — ACTIVE IMPLEMENTATION DEFECT OR UNAUTHORIZED AUTHORITY DRIFT REQUIRES ACTION
```

This verdict does **not** authorize implementation. `implementation = BLOCKED` remains true unless promoted authority later changes it.

Also answer separately:

```text
CURRENT IMPLEMENTATION FULLY MATCHES PROMOTED TARGET = YES | NO

UNAUTHORIZED POST-PROMOTION PRODUCT IMPLEMENTATION = YES | NO

ACTIVE SECURITY/CORRECTNESS BLOCKER NOW = YES | NO

SAFE TO CONTINUE R10 DESIGN = YES | NO

IMPLEMENTATION GATE MAY OPEN = NO   // unless existing authority itself says otherwise; reviewer may not open it
```

---

# 24. Required output structure

The final review artifact MUST contain:

1. bootstrap/read-order proof;
2. reviewed HEAD / branch history summary;
3. exact verdict;
4. BLOCKER / MAJOR / LOW counts;
5. unauthorized-implementation-drift verdict;
6. promoted-authority conformance matrix;
7. active-defect findings;
8. expected legacy/cutover debt findings;
9. missing-proof findings;
10. single-company/Tenant-root verdict;
11. B1 relational/database-security verdict;
12. Authentication/B2-1 verdict;
13. Organization/AuthZ current-state verdict;
14. transaction/Audit/durable-intent verdict;
15. async/external-effects verdict;
16. privacy/retention/restore verdict;
17. Artifact/storage/malware verdict;
18. Controlled Information/Approval/Release verdict;
19. Documentary Context/Records/Distribution/Interchange verdict;
20. API/frontend contract verdict;
21. verification/test evidence and exact command results;
22. current safety properties that must survive migration;
23. R10-F cutover inventory;
24. 20 mandatory adversarial scenario dispositions;
25. structural inversion result;
26. subtractive/YAGNI result;
27. exact stop/go recommendation for continuing R10 design;
28. whether any promoted authority must reopen;
29. whether a bounded follow-up is sufficient or another broad review is required.

Findings must cite concrete file/table/function/test evidence.

---

# 25. Scope / write authorization

This is a review-only task.

AUTORIZATION is limited to:

- reading repository/docs/code/schema/OpenAPI/tests/runtime evidence;
- running read-only verification/tests/checks;
- creating exactly one review artifact:
  `docs/superpowers/analysis/2026-08-17-whole-system-implementation-conformance-independent-fable-review.md`;
- committing that review artifact to the current branch;
- pushing that commit to the same branch.

DO NOT:

- modify target authority (`wiki/architecture/*`, handoff, ledger);
- modify code/schema/migrations/OpenAPI/frontend/tests merely to make findings disappear;
- implement B2/B3/B4/etc;
- alter Keycloak/deployment config;
- update PR body;
- merge;
- force-push;
- open the implementation gate;
- treat evidence-only B2-2 candidates/reviews as promoted law.

If a check fails, record the failure truthfully. Do not fix it in this review task.

---

# 26. Completion report

When finished, report succinctly:

```text
overall verdict
BLOCKER / MAJOR / LOW
unauthorized drift = YES/NO
active current defect blocker = YES/NO
safe to continue R10 design = YES/NO
current implementation fully target-conformant = YES/NO
most important 5 findings
most important 5 current safety properties to preserve
whether authority must reopen
whether broad follow-up is required
remote commit SHA
```

The review artifact is evidence only. Operator adjudication remains the next authority gate.