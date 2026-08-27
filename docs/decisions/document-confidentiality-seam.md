---
id: document-confidentiality-seam
kind: authority
owner: architecture
summary: Bounded future seam for in-Area document confidentiality through a governed classification plus clearance grants, fully decided now with zero Launch capability delta.
---

# Document Confidentiality future seam

> **Status:** OPERATOR-RATIFIED / SEMANTIC MODEL CURRENT — FUTURE-SEAM DISPOSITION UNDER OPEN PROMOTION REOPEN (`document-confidentiality-launch.md`)
> **Ratified:** 2026-08-26
> **Origin:** T11 FP2-F3 — operator finding after P11 ("documento da Produção legível só por gerente e diretoria")
> **Method:** Engineering Method v1.0.0 + Frontend Method v2.3 §3.10A / §26 (category 7)
> **Implementation:** BLOCKED.
> **Current Launch capability delta:** ZERO.
> **Precedent shape:** `governance-review-layer-seam.md`.

## 1. Decision outcome

```text
CURRENT LAUNCH ACCESS STRUCTURE CONFIRMED
+ BOUNDED FUTURE CONFIDENTIALITY SEAM DECIDED AND PRESERVED
+ ZERO dormant module, table, permission, job or UI in Launch
```

The need is **real and material**, proven by an ordinary company situation and by three mature reference platforms. It is **not** Launch V1 scope. This page decides the future shape now so Launch grows into it instead of being retrofitted against it.

## 2. Proven human job

```text
a Document belongs organizationally to an Area shared by many readers,
while its content must be legible only to a subset (management, direction)
without moving it out of its organizational context and without losing
identity, numbering, revision history or audit continuity
```

## 3. Insufficiency of current authority (proved, not assumed)

```text
T3 read equation          document.read_effective in scope Company | Area
viewer bundle             { document.read_effective }
RoleAssignmentScopeKind   company | area — no further dimension exists
document predicates       constrain state/relationship/governance, never sensitivity
```

Two Documents in the same Area are therefore **indistinguishable** for read authorization. Current authority cannot express §2.

## 4. Reference convergence (Evidence, never authority)

```text
Veeva Vault      Dynamic Access Control — sharing rules driven by document METADATA assign
                 groups automatically; manual per-document assignment exists as exception
Qualio           private tags bound to User Groups; tag→group mapping administered centrally;
                 unrestricted by default
M-Files          automatic permissions driven by metadata property/class/value; named ACLs
                 remain a secondary option

CONVERGENT PRODUCT DECISION
  confidentiality is a GOVERNED CLASSIFICATION on the document, mapped centrally to groups
  it is NOT a free per-document people picker
```

## 5. Decided future semantic model

### 5.1 ConfidentialityClass

```text
Company-configured vocabulary of classes, exactly like Area and DocumentType are configured data
authorization semantics of a class are PRODUCT-OWNED and static, exactly like Role semantics
every Document carries exactly one class
one distinguished default class means "no restriction beyond Area/Company grants"
Launch behaves as if every Document carries that default class, with no vocabulary materialized
```

### 5.2 Clearance grant — a second independent axis

```text
ConfidentialityGrant { subject: User | Group, class: ConfidentialityClass, scope: Company | Area }
```

The T3 equation gains **one conjunctive term**, never a restructure:

```text
enabled User
+ (direct RoleAssignments ∪ Group RoleAssignments)
+ static Role → Permission bundle
+ scope match
+ Controlled Documents predicates
+ CLEARANCE: document class = default  OR  actor holds a current grant for that class in scope
= ALLOW
```

Default DENY is preserved. Evaluation stays live and compositional. Administration stays in `access.manage`, so **no new permission family is introduced**.

### 5.3 Non-identity law — the seam that makes this implementable

```text
ConfidentialityClass NEVER enters Document.code
ConfidentialityClass NEVER becomes a numbering scope or identity component
reclassification is a governed, audited operation that changes NEITHER identity, code,
  revision ordinals, effectivity NOR history
```

This is the decisive law. Confidentiality must be **mutable and non-identity** precisely because organizational placement (Area) is neither.

### 5.4 Explicitly excluded, permanently

```text
per-document ACL / people picker      contradicts T3 "materialized ACL" exclusion, destroys the
                                      auditable "who could see what, when", produces permission drift
external / guest / public-link access  contract §1 + §8: single-company deployment, authenticated users
confidentiality as an Area            falsified in §7
frontend-computed visibility           frontend never evaluates authorization (T3 + frontend.md)
```

