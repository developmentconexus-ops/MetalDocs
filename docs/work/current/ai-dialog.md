# Fable Review — T11 Discussion / Mention / Notifications bounded reopen

## Review target

Repository: `developmentconexus-ops/MetalDocs`  
Candidate branch: `arch/t11-implementation-program`  
Exact candidate HEAD: `a9047924aa2e31aaa1418a15c8786b7e9ad2967f`  
Required candidate CI: `#1263 SUCCESS`  
Review branch: `review/t11-notifications-fable`

Follow the canonical DevelopmentConexus Fable workflow and METHOD v1.0.0.

Read repository authority in the normal bounded order:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ the smallest owning authority/work package needed for this review
```

Primary candidate package:

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

Current upstream authorities should be read only as required to challenge the candidate, especially:

```text
docs/product/contract.md
docs/product/journeys.md
docs/architecture/ownership.md
docs/architecture/domain-model.md
docs/architecture/authorization-and-audit.md
docs/architecture/async-and-search.md
docs/architecture/backend.md
docs/architecture/interfaces.md
docs/architecture/persistence.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
docs/architecture/runtime.md
```

The current integrated authority remains 4 business + Audit, 10 SPA routes, 78 application operations, 10 Idempotency-Key creations. The package under review is a bounded material reopen candidate; do not treat its new counts as already promoted authority.

## Candidate result to attack

```text
Product
  stable-Document Discussion
  immutable DiscussionMessage
  chronological timeline + optional one-message reply reference
  semantic Mention(user_id)
  in-app Notification for explicit Mention
  no Launch email/push

Authorization
  new document.discuss Permission for writing
  reading Discussion follows current Document disclosure
  Mention targets need current ability to receive/read exact Discussion context

Ownership
  Notifications becomes second supporting semantic owner
  4 business + 2 supporting owners

Inbox
  seen_at monotonic
  read_at reversible
  archived_at reversible
  unseen badge
  unread filter/count
  archive/unarchive
  mark all read

Frontend IA
  global bell + Quick Inbox
  stable /notifications route
  sidebar unchanged

Wire candidate
  +3 Discussion/Mention operations
  +4 Notification state operations
  +1 SSE invalidation operation
  78 -> 86 application operations
  10 -> 11 Idempotency-Key creations
  13/13 ETag domains unchanged
  4 exact-byte resources unchanged

Atomicity
  accepted explicit Mention <=> required DOCUMENT_MENTION Notification exists
  one local PostgreSQL Scope coordinated by application
  Authorization alone decides final ALLOW/DENY
  protected author + target eligibility serializes with offboarding

Disclosure
  Notification source disclosure is recomputed server-side
  presentability is applied before public pagination/counts
  no copied ACL/presentable authority

Realtime
  SSE is best-effort invalidation only
  transport -> application -> narrow realtime mechanism
  in-process coalescing hub for one-replica Launch
  wake only after committed Notification state changes
  River remains the only durable future-work mechanism
  no generic EventBus / external broker / Redis baseline

Mechanisms
  Lexical core + @lexical/react for PlainText + custom Mention node/typeahead
  Lexical state never persists as Product truth
  native MetalDocs Inbox over OpenAPI + TanStack Query
  native browser EventSource
  narrow Go stdlib SSE realization
  LISTEN/NOTIFY only as a future multi-replica candidate
  Watermill only as a future EventStore candidate if real multi-consumer pressure appears
