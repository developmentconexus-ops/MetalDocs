from pathlib import Path

p = Path("docs/work/current/proposal.md")
s = p.read_text()


def replace_once(old: str, new: str, label: str) -> None:
    global s
    count = s.count(old)
    if count != 1:
        raise SystemExit(f"{label}: expected exactly one anchor, found {count}")
    s = s.replace(old, new, 1)


def replace_section(start: str, next_heading: str, replacement: str) -> None:
    global s
    i = s.find(start)
    if i < 0:
        raise SystemExit(f"section start missing: {start}")
    j = s.find(next_heading, i)
    if j < 0:
        raise SystemExit(f"next heading missing: {next_heading}")
    s = s[:i] + replacement.rstrip() + "\n\n---\n\n" + s[j:]


replace_once(
    """ProviderSubjectRef
  opaque/nonblank, maxLength=2048
  server-resolvable anti-corruption handle for exact issuer+subject
  unchanged binding returns byte-stable ref
  client never parses it or treats it as Product identity
```""",
    """ProviderSubjectRef
  opaque/nonblank, maxLength=2048
  server-resolvable anti-corruption handle for exact issuer+subject
  unchanged binding returns byte-stable ref
  client never parses it or treats it as Product identity

EmailAddress
  trim surrounding whitespace
  minLength=3; maxLength=254; OpenAPI format=email
  no case-folding/canonicalization, uniqueness or verification claim
  profile/contact metadata only; never authentication or Authorization identity
```""",
    "EmailAddress scalar",
)

replace_once(
    "|46|`createDocument`|`POST /api/v1/documents`|`CreateDocumentRequest` / `IDEMPOTENT_CREATE`|`201 CreateDocumentResult`|`JSON_NO_STORE`|none|`U + J + I + S`|",
    "|46|`createDocument`|`POST /api/v1/documents`|`CreateDocumentRequest` / `IDEMPOTENT_CREATE`|`201 CreateDocumentResult`|`JSON_NO_STORE`|none|`U + J + I + S + X`|",
    "createDocument integrity error",
)
replace_once(
    "|52|`createDocumentRevision`|`POST /api/v1/documents/{document_id}/revisions`|no body / `IDEMPOTENT_CREATE`|`201 CreateRevisionResult`|`JSON_NO_STORE`|none|`U + N + I + S`|",
    "|52|`createDocumentRevision`|`POST /api/v1/documents/{document_id}/revisions`|no body / `IDEMPOTENT_CREATE`|`201 CreateRevisionResult`|`JSON_NO_STORE`|none|`U + N + I + S + X`|",
    "createDocumentRevision integrity error",
)
replace_once(
    "|62|`createSubmission`|`POST /api/v1/revisions/{revision_id}/submissions`|no body / `SUBMISSION_CREATE`|`201 SubmissionCreateResult`|`JSON_NO_STORE`|none|`U + N + I + D + S + validation.failed + validation.content_malicious + dependency.malware_inspector_unavailable`|",
    "|62|`createSubmission`|`POST /api/v1/revisions/{revision_id}/submissions`|no body / `SUBMISSION_CREATE`|`201 SubmissionCreateResult`|`JSON_NO_STORE`|none|`U + N + I + D + S + X + validation.failed + validation.content_malicious + dependency.malware_inspector_unavailable`|",
    "createSubmission integrity error",
)