## 6. Seams Launch V1 must preserve (the actual obligation)

Launch builds nothing for this. Launch must, however, avoid foreclosing it. Each item below already follows from current authority; listing them makes them **seam-critical** rather than incidental:

```text
S1  authorization is never materialized, denormalized or cached as truth
    (T3 negative list) — a future clearance term must be evaluable live
S2  read authorization is conceptually PER DOCUMENT, even where today's outcome is uniform
    across an Area; no read model may encode "grant in Area ⇒ every document of the Area"
    as a semantic shortcut
S3  no NEW access dimension may enter Document identity or code; Area's presence in a
    numbering scope is numbering provenance, never a statement about who may read
S4  Audit facts stay typed and additive, so a future classification-changed fact composes
    without rewriting historical evidence
S5  read projections never let the client infer visibility from field absence; disclosure
    stays server-decided (existing presence/disclosure law)
S6  Area is never presented — in product, documentation or guidance — as a confidentiality
    mechanism
```

## 7. Why Area must never be the workaround (falsified with evidence)

```text
IRREVERSIBLE      `area_id` is supplied only at op46 createDocument; NO operation changes a
                  Document's Area. Reclassification would require obsoleting and recreating,
                  breaking the identity/history continuity the product exists to guarantee
IDENTITY LIES     with numbering_scope=document_type_area the code embeds the Area, and
                  `committed Document.code = unique and never reused` makes it permanent —
                  the document's own code would carry its secrecy level forever
ORGANIZATION      Area is an ORGANIZATIONAL concept; duplicating Areas by sensitivity fragments
POLLUTED          the org chart, the numbering space and navigation to solve a security problem
```

Recommending it would create irreversible damage in a customer's estate. Hence S6.

## 8. Current Launch consequence

```text
capability delta            ZERO
new module/table/permission NONE
new operation               NONE — census remains 89
UI surface                  NONE; the creation surface exposes no confidentiality control
```

Honest V1 limitation, stated rather than disguised: a company needing in-Area confidentiality either keeps such documents outside MetalDocs until this capability ships, or accepts Area-level visibility. It must **not** be told to create a restricted Area (§7).

The B13 creation surface must, within current authority, state the read consequence of the Area choice as a **rule** ("readers holding permission in this Area or Company-wide will be able to read it once effective") and never as an enumeration of people.

## 9. Proof obligations for any future promotion

```text
1. clearance is a conjunctive term: removing a clearance immediately removes read, live
2. a reclassified Document keeps code, identity, revision ordinals, effectivity and history
3. collection reads and Search never leak restricted existence through counts, cursors or gaps
4. Audit answers "who could see this document at instant T" from typed facts alone
5. no materialized ACL, permission cache or frontend-evaluated visibility appears
6. Group membership drift changes access live, without rewriting historical audit meaning
```

## 10. Reopen / promotion triggers

```text
a named customer or regulatory consumer requires in-Area confidentiality
a Launch+ capability (Distribution / Periodic Review) acquires a restricted-audience requirement
a real estate of sensitive documents is blocked from onboarding by this limitation
```

## 11. Trigger status

```text
2026-08-26  TRIGGER §10.1 FIRED BY OPERATOR
            named consumer: Comercial Area — documents legible only to vendedores,
            others only to gerência, others only to diretoria, all inside one Area
            → the need is ORDINARY, not exceptional; the "keep it outside MetalDocs"
              fallback of §8 is not viable for this estate
```

Firing the trigger does **not** by itself change scope. Promotion from future seam to
Launch scope is a Product/T6 bounded reopen and requires explicit operator ratification.
It is registered as `SEC-07` in `forward-obligations.md`.

```text
2026-08-27  PROMOTION REOPEN OPENED / OPERATOR-AUTHORIZED
            → docs/decisions/document-confidentiality-launch.md
            The operator authorized opening the Product/T6 bounded reopen BEFORE B13
            LOCK, so the creation surface absorbs the capability instead of being
            retrofitted. This page's semantic model, exclusions and proof obligations
            are carried forward unchanged; only the FUTURE disposition is superseded.
```

Until that reopen is ratified, Launch scope stays at 89 operations and no confidentiality
capability is implemented.
