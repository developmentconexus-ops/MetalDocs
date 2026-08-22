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
      Lead GCR Round 2           CONVERGED — MATERIAL=0 / IMPORTANT=0
      Fresh Fable challenge      CURRENT GATE

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

Until the Discussion/Notifications reopen is independently challenged, consolidated and promoted, current integrated authority remains:

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

Operator-ratified Product/UX/technical candidate:

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

Corrected candidate result:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

Lead GCR corrections now binding inside the candidate:

```text
Authorization alone owns final Mention-target ALLOW/DENY
protected author/target eligibility serializes with offboarding in deterministic user_id order
Notification presentability is composed server-side before public pagination/counts
batch seen cannot become a disclosure oracle
completed message replay does not rerun historical Mention-target eligibility
Discussion/Notifications do not duplicate Audit/History truth
SSE call graph remains transport -> application -> mechanism
all Notification-changing wake-ups happen only after commit
OpenAPI server-side text/event-stream proof remains a closure gate
```

These values are still **candidate reopen results, not current integrated authority**.

## Exact next action

```text
1. Verify corrected candidate HEAD with required CI.
2. Run fresh independent Fable challenge over the whole coherent D0→D8 + corrected GCR package.
3. Adjudicate every Fable MATERIAL/IMPORTANT finding; Round 2 only if a real contradiction survives.
4. Consolidate the smallest approved delta into Product / T1 / Ownership / T3 / T5 / T6 / T8-B→G / T9 / T11 authorities.
5. Only then promote exact 4+2 / 11 routes / 86 operations / 11 Idempotency-Key creations as current authority.
6. Render/adjudicate smallest B01 P8 notification delta; only operator may re-LOCK the implicated scope.
7. Resume B03 P8 with real Discussion/Notification semantics; resolve B03-F1; only operator may LOCK B03.
8. Do not open B04+, T12, or Product implementation before their normal gates.
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
