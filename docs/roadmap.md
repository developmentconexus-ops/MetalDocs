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
T1 → T10              CLOSED / OPERATOR-RATIFIED / INTEGRATED; bounded T11 collaboration amendment CURRENT
T11                   OPEN / ACTIVE
T12                   NOT OPEN
IMPLEMENTATION         BLOCKED
Draft PR               #162
branch                 arch/t11-implementation-program
opening main           cae6ba48df5d611959c0390e0f2b9b8194d62a9d
```

Repository-local Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`. The bounded current cross-layer Discussion/Notifications amendment is `docs/decisions/discussion-notifications-launch.md`. `docs/work/current/**` remains temporary Draft planning/evidence and cannot survive a merge candidate.

## Frontend Product Experience gate

Method:

```text
docs/development/functional-html-wireframe-method.md
Frontend Product Experience Planning Method v2.1
```

Current block state:

```text
B01  App Shell + Global IA       LOCKED / OPERATOR-RATIFIED baseline
      notification delta         CURRENT ARCHITECTURE / P8 NOT YET RE-LOCKED

B02  Library / discovery         LOCKED / OPERATOR-RATIFIED

B03  Document Official           OPEN / CANDIDATE / NOT LOCKED
      Discussion/Notifications   CURRENT / OPERATOR-RATIFIED / GCR+FABLE CONVERGED

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

The current Notifications amendment keeps the sidebar unchanged and adds only utility-header bell/unseen badge, Quick Inbox and stable `/notifications`. The implicated B01 structural scope still requires the method-mandated rendered P8 delta and explicit operator re-LOCK.

B03 remains record/ficha-first with a deliberate distinct read-only official-content viewer and stable-Document Discussion. B03-F1 remains open for server-derived Document Official `allowed_actions` precision. B04+, T12 and Product implementation remain blocked.

## Current accepted system invariants

The Discussion / `@mention` / Notifications bounded reopen is now current repository authority through the routed decision register, ownership/domain authorities, forward-obligation disposition and API census.

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
Product implementation           BLOCKED
```

`docs/decisions/api-operation-census.md` is the sole current numeric census authority. Earlier `78` / `operation 79 absent` statements are historical stage snapshots or bounded clauses superseded by the T11 current amendment; they do not override the current census.

Existing operator-approved T11 precisions still awaiting final T11 durable absorption/cleanup:

```text
T8-E-RO  responsible_owner_candidates on getDocument
B02-LD   first-page Library discovery options
```

## Current Discussion / Mention / Notifications authority

Current Product/UX/technical result:

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

Current enforcement precision:

```text
Authorization alone owns final Mention-target ALLOW/DENY
protected author/target eligibility serializes with offboarding in deterministic user_id order
DocumentDiscussionDisclosure is the one named disclosure composition reused across read/Mention/presentability
Notification presentability is composed server-side before public pagination/counts
batch seen cannot become a disclosure/cardinality oracle
completed message replay does not rerun historical Mention-target eligibility
Discussion/Notifications do not duplicate Audit/History truth
anchor_message_id remains one Discussion-list pagination authority
message bounds = 1..64 segments / <=20 unique Mention targets / <=4096 aggregate Text code points
batch-seen request <=100 ids
SSE call graph remains transport -> application -> mechanism
all Notification-changing wake-ups happen only after commit
transient deploy-overlap wake loss is tolerated because SSE is non-authoritative
OpenAPI server-side text/event-stream proof remains an implementation-readiness closure gate
```

## Review / promotion proof

```text
Lead GCR R1      NOT CONVERGED   MATERIAL=3 / IMPORTANT=6
operator         accepted all bounded corrections
Lead GCR R2      CONVERGED       MATERIAL=0 / IMPORTANT=0
Fable PR #165    CONVERGED       MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
operator         F-1 + O1→O3 ACCEPTED
Fable Round 2    NOT JUSTIFIED
Evidence PR #165 CLOSED / UNMERGED
promotion HEAD   29599eab33a5479adf3ee92741f0a9b5ee8720a9
promotion CI     #1285 SUCCESS
```

Fable explicitly confirmed survival of 4+2 owners, 11 routes, 16 permissions, 86 operations, 11 Idempotency-Key creations, same-Scope Mention→Notification, presentability-before-paging/counts, Lexical, SSE/in-process wake-up, River-only durable async and absence of generic EventBus/broker/Redis.

## Exact next action

```text
1. Run post-promotion whole-current-authority coherence proof and correct only genuine live contradictions.
2. Render the smallest B01 P8 notification delta:
     utility-header bell + unseen badge
     desktop Quick Inbox
     narrow/mobile transformation
     transition to /notifications
3. Operator visually adjudicates and may re-LOCK only the implicated B01 delta.
4. Resume B03 P8 using the now-current Discussion/Notification semantics.
5. Resolve B03-F1 before any B03 LOCK.
6. Only operator may LOCK B03.
7. Do not open B04+, T12, or Product implementation before their normal gates.
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

Accepted Product/R10 decisions reopen only on material evidence under the DevelopmentConexus Engineering Method. Preference, sunk cost, framework availability, hypothetical scale or infrastructure fashion are not reopen triggers.
