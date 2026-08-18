# MetalDocs — Whole-Product Global Coherence Review

> **Status:** NON-AUTHORITATIVE GCR — OPERATOR ADJUDICATION PENDING  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **Method:** DevelopmentConexus Engineering Method v1.0.0  
> **Implementation:** BLOCKED  
> **Technical descent:** BLOCKED — R10-C remains paused

This review was performed **from the accepted Product Contract outward**. It does not repair the paused R10-C candidate, does not promote a replacement ownership topology, and does not author storage/schema/package/API/implementation design.

The review asks only:

```text
what product capability is required
→ what end-to-end journey must be true
→ what invariant follows
→ what complexity is essential vs accidental
→ what authority must exist
→ which prior decisions still deserve to survive
```

Only after operator adjudication may ownership/topology be re-derived.

---

# 1. Authority reconstruction

Current authority order for this review:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — **accepted product authority**
5. `wiki/architecture/whole-product-alignment-review.md` — active review routing
6. R9.5 frozen ledger — historical/product-domain evidence only where not contradicted
7. R10-A/B/C material — prior accepted/promoted/candidate architecture evidence only
8. runtime/schema/OpenAPI/code — current-state/migration evidence only when required

The Product Contract now decides Launch capability. Earlier architecture cannot restore a capability merely because that capability already has entities, permissions, transactions, lock laws or review history.

---

# 2. Executive verdict

## Whole-product Method outcome

> **RESTRUCTURE NOW — at the whole-product target-design level.**

This is **not** a rejection of the controlled-document kernel. It is a rejection of carrying the previous R10 ownership topology and downstream technical shape forward unchanged after the accepted product-scope reduction.

The following product kernel survives the review:

```text
single-company identity/access
Document Type + numbering
Controlled Document stable identity
Business Revision
mutable DRAFT Working Content
immutable Submission attempt
one sequential governance Step semantic
NoHumanApproval or governed human route
feedback + ACCEPT / RETURN_FOR_CHANGES
withdraw Submission attempt
cancel open Revision
system-owned Release/effectivity
EFFECTIVE / SUPERSEDED
optional required official Rendition
explicit governed obsolescence without replacement
revision/history
current-effective discovery/read/download
Audit
truthful historical migration/cutover
backup/restore correctness
```

The previous **8 business contexts + 3 supporting semantic owners** topology does **not** survive as a Launch conclusion because several owners now have no Launch consumer and one supporting owner (`Artifact`) exists primarily to mediate storage/content mechanics that the Product Contract explicitly keeps out of business identity.

R10-C remains paused and must not be repaired in place.

---

# 3. External-reference falsification evidence

Reference products were used only to falsify claims that a MetalDocs abstraction is unavoidable.

Verified current official documentation on 2026-08-18:

- SharePoint versioning/content approval distinguishes draft/pending content from published/approved reader-visible content, including cases where ordinary readers continue to see the last approved/major version while a newer draft exists:  
  https://support.microsoft.com/en-us/sharepoint/lists/documents-and-library/how-versioning-works-in-lists-and-libraries
- M-Files makes a normal document/object a template through an `Is template` property and maintains ordinary object version history; this supports a template **role**, not a parallel template lifecycle:  
  https://userguide.m-files.com/user-guide/latest/eng/using_template.html  
  https://userguide.m-files.com/user-guide/latest/eng/object_history.html
- Veeva QualityDocs exposes Periodic Review, Read & Understood/Training, Change Control and obsolescence-related processes as additional feature/process layers around document lifecycle; their existence does not imply those layers are prerequisites for a smaller controlled-document launch:  
  https://quality.veevavault.help/en/lr/72024/  
  https://quality.veevavault.help/en/lr/37406/  
  https://quality.veevavault.help/en/lr/16134/
- Qualio separately exposes document review/approval/effectivity/retirement, periodic review, controlled export and audit capabilities. This is evidence that mature platforms can layer those concerns rather than proof that MetalDocs must ship all of them together:  
  https://docs.qualio.com/en/articles/6526420-user-permissions  
  https://docs.qualio.com/en/articles/11122-audit-trail-overview

Reference conclusion:

> Mature platforms support the accepted core distinctions, but their additional QMS/records/training/change-control machinery is **falsification evidence against “we must also build it”**, not a checklist.

