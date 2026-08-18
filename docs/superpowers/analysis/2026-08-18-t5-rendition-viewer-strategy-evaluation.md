# T5 Subgate — Rendition / Viewer Strategy Evaluation

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — OPERATOR DECISION PENDING  
> **Date:** 2026-08-18  
> **Parent stage:** T5 — Durable Async, Search & External Effects  
> **Implementation:** BLOCKED

## 1. Why this subgate exists

The active T5 candidate originally treated `official_rendition_render` as one of the mandatory durable job types. The operator challenged the implied product assumption: MetalDocs will ingest both native PDF and DOCX, and a PDF used only to make DOCX easy to view is not automatically the same thing as a governed `OfficialRendition` that must be persisted and gate Release.

Therefore T5-D/T5-E and the exact meaning of `official_rendition_render` are **paused for this subgate**. T5-A→T5-P must not be adjudicated as a whole until this distinction is settled.

## 2. Binding baseline from T1→T4

```text
source content remains exact semantic truth on WorkingContent / Submission
OfficialRendition exists only when representation policy requires it
SourceOnly is a valid representation policy
preview/viewer mechanism is not semantic authority
storage/provider identity is mechanism only
PDF source and DOCX source are both valid exact content formats
```

T4 already guarantees that if an OfficialRendition exists, it freezes its own exact descriptor + managed-content handle. This subgate decides **when one is actually needed** and what viewing mechanism to use otherwise.

## 3. Reference-software patterns observed

### Veeva Vault — persistent derived viewable PDF

Veeva automatically creates a PDF `Viewable Rendition` for each document version. The rendition is used for inline viewing/annotations/download and is a version-specific derived representation. Reference: Veeva Vault Help, `About Auto-Generated Viewable Renditions`.

This pattern is justified by named consumers such as annotations and standardized view/download behavior.

### M-Files — native preview plus optional PDF

M-Files previews Microsoft Word/Excel/PowerPoint and PDF directly. PDF conversion is separately configurable/optional and can keep the original file or add a separate PDF. References: M-Files User Guide, `Document preview` and `Convert to PDF format`.

### DocuWare — multi-format viewer / Office for Web

DocuWare's Viewer supports PDF and Microsoft Office formats including DOCX. Its current viewer renders PDFs client-side and can open Microsoft Office documents through Microsoft for the Web/native applications. References: DocuWare Knowledge Center, `File formats supported by the Viewer` and `Viewing documents in new DocuWare`.

### Microsoft 365 / WOPI — direct DOCX browser rendering

Microsoft's WOPI viewing scenario renders `.docx` directly in the browser from the source file through a WOPI client. Reference: Microsoft Learn `[MS-WOPI]: Viewing a Document`.

### ONLYOFFICE — direct DOCX viewer/editor plus explicit conversion service

ONLYOFFICE Docs views/edits DOCX in the browser and exposes document conversion as a separate service, including DOCX→PDF. References: ONLYOFFICE Docs API `How it works`, `Converting and downloading file`, and Conversion API.

### Gotenberg / LibreOffice — simple server-side PDF conversion

Gotenberg exposes self-hosted Office→PDF conversion using LibreOffice. Its troubleshooting docs explicitly show that fonts/resource constraints can affect conversion behavior, so fidelity must be empirically proven. References: Gotenberg `Convert to PDF`, `Installation`, and `Troubleshooting`.

### EigenPal / docx-editor — browser-native DOCX editing/viewing

EigenPal's open-source DOCX editor is client-side and edits DOCX natively. It remains the lowest-infrastructure candidate already preserved by the MetalDocs decision registry for DRAFT and read-only DOCX rendering, subject to a fidelity corpus. Reference: `eigenpal/docx-editor` GitHub repository.

### PDF.js — direct PDF browser viewer

PDF.js provides a browser PDF viewer/display layer. For source PDF content, a second generated PDF is unnecessary unless a separate business representation is required. Reference: Mozilla PDF.js Getting Started.

## 4. Four architecture options

### Option A — convert every DOCX to a persisted PDF

Pros:
- stable viewing format;
- fast repeat reads;
- easy print/download;
- strongest similarity to regulated Vault products.

Cons:
- creates storage + async work for every DOCX even when no business consumer requires a PDF;
- rendering failures can block otherwise valid source documents;
- risks making a viewing copy look like semantic truth;
- duplicates content for a capability that may be satisfied by native DOCX viewer.

**Reject as universal Launch default unless the operator deliberately wants PDF to be the official human-readable representation for every controlled DOCX.**

### Option B — client/native viewer only; never persist generated PDF

Pros:
- minimum mechanism;
- no render queue/storage duplication;
- DOCX remains the only source truth.

Cons:
- viewing fidelity depends on viewer/runtime/fonts;
- repeated server/on-the-fly conversion may cost CPU if chosen viewer internally converts;
- no stable PDF for controlled print/export where that becomes a requirement.

**Credible Launch baseline for `SourceOnly`.**

### Option C — rebuildable ViewableRendition cache, not semantic OfficialRendition

