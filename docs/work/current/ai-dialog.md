# Fable Review — T11 B09 Audit P7/P8 adversarial gate

## Review target

Repository: `developmentconexus-ops/MetalDocs`
Candidate branch: `arch/t11-implementation-program`
Exact candidate HEAD: `b62957501a56457a8f46cb081d24ee93f02aac45`
Required candidate CI: `required` SUCCESS on exact HEAD (confirmed via gh check-runs)
Review branch: `review/t11-b09-audit-fable`
Review brief: `docs/work/current/t11-b09-fable-adversarial-review-request.md` (canonical; read in full)

Follow the canonical DevelopmentConexus Fable workflow, Frontend Product Experience Planning Method v2.3 and the DevelopmentConexus Engineering Method. Reviewer output is Evidence, never authority. Only this file differs from the candidate.

---

## Fable response

### 0. Review identity and verified premises

Fresh independent Fable adversarial review of the B09 P7 exit and P8 realization plan before any P8 execution. Verified before evaluation:

```text
origin/main                          cae6ba48  (matches roadmap "opening main")
candidate HEAD                       b6295750  (matches handoff; checked out locally)
PR #162                              OPEN / DRAFT / base main / head arch/t11-implementation-program
CI on exact HEAD b6295750            required = SUCCESS (completed)
routing authority                    AGENTS.md → docs/roadmap.md sole status authority — consistent
census                               89 ops / op78 REFINED + op87-89 SAFE_READ — proof arithmetic checks (76+2+8+3)
review branch − candidate            docs/work/current/ai-dialog.md only
```

Full authority pack read: Method v2.3 (`docs/development/functional-html-wireframe-method.md`), `docs/roadmap.md`, `docs/decisions/audit-investigation-read.md`, `docs/decisions/api-operation-census.md`, `t11-b09-audit-upstream-replan.md`, `t11-b09-f1-rebaseline-proof.md`, `t11-b09-audit-r1.md`, `t11-b09-p7-exit.md`, `t11-b09-p8-realization-plan.md`, plus cross-block coherence artifacts `t11-b07-document-history-r1.md`, `t11-b08-notifications-full-inbox-r1.md`, `t11-b08-pattern-consolidation.md`.

### 1. Falsification work actually performed (not just re-read)