---

# 4. Finding GCR-01 — Product authority vs inherited architecture

## Evidence

R9.5 and R10 accumulated product semantics, technical owners and implementation-proof obligations in the same chain. After successive Launch reductions, the old target still carried owners/permissions/transactions for Distribution, Periodic Review, Dossier, Evidence, Records Governance, Governed Export and repository interchange.

The accepted Product Contract now places:

```text
Distribution / Read & Acknowledge → Launch+
Periodic Review                   → Launch+
Dossier                           → Future
Evidence                          → Future
Retention/Hold/Disposition        → Future
Governed Export                   → Future
External repository copy/sync     → Future
```

## Root cause

Prior accepted architecture was being treated as a source of product requirements.

## Target invariant

> Product capability authority is singular: the accepted Product Contract. Earlier architecture survives only where it still satisfies that contract and the Method.

## Outcome

**RESTRUCTURE NOW** — authority/routing already corrected. Prior R9.5/R10 material remains evidence, not a mechanism for resurrecting Launch scope.

Strongest counterargument: many prior decisions were independently reviewed and operator accepted.  
Answer: review quality proves those decisions against their then-current premises; it does not preserve premises invalidated by a later accepted Product Contract.

Reopen trigger: only a named Launch consumer, requirement or production failure mode.

---

# 5. Finding GCR-02 — Document ≠ Revision ≠ Working Content ≠ Submission

## Evidence

Every core journey depends on four different identity/mutation laws:

```text
Document        = stable official identity
Revision        = one business change cycle / non-reused ordinal
Working Content = mutable DRAFT truth + autosave/concurrency
Submission      = immutable exact governed attempt
```

Return/resubmit specifically requires the same Revision to produce another immutable Submission. A newer open Revision must coexist with the older EFFECTIVE Revision without changing reader truth.

SharePoint's draft/published separation is compatible evidence that moving work and published reader truth need not collapse into one version concept.

## Structural inversion

If any pair is merged, product semantics become harder, not simpler:

- autosave would consume business revisions;
- return would have to mutate submitted history;
- a new draft could replace reader truth too early;
- stable document identity would become a file/version identity.

## Outcome

**CURRENT STRUCTURE CONFIRMED** at product-semantic level.

This is essential complexity and remains the core architectural spine.

Reopen only on a material product counterexample, not table-count pressure.

---

# 6. Finding GCR-03 — Standalone `Artifact` semantic owner is accidental complexity

## Evidence

The prior design progressively gave `Artifact` responsibility for:

```text
exact-byte identity
hash / size / format
semantic confirmation
no-orphan law
single retention root
staging/promotion
physical-store mapping semantics
relocation/restore integrity
```

B5 then needed a closed global reference catalog and one-semantic-retention-root law. R10-C added `ArtifactStage`, opaque semantic-root binding, promotion and GC fences. Much of this complexity existed to make a separately-owned Artifact safe in the presence of Records disposition and cross-root reuse.

The Product Contract instead says:

```text
storage/provider identity never becomes Document/Revision/Submission identity
Submission freezes exact governed content
Rendition is an exact derived representation when required
confirmed governed history is preserved in Launch
Records disposition is not Launch
```

## Root cause

An exact-content **fact** was promoted into a standalone semantic **owner**, after which other domains needed relations and lifecycle rules merely to tell that owner what the bytes meant.

## Structural inversion

Without the standalone owner:

```text
Submission        owns the exact source-content facts it freezes
Rendition         owns the exact derived-content facts it freezes
imported Revision owns exact imported-content facts/provenance
Working Content   owns current mutable DRAFT content state
storage/staging   persists/proves bytes as mechanism
```

A provider-neutral storage handle may still be required technically. That does not create an independent business identity or bounded semantic owner.

Hash/size/format remain useful exact-content proof. They do not require a global Artifact aggregate.

## Outcome

**RESTRUCTURE NOW.**

Remove standalone `Artifact` semantic ownership from the Launch target. Do **not** infer from this that storage, integrity validation, malware inspection, no-overwrite or restore proof disappear; those re-enter later as technical mechanism requirements derived from the semantic records that actually freeze content.

Strongest counterargument: one Artifact row centralizes deduplication, integrity and references.  
Answer: provider deduplication is mechanism freedom; global semantic reuse created retention/root/late-reference complexity with no Launch consumer. Centralized integrity services do not require centralized semantic ownership.