section7 = r'''# 7. Document admission limits — measured Launch candidate

The Launch candidate freezes exactly three resource ceilings:

```text
DOC_RAW_MAX_BYTES        = 104857600  // 100 MiB; DOCX and PDF
DOCX_EXPANDED_MAX_BYTES  = 268435456  // 256 MiB; streamed top-level OPC expansion
DOCX_MAX_ZIP_ENTRIES     = 4096       // top-level ZIP entries
```

These are application admission limits, not claims about the theoretical maximum accepted by Microsoft Word, S3, a DAM, or another provider.

Measured real corpus supplied during T8-E:

```text
ForgeFlow_Arquitetura_Base_v01.docx
  raw bytes              22,863
  ZIP entries            24
  file entries           20
  expanded bytes         284,172
  embedded media entries 0
  expanded/raw           12.43x

PO-05-04 Projeto e Desenvolvimento.pdf
  raw bytes              445,131
  pages                  11
  encrypted              no
```

The largest measured real sample therefore sits below the candidate ceilings by approximately:

```text
raw bytes       235x
DOCX expansion  944x
ZIP entries     170x
```

The deliberately large headroom is necessary because the supplied DOCX contains tables/formatting but no embedded media, while ordinary future controlled documents may contain images, headers/footers, page/section breaks and richer OOXML parts. The limits are still materially below generic DAM/large-media scales.

Adversarial disposable probes demonstrate that each control protects a different resource:

```text
expanded_above_256m.docx
  raw       306,139 bytes
  expanded  314,572,846 bytes
  -> reject by DOCX_EXPANDED_MAX_BYTES

many_entries.docx
  raw       628,056 bytes
  entries   5,002
  expanded  320,040 bytes
  -> reject by DOCX_MAX_ZIP_ENTRIES

duplicate_parts.docx
  duplicate canonical ZIP part name
  -> reject structurally

traversal.docx
  parent-traversal ZIP path
  -> reject structurally

expanded_bomb.docx
  raw       130,838 bytes
  expanded  134,217,774 bytes
  -> high compression ratio alone is NOT invalid because actual resource use remains inside the explicit expanded-byte budget
```

Industry comparables are sanity bounds only, never Product authority: controlled-document products commonly admit individual files around the 100 MB class, while DAM/large-media products allow materially larger files. MetalDocs therefore keeps a conservative controlled-document Launch ceiling rather than importing DAM-scale behavior.

Closed structural laws:

```text
DOCX = valid top-level OOXML/OPC ZIP with WordprocessingML main document
duplicate canonical ZIP part names rejected
absolute/parent-traversal paths rejected
symlink entry extraction rejected
no recursive expansion of embedded archives
stream expansion while enforcing cumulative expanded-byte + entry-count ceilings
encrypted/password-protected or macro-enabled/non-DOCX Office packages rejected as DOCX
PDF = structurally parseable PDF; encrypted/password-protected PDF rejected at Launch
client filename/Content-Type never decides actual ContentFormat
```

Validation does not recursively unpack arbitrary embedded archives, so application archive depth is exactly one. No generic nested-archive framework or compression-ratio threshold is added. Malware inspection remains a separate exact-byte governed-boundary control.

Boundary behavior is exact:

```text
expected_size_bytes > DOC_RAW_MAX_BYTES
  -> allocation rejected before provider capability exists

actual provider bytes != expected_size_bytes
  -> completion rejected; READY not established

actual DOCX expanded bytes > DOCX_EXPANDED_MAX_BYTES
  -> 422 validation.content_invalid

actual DOCX ZIP entries > DOCX_MAX_ZIP_ENTRIES
  -> 422 validation.content_invalid

structural path/duplicate/encryption/package violation
  -> 422 validation.content_invalid
```

`DraftUploadAllocation.max_bytes = DOC_RAW_MAX_BYTES`; `StartDraftUploadRequest.expected_size_bytes <= DOC_RAW_MAX_BYTES`.

Multipart, recursive archive inspection, compression-ratio thresholds and DAM-scale upload machinery remain absent. A later measured ordinary controlled document that cannot fit these ceilings is a bounded admission-limit reopen, not permission to raise limits silently.'''
replace_section(
    "# 7. Document admission limits — measured prerequisite",
    "# 8. Bounded upstream findings exposed by T8-E — RESOLVED",
    section7,
)

