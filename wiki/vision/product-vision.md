# Product Vision

> **Last verified:** 2026-05-01
> **Status:** Stub. Expand when product brief is finalized.
> **Scope:** What MetalDocs is, the problem it solves, why it exists.

## One-liner

MetalDocs is an ISO-bound controlled-document platform for organizations that need versioned, signed, immutable operational documents (procedures, job descriptions, policies) with full audit trail and PDF distribution.

## Problem

ISO 9001 / 14001 / 45001 / 27001-bound organizations must:

- Maintain controlled versions of every operational document.
- Enforce review + approval workflows with segregation of duties.
- Freeze approved versions so they cannot be silently mutated.
- Distribute the canonical PDF to operators who execute the procedure.
- Prove compliance during audits (who approved what, when, with what content).

Generic editors (Word + SharePoint, Google Docs) lack:

- Built-in revision counters tied to controlled-document codes.
- Approval routes with quorum and ISO segregation enforcement.
- Cryptographic freeze (hashes prove the artifact is the one approved).
- Token substitution (`{doc_code}`, `{author}`, `{effective_date}`) auto-resolved at freeze.

## What MetalDocs does

- Templates with the 7 fixed tokens; templates are versioned, approved, published.
- Document profiles (Tipos Documentais) bind a category to a template.
- Controlled documents get auto-generated codes per (profile, area) sequence.
- Documents go through `draft → under_review → approved → frozen` with signoffs.
- Freeze resolves tokens, computes content/values/schema hashes, stores immutable DOCX.
- Async fanout converts to PDF for distribution.

## What MetalDocs is NOT

- A general-purpose word processor (eigenpal handles editing; we layer governance).
- A DMS / file-share replacement.
- A workflow-engine-for-everything (approval routes are scoped to controlled docs).

## See also

- [vision/target-users.md](target-users.md)
- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — how users actually experience this
- [decisions/0008-placeholder-fixed-catalog.md](../decisions/0008-placeholder-fixed-catalog.md) — why fixed 7-token catalog