Reopen trigger: a real independent content object lifecycle/consumer that cannot be represented as content of Document/Submission/Rendition/imported history.

---

# 7. Finding GCR-04 — R10-C discovered a valid autosave invariant but put it under the wrong authority

## Evidence

R10-C correctly found that autosave approximately every editor debounce must not create preserved governed history. Its `ArtifactStage` candidate separates reclaimable DRAFT candidates from immutable governed content and preserves server-side hash/format/malware checks at the governed boundary.

## Essential insight that survives

```text
DRAFT autosave != governed immutable history
provider upload success != governed confirmation
untrusted bytes require the production safety gate before governed use
provider location != business identity
restore must fail closed on missing/corrupt required bytes
```

## Accidental part

Making `ArtifactStage` part of an Artifact semantic ownership model, with semantic-root binding and a Stage→Artifact promotion lifecycle, is not justified once standalone Artifact authority is removed.

## Outcome

**RESTRUCTURE NOW** for the ownership/model; **CURRENT STRUCTURE CONFIRMED** for the safety/integrity principles.

Do not author the replacement physical design until ownership is re-derived after operator adjudication.

---

# 8. Finding GCR-05 — Template as ordinary governed Document survives; template-side mini-platforms do not automatically survive

## Evidence

Product requirement:

```text
current EFFECTIVE eligible Template
→ seeds a new independent Document
→ later Template changes never rewrite the derived Document
```

M-Files independently demonstrates the simpler pattern: an ordinary document/object can carry a template role while retaining normal version history.

## Outcome

**CURRENT STRUCTURE CONFIRMED**:

- Template is a role of a governed Document;
- source template must be the exact eligible current EFFECTIVE content at creation;
- derived Document becomes independently authoritative immediately.

**DEFER SAFELY unless a named Launch consumer is produced**:

- independent structured `TemplateSpec` product semantics;
- a generic structured-authoring/template-schema platform;
- retention-driven weakening of template provenance solely so source governed history can later be physically disposed.

Launch has no governed physical disposition, so B5's source-retention counterexample cannot force Launch provenance complexity.

---

# 9. Finding GCR-06 — One sequential governance Step survives; the current Approval feature surface does not inherit authority wholesale

## Confirmed core

B4's removal of structural `REVIEW | APPROVAL` Step types survives the Product Contract.

One Step semantic is sufficient for the required route:

```text
ordered governance Step
→ participant(s) inspect exact governed candidate
→ feedback
→ ACCEPT or RETURN_FOR_CHANGES
```

`NoHumanApproval` remains explicit and must never fabricate an approver.

## Unsupported inherited dimensions

The Product Contract does **not** independently establish Launch requirements for all current B4 dimensions:

```text
ANY vs ALL configurable quorum
NamedUser | Group | RoleInArea as a three-way policy language
strict global submitter/creator SoD
same-user-cannot-accept-two-Steps law
requires_reauthentication
fresh-auth evidence snapshots
due_in_days
approval.oversee
approval.reassign
administrative approval.cancel
structured annotation/suggestion payload schemas
provider-neutral anchored suggestion system
```

Some may be desirable. They are not entitled merely because prior R9.5/B4 already designed them.

At least one participant-selection rule is essential; the exact rule vocabulary/quorum/SoD model must be re-derived from real Launch journeys rather than inherited.

## Outcome

**CURRENT STRUCTURE CONFIRMED** for one sequential Step + exact candidate + ACCEPT/RETURN.  
**DEFER SAFELY / RE-DERIVE** unsupported policy dimensions unless operator confirms a named Launch requirement during adjudication.

Strongest counterargument: these controls make governance more professional and flexible.  
Answer: flexibility is not free; each adds state, permissions, race laws and UI. `NoHumanApproval` already gives an explicit path when no human route is required. Launch should implement only governance dimensions with a concrete consumer.

---

# 10. Finding GCR-07 — Required obsolescence is a real counterexample to Submission-only Approval execution

## Evidence

The Product Contract requires a distinct end-to-end journey:

```text
current EFFECTIVE Revision
→ explicit reason
→ governed obsolescence attempt
→ existing Revision remains EFFECTIVE while pending
→ success: Revision becomes OBSOLETE
→ no current EFFECTIVE Revision remains
```