section9 = r'''# 9. Generation / provider feasibility and runtime conformance proof

Disposable probe pins are evidence only; they do not pre-authorize T8-G runtime/toolchain choices:

```text
Go          oapi-codegen v2.8.0 strict-server
TypeScript  openapi-typescript 7.13.0 paths/components
S3 probe    AWS SDK for Go v2 service/s3 v1.106.2
```

## 9.1 Generated boundary feasibility — PASS

A disposable OpenAPI 3.0.3 probe exercised the actual T8-E encoding patterns rather than a scalar-only toy schema:

```text
additionalProperties:false
required nullable member
optional non-nullable member
closed string enum
safe-integer maximum 9007199254740991
oneOf + discriminator union
multiple success responses 200 + 201
operation-specific RFC9457 response schemas
strict Go server response objects
TypeScript paths/components
```

Execution evidence:

```text
oapi-codegen v2.8.0
  generator built successfully
  runner Go 1.24.13 automatically acquired Go 1.26.7 because generator requires Go >=1.25
  generated Go compiled
  generated Go tests PASS

openapi-typescript 7.13.0
  generated declarations successfully
  TypeScript 5.9.2 strict noEmit probe PASS
```

Observed generated properties:

```text
Go
  fixed objects gained no AdditionalProperties field
  required nullable stayed non-omitempty
  optional member stayed omitempty
  required nullable Page.next_cursor serialized as explicit null
  strict response set contained distinct 200 and 201 JSON response objects

TypeScript
  closed enum -> "active" | "retired"
  next_cursor -> string | null
  required_nullable -> required string | null
  optional_nonnullable -> optional string
  200 and 201 response keys both present
  negative compile probes rejected undeclared enum, missing required nullable,
    extra fixed-object member and unknown discriminator
```

`oapi-codegen` represents `oneOf` internally with private `json.RawMessage` union storage plus typed conversion helpers. That is generated encoding mechanism, not a public `any`/map, provider identifier or second DTO authority.

The generator's build-time Go requirement is **not** a MetalDocs runtime Go-version decision; T8-G remains free to select the compatible runtime/toolchain floor.

## 9.2 Direct S3 PUT constraint feasibility — PASS

A disposable AWS SDK Go v2 presign probe used:

```text
PutObjectInput:
  ContentLength = 12345
  IfNoneMatch   = "*"
```

and produced:

```text
Content-Length=["12345"]
If-None-Match=["*"]
X-Amz-SignedHeaders=content-length;host;if-none-match
```

Therefore the concrete reference provider can bind both exact body length and create-only precondition in the signed PUT request without POST-form/multipart.

Browser feasibility also closes the public contract:

```text
If-None-Match   browser-settable request header; not Fetch-forbidden
Content-Length  script-forbidden but user-agent generated from the fixed Blob/File body
cross-origin PUT uses normal provider CORS preflight for returned browser-settable headers
```

Exact provider CORS configuration belongs T8-G. T8-E requires only that `required_headers` be returned and applied verbatim and that the upload body have known fixed length.

## 9.3 Closed ledger fixture proof — PASS

A mechanical Lead proof compared the candidate ledger against the canonical Product/T6 census and special profiles:

```text
ledger rows                         78
row numbers                         exact 1..78
method+path census                  exact match; zero missing/extra
operationId                         78 unique
family partition                    3 / 26 / 4 / 10 / 34 / 1
Idempotency-Key POST creations      exact accepted 10
ETag concurrency domains            13 GET/mutation domains
exact-byte application resources    exact accepted 4
operation 79                         absent
```

The proof also checks that unsafe application operations use a CSRF-bearing request profile and that JSON-body operations carry the closed structural/media/size validation family where applicable.

The final OpenAPI must expand proposal macros (`A`, `B`, `U`, `J`, etc.) into explicit per-operation response schemas; macros never survive as executable ambiguity.

## 9.4 Runtime conformance contract

T8-E closes the proof architecture; it does not fabricate an application runtime while implementation is blocked:

```text
raw HTTP
-> route/raw envelope limit
-> ApplicationSession
-> CSRF for unsafe request
-> central OpenAPI + strict request validation
-> generated typed request boundary
-> semantic application
-> generated typed response boundary
-> HTTP
-> contract fixture validates exact status + headers + body/Problem
```

Required negative/edge fixture classes:

```text
all 78 rows and no 79th
unknown path ->404; undeclared method ->405 exact Allow
unknown JSON/query member and duplicate scalar/member rejection
bodyless operation rejects a body
wrong media/content coding and 65,536-byte JSON ceiling
role bundles/scope matrix
wrong-domain/tampered/stale ETags + exact-current PUT exception
stale DRAFT always412
PROFILE_REPLACE If-Match/If-None-Match matrix
Idempotency-Key replay/different fingerprint/24h expiry/current-AuthZ recheck
cursor tamper/filter replay/ordering
complete creation/options arrays
upload exact Content-Length/create-once/shared15min expiry + completion size re-proof
100 MiB raw / 256 MiB expanded / 4096-entry admission boundaries
duplicate/traversal/encrypted/invalid package rejection
Governance Step historical label snapshot survives current-route relabel
Audit operation/resource/facts combinations exclude provider_binding.disabled
exact bytes verified before response commit
Range/redirect/206/304/compression absent
Content-Digest == exact body SHA-256
corrupt semantic bytes ->500 internal.content_integrity with zero success bytes
semantic byte-copy mutations may emit only their declared internal.content_integrity path
ReplaySnapshot <=2048
PDF-source RequireOfficialRendition creates no duplicate bytes/job
```

No generic production response-buffer validator is added. Generated typed output plus targeted contract tests remain the accepted minimum. Actual runtime execution of these fixtures belongs to the later validation/implementation program once a runtime exists.

External evidence checked during T8-E includes OpenAPI 3.0.3, RFC9110, RFC9457, RFC9530, Fetch forbidden-header behavior, OWASP archive/upload resource controls, current AWS S3 PutObject/presigning behavior, controlled-document upload limits, Stripe/Adyen idempotency practice, and current `oapi-codegen` / `openapi-typescript` behavior.'''
replace_section(
    "# 9. Generation / runtime conformance proof",
    "# 10. Structural Inversion / subtractive checkpoint",
    section9,
)

section11 = r'''# 11. Remaining closure gate

The measurement, generated-boundary feasibility, provider presign feasibility and ledger-census fixture obligations are closed at candidate level.

Remaining Lead gate:

```text
A. run one final whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence attack
B. close every surviving Lead finding without speculative capability
C. revalidate exact candidate HEAD + intended 5-file durable/work diff + required CI
D. only if A→C converge, create review/t8e-fable from that exact candidate HEAD
E. independent Fable challenge
F. Lead adjudication of Fable evidence
G. explicit operator ratification
```

Until A→C converge:

```text
T8-E ACTIVE
T8-F NOT OPEN
implementation BLOCKED
Fable NOT STARTED
```'''
i = s.find("# 11. Remaining closure gate")
if i < 0:
    raise SystemExit("section 11 missing")
s = s[:i] + section11.rstrip() + "\n"

p.write_text(s)
