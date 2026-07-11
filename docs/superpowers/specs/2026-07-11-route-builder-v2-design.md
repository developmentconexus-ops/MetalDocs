# Route Builder v2 — design (unit 2.5)

**Status:** drafted 2026-07-11 by orchestrator session; scope LOCKED by chip.
**Owning surface:** frontend `features/approval/pages/route-admin` (consumes taxonomy `governance_class` G1, already merged on main).
**Governing spec:** `docs/superpowers/specs/2026-07-10-review-approval-workflow-model.md` (R1–R5).
**Design reference:** v2 mock file `route_builder_mock_v2` **does not exist in the repo** (only v1
`frontend/apps/web/design-source/route-admin/`). The ratified workflow spec §mock enumerates every
v2 element explicitly, so the SPEC is the design authority here; v1 design-source is the visual base
(Approval-slate palette, tokens). **Bounded deviation carried for HS-1.**

## Global-maximum judgment
The v1 route-admin data/mutation/transport layer is sound and tested — contract-first codegen types,
`mutate()` idempotency/If-Match/problem+json, optimistic create/update/deactivate, error mapping,
stable stage `uid` keys. The gap is purely **presentational + one domain concept** (`stage_kind` +
per-profile signature policy). So v2 = **evolve the presentational layer in place**, reuse the data
layer. Forking a parallel `*V2.tsx` that re-wires transport would duplicate tested infra = local
maximum / regression. The named next consumer (approval-remediation M4 ActorSelector) gets its
extension point via component decomposition, not a config flag (YAGNI-professional).

## Contract facts (verified)
- `openapi.yaml:5619` — `DocumentProfileItem.governance_class` enum `[controlado,simples,livre]`, **required**.
- FE generated types `src/lib/api-types/index.d.ts` are **stale** — `DocumentProfileItem` lacks
  `governance_class`. Slice 0 regenerates via `pnpm gen:api` (contract-first).
- `StageRequest.stage_kind?: "review" | "approval"` **already in generated types** (optional, defaults
  approval at persistence). v1 editor never exposes it → every stage is implicitly approval today.
- `features/taxonomy/api/taxonomy.ts:toDocumentProfile` **drops** governance_class (narrowing map).
  Must thread `governanceClass` through the camelCase `DocumentProfile` type + mapping.

## Domain rules (from workflow spec)
| governance_class | Route allowed? | Signature (approval-kind) stage | Builder UX |
|---|---|---|---|
| controlado | yes | **≥1 required** | badge "obrigatório ≥1 assinatura"; block save if 0 approval stages |
| simples | yes | optional (review-only OK) | badge "opcional" |
| livre | **no** | — | route creation blocked; friendly disabled state (backend also 409/422) |

- R1: route = review* → approval*, ≥1 stage total. `stage_kind` distinguishes.
- R2: multiple stages = sequential rounds; people in a round parallel w/ quorum (`any_1_of`/`all_of`/`m_of_n`).
- R4: author auto-excluded from every stage (SoD) → informational note in builder.
- R5: overlap (same person review+approval) → "Aprovar já" fast-forward → informational note.

## Slices (task board #1–#5)
- **S0** contract regen + governance_class threading (foundation; unblocks S1,S2).
- **S1** `routeGovernance.ts` — pure RoutePolicy derivation + stage_kind labels + controlado ≥1-approval validation.
- **S2** presentational blocks — `StageKindControl`, `QuorumPills`, `StageActorSlot` (M4 extension slot),
  `RouteFlowPreview`, `ApprovalPolicyBadge` + SoD/Aprovar-já notes. Design tokens only.
- **S3** assemble RouteEditorDialog v2 (evolve in place); wire stage_kind into StageRequest; livre block; controlado validation.
- **S4** independent reviews + ladder L0/L1 + browser QA :80 + evidence + commit.

## Constraints (non-negotiable)
Generated API types only (ADR 0035). Design tokens only — no new tokens, no inline `style=`. Never
read/print/commit `.env` (FE-only; `.env.development` present, no root secret needed). Commit when
verified; NEVER push. Subagent implements each slice; independent reviewer per slice (implementer ≠ reviewer).
