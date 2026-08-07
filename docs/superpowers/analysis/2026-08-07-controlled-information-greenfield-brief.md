# Brief — greenfield aggregate design for a controlled-document QMS. Status quo is INADMISSIBLE evidence.

## Why this brief exists

A previous round asked "is the documents / controlleddocuments / templates 3-way module split a domain truth or
an implementation accident?" Two independent advisors answered — and BOTH answered largely with evidence about
the code that exists: table ownership, current route topology, existing ADRs, the shape of the current
transaction. The operator rejected both answers, correctly, on this ground:

> "você está baseando num máximo local — baseando numa coisa que já está implementada em vez de analisar o
> melhor pra ser implementado. Tentar fazer o melhor porque já está implementado, e o que está implementado
> for ruim, não faz sentido. A gente tem que analisar o global."

He is right. "The tables are disjoint" is the *consequence* of the split, not an argument for it. "There is no
POST /documents" is current route topology. "ADR 0011 / 0082 already decided this" is appeal to precedent. All
of it is circular.

## THE RULE — read this twice

**Evidence about the current MetalDocs implementation is INADMISSIBLE as an argument for or against any model.**

Specifically inadmissible:
- the current database schema, table ownership, FKs, or constraints
- the current module layout, package structure, or import graph
- the current HTTP route topology or OpenAPI spec
- any existing ADR, design spec, or ratified decision in this repo
- the current transaction boundaries
- doc-comments in the code describing what the code is (the code describing itself is not domain evidence)
- migration cost, refactor size, or "how much work it would be"

You MAY read the code for exactly one purpose: to learn **what the product actually does for its users** —
the observable behaviours, the workflows, the rules the business needs. Extract requirements, not structure.
State plainly when you are doing this.

Admissible evidence:
- the regulatory/domain reality of controlled-document management: ISO 9001:2015 §7.5 (documented information),
  ISO 13485 §4.2, 21 CFR Part 11 (electronic records/signatures), EU GMP Annex 11, GxP document control
- how mature systems in this exact field model it (Veeva Vault QualityDocs, MasterControl, Qualio, Greenlight
  Guru, SharePoint/M-Files document control, DocuWare) — cite what you actually know, mark uncertainty
- general aggregate-design theory (DDD aggregate boundaries: what must be transactionally consistent, what is
  eventually consistent, what has independent lifecycle and independent identity)
- the product's actual user-facing behaviour, extracted as requirements

## The question

**Designing from zero, for a multi-tenant controlled-document QMS: what are the correct aggregates and bounded
contexts covering document authoring, document control (identity/numbering/revision governance), and templates?**

Answer these:

**G1 — The aggregates.** Name them. For each: its identity, its invariants, what MUST be transactionally
consistent inside it, and what is merely referenced. Justify each boundary from the domain, not from code.

**G2 — Controlled identity vs content.** Is "the controlled document" (a stable number/code, a governance class,
an owning area, a revision history) the same aggregate as "the content that is being revised", or two?
The regulatory framing matters here: in ISO/GxP document control, what is the *controlled entity* — the number
or the revision? Does that force one model? Answer from the norm and from how field systems do it.

**G3 — Templates.** Is a template (a) a kind of controlled document whose content happens to be a skeleton,
(b) a separate aggregate with its own identity and lifecycle, or (c) a projection/role that any controlled
document can play? Consider: templates in a QMS are themselves usually controlled documents (a form template
is approved, versioned, and its revision is traceable on every record produced from it). Does that collapse
the distinction? What breaks if you model it either way?

**G4 — The revision-mutability question, asked honestly.** In the current system, template versions are
draft-mutable and overwritten in place; document revisions are append-only and content-hash addressed. DO NOT
treat this as a requirement. Ask: in a regulated QMS, *should* a draft template version be mutable in place?
What does Part 11 / audit-trail integrity actually demand of a draft? If both should be append-only, that
difference is a defect, not a domain distinction — say so.

**G5 — Approval as evidence.** In this product, approval was independently modelled as subject-generic: one
approval kernel serving both documents and templates, discriminated by subject kind. Treating that ONLY as an
observation about what the process does (not as an endorsement of the current code): when two entity types
require the identical approval/signoff/lifecycle process, is that evidence they are one aggregate type with a
discriminator, or is it normal for distinct aggregates to share a process kernel? Argue it as a modelling
question. This is the strongest non-status-quo argument available in either direction — treat it seriously.

**G6 — The verdict, and the cost of being wrong.** Give the greenfield model. Then, and only then, note how
far the current implementation is from it, and which direction the gap points. Say explicitly which parts of
your greenfield answer would survive if you learned the current implementation were the opposite of what it is
— that is the test of whether you actually reasoned from the domain.

**G7 — Second-order.** Where do the following live in your model, and does that change any boundary above:
areas/taxonomy, governance class, approval routes, distribution/acknowledgement, review-and-expiry
(periodic review), obsolescence/withdrawal, and the rendered artifact (PDF/DOCX)?

## Output discipline

Terse, dense. Argue; do not enumerate options neutrally — commit to a model and defend it. When you cite a
regulation or a commercial system, say what you actually know and flag what you are unsure of; do not invent
citations. If your honest answer is "the current split is right", say so — but the argument must survive with
every status-quo fact deleted. End with: the model, in one paragraph, plus the single strongest argument
AGAINST your own model.
