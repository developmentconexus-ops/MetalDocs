# T11 — B09 Fable Adversarial Review Request

> **Status:** REVIEW REQUEST / P8 BLOCKING GATE.  
> **Block:** B09 — Audit.  
> **Reviewer:** Fable — independent/adversarial review requested by the operator.  
> **Method:** `docs/development/functional-html-wireframe-method.md` v2.3 + DevelopmentConexus Engineering Method.  
> **P7:** CLOSED / OPERATOR-RATIFIED, subject to independent adversarial challenge before P8 execution.  
> **P8:** BLOCKED pending Fable verdict + operator adjudication.  
> **Implementation:** BLOCKED.

## 1. Review instruction

Start fresh. Repository current authority > this request/handoff.

Before reviewing:

```text
revalidate main
revalidate arch/t11-implementation-program current HEAD
revalidate Draft PR #162
revalidate exact CI on current HEAD
read AGENTS.md / repository routing authority
read docs/development/functional-html-wireframe-method.md v2.3
read docs/roadmap.md
```

Do not assume the embedded history is current if the repository differs.

Stable review baseline before B09 P7 work:

```text
00d7779eae46a4d5f1d249699ce59f07f17fb704
```

Pre-review-request P7/P8-planning checkpoint:

```text
ab7bcb9353fbe396b983bca4ef28bb47c81bbff7
```

Use current branch HEAD as the actual review target.

## 2. Required authority pack

Review at minimum:

```text
docs/development/functional-html-wireframe-method.md
docs/roadmap.md
docs/decisions/audit-investigation-read.md
docs/decisions/api-operation-census.md
docs/work/current/t11-b09-audit-upstream-replan.md
docs/work/current/t11-b09-f1-rebaseline-proof.md
docs/work/current/t11-b09-audit-r1.md
docs/work/current/t11-b09-p7-exit.md
docs/work/current/t11-b09-p8-realization-plan.md
```

For cross-block coherence, inspect relevant locked B01–B08 artifacts where a B09 choice claims reuse or separation, especially:

```text
docs/work/current/t11-b07-document-history-r1.md
docs/work/current/t11-b08-notifications-full-inbox-r1.md
docs/work/current/t11-b08-pattern-consolidation.md
```

## 3. Review objective

Challenge whether B09 is truly ready to enter P8 under Method v2.3.

Do not optimize for agreement with the current proposal. Try to falsify it.

The review must answer whether the ratified P7 and P8 realization plan preserve all of the following simultaneously:

```text
human-job-first frontend planning
Global Maximum rather than local-minimum API preservation
YAGNI without suppressing a proven material need
NO screen-shaped backend
NO backend-shaped UX
Audit evidence != current-state reconstruction
Audit != Document History
server-complete filtering before pagination
stable identity != mutable recognition
Query Assist != Authorization
read-only Audit domain boundary
accessible + responsive interaction structure
P8 as functional Evidence only, not implementation
```

## 4. Adversarial questions — Method v2.3

### 4.1 P7 process integrity

Check whether P7 actually satisfies the method rather than merely declaring compliance:

```text
Were real Auditor jobs explicit before layout?
Were credible alternatives compared only where ambiguity was real?
Were task completion, scanability, density, cognitive load, context preservation,
accessibility, responsive viability, scale, error recovery and backend truth fit considered?
Did the leading hypothesis state fields/summaries, identity sources, pagination/scale,
sort/filter needs, preview/content truth and material writes?
Was every material requirement classified PRESENT / UPSTREAM FINDING / REJECTED / DEFERRED?
Is any rejected/deferred capability justified by Product evidence rather than current API absence?
Is there any hidden material need that should reopen Product/backend authority before P8?
```

Any material unmet authority need is a **BLOCKING UPSTREAM FINDING**, not a P8 fixture opportunity.

### 4.2 Global Maximum / YAGNI

Try both directions:

```text
Has the design been weakened to fit op78/op87-op89?
Has the design invented platform/generalization work without a proven Launch consumer?
```

Challenge especially:

```text
free-text search
export
saved searches
custom sort
total counts/page numbers
Company-only historical-scope filter
annotations/case management
generic entity/reference-data/deep-link abstractions
admin-directory selector dependencies
```

If any deserves a different disposition, explain the human job and whole-Product consequence.

## 5. Adversarial questions — Audit semantics and wire truth

Verify exact agreement with durable authority:

```text
op78 = sole Audit evidence traversal
op87 = Area Query Assist
op88 = Actor Query Assist
op89 = Resource Query Assist
```

Challenge:

```text
recent-first with no hidden time cutoff
fixed occurred_at DESC,event_id DESC ordering
first-page predicates vs cursor continuation law
actor_kind/user_id semantics
resource_kind/resource_id dependency
operation_codes multi-select semantics
historical Area semantics excluding Company-attributed events
Query Assist candidate completeness and non-authorization role
recognition never filtering/reordering/authorizing
recognition erasure/unavailability fallback
owner handoffs independently reauthorizing
no detail endpoint
no URL/links resolver in Audit wire
```

Flag any frontend behavior that would falsely claim evidence completeness or current-state truth.

## 6. Adversarial questions — P7 experience hypothesis

Try to falsify the **Audit Investigation Ledger** structure:

```text
recent-first entry
horizontal investigation bar + progressive disclosure
draft filters != applied query
applied chips + canonical URL
dense Evidence Ledger
local-day visual separators
explicit Load older events cursor continuation
Detail Drawer on desktop / full-surface detail on narrow screens
same actor/resource/action investigation shortcuts
secondary Current Context owner handoffs
```

Ask whether an Auditor can:

```text
scan recent evidence quickly
construct an exact evidence question without technical IDs as normal workflow
understand which query is currently authoritative
recover after invalid draft / Query Assist failure / main query failure / continuation failure
distinguish known-empty from unknown/failure/not-authorized
inspect exact canonical evidence without losing ledger context
understand historical scope vs current Area/resource context
operate with keyboard/screen-reader semantics
operate coherently on narrow/mobile layouts
continue through large histories without fake page/total metaphors
```

If another structure materially dominates H1 at whole-Product level, demonstrate why rather than preferring a different aesthetic.

## 7. Adversarial questions — P8 realization plan

Review `t11-b09-p8-realization-plan.md` as a falsification plan, not implementation design.

Confirm that it requires one browser-operable low-fidelity HTML artifact and materially exercises:

```text
recent-first default state
all five structured filter dimensions
draft/applied separation
canonical URL simulation
Period validation and presets
op87/op88/op89 Query Assist states
37-action local grouped selector
kind-first Resource Assist
recognition fallbacks
historical-scope presentation
ledger density + local-day separators
20-item cursor continuation + end/failure/retry
detail drawer/full-surface detail
typed facts semantic presentation
same actor/resource/action immediate investigations
admitted owner-lens boundaries
loading / known-empty / failure distinctions
route-authorization boundary review state
sticky/context preservation
responsive/mobile transformation
keyboard/focus/accessibility behavior
```

Reject the plan if it can produce a visually plausible HTML without actually making these material interactions operable.

Also verify the plan does **not** authorize:

```text
React/production frontend
real API/OpenAPI integration
real Authorization evaluator
business persistence
Product writes
new Audit endpoint/schema
B03/B06/B07 reimplementation
B10-B12 design
final visual design
generic filter/search/entity platform
```

## 8. Cross-product coherence challenge

Check B09 against the locked IA and prior patterns:

```text
Início       = current operational situation
Minha Caixa  = assigned work
Documentos   = official document truth / creation
Gestão       = system configuration
Evidência    = audit/evidence
```

Ensure B09 does not revive patterns previously rejected in B08 such as:

```text
generic Activity/Event feed
generic Inbox framework
generic filter engine
generic deep-link resolver
generic realtime entity sync
```

Check that B09 remains evidence-oriented and does not become:

```text
analytics dashboard
Document History duplicate
administration console
current-state inspector
generic enterprise search
case-management system
```

## 9. Required output format

Return findings ordered by severity:

```text
BLOCKING
  violates Method / authority / Product truth / P8 gate

IMPORTANT
  material design or plan weakness that should be corrected before P8

MINOR
  non-blocking clarity/precision issue
```

For every finding provide:

```text
exact file + section/line when possible
what is wrong
why it matters under Method v2.3 / Global Maximum / YAGNI / authority
smallest lawful correction or required upstream reopen
```

Then provide one explicit verdict only:

```text
PASS TO P8
```

or

```text
HOLD BEFORE P8
```

A PASS means no unresolved BLOCKING or IMPORTANT finding remains. A HOLD must identify the exact blocking/important item(s).

## 10. Hard review constraints

```text
do not implement P8 HTML
do not modify Product/runtime code
do not open B10+
do not open T12
do not merge PR #162
do not mark B09 LOCKED
do not treat reviewer preference as Product authority
do not weaken a proven user need merely because current API lacks it
do not invent backend capability merely because a reference product has it
```

The operator will adjudicate the Fable verdict/findings before P8 is unblocked.