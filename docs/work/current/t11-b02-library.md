# T11 — B02 Library / Discovery

> **Status:** LOCKED / OPERATOR-RATIFIED on 2026-08-22.  
> **Block:** B02 — Library / discovery.  
> **Predecessor:** B01 App Shell + Global IA is LOCKED / OPERATOR-RATIFIED.  
> **Selected structure:** C3 — Discovery-first Library with `Por tipo` + `Por área`.  
> **Canonical rendered artifact:** `docs/work/current/t11-b02-library-wireframe.html`.

## 1. Boundary

B02 owns only official-document discovery inside `/documents`.

It does not design Document Official, Document Work, Governance Case, History/Audit detail or Administration detail. Opening a result may navigate to the accepted Document Official route, but B03 owns that destination structure.

Inherited invariants remain:

```text
application operations        78
operation 79                  ABSENT
stable Product SPA paths      unchanged
Product implementation        BLOCKED
T12                           NOT OPEN
```

## 2. Operator adjudication

The initial wide-table-first hypothesis was rejected as the preferred baseline because it exposed a dense register before the user had expressed search/browse intent.

The operator then selected the search-first C direction, refined it with visual Document Type cards, and finally added Area as an equally important business navigation axis.

Final B02 structural decision:

```text
prominent code/title search
→ Explore documents
   [Por tipo] [Por área]
→ optional secondary filters
   Responsável / Status
→ structured result list only after user intent
```

Decision:

```text
A — table first                         REJECTED as preferred baseline
B — local filter rail + table          REJECTED as preferred baseline
C2 — search + Type browse              SUPERSEDED by C3
C3 — search + Type/Area browse         LOCKED
B02                                     LOCKED
```

## 3. User mental model

B02 supports three ordinary discovery intents:

```text
I know roughly what it is
→ search by official code/title

I know what kind of document I need
→ browse by Document Type

I know which business area I care about
→ browse by Area
```

These are two browse views over one Library collection, never two Product authorities.

Examples:

```text
Por tipo → Procedimentos
Por área → Comercial
Por área Comercial + Tipo Procedimentos
```

All resolve through the same `/documents` lens and accepted `listDocuments` filters.

## 4. Accepted Library truth

```text
route                          /documents
operation                      listDocuments (45)
default catalog lens           status=effective
search q                       code + current EFFECTIVE title only
filters                        document_type_id
                               area_id
                               responsible_owner_user_id
                               status
status modes                   effective; obsolete/cancelled only where authorized
ordering with q                exact code
                               → code prefix
                               → title prefix
                               → title contains
                               → code
                               → document_id
ordering without q             code → document_id
pagination                     seek cursor; no offset; no total count
```

`DocumentSummary` supplies:

```text
document id + code
official title/revision when present
Document Type reference
Area reference
responsible-owner User reference
catalog status
```

DRAFT/SUBMITTED never become ordinary Library official results.

## 5. Locked discovery structure

Initial state, in reading order:

```text
1. Library identity + New document entry
2. prominent code/title search
3. Explore documents
   3a. Por tipo
   3b. Por área
4. Ver todos os documentos
5. Mais filtros
6. no automatic result dump
```

A result collection appears only after one of:

```text
q entered
Document Type selected
Area selected
explicit View all documents
accepted secondary filter applied
```

### Por tipo

Each tile uses server-returned:

```text
DocumentTypeReference.code
DocumentTypeReference.name
```

No type counts and no hardcoded semantic icon mapping. Document Type is configurable; `Policy = shield`, `Procedure = gear`, etc. are not Product authority.

### Por área

Each tile uses server-returned:

```text
AreaReference.code
AreaReference.name
```

Selecting an Area applies the existing `area_id` Library filter. It does not create an Area page or a new route.

### Combining axes

Type and Area may be combined because `listDocuments` already admits both filters.

```text
Área Comercial
+ Tipo Procedimento
→ same /documents lens with both semantic filters
```

### Secondary filters

```text
Responsável
Status
```

Responsible owner remains useful but is not promoted to a primary browse axis. Status defaults to Vigentes/effective; historical catalog modes remain authorization-bound.