Existing B4 Approval execution is structurally bound to one `RevisionSubmission`. A no-replacement obsolescence attempt has no new Submission candidate to approve.

## Root cause

The prior governance execution model was specialized to “approve a submitted new revision”, while the product now has a second required governance journey.

## Target invariant

> Keep **one sequential governance Step semantic**, but do not create a generic BPM or arbitrary `subject_type/id` workflow merely to accommodate obsolescence.

The later technical design must support two explicit product journeys:

1. Submission governance;
2. governed obsolescence of the current EFFECTIVE Revision.

How policy definition/Step mechanics are reused is intentionally **not** decided in this GCR.

Unknown requiring later product/technical resolution: whether obsolescence may use `NoHumanApproval` or always requires a human route.

## Outcome

**STOP / SPLIT PREREQUISITE** for technical descent until operator accepts this product-level correction. Then **RESTRUCTURE NOW** the governance execution boundary during topology re-derivation.

---

# 11. Finding GCR-08 — Release/effectivity survives; Distribution must not be part of Launch Release atomicity

## Confirmed core

System-owned Release remains the sole normal effectivity authority:

```text
all required gates satisfied
→ one winning Release
→ candidate EFFECTIVE
→ prior EFFECTIVE SUPERSEDED when present
```

No user “publish latest file” operation.

## Conflict

B4 currently makes winning Release atomically resolve Distribution groups and create Distribution obligations. The Product Contract moved Distribution / Read & Acknowledge to Launch+.

That means current B4 makes a non-Launch capability a prerequisite participant in the core effectivity transaction.

## Outcome

**RESTRUCTURE NOW** — Launch Core Release must have no dependency on Distribution existence/configuration/availability.

**DEFER SAFELY** Distribution / Read & Acknowledge to Launch+ with no dormant Launch tables, permissions, jobs or mandatory Release locks.

Later Launch+ design may attach to Release through an explicit seam without taking effectivity authority.

Reference falsification: Veeva treats Read & Understood as an additional feature set around effective documents; it is not evidence that reader acknowledgement must be part of the transaction that establishes document effectivity.

---

# 12. Finding GCR-09 — Periodic Review is Launch+ and must stop shaping Launch Core

## Conflict

B3 gave Periodic Review first-class policy/records and B3/B4 created Release serialization obligations specifically to prevent a periodic-review race.

The Product Contract now places Periodic Review in Launch+.

## Outcome

**DEFER SAFELY.**

Launch Core must contain no PeriodicReview policy/record/permission/job or Release lock obligation. Reopen at Launch+ from the stable EFFECTIVE Revision identity.

Veeva and Qualio both expose periodic review as a distinct feature/process layer, supporting that it can attach later without redefining core Document/Revision/Release identity.

---

# 13. Finding GCR-10 — Dossier, Evidence and Records Governance are Future; their backward pressure on the core must be removed

## Evidence

B5 was internally coherent under its old requirements, but it introduced or forced:

- Dossier/Evidence contexts and permissions;
- Evidence lifecycle/capture;
- retention policy/bindings/extensions;
- LegalHold materialization;
- disposition fences/records;
- a global Artifact retention-root law;
- template-provenance changes to permit later source disposal;
- Revision payload/skeleton separation to permit later physical disposition;
- substantial cross-owner locks/transactions.

The accepted Product Contract makes all of these Future and explicitly says Future capabilities create no dormant Launch module/table/permission/job.

## Outcome

**DEFER SAFELY** the capabilities.  
**RESTRUCTURE NOW** any Launch design that still exists only to support their future disposal/hold/context semantics.

Consequences for later re-derivation:

- no Documentary Context Launch bounded context entitlement;
- no Evidence Launch owner/lifecycle;
- no Records Governance Launch owner;
- no Retention/Hold/Disposition transaction graph;
- no Artifact “one semantic retention root” law;
- no B5-driven persistence split retained solely for future deletion;
- no Evidence/Dossier imported-history branches in Launch migration.

Do not delete useful historical analysis; it remains Future evidence and a future reopen packet.

---

# 14. Finding GCR-11 — `Interchange` is an abstraction grouping unrelated transfer capabilities

## Evidence

