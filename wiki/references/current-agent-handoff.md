# Current Agent Handoff

> **Last verified:** 2026-08-19  
> **Status:** ACTIVE — **T1→T8-B CLOSED / OPERATOR-RATIFIED; T8-C ACTIVE / ROUND-1 REVIEW + LEAD ADJUDICATION COMPLETE / CORRECTED CANDIDATE MATERIALIZED / BOUNDED ROUND-2 NEXT; T8-D→T12 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/r10-technical-architecture.md` — sole status/next-action router
5. Product Contract REV001 + Whole-Product GCR + 4+1 ownership
6. T1→T8-B durable authorities
7. Decision Registry + amendments through T8-B
8. `wiki/architecture/r10-post-t6-implementation-readiness-program.md`
9. `wiki/architecture/r10-technical-realization-reconciliation-baseline.md`
10. `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-bootstrap.md`
11. `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md`
12. `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-independent-fable-review.md`
13. `docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-adjudicated-corrected-candidate.md`
14. current interfaces/code only when a concrete T8-C evidence claim needs them

Do not route target design through superseded/historical architecture or current module/interface existence.

## Current checkpoint

```text
Product Contract                         REV001 / OPERATOR-APPROVED
Whole-Product GCR / 4+1 ownership       CLOSED / OPERATOR-APPROVED
T1→T8-B                                  CLOSED / OPERATOR-RATIFIED
Decision Registry                        CURRENT + amendments through T8-B
Post-T6 Stage-Decomposition GCR          RESTRUCTURE NOW / OPERATOR-RATIFIED
TRRB                                     CLOSED / OPERATOR-RATIFIED / PROMOTED

T8-C                                     ACTIVE / CORRECTED CANDIDATE / BOUNDED ROUND-2 NEXT
T8-D                                     NOT OPEN
T8-E                                     NOT OPEN
T8-F                                     NOT OPEN
T8-G                                     NOT OPEN
T8-H                                     NOT OPEN
T9                                       NOT OPEN
T10                                      NOT OPEN
T11                                      NOT OPEN
T12                                      NOT OPEN
implementation                           BLOCKED
```

## Binding execution law

> **No Writer task may contain a material architecture decision that should have been decided before execution.**

## Reference-backed decision law

For material technical decisions:

```text
MetalDocs semantic/product authority
→ DevelopmentConexus Engineering Method
→ current repository evidence
→ primary/current standards and official tool/library documentation
→ relevant reference products/patterns as falsification evidence
→ credible alternatives + Global Maximum
→ proof/adversarial review
→ operator ratification
```

Do not reinvent technical mechanisms already solved by the selected stack unless a concrete MetalDocs invariant requires an additional boundary. Do not import an external best practice when it conflicts with MetalDocs authority.

## T7 — CLOSED

```text
NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH
```

Current MetalDocs DB/content/history is DEV/test/throwaway and creates no business-data compatibility entitlement.

## T8-A — CLOSED

Durable authority:

`wiki/architecture/r10-t8a-technical-authority-legacy-disposition.md`

Current implementation remains evidence only. PRESERVE requires all five T8-A proofs.

## T8-B — CLOSED

Durable authority:

`wiki/architecture/r10-t8b-backend-module-package-topology.md`

Ratified Global Maximum:

```text
ONE GO MODULE FOR BACKEND GO CODE
+ OWNER-FIRST MODULAR MONOLITH
+ ONE IMPORTABLE PUBLIC SURFACE PER SEMANTIC OWNER
+ STATELESS APPLICATION LEAF ORCHESTRATION
+ ONE SEMANTIC INBOUND DOOR THROUGH APPLICATION
+ NON-SEMANTIC PLATFORM MECHANISMS
+ WIRING-ONLY COMPOSITION ROOT
+ CLOSED-WORLD / DEFAULT-DENY FIRST-PARTY DEPENDENCY GRAPH
```

Semantic homes remain exactly:

```text
Authentication
Organization
Authorization
Controlled Documents
Audit
```

No direct owner→owner imports, foreign SQL, mechanism-as-authority or second Authorization evaluator.

## T8-C — ACTIVE

### Original candidate

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-global-maximum-candidate.md`

### Round-1 Fable evidence

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-independent-fable-review.md`

Round-1 verdict:

```text
APPROVE T8-C GLOBAL MAXIMUM CANDIDATE WITH MATERIAL FIXES
BLOCKER 5 / MAJOR 6 / LOW 5
GLOBAL MAXIMUM CLASS CONFIRMED
T8-B REOPEN NO
T1→T7 REOPEN NO
```

### Lead adjudication

```text
B1 ACCEPT
B2 ACCEPT
B3 ACCEPT
B4 ACCEPT
B5 REJECT AS BLOCKER after PostgreSQL primary-evidence verification

M1 ACCEPT
M2 ACCEPT
M3 ACCEPT WITH NARROWING
M4 ACCEPT
M5 ACCEPT as real unmade decision -> SELECT PII-FREE REPLAY SNAPSHOT BY CONSTRUCTION
M6 ACCEPT

L1-L5 ACCEPT
```

Reviewer evidence never became authority. The selected Global Maximum class remains confirmed.

### Current corrected candidate / Round-2 input

`docs/superpowers/analysis/2026-08-19-r10-t8c-internal-communication-contracts-adjudicated-corrected-candidate.md`

Key corrected laws:

```text
database/sql-family transaction substrate deliberately selected
sealed txscope + platform-only native SQL binding for River
semantic owners still receive only txscope.Scope
Audit AppendIn + Authorization-scoped historical ListEvents
Authorization AuthorizedScopes from same canonical evaluator
ProtectedSecuritySubjectIn serializes eligibility against offboarding semantically
owner VersionToken + expected-version mutation law; ETag remains T8-E
ProviderClient crosses verified primitive issuer+subject, not raw claim bag
ManagedContent admission claims + DeleteReclaimable + T5-J GC contracts
MalwareInspector returns digest of exactly inspected bytes
OfficialRendition intent remains named; no EventBus
idempotency concurrent outcome frozen, SQL realization remains T8-D
PII-free ReplaySnapshot by construction; no Launch purge/redaction subsystem
OfficialRendition content read explicitly covered
L1-L5 operation-census precision closed in candidate
```

### B5 Lead rejection

The target contract requires concurrent same-key requests to serialize and return claim/replay/conflict without leaving Scope unusable. It does **not** mandate a unique-violation path. PostgreSQL current primary documentation supports conflict handling through `ON CONFLICT DO NOTHING` under READ COMMITTED; exact SQL remains T8-D.

### Exact next action

```text
bounded Fable Round 2 on the adjudicated corrected candidate
→ attack only the material corrected delta
→ verify txscope/River binding
→ verify Audit read + AuthorizedScopes + eligibility serialization + VersionToken
→ verify ManagedContent claim/GC + malware correlation + rendition read
→ directly challenge B5 Lead rejection using PostgreSQL primary evidence
→ directly challenge new PII-free replay decision
→ verify operation-census delta closure and T8-D/T8-E boundaries
→ final Lead adjudication
→ explicit operator ratification before durable T8-C promotion
```

Do **not** start T8-D from the corrected candidate.

### Do not decide by stealth

```text
schema/tables/constraints/indexes/lock SQL → T8-D
exact OpenAPI/wire/ETag encoding            → T8-E
frontend realization                       → T8-F
runtime/process/deployment                  → T8-G
transition/deletion                         → T10
implementation tasks                        → T11
```

T8-C may reopen T8-B only on a concrete required-contract contradiction, not preference.

Implementation remains **BLOCKED**.