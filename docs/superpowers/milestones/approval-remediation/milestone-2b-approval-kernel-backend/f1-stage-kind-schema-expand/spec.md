# Feature F1 — Spec

> **Milestone:** 2b — Approval Kernel Backend  ·  **Folder:** `f1-stage-kind-schema-expand`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-07 — operator (via ratified governing spec §2.1/§7; this
> feature's contract is fully prescribed by the spec, no open design decision remains)

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — the governing spec (§2.1, §7) and the plan (F1) already fix the exact enum values (`review`/`approval`), field names (`StageKind`, `frozen_content_hash`, `cancel_reason`, `signature_meaning`), and placement (route-stage definition + instance-stage snapshot + instance + signoff). Baseline verification (this session) confirmed the real table is `approval_stage_instances`, not `approval_instance_stages` as the plan first assumed. No consumer-contract ambiguity remains to interview. | — |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** F2 (route versioning admin reads/writes stage kind on new versions), F3 (tier-2
  `authz.Require` will branch on stage kind), F4 (review-verdict service requires `StageKind ==
  review` at the domain layer), F5 (freeze executor reads stage kind to know when the boundary
  fires), F6 (reads `approval_instances.frozen_content_hash`), F7 (persists `signature_meaning`
  read from this column), F8 (reads `due_in_days`/`due_at`). All are later features in this same
  milestone — this feature is purely additive substrate they depend on.
- **Contract:**
  - `domain.StageKind` string type with exactly two values `"review"` and `"approval"`, plus a
    `Validate() error` method returning a sentinel `ErrInvalidStageKind` for anything else.
  - `approval_route_stages.stage_kind` (text, NOT NULL DEFAULT `'approval'`, CHECK IN
    `('review','approval')`) and `approval_route_stages.due_in_days` (integer, nullable, CHECK
    `due_in_days IS NULL OR due_in_days > 0`).
  - `approval_stage_instances.stage_kind` (text, NOT NULL DEFAULT `'approval'`, same CHECK) and
    `approval_stage_instances.due_at` (timestamptz, nullable).
  - `approval_instances.frozen_content_hash` (text, nullable, CHECK matches `^[0-9a-f]{64}$` when
    present) and `approval_instances.cancel_reason` (text, nullable).
  - `approval_signoffs.signature_meaning` (text, NOT NULL DEFAULT `'approval'`, CHECK IN
    `('approval','rejection')`).
  - Go struct fields on the route-stage struct (`domain/route.go`), the instance stage-snapshot
    struct (`domain/instance.go`), and the signoff struct (`domain/signoff.go`) mirroring each new
    column, at the placement the existing struct field ordering suggests (append after the
    corresponding existing snapshot/definition fields).
- **Source of truth for the contract:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md`
  §2.1/§7; runtime-truth corrections in `../milestone.md` (table is `approval_stage_instances`).

## What this feature implements

Expand-only migration `0286_approval_stage_kinds_expand.sql` adding the six columns above (all
NOT NULL-with-default or nullable — no backfill of existing rows required beyond the DEFAULT), plus
the `domain.StageKind` Go enum + `Validate()` + `ErrInvalidStageKind`, plus the corresponding struct
field additions on `RouteStage`, the instance stage-snapshot type, and `Signoff`.

## Non-goals (mandatory)

- No behavior change — `stage_kind` defaults to `'approval'` everywhere, meaning existing routes/
  instances behave exactly as before this migration. Wiring stage-kind semantics into services
  (review-verdict routing, freeze trigger, capability gating) is F3/F4/F5, not F1.
- No route versioning, no capability changes, no hash-chain changes — those are F2/F3/F6.
- No population of `frozen_content_hash` (that's F5) or `signature_meaning` derivation logic
  (that's F7) — F1 only adds the column and the DEFAULT.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|---|---|---|
| Go enum has exactly the two spec values and rejects unknown values | `go test ./internal/modules/documents/approval/domain/... -run TestStageKind -v` | real |
| DB CHECK rejects an unknown `stage_kind` on `approval_route_stages` and `approval_stage_instances` | `go test ./tests/integration/approval/... -run StageKind -v` (testdb factory, DB up) | real |
| Migration applies cleanly on top of `0285` and existing rows get the DEFAULT | same integration test, asserting a pre-existing seeded row now reads `stage_kind='approval'` | real |
| No regression | `go build ./...` clean; existing approval domain/application test suites still green | real |

## ADR needed?

- [x] No durable decision — skip (schema-expand only; the durable decisions — route versioning,
  freeze boundary, oversee capability, delegation — are recorded as ADRs in F2/F5(via F10 writeup)/F9).
