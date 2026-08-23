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

Current Product/architecture authority lives in `docs/product/**`, `docs/architecture/**`, and `docs/decisions/**`. The bounded current Discussion/Notifications amendment is `docs/decisions/discussion-notifications-launch.md`. `docs/work/current/**` remains temporary Draft planning/evidence and cannot survive a merge candidate.

## Frontend Product Experience gate

Method:

```text
docs/development/functional-html-wireframe-method.md
Frontend Product Experience Planning Method v2.1
```

Current block state:

```text
B01  App Shell + Global IA       LOCKED / OPERATOR-RATIFIED baseline
      notification delta         P8 CURRENT / NOT YET RE-LOCKED

B02  Library / discovery         LOCKED / OPERATOR-RATIFIED

B03  Document Official           OPEN / CANDIDATE / NOT LOCKED
      Discussion/Notifications   CURRENT / OPERATOR-RATIFIED / GCR+FABLE+PROMOTION COHERENT

B04+                             NOT OPEN
```

B01 preserved mental model:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

The current Notifications amendment keeps the sidebar unchanged and adds only:

```text
utility-header bell + unseen badge
desktop Quick Inbox
narrow/mobile accessible transformation
stable /notifications full Inbox route
```

Only the implicated B01 structural scope is reopened. A rendered P8 delta + explicit operator adjudication are required before it may be re-LOCKED.

B03 remains record/ficha-first with a deliberate distinct read-only official-content viewer and stable-Document Discussion. B03-F1 remains open for server-derived Document Official `allowed_actions`. B04+, T12 and Product implementation remain blocked.

## Current accepted system invariants

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

`docs/decisions/api-operation-census.md` is the sole current numeric census. Earlier `78` / `operation 79 absent` statements are historical stage snapshots or bounded clauses superseded by the current T11 amendment.

Existing operator-approved T11 precisions still awaiting final T11 durable absorption/cleanup:

```text
T8-E-RO  responsible_owner_candidates on getDocument
B02-LD   first-page Library discovery options
```

## Current Discussion / Mention / Notifications authority

Binding result:

```text
stable-Document Discussion
immutable DiscussionMessage
semantic Mention(user_id)
purpose-built disclosure-safe mention autocomplete
document.discuss write Permission
Notifications = second supporting owner
seen / read-unread / archive-unarchive Inbox
bell + Quick Inbox + /notifications
same-Scope accepted Mention -> Notification
server-side presentability before paging/counts
Lexical replaceable composer mechanism
SSE invalidation only
River remains sole durable future-work mechanism
no generic EventBus / external broker / Redis baseline
```

Important precision:

```text
DocumentDiscussionDisclosure = one named disclosure composition reused across read/Mention/presentability
protected author/target eligibility locks in deterministic user_id order
anchor_message_id stays within one Discussion-list pagination authority
message: 1..64 segments / <=20 unique Mention targets / <=4096 aggregate Text code points
batch seen: <=100 ids, bodyless/no cardinality result
all Notification wake-ups occur post-commit
transient wake loss is tolerated because SSE is non-authoritative
OpenAPI server-side text/event-stream remains an implementation-readiness proof gate
```

## Review / promotion proof

```text
Lead GCR R1      NOT CONVERGED   MATERIAL=3 / IMPORTANT=6
Lead GCR R2      CONVERGED       MATERIAL=0 / IMPORTANT=0
Fable PR #165    CONVERGED       MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
operator         F-1 + O1→O3 ACCEPTED
Fable Round 2    NOT JUSTIFIED
PR #165          CLOSED / UNMERGED
promotion CI     #1286 SUCCESS
post-promotion coherence PASS / MATERIAL=0 / IMPORTANT=0
```

Post-promotion evidence:

```text
docs/work/current/t11-discussion-notifications-promotion-coherence.md
```

## Exact next action

```text
1. B01 Notifications P8 — render the smallest structural delta only:
     utility-header bell + unseen badge
     desktop Quick Inbox
     narrow/mobile transformation
     transition to /notifications
     sidebar unchanged
2. Operator visually adjudicates hierarchy/size/position/discoverability/accessibility/responsive behavior.
3. Iterate only the implicated B01 delta until operator re-LOCK.
4. Resume B03 P8 with current Discussion/Notification semantics.
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
