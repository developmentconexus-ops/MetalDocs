---
id: document-confidentiality-launch
kind: authority
owner: architecture
summary: Bounded Product/T6 reopen promoting in-Area document confidentiality from future seam to Launch V1 scope through a governed non-hierarchical ConfidentialityClass plus clearance grants.
---

# Document Confidentiality — Launch bounded reopen

> **Status:** OPEN / OPERATOR-AUTHORIZED 2026-08-27 / NOT YET RATIFIED
> **Supersedes on ratification:** `document-confidentiality-seam.md` (future-seam disposition only; its semantic model, exclusions and proof obligations are carried forward unchanged)
> **Origin:** operator finding after FP2/P11; trigger §10.1 of the seam decision fired
> **Method:** Engineering Method v1.0.0 + Frontend Method v2.3 §26 category 7 (missing Product capability) + §5.3 bounded rebaseline
> **Implementation:** BLOCKED.

## 1. Why the disposition changes

The seam decision deferred this because no named consumer required it and the honest
fallback ("keep such documents outside MetalDocs") was viable. The operator has falsified
both premises:

```text
NAMED CONSUMER    a Comercial Area holding documents legible only to vendedores,
                  others only to gerência, others only to diretoria — one Area,
                  three audiences, simultaneously and permanently
ORDINARY, NOT     the same shape recurs in Produção, RH and Diretoria. A product whose
EXCEPTIONAL       Areas cannot hold differently-legible documents cannot hold a real
                  company's document estate
FALLBACK DEAD     "keep it outside MetalDocs" excludes a routine fraction of every Area,
                  which defeats the North Star of a single controlled estate
```

Deferring further would not save Launch work; it would guarantee a retrofit against
identity, numbering, Audit and every read projection at the exact moment those are
hardest to change.

## 2. What is promoted unchanged

The whole semantic model of `document-confidentiality-seam.md` is carried forward as-is:
§5.1 ConfidentialityClass, §5.2 clearance grant as **one conjunctive term**, §5.3
non-identity law, §5.4 permanent exclusions, §9 proof obligations. Nothing about the
model is reopened. Only its **disposition** (future → Launch) changes.

### 2.1 Non-hierarchical law (decided)

```text
ConfidentialityClasses are ADDITIVE and NON-HIERARCHICAL
no class implies, dominates or inherits another
no total or partial order over classes exists in product semantics
```

This mirrors the existing Launch Role law (`authorization-and-audit.md` §3, "Roles are
additive and non-hierarchical"). Reference convergence agrees: Veeva, Qualio and M-Files
all drive access from unordered metadata mappings, not ordered levels.

Rejected alternative — ordered levels (público < vendedores < gerência < diretoria):

```text
NO GLOBAL ORDER   levels in Comercial and levels in RH are not comparable; a single
                  total order would make "diretoria" in one Area dominate an unrelated
                  Area's classes by arithmetic accident
NOT EXTENSIBLE    inserting a class into a total order retroactively changes the meaning
                  of every grant already issued
UNAUDITABLE       proof obligation §9.4 ("who could see this document at instant T from
                  typed facts alone") degrades into re-deriving a historical order
RECOVERABLE       ordered convenience is expressible ON TOP of independent classes as an
                  administration affordance (grant a Group several classes at once).
                  The reverse is not. Therefore the general model is the correct Launch model.
```

Administration ergonomics — issuing several clearances to one Group in a single
administrative act — are a **UX obligation on B11**, never a semantic hierarchy.

## 3. Launch capability delta (proposed, to be proved at P6/P7)

### 3.1 Concepts

```text
+ ConfidentialityClass      Company-configured vocabulary; product-owned static semantics
+ ConfidentialityGrant      { subject: User | Group, class, scope: Company | Area }
~ Document                  gains exactly one class; default class = unrestricted
```

### 3.2 Permissions — none added

```text
access.manage    administers ConfidentialityGrants and the class vocabulary
```

No new permission family. This is a hard constraint of the promotion, inherited from
seam §5.2.

### 3.3 Operations — census 89 → 97 (proposed)

```text
 90 GET    /api/v1/confidentiality-classes            listConfidentialityClasses
 91 POST   /api/v1/confidentiality-classes            createConfidentialityClass      [Idempotency-Key]
 92 PATCH  /api/v1/confidentiality-classes/{id}       updateConfidentialityClass      [ETag]
 93 PUT    /api/v1/confidentiality-classes/{id}/state archiveConfidentialityClass     [ETag]
 94 GET    /api/v1/confidentiality-grants             listConfidentialityGrants
 95 POST   /api/v1/confidentiality-grants             createConfidentialityGrant      [Idempotency-Key]
 96 DELETE /api/v1/confidentiality-grants/{id}        revokeConfidentialityGrant
 97 PUT    /api/v1/documents/{id}/confidentiality     setDocumentConfidentiality      [ETag]

REFINEMENTS (+0 operations)
 op46 createDocument   accepts optional confidentiality_class_id
 op47 getDocument      projects the class + whether the actor may reclassify
 library / My Work / Governance reads project the class as a recognizable label
```

Supporting census deltas:

```text
application operations        89 → 97
Idempotency-Key creations     11 → 13   (op91, op95)
ETag read/mutation domains    13 → 14   (ConfidentialityClass; Document already a domain)
exact-byte resources           4 → 4    unchanged
```

`api-operation-census.md` §"operation 90 or later requires a new explicit bounded
Product/T6 reopen" is satisfied by **this** instrument once ratified — not before.

### 3.4 Audit

New typed, additive facts only: class created / updated / archived, grant issued /
revoked, document reclassified. No historical Audit meaning is rewritten (seam S4).

## 4. Open questions for P6/P7 (not decided here)

```text
Q1  WHO CLASSIFIES AT CREATION
    (a) the author, restricted to classes they themselves hold clearance for — prevents
        an author making a document unreadable to everyone including themselves
    (b) access.manage only, with creation always defaulting to unrestricted
    Reclassification after creation is access.manage in both hypotheses.

Q2  SCOPE OF A CLASS VOCABULARY
    Company-wide vocabulary, or classes declared per Area? Company-wide is simpler and
    matches DocumentType; per-Area risks re-fragmenting the org chart (seam §7).

Q3  DEFAULT-CLASS REPRESENTATION
    Is "unrestricted" a materialized vocabulary row or the absence of a class? Affects
    read projections and the honest-collection law.

Q4  COLLECTION READS
    Restricted documents must not leak through counts, cursors or gaps (proof §9.3).
    Whether the frontend shows "N results hidden" at all is a UX adjudication.

Q5  GOVERNANCE PARTICIPANTS
    Does an approver routed to a restricted document need clearance, or does routing
    itself confer read? Routing-confers-read is the likely answer but must be decided,
    audited and stated — it is a genuine confidentiality hole if left implicit.
```

## 5. Blocks reopened (bounded rebaseline, Frontend Method §5.3)

```text
B13  Document Creation      OPEN — absorbs the class field before LOCK. No rework:
                            this reopen was authorized BEFORE B13 P8 was locked.
B02  Library / Discovery    class label + filter; hidden-result honesty
B03  Document Official      class display + reclassify affordance
B11  Access Administration  ConfidentialityGrant administration + multi-class issuance
B10 / B12  Administration   class vocabulary administration (owner adjudicated at P6)
FP2 / P11                   re-integration after the above
```

Each reopened block re-enters at P6 with a bounded delta only; already-proved regions
are not re-planned. Their existing LOCK Evidence remains valid for everything the delta
does not touch.

## 6. Invariants this reopen may not violate

```text
seam §5.3 non-identity law      class never enters Document.code, numbering or identity
seam §5.4 permanent exclusions  no per-document ACL, no external/guest, no Area-as-secrecy,
                                no frontend-computed visibility
T3 negative list                no materialized ACL or permission cache becomes authority
no new permission family        administration stays in access.manage
default DENY                    clearance is conjunctive; removing it removes read live
```

## 7. Ratification gate

This instrument is OPEN. It becomes current authority only after:

```text
P6 reference study + P7 blocking law for the reopened blocks
Q1-Q5 adjudicated by the operator
census / idempotency / ETag deltas proved against api-operation-census.md
T3 equation delta accepted in authorization-and-audit.md
Product contract §4 concept + §6 scope tier updated
explicit operator ratification
```

Until then Launch scope remains 89 operations and no confidentiality capability is built.
