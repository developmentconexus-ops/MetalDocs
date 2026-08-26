---
id: content-format-vocabulary
kind: authority
owner: architecture
summary: Bounded T11 authority restoring the Product-intended open-ended source-format scope by widening the closed ContentFormat vocabulary and binding official-rendition requirement to converter availability.
---

# Content Format Vocabulary — bounded T11 authority

> **Status:** CANDIDATE / AWAITING OPERATOR RATIFICATION.
> **Trigger:** FP2-F2, post-P11 whole-product coherence alignment.
> **Method:** `docs/development/engineering-method.md` v1.0.0 + `docs/development/frontend-product-experience-planning-method.md` v2.3 §3.10A.
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists — the contradiction

The Product Contract never closed content format:

```text
contract §4 Rendition   "A Document Type may be source-only or require one derived official
                        representation SUCH AS PDF"      → example, not a closed rule
contract §4 Submission  "freezes exact content"          → no prescribed format
contract §6 Launch Core "source/official representation read/download"  → no format list
```

The closure exists only in the realization layer:

```text
wire-contract §2      ContentFormat = docx | pdf
content-integrity §3  content_format = closed vocabulary (as realized)
journeys              "SourceOnly DOCX"
wire-contract §2.9    exact-byte Content-Type table lists docx + pdf only
```

`CNT-14 — PRESERVE` states the DOCX adapter (EigenPal) **never owns semantic truth**; it is an in-app **editing** mechanism. A closed two-value vocabulary therefore let an editing-mechanism constraint narrow Product scope without a Product decision authorizing it — the `NO backend-shaped UX` failure named by Frontend Method §3.10A.

This decision restores Product intent. It introduces no new Product concept, operation, permission or lifecycle.

## 2. Proven human job

A controlled document estate is not only Word text. Companies govern spreadsheets (calculation sheets, matrices), presentations, scanned/native PDFs and image-based records under the same identity, numbering, approval, effectivity, audit and access rules. Forcing those out of MetalDocs either loses them to an ungoverned drive or fakes them as DOCX.

In-app editing remains DOCX-only; other formats are governed through upload/download of exact bytes.

## 3. Application-operation delta

```text
none — no operation is added, removed or resemanticized
```

Current cross-contract census remains:

```text
application operations           89
Idempotency-Key creations        11
ETag read / mutation domains     13 / 13
exact-byte resources             4
```

## 4. Widened ContentFormat vocabulary

`ContentFormat` remains a **closed, server-owned, extensible** vocabulary. Launch value set:

```text
docx   application/vnd.openxmlformats-officedocument.wordprocessingml.document
xlsx   application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
pptx   application/vnd.openxmlformats-officedocument.presentationml.presentation
pdf    application/pdf
png    image/png
jpeg   image/jpeg
txt    text/plain
csv    text/csv
```

Laws preserved unchanged from `content-integrity.md`:

```text
actual ContentFormat is derived server-side from the exact admitted bytes
client filename / client Content-Type never decide ContentFormat
SHA-256 + size_bytes + content_format remain the exact-content descriptor
malware inspection gates every format identically before governed admission
create-once / no-overwrite, OPEN→READY admission and restore verification unchanged
a format outside the vocabulary is rejected at admission with validation.failed
```

### 4.1 Structural admission laws per format

The existing closed structural laws are DOCX/PDF-shaped. Widening the vocabulary requires each
format to carry its own explicit structural law; no format is admitted with weaker protection.

```text
xlsx / pptx   valid top-level OOXML/OPC ZIP with the format's own main part
              (SpreadsheetML workbook / PresentationML presentation)
              INHERIT UNCHANGED from the DOCX law: duplicate canonical part names rejected,
              absolute/parent-traversal paths rejected, symlink entries rejected,
              no recursive expansion of embedded archives (archive depth exactly one),
              streamed expansion enforcing the cumulative expanded-byte + entry-count ceilings
              (DOCX_EXPANDED_MAX_BYTES / DOCX_MAX_ZIP_ENTRIES become the OOXML ceilings,
               renamed by mechanism decision without changing their values)
              encrypted/password-protected packages rejected
              MACRO-ENABLED packages rejected — xlsm/pptm/docm are NOT vocabulary values

pdf           unchanged: structurally parseable; encrypted/password-protected rejected

png / jpeg    valid decodable image container; declared dimensions and decoded pixel budget
              bounded before decode (decompression-bomb defense); embedded thumbnails/metadata
              never recursively expanded

txt / csv     byte-size bounded by DOC_RAW_MAX_BYTES; no parsing, no formula evaluation,
              no encoding transformation — exact bytes are preserved verbatim
```

