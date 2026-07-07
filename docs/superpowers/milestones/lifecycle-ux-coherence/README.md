# Program — Lifecycle & UX Coherence

**Governing spec:** `docs/superpowers/specs/2026-07-06-lifecycle-ux-coherence-design.md`
**Origin:** post-GMR live E2E QA; 23-finding gap register from 3 read-only investigations.
**Standard:** global-maximum, YAGNI, one-implementation-N-entry-points (spec §1 R1–R5).

## Milestones

| # | Milestone | Objective | Status |
|---|---|---|---|
| M1 | `milestone-1-canonical-submit-backend` | Author submit (REV0 + REV≥1) succeeds on canonical /submit with in-tx server resolution; finalize chain deleted; idempotency map complete (findings 1–5, 16–17) | validated — pending HS-1 |
| M2 | `milestone-2-fe-surface-ownership` | One submit implementation per artifact kind, author surfaces only; cockpit approver-only (findings 6–8, 13–14) | planned |
| M3 | `milestone-3-journey-closure` | Deep links close every journey (cockpit↔detail, notifications, fanout CTA); dead FE affordances deleted (findings 9–12, 20) | planned |
| M4 | `milestone-4-template-inbox` | Template reviews visible in the single approver worklist, contract-first (finding 15) | planned |

## Deferred register (bounded, with triggers)

| Finding | Defer trigger |
|---|---|
| 18, 19 (ETag return uniformity / route-admin domain) | first external API consumer chaining mutations |
| 21 (template publish/archive endpoints unused) | product decision on archive/direct-publish UI |
| 22 (bespoke idempotency stores) | next signoff/route-admin contract change |
| 23 (template contract-level OCC) | concurrent-editing complaints on templates |

## Pre-existing work absorbed (dirty tree at program start)

OpenAPI /finalize removal + /submit extension, oapi-codegen + FE api-types regen,
ADR 0073, parseIfMatch v0 fix, editor→/submit migration (vitest pending) — absorbed
by M1/M2, not reverted.

## Close-out

_(filled at program end)_
