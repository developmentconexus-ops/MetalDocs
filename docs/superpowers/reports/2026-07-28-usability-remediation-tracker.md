# Usability Remediation Tracker — 2026-07-28

Objetivo do operador: sistema **usável de ponta a ponta** pensando como usuário
que começa do zero. Foco: funcionamento + código. Polish visual fica para
depois. Operação: Claude orquestra, subagentes implementam, Codex (GPT-5.6 Sol)
revisa decisões estruturais; passo a passo, área por área.

## Decisões de produto ratificadas (operador, 2026-07-28)

| ID | Decisão |
|---|---|
| D1 | **Config-first**: ordem canônica de onboarding = áreas → perfis → rotas de aprovação → templates/documentos. Sistema guia essa ordem. |
| D2 | **Gate de criação (hard-block)**: perfil sem rota ativa = INCOMPLETO — não selecionável no wizard de documento/template; badge + link "configurar rota". Sem rota standard/fallback (fail closed). |
| D3 | **Publicar = capability**: gate da UI = `approved` + `document.publish`; backend expõe contexto de publish (content_hash/If-Match) a quem tem a capability, sem exigir membership de área. Sem capability, botão não renderiza. |
| D4 | **PDF viewer embutido** na tela; download é ação separada. |
| D5 | **Push**: autorizado para o fim de 2026-07-28 se build+testes+QA verdes. |
| D6 | **Rota de templates na configuração** (2026-07-28): re-modelar rota de aprovação de template para a fase de config (nível tenant/perfil, ANTES de existir template); criação de template hard-blocked sem ela, igual documentos. ADR necessário. Substitui o modelo atual keyed por template_id. |
| D3′ | **D3 SUSPENSO para re-exame**: operador questionou a existência do passo publish pós-aprovado ("no momento que foi aprovado já não era para estar publicado?… isso vem desde o legacy"). Verificar prática de mercado (eQMS) + pipeline atual → proposta. Investigação em andamento. |

## Fila de etapas (executar em ordem; 1 etapa = 1 ciclo implementa→revisa→verifica→commit)

