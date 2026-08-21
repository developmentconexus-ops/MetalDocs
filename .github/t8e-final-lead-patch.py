from pathlib import Path

p = Path("docs/work/current/proposal.md")
s = p.read_text()

def repl(old, new, count=1):
    global s
    n = s.count(old)
    if n != count:
        raise SystemExit(f"anchor count mismatch expected={count} got={n}: {old[:100]!r}")
    s = s.replace(old, new, count)

repl(
"UUID textual case does not create a distinct key. Fingerprint is over complete validated normalized semantic command fields, not raw HTTP bytes. `createSubmission` additionally fingerprints the authenticated semantic DRAFT precondition token value; CSRF/session transport material is not fingerprint input.",
"""UUID textual case does not create a distinct key. Fingerprint is over **every validated normalized semantic input that can change the command effect/result**, never raw HTTP bytes:

```text
canonical path identifiers
+ semantic query values when an Idempotency-Key operation ever has them
+ normalized JSON command fields
+ semantic conditional value only where command meaning depends on it
```

This makes bodyless resource-scoped creations (`createDocumentRevision`, `createSubmission`, `createGovernanceFeedback`, `createObsolescenceRequest`) distinct across their path resources. `createSubmission` additionally fingerprints the authenticated semantic DRAFT `If-Match` token value. CSRF/session transport material is not fingerprint input."""
)

repl(
"Other human text is nonblank where stated and is bounded by the aggregate JSON ceiling rather than unrelated guessed per-field maxima.",
"Except for scalars that explicitly define normalization above (`CodeInput`, `SearchQuery`, `EmailAddress`), accepted human text is **not** silently trimmed, case-folded or Unicode-normalized by convention. `nonblank` means the supplied value contains at least one non-whitespace code point. Human text is bounded by the aggregate JSON ceiling rather than unrelated guessed per-field maxima."
)

repl(
"Explicit absent/non-disclosable filter ->404; semantically inapplicable valid combination ->422. If real scale makes these complete arrays unsustainable, that evidence reopens T6; T8-E does not truncate or add operation79.",
"An explicitly requested absent/non-disclosable `area_id` or `document_type_id` ->404. Every pair of individually usable Area + DocumentType is semantically admissible at this read surface; inapplicable templates/candidates yield empty/absent sublists rather than an invented 422 mode. If real scale makes these complete arrays unsustainable, that evidence reopens T6; T8-E does not truncate or add operation79."
)

repl("`PermissionCode` is exactly the accepted 14-value T3 dot-spelled vocabulary.",
     "`PermissionCode` is exactly the accepted **15-value** T3 dot-spelled vocabulary.")

repl(
"""ProviderSubjectOption { provider_subject_ref:ProviderSubjectRef, display_hints:string[] }
  display_hints maxItems3; each nonblank <=256
ProviderSubjectSearchView { items:ProviderSubjectOption[] } // maxItems20, provider order""",
"""ProviderSubjectOption { provider_subject_ref:ProviderSubjectRef, display_hints:string[] }
  display_hints maxItems3; each nonblank <=256
ProviderSubjectSearchView { items:ProviderSubjectOption[] } // maxItems20; provider relevance/enumeration order
```

This is a bounded **selection preflight**, not a general directory listing: the Launch provider adapter requests at most 20 matches and exposes at most three presentation hints per subject; the caller refines the required query instead of paginating an administrative provider directory.

```text"""
)

repl(
"""DraftUploadAllocation
  upload_id:Uuid
  upload_url:URI capability
  expires_at:UtcInstant = allocation_time + 15 minutes
  max_bytes:DOC_RAW_MAX_BYTES
  required_headers:map<string,string>""",
"""DraftUploadAllocation
  upload_id:Uuid
  upload_url:URI capability
  expires_at:UtcInstant = allocation_time + 15 minutes
  required_headers:map<string,string>"""
)

