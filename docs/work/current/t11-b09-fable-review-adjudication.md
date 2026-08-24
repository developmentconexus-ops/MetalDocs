# T11 — B09 Fable Adversarial Review Adjudication

> **Status:** OPERATOR-ADJUDICATED / REVIEW FINDINGS RESOLVED IN PLANNING AUTHORITY.  
> **Block:** B09 — Audit.  
> **Method:** Frontend Product Experience Planning Method v2.3 + DevelopmentConexus Engineering Method.  
> **Review Evidence PR:** #166 — `review/t11-b09-audit-fable` — Evidence only / never merge.  
> **Review base:** `b62957501a56457a8f46cb081d24ee93f02aac45`.  
> **Review HEAD:** `418f7fb7c3476f83501412887232b85bd79dd401`.  
> **Review evidence file:** `docs/work/current/ai-dialog.md` on the review branch only.  
> **P8 HTML:** NOT STARTED.  
> **Implementation:** BLOCKED.

## 1. Review verdict recovered

Fable independently reviewed Method v2.3 compliance, the B09 P7 exit, durable Audit authority and the P8 realization plan.

Returned verdict:

```text
HOLD BEFORE P8

BLOCKING   0
IMPORTANT  1   I-1
MINOR      7   M-1..M-7
```

Fable explicitly found:

```text
P7 architecture failure                    NONE
upstream Product/backend reopen             NOT REQUIRED
screen-shaped API invention                 NONE
backend-shaped UX suppression               NONE
Audit/History/current-state conflation      NONE
op78/op87/op88/op89 contradiction           NONE
premature generic framework                 NONE
B01-B08 / B10-B12 coherence break           NONE
```

Its sole Important finding concerned falsifiability of the P8 plan's applied-URL behavior, not B09 Product architecture.

## 2. Operator adjudication

The operator approved the complete adjudication of I-1 and M-1..M-7.

Binding disposition:

```text
I-1   ACCEPT — required P8-plan correction
M-1   ACCEPT — compound semantic chips
M-2   ACCEPT — relative period labels are draft-only
M-3   ACCEPT — Area Query Assist failure path
M-4   ACCEPT — op87 fixture candidates must occur in admitted evidence
M-5   ACCEPT — max-20 non-claiming refinement affordance
M-6   ACCEPT — all-human actor category filter REJECTED for Launch
M-7   ACCEPT — P8-only defensive History API evidence hardening

Upstream reopen     NO
P7 redesign         NO
operation 90+       NO
Product change      NO
P8 HTML             NOT STARTED
```

## 3. I-1 — applied URL round-trip

Finding:

```text
P8 plan serialized applied query into location.search
but did not require the reverse path.
```

Without correction, a builder could produce a visually plausible artifact where URL text changed but:

```text
refresh
browser Back / Forward
pasted/copied canonical query
```

failed to reconstruct the same investigation.

Resolution in `t11-b09-p8-realization-plan.md`:

```text
serialize applied stable IDs/enums/UTC instants
render visible canonical-query review evidence
write History API defensively

initial load
  -> parse location.search
  -> appliedQuery
  -> draftQuery
  -> first-page query

popstate
  -> parse location.search
  -> appliedQuery
  -> draftQuery
  -> close detail
  -> reset continuation
  -> first-page query
```

The Task 8 falsification matrix now explicitly requires:

```text
refresh preserves applied query
browser back/forward navigates applied queries
pasted canonical query string reproduces investigation
visible canonical query evidence matches applied URL/query state
```

No production router or new Product authority is introduced.

## 4. M-1 — chip granularity

Resolution in the ratified P7 design and P8 plan:

```text
Period chip
  = occurred_at_from + occurred_at_before

USER Actor chip
  = actor_kind=user + actor_user_id

exact Resource chip
  = resource_kind + resource_id
```

Removing a chip clears the full semantic dimension immediately. The UI cannot manufacture wire-invalid dependent combinations through chip removal.

## 5. M-2 — relative period labels

