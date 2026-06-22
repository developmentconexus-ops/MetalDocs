# Feature F2.3 — Spec (distribuicao-wire)

> **Milestone:** 2 — Distribuição coverage-scope  ·  **Folder:** `f2.3-distribuicao-wire`
> **Status:** Approved (pre-code) — 2026-06-22 (operator standing approval via the re-decomposition; the consumer contract is the same one F2.1c froze and F2.2 now serves live).

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

> **Engine note:** F2.3 is a **pure consumer**. The producer is real (F2.2, gate green) and the FE
> types are generated and committed (`frontend/apps/web/src/lib/api-types/index.d.ts` —
> `DistributionSummaryResponse`, `DistributionRecipient`, `DistributionRecipientsResponse`,
> `DistributionAreaCoverage`). No contract is authored here. The interview below records the only open
> questions, which are **presentation/UX**, not data-contract.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Which surfaces wire to **live** data? | Exactly the three **denominator** surfaces the producer serves: **total** (`GET …/distribution` → `total_targets`), **recipient list** (`GET …/distribution/recipients` → `items[] {user_id, name, area_code\|null, area_name\|null, source}` + `page` cursor), **by-area coverage** (`GET …/distribution/coverage` → `{area_code, area_name, total}[]`). Nothing else has a real producer. |
| 2 | What happens to the **numerator** surfaces (read %, acknowledged %, overdue, pending, adoption timeline, per-recipient read/ack status & timestamps)? | Rendered as an explicit, labeled **"tracking pending"** state — *"Acompanhamento de leitura/ciência ainda não disponível"* — pointing at the parked mission. **Not** illustrative data, **not** a fabricated `0%`/`0 de N`, **not** the old watermark block. The honest empty-of-meaning state. |
| 3 | What is the precise copy for the tracking-pending state? | Portuguese, operator-locked tone: a short heading + one line that says reading/acknowledgement tracking is not yet recorded by the backend and is planned. No fake numbers anywhere inside it. Single shared presentational component so the grep for old literals stays at 0 and the disclosure reads identically across all numerator surfaces. |
| 4 | Recipient pagination UX? | Keyset cursor (producer is keyset on `(area_name NULLS LAST, user_id)`). FE consumes `page.next_cursor` / `page.has_more`. Default `limit` follows the contract default (20). UX: a **"Carregar mais"** affordance (append-on-fetch via `useInfiniteQuery`) — no page-number UI, since the producer is cursor-only and has no total count for the recipient list. The old status tabs (`Pendentes/Em atraso/...`) are **removed** (they are numerator filters with no producer). |
| 5 | What does each recipient **row** show now? | Denominator-only columns the producer actually returns: **name** (`name`, already COALESCE'd to `user_id` server-side), and an **origin** chip derived from `source` (`area_grant` → área `area_name`; `user_grant` → "Concessão direta"; `company_scope` → "Toda a empresa"). The read/ack **status** column, "last opened", and timestamps are dropped from the live list (numerator → tracking-pending note above the list, not per-row fabrication). |
| 6 | What happens to the 4 hero CTAs (Lembrete em massa, Exportar relatório, Adicionar destinatários, Política de fanout) and any row/bulk actions? | Stay **`aria-disabled` deferred-with-trigger**, `title`/note pointing at the parked mission (`wiki/backlog/document-distribution-mission.md`). Action layer is out of scope (HS-6). No behavior wired. |
| 7 | What is **removed at root** (not flag-hidden)? | `MOCK_DISTRIBUTION` and the mock interfaces it feeds; the `IllustrativeBlock` watermark wrapper and the `Dados ilustrativos · Em breve` literal; the `Em breve` hero badge over the denominator surfaces; the "todos os números… são ilustrativos" banner. Numerator mock data (`dailyReads`, `read`, `acknowledged`, `pending`, `overdue`, recipient `status`/`when`) is deleted, not repurposed. |
| 8 | Empty / loading / error states for the live surfaces? | Standard: skeleton/`role=status` while loading; honest empty ("Nenhum destinatário obrigatório" — a real zero is meaningful for company-scope-free docs); `role=alert` + retry on error (mirrors the existing `DocumentDistributionPage` doc-detail error state). A real `total_targets: 0` renders "0", not a placeholder. |
| 9 | Tenant / auth? | Same as every other screen — IAM cookies; the tier-1 route guard (`CapDistributionRead`) is enforced server-side by F2.2. A 403 surfaces the standard "sem permissão" error state; no client-side cap logic. |
| 10 | What must **not** change? | No backend change. No contract/type regen (types are frozen). No new shared primitive/token (HS-2). No touch to `useDocumentDetailQuery` semantics. No M0/M1 surface. No action-layer behavior. |

## Consumer contract (frozen — from generated types)

```ts
// frontend/apps/web/src/lib/api-types/index.d.ts
DistributionSummaryResponse  = { total_targets: number }
DistributionRecipient        = { user_id: string; name: string; area_code: string|null; area_name: string|null; source: "area_grant"|"user_grant"|"company_scope" }
DistributionRecipientsResponse = { items: DistributionRecipient[]; page: { next_cursor: string|null; has_more: boolean } }
DistributionAreaCoverage     = { area_code: string; area_name: string; total: number }   // GET coverage → DistributionAreaCoverage[]
```

Endpoints (all under `/api/v1`, IAM-cookie auth, tier-1 `CapDistributionRead`):
- `GET /documents/{id}/distribution` → `DistributionSummaryResponse`
- `GET /documents/{id}/distribution/recipients?cursor=&limit=` → `DistributionRecipientsResponse`
- `GET /documents/{id}/distribution/coverage` → `DistributionAreaCoverage[]`

The FE consumes the **generated** types only — no hand-rolled shapes (HS-3).

## Validation Gate (acceptance — each objectively checkable)

| # | Criterion | How proven |
|---|-----------|------------|
| V1 | **Illustrative literals gone at root** | `grep -nE "Dados ilustrativos\|MOCK_DISTRIBUTION\|Em breve" frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx` = **0**; `MOCK_DISTRIBUTION` deleted from `distributionMeta.ts` (grep over `src/features/documents` for `MOCK_DISTRIBUTION` = 0). |
| V2 | **Denominator renders live** | Runtime proof (preview): total, recipient list (name + origin chip), and by-area totals render values sourced from the real endpoints (consumer shape == generated types). Screenshot on record. |
| V3 | **Numerator is honest tracking-pending** | The read/ack donut, timeline, and per-recipient read/ack columns render the shared **"acompanhamento … ainda não disponível"** state — not a fabricated metric, not the old watermark. A grep proves no `0%`/fake-number fabrication was substituted. |
| V4 | **CTAs deferred-with-trigger** | The 4 hero CTAs remain `aria-disabled` with a trigger note referencing the parked mission; no action handler wired. |
| V5 | **Typed cursor pagination** | "Carregar mais" appends the next keyset page via `page.next_cursor`/`has_more`; no page-number UI; no skip/dup. |
| V6 | **Hook test green** | A query-hook / page test passes against fixtured responses matching the generated types (loading → data → empty → error; pagination append). |
| V7 | **Type safety** | `npm run typecheck` (web) clean for the touched files; the API layer references `components['schemas']['Distribution*']`, never inline duplicates. |
| V8 | **FE suite holds** | `make test` (or the web vitest suite) green at the operator-accepted baseline; no regression in `DocumentDistributionPage.test.tsx`. |
| V9 | **Both reviewers APPROVE on record (D2)** | `frontend-screen-reviewer` (visual/architectural parity) **APPROVE** + `frontend-code-reviewer` (code/maintainability) **APPROVE**, both recorded in `evidence.md`. |
| V10 | **No out-of-scope change** | `git diff` touches only `frontend/apps/web/src/features/documents/**` (+ `lib/queryKeys.ts`, the new api/query/test files); no backend, no contract regen, no shared primitive/token, no M0/M1 surface. |

## Non-goals (parked → `wiki/backlog/document-distribution-mission.md`)

Read/ack numerator producer & UI; adoption timeline; per-recipient read status; mass reminder; export;
add recipients; fanout policy; status-tab filtering of recipients. All render as honest
tracking-pending / deferred-with-trigger, never fabricated (HS-6).