repl(
"""now < expires_at + OPEN                        completion may proceed
now < expires_at + READY/unconsumed            exact completion repeat ->204; DRAFT attach may proceed
now >= expires_at + OPEN or READY/unconsumed   410 state.upload_expired; content reclaimable
consumed attachment                            upload claim is no longer an attachable resource""",
"""OPEN + now < expires_at
  completion may proceed

OPEN + now >= expires_at
  410 state.upload_expired; content reclaimable

READY
  exact completion repeat is recognized before claim-expiry handling ->204
  repeat never extends/revives the admission claim

READY + unconsumed claim + now < expires_at
  DRAFT attach may proceed

READY + unconsumed claim + now >= expires_at
  DRAFT attach ->410 state.upload_expired; content reclaimable when no semantic reference protects it

consumed attachment
  upload claim is no longer an attachable resource"""
)

repl("DraftUploadAllocation { upload_id:Uuid, upload_url:URI, expires_at:UtcInstant, max_bytes:ByteCount, required_headers:map<string,string> }",
     "DraftUploadAllocation { upload_id:Uuid, upload_url:URI, expires_at:UtcInstant, required_headers:map<string,string> }")

repl("|44|`getDocumentCreationOptions`|`GET /api/v1/document-creation/options`|`SAFE_READ`|`200 DocumentCreationOptionsView`|`JSON_NO_STORE`|optional area_id,document_type_id; §2.7|`A + N + validation.failed`|",
     "|44|`getDocumentCreationOptions`|`GET /api/v1/document-creation/options`|`SAFE_READ`|`200 DocumentCreationOptionsView`|`JSON_NO_STORE`|optional area_id,document_type_id; §2.7|`A + N`|")

repl("|45|`listDocuments`|`GET /api/v1/documents`|`SAFE_READ`|`200 DocumentPage`|`JSON_NO_STORE`|first page q,document_type_id,area_id,responsible_owner_user_id,status,limit; §2.3/2.7|`A + validation.failed`|",
     "|45|`listDocuments`|`GET /api/v1/documents`|`SAFE_READ`|`200 DocumentPage`|`JSON_NO_STORE`|first page q,document_type_id,area_id,responsible_owner_user_id,status(default=`effective`),limit; §2.3/2.7|`A + validation.failed`|")

repl("|60|`completeRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete`|no body / `UNSAFE_CSRF`|`204` live READY repeat|`NO_STORE`|none|`U + N + S + state.upload_expired + validation.content_invalid`|",
     "|60|`completeRevisionDraftUpload`|`POST /api/v1/revisions/{revision_id}/draft/uploads/{upload_id}/complete`|no body / `UNSAFE_CSRF`|`204`, including exact READY repeat without claim revival|`NO_STORE`|none|`U + N + S + state.upload_expired + validation.content_invalid`|")

repl("`DraftUploadAllocation.max_bytes = DOC_RAW_MAX_BYTES`; `StartDraftUploadRequest.expected_size_bytes <= DOC_RAW_MAX_BYTES`.",
     "`StartDraftUploadRequest.expected_size_bytes <= DOC_RAW_MAX_BYTES`. The allocation does not echo a redundant global `max_bytes`; the client already supplied the exact intended length and the schema owns the Launch ceiling.")

runtime_anchor = """raw HTTP
-> route/raw envelope limit
-> ApplicationSession
-> CSRF for unsafe request
-> central OpenAPI + strict request validation
-> generated typed request boundary
-> semantic application
-> generated typed response boundary
-> HTTP
-> contract fixture validates exact status + headers + body/Problem
```"""
runtime_new = runtime_anchor + """

An executable `kin-openapi v0.142.0` request-validation probe closes the validator split rather than assuming it:

```text
OpenAPI validator
  rejects additionalProperties:false unknown JSON member
  accepts declared JSON member

OpenAPI validator alone does NOT reject
  unknown query parameter
  duplicate scalar query parameter
  duplicate JSON object member
  body on bodyless operation

minimal central envelope guard
  rejects exactly those four missing classes before typed semantic handling
```

Therefore MetalDocs does **not** add a second schema/validation framework. The envelope guard owns only raw/request-shape properties OAS/kin-openapi does not enforce; OpenAPI remains the schema authority."""
repl(runtime_anchor, runtime_new)

p.write_text(s)