### Etapa 1 — Onboarding/configuração operável (D1+D2)
| Item | Defeito | Forma global-máxima |
|---|---|---|
| 1.1 | F-QA4-3 submit 409 sem rota (silencioso) | Conceito **readiness de perfil** no backend: perfil operável ⇔ tem rota de aprovação ativa. Exposto no contrato (lista de perfis/áreas elegíveis p/ criação). Wizard consome elegibilidade, não `taxonomy.view` cru. |
| 1.2 | F-QA4-1 preview-code 403 engolido (`it-usinagem-???`) | Dropdown de áreas por **elegibilidade** (capability×área do usuário), erro de preview surfaceado com estado explícito. |
| 1.3 | Tela taxonomia não sinaliza perfil incompleto | Badge INCOMPLETO + link para rotas; contagem de rotas ativas por perfil no contrato de listagem. |
| 1.4 | F-QA4-12 nav: "Auditoria" morto; /admin/* sem entrada | Grupo "Administração" no rail filtrado por capability; item Auditoria removido ou apontando para algo real. |

### Etapa 2 — Publicação + artefato (D3+D4)
| Item | Defeito | Forma global-máxima |
|---|---|---|
| 2.1 | F-QA4-5 gate invertido (admin bloqueado, autor sem capability vê botão vivo no-op) | Gate = `approved` + `document.publish`. Contexto de publish legível por capability holder (endpoint ou campo na ficha). Ação não renderiza sem capability (padrão da tela de Distribuição). |
| 2.2 | Visualizar PDF só baixa | Viewer embutido (painel/modal) com o `final.pdf`/export; download separado. |
| 2.3 | F-QA4-11 `values_frozen_at` sempre null | **CORRIGIDO** `4bf6b154`. |
| 2.4 | F-QA4-7 `signoff_id` vazio | Devolver o id persistido. |
| 2.5 | F-QA4-6 Idempotency-Key inconsistente | **Decidido (Codex Q4)**: UUID em toda rota idempotente; spec `format: uuid`; validador compartilhado único (middleware), nunca regra por handler. |

### Etapa 3 — Editor/documento
| Item | Defeito | Forma global-máxima |
|---|---|---|
| 3.1 | F-QA4-9 autosave/commit 500 (`form_data_snapshot` opcional⊥NOT NULL) | Contrato e escrita coerentes: default `{}` na escrita + 400 nunca 500. |
| 3.2 | F-QA4-4 painel IDENTIFICAÇÃO `---` | Ligar snapshots já devolvidos pela API ao painel (dado, não estética). |
| 3.3 | F-QA4-8 inbox mostra UUID; `?` em próximos aprovadores | Card usa `code`; resolver elegível pré-decisão. |

### Etapa 4 — Template do zero (jornada nunca testada de ponta a ponta)
| Item | Escopo |
|---|---|
| 4.1 | QA browser: criar template do zero (com e sem placeholder) → aprovar → usar no wizard → documento a partir dele. Gate D2 vale para template também. |
| 4.2 | Corrigir o que a jornada revelar (mesma disciplina). |

### Etapa 5 — Higiene/integridade
| Item | Escopo |
|---|---|
| 5.1 | F-QA4-10 renomear `content_hash`→`values_hash` nos outboxes (ou gravar hash real). |
| 5.2 | Lineage hashes revisão→frozen.docx→PDF (resto da opção (a) do F-QA3-1). |
| 5.3 | Expurgo dos artefatos congelados inválidos pré-fix (`ba24c4f2…`, `45c9e784…`, `d18fbfdf…`). |
| 5.4 | F-QA4-2 enums de role hand-synced → fonte única do contrato. |

## Estado

| Etapa | Estado | Evidência |
|---|---|---|
| 1 | EM ANDAMENTO | — |
| 2 | FILA | — |
| 3 | FILA | — |
| 4 | FILA | — |
| 5 | FILA | — |

Log de execução no fim deste arquivo; achados novos continuam em
`2026-07-24-qa4-browser-qa-evidence.md`.

## Protocolo

- Cada etapa: investigar → desenhar (global máx) → **Codex revisa** → subagentes
  sonnet implementam → build+vet+testes → QA live no stack :80 → commit.
- Decisão de lógica nova que não coberta por D1–D5 → perguntar ao operador
  (que responde à distância); enquanto isso, avançar no que não depende.
- Push: só ao fim do dia (D5), tudo verde.

## Veredito Codex (task-ms4j2a0w-xglo2o, GPT-5.6 Sol xhigh) — reconciliado

- **Q1 AGREE**: creation-context é read model do controlleddocuments; approval expõe
  interface publicada bulk (nunca SQL); re-check de rota **dentro da tx de criação**;
  áreas devolvidas já autorizadas server-side (nunca lista crua p/ cliente filtrar);
  anotação em taxonomy = badge admin apenas, não autoridade do wizard; regra tier-1
  específica `GET /controlled-documents/creation-context → controlled_documents.create`
  ANTES da regra genérica de prefixo (senão herda `document.view`).
- **Q2 DISAGREE parcial com D3**: `document.publish` é **area-grade** hoje
  (`iam/domain/capability_scope.go:43`). Endpoint próprio `publish-context` é o seam
  certo, mas ou avalia capability×área in-tx (bypass system_admin explica o caso QA)
  ou D3 reclassifica publish como tenant-grade via ADR. → **pergunta ao operador**.
- **Q3 reordenação**: fatia atual ganha F-QA4-9 (500) e UX explícita do 409
  `approval_route_missing` no submit; decisão de semântica D2-template antes de fechar
  Etapa 1; **expurgo/quarentena dos artefatos congelados pré-fix = gate de release
  antes do push D5** (não higiene tardia). 2.3 já estava corrigido (tabela ajustada).
- **Q4 AGREE**: UUID everywhere (ver 2.5).
- **Q5 traps**: (a) rota de template keyed por `template_id`
  (`template_submit_service.go:194`) criada DEPOIS do template — D2 literal não se
  aplica a template sem re-modelar → **pergunta ao operador**; (b) contexts de
  leitura são advisory: mutação re-valida invariante in-tx sempre (já no desenho).

## Etapa 1 — desenho (ratificado com Codex)

Fatos mapeados (investigador):
- Rota ativa por perfil: `approval` module, `approval_routes(tenant_id, subject_kind, subject_key, active)`;
  `LoadActiveRouteIDByProfile` (`postgres_approval_repository.go:2018`, subject_kind=`document`, subject_key=profile code).
- `state.approval_route_missing` levantado em `submit_service.go:407-437` → 409 (`http/errors.go:84`).
- `GET /documents/{id}/approval-preview` já devolve `route_id: null` sem rota (informacional).
- preview-code com gate assimétrico: tier-1 `document.view` (prefixo), tier-2 `controlled_documents.create` (`service.go:575`).
- Template create SEM checagem de rota (`templates/application/create.go:38-100`, ADR 0082).
- Wizard consome listas cruas de taxonomia (`useProfilesQuery`/`useAreasQuery` → `taxonomy.view`).

Forma proposta:
1. **Interface publicada no módulo approval**: leitura bulk "quais subject_keys têm rota ativa"
   (`subject_kind='document'`) — consumível por outros módulos via service, nunca via SQL.
2. **Hard gate no backend (D2)**: `atomicCreateControlledDocument` falha fechado 409
   `state.approval_route_missing` quando o perfil não tem rota ativa. Mesmo problem+json do submit.
3. **Contrato de contexto de criação**: endpoint composto no módulo controlleddocuments
   (`GET /controlled-documents/creation-context`): perfis anotados `has_active_route` + áreas anotadas
   `eligible` (capability×área do chamador). Wizard passa a consumir isso (não mais listas cruas).
4. **Tier-1 do preview-code** alinhado a `controlled_documents.create` (simetria com tier-2).
5. **Taxonomia admin**: listagem de perfis anotada com `has_active_route` → badge INCOMPLETO + link rotas.
6. **Nav**: grupo Administração por capability; item Auditoria morto removido (F-QA4-12 — **FEITO** `41c5e892`).
7. Erro de preview-code surfaceado no banner (estado de erro explícito, sem `???` ambíguo).
8. (Q3) F-QA4-9 entra nesta fatia: snapshot ausente → **preserva** form_data existente
   (partial update, sem wipe, sem default substituto); inválido → 400; nunca 500.
9. (Q3) UX explícita no submit para 409 `state.approval_route_missing` (frontend, junto do item 7).

Perguntas abertas ao operador: escopo de `document.publish` (tenant-grade via ADR vs
area-grade avaliado no publish-context); modelo de gate D2 para templates.

## Log

- 2026-07-28: tracker criado; D1–D5 ratificadas; Etapa 1 iniciada.
- 2026-07-28: F-QA4-11 corrigido (`4bf6b154`) — values_frozen_at em GetDocument + lista paginada.
- 2026-07-28: modelo de execução ajustado pelo operador: Fable/Sol-xhigh orquestram+revisam; código = Opus med / Sol low.
- 2026-07-28: F-QA4-12 (nav) CORRIGIDO `41c5e892` — grupo Administração por capability
  (Taxonomia/Membros/Rotas), Auditoria morta removida; typecheck 0, shell 14/14.
- 2026-07-28: veredito Codex recebido e reconciliado (seção acima); desenho Etapa 1 ratificado.
- 2026-07-28: F-QA4-9 CORRIGIDO `06d0d17e` — commit de autosave preserva form_data quando
  snapshot ausente (partial update, sem `{}` substituto); 400 p/ não-objeto; 6 testes novos.
- 2026-07-28: Etapa 1 backend CORRIGIDO `90a6fae3` — gate D2 in-tx (409
  state.approval_route_missing na criação), interface publicada RouteReadinessReader,
  GET /controlled-documents/creation-context (áreas narrowed server-side via novo
  iam AreaCapabilityReader), permissions tier-1 creation-context+preview-code →
  controlled_documents.create (F-QA4-1), has_active_route no listTaxonomyProfiles.
  37 testes novos; build/vet/suites verdes. Frontend (wizard+badge+erros) em implementação.
- 2026-07-28: operador ratificou D6 (rota de template na configuração) e suspendeu D3
  para re-exame (publish pós-aprovado pode ser legacy) → investigação do pipeline
  de publish em andamento; Etapa 2 re-desenha depois da resposta.