Resolution:

```text
Hoje / Últimos 7 dias / Últimos 30 dias
  = draft-editor conveniences only

applied authority
  = exact canonical UTC from/before interval

refresh/copied URL
  = reconstruct exact interval label in local presentation time
  != recalculate relative preset name from the later clock date
```

The same canonical URL therefore preserves the same human question over time.

## 6. M-3 — Area Query Assist failure

P8 review controls now include:

```text
failAreaAssist
failActorAssist
failResourceAssist
```

Area/Actor/Resource Query Assist must each prove distinct:

```text
loading
known-empty
failure + retry
```

Failure never masquerades as an empty option set.

## 7. M-4 — truthful op87 fixtures

Every simulated Area option must have at least one corresponding historically visible Audit fixture event.

```text
op87 option
  -> identity occurs in admitted Audit evidence fixture
```

This keeps the P8 fixture model inside the durable op87 candidate law rather than inventing server behavior.

## 8. M-5 — max-20 refinement affordance

When simulated op88/op89 returns exactly 20 results, the UI may say:

```text
Mostrando até 20 opções. Refine a busca para localizar outro resultado.
```

It must not say or imply:

```text
more results definitely exist
has_more=true
page 2 available
```

because no such authority exists.

## 9. M-6 — all-human actor category

Explicit Launch disposition added to P7:

```text
actor_kind=user without actor_user_id
REJECTED — Product reason
```

Reason:

```text
ratified point investigation
  -> exact USER identity

SYSTEM investigation
  -> closed SYSTEM actor

human-vs-system category analysis
  -> no distinct Launch Auditor job proven
```

This is a Product/YAGNI rejection, not an API-absence shortcut. A future proven automation-vs-human analysis job may trigger the smallest lawful reopen.

## 10. M-7 — History API hardening

P8 History API use remains Evidence only.

The plan now requires:

```text
try/catch around pushState / replaceState
visible canonical query string in review evidence
local interaction remains operable if History API update fails
```

This does not create production routing architecture.

## 11. Method compliance after adjudication

The adjudicated package preserves the Method v2.3 P7 exit laws:

```text
fields/summaries           PRESENT-IN-AUTHORITY
identity sources           PRESENT-IN-AUTHORITY
pagination/scale           PRESENT-IN-AUTHORITY
sort/filter                PRESENT-IN-AUTHORITY or explicit REJECTED/DEFERRED
preview/content truth      PRESENT-IN-AUTHORITY
material writes            none in Audit business domain
unresolved upstream finding 0
```

The Fable review itself concluded that P7 genuinely demonstrated compliance rather than merely declaring it.

## 12. Review-round disposition

A second Fable round is not required by the current evidence because the reviewer stated that:

```text
I-1 is the sole blocker to PASS
I-1 requires only a small P8-plan amendment
P7 itself stands
no upstream reopen is required
once operator adjudicates I-1, no BLOCKING/IMPORTANT item remains
```

The operator accepted I-1 and all seven Minors, and the corresponding planning artifacts were amended on the candidate branch.

This does not retroactively change Fable's historical review verdict. PR #166 remains truthful Evidence of `HOLD BEFORE P8` against base `b6295750...`; the candidate branch records the later operator adjudication and corrections.

## 13. Gate

After exact-HEAD diff/CI verification of this adjudicated package:

```text
B09 P7                 CLOSED / OPERATOR-RATIFIED
Fable review           COMPLETED / OPERATOR-ADJUDICATED
unresolved BLOCKING    0
unresolved IMPORTANT   0
P8 plan                READY
P8 HTML                NOT STARTED
P8 execution           requires explicit operator authorization
P9-P10                 NOT OPEN
B10-B12                NOT OPEN
T12                    NOT OPEN
Product implementation BLOCKED
merge                  NOT AUTHORIZED
```

PR #166 must be closed unmerged after the candidate correction package is verified. Its `ai-dialog.md` must never enter `arch/t11-implementation-program` or `main`.