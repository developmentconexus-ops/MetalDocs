# ADR 0073 — Remove `/documents/{id}/finalize`; canonical `/submit` owns in-tx prereq resolution

> **Status:** Accepted 2026-07-06
> **Scope:** Post-GMR live-QA remediation — draft→under_review submission contract.
> **Supersedes:** the CON-01/DEC-01 decision to *retain* `/finalize` as a deprecated
> convenience wrapper (the DEC-01 "submit is canonical" ruling itself stands and is
> completed by this ADR).

## Context

Live end-to-end QA (2026-07-06, docker production stack) found that "Submeter para
revisão" on a freshly created document failed with 400 `VALIDATION_ERROR`. Root cause
chain, established from runtime + code evidence:

1. A fresh draft rests at `revision_version = 0`. Autosave commit
   (`repository.CommitUpload`) deliberately never increments the OCC counter — only
   the submit transition does (`revision_version = revision_version + 1`).
2. The editor submitted via the deprecated wrapper `POST /documents/{id}/finalize`,
   sending `If-Match: "v0"`.
3. `parseFinalizeIfMatch` (wrapper) rejected `v0` and did not accept `*`; the
   canonical `/submit` parser (`parseIfMatch`) accepted `*` but also rejected a
   literal `v0`. **No submit entrypoint accepted the true OCC state of a fresh
   draft.** The service layer itself accepts `RevisionVersion: 0` (integration
   tests exercise it).
4. The FE could not migrate to canonical `/submit` because that contract required
   `route_id` + `content_hash`, which an author cannot supply: route resolution
   needs approval-route admin reads the author has no capability for, and the
   content hash is server-authoritative state (autosave commit computes it).
   `/finalize` existed solely to resolve both server-side (`GetFinalizePrereqs`)
   — plus a bespoke copy of idempotency replay — and it resolved them **off-tx**,
   leaving a TOCTOU window between the prereq read and the submit transaction.

Retaining a deprecated wrapper whose only substance is prereq resolution + a second
idempotency implementation is a local maximum; the global maximum is one canonical
endpoint that owns the complete use-case.

## Decision

1. **`POST /documents/{id}/finalize` is removed** from the OpenAPI spec, generated
   code, handler, its private If-Match parser, its bespoke idempotency store wiring,
   the direct-mux registration/`skippingMux` exception, and the
   `GetFinalizePrereqs` chain (handler port → application → repository →
   `domain.FinalizePrereqs`). The domain sentinels (`ErrDocumentNotDraft`,
   `ErrProfileNotConfigured`, `ErrApprovalRouteMissing`) survive — the submit
   service now returns them.
2. **`POST /documents/{id}/submit` is the sole submit entrypoint** and gains in-tx
   server-side resolution:
   - `route_id` optional — when omitted, the service resolves the single active
     approval route for the document's controlled-document profile **inside the
     submit transaction** (documents row → `CDFieldReader.ProfileCode` port →
     `approval_routes WHERE active`), closing the wrapper-era TOCTOU.
   - `content_hash` optional — when omitted, the service binds the head
     revision's stored content hash in the same transaction.
   - `revision_title` added to the request body (was finalize-only); REV>=1
     requiredness stays enforced in the service against the in-tx governed
     revision number.
   - Explicit client-supplied values keep their exact prior semantics.
   Resolution reads are plain non-recording SELECTs (HS-PRE-1 respected); the
   narrow `SubmitDefaultsResolver` port keeps the big `ApprovalRepository`
   interface untouched (ISP).
3. **`parseIfMatch` accepts `"v0"`** (`version < 0` is malformed, not `<= 0`).
   The parser's domain now equals the OCC counter's domain, which starts at 0.
   `*` keeps meaning "expected version 0" as before.
4. **FE editor calls canonical `/submit`** with `If-Match: "v<N>"` (N ≥ 0),
   `Idempotency-Key`, and body `{revision_title?}` — no route/hash echo.

## Consequences

- One submit code path, one idempotency mechanism (platform middleware +
  `approval_instances` UNIQUE backstop, F-D4); the duplicated per-handler replay
  store copy dies with the wrapper.
- First-submit works from the true draft state (`v0`) without a wildcard footgun;
  OCC semantics are uniform across approval endpoints.
- Clients that *can* assert `route_id`/`content_hash` (future integrations,
  signoff-style content binding) still may — optionality is additive, not a
  weakening: the enforced integrity guard for submit remains the atomic
  `WHERE status='draft' AND revision_version=$N` CAS.
- Breaking change for any external `/finalize` caller: none exist (FE was the
  only consumer; pre-v1 product, no external API consumers).

## References

- REQ: `wiki/architecture/backend-target-architecture.md` (contract-first,
  fixed lifecycle, DB-enforced invariants).
- ADR 0072 (`documents/approval` nested exception — sanctions the approval
  repo's in-module reads used by the resolver).
- Grade-A CON-01/DEC-01 notes (submit canonicalization; wrapper retention now
  superseded).
- Live QA evidence: document `48e00392-8718-457f-bc3e-4e48b425c99a` (PO-RH-001),
  400 on finalize with `If-Match: "v0"`, 2026-07-06.
