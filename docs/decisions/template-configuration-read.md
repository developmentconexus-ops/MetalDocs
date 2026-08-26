---
id: template-configuration-read
kind: authority
owner: architecture
summary: Bounded T11 authority for server-side template-configuration filtering and search on operation 43 discovered during B12 frontend planning.
---

# Template Configuration Read — bounded T11 authority

> **Status:** OPERATOR-RATIFIED / BOUNDED T11 REOPEN.
> **Ratified:** 2026-08-26.
> **Trigger:** B12 functional P8 operator walkthrough (B12-F2).
> **Method:** `docs/development/engineering-method.md` v1.0.0 + `docs/development/frontend-product-experience-planning-method.md` v2.3.
> **Precedent:** mirrors the B11-F1 op31 read-precision pattern (`access-assignment-read.md`).
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Authority and supersession

This page is the single bounded current authority for the B12 template-configuration read precision exposed by frontend planning.

It does not replace Product/T1→T10 wholesale. It supersedes only conflicting current-tense clauses concerning operation 43 `listTemplateConfigurations`, its first-page query and its collection completeness law in:

```text
../product/journeys.md
../architecture/wire-contract.md
../architecture/frontend.md
api-operation-census.md
```

All unchanged template semantics remain current: template role ownership (op50/51 on the Document), per-type eligibility ownership (op40/41 on the DocumentType, ETag of the type), `TemplateCreationOption` creation-time projection, Audit facts and disclosure laws.

Historical stage snapshots remain truthful for the stage at which they were ratified.

## 2. Proven human jobs

The B12 Modelos lens must let an administrator find and organize the template estate at real scale through canonical configuration:

```text
find a template by name/code without crawling pages
see which templates are usable for one selected DocumentType
distinguish templates with / without an effective revision
distinguish documents flagged as template from ordinary documents
```

### Generality is a derived fact, not a concept

A "general template" does not exist as Product semantics. A template is eligible in the explicit per-type sets owned by op40/41; being eligible in many (or all) types is a **derived fact** the UI may present ("eligible in N types"). No `is_general` flag, no implicit all-types eligibility, no future-type auto-inclusion is introduced. Explicit per-type eligibility remains the controlled-document control model.

### Templates are not Area resources

Templates carry no Area scope in current authority. No Area filter or Area ownership is introduced by this decision.

## 3. Global-Maximum boundary

The smallest sustainable correction is:

```text
refine existing operation 43
+ exact server-side filters and search before pagination
```

Explicitly not required by current Evidence:

```text
new application operation
new write operation or mutation change
generic document-search DSL reuse inside op43
"general template" Product concept or flag
Area-scoped templates
browser crawl of all template pages
client-side post-filter presented as complete
```

The existing Search owner remains the discovery authority for the Library; op43 filtering is administration-scoped configuration reading, not a second search engine.

## 4. Application-operation delta

No operation is added or removed.

```text
43  GET /api/v1/document-governance/templates
    operationId listTemplateConfigurations
    REFINED — same semantic collection/read owner
```

Current cross-contract census remains:

```text
application operations           89
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

## 5. op43 first-page query

When `cursor` is absent, operation 43 admits:

```text
q?                          ShortText; matches document code and current effective title
eligible_document_type_id?  Uuid
template_role?              boolean
has_effective_revision?     boolean
limit?                      integer 1..100; default 20
```

Filter laws:

```text
all supplied filters are conjunctive
q matches case-insensitively against document.code and current_effective_title
q never matches non-effective draft content
eligible_document_type_id filters by current explicit eligibility membership
```

Examples:

```text
?eligible_document_type_id=<T>
  every template configuration currently eligible for DocumentType T

?template_role=true&has_effective_revision=false
  flagged templates not yet usable (no effective revision)

?q=contrato
  configurations whose code or current effective title matches "contrato"
```

Filters execute server-side before seek pagination.

Collection order remains:

```text
document.code ASC, document_id ASC
```

A syntactically valid filter that currently matches no disclosable configuration returns an ordinary empty page. Existence of a DocumentType remains owned by Document Governance type reads; an unknown `eligible_document_type_id` value is an ordinary empty result, not a type-existence oracle.

## 6. Cursor law

The existing global pagination law remains current.

For a continuation request:

```text
cursor + optional limit only
```

The cursor authenticates:

```text
operationId
+ normalized op43 filters
+ seek position
```

Repeating `q`, `eligible_document_type_id`, `template_role` or `has_effective_revision` with a cursor is `400 request.invalid`. Changing `limit` on continuation remains permitted by the existing global pagination law.

## 7. Read projection

`TemplateConfigurationItem` and `TemplateConfigurationPage` remain unchanged:

```text
TemplateConfigurationItem {
  document: DocumentReference,
  template_role: boolean,
  has_effective_revision: boolean,
  current_effective_title?: LongText,   // present iff has_effective_revision=true
  eligible_document_type_ids: unique Uuid[]  // UUID ASC
}
```

No new reference type, label authority or projection is required; this reopen is query-only.

## 8. B12 consumption law

The Modelos lens renders exactly the returned server page under the active filter identity (P10-S3). Type recording (`eligible_document_type_id`) and search (`q`) are server-side; the lens never crawls pages to emulate them and never presents a post-filtered page as complete.

Writes remain unchanged: the lens mutates only the Document's template role (op50/51); eligibility is written only in the owning DocumentType section (op41).

## 9. Proof strategy

The B12 functional P8 must be capable of falsifying this decision with deterministic fixtures:

```text
1. more templates than one page under each filter identity; continuation remains real;
2. type filter returns exactly the explicit eligibility membership, including an empty set;
3. q matches code and effective title only; a draft-only title never matches;
4. template_role/has_effective_revision filters compose conjunctively;
5. changing any filter starts a new first-page identity; stale cursors are not reused;
6. no screen claims a complete template universe from a filtered page.
```

## 10. Reopen triggers

Reopen only on material Evidence such as:

```text
Product introduces a real template taxonomy/category concept
a proven consumer requires cross-field ranking/relevance beyond conjunctive filters
template eligibility acquires new scope semantics (e.g. Area)
real scale proves the admin lens needs the Search owner rather than filtered reads
```

Preference for a richer template-management console is not a reopen trigger.
