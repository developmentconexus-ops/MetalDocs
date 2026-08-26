---
id: t11-p11-acceptance-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for the operator-accepted FP2/P11 integrated low-fidelity product Evidence preserved outside the merge candidate.
---

# T11 FP2/P11 Acceptance Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.
> **Scope:** FP2/P11 — operator-accepted integrated low-fidelity product assembly only.
> **T11:** remains OPEN.
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists

Merge candidates contain no `docs/work/**`, while later P12 adversarial walkthrough and P13/P14 conformance require the exact operator-accepted assembled product. The complete P11 tree is preserved under one immutable Evidence ref before temporary work leaves the acceptance candidate.

This locator routes Evidence only. It is not Product authority, a second roadmap or permission to reopen a LOCK.

## 2. Preserved Git identity

```text
repository   developmentconexus-ops/MetalDocs
source       claude/repo-context-technical-design-69t84i
evidence ref evidence/t11-p11-r1-accept-proofs-20260826
exact commit 4846ad445e251d02673bda6b54fac461d3548e54
```

The remote ref resolves to the exact commit above. The earlier ref `evidence/t11-p11-r1-accept-20260826` (commit a1b3a3d) is superseded provenance of the same acceptance without the §17 negative-flow proof record. It MUST NOT move while current T11/P12/P13/P14 proof depends on P11.

## 3. Canonical P11 Evidence

| Evidence | Path on exact Evidence commit | Git id |
|---|---|---|
| Integration shell (global IA nav, hash deep links, journey guide) | `docs/work/current/p11/t11-p11-integrated-product.html` | blob `748b87ee91908e9dd95feaa5f823e5fa6958568f` |
| Assembled LOCKED block artifacts (byte-exact copies of B01–B12 P8 Evidence) | `docs/work/current/p11/blocks/` | tree `4cb8327afbd70183371d9221febc025b2b66f091` |
| P11 planning, assembly identity, S1 adjudication, operator acceptance + cross-block negative/recovery proof (method §17) | `docs/work/current/p11/t11-p11-planning.md` | blob `d563020a4b7194d7665b586bdb5a6e74d558e2e7` |

The 16 files under `blocks/` are unmodified byte-exact copies of the LOCKED P8 Evidence already routed by `t11-b01-b09-lock-evidence.md`, `t11-b10-lock-evidence.md`, `t11-b11-lock-evidence.md` and `t11-b12-lock-evidence.md`; the block locators remain the canonical block authority.

## 4. Protected P11 meaning

```text
one integrated low-fi product assembling all 13 LOCKED blocks byte-exact (no block redesign)
global navigation carrying the accepted B01 IA groups
per-route hash deep links that survive reload
six operator-testable cross-block journeys (consumption, authoring, approval,
  document-governance administration, access/organization, evidence)
13/13 routes proven in headless Chromium (render, nav state, deep-link reload, zero errors)
eight cross-block negative/recovery flows proven through the integrator (denial panel + global-nav
  recovery + fresh re-entry, 404 non-disclosure with recovery, continuation failure then cross-block
  navigation, unknown-deep-link fallback, deep-link reload into admin surface)
S1 adjudicated by the operator: in-block boundary terminators remain the LOCKED Evidence
  terminators; cross-block traversal is the integrator's global navigation (no per-block shim)
operator acceptance: R1, 2026-08-26
```
