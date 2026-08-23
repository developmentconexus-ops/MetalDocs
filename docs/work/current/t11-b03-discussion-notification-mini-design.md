# T11 — B03 Discussion / Mention / Notification Mini-Design

> **Status:** CANDIDATE / PARTIALLY OPERATOR-RATIFIED.  
> **Parent:** B03 — Document Official / Ficha do Documento.  
> **Purpose:** close the smallest Product/UX semantics required by the operator-mandated Launch V1 Document Discussion + `@mention` + in-app Notification capability before upstream Product/T3/T5/T6/T8-E/T8-F consolidation.  
> **Implementation:** BLOCKED.  
> **Census:** existing 78 operations / 10 Idempotency-Key creations remain current accepted authority until the bounded reopen is fully designed and ratified.  
> **Reasoning authority:** `developmentconexus-ops/conexus-methodology/METHOD.md` — DevelopmentConexus Engineering Method v1.0.0.

## 1. Analysis law

Every decision in this mini-design follows the canonical METHOD decision core proportionally:

```text
Evidence
→ Known / Inferred / Unknown / Deferred
→ Root Cause
→ Target Invariant
→ Constraints
→ Credible Alternatives
→ Local Maximum vs Global Maximum
→ Essential vs Accidental Complexity
→ YAGNI / Overengineering / Future Cost
→ Authority / Boundary when relevant
→ Enforcement
→ Proof Strategy
→ Adversarial Challenge
→ Decision
→ Reopen Triggers
```

Repository-local application:

```text
current MetalDocs authority first
→ real user need / named consumer
→ alternatives independent of current implementation shape
→ legacy/reference evidence only as falsification/input, never authority
→ structural inversion/adversarial test
→ Global Maximum comparison
→ smallest sustainable semantic change that closes the need
→ bidirectional impact across Product / T1–T10 / frontend / API
→ operator adjudication
→ upstream consolidation only after the mini-design closes
```

METHOD laws particularly binding here:

```text
mechanism != authority
YAGNI must remove unsupported capability, not required invariants or justified seams
prepare the seam, not the entire future capability
new real consumer / changed requirement is a valid bounded-reopen trigger
Global Maximum is not maximum abstraction or infrastructure
```

Technology/framework/library choice is intentionally deferred until Product semantics are frozen. A framework may implement the accepted model; it may not define the model.

## 2. Fixed Product requirement

Operator-required for Launch V1:

```text
stable-Document human discussion
+ explicit @mention of MetalDocs Users
+ in-application Notification for mention
```

Semantic separation:

```text
Document Discussion != DRAFT/editor comments
Document Discussion != SubmissionFeedback
Document Discussion != GovernanceCase feedback
Notification != access grant
Notification != lifecycle/history authority
```

Legacy evidence supports the user need: the published-document backlog identified a separate display-side `CommentsCard` / `discussion` model because editor comments were not appropriate for the published Document page. Legacy also contained a Notifications feature. Neither legacy model is current authority.

## 3. D0 — delivery channel baseline — OPERATOR-RATIFIED

**Decision:** Launch V1 `@mention` requires **in-app Notification only**.

```text
in-app notification   REQUIRED
email                 NOT Launch V1 baseline
push                  NOT Launch V1 baseline
```

This intentionally avoids promoting email-delivery/template/retry/bounce/preference infrastructure without a named Launch requirement. Later channels may attach to the same accepted notification intent without redefining Document Discussion.

## 4. D1 — Discussion read/write authorization — OPERATOR-RATIFIED

### Read

Reading Discussion follows the actor's current ability to receive/open the Document Official lens. Discussion is not an access-grant mechanism.

```text
current Document disclosure/access required
notification cannot preserve or expand access
loss of access -> discussion becomes non-disclosable even if an old notification exists
```

### Write / reply / mention

A distinct Product Permission is required:

```text
document.discuss
```

Writing requires:

```text
enabled User
+ current document.discuss grant in matching scope
+ current Document access/disclosure
+ any accepted Discussion-specific state predicate
```

Candidate role-bundle delta approved for the bounded reopen:

```text
viewer              + document.discuss
author              + document.discuss
approver            + document.discuss
area_manager         + document.discuss

governance_viewer   no document.discuss by default
governance_admin    no document.discuss by default
```

Rationale:

- `document.read_effective` remains read semantics and does not silently become write authority.
- `governance_viewer` preserves its deliberate read-only governance/auditor posture.
- `governance_admin` still receives no content participation solely from configuration authority.
- a future read-only content role can exist without redefining Discussion semantics.

Commands must always recheck current authorization; frontend affordance presence is UX guidance only.

## 5. D2 — message / reply semantics — OPERATOR-RATIFIED

**Decision:** Discussion is one chronological linear timeline over the stable Document. A message may optionally reference one prior message, but replies do not create a separate semantic Thread aggregate or an arbitrarily nested tree.

Candidate semantic shape:

```text
DocumentDiscussionMessage
  message_id
  document_id
  author_user_id
  created_at
  body
  reply_to_message_id?       // optional reference to one prior message
  official_revision_at_post? // contextual snapshot, not ownership
```

Binding laws:

```text
message belongs to stable Document identity
reply remains an ordinary message in the same chronological timeline
reply_to_message_id, when present, must reference a message in the same Document Discussion
no separate Thread owner/lifecycle is introduced
no semantic nesting depth exists; a reply to a reply is still one new message with one reference
```

