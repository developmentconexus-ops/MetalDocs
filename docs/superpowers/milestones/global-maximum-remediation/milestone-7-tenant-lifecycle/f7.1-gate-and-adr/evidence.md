# F7.1 — gate-and-adr · evidence

> **Feature:** F7.1 (developing-new-work gate + tenant-lifecycle ADR) · **Status:** CLOSED 2026-07-05
> **Acceptance (milestone.md):** gate artifact committed, verdict Green/Yellow; ADR accepted, cited by
> the milestone.

## Deliverables

| Artifact | Path | Commit |
|---|---|---|
| System-impact analysis (developing-new-work gate) | `docs/superpowers/analysis/2026-07-05-m7-tenant-lifecycle-system-impact.md` | `14c631c6` |
| ADR 0070 (tenant lifecycle — Accepted 2026-07-05) | `wiki/decisions/0070-tenant-lifecycle-onboarding-export-crypto-shred-erasure.md` | `d4a90bde` |

## Gate outcome
- **Verdict: 🟡 Yellow** — fits the architecture; no AS-1/AS-2/AS-3 hard-stop. ADR mandated (D7) and
  carries the decided erasure strategy. (Red would have been HS-8 → BLOCKED; not triggered.)
- Runtime-truth verified (5-anchor sweep): no onboarding API; 34 `tenant_id` tables; audit
  append-only + hash chain (PII in `payload`, skeleton survives); dev/CI role SUPERUSER+BYPASSRLS
  (owner-bypass trap); 33/34 FORCE RLS, `approval_signoffs` sole-trigger gap surfaced.

## Decisions locked (operator, 2026-07-05) → recorded in ADR 0070
1. Erasure = **crypto-shred per-tenant key + immutable audit skeleton** (ratified after explanation of the audit×GDPR tension).
2. Onboarding = **API route (contract-first)**.
3. F7.3 depth = **Implemented**.
4. Orchestrator home = **iam application service** (per-module `TenantDataPort` fan-out).
5. `approval_signoffs` = add real FORCE RLS + policy (no special-case).

## Review/QA disposition
Gate + ADR are documentation artifacts; no code. Self-reviewed against the developing-new-work
template (10 sections filled, feature-class module-only rows marked conditionally-N/A with reason) and
the ADR template (status/module/REQ/supersedes header + Context/Decision/Consequences/Validation).

## Bounded defers
None for F7.1.