- Cross-checked every P7 claim in `t11-b09-audit-r1.md` against wire law in `audit-investigation-read.md`: first-page predicates, actor/resource dependency laws, `operation_codes` form/explode=false + canonical enum-order serialization, cursor-continuation law, order `occurred_at DESC,event_id DESC`, op87 no-params complete set, op88/op89 max-20 no-pagination, recognition non-filtering role, §11 handoff whitelist. No contradiction found.
- Tested the local-maximum direction both ways (weakened-to-fit-API vs invented-for-UI): op87–89 came from a ratified upstream reopen (B09-F1) with Product reasoning, bounded to Audit-visible candidates, no generic entity/search platform; conversely every rejection/deferral in the R1 §19 matrix carries a Product/scope reason, not "current API lacks it".
- Empirically tested the one plan mechanism that could silently break functional evidence: `history.pushState/replaceState` with query strings on `file://` origin (the operator's handoff model is a locally opened `.html`). Result: works in current Chromium (headless verification; `location.search` updated after both calls). The plan's History-API choice is viable for a locally opened artifact.
- Checked B09 against B08's rejected patterns (generic Activity/Event feed, generic Inbox framework, generic filter engine, generic deep-link resolver, generic realtime sync): B09 keeps the investigation bar Audit-specific, keeps handoffs to the bounded §11 whitelist, defers admin deep-links to B10–B12. No revival.
- Checked B09 does not become: analytics dashboard (rejected in matrix), Document History duplicate (Escopo histórico / Contexto atual labeling actively prevents conflation; History reached only via bounded handoff that reauthorizes), administration console (no writes; admin directories rejected), current-state inspector (recognition explicitly non-historical and secondary), generic enterprise search (free-text deferred; assist bounded to Audit-visible evidence), case management (deferred with reason).

### 2. Findings

#### BLOCKING

None.

#### IMPORTANT

**I-1 — Canonical-URL simulation is write-only in the P8 plan; the URL round-trip is never required or verified.**

```text
location                   docs/work/current/t11-b09-p8-realization-plan.md
                           Task 2 Step 2 + Task 8 Step 4
what is wrong              Task 2 requires serializing the applied query and calling
                           pushState/replaceState "to prove refresh/back-forward/query-copy
                           structure", but no task step requires deserializing
                           location.search into appliedQuery on load, none requires a
                           popstate handler, and the Task 8 Step 4 falsification matrix has
                           no line for refresh-preserves-applied-question,
                           back/forward-navigates-applied-questions, or
                           copied-URL-reproduces-the-question. P7 R1 §8 declares exactly
                           those three as binding properties of applied-URL truth.
why it matters             This is precisely the "visually plausible HTML without operable
                           material interaction" failure the plan must be rejected for. A
                           builder can satisfy every written checkbox while the URL remains
                           decorative output — F5 silently loses the applied investigation,
                           and the H1 pillar "applied query = URL + chips + ledger truth"
                           ships unfalsified. Task 8's matrix is the handoff gate; anything
                           absent from it can be claimed ready untested.
smallest lawful correction Plan amendment only; no upstream reopen; no authority change.
                           (a) Task 2: add explicit step "on load and on popstate, parse
                           location.search into appliedQuery and run the first-page query".
                           (b) Task 8 Step 4: add three matrix lines —
                               refresh preserves applied query
                               back/forward navigates applied queries
                               pasted canonical query string reproduces the investigation
```

#### MINOR

**M-1 — Chip granularity for dependent predicates unspecified.** `t11-b09-audit-r1.md` §6.1 makes applied-chip removal an immediate query action but never states whether `resource_kind`+`resource_id`, `actor_kind=user`+`actor_user_id` and the period pair are one compound chip or separate chips. Separate chips would let removal produce wire-invalid combinations (`resource_id` without `resource_kind` → 400 per authority §5.2). Correction: one compound chip per query dimension; removing it clears the whole dimension.

**M-2 — Period-preset chip labels are apply-time-relative but the URL is canonical instants.** R1 §7.1 converts `Hoje`/`Últimos 7 dias` at Apply time; after reload/copy only UTC instants exist, so a chip re-labeled `Hoje` tomorrow would be false. Correction: chips reconstructed from URL render the exact interval; preset names are draft-editor conveniences only.

**M-3 — `reviewMode` lacks an area-assist failure toggle.** Plan Task 1 Step 3 defines `failActorAssist`/`failResourceAssist` only, yet Task 6 Step 4 requires area assist to expose failure+retry, and the Task 8 matrix says only "Scope op87-style selection" with no failure line. Correction: add `failAreaAssist` + a matrix line.

**M-4 — Simulated op87 options must obey their own candidate law.** Plan Task 3 Step 2 offers COM/FIN/RH, but op87 candidates must occur in historically visible evidence (authority §7). If fixtures contain no RH event, the prototype demonstrates unlawful server behavior. Correction: every offered area has at least one fixture event.

**M-5 — Query Assist 20-item truncation has no affordance.** op88/op89 return max 20 with no pagination and no more-marker; an exactly-20 list can read as a complete candidate set. Frontend-only mitigation, no authority change: when 20 items return, show a refine-search hint. Not dispositioned anywhere; worth one line in P7/P8.

**M-6 — "All human actors" filter (`actor_kind=user` without id) is wire-inexpressible and was never dispositioned.** The actor law forbids kind-only user filtering while permitting kind-only SYSTEM — an asymmetry the R1 §19 matrix is silent on. No evidence of a material Launch job (both ratified jobs A/B are satisfied), so this is not an upstream finding — but silence is not a disposition under Method §13. Correction: explicit REJECTED/DEFERRED matrix row with the Product reason; escalate to upstream finding only if the operator asserts the job.

**M-7 — Defensive guard on History API (optional hardening).** Verified working on `file://` in current Chromium, but engines/policies vary. Cheap: wrap pushState in try/catch and always render the canonical query string as visible in-page text (which I-1 evidence benefits from anyway).

### 3. Assessment against the brief's core questions

```text
P7 genuinely satisfies Method v2.3        YES — real hypotheses (H2/H3 rejected with
                                          substantive reasons), real §19 disposition matrix
                                          with all four categories, all six §13 pre-P8
                                          declarations present; compliance is demonstrated,
                                          not merely declared
material Auditor need suppressed          NONE FOUND — total counts / custom sort /
                                          free-text / export / Company-only scope all carry
                                          Product or scope reasons; M-6 is the only
                                          undispositioned absence and lacks materiality
                                          evidence
API/platform capability invented for UI   NONE — op87-89 bounded, Audit-visible-only,
                                          ratified upstream; no detail endpoint, no
                                          resolver, no generic platform; census verified
Audit vs History vs current-state         HOLD — boundaries preserved and actively labeled;
                                          handoffs reauthorize at destination
op78/op87/op88/op89 + cursor semantics    EXACT AGREEMENT with durable authority
Query Assist vs Authorization             PRESERVED — assist is guidance; op78 rechecks
stable identity vs mutable recognition    PRESERVED — recognition never filters/reorders/
                                          authorizes; fallback chain truthful
scale/scanability/a11y/responsive         STRUCTURALLY COVERED (R1 §9/§16/§17, plan Task 7)
P8 plan produces functional evidence      YES with I-1 corrected — Task 8 matrix then
                                          materially exercises the brief's full list;
                                          forbidden-scope scan complete
premature generic frameworks              NONE — consistent with B08 P10 anti-abstraction
coherence with locked B01-B08 / B10-B12   PRESERVED — no unopened-block design, admin
                                          deep-links deferred
```

### 4. Verdict

```text
HOLD BEFORE P8

BLOCKING   0
IMPORTANT  1   (I-1)
MINOR      7   (M-1..M-7)
```

Sole blocker to a PASS: **I-1** — a small amendment to `t11-b09-p8-realization-plan.md` (URL deserialization on load/popstate + three Task 8 matrix lines). No upstream reopen is required; P7 itself stands. Once the operator adjudicates I-1 (amend, or explicitly disposition it), no BLOCKING/IMPORTANT item remains and P8 is lawfully executable. M-1–M-7 can be folded into the same amendment or resolved during the P8 iteration loop at operator discretion.

Hard constraints respected: no P8 HTML written, no Product/runtime code touched, no B10+/T12 opened, PR #162 untouched, B09 not marked LOCKED. Reviewer output is Evidence, never authority; adjudication belongs to the operator.
