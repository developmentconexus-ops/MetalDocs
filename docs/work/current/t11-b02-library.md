# T11 — B02 Library / Discovery

> **Status:** CANDIDATE / NOT LOCKED.  
> **Block:** B02 — Library / discovery.  
> **Predecessor:** B01 App Shell + Global IA is LOCKED / OPERATOR-RATIFIED.  
> **Boundary:** official-document discovery only; no B03+ screen is opened by this record.

## 1. User need / mental model

B02 exists so an authenticated user can find the current official controlled document without needing to understand backend topology, draft/governance state, or technical identifiers.

Primary jobs:

```text
When I know roughly which official document I need,
I need to find it quickly by code or title,
so that I can open the current official truth.

When I do not know the exact code/title,
I need to narrow the official collection using business metadata I understand,
so that I can recognize the right document without opening many wrong records.

When I browse the Library,
I need drafts/submissions to stay out of the ordinary official collection,
so that I do not confuse work-in-progress with valid company truth.
```

Operator evidence from B01 also establishes a product-level preference for direct access to documents and a richer, task-oriented experience rather than a thin route launcher.

## 2. Bounded accepted authority

B02 uses only current accepted Library authority:

```text
route                          /documents
operation                      listDocuments (45)
default catalog lens           status=effective
search q                       code + current EFFECTIVE title only
filters                        document_type_id
                               area_id
                               responsible_owner_user_id
                               status
status modes                   effective; obsolete/cancelled only where current authority permits
ordering with q                exact code
                               → code prefix
                               → title prefix
                               → title contains
                               → code
                               → document_id
ordering without q             code → document_id
pagination                     cursor / seek; no offset; no total count
```

`DocumentSummary` supplies:

```text
document.code
official_revision.title when official truth exists
official_revision.revision.ordinal
Document Type reference
Area reference
responsible-owner User reference
catalog status
document_id for navigation
```

The Library never presents DRAFT/SUBMITTED as equivalent official truth. Those belong My Work / Document Work / Governance lenses.

## 3. Hard UX consequences of authority

The current contract means B02 MUST NOT invent:

```text
body/OCR/vector/full-text search
fuzzy/accent-folded search
tags
favorites
recently accessed documents
saved views
arbitrary sort controls
sortable columns that imply unsupported server order
offset/page-number pagination
total result count
thumbnail/preview imagery absent from DocumentSummary
client-derived lifecycle truth
```

Search and filters are URL/navigation context, not Product authority.

## 4. Reference study — P6

References are evidence only.

| Reference | Source observation | Benefit | Mismatch / disconfirming evidence | Candidate lesson |
|---|---|---|---|---|
| Qualio Documents Library / 2024+ UX | Effective documents live in a dedicated Library; current UX emphasizes wider document tables, Title before Type, search/filter and action-oriented document handling. | Strong scanability for a controlled-document collection. | Qualio also has tags, full-text search and broader QMS capabilities that MetalDocs Launch does not admit. | Prefer a wide metadata table and prominent search, but only with MetalDocs fields/operations. |
| Veeva Vault Library/Search | Search leads into document results and users can refine with metadata filters; search/filter state is prominent. | Good model for known-item lookup followed by narrowing. | Advanced search, auto-filtering, suggestions and configurable metadata are materially broader than MetalDocs. | Keep search + typed filters visible, but avoid advanced-search machinery. |
| M-Files Search / Views | Search plus metadata filters/facets and views are central to finding documents/objects. | Shows value of metadata-driven browse when exact title/code is unknown. | Saved views, arbitrary metadata facets and personalized defaults require authority MetalDocs does not have. | Metadata filters are useful, but fixed Launch filters should stay simple and bounded. |
| Carbon Data Table | Data tables are appropriate when users navigate many resources and compare structured attributes; toolbar can host search/filter/actions. | Fits code/title/type/area/owner comparison. | Sortable headers and generic table tooling are not automatically valid; MetalDocs server order is fixed. | Use table semantics without implying unsupported sorting or total-count pagination. |

Reference URLs:

```text
https://docs.qualio.com/en/articles/9612864-august-24-new-qualio-user-experience
https://quality.veevavault.help/en/lr/442/
https://userguide.m-files.com/user-guide/latest/eng/searching.html
https://carbondesignsystem.com/components/data-table/usage/
```

## 5. P7 — genuine structural hypotheses

### A — Wide metadata table + top search/filter toolbar — LEADING

```text
page identity + New document
large code/title search
visible fixed filter controls/chips
wide official-document table
cursor-compatible incremental continuation
```

Strengths:

```text
best cross-row comparison
maximum use of available content width
search and browse coexist without another navigation layer
business metadata stays visible while scanning
responsive degradation can move rows to structured list at narrow width
```

### B — Filter rail + wide result table

```text
search across top
left local filter rail
results table on right
```

Strength: filters stay continuously visible and work well for browse-heavy users.

Cost: B01 already owns a global left sidebar; a second persistent left rail compresses the Library and creates competing vertical navigation/filter regions. It becomes increasingly expensive on laptop widths.

### C — Search-first structured result list

```text
large search
compact filter row
one rich result row/card per Document
metadata wraps below title
```

Strength: recognition is strong, responsive behavior is simple and title can dominate.

Cost: poorer comparison across Type/Area/Owner/Revision, lower information density and more scrolling for a document register. No visual thumbnail/preview truth exists to justify a card/grid model.

## 6. Leading-candidate data feasibility

| Need | Required truth | Feasibility |
|---|---|---|
| Row identity/navigation | `document_id`, `document.code` | PRESENT-IN-AUTHORITY |
| Official title/revision | `official_revision` | PRESENT-IN-AUTHORITY |
| Type | `DocumentTypeReference` | PRESENT-IN-AUTHORITY |
| Area | `AreaReference` | PRESENT-IN-AUTHORITY |
| Responsible owner | `UserReference` | PRESENT-IN-AUTHORITY |
| Catalog status | derived catalog status returned by server | PRESENT-IN-AUTHORITY |
| Code/title lookup | `q` | PRESENT-IN-AUTHORITY |
| Filter application | type/area/owner/status query inputs | PRESENT-IN-AUTHORITY |
| Pagination | seek cursor + optional limit | PRESENT-IN-AUTHORITY |
| Arbitrary sorting | no admitted sort parameter | DELIBERATELY ABSENT |
| Total result count | no admitted count | DELIBERATELY ABSENT |
| Human-readable complete filter option sources | complete, disclosure-safe Type/Area/Owner choices usable by an ordinary Library reader | **FINDING B02-F1** |

### B02-F1 — Library filter option source

The Product authority names Document Type, Area and responsible owner as core Library filters, and `listDocuments` accepts their opaque IDs. Current B02 authority does not yet prove a complete disclosure-safe source of human-readable selector options for an ordinary Library reader.

Rejected shortcuts:

```text
ask users to paste UUIDs
reuse Admin directories regardless of access
reuse document-creation/options outside its least-privilege creation purpose
build options from only the current result page
fetch the entire Library client-side and derive facets
invent operation 79 silently
```

This is a SCREEN-CONTRACT / read-composition finding. It does not justify changing Product semantics or the 78-operation census by preference. Before B02 can be LOCKED, resolve it through the smallest accepted owner/read-model precision or prove an existing disclosure-safe source already owns it.

## 7. Candidate terminology

```text
Library                  → Biblioteca
EFFECTIVE catalog mode   → Vigentes       (CANDIDATE user-facing label)
OBSOLETE catalog mode    → Obsoletos      (CANDIDATE)
CANCELLED catalog mode   → Cancelados     (CANDIDATE)
Document Type            → Tipo de documento
Area                     → Área
Responsible owner        → Responsável
```

Technical lifecycle values remain server truth; these labels are presentation terminology only and remain candidate until operator adjudication.

## 8. P8 disposition

Render A/B/C only for the B02 Library collection frame. The leading A candidate may be visually discussed, but **B02 cannot be operator-LOCKED while B02-F1 remains unresolved**.

Do not pre-design Document Official, Document Work, Governance, History/Audit detail or Administration from this block.
