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
evidence ref evidence/t11-p11-r1-accept-20260826
exact commit a1b3a3d13d7ebaa2dd454df31110c3abc4c2d9f5
```

The remote ref resolves to the exact commit above. It MUST NOT move while current T11/P12/P13/P14 proof depends on P11.

## 3. Canonical P11 Evidence

| Evidence | Path on exact Evidence commit | Git id |
|---|---|---|
| Integration shell (global IA nav, hash deep links, journey guide) | `docs/work/current/p11/t11-p11-integrated-product.html` | blob `748b87ee91908e9dd95feaa5f823e5fa6958568f` |
| Assembled LOCKED block artifacts (byte-exact copies of B01–B12 P8 Evidence) | `docs/work/current/p11/blocks/` | tree `4cb8327afbd70183371d9221febc025b2b66f091` |
| P11 planning, assembly identity table, S1 adjudication + operator acceptance record | `docs/work/current/p11/t11-p11-planning.md` | blob `174fc627d86542e4074235f917e8d93d16b2edc4` |

The 16 files under `blocks/` are unmodified byte-exact copies of the LOCKED P8 Evidence already routed by `t11-b01-b09-lock-evidence.md`, `t11-b10-lock-evidence.md`, `t11-b11-lock-evidence.md` and `t11-b12-lock-evidence.md`; the block locators remain the canonical block authority.

## 4. Protected P11 meaning

```text
one integrated low-fi product assembling all 13 LOCKED blocks byte-exact (no block redesign)
global navigation carrying the accepted B01 IA groups
per-route hash deep links that survive reload
six operator-testable cross-block journeys (consumption, authoring, approval,
  document-governance administration, access/organization, evidence)
13/13 routes proven in headless Chromium (render, nav state, deep-link reload, zero errors)
S1 adjudicated by the operator: in-block boundary terminators remain the LOCKED Evidence
  terminators; cross-block traversal is the integrator's global navigation (no per-block shim)
operator acceptance: R1, 2026-08-26
```