R10-A's `Interchange` supporting semantic owner grouped:

```text
Historical Migration       // Launch Core cutover concern
Governed Subject Export    // Future
External IMPORT_COPY       // Future
External PUBLISH_COPY      // Future
```

Backup/Restore is separately an operations continuity concern.

B6 itself correctly states that ongoing imported business truth must become target-owner state so Controlled Information does not depend forever on migration-process tables.

## Root cause

Capabilities were grouped because they all “move data”, not because they share one product lifecycle/authority.

## Outcome

**RESTRUCTURE NOW.**

Historical Migration survives as a bounded **go-live/cutover capability**, but does not automatically justify a permanent generic Interchange semantic owner.

**DEFER SAFELY**:

- Governed Subject Export;
- generic repository connections/transfers/receipts;
- IMPORT_COPY/PUBLISH_COPY runtime product surfaces.

Target imported truth remains essential:

```text
native vs imported history must remain distinguishable
trusted source ordinals/state/provenance must be preserved
unknown must remain unknown
native Submission/Approval/Release/User history must never be fabricated
```

Concepts such as preserving a reliable missing-content ordinal so it cannot later be reused may remain necessary. Exact migration process persistence, idempotency state and ownership are re-derived later from the cutover consumer, not inherited as an `Interchange` domain.

---

# 15. Finding GCR-12 — Audit separation survives; the global cryptographic chain is not yet proved as essential

## Confirmed product authority

```text
Domain history owns lifecycle facts.
Audit owns transversal action/timeline evidence.
Audit is never queried to derive current lifecycle state.
```

Critical governed actions must not report success while their required Audit evidence failed to persist. Same-local-commit append-only Audit is therefore independently justified.

## Bounded challenge

B6 additionally made a deployment-wide cryptographic `AuditChainHead` the final lock of every audited transaction and shaped the global lock DAG around it.

The Product Contract requires trustworthy Audit; it does not currently name a threat/compliance requirement for a bespoke in-database cryptographic chain or external non-repudiation.

An in-database hash chain can detect some accidental/unauthorized mutation, but without a separately protected external anchor it is not equivalent to proof against a fully privileged database operator. Therefore the assurance gain must be weighed against making every governed transaction contend on one global semantic lock.

## Outcome

**CURRENT STRUCTURE CONFIRMED**:

- separate Audit authority;
- append-only immutable events;
- bounded, PII-minimized facts;
- same-commit Audit for required governed operations;
- Audit never becomes event sourcing/domain state.

**DEFER SAFELY unless the operator names the Launch assurance requirement**:

- global cryptographic hash-chain head as a product requirement;
- transaction-wide lock ordering driven by that chain;
- stronger external anchoring/non-repudiation machinery.

This is a bounded adjudication item, not permission to weaken ordinary Audit immutability.

---

# 16. Finding GCR-13 — Search is correctly a projection, but current-effective truth needs an explicit fail-closed journey invariant

## Confirmed boundary

Search is rebuildable discovery and never grants access. That survives.

## Missing end-to-end proof target

The Product Contract requires ordinary readers to find **current EFFECTIVE** truth, while draft/submitted work belongs in author/governance workspaces.

An eventually consistent index may temporarily be stale. Therefore later technical design must prove:

```text
search hit != authority to serve bytes
search hit != authority that Revision is still EFFECTIVE
open/read/download re-resolves canonical Document + current EFFECTIVE Revision + AuthZ
stale hit cannot serve SUPERSEDED/OBSOLETE content as current
history is an explicit authorized journey, not ordinary-reader fallback
```

Obsolescence success must make ordinary current search cease presenting the Document as active; projection delay may never convert stale metadata into canonical content truth.

## Outcome

**CURRENT STRUCTURE CONFIRMED** for Search-as-projection.  
**STOP / SPLIT PREREQUISITE** for later API/frontend design until the above canonical read invariant is explicit in the re-derived technical architecture.

---

# 17. Finding GCR-14 — Current AuthZ architecture mostly survives; current role/permission catalog does not

## Surviving boundary

The accepted Product Contract explicitly preserves:

```text
Authentication provider may own credentials
MetalDocs owns product User identity + Authorization
Users / Areas / Groups + scoped access
search/context links never grant access
```

B1/B2's one-company posture, provider-role anti-corruption boundary and canonical product AuthZ remain coherent.