Delivery hardening for the widened set (`wire-contract §2.9`):

```text
Content-Disposition attachment for every format whose bytes a browser would otherwise
render in the application origin (txt, csv, png, jpeg, and any later scriptable format);
docx/xlsx/pptx/pdf keep their existing disposition behavior
X-Content-Type-Options nosniff and no-transform remain mandatory for every format
```

Adding a later value is an ordinary bounded vocabulary decision (detector support + Content-Type + viewer/converter disposition); it is never implicit.

## 5. Editing versus governing

```text
EDIT IN APP        docx only (CNT-14 adapter evidence; never semantic truth)
GOVERN / UPLOAD    every vocabulary format — identity, numbering, revisions, submission,
                   governance route, effectivity, obsolescence, audit, access are format-neutral
READ / DOWNLOAD    exact bytes for every format under the §2.9 delivery law, with that format's
                   Content-Type added to the exact-byte table
```

No format receives a weaker governance, audit or access rule.

## 6. Official rendition and converter availability

`RepresentationPolicy` keeps its two accepted kinds. The requirement binds to converter availability:

```text
source already PDF + required PDF      -> existing reuse path (no copy, no renderer)
source DOCX + required PDF             -> existing render path
source in a format with an accepted converter -> same render path
source in a format with NO accepted converter -> require_official_rendition is INADMISSIBLE
                                                 for that source; source_only applies
```

Launch accepted converters: `docx -> pdf` (existing). `xlsx`/`pptx` converters may be added later as bounded mechanism decisions without changing this law.

Consequences that must remain visible to humans rather than silently inferred:

```text
the creation/upload surface states, before the author commits bytes, that a Document Type
requiring an official PDF accepts only convertible sources

a non-convertible source under a PDF-requiring type fails with a named validation problem;
the frontend never silently downgrades the type's representation policy
```

This preserves `contract §4` ("a representation cannot silently change the governed content") and adds no per-format policy matrix to the Document Type configuration surface (B12 stays LOCKED).

## 7. Supersession

Supersedes only the conflicting current-tense format-closure clauses in:

```text
../architecture/wire-contract.md   (ContentFormat enum; §2.9 exact-byte Content-Type table)
../architecture/content-integrity.md (realized closed vocabulary examples)
../product/journeys.md             ("SourceOnly DOCX" read as docx-only rather than source-only)
```

Unchanged: every byte-exactness, malware, admission, GC, restore, ETag, idempotency, authorization and audit law.

## 8. Explicitly NOT introduced

```text
generic file drive / uncontrolled document class   (excluded by contract §1 North Star)
per-format Document Type policy matrix
in-app editing for non-DOCX formats
format-specific permissions, lifecycle or numbering
client-side format detection as authority
unbounded "any binary" acceptance
```

## 9. Proof strategy

```text
1. a governed xlsx Document completes create → submit → route → EFFECTIVE with a source_only type;
2. its exact bytes download with the correct Content-Type, digest and no transformation;
3. malware rejection behaves identically for xlsx and docx;
4. a non-convertible source under a PDF-requiring type fails with the named validation problem
   and mutates nothing;
5. a mislabelled file (client Content-Type xlsx, real bytes docx) is classified by server detection;
6. audit/history/access read identically regardless of format.
```

## 10. Reopen triggers

```text
a named consumer requires in-app editing of a non-DOCX format
a converter for xlsx/pptx becomes a committed requirement
a proven need arises for a format outside the vocabulary
records/retention semantics start differing by format
```
