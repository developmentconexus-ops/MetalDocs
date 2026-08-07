# ADR 0093 — Controlled Information is one bounded context; a template is a version-scoped role, not a peer aggregate

> **Status:** Accepted as a design ruling 2026-08-07. Not implemented — no code or schema change is authorized by this ADR alone.
> **Scope:** Aggregate and bounded-context boundaries for document authoring, document control, and templates. Not the grant model (see [`0092`](0092-authz-grant-unification.md), pending), not the approval kernel's internals (see [`0082`](0082-approval-kernel-extraction.md)).
> **Supersedes:** the "no backend module merge" ruling in `docs/superpowers/specs/2026-06-30-template-document-parity-design.md` §D1.

## Context

Three modules own what this ADR treats as one domain: `documents` (27,024 Go LOC), `controlleddocuments`
(11,131), `templates` (17,339). Four of the seven module-level import cycles in the backend run between them
and `approval`. `templates/domain/errors.go` documents itself as mirroring `controlleddocuments`' identically
named sentinels; seven sentinel names, twin status-transition validators, and a byte-identical objectstore
error switch are defined twice across `documents` and `templates`.

The operator asked whether the three-way split is a domain truth or an artifact of implementation history.
**The first two independent answers were rejected, and the reason is the durable part of this ADR.** Both
advisory arms answered with evidence about the system as built: table ownership is disjoint, no `POST
/documents` route exists, a prior design already ratified "no merge", the current transaction creates slot and
first document together. Every one of those is a *consequence* of the split being examined. An argument of
that shape can only ever conclude "leave it as it is", because it takes the artifact under examination as its
own premise. This is the Local-Maximum failure named in `CLAUDE.md`, appearing in the one place hardest to
catch — inside an analysis commissioned specifically to avoid it.

The question was re-asked with status-quo evidence ruled **inadmissible**: current schema, module layout,
import graph, route topology, existing ADRs, transaction boundaries, doc-comments describing the code, and
migration cost were all barred. Admissible: the regulatory domain (ISO 9001:2015 §7.5, ISO 13485:2016 §4.2,
21 CFR Part 11, EU GMP Chapter 4 and Annex 11), how mature field systems model it (Veeva Vault, M-Files,
MasterControl, Qualio), aggregate-design theory, and the product's observable user-facing behaviour. Two
independent arms answered the reframed question and converged.

## Decision

### One bounded context: Controlled Information

**`ControlledDocument`** — the stable governed lineage. Identity `(tenant, documentId)` with an immutably
bound business code. Owns: governance class, owning area, active/withdrawn state, revision-slot allocation,
the review schedule, and **the sole effective-revision pointer**. Revision labels are never reused.

**`DocumentRevision`** — the content candidate and, once released, the controlled issuance. Owns: the mutable
draft head, immutable submission attempts, exact content digest, template provenance, and the approval
receipt. Rejection preserves the failed attempt.

These are **two aggregates in one bounded context**, not one aggregate and not two contexts. The one
transactional invariant that ISO 9001 §7.5.3 makes non-negotiable — never two simultaneously effective
revisions — is satisfied by making exactly one step atomic:

1. the revision freezes its submission digest;
2. approval completes against that digest;
3. the revision accepts the matching receipt;
4. `ControlledDocument` switches the effective pointer — **atomic**.

The previous revision remains effective until step 4, so the deliberately eventual steps 1–3 never expose an
unapproved candidate. Collapsing both into one aggregate would place identity changes, editing, approval and
distribution inside one contention boundary for no invariant that requires it.

**`AuthoringWorkspace`** is the mutable draft surface, joined to governance only by hash-freeze at each
transition. Keystroke-rate writes must not contend on the row guarding effectivity, and no governance
invariant depends on an intermediate draft state.

**`NumberSeries`** guards scope-wide uniqueness for `(tenant, numberingScope)`. Uniqueness spans every
document in the scope, so it cannot live inside any one of them. Issuance — allocate and bind — is the one
sanctioned cross-aggregate transaction: an issued code with no document, or a code issued twice, are both
audit findings.

### A template is a version-scoped role, not a type

A QMS blank form or authoring skeleton **is a controlled document**. ISO 13485 §4.2.4 governs documents and
§4.2.5 governs records; the blank form sits on the document side precisely so every record traces to the form
revision it was captured on. EU GMP Chapter 4 requires controls over electronic templates, forms and master
documents; EMA GMP Q&A requires a blank-form master to be approved and to carry a unique reference including
version.

`TemplateUsePolicy` is the aggregate that carries the role: it designates which **released revision** may seed
which document classes within a scope, validates the placeholder/schema contract, and binds consumers to an
**exact revision and hash**. Field practice matches: Veeva designates an approved document as a controlled-
document template and records the exact template document number *and version* on every derived document, and
a later approved version can take over the designation; M-Files carries template status as a property on an
otherwise ordinary versioned object.

Template is therefore a **relationship**, not an essence. A permanent subtype would make a document a template
forever and make re-designation to a newer approved version awkward — the operation Veeva treats as routine. A
bare boolean or an uncontrolled projection would permit silent rebinding and a floating "latest template"
reference, which breaks record-to-form-revision traceability.

### One draft-integrity policy, for every governance class

21 CFR Part 11 §11.10(e) requires time-stamped audit trails that do not obscure prior information; Annex 11 §9
requires GMP-relevant changes and their reasons to be auditable; §11.70 / Annex 11 §14 make signed content
permanently linked and immutable. Neither requires a physically append-only store for drafts. So:

- draft head — mutable, concurrency-controlled;
- saved content commits and audit evidence — append-only or equivalently reconstructible;
- submitted snapshot — immutable and locked; any post-submission change withdraws the attempt and creates a
  new immutable submission;
- approved/effective revision — immutable; change requires a new revision.

**Uniformly, for templates and every other governance class.** Draft mutability is not the defect.
Destructively overwriting one class's draft history while an equivalent class retains reconstructible history
is: it is two answers to one regulatory question, and by the QMS's own logic the entities are not different.

### The approval kernel stays separate and subject-generic

A shared approval process is **not** evidence that its subjects are one aggregate type — distinct aggregates
routinely share approval, payment, or case-management protocols, and Part 11 §11.50/§11.70 constrain the
signature record, not the domain type of what is signed. `ApprovalCase` remains its own aggregate with its own
immutability law, binding an immutable `(subject-kind, subject-id, submission-id, digest)`. Its subject
genericity ([`0082`](0082-approval-kernel-extraction.md)) is correct and is retained; under this model it
serves mostly one subject kind, and earns its genericity on genuinely foreign future subjects (CAPA, change
control, supplier records).

Route *definitions* are versioned policy; each case snapshots the route version at submit, so later edits
cannot rewrite an in-flight or historical approval. That materialization is an evidentiary requirement, not a
cache.

### Second-order placements

Taxonomy is a separate classification context; a revision snapshots audit-relevant labels so reorganizing an
area never rewrites history. Governance class is versioned policy data — numbering scheme, required controls,
review cadence, permitted template policies — and is what lets one model serve every document kind without
code forks. Distribution and acknowledgement are per **released revision** and recipient, never per lineage,
and are eventually consistent. Periodic review references the stable document and its currently effective
revision. Obsolescence is a lifecycle state of `ControlledDocument`, not a separate concept. A rendered
DOCX/PDF is an immutable derived `Rendition` keyed by revision, submission manifest, renderer version and
hash — a projection with no independent invariants; evidentiary truth remains the signed revision manifest,
and approval must never wait inside a rendering transaction.

## Inversion trigger

Both arms independently identified the same condition under which this ruling is wrong, by the same reasoning.
It is recorded here as a standing trigger, not as a caveat:

> **If the template content payload develops an independent lifecycle — an executable placeholder schema,
> validation logic, renderer-compatibility rules, schema evolution, or migration semantics for in-flight
> documents bound to a superseded template revision — then `TemplateDefinition` earns its own aggregate.**

Folding templates in makes the core aggregate payload-polymorphic, and payload polymorphism is where unified
models rot: class-conditional branches get pushed into shared governance code, and the instantiation edge
becomes a self-referential link between instances of one aggregate type, which is harder to reason about,
test, and access-control than a clean edge between two named aggregates.

The trigger fires when **template-specific behaviour grows faster than governance-shared behaviour**. When it
fires, the correct move is `TemplateDefinition` **paired with** — never replacing — the controlled-document
revision that supplies its regulated approval and provenance. This ADR is superseded at that point, not
patched.

Whoever next touches template payload behaviour is responsible for evaluating this trigger and recording the
evaluation. That obligation has no firing mechanism yet; until one exists this is a level-5 (discipline)
control and should be treated with the suspicion `docs/engineering/mechanical-enforcement-register.md` demands
of that level.

## Consequences

- The three-module split does not survive. The target is one Controlled Information context with the
  aggregates above, plus `NumberSeries`, `TemplateUsePolicy`, and the separate Approval & Evidence context.
- Four of the seven module-level import cycles are dissolved by the boundary change rather than repaired.
- The duplicated lifecycle machinery between `documents` and `templates` becomes one implementation. Two
  implementations of one regulation, maintained separately, drift where auditors look.
- Approval's ownership of document status writes and `documents`' reads of approval tables are re-expressed as
  intra-context relationships or explicit ports, not cross-module reach.
- The draft-integrity asymmetry becomes a defect with a named fix, not a modelling difference.
- Nothing here authorizes implementation. Sequencing against the other architecture axes is decided
  separately; both advisory arms, in both rounds, held that the grant-model axis ([`0092`](0092-authz-grant-unification.md))
  must not be shelved behind this one.

## Rejected alternatives

**Keep the three modules; extract a shared lifecycle kernel.** Rejected: it preserves two governance
implementations and adds a third construct to keep them in step. The duplication is the symptom; the boundary
is the cause.

**One aggregate owning identity and revisions.** Rejected: it forces identity changes, editing, approval and
distribution into one contention boundary to protect an invariant that only the effective-pointer switch
actually requires.

**Template as a permanent subtype or governance class.** Rejected: makes template an exclusive essence rather
than a use relationship, and makes re-designation to a newer approved revision awkward.

**Merge templates into documents with no policy aggregate.** Rejected: permits floating "latest template"
references and silent rebinding, breaking record-to-form-revision traceability.

## Evidence basis

Two independent advisory arms (Fable 5; GPT-5.6 Sol) answered the same brief with status-quo evidence ruled
inadmissible, and converged on: templates are not a peer content aggregate; the draft-mutability asymmetry is
a defect; the approval kernel stays separate and subject-generic; renditions are derived projections;
distribution binds to a released revision. They diverged on one-vs-two aggregates and on class-vs-role for
templates; this ADR takes the two-aggregate and role positions, reasoned above. Both arms explicitly passed an
inversion test — stating that their conclusions would not move if the current implementation were the opposite
in every respect.

Brief: `docs/superpowers/analysis/2026-08-07-controlled-information-greenfield-brief.md`.
