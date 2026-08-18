# R10-T6 — Canonical API / Frontend Journeys — Stage Bootstrap

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **T6 OPEN / DESIGN NEXT**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product Contract:** `wiki/architecture/launch-v1-product-contract.md` — REV001  
> **Decision baseline:** `wiki/architecture/rebaseline-decision-registry.md`  
> **Implementation:** BLOCKED

This file opens T6 after explicit operator closure of the post-T5 independent Fable checkpoint. It is a routing/bootstrap artifact only: it does **not** pre-decide API routes, frontend screens, DTOs, provider choices, SQL, package layout or implementation plan.

## 1. Authority order

Read and obey:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. this T6 staging file
15. current API/frontend/runtime only as evidence needed to falsify or validate a concrete T6 claim

Historical/current implementation shape is not target authority.

## 2. Binding method laws

```text
smallest sustainable solution
one authority per meaning
mechanism != authority
proof before implementation
revalidation does not mean reinvention
prepare the seam, not the dormant capability
```

T6 must encode ratified product semantics into user/API journeys without creating a second lifecycle, Authorization, exact-content, Search or Audit authority.

## 3. T1→T5 baseline that T6 may not casually reopen

```text
Document != Revision != WorkingContent != Submission
REV000 initial issuance / REV001 first revision
human-readable title = Revision-governed metadata
WorkingContent OCC/CAS = DRAFT concurrency authority
Submission immutable exact attempt
one sequential Governance Step model
Release = sole normal effectivity authority
bounded withdrawal of active human-governed obsolescence request
current Authorization = live grants + scope + domain predicates
Audit = action evidence, not current state
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
managed_content_id = retrieval mechanism only
OPEN→READY/admission/malware laws remain T4 authority
viewer/preview != OfficialRendition
Search journey required; baseline = canonical PostgreSQL query/view
materialized Search/search_refresh only if T6 proves a derived/expensive/measured consumer
no mandatory Launch notifications/event bus
no generic integration/event platform
```

## 4. Official T6 REOPEN set

T6 owns only the current Decision Registry REOPEN set:

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
source upload / T4 admission UX
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
exact Search field/ranking UX + prove whether any derived/expensive fact activates materialized Search seam
EditorSession/UX lease only if a real editor-integration consumer requires it
```

## 5. Required T6 proof questions

T6 must eventually prove, without implementation:

1. **Canonical command/query surface:** Which public operations are actually required by the Launch journeys, and which are merely current-code artifacts?
2. **Reader truth:** How do ordinary readers always land on current EFFECTIVE truth while authors/governance actors can inspect authorized DRAFT/SUBMITTED/history without ambiguity?
3. **Title semantics:** How does DRAFT retitle participate in an existing T2 concurrency law without creating an out-of-band metadata race?
4. **Upload/admission:** How does the browser journey express T4 `OPEN→READY`, malware/preflight failures and retry without frontend/provider state becoming semantic truth?
5. **Viewer/source:** What is shown in-product versus downloaded exactly, preserving `viewer/preview != OfficialRendition`?
6. **Governance work:** How do active participants reach only the exact current case context they are authorized to act on?
7. **Admin:** What is the smallest coherent admin surface for Users/Profiles/Areas/Groups/Memberships/RoleAssignments/DocumentTypes/routes/templates without resurrecting removed platform concepts?
8. **Numbering:** What grammar/configuration and preview behavior is necessary, while preview reserves nothing and committed codes never reuse?
9. **Search:** Are canonical fields sufficient? If T6 proposes any derived/expensive field, what named user journey proves materialized Search is necessary?
10. **History/Audit:** How are domain history and Audit presented distinctly so Audit never becomes lifecycle reconstruction authority?
11. **Errors/idempotency:** Which commands need public idempotency and what stable conflict/error semantics expose T2/T3/T4 failures truthfully?
12. **EditorSession:** Does the chosen editor integration prove a bounded UX lease/session is needed, or can T2 WorkingContent OCC remain the only concurrency mechanism?

## 6. Explicit non-decisions at T6 opening

Not yet decided:

```text
endpoint paths / HTTP verbs
request/response DTO shapes
OpenAPI schema
frontend route tree / component hierarchy
exact editor/viewer product selection
exact renderer product selection
final Search fields/ranking/materialization
SQL/indexes
Go package placement
queue/process topology
Historical Migration API/execution
implementation plan or code
```

## 7. Stage protocol

```text
read Decision Registry
→ consume CURRENT / PRESERVE / REFINED
→ design only T6 REOPEN set
→ credible alternatives / Global Maximum analysis
→ candidate decisions
→ operator material-decision adjudication
→ platform-facing T6 summary
→ explicit operator summary ratification
→ promote durable T6 authority
→ update Decision Registry
→ remove completed T6 staging
→ only then open T7
```

## 8. Current gate

```text
Product Contract REV001        OPERATOR-APPROVED
T1→T5                          CLOSED / OPERATOR-RATIFIED
Post-T5 Fable checkpoint       CLOSED / OPERATOR-APPROVED
Decision Registry              CURRENT / OPERATOR-RATIFIED
T6                             OPEN / DESIGN NEXT
T7                             NOT OPEN
implementation                 BLOCKED
```

The next work is T6 architecture/design only. No product implementation is authorized.