A DOCX may be converted to PDF on first view or asynchronously for UX and cached by exact source descriptor + renderer profile. It may be regenerated or discarded because the source Submission remains truth.

Pros:
- consistent/faster read UX after first render;
- avoids making PDF a Release gate;
- mirrors the regenerable/viewing-rendition pattern used by document platforms.

Cons:
- adds cache invalidation/renderer-profile mechanism that Launch may not need;
- still stores duplicate content.

**Defer until native DOCX viewer performance/fidelity proves inadequate.**

### Option D — hybrid: native viewer by default; persisted OfficialRendition only by DocumentType policy

```text
PDF source
  → direct PDF viewer
  → source itself remains exact truth
  → no duplicate generated PDF unless post-processing/business policy requires it

DOCX source + SourceOnly
  → direct read-only DOCX viewer
  → no governed PDF generated

DOCX source + RequireOfficialRendition(PDF)
  → exact Submission source
  → durable server-side PDF render
  → T4 admission
  → immutable OfficialRendition stored in managed content
  → Release gate satisfied only by that exact rendition
```

**Recommended Global Maximum.**

It preserves the existing T1/T2 representation-policy seam instead of inventing a universal PDF policy.

## 5. Important semantic distinction

```text
Source content
  = exact content the author edited/submitted

Viewable preview / viewable rendition
  = derived display mechanism
  = may be direct client render, service render, or rebuildable cache
  = not automatically a Release gate

OfficialRendition
  = policy-required immutable representation
  = has its own T4 exact descriptor/handle
  = may gate Release
```

A PDF used only to make a DOCX convenient to read must **not** silently acquire the meaning of `OfficialRendition`.

## 6. Storage law

If a generated PDF must be retained, its bytes do **not** belong in PostgreSQL. They live in the T4 `ManagedContentStore`; PostgreSQL keeps only the semantic `OfficialRendition` reference, exact descriptor and opaque handle.

A preview/cache PDF, if later added, is mechanism-only and remains rebuildable from the exact source.

## 7. Recommended Launch viewer mapping

```text
source = PDF
  viewer = PDF.js / equivalent client PDF viewer
  generated PDF = NO by default

source = DOCX, DRAFT
  viewer = EigenPal/docx-editor editing/view mode
  generated PDF = NO

source = DOCX, SUBMITTED/EFFECTIVE, SourceOnly
  first candidate = EigenPal read-only viewer
  stronger self-hosted candidate = ONLYOFFICE viewer
  generated persistent PDF = NO by default

source = DOCX, RequireOfficialRendition(PDF)
  renderer candidates = Gotenberg/LibreOffice vs ONLYOFFICE conversion
  output = persisted managed-content PDF + OfficialRendition descriptor
```

Microsoft 365/Office for the Web is a high-fidelity reference pattern, but a standalone MetalDocs integration introduces an external Microsoft/WOPI product dependency and should not be Launch default without a deliberate product decision.

## 8. Renderer selection must be empirical

Do not choose Gotenberg, ONLYOFFICE or a client renderer from marketing claims alone. Build a representative MetalDocs DOCX conformance corpus and compare against Microsoft Word's PDF output/reference rendering.

Required discriminator documents should cover at least:

```text
headers/footers + page numbers
company fonts
multi-page tables / repeated table headers
images and wrapping
styles/headings/TOC
manual page/section breaks
landscape sections
links/bookmarks
tracked changes/comments when present
fields/placeholders used by MetalDocs
long Portuguese text with accented characters
```

Score:

```text
pagination
line wrapping
font substitution
image placement
header/footer fidelity
table fidelity
link/bookmark behavior
conversion latency
CPU/RAM
failure visibility
security/isolation
operational burden
```

## 9. Proposed correction to T5

The active T5 recommendation is refined from:

```text
mandatory durable jobs = official_rendition_render + search_refresh
```

to:

```text
always-required durable projection job:
  search_refresh

conditional durable job when current frozen representation policy requires it:
  official_rendition_render

preview/viewer:
  synchronous/client/service mechanism owned by T6 UX design;
  no durable job or persisted PDF is required merely to view DOCX.
```

Managed-content GC remains a periodic reconciliation mechanism.

## 10. Remaining material choices for operator

```text
RV-1 — ACCEPT hybrid Option D as Launch policy.
RV-2 — For SourceOnly DOCX, start with EigenPal read-only rendering; keep ONLYOFFICE as a stronger candidate if conformance fails.
RV-3 — PDF source is viewed directly; do not duplicate PDF by default.
RV-4 — Persistent generated PDF exists only for RequireOfficialRendition(PDF) or another named future consumer.
RV-5 — Renderer product choice remains evidence-driven through a DOCX fidelity corpus before implementation; architecture does not freeze Gotenberg/ONLYOFFICE yet.
RV-6 — Refine T5 job census so `official_rendition_render` is conditional, not universally mandatory.
```

Until RV-1→RV-6 are adjudicated, **T5-A→T5-P as a whole is paused** and T6 remains NOT OPEN.