## Permission leakage

The promoted catalog still contains Launch+/Future capabilities including:

```text
document.review_periodic
distribution.manage
distribution.oversee
Evidence / Dossier permissions
retention.extend
legal_hold.manage
disposition.manage
governed_subject.export
external_repository.publish
```

It also carries Approval administration dimensions that the Product Contract does not yet establish (`oversee`, `reassign`, administrative approval cancel).

## Missing persona path

The accepted Product Contract has both:

- ordinary **Reader**;
- **Auditor / Governance Viewer**, who must reconstruct lifecycle/action history without becoming an administrator.

The current exact five-role bundles do not provide a clean least-privilege Auditor/Governance Viewer path: `viewer` is effective-read oriented while `audit.read` is currently held only by the all-powerful tenant-owner bundle.

## Outcome

**RESTRUCTURE NOW** the role/permission catalog **after** product adjudication. Do not patch the current 5×43 catalog by subtraction while product decisions are still open.

The later catalog must be regenerated from accepted Launch journeys and include a least-privilege way to satisfy the Auditor/Governance Viewer persona without conflating it with tenant administration.

Exact role count/bundles are intentionally not decided in this GCR.

---

# 18. Finding GCR-15 — Several B3/B4 adjuncts have no accepted Launch consumer

The following are not Product Contract capabilities merely because they already exist in prior design:

```text
DocumentTypeCategory navigation taxonomy
Tenant Dictionary / System Value Catalog as editable product capability
structured TemplateSpec platform
DRAFT EditorialComment collaboration system
mandatory REV002+ reason-for-change rule
scheduled ReleasePlan.not_before
auxiliary semantic Rendition for SourceOnly documents
structured anchored suggestion payloads
Approval due-date/SLA machinery
fresh-auth as universal configurable Step dimension
```

Some may later prove useful. The Product Contract explicitly names reasons where they are essential (withdraw/cancel/obsolescence) but does not authorize a generic “every possible governance metadata” platform.

## Outcome

**DEFER SAFELY unless a named Launch journey/consumer is produced during operator adjudication.**

Two refinements:

1. numbering itself is core; only unrelated taxonomy/configuration machinery is challenged;
2. a preview generated for UI fidelity may exist as a technical cache/mechanism without becoming a governed semantic Rendition when the Document Type is SourceOnly. A semantic Rendition is justified when the product treats it as the official required representation.

---

# 19. Finding GCR-16 — Physical integrity/restore remains mandatory, but must be re-derived from semantic content owners

The Product Contract keeps backup/restore correctness and exact-content history. R10-C contains useful safety conclusions independent of `Artifact`:

```text
provider success never equals semantic success
provider keys/URLs are not business identity
admitted writes must not overwrite governed bytes
server verifies exact bytes/hash/size/format
production untrusted content cannot silently bypass required safety inspection
temporary failed/stale DRAFT data may be reclaimed
governed confirmed history is not physically disposed in Launch
restore cannot serve while required governed bytes are missing/corrupt
```

## Outcome

**CURRENT STRUCTURE CONFIRMED** for these technical obligations.  
**RESTRUCTURE NOW** the later storage/restore architecture so proofs start from the semantic records that own exact content rather than from a global Artifact registry.

No custom backup platform, WORM/ObjectLock, multi-cloud/BYOS or governed delete workflow is justified by Launch.

---

# 20. Whole-product essential vs accidental complexity

## Essential Launch complexity

```text
single-company identity/access
stable Document identity
business Revision non-reuse
mutable DRAFT Working Content + OCC/recoverability
immutable exact Submission attempts
sequential governance over exact attempt
feedback + ACCEPT / RETURN
withdraw attempt vs cancel Revision distinction
system-owned Release/effectivity
one current EFFECTIVE truth
optional required official Rendition
governed no-replacement obsolescence
current-effective reader discovery/read
history + Audit with distinct authorities
truthful imported/native history distinction when migrating
exact content integrity + restore fail-closed
```

## Accidental / not Launch-entitled