## 6. Result structure

After the actor expresses intent, results use a structured list rather than a card gallery.

Each item prioritizes:

```text
official title
code
current official Revision
Area
responsible owner
catalog status
Open
```

This intentionally separates:

```text
recognize a discovery context → search / metadata tiles
compare actual documents      → structured list
```

Pagination uses cursor continuation (`Carregar mais documentos`). No `page 1 of N`, total count or free sort UI is implied.

## 7. B02-LD — bounded Library discovery read-composition precision

### Problem closed

The accepted Product already requires Type, Area and responsible owner filters, but opaque filter IDs alone are not human-operable. Admin directories and creation options cannot be silently reused for ordinary Library readers.

B02-F1 is resolved by a bounded read-model precision on the existing Library operation. This creates no operation, owner, Permission or route.

### Wire/read-model precision

Existing operation 45 remains:

```text
GET /api/v1/documents
operationId = listDocuments
```

The first-page `DocumentPage` gains a discovery composition:

```text
DocumentLibraryDiscoveryOptions {
  document_types: DocumentTypeReference[]
  areas: AreaReference[]
  responsible_owners: UserReference[]
}

DocumentPage {
  items: DocumentSummary[]
  page: Page
  discovery_options?: DocumentLibraryDiscoveryOptions
}
```

Presence law:

```text
request cursor absent  → discovery_options required
request cursor present → discovery_options absent
```

Thus continuation pages do not repeat the discovery payload.

### Discovery universe

`discovery_options` is derived from the actor's currently disclosable Library universe for the selected catalog `status`.

When deriving options, the server:

```text
APPLIES
  current authentication
  current Authorization/disclosure
  selected status lens

IGNORES for option-universe derivation
  q
  document_type_id
  area_id
  responsible_owner_user_id
```

This keeps primary browse choices stable while the result query is narrowed.

An option appears only if at least one Document in that status universe is currently disclosable to the actor and references that Type / Area / responsible owner.

Consequences:

```text
no empty global Admin catalog leak
no disclosure of unrelated Users
no inference from hidden Documents
no client-side facet authority
```

### Completeness and ordering

Lists are complete for that actor/status universe and never silently truncated.

```text
document_types      → code ASC, document_type_id ASC
areas               → code ASC, area_id ASC
responsible_owners  → user_id ASC
```

No counts are supplied.

`UserReference.display_name` remains optional under existing privacy/history rules. The frontend must not invent profile data; an unavailable display name is rendered as unavailable identity while retaining the stable User reference.

### Scale posture

The precision deliberately avoids three new paginated endpoints. Discovery options are carried only on the first Library page and contain compact references, not full Admin models.

A synthetic oversize probe (hundreds of Types/Areas and thousands of responsible owners) remains technically representable as a first-page reference payload without requiring a new semantic operation. This is not a Product maximum or performance guarantee; implementation proof must measure realistic payload/query p95. Material production evidence of an unsustainable dimension set reopens only this read composition.

The UI itself does not assume a six-card universe:

```text
small option set  → direct tile grid
large option set  → same complete options + local quick-filter / bounded-height browse region
                     with all choices reachable
```

No arbitrary hidden top-N or popularity ranking is introduced.

### Rejected repairs

```text
Admin listDocumentTypes/listAreas/listUsers as ordinary reader selectors
DocumentCreationOptionsView outside create semantics
hardcoded Type/Area catalogs
UUID entry
page-local facet derivation
load whole Library into browser
generic Search/facet service
operation 79
```

Reconciliation:

```text
new application operations    0
new semantic owners           0
new stable Product paths      0
operation 79                  ABSENT
B02-F1                        RESOLVED
```

This precision must be consolidated into the effective T6/T8-E/T8-F durable owners before T11 integration, just like other approved T11-discovered frontend-read precision.

## 8. Responsive/accessibility structure

Desktop:

```text
B01 global sidebar remains unchanged
Library discovery uses the wide content region
Type/Area tiles form a grid
results become a structured list
```

Narrow widths:

```text
B01 sidebar → drawer
Type/Area grid → fewer columns → one column
filters stack
result metadata wraps below title
```

`Por tipo` / `Por área` are semantic tabs/buttons with keyboard and focus behavior, not color-only switches. Whole tiles are keyboard-activatable controls. No essential action relies on hover.

Large dimension sets use a keyboard-operable local quick-filter/browse region rather than inaccessible visual clipping.

## 9. P9 — Screen Contract / vertical trace

| Region/control | User goal | Accepted truth / operation | Navigation/identity | Material failure intent | Forbidden frontend authority |
|---|---|---|---|---|---|
| Search | known-item lookup | `listDocuments(q)` | URL `q` | validation/denied remain server truth | body/OCR/fuzzy search |
| Por tipo tiles | browse by recognizable document kind | `DocumentPage.discovery_options.document_types` | returned `document_type_id` | option disappearance means current actor/status universe changed | hardcoded Type catalog/counts |
| Por área tiles | browse by business area | `DocumentPage.discovery_options.areas` | returned `area_id` | same | Area page/parallel hierarchy |
| Responsible filter | narrow by responsible owner | `discovery_options.responsible_owners` + existing filter | returned `user_id` | missing display enrichment never invents PII | general User directory |
| Status | choose admitted official catalog lens | existing status filter | URL state | unauthorized historical modes fail closed | client lifecycle authority |
| Results | recognize official document | `DocumentSummary[]` | `document_id` | empty means known-empty for current query, not unknown | DRAFT-as-official |
| Open | reach official lens | stable `/documents/:document_id` | returned document id | destination rechecks current disclosure | B02 owning B03 truth |
| Load more | continue same query | cursor | returned next_cursor | invalid/stale cursor follows server Problem | offset/page-number emulation |
| New document | begin accepted create journey | existing Library creation entry | B01/B02 task entry | destination owns create eligibility | separate create route/owner |

## 10. Bidirectional trace

Product/backend → B02:

```text
listDocuments filters       → search + Type + Area + Responsible + Status
DocumentSummary             → result row
DocumentPage.page           → cursor continuation
B02-LD discovery_options    → human-operable browse/filter selectors
/document/:document_id      → result destination only
```

B02 → Product/backend:

```text
search                       → Controlled Documents / listDocuments(q)
Por tipo                     → listDocuments(document_type_id)
Por área                     → listDocuments(area_id)
combined Type + Area         → same operation with both filters
Responsável                  → listDocuments(responsible_owner_user_id)
Status                       → listDocuments(status)
Open                         → accepted Document Official route
```

## 11. P10 bounded pattern consolidation

B02 contributes local evidence for:

```text
metadata-browse tile
search-first discovery frame
structured official-result row
filter-context chip
cursor continuation
```

No shared Product component/pattern is frozen merely because B01 and B02 have similarly styled boxes. Pattern graduation still requires repeated locked semantic purpose, protected behavior and accessibility/state ownership.

## 12. Deliberate absences

B02 does not add:

```text
body/OCR/vector/full-text search
fuzzy/accent-folded search
tags
favorites
recently accessed
saved views
popular documents
Type/Area counts
arbitrary sorting
page-number pagination
total result count
thumbnail previews
standalone Area route
standalone Template route
frontend permission matrix
operation 79
```

## 13. Reopen law

Reopen only the smallest B02 decision if evidence shows, for example:

```text
direct users cannot find documents through search/Type/Area
real Type/Area scale makes the tile interaction unusable
real discovery-option payload/query cost falsifies B02-LD
responsible-owner filtering requires a different disclosure model
later B03 evidence proves result identity/context is insufficient
responsive/accessibility testing falsifies the structure
```

Visual styling alone belongs later unless it changes hierarchy, grouping, reading order, density class or interaction meaning.

## 14. Progression

B02 is operator-LOCKED. P9 and bounded P10 are complete in this record.

B03 — Document Official is now eligible to be opened as the next CANDIDATE, but is not opened by this record itself.

Do not pre-generate B03+ as baseline. Product implementation remains blocked.
