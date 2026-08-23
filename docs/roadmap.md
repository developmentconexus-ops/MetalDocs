---
id: repository-roadmap
kind: authority
owner: architecture
summary: Sole mutable MetalDocs stage, gate, implementation-status, and next-action authority.
---

# MetalDocs roadmap

## Current state

```text
REPOSITORY MODE       CLEAN-SLATE / ARCHITECTURE-FIRST
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED
T11                   OPEN / ACTIVE CANDIDATE
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
Draft PR               #162
branch                 arch/t11-implementation-program
opening main           cae6ba48df5d611959c0390e0f2b9b8194d62a9d
```

Repository-local Product/architecture authority remains in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`. `docs/work/current/**` is temporary Draft planning/evidence and cannot survive a merge candidate.

## Frontend Product Experience gate

Method:

```text
docs/development/functional-html-wireframe-method.md
Frontend Product Experience Planning Method v2.1
```

Current block state:

```text
B01  App Shell + Global IA       LOCKED / OPERATOR-RATIFIED
      notification delta         REOPEN CANDIDATE / OPERATOR-RATIFIED / P8 NOT YET RE-LOCKED

B02  Library / discovery         LOCKED / OPERATOR-RATIFIED

B03  Document Official           OPEN / CANDIDATE / NOT LOCKED
      Discussion/Notifications   D0→D8 OPERATOR-RATIFIED
      Lead GCR Round 2           CONVERGED — MATERIAL=0 / IMPORTANT=0
      Fresh Fable                CONVERGED — MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
      Fable adjudication         F-1 + O1→O3 ACCEPTED / ROUND 2 NOT JUSTIFIED
      Upstream consolidation     CURRENT GATE

B04+                             NOT OPEN
```

B01 preserved mental model remains:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

The notification reopen candidate keeps the sidebar unchanged and adds only utility-header bell/unseen badge, Quick Inbox and stable `/notifications`. A smallest-scope rendered P8 delta is required before the implicated B01 scope may be re-LOCKED.

B03 remains record/ficha-first with a deliberate distinct read-only official-content viewer and a stable-Document Discussion surface. B03-F1 remains open for server-derived Document Official `allowed_actions` precision. B04+, T12 and Product implementation remain blocked.

## Current accepted system invariants

Until the Discussion/Notifications reopen is coherently consolidated and promoted across every owning authority, current integrated authority remains:

```text
semantic owners                  4 business + Audit supporting
stable SPA routes                10
application operations           78
operation 79                     ABSENT
Idempotency-Key creations        10
ETag read / mutation domains     13 / 13
exact-byte resources             4
Product implementation           BLOCKED
```

Existing operator-approved T11 precisions remain candidate for T11 consolidation:

```text
T8-E-RO  responsible_owner_candidates on getDocument
B02-LD   first-page Library discovery options
```

## Discussion / Mention / Notifications bounded reopen — converged candidate

Operator-ratified and independently challenged Product/UX/technical result:

```text
stable-Document Discussion
immutable Launch DiscussionMessage
chronological timeline + optional one-message reply reference
semantic @Mention(user_id)
purpose-built disclosure-safe mention autocomplete
in-app Notification only; no Launch email/push
new document.discuss write Permission
Notifications = second supporting semantic owner
seen / read-unread / archive-unarchive Inbox lifecycle
current source disclosure always rechecked
bell + Quick Inbox + /notifications
same-local-transaction accepted Mention -> Notification
Lexical as replaceable composer mechanism
SSE invalidation only; canonical Notification truth stays persistent
River remains the one durable future-work mechanism
no generic EventBus / external broker / Redis baseline
```

Converged candidate result:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

Converged enforcement precisions:

```text
Authorization alone owns final Mention-target ALLOW/DENY
protected author/target eligibility serializes with offboarding in deterministic user_id order
one named canonical Discussion-disclosure predicate is reused across read/Mention/presentability
Notification presentability is composed server-side before public pagination/counts
batch seen cannot become a disclosure/cardinality oracle
completed message replay does not rerun historical Mention-target eligibility
Discussion/Notifications do not duplicate Audit/History truth
anchor_message_id remains one list operation/filter/cursor authority
message segment + unique-Mention counts receive explicit executable-wire bounds
SSE call graph remains transport -> application -> mechanism
all Notification-changing wake-ups happen only after commit
transient deploy-overlap wake loss is tolerated because SSE is non-authoritative
OpenAPI server-side text/event-stream proof remains a closure gate
```

These values are still **candidate reopen results, not current integrated authority**, until the current consolidation gate completes.

## Independent review proof

```text
Lead GCR R1      NOT CONVERGED   MATERIAL=3 / IMPORTANT=6
operator         accepted all bounded corrections
Lead GCR R2      CONVERGED       MATERIAL=0 / IMPORTANT=0
Fable PR #165    CONVERGED       MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
operator         F-1 + O1→O3 ACCEPTED
Fable Round 2    NOT JUSTIFIED
```

Fable explicitly confirmed survival of 4+2 owners, 11 routes, 16 permissions, 86 operations, 11 Idempotency-Key creations, same-Scope Mention→Notification, presentability-before-paging/counts, Lexical, SSE/in-process wake-up, River-only durable async and absence of generic EventBus/broker/Redis.

## Exact next action

```text
1. Consolidate only the implicated current authorities:
     Product contract
     T1 semantic state
     Ownership
     T3 Authorization/Audit
     T5 async/realtime
     T6 journeys/routes/API meaning
     T8-B→G
     T9 validation/golden-flow obligations
     T11 planning
     docs/decisions/forward-obligations.md
     docs/decisions/api-operation-census.md
2. Re-run whole-package coherence over the consolidated current authorities.
3. Promote 4+2 / 11 routes / 86 operations / 11 Idempotency-Key creations only when every current authority agrees.
4. Render/adjudicate smallest B01 P8 notification delta; only operator may re-LOCK the implicated scope.
5. Resume B03 P8 with real Discussion/Notification semantics; resolve B03-F1; only operator may LOCK B03.
6. Do not open B04+, T12, or Product implementation before their normal gates.
```

## Hard stops

```text
no Product code/schema/OpenAPI implementation/runtime/deploy work
no T12 work
no framework/library allowed to redefine Product semantics
no generic EventBus/broker/Redis without a named material trigger
no frontend Authorization matrix
no legacy implementation restoration by sunk cost
no merge authorization implied
```

## T11 closure / implementation gate

Implementation remains blocked until:

```text
T11 CLOSED / OPERATOR-RATIFIED
T12 CLOSED / OPERATOR-RATIFIED
Integrated Whole-R10 coherence = PASS
fresh independent challenge = converged
operator implementation authorization = explicit
```

## Reopen law

Accepted Product/R10 decisions reopen only on material evidence under the DevelopmentConexus Engineering Method. The operator-required Discussion/@Mention/Notification capability is valid new-consumer evidence. Preference, sunk cost, framework availability, hypothetical scale or infrastructure fashion are not reopen triggers.