```text
standalone Artifact semantic owner
Artifact retention-root graph
ArtifactStage as semantic domain lifecycle
Distribution inside Release atomicity
Periodic Review inside core CI/Release locks
Dossier/Evidence/Records Governance Launch contexts
retention/hold/disposition-driven core persistence splits
generic Interchange owner
Governed Export/repository copy runtime machinery
workflow dimensions without named Launch consumers
scheduled effectivity without requirement
SourceOnly auxiliary preview as governed Rendition
generic dictionary/taxonomy/structured-template subplatforms without consumer
global Audit hash-chain lock without named assurance requirement
```

---

# 21. Smallest sustainable Launch product — capability conclusion only

This is deliberately **not** a bounded-context/package topology.

The smallest sustainable Launch must be able to complete this product loop truthfully:

```text
configure company identity/access + controlled-document policy
→ create stable Document / REV001 DRAFT
→ author/autosave recoverably
→ optionally seed from exact current EFFECTIVE Template
→ freeze immutable Submission
→ apply no-human or sequential human governance
→ return/change/resubmit without rewriting history
→ withdraw an attempt or cancel a change cycle correctly
→ system-release exact accepted content
→ keep prior effective truth while a successor is open
→ replace EFFECTIVE atomically on successful successor Release
→ obsolete current EFFECTIVE through explicit governance with no successor
→ let readers find/read/download canonical current EFFECTIVE truth
→ let authorized governance/audit viewers reconstruct history
→ migrate historical truth without fabricating native history
→ restore without serving inconsistent semantic/content state
```

If a proposed semantic owner/entity/capability does not serve this loop or an explicit product invariant, it is presumptively YAGNI for Launch.

---

# 22. Proposed operator adjudication

The GCR recommends the operator accept the following as one coherent disposition:

### A1 — Accept

**Remove standalone `Artifact` semantic ownership from the Launch target.** Exact-content facts live with the semantic record that freezes them; storage/staging/integrity remain mechanisms.

### A2 — Accept

**Keep Document ≠ Revision ≠ Working Content ≠ Submission** and keep system-owned Release as effectivity authority.

### A3 — Accept

**Keep one sequential governance Step semantic**, but do not inherit unproven B4 dimensions by default. Re-derive participant/quorum/SoD/fresh-auth/overseer/reassignment details only from named Launch journeys.

### A4 — Accept

**Treat governed obsolescence as a second explicit governance journey**, while refusing a generic arbitrary-subject BPM/workflow engine.

### A5 — Accept

**Remove Distribution and Periodic Review from Launch Core architecture** and remove their mandatory Release/core transaction coupling. They remain Launch+.

### A6 — Accept

**Defer Dossier, Evidence and Records Governance completely for Launch**, including all backward pressure they created on Artifact ownership, retention-root laws and disposition-driven persistence splits.

### A7 — Accept

**Break the generic Interchange grouping.** Keep truthful Historical Migration as a cutover capability; Governed Export and generic repository IMPORT/PUBLISH remain Future.

### A8 — Accept

**Regenerate Launch AuthZ roles/permissions from the Product Contract after this adjudication**, including a least-privilege Auditor/Governance Viewer path. Do not preserve the current exact 5×43 catalog by sunk cost.

### A9 — Accept unless a named Launch assurance requirement is supplied

Keep same-commit append-only Audit, but **defer the global cryptographic AuditChainHead/lock law** until a concrete tamper-evidence/non-repudiation requirement justifies it.

### A10 — Accept unless a named Launch consumer is supplied

Defer residual uncontracted adjuncts such as editable Dictionary/System Values, DocumentTypeCategory, structured TemplateSpec, DRAFT comment platform, scheduled release, auxiliary semantic SourceOnly rendition and advanced Approval policy dimensions.

---

# 23. What this GCR explicitly does not decide

Until the operator adjudicates A1–A10, do not decide:

```text
new bounded-context count/names
module/package layout
DB tables/schema/constraints
storage provider port shape
staging model
Submission manifest physical representation
new Approval/Obsolescence persistence shape
exact role count/permission catalog
async/outbox/job topology
API routes/frontend pages
migration implementation tooling
implementation plan
```

Those are downstream of accepted product/authority conclusions.

---

# 24. Gate / next exact step

```text
this Whole-Product GCR
→ OPERATOR ADJUDICATION
→ re-derive ownership/topology from accepted findings
→ re-derive remaining technical architecture
→ Whole-R10 Global Coherence Review
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

Implementation remains **BLOCKED**.