The server also records the official Revision that existed when the message was accepted, when one exists:

```text
current official Revision exists
→ official_revision_at_post = exact current official Revision identity

no official Release exists
→ official_revision_at_post absent
```

This contextual snapshot does not move Discussion ownership from Document to Revision and does not bind the message to WorkingContent/DRAFT. A later Release never rewrites the stored context of an older message.

Rationale / Global Maximum result:

- preserves stable-Document conversation across Revision changes;
- preserves historical conversational context without turning Discussion into Revision history;
- supports direct message references, mentions and future notification routing;
- avoids hidden thread state, deep reply trees and a Slack/forum domain that Launch does not need;
- remains useful even if the visual composer or chat library changes later.

Message edit/delete semantics remain deliberately open for D5. `@mention` parsing/identity belongs D3. Pagination/API mechanics belong D7.

## 6. D3 — @mention discovery + accepted-message validation — OPERATOR-RATIFIED

**Decision:** `@mention` is a server-derived, Document-scoped, disclosure-safe interaction. A candidate need not have `document.discuss`; the relevant admission criterion is that the User can currently receive/read the Discussion on that exact Document.

Candidate eligibility:

```text
existing MetalDocs User
+ same Company
+ ENABLED
+ currently eligible to receive/read this exact Document Discussion
+ candidate != message author
```

`document.discuss` is deliberately **not** required of the mentioned User. This preserves D1's separation between reading and writing: a read-only actor may be mentioned to receive context without thereby gaining the ability to reply.

### Purpose-built discovery

The frontend must not use an administrative User directory or reconstruct disclosure rules. Mention discovery is purpose-built for the exact Document and returns only bounded human-reference data needed for recognition/selection.

Conceptually:

```text
@bea
→ server-side search within currently mention-eligible Users for this Document
→ bounded UserReference-like results
```

No email, Role/Permission set, administrative profile payload or explanation of excluded Users is required merely to populate the composer.

### Stable mention identity

A mention is not authoritative merely because message text contains characters such as `@Beatriz`. The accepted message must carry a semantic mention token/reference bound to stable `user_id`; display text is presentation.

Candidate content model remains deliberately minimal rather than reusing ProseMirror or an arbitrary rich-text document model:

```text
MessageContent
  = Text
  | Mention(user_id)
```

Equivalent wire encoding may be chosen later in D7, but it must preserve one authoritative User identity per Mention and cannot rely on reparsing display text after acceptance.

### Accepted-message revalidation

Autocomplete results are UX guidance only. Message acceptance rechecks current truth atomically:

```text
author still ENABLED
+ author still has document.discuss in scope
+ author can still receive the exact Document Discussion
+ every Mention target still:
    exists
    is same-Company
    is ENABLED
    can currently receive/read this exact Document Discussion
```

If any explicit Mention is no longer admissible:

```text
reject whole message command
zero DiscussionMessage
zero Notification
preserve local composer input for explicit reconciliation
```

The server must not silently publish the message while dropping an invalid Mention, because that would falsely communicate to the author that the target was notified.

### Notification trigger law for Launch V1

```text
explicit accepted Mention -> one in-app Notification intent for that target/message
same User mentioned multiple times in one accepted message -> at most one Notification for that message
author self-mention -> not admitted / not offered as candidate
reply without explicit Mention -> no Notification solely because it is a reply
reply with explicit Mention -> normal Mention notification law
```

Mention does not grant access, create a governance participant, change lifecycle state, or make a User a persistent Discussion member.

Rationale / Global Maximum result:

- preserves D1 read/write separation;
- avoids information leakage from generic Company/User directories;
- avoids frontend Authorization/disclosure duplication;
- uses stable Product User identity rather than mutable names/text parsing;
- prevents false-positive notification UX under races/offboarding/access changes;
- keeps the initial Notification consumer closed to explicit mentions instead of inventing broad event-subscription semantics;
- remains valid if the visual composer or chat library changes later.

## 7. Reference evidence for D4+ — NOT AUTHORITY

Targeted external study is admitted only to test whether the candidate Notification model is a local custom invention or a stable industry pattern.

Observed mature Inbox patterns:

```text
Knock      unseen / seen / read / archived; badge/filter/feed semantics
MagicBell  unseen / unread / read / archived; mark-all + per-item state actions
Novu       seen vs read distinction; read/unread + mark-all
```

This evidence supports evaluating a richer standard Inbox engagement model rather than assuming `UNREAD/READ` alone is globally sufficient. It does **not** authorize importing any vendor workflow, channel, subscriber, preference or delivery platform into MetalDocs.

Mechanism study note:

```text
Novu self-host currently brings its own MongoDB + Redis + worker/WebSocket/service topology.
MetalDocs current technical authority already selects PostgreSQL product state + one PostgreSQL-backed durable-job mechanism.
```

Therefore a full notification platform is not the default solution merely because its Inbox state model is useful reference evidence. D7/D8 must compare reuse only against the actual frozen MetalDocs requirements and the METHOD mechanism/authority law.

## 8. Open decisions

Close one at a time before upstream consolidation:

```text
D4 Notification engagement lifecycle / identity / navigation
D5 disclosure, offboarding, deletion/edit immutability behavior
D6 smallest B01 shell impact
D7 exact owner/state/API/async boundary and exact census delta
D8 technology-spike requirements + framework/repo/library evaluation
```

No B04+ baseline opens while a material B03 dependency remains unresolved.
