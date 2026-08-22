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
      Discussion/Notifications   D0→D8 OPERATOR-RATIFIED CANDIDATE
      Lead GCR                   NOT CONVERGED — MATERIAL=3 / IMPORTANT=6

B04+                             NOT OPEN
```

B01 preserved baseline:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

The notification reopen candidate keeps the sidebar unchanged and adds only:

```text
utility-header bell + unseen badge
Quick Inbox
stable /notifications full Inbox route
```

A smallest-scope rendered P8 delta is required before the implicated B01 scope may be re-LOCKED.

B03 direction remains:

```text
/documents/:document_id
= stable Document ficha/record first
→ explicit read-only official-content viewer
→ Document Discussion on the stable Document

Document Work   remains separate B04
Governance Case remains separate B06
History         remains separate B07
```

B03-F1 remains open: server-derived Document Official `allowed_actions` precision for create Revision / owner replacement / obsolescence actions without frontend AuthZ duplication.

## Current accepted system invariants

Until the bounded Discussion/Notifications reopen is fully corrected, independently challenged, consolidated and promoted, current integrated authority remains:

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

## Discussion / Mention / Notifications bounded reopen candidate

Operator-ratified Product/UX candidate:

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
```

Candidate technical result after D7/D8:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4

generic EventBus                 absent
external broker                  absent
Redis baseline                   absent
River durable-work mechanism     preserved
Notification creation            same local PostgreSQL Scope as accepted Mention
Notification realtime            SSE invalidation only
Launch wake-up                   in-process coalescing mechanism candidate
multi-replica wake-up seam       PostgreSQL LISTEN/NOTIFY first candidate
Discussion composer              Lexical mechanism; Product persists Text | Mention(user_id)
```

These values are **candidate reopen results, not yet promoted current authority**.

Durable candidate records:

```text
docs/work/current/t11-b03-discussion-notification-mini-design.md
docs/work/current/t11-b03-notification-ownership-reopen.md
docs/work/current/t11-b03-notification-engagement.md
docs/work/current/t11-b03-discussion-notification-d5.md
docs/work/current/t11-b01-notifications-reopen.md
docs/work/current/t11-b03-discussion-notification-d7-contract.md
docs/work/current/t11-b03-notification-technology-spike.md
docs/work/current/t11-notifications-global-coherence-review.md
```

## Lead Global Coherence Review

Current Lead GCR verdict:

```text
NOT CONVERGED
MATERIAL   3
IMPORTANT  6
```

The three material findings do **not** change the Product capability set or the candidate `4+2 / 11 routes / 86 ops / 11 idempotent creates` result. They require authority/enforcement corrections:

```text
M1  Mention eligibility:
    Organization authors User facts;
    Controlled Documents authors resource predicate facts;
    Authorization alone decides ALLOW/DENY;
    application coordinates.
    Author + Mention-target ENABLED eligibility must serialize with offboarding.

M2  Notification presentability:
    current source disclosure must be composed server-side before public page/count results;
    no frontend post-filter and no copied ACL/presentable authority in Notifications.

M3  SSE/wake-up:
    preserve transport -> application inbound direction;
    platform realtime remains mechanism only;
    wake after successful Notification creation/engagement commit, never from semantic owners.
```

Important corrections cover race-safe batch seen, completed idempotency replay, Audit/History non-duplication, persistence constraints, OpenAPI SSE proof, and visual/upstream sequencing. Exact text is owned by `docs/work/current/t11-notifications-global-coherence-review.md`.

## Exact next action

```text
1. Operator adjudicates Lead GCR M1→M3 + I1→I6.
2. Apply accepted bounded corrections to the D0→D8 candidate package.
3. Re-run Lead Global Coherence Review; require MATERIAL=0 and IMPORTANT=0 for convergence.
4. Run METHOD-required fresh independent challenge on the corrected package.
5. Adjudicate reviewer evidence; repeat only if material/important findings remain.
6. Consolidate the smallest approved delta into Product / T1 / Ownership / T3 / T5 / T6 / T8-B→G / T9 / T11 authorities.
7. Only then promote exact 4+2 / 11 routes / 86 operations / 11 Idempotency-Key creations as current authority.
8. Render and adjudicate smallest B01 P8 notification delta; only operator may re-LOCK it.
9. Resume B03 P8 with real Discussion/Notification semantics; resolve B03-F1; only operator may LOCK B03.
10. Do not open B04+, T12, or Product implementation before their normal gates.
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

T11 remains open until the frontend planning method completes through assembled closure and all bounded upstream findings are durably reconciled.

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
