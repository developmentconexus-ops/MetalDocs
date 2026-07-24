# Full QA — verdict consolidado (2026-07-24)

Escopo acordado com o operador: 6 etapas. 1–4 autônomas, 5 (browser fim-a-fim)
guiada, 6 = este verdicto. Stack: docker compose completo via gateway `:80`,
HEAD `21f52f29` + fix do `noopPresigner`.

## Resultado por etapa

| # | Etapa | Verdicto | Evidência |
|---|---|---|---|
| 1 | Gates estáticos (`go build ./...`, `go vet -tags integration`, lints de contrato) | PASS | build limpo |
| 2 | Suíte de integração backend (`scripts/test-integration.ps1`) | **FAIL** | 10+ testes, ver abaixo |
| 3 | Testes frontend (`vitest`) | **FAIL (1)** | 931/932 · 145/146 arquivos |
| 4 | Subida da stack Docker | PASS | todos os containers Healthy |
| 5 | QA de browser fim-a-fim "empresa do zero" | **FAIL** | [QA-4 evidence](2026-07-24-qa4-browser-qa-evidence.md) |
| 6 | Consolidação | este documento | — |

## Etapa 2 — integração

Falhas (amostra): `TestActiveUserAreasView_ParityWithBaseActiveNow`,
`TestRoleProvider_UserActiveInTenant_Live`,
`TestIntegration_SLASurfacer_FullTick_IteratesAllTenants`,
`TestIntegration_AuditValidator_P3_DetectsTamperedChain`,
`TestIntegration_Surfacer_*`, `TestIntegration_Janitor_P2_*`.

Padrão de duração é diagnóstico: o primeiro teste de cada pacote gasta ~160 s e
os seguintes morrem em exatos 15,00 s — assinatura de contenção/timeout no
provisionamento de banco (template testdb), não de asserção de produto. Precisa
de re-run isolado antes de classificar como regressão funcional; até lá fica
**INDETERMINADO com suspeita de infra de teste**, não green.

## Etapa 3 — frontend

Uma falha: `src/features/documents/routes.test.tsx` — "F2d.5 S3 —
single-artifact route flip > renders DocumentDetailLayout (record surface) at
/documents/:id/details". Escopo isolado, 931 testes passam.

## Etapa 5 — browser fim-a-fim

Jornada completa numa empresa criada do zero (família `operacoes` → perfil `it`
→ área `usinagem` → documento IT-USINAGEM-001 com conteúdo real → rota de
aprovação → submit → aprovação assinada → publicação → artefato).

Chegou ao fim, mas **o produto final sai em branco**. 8 achados novos
(F-QA4-1…-8) + reprodução dura do blocker F-QA3-1. Dois deles bloqueiam a
operação sem contornar por API (F-QA4-3 submit silencioso, F-QA4-5 publicar
inalcançável pela UI).

## Verdicto

**FULL QA: FAIL — HOLD mantido.**

O verdicto de release de 2026-07-23 (HOLD, 3 blockers) não só se sustenta como
ganha prova nova: F-QA3-1 foi reproduzido com conteúdo autoral real, e o
`content_hash` de materialização é literalmente o SHA-256 da string vazia. O
documento assinado e publicado — o artefato que vale legalmente num QMS — não
contém o texto aprovado.

Fila de correção:

1. **F-QA3-1** (blocker) — `FreezeService.Materialize` deve materializar
   `current_revision_id`, não o snapshot do template. Opção (a) já acordada.
2. **F-QA4-5** — gate de publish na UI derivado de recurso escopado por área.
3. **F-QA4-3** — submit engole `409 state.approval_route_missing`; perfil sem
   rota não é sinalizado na taxonomia.
4. **F-QA4-1** — preview-code engole `403`; dropdown de áreas não reflete
   capability×área.
5. Etapa 2 — re-run isolado da suíte de integração para separar infra de
   regressão; etapa 3 — corrigir `routes.test.tsx`.
6. F-QA4-2, -4, -6, -7, -8 — contrato/UX, sem bloqueio operacional.