```

## Lead GCR history

Lead GCR Round 1 found `MATERIAL=3 / IMPORTANT=6`; operator approved all corrections. Corrected D7/technology records now encode them. Lead GCR Round 2 says `CONVERGED / MATERIAL=0 / IMPORTANT=0`.

Do **not** accept that self-review on trust. Reconstruct and attack it independently.

The three former material classes were:

```text
M1 Authorization ownership + offboarding serialization
M2 current disclosure before Notification pagination/counts
M3 SSE call graph + post-commit invalidation
```

The former important classes covered batch-seen disclosure, completed idempotency replay, Audit/History duplication, persistence constraints, OpenAPI SSE proof and visual/upstream sequencing.

## Mandatory adversarial questions

### 1. Root cause / Product boundary

Is stable-Document Discussion actually the smallest coherent product concept, or is it accidentally recreating DRAFT editor comments, SubmissionFeedback, chat, activity feed or generic collaboration?

Does Discussion ownership at stable Document survive Revision changes and first-release/no-release states correctly?

### 2. Notifications ownership

Does `seen/read/archive` truly justify a second supporting semantic owner, or is this an unnecessary owner boundary? Test Controlled Documents, Organization and rebuildable-projection alternatives seriously.

If Notifications is a real owner, is the proposed boundary complete enough without becoming a generic notification platform?

### 3. Authorization/disclosure

Attack the exact split:

```text
Organization -> subject facts
Controlled Documents -> resource predicate facts
Authorization -> final ALLOW/DENY
application -> choreography
```

Look for hidden second permission matrices in Mention autocomplete, Discussion reads, Notification presentability, batch seen, direct engagement, Quick Inbox or SSE.

Test offboarding/access races and whether deterministic protected-user acquisition is actually sufficient under current T3/T8-D laws.

### 4. Cross-owner transaction

Try to falsify the same-Scope Mention -> Notification invariant.

Is synchronously creating Notifications in the same PostgreSQL transaction the Global Maximum, or would an outbox/EventStore/River subscriber produce lower total complexity or better recovery?

Conversely, would introducing an event bus here be accidental complexity?

### 5. Pagination/count correctness

Attack the candidate-scan -> batch disclosure -> continue-until-presentable-page algorithm.

Can it preserve deterministic cursor semantics, `has_more`, unseen/unread counts, non-disclosure and bounded work without copied access authority?

Look for denial-of-service/scale traps or observable cardinality leaks. If optimization is required now, prove the current candidate structurally insufficient rather than assuming future scale.

### 6. Idempotency

Attack the 11th Idempotency-Key operation.

Does completed replay correctly recheck only current caller authorization/disclosure without rerunning historical Mention-target eligibility? Is the fingerprint sufficient and free of replay ambiguity? Can retries ever duplicate Notification or expose now-forbidden source truth?

### 7. Message immutability / retention

Is no-edit/no-delete the correct Launch Global Maximum after mentions/replies/notifications, or does it create unreasonable user harm? If edit/delete is required, specify the smallest semantics that preserve truthful notifications/replies.

Check offboarding/profile erasure and future retention/privacy compatibility.

### 8. Inbox engagement

Attack `seen_at`, `read_at`, `archived_at` as independent dimensions, badge=`unseen`, unread reversible, archive reversible, no mark-unseen.

Look for contradictory states, concurrency needs, duplicate counters or conflict with future Read & Acknowledge.

### 9. API surface / census

Try to reduce the candidate 8 new operations without:

```text
screen-shaped APIs
generic /actions
frontend AuthZ
hidden semantics
per-item network explosion
loss of deep-link/realtime behavior
```

Also test whether any missing operation/read model makes the 86 count incomplete.

### 10. SSE / runtime

Attack whether SSE belongs in `/api/v1` application census, whether the OpenAPI/codegen proof is realistic, and whether the call graph preserves the one semantic inbound door.

Challenge native EventSource, Go stdlib SSE, heartbeat/reconnect/resource limits, one-replica in-process hub, multi-tab behavior and post-commit wake-up.

Could polling be materially smaller? Could WebSocket be required? Could LISTEN/NOTIFY be needed now? Require evidence.

### 11. Technology choices

Challenge Lexical against Tiptap/ProseMirror, a smaller contenteditable/textarea solution, react-mentions-style libraries and custom editor code.

Challenge native Inbox against Novu/Knock/MagicBell or another mature reusable library/service. Mechanism reuse must not import competing Product authority.

Challenge the absence of Watermill/EventBus/broker/Redis under both current and foreseeable declared requirements.

### 12. Persistence / proofability

Check whether current T8-D architecture can admit a `notifications.*` owner namespace, identity-only cross-owner references, message immutability, same-Document reply constraint, Notification uniqueness and engagement constraints without violating foreign-SQL laws.

Call out any invariant that cannot actually be enforced or falsified.

### 13. B01/B03 UX coherence

Does bell + Quick Inbox + full `/notifications` preserve `Minha Caixa = assigned work`, or is the global IA still locally optimal rather than globally coherent?

Does Discussion belong on B03 ficha rather than the exact-content viewer/work/governance/history lenses?

### 14. YAGNI / Global Maximum

Attack both directions:

```text
under-design: retrofit traps hidden by “YAGNI”
over-design: owner/framework/event/realtime machinery not required by current consumer
```

Do not reward complexity merely because it resembles large-scale platforms.

## Required verdict

Use exactly:

```text
VERDICT: CONVERGED | NOT CONVERGED
MATERIAL: <count>
IMPORTANT: <count>
OPTIONAL: <count>
UNSUPPORTED_PREFERENCE: <count>
```

For each MATERIAL/IMPORTANT finding provide:

```text
finding id
exact candidate/authority location
concrete counterexample/failure mode
protected property
why current candidate is insufficient
smallest sustainable correction
whether it changes Product capability, owner count, route count, operation count, Permission count, mechanism choice, or only enforcement precision
```

Explicitly state whether the following survive your challenge:

```text
4+2 owners
11 stable SPA routes
16 PermissionCode values
86 application operations
11 Idempotency-Key creations
same-Scope Mention -> Notification
server-side presentability before paging/counts
Lexical
SSE + in-process wake-up
River as sole durable async
no generic EventBus/broker/Redis
```

Do not implement code. Do not modify any file except this review dialogue. Reviewer output is Evidence, never authority.

---

## Fable response

### 0. Review identity and verified premises

Fresh independent Fable challenge under METHOD v1.0.0 and Repository Standard v1.0.0. This output is Evidence, never authority.

Verified before evaluation:

```text
candidate HEAD                          a9047924aa2e31aaa1418a15c8786b7e9ad2967f  (confirmed)
candidate CI                            #1263-class run on arch/t11-implementation-program SUCCESS (confirmed via gh)
review branch − candidate               docs/work/current/ai-dialog.md only (git diff --stat: 1 file)
review CI                               SUCCESS on review/t11-notifications-fable
PR #165                                 DRAFT, base=arch/t11-implementation-program, evidence-only (confirmed)
```

Authority reconstructed through the normal route: `AGENTS.md → docs/index.md → docs/roadmap.md`, then the full candidate package (8 files) and the implicated upstream authorities: Product contract, journeys (§5/§6/§29/§31), ownership, domain model (T1), T3, T5, T8-B, T8-C, T8-D (laws + §22 lock law + transaction census), T8-E wire SSOT (full), T8-F, T8-G (topology/replica/consumer census), `decisions/api-operation-census.md`, `decisions/forward-obligations.md`. External METHOD.md and REPOSITORY-STANDARD.md read at v1.0.0 from the methodology repository. The Lead GCR was read last and treated as a claim to falsify, not as input.

### 1. Independent re-falsification of the Lead corrections

I re-derived each former MATERIAL class from upstream authority before reading the Lead's closure text, then attacked the corrected candidate.

**M1 — Authorization ownership / offboarding serialization.** The corrected D7 flow is exactly the T8-C D07/D08/D29 partition: Organization authors subject facts (protected form), Controlled Documents authors predicate facts, `Authorization.DecideIn/DecideManyIn` alone computes ALLOW/default-DENY, application choreographs. The multi-user acquisition (`unique actor + targets → dedupe → user_id ASC → strongest mode`) is byte-compatible with the already-ratified T8-D §22 User-lock-ordering law, so no new deadlock class is introduced and offboarding (single `FOR UPDATE` root) cannot form a cycle against ASC `FOR SHARE` sets. Residual race I attacked: a concurrent **non-offboarding access revocation** (RoleAssignment revoke / GroupMembership removal takes no mention-target lock) can commit between eligibility check and message commit, yielding a Notification whose target just lost access. This is not a defect: T3 §11 explicitly scopes serialization to eligibility/offboarding, tolerates linearized-first commits, and D5 presentability fails closed at read time with zero metadata leak. The invariant protected (no false "target was notified" success, no disclosure) survives. **CLOSED — confirmed independently.**

**M2 — presentability before pagination/counts.** The candidate-scan → batch-facts → Decide/DecideMany → retain → bounded continue-until-page+lookahead algorithm is the only structure compatible with all four constraints simultaneously (§2.7 cursor law, no sparse pages, no copied ACL, no frontend post-filter). I attacked: (a) cursor determinism — cursor seeks on canonical Notification ordering, not presentable ordinal, so disclosure drift between pages cannot corrupt the seek; (b) `has_more` — lookahead of one presentable candidate decides it deterministically; (c) oracle surface — hidden candidates are never returned and counts are computed over the same presentable composition; (d) worst-case O(all-candidates) scan for a fully-revoked recipient — real, but bounded per scan, Launch-scale-appropriate, with a measured-scale reopen trigger and a rebuildable-projection successor already named. Declaring the current candidate structurally insufficient would require scale evidence that does not exist; forcing a durable presentability projection now would duplicate AuthZ authority — exactly the defect M2 removed. **CLOSED.**

**M3 — SSE call graph / post-commit wake-up.** `transport/http → application/notifications → narrow subscription/wake port → platform realtime` conforms to T8-B §5/§9 (transport→application only; platform owns no semantics; consumer-owned port per T8-C §2.2). The wake-up census (creation + all six engagement transitions, post-commit only, best-effort, never owner-invoked, never River-retried) is complete against the operation set: every state change observable by another tab is listed. Wake-up failure or rolling-deploy replica overlap (two processes transiently alive; old-process subscribers miss cross-process wakes) cannot violate any invariant because SSE is non-authoritative and canonical GET reconciles — consistent with T8-G's "overlap must not corrupt semantics". **CLOSED.**

**Former IMPORTANT classes (I1–I6)** re-attacked: batch-seen intersection law forbids per-id **and cardinality** detail (single-id probe oracle is closed only if the response stays count-free — consolidation must encode 204/no-count, which the current wording already requires); completed replay correctly instantiates the pre-existing global replay law (§2.5 "historical lifecycle/preconditions NOT re-run") — re-running historical target eligibility would be a *new* second mention command, so the correction is the law, not an exception to it; Audit/History disposition mirrors the SubmissionFeedback precedent and T3 §15 already lists "notification delivery" as non-mandatory Audit; persistence obligations (unique per recipient+message, `read_at→seen_at`, `archived_at→seen_at`, immutability, same-Document reply via composite FK) are all expressible with T8-D-class declarative constraints; OpenAPI SSE proof is honestly held as a closure gate with unknown-stays-unknown posture; sequencing keeps B01 P8 re-LOCK operator-owned. **All CLOSED.**

### 2. Adversarial sweep — the fourteen mandated attacks

**Q1 Product boundary.** Stable-Document Discussion is the smallest coherent concept: T1 explicitly excludes DRAFT EditorialComment; SubmissionFeedback binds an exact immutable Submission; GovernanceCase feedback binds an attempt; none can host cross-Revision conversation without semantic corruption. `official_revision_at_post` as contextual snapshot (absent pre-first-Release) keeps Discussion at Document identity across Revision changes, no-release, obsolescence and cancellation; read access degrades/fail-closes with the lens in every state. Discussion on OBSOLETE/CANCELLED documents remains admissible (no state predicate forbids it) — deliberate and harmless; a future predicate is additive.

**Q2 Notifications ownership.** The promotion argument survives inversion: `seen/read/archive` is recipient-controlled state that is (a) not rebuildable from the Mention (kills alternative C), (b) not document meaning (kills A — Controlled Documents would own recipient-personal inbox state that outlives any document source family), (c) not organizational identity (kills B). A projection-only Notification would have been a mechanism; the ratified persistent engagement lifecycle is what earns the owner. The boundary is complete without platformization: closed kind union, identity-only source refs, no delivery/transport/copied-source ownership. **4+2 justified by current lifecycle, not by speculative source count.**

**Q3 Authorization/disclosure.** I hunted a hidden second permission matrix in all seven named surfaces. None exists: autocomplete composes Organization facts + CD predicates + Decide; list/counts/batch-seen/engagement reuse one presentability composition; Inbox self-access is structural recipient scoping (precedent: ops 54/55 My Work projections with `B` problem set), not an application-computed grant; SSE carries zero decisions. `document.discuss` for `viewer` widens the default viewer posture — operator-adjudicated, and `governance_viewer`/`governance_admin` correctly preserve their deliberate postures. Mention eligibility deliberately not requiring `document.discuss` preserves the D1 read/write split.

**Q4 Cross-owner transaction.** Same-Scope creation is the Global Maximum, not a local one: the protected property is a **biconditional** (accepted Mention ⇔ Notification exists), and only same-transaction composition enforces it with zero new machinery. An outbox/River subscriber converts the biconditional into eventual consistency, adds a consumer + retry + dedup surface, and still needs the same uniqueness constraint as backstop — strictly more complexity for a weaker invariant. An EventBus for one producer→one consumer inside one process violates the METHOD abstraction test on all five questions. Conversely nothing in the candidate smuggles a bus in: the wake port is signature-narrow (`Subscribe/Wake(user_id)`), carries no payloads, and cannot grow into one without breaking its port contract.

**Q5/Q6** — covered under M2/I2 above; additionally the fingerprint (document + reply ref + ordered normalized Text/Mention sequence) is collision-exact over everything that changes command effect, ReplaySnapshot=`message_id` satisfies self-contained reconstruction (D36) for a `201 {message_id}` body, and the `unique(recipient, message)` constraint backstops any replay/concurrency residue against duplicate Notifications.

**Q7 Immutability.** Immutable-accepted-message is correct for Launch: after Mention→Notification and reply references exist, edit/delete drags in retraction, re-notification, tombstone and edit-history semantics with no named consumer (alternative B is coherent but unpurchased). Correction-as-new-message plus the named reopen triggers (real usage harm, moderation/legal, retention) is the smallest sustainable form. Erasure compatibility holds through stable `user_id` + erasable profile + neutral rendering; hard-deletion is honestly deferred as a material privacy/retention reopen, not claimed solved.

**Q8 Engagement.** The three independent timestamps admit no contradictory state: direct-engagement⇒seen kills archived-but-unseen; read⇒seen and archived⇒seen become CHECKs; mark-unread preserves seen so the badge (unseen) cannot be resurrected — badge and unread-filter semantics stay disjoint by construction. Per-property last-accepted-wins without an ETag domain is right for personal reversible state (no cross-property lost update is possible; 13/13 domains correctly unchanged). No stored counter becomes authority. `Notification READ ≠ Read & Acknowledge` keeps the Launch+ seam clean.

**Q9 API surface.** I attempted reduction and expansion. Reduction: folding 84 (batch seen) into 83 explodes per-item network on viewport presentation; folding 85 (mark-all) into anything generic creates the forbidden `/actions` shape; counts already ride the first list page instead of a ninth operation; message deep-link correctly rides an anchor navigation mode on 79 instead of a message-detail operation; the bell reuses list-page-1 counts instead of a count-only operation. Expansion: no accepted flow is orphaned — compose (81→80), read (79 + anchor), inbox (82–85), freshness (86), deep-link navigation to the existing B03 route. **86 is minimal-complete; 11th Idempotency-Key creation is exactly the one non-idempotent semantic creation.**

**Q10 SSE/runtime.** Keeping 86 inside the `/api/v1` census is correct: unlike OIDC (non-application semantics), the stream is session-authenticated recipient-scoped application behavior, and excluding it would legalize exactly the manual-parallel-route drift the wire SSOT exists to kill; the OpenAPI server-side `text/event-stream` proof is properly a closure gate with a bounded tooling-reopen fallback. Polling would drop operation 86 but buys permanent per-tab interval load, worse freshness, and still needs focus/refetch logic — not materially smaller, and the operator ratified the realtime UX. WebSocket is unjustified (server→browser only). LISTEN/NOTIFY now would be speculative-scale machinery for a one-replica baseline; the mechanism-swap seam is the correct preparation. EventSource (native, GET, cookie, no CSRF needed on safe method) is the smallest client.

**Q11 Technology.** Lexical core + `@lexical/react` + owned MentionNode beats: Tiptap/ProseMirror (imports a richer document model than `Text|Mention(user_id)` needs), react-mentions lineage (string-markup identity — precisely the display-text-as-authority defect D3 forbids), and bare textarea/contenteditable (either reintroduces markup identity or hand-rolls IME/a11y/composition at higher proof cost than the substrate it avoids). The boundary law (EditorState/JSON never persisted; wire = closed segment sequence) makes it replaceable with zero content migration — mechanism, not authority. Native Inbox over Novu/Knock/MagicBell is compelled by the baseline: Novu self-host imports MongoDB+Redis+worker topology against a one-monolith/one-PostgreSQL authority; SaaS inboxes would hold Product state externally. Their state models were correctly consumed as evidence only. Watermill/broker/Redis absence is correct under both current and declared-foreseeable requirements; triggers are named and non-preferential.

**Q12 Persistence/proofability.** A `notifications.*` schema extends the closed catalog through the bounded reopen; identity-only cross-owner FKs (recipient→`org.users`, source→`controlled_docs` message identity) conform to the T8-D law; every named invariant is declaratively enforceable (partial-unique per kind, timestamp-implication CHECKs, composite `(document_id, message_id)` FK for same-Document reply, privilege-enforced immutability per the Audit precedent). No unfalsifiable invariant found: every D7 §13 proof line names a concrete negative fixture class.

**Q13 B01/B03 UX.** Bell + Quick Inbox + `/notifications` preserves the B01 mental model precisely because attention ≠ assigned work (`Minha Caixa` untouched); the four rejected IA alternatives are correctly rejected; discoverability rides the persistent utility-header bell (industry-consistent). Discussion belongs on the B03 ficha, not the content viewer (viewer is content-only by operator correction), not work/governance/history lenses (each owns different truth). Notification deep-link into the existing B03 route with message-highlight as navigation context correctly avoids a 12th route.

**Q14 YAGNI both directions.** Under-design retrofit traps are all seam-covered: email/push (persistent intent + River per T5), new kinds (closed-union expansion), multi-replica (wake-mechanism swap), edit/delete (analyzed alternative B + trigger), count scale (rebuildable projection + trigger). Over-design: the second supporting owner, SSE machinery and Lexical each trace to a ratified current consumer; nothing rewards platform resemblance.

### 3. Findings

**F-1 — IMPORTANT — consolidation-target map omits the decision-register authorities that hard-code the superseded state**

```text
finding id                 F-1
location                   docs/work/current/t11-b03-notification-ownership-reopen.md §9 (downstream coherence list)
                           + docs/roadmap.md "Exact next action" step 4
                           versus docs/decisions/forward-obligations.md ASY-02
                           and docs/decisions/api-operation-census.md
