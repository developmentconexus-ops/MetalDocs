# T11 — Discussion / Mention / Notifications post-promotion coherence

> **Status:** PASS / CURRENT T11 EVIDENCE.  
> **Scope:** current-authority coherence after promotion of the operator-ratified Discussion / `@mention` / Notifications bounded reopen.  
> **Method:** DevelopmentConexus Engineering Method v1.0.0.  
> **Implementation:** BLOCKED.

## 1. Promotion under review

Promoted current result:

```text
semantic owners                  4 business + 2 supporting
stable SPA routes                11
PermissionCode values            16
application operations           86
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

Core current authority:

```text
docs/decisions/discussion-notifications-launch.md
```

It is a bounded current cross-layer authority: it supersedes only conflicting pre-consumer clauses in the larger Product/T1→T9 authorities and leaves every unrelated ratified statement untouched.

## 2. Current-authority alignment check

### Routing

```text
docs/index.md
  routes Discussion/@Mention/Notifications directly to the bounded current authority
  affected Product/T3/T5/T6/T8/T9 tasks pair their base authority with the amendment
```

Result: **PASS** — a fresh actor is not sent to a stale base clause alone for this capability.

### Decision register

```text
docs/decisions/index.md
  PRODUCT        bounded T11 reopen current
  OWN            4 business + Audit + Notifications
  T1             Discussion/Mention/Notification families current
  T3             document.discuss + canonical Discussion disclosure current
  T5             persistent local Notification + ephemeral SSE wake-up current
  T6             /notifications + Discussion current
  T6-API         86 operations / 11 Idempotency-Key creations
  T8-B→G         bounded current amendment routed
  T8-H/T9        bounded current refinement routed
  T11-COLLAB     CURRENT / OPERATOR-RATIFIED / FABLE-CONVERGED
```

Result: **PASS**.

### Forward obligations

```text
docs/decisions/forward-obligations.md
  ASY-02 no longer DEFERRED
  consumed by current Launch requirement
  email/push/preferences remain deferred inside the bounded authority

PRESERVE 21
REOPEN    3
DEFERRED 26
TOTAL    50
```

Result: **PASS** — no live `Notifications are deferred` contradiction remains in the forward register.

### Numeric census

```text
docs/decisions/api-operation-census.md
  current operations                86
  current Idempotency-Key creates   11
  ETag domains                      13 / 13
  exact-byte resources              4
  operation 87+ requires new lawful basis
```

The census authority explicitly classifies earlier `78 / operation79-absent` statements as historical stage snapshots or bounded clauses superseded by the T11 amendment.

Result: **PASS** — one current numeric authority.

### Semantic ownership

```text
docs/architecture/ownership.md
  BUSINESS     Authentication / Organization / Authorization / Controlled Documents
  SUPPORTING   Audit / Notifications

Controlled Documents owns DiscussionMessage/Mention
Notifications owns persistent recipient attention/engagement
realtime remains mechanism
```

Result: **PASS**.

### Semantic state

```text
docs/architecture/domain-model.md
  Controlled Documents + DocumentDiscussionMessage + Mention
  Notifications + Notification engagement
```

Result: **PASS**.

### Program status

```text
docs/roadmap.md
  current system invariants = 4+2 / 11 / 16 / 86 / 11 / 13-13 / 4
  implementation remains BLOCKED
  B01 Notification structural delta not yet re-LOCKED
  B03 remains OPEN / NOT LOCKED
```

Result: **PASS**.

## 3. Supersession versus stale contradiction

Large previously ratified Product/T3/T5/T6/T8-B→G/T9 documents retain historical/pre-consumer sentences such as `78 operations`, `operation 79 absent` or `Notifications not a Launch consumer`.

Those sentences are not treated as competing current authorities because:

```text
1. discussion-notifications-launch.md explicitly names the exact bounded clauses it supersedes;
2. docs/decisions/index.md routes the base authority + bounded amendment as current meaning;
3. docs/index.md routes affected tasks to the bounded amendment;
4. api-operation-census.md is the sole current numeric census and explicitly closes prior 78 snapshots;
5. forward-obligations.md consumed ASY-02 rather than leaving a contradictory DEFERRED disposition;
6. unrelated content in those large authorities remains current and is not rewritten merely for cosmetic normalization.
```

This is a bounded-authority amendment, not an attempt to make old closure prose disappear from Git/current documents. A future substantive rewrite may absorb the amendment, but it cannot silently change its meaning.

Result: **PASS** — no live ambiguous authority path remains for the bounded subject.

## 4. Review chain

```text
Lead GCR R1      NOT CONVERGED / MATERIAL=3 / IMPORTANT=6
operator         accepted all bounded corrections
Lead GCR R2      CONVERGED / MATERIAL=0 / IMPORTANT=0
Fable PR #165    CONVERGED / MATERIAL=0 / IMPORTANT=1 / OPTIONAL=3
operator         accepted F-1 + O1→O3
Fable Round 2    NOT JUSTIFIED
PR #165          CLOSED / UNMERGED
```

Fable's sole IMPORTANT was the exact register/census omission corrected before promotion.

## 5. Repository verification

Promotion checkpoint before this evidence record:

```text
HEAD  588fe56dc179b74e8777bc879f147ade2ba0cdcc
CI    #1286 SUCCESS
```

The required check covers, among other Repository Standard properties:

```text
bootstrap budget
forward-obligation count proof
durable documentation routing/link integrity
architecture-first tracked-path allowlist
Draft-only docs/work HTML rule
```

No Product/runtime implementation was introduced.

## 6. Global coherence verdict

```text
VERDICT: PASS
MATERIAL contradictions: 0
IMPORTANT contradictions: 0
```

Current durable meaning is coherent:

```text
4+2 owners
11 stable SPA routes
16 PermissionCode values
86 application operations
11 Idempotency-Key creations
13/13 ETag domains
4 exact-byte resources
```

No further architecture/review round is justified before the frontend structural gate.

## 7. Next gate

```text
smallest B01 P8 Notification delta
→ render utility-header bell + unseen badge
→ render desktop Quick Inbox
→ render narrow/mobile transformation
→ preserve sidebar unchanged
→ operator visual adjudication
→ only operator may re-LOCK implicated B01 delta
→ resume B03 P8 afterward
```

B04+, T12 and Product implementation remain blocked.