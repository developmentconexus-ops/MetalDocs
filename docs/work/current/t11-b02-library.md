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
I need to browse by a business concept I recognize, especially Document Type,
so that I can narrow the collection before scanning records.

When I browse the Library,
I need drafts/submissions to stay out of the ordinary official collection,
so that I do not confuse work-in-progress with valid company truth.
```

Operator evidence on 2026-08-22 rejected a collection-dump-first experience for B02 and preferred the search-first structured-list direction, strengthened with visual Document Type browse cards before result disclosure.

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

`DocumentSummary` supplies document code/id, official title/revision when present, Document Type reference, Area reference, responsible-owner User reference and catalog status.

The Library never presents DRAFT/SUBMITTED as equivalent official truth. Those belong My Work / Document Work / Governance lenses.

## 3. Hard UX consequences of authority

B02 MUST NOT invent:

```text
body/OCR/vector/full-text search
fuzzy/accent-folded search
tags
favorites
recently accessed documents
saved views
arbitrary sort controls
offset/page-number pagination
total result count
thumbnail/preview imagery
per-type document counts without admitted aggregation
client-derived lifecycle truth
hardcoded semantic icons for configurable Document Types
```

Document Type is configurable. A robust visual tile therefore uses accepted `DocumentTypeReference.code + name` as its identity. A decorative generic document glyph is permissible; fixed semantic mappings such as `Policy = shield` are not Product authority.

## 4. Reference study — P6 update

References are evidence only.

| Reference | Source observation | Candidate lesson | Mismatch / disconfirming evidence |
|---|---|---|---|
| Qualio Documents Library | Qualio explicitly warns an all-effective-document Library can be overwhelming and directs users to search/filter; later UX widened result tables after narrowing. | Do not make the unfiltered collection the only first impression. | Qualio has tags/full-text search that MetalDocs Launch does not. |
| Veeva Vault search | Document Type can be selected as a search context before or while searching. | Document Type is a legitimate first-class discovery choice, not merely an advanced filter. | Veeva custom tabs/search collections are much broader than MetalDocs. |
| M-Files Views | Objects can be categorized into views/grouping levels from metadata such as object type/class. | Metadata-led browse is valuable when the exact item is unknown. | Saved/configurable views and arbitrary metadata grouping exceed Launch authority. |
| Carbon / enterprise collection patterns | Dense tables/lists remain useful after the user has narrowed a collection; clickable/selectable tiles are appropriate for bounded navigation/options. | Preserve structured comparison in the result state; use simple whole-tile interaction for Document Type choice. | Generic sortable/faceted table/card features are not automatically admitted. |

Product inference:

```text
search-first + Document Type browse
→ optional secondary filters
→ structured result list
```

The Type-card treatment itself is a MetalDocs hypothesis derived from these patterns; it is not copied reference authority.

## 5. P7 adjudication and C2 hypothesis

### A — Wide table first

Structurally feasible but **not operator-preferred** because it exposes a dense register before the user has expressed search/browse intent.

### B — Local filter rail + table

Not preferred because B01 already owns a persistent global left navigation region; a second rail competes for width and orientation.

### C — Search-first structured results

**OPERATOR-PREFERRED DIRECTION**, refined to C2.

### C2 — Discovery-first Library

Initial state:

```text
Library identity + New document
↓
prominent code/title search
↓
Explore by Document Type cards
  card identity = DocumentType.code + DocumentType.name
  no counts
  no hardcoded semantic icon authority
↓
More filters (Area / Responsible / status) as secondary narrowing
↓
no result dump yet
```

A result collection appears only after one of:

```text
search q entered
Document Type selected
explicit "View all documents"
```

Result state uses a structured list, not a preview card grid:

```text
official title primary
code + official Revision
Area
Responsible owner
catalog status
Open action
cursor continuation
```

This separates two tasks deliberately:

```text
recognize a discovery context → cards / search
compare actual documents      → structured list
```

## 6. C2 data feasibility

| Need | Required truth | Feasibility |
|---|---|---|
| Result identity/navigation | `document_id`, `document.code` | PRESENT-IN-AUTHORITY |
| Official title/revision | `official_revision` | PRESENT-IN-AUTHORITY |
| Result Type/Area/Owner/status | `DocumentSummary` references/status | PRESENT-IN-AUTHORITY |
| Search | `q` | PRESENT-IN-AUTHORITY |
| Apply Type/Area/Owner/status filters | existing `listDocuments` query inputs | PRESENT-IN-AUTHORITY |
| Result pagination | seek cursor | PRESENT-IN-AUTHORITY |
| Type card identity | `DocumentTypeReference.code + name` | PRESENT AS A SHAPE; **complete reader option source unresolved** |
| Type counts | aggregation/total per type | DELIBERATELY ABSENT |
| Arbitrary sorting | no admitted sort | DELIBERATELY ABSENT |
| Complete human-readable filter option sources | reader-safe Type/Area/Owner choices | **FINDING B02-F1** |

## 7. B02-F1 — Library discovery-option source

The accepted Product already names Document Type, Area and responsible owner as core Library filters, and `listDocuments` accepts their opaque IDs. Yet the current frontend/wire authority does not prove how an ordinary reader obtains the complete, human-readable, disclosure-safe choices needed to use those filters or C2 Type cards.

Existing surfaces are not silently reusable:

```text
Document Governance list/read → administration/configuration authority
Organization User/Area lists  → organization administration authority
DocumentCreationOptionsView   → purpose-built create eligibility; excludes reader-only users
current DocumentPage items    → incomplete page-local sample, not a complete selector source
```

Rejected repairs:

```text
hardcode Document Types
ask users for UUIDs
derive facet values from one result page
load the whole Library client-side
give ordinary readers Admin-directory access
repurpose creation/options outside its trust purpose
invent operation 79 without a bounded reopen
```

Resolution order before B02 LOCK:

```text
1. prove an existing disclosure-safe read already owns the choices
2. otherwise test the smallest read-model/read-composition precision on accepted Library authority
3. prove scale/disclosure/pagination for Type/Area/Owner options
4. only if the 78-operation contract cannot realize the accepted Library filters coherently,
   classify a material bounded upstream reopen rather than introducing a workaround
```

Operation 79 remains absent unless the repository's normal material-reopen law is actually satisfied and operator-authorized.

## 8. Assumption register delta

| Assumption | Evidence | Influence | Probe / status |
|---|---|---|---|
| Most deployments have a modest enough Document Type set that visual browse is useful | operator/domain hypothesis; Product permits configurable types but sets no count | C2 Type-card region | OPEN — test card behavior with small and large synthetic sets before LOCK |
| Users benefit from narrowing intent before seeing the full register | direct operator evidence + external reference pattern | C2 initial state | VALIDATED for current candidate direction |

A large type set must not force hidden arbitrary priorities. If the grid does not scale under realistic type counts, the Type browse interaction must be refined before B02 LOCK.

## 9. Candidate terminology

```text
Library                  → Biblioteca
Explore by Document Type → Explorar por tipo de documento
EFFECTIVE                 → Vigentes       (candidate label)
OBSOLETE                  → Obsoletos      (candidate label)
CANCELLED                 → Cancelados     (candidate label)
Document Type             → Tipo de documento
Area                      → Área
Responsible owner         → Responsável
```

Technical lifecycle values remain server truth.

## 10. P8 disposition

Current rendered candidate is **C2 — Discovery-first Library**. The operator-viewable HTML is being iterated as the visual reference for this block; the repository-canonical HTML is frozen only when the selected B02 candidate is operator-approved.

B02 cannot be LOCKED while B02-F1 or the material Document-Type-scale assumption remains unresolved.

Do not pre-design Document Official, Document Work, Governance, History/Audit detail or Administration from this block.