counterexample             Execute consolidation exactly as the written target list
                           (Product / T1 / Ownership / T3 / T5 / T6 / T8-B→G / T9 / T11).
                           Afterward, docs/decisions/forward-obligations.md still asserts
                           "ASY-02 — DEFERRED — Notifications remain a delivery-projection
                           concept only; no Launch inbox ... machinery", and
                           docs/decisions/api-operation-census.md still asserts
                           "current census 78" + "T8-E must use this 78-operation census",
                           while promoted authority asserts a Notifications owner and 86.
                           Two current authorities then contradict one meaning.
protected property         Repository Standard §2/§8: one current authority per meaning;
                           decision dispositions must be current; a stale DEFERRED/78 register
                           is a live false authority for the next fresh actor.
why candidate insufficient The package's own consolidation maps (ownership-reopen §9, D7 §1,
                           roadmap step 4) never name the two register documents; ASY-02 is
                           cited nowhere in the package, so no mechanical trace forces its flip.
                           (journeys §29 does route to the census doc, so that half is likely
                           caught; ASY-02 has no inbound reference from the candidate at all.)
smallest correction        Add the two explicit targets to the consolidation scope: flip
                           ASY-02 (DEFERRED → REFINED/SUPERSEDED by the ratified reopen) and
                           update docs/decisions/api-operation-census.md (78 → 86 with the
                           bounded-reopen basis) in the same consolidation commit. One list
                           edit now; no candidate semantics change.
