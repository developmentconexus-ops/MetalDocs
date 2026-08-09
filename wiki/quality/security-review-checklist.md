# Security Review Checklist (OWASP ASVS)

> **Last verified:** 2026-08-08
> **Scope:** Operationalizes REQ-SEC-3 (`wiki/architecture/backend-target-architecture.md` §9) — "OWASP ASVS is the review checklist for any change touching auth, input handling, file paths, crypto, or queries." Gives a reviewer a concrete ASVS section per trigger area and one answer to "does this change touch a trigger area?"
> **Evidence for REQ-SEC-3:** `kind: commit` in `wiki/architecture/req-trace-map.yaml`, pointing at the commit that adds this file and its `.github/PULL_REQUEST_TEMPLATE.md` wiring — see that entry's `note` for why this is process evidence, not an executable assertion.

## Status: transitional — this is a prompt, not a gate

**This checklist is enforced by a reviewer choosing to fill it in.** A PR template block can be ticked without the boxes being true; nothing here mechanically verifies that ASVS was actually consulted. That is a known, accepted limitation of this artifact, not an oversight — see "Global maximum this defers" below.

## When a change is in scope

A change is in the **five trigger areas** if it does any of the following. This list is the trigger's operable definition — "does this change touch crypto?" has exactly one answer: check this list against the diff.

| Trigger area | Triggers when the diff... |
|---|---|
| **Auth** | touches `internal/modules/auth/**`, `internal/platform/authn/**`, session/token issuance, verification, revocation, password hashing/verification, or login/lockout logic |
| **Input handling** | adds or changes a request-body/query/path parameter parser, a validation rule, a file-upload handler, or anything that deserializes external input before it reaches an application service |
| **File paths** | constructs, joins, or resolves a filesystem or blob-store path/key from any value that originates outside the process (user input, another service's response, a DB row not written by this process in this same transaction) |
| **Crypto** | touches `internal/platform/passwordhash/**`, `internal/platform/idempotency/**`'s key derivation, any `crypto/*` import, HMAC/signing code (e.g. `internal/modules/auth/application/service.go`'s `signToken`/`hashToken`), tenant DEK handling, or TLS/cert configuration |
| **Queries** | adds or changes a hand-built SQL string, a new call site passing external input into a query (parameterized or not), or anything touching `internal/platform/sqlescape` |

If none of the five apply, this checklist does not apply and the PR template's checklist block may be marked N/A.

## ASVS sections by trigger area

[OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) (this checklist targets ASVS 4.0.3, the version current at time of writing) organizes controls into numbered chapters. Run the listed chapters for each trigger area the diff hits — not the whole standard, so the checklist stays runnable in review time.

| Trigger area | Run these ASVS chapters | Focus |
|---|---|---|
| **Auth** | V2 (Authentication), V3 (Session Management) | Credential storage (memory-hard KDF, REQ-AUTHN-1), session token opacity/entropy/revocation (REQ-AUTHN-3, ADR 0094), lockout, re-auth for sensitive ops (REQ-AUTHN-2) |
| **Input handling** | V5 (Validation, Sanitization and Encoding) | Boundary validation against the contract schema (REQ-API-6), output encoding, deserialization safety |
| **File paths** | V12 (File and Resources) | Path traversal, tenant-namespaced blob keys (REQ-TEN-4), upload type/size constraints |
| **Crypto** | V6 (Stored Cryptography), V9 (Communications) | Algorithm choice and parameters (e.g. Argon2id cost params), key/secret handling (REQ-SEC-1), constant-time comparison, TLS configuration |
| **Queries** | V5.3 (Output Encoding and Injection Prevention) | Parameterization (REQ-DATA-3), `sqlescape` usage for identifiers that can't be parameters, tenant-scope predicate present in every query (REQ-TEN-1) |

## Checklist (fill in for each triggered area)

For every trigger area the diff hits:

- [ ] the relevant ASVS chapter(s) above were read against this diff, not just against memory of the standard
- [ ] any deviation from a Level-1 ASVS control is either fixed or named as an explicit, linked exception
- [ ] the change does not introduce a second, competing implementation of a control this repo already has one canonical implementation of (e.g. a new hashing call site, a new ad-hoc SQL builder) — reuse the existing platform package instead
- [ ] tests exist for the security-relevant behavior, not only the happy path

## Evidence expectation

Record in the PR body (the template's checklist block):

- which of the five trigger areas applied
- which ASVS chapters were checked
- any accepted deviation, with a one-line reason and a link if it needs its own tracking

## Global maximum this defers

**Named structure:** a CI check that parses the diff for the five trigger-area signals above (same patterns as the table) and, when any signal fires, requires the PR body to contain a completed checklist block matching this file's format — a red build on a missing or unfilled block, not a request. This turns "reviewer prompt" into "mechanically enforced gate," closing the gap named in Status above.

**Why it is not built now:** the required-checks closure for this branch is pinned by `scripts/required-gate.jq` against a live ruleset that still requires 21 legacy contexts with `bypass_actors: []`. Adding a new required job here changes the aggregator's `needs:` closure before the Phase 4 ruleset swap, which can deadlock every open PR against the old ruleset. This checklist ships unenforced, deliberately, until that swap lands.

**Milestone that replaces the prompt with the gate:** the CI-restructure branch's Phase 4 ruleset swap (the point at which the required-checks aggregator's context set is renegotiated against a fresh branch-protection ruleset). Track the CI-check build as a named follow-up item at that milestone, not before.
