# T11 — B03 Discussion / Mention / Notification Mini-Design

> **Status:** CANDIDATE / PARTIALLY OPERATOR-RATIFIED.  
> **Parent:** B03 — Document Official / Ficha do Documento.  
> **Purpose:** close the smallest Product/UX semantics required by the operator-mandated Launch V1 Document Discussion + `@mention` + in-app Notification capability before upstream Product/T3/T5/T6/T8-E/T8-F consolidation.  
> **Implementation:** BLOCKED.  
> **Census:** existing 78 operations / 10 Idempotency-Key creations remain current accepted authority until the bounded reopen is fully designed and ratified.

## 1. Analysis law

Every decision in this mini-design follows:

```text
current repository authority first
→ real user need / named consumer
→ alternatives independent of current implementation shape
→ legacy/reference evidence only as falsification/input, never authority
→ inversion/adversarial test
→ global-maximum comparison
→ smallest coherent semantic change that closes the need
→ bidirectional impact across Product / T1–T10 / frontend / API
→ operator adjudication
→ upstream consolidation only after the mini-design closes
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

Rationale / global-maximum result:

- preserves stable-Document conversation across Revision changes;
- preserves historical conversational context without turning Discussion into Revision history;
- supports direct message references, mentions and future notification routing;
- avoids hidden thread state, deep reply trees and a Slack/forum domain that Launch does not need;
- remains useful even if later visual presentation stops resembling a chat.

Message edit/delete semantics remain deliberately open for D5. `@mention` parsing/identity belongs D3. Pagination/API mechanics belong D7.

## 6. Open decisions

Close one at a time before upstream consolidation:

```text
D3 @mention candidate discovery + accepted-message validation
D4 Notification identity / unread-read lifecycle / navigation
D5 disclosure, offboarding, deletion/edit immutability behavior
D6 smallest B01 shell impact
D7 exact owner/state/API/async boundary and exact census delta
D8 technology-spike requirements (framework/repo/library evaluation criteria only)
```

No B04+ baseline opens while a material B03 dependency remains unresolved.