impact                     enforcement precision only — zero change to Product capability,
                           owner count, route count, operation count, Permission count,
                           or mechanism choice.
```

**O-1 — OPTIONAL** — `anchor_message_id` navigation mode: wire consolidation must close ordering/continuation direction under the §2.7 cursor law (anchor as first-page filter; continuation cursor authenticates the same operation; one named fixture), so deep-link paging cannot drift into a second pagination authority.

**O-2 — OPTIONAL** — wire consolidation should bound message composition explicitly (max segments / max unique Mention targets per message) rather than relying only on the 65,536-byte JSON ceiling; the bound also caps protected-lock fan-out and per-message Notification fan-out.

**O-3 — OPTIONAL** — at T3/T6 consolidation, bind "ability to receive/read the exact Document Discussion" to one named citable predicate (the Document Official lens disclosure rule), so Discussion read and Mention eligibility share a single referenced source instead of a described one; and record in T8-G that transient deploy-overlap wake loss is tolerated because SSE is non-authoritative.

No UNSUPPORTED_PREFERENCE findings: every mechanism I would personally weigh differently (polling-only Launch, textarea composer) is either operator-ratified with a defensible basis or demonstrably not smaller.

### 4. Verdict

```text
VERDICT: CONVERGED
MATERIAL: 0
IMPORTANT: 1
OPTIONAL: 3
UNSUPPORTED_PREFERENCE: 0
```

F-1 is a bounded consolidation-completeness correction inside the already-planned consolidation step; it does not change the candidate result and does not require a Round 2 unless the Lead contests it.

Explicit survival statement — all challenged and **SURVIVE**:

```text
4+2 owners                                   SURVIVES
11 stable SPA routes                          SURVIVES
16 PermissionCode values                      SURVIVES
86 application operations                     SURVIVES
11 Idempotency-Key creations                  SURVIVES
same-Scope Mention -> Notification            SURVIVES
server-side presentability before paging/counts SURVIVES
Lexical                                       SURVIVES
SSE + in-process wake-up                      SURVIVES
River as sole durable async                   SURVIVES
no generic EventBus / broker / Redis          SURVIVES
```

Reviewer output is Evidence, never authority. Adjudication of F-1 and the OPTIONAL precisions belongs to the Lead/operator; upstream consolidation may proceed per the roadmap once F-1's one-line scope addition is adjudicated.
