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
| D3′ | **D3 SUSPENSO para re-exame**: operador questionou a existência do passo publish pós-aprovado ("no momento que foi aprovado já não era para estar publicado?… isso vem desde o legacy"). Verificar prática de mercado (eQMS) + pipeline atual → proposta. **Superseded por D7.** |
| D7 | **Publish manual MORRE — "ADR + coordenador já"** (2026-07-28): aprovado ⇒ publicado (com gates de data efetiva/prontidão, alinhado a Qualio `effectiveOnApproval`/MasterControl/Veeva). Etapa 2 = ADR do **release coordinator** idempotente (fatos duráveis de outbox: aprovação + artefato pronto; predicado approval×artefato×data-efetiva×supersession-head; transição única via CAS) + implementação. Endpoint/botão/capability `document.publish` DELETADOS (não repropostos); `/supersede` removido ou re-desenhado. Plano de publicação (data efetiva + supersede) declarado na submissão. F-QA4-13/14 corrigidos dentro do redesenho. Push D5 leva só o que fechar verde. |

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
| 1 | **PASSED** | QA live :80 2026-07-28 — ver log |
| 2 | EM ANDAMENTO (D7: ADR release coordinator) | — |
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

## Etapa 4 — fatos D6 (investigador, 2026-07-28)

- Rota de template é keyed **por instância** (`subject_key = template_id`,
  `subject.go:47-57`), oposto do modelo documento (por perfil). Ninguém cria
  automaticamente: admin precisa POST /routes manual com subject_kind=template —
  e o FE route-builder **não suporta** subject_kind=template
  (`routeDraft.ts:12-16,114` exige profileCode) → hoje só via API crua.
- `templates_template.doc_type_code` já é a chave de classe taxonômica
  (== profile code) usada em SetDefaultTemplate + IsPublished
  (`profile_service.go:277-333`, `template_version_reader.go:15-62`) → alvo
  natural de re-key para config-first.
- Re-key toca: `route_admin_service.go:232-301` (branch template pula
  profile-FK+policy; CHECK 0297 força profile_code NULL em rota template),
  `template_submit_service.go:201` (lookup literal por template_id), decisão
  sobre herdar RoutePolicy do perfil, e FE novo.
- Elegibilidade de criação de documento = `DefaultTemplateVersionID` do perfil
  apontando p/ versão `published` com doc_type_code igual (`service.go:466`).
- Release coordinator é document-only por ADR 0085 (`release_coordinator.go:123-129`
  recusa non-document); publish de template segue `PublishTemplateVersion` direto
  (5 status draft→under_review→approved→published→obsolete, SoD no publish).
  D6 = governança de rota (keying), NÃO timing de release, salvo ampliação explícita.
- Instância de aprovação continua keyed por template_version_id independente do
  re-key da rota (two-level keying já existente).

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
- 2026-07-28: Etapa 1 frontend CORRIGIDO `cc175c01` — wizard consome creation-context
  como autoridade única de elegibilidade (catalogQueries só enriquecem display),
  card de perfil sem rota desabilitado + marcador "Incompleto — sem rota de aprovação"
  + link "Configurar rota" (gated `route.manage`), deep-link protegido, banner
  preview-code com estado de erro explícito + retry, mensagem PT-BR para 409
  `state.approval_route_missing`, pill INCOMPLETO na taxonomia admin.
- 2026-07-28: **operador ratificou D7** ("ADR + coordenador já") após análise
  Claude+Codex alinhada (task-ms4lce7v-ov1vmy: mercado = approval-driven effectiveness
  com gates explícitos; correções Codex: frozen_content_hash ≠ artifact-readiness,
  não existe fato durável "artefato materializado", rota review-verdict bypassa
  FreezeService.Pin). F-QA4-13 (effective_from gap → docs escapam review periódico)
  e F-QA4-14 (review-verdict pula materialização) registrados em
  `2026-07-24-qa4-browser-qa-evidence.md` (commit `2c9b3ee7`).
- 2026-07-28: **Etapa 1 QA live PASSED** (stack :80 rebuilt):
  - API: create sob perfil sem rota (`fmea`) → **409 `state.approval_route_missing`**;
    create sob `it` (com rota) → **201** (controle); creation-context 200 com
    `has_active_route` + áreas narrowed.
  - Browser: nav grupo Administração (Taxonomia/Membros/Rotas) visível, Auditoria
    morta ausente; wizard: card `fmea` disabled + "Incompleto — sem rota de aprovação"
    + link Configurar rota, clique não seleciona (aria-checked false).
  - Taxonomia admin aba Perfis: pill `INCOMPLETO` (title "Sem rota de aprovação
    ativa") + link "Configurar rota" apenas na linha `fmea`; `it`/`po` limpos.
    Verificado via probe DOM live (pane sem display p/ screenshot — operador remoto).
- 2026-07-28: 2.4 (signoff_id) + 2.5 (Idempotency-Key UUID) implementados por agente
  (opus): `SignoffResult.SignoffID` fresh+replay (envelope `SignoffReplay` ganha
  `signoff_id`, legacy decodável), spec `format: uuid` nos 27 params Idempotency-Key
  + regen completo, `ValidateKey` backstop (`IDEMPOTENCY_KEY_INVALID`), E2E keys →
  UUID; allowlist seed-chokepoint re-ancorada (drift de 06d0d17e). Commit pendente.
  Follow-up aberto: `review_verdict_handler.go:110` `VerdictID: ""` (mesma classe).
- 2026-07-28: 2.4/2.5 COMMITTED `123a56db`; api container rebuilt (imagem parcial
  descartada), smoke pós-restart 200. D4 viewer embutido COMMITTED `ff5d2cea` —
  modal blob+iframe, download separado, review 1🟡
  pré-existente (focus-trap classe codebase), FE suite 957/957.
- 2026-07-28: **ADR 0085 ACEITA, Codex ALIGN (rev 2)** — commit `2e9f9b68`.
  Rodada 1 do Codex: NOT ALIGN com 8 emendas obrigatórias (chave de geração
  durável compartilhada + fatos fail-closed; artifact-ready = DOCX+PDF em tx;
  tx de release completa com lock determinístico de alvos + rollback; split
  `planned_effective_from`/`effective_from` actual + emenda ADR 0069; backfill
  executável com Pin-repair; inventário hard-break completo + `document.supersede`
  retida re-homed p/ autorização na submissão; correções cross-ADR 0067/0069/0082
  + modelo 8-status/11-arcos; projeção de readiness + atribuição de ator +
  reconciliação River alert-only). Rev 2 incorporou tudo → ALIGN.
  Implementação em estágios a seguir (A: núcleo backend; B: retirement +
  contrato de submissão; C: backfill + projeção + FE).
- 2026-07-28: **Stage A (núcleo backend ADR 0085) IMPLEMENTADO** (~45 arquivos,
  working tree): migração `0310_release_coordinator.sql` (release_generations
  identidade 7-col UNIQUE + FORCE RLS; split `planned_effective_from` com data-move
  + CHECKs published/scheduled; `release_generation_id` nos dois outboxes de
  dispatch), fatos fail-closed em tx (`release_facts.go`), recorder terminal único
  em decision+review_verdict (`release_terminal_approval.go` — F-QA4-14
  estruturalmente irrepresentável), predicado puro (`domain/release.go`),
  coordinator idempotente CAS + lock ordenado + rollback supersede_conflict
  (`release_coordinator.go`), River `release-evaluate` (bypass authz de background),
  facts DOCX/PDF na tx de artefato (materialize/pdf job runners; caminho PDF legacy
  falha fechado com geração presente), threading generationID em eventos/fanout/
  dispatch (chaves idempotência legadas byte-idênticas). Verificação: ladder 114
  pkgs ok, 7 testes integração novos verdes, boundaries OK, api-lint 0.
  **Desvio aceito**: CHECKs de DB + correção dos caminhos legacy de escrita já em
  Stage A (F-QA4-13 fechado no nível DB); legacy morre em Stage B.
  **Defer bounded**: sync db-dictionary → próximo baseline fold.
  Reparo carona: `manual_code_create_integration_test.go` (quebrado pelo gate D2
  `90a6fae3`) — fixture semeia rota ativa (`testdb.NewApprovalRoute`) + wire
  `RouteReadinessReaderPG` real; pacote verde. Review Codex (Sol xhigh) do diff
  em curso; commit após veredito.
- 2026-07-28: **Stage A review Codex rodada 1 NOT-ALIGN → 4 fixes (opus)**, todos
  verificados reais por mim antes de agir: (1) blocker supersede legacy sem
  `effective_from` (CHECK 0310) → COALESCE(now()); (2) ordem de lock
  não-determinística (fonte lockada antes da descoberta de alvos;
  LoadCurrentPublishedHead com FOR UPDATE) → evaluateTx re-estruturado: geração
  FOR UPDATE primeiro, leitura da fonte SEM lock, lockAndRedecide locka
  {fonte}∪alvos ordenado 1-a-1 + re-decisão sob lock; novo
  LoadCurrentPublishedHeadNoLock; (3) AB-BA runners×coordinator →
  RecordArtifactFactTx ANTES da escrita em documents nos 2 runners
  (ordem global geração→documento); (4) evento governance `document_superseded`
  por alvo (constante única EventTypeDocumentSuperseded, meta-defeito de
  enumeração evitado). Suites approval+worker+integration verdes, boundaries OK.
- 2026-07-28: **rodada 2 NOT-ALIGN → 3 fixes (opus)**: (5) evento lifecycle do
  alvo cross-CD carregava CD da fonte → per-target ControlledID retido do lock;
  (6) legado SchedulePublish/RunScheduledPublishJob lockava head→fonte (AB-BA
  contra coordinator) → lockDocumentRowsInIDOrder compartilhado (descoberta
  lock-free + par ordenado + re-read autoritativo); (7) **blocker**: nenhum
  índice de head published único — legacy PublishApproved podia criar 2
  published no mesmo CD (Codex provou corrida coordinator-only inalcançável via
  ux_documents_cd_active) → 0310 §3b pre-repair idempotente + UNIQUE
  ux_documents_published_head (tenant_id, controlled_document_id) WHERE
  status='published'; MapPgError 23505→ErrStaleRevision (409 existente, zero
  churn de contrato). +3 testes integração (cross-CD, double-head, lock-order);
  fixture pré-existente com 2 heads reparada (estado agora irrepresentável).
  Residual aceito (morre Stage B): scheduler NULL-target pode bater no índice →
  River retry loop, não mapeado. Rodada 3 Codex em curso.
- 2026-07-28: **rodada 3 NOT-ALIGN → 2 fixes (opus)**: (8) **blocker**: 0310
  §3b pre-repair UPDATEia documents sem setar GUC `metaldocs.asserted_caps` →
  tripwire trg_require_cap_asserted P0001 em qualquer row legado (arm atual
  0301:93-104 aceita document.edit, match cap-only) → set_config tx-local
  '[{"cap":"document.edit"}]' antes do primeiro data-write (precedente
  0267:49); (9) major: supersede legacy publicava novo ANTES de rebaixar prior
  → ux_documents_published_head rejeita (endpoint inoperante) + sem lock
  ordenado + sem MapPgError → lockDocumentRowsInIDOrder({prior,novo}) pós-authz
  (H-PRE-1), reordena prior→superseded primeiro, ExecContext via MapPgError.
  Findings 3-5 da rodada = confirmações (per-target CD ok, lock order global ok,
  repair idempotente ok), sem ação. Fixes 8/9 aplicados: build+vet+vet-integration
  clean, ladder approval completa PASS (F8 exercitado live — ladder aplica
  migrations do zero); testes supersede fortalecidos (ordem observada, não
  assumida).
- 2026-07-28: **rodada 4 NOT-ALIGN → 1 fix (opus)**: F8/F9 confirmados corretos
  pelo Codex; (10) major: re-read under-lock do revision_version prior era
  condicional em `s.repo != nil` (produção sempre injeta repo — services.go:90 —
  mas branch nil usava valor pré-lock do caller, furo OCC) → repo obrigatório
  fail-closed, re-read sempre, teste prova valor under-lock no CAS.
- 2026-07-28: **rodada 5 VERDICT: ALIGN** — ciclo adversarial fechado (5
  rodadas, 8 defeitos corrigidos: 3 blockers, 5 majors). **Stage A COMMITADO
  `616c82ca`** (63 arquivos, +4796/−364). Deploy: api/jobs/worker rebuilt +
  up; migração 0310 aplicada live (50/50), ux_documents_published_head
  presente, logs limpos. Pendente Etapa 2: Stage B (retirement inventory) +
  Stage C (backfill + readiness projection).
- 2026-07-28: **Stage B1 backend implementado (opus, working tree)**: 3 paths
  deletados do spec + regen full (11 api.gen.go); SubmitDocumentRequest ganhou
  plano opcional (planned_effective_from/effective_to/review_due_at/
  superseded_document_id, wire=coluna 1:1); submit tx persiste plano + authz
  document.supersede na ÁREA DO ALVO (H-PRE-1 ok) + código problema renomeado
  submit.invalid_supersede_target; 18 arquivos removidos (services publish/
  scheduler/supersede + handlers + contracts + job scheduled-publish);
  capability document.publish extinta (registry+seed+0311 delete
  role_capabilities 2 rows, golden 112→110 pegou); LoadCurrentPublishedHead-
  ForDocument deletado; lockDocumentRowsInIDOrder removido (coordinator usa
  lockDocumentForRelease inline — verificado). Desvio ratificado: predicado
  ValidateScheduledSupersedeTarget (same-CD) INVERTIDO p/
  ValidateCrossDocumentSupersedeTarget (CD distinto + published) — ADR 0085 diz
  same-doc é implícito, nunca nomeado. e2e re-trabalhado p/ fato durável +
  outcome legal (materializing hold sem artefatos). Ladder: build/vet/
  vet-integration/api-lint 0/boundaries OK/approval+documents+iam+scenarios
  PASS. Defer: check-contract-sync-all FAIL pré-existente (/finalize drift,
  chip task_8bcc6a21). B2 FE em curso.
- 2026-07-28: **Stage B2 FE implementado (opus, working tree)**: approvalApi
  publish/schedulePublish/supersede deletados; SupersedePublishDialog (+test
  +css) deletado; DocumentDetailRoute sem botão Publicar/Agendar/banner;
  useDocumentArtifact sem canPublish/publishContextNotice/activeDocument/
  refetchAll (consumidor único era o dialog); tipos FE regenerados (pnpm
  gen:api — operações/schemas mortos sumiram, SubmitDocumentRequest com plano);
  error-codes.generated.json regenerado; e2e scheduled_publish.spec deletado +
  projeto serial-clock removido dos configs; happy_path/quorum aceitam
  Aprovado|Publicado (coordinator assíncrono). tsc limpo (2 configs), vitest
  147 files/959 tests PASS, eslint só 4 pré-existentes. **Pendência operador:
  14 strings PT-BR de erro autoradas** (error-codes.generated.json estava
  stale, guard false-green; códigos alheios ao 0085 sem mapping) — revisar
  copy em errorMessages.ts. Defers bounded: playwright não rodado (stack
  live), e2e fora de tsconfig (pré-existente), mojibake happy_path
  pré-existente, UI de plano = trabalho futuro.
- 2026-07-28: **Stage B rodada 1 Codex NOT-ALIGN → 4 fixes (opus)**: (11)
  blocker: kind River `scheduled_publish_cutover` órfão pós-deploy → 0311
  ganhou DO block to_regclass-guarded deletando rows não-terminais
  (state::text — 'pending' pode não existir no enum instalado; 'running'
  excluído de propósito) + teste replay/idempotência em
  tests/integration/migrations/migration_0311_test.go; (12) wiki sync 9 docs
  (documents.md rotas/transições/glossário → modelo coordinator; deep-qa
  runbook/matrix com banner de emenda datado; corrigida alegação falsa "jobs
  sem Dockerfile/compose" — ambos existem); (13) prova full-release: subtest
  CoordinatorReleasesOnceArtifactFactsLand no e2e_happy (facts via
  RecordArtifactFactTx produção + Evaluate direto → published+released_at+
  eventos) — NÃO executado localmente (gate METALDOCS_E2E_URL; equivalentes
  determinísticos passam no release_coordinator_integration_test) → defer p/
  QA live pós-rebuild; (14) idempotency_middleware_test derivado de
  router.idempotentRoutes (cobriu /submit,/review omitidos). Ladder green
  (flake TestGoVetPasses = timeout compartilhado, chip task_231e6db1).
- 2026-07-28: **Stage B rodada 2 Codex: 2 findings só-docs** (High-risk tally
  jobs.md contradizia a própria tabela; matrix C6/C7 ainda prescreviam CTA de
  publicar) → corrigidos inline (incl. linha lease_reaper igualmente stale).
  **Stage B COMMITADO `ff5851c7`** (tree limpa). Rebuild api/jobs/worker/web
  em curso → e2e HTTP live fecha defer F13.
- 2026-07-29: **Deploy Stage B verificado + F13 = defer limitado.** Stack
  rebuildada de ff5851c7 e saudável; migração 0311 aplicada live (51/51):
  0 rows document.publish em metaldocs.role_capabilities, 0 river_job
  não-terminal de scheduled_publish_cutover. e2e HTTP
  (TestE2E_HappyPath_HTTP) FALHOU com 401 em todo subtest autenticado —
  causa: o teste autentica via headers X-Tenant-ID/X-User-ID/X-User-Roles,
  mas o authn de produção os REMOVE incondicionalmente
  (internal/modules/iam/delivery/http/middleware.go:70-81, defesa contra
  identidade fornecida pelo cliente; sem toggle dev). Harness legado
  pré-sessão — defeito pré-existente, NÃO regressão Stage B. Mitigação: os
  equivalentes determinísticos do caminho full-release passam em
  tests/integration/approval/release_coordinator_integration_test.go (incl.
  hold materializing → release). Chip aberto para modernizar o harness p/
  login de sessão (task_8f8ac5a3). F13 permanece defer limitado até QA live
  via login curl dev-seed.
- 2026-07-29: **Follow-up VerdictID fechado** (opus): id do ledger de verdict
  agora atravessa ReviewVerdictResult → handler (fresh E replay; replay reusa
  o slot SignoffID do envelope de idempotência, mecanismo idêntico ao signoff
  — sem formato novo persistido). Contrato já declarava verdict_id required —
  zero mudança de spec. NOTE falso removido. +4 testes (incl. primeiro
  handler test da rota review-verdict). Ladder approval PASS.
- 2026-07-29: **Stage C design alinhado com Codex (rodada 1 NOT-ALIGN → 6
  correções aceitas)**: (1) Pin early-return em doc já frozen → backfill usa
  novo FreezeService.RepairMaterialization (enqueue materialize
  generation-aware direto); (2) identidade autoritativa = frozen_content_hash
  da instância (= ContentHashAtSubmit, freeze.go:58) — preflight aborta em
  NULL/não-pinned, NÃO compara com content_hash atual (produção também não);
  fast-forward de ponteiros legados EXCLUÍDO do backfill (no-fallback:
  presença de coluna ≠ readiness; 02bef5ae re-materializa fresh); (3)
  quarentenados d18fbfdf/45c9e784 FORA do backfill até Etapa 5 (tool com
  allowlist explícita -docs, re-rodável pós-expurgo); (4) segurança da tool =
  mesmo caminho background do release_evaluate job (SeedTxTenant +
  authz.BypassSystem tx-local), testado contra tripwire; (5) sweep de
  reconciliação (ADR §200-204) NÃO existia → vira entregável Stage C
  (release-hold-reconciler, alert-only ADR 0068, dual-define ADR 0067); (6)
  projeção release via LEFT JOIN LATERAL na query única do GetDocument
  (repository.go:285), não segunda query. Posições: go-run one-shot OK;
  `state` derivado server-side; enqueue de avaliação imediato (hold
  materializing honesto). 3 implementadores opus em voo (W1 backfill+repair,
  W2 sweep, W3 projeção); W4 FE aguarda W3.
- 2026-07-29: **Stage C implementado (4 pacotes opus) + 2 rodadas Codex →
  ALIGN.** W1: scripts/release-backfill (go-run, -docs allowlist, dry-run
  default via rollback de sentinela, SeedTxTenant+BypassSystem, sem
  fast-forward de ponteiros legados) + FreezeService.RepairMaterialization
  (:249, fail-closed p/ não-pinned) + 6 testes integração. W2:
  release-hold-reconciler (15min/threshold 30min, alert-only + evento
  governança release.generation.stuck_alert; SEM re-enqueue — avaliação
  carimba last_evaluated_at, o próprio sinal do detector; predicado com 3
  conjuntos extras testados: status liberável, timer futuro armado não
  alerta, só freeze-head; read-port ReleaseHoldReader publicado). W3:
  DocumentDetailResponse.release required+nullable (regra
  SHAPE-NULLABLE-NOT-REQUIRED) via LEFT JOIN LATERAL na query única
  GetDocument (precedente active_instance_reader; boundaries OK,
  contract-sync OK, allowlist seed-chokepoint shift mecânico +65). W4:
  documentReleasePresentation.ts resolver + bloco InlineAlert sem CTA
  (anomalia tone=warning por aria); copy PT-BR listada p/ revisão operador.
  Rodada 1 Codex: 3 findings → fix Major (poll gate único
  isDocumentLifecycleSettling — holds transientes 5s;
  awaiting_effective_date não faz poll, data pode estar semanas à frente) +
  fix Minor (publicationTimestamp = release.released_at; hold = SEM data;
  fallback approval.completed_at só release null) + Nit dedup de alerta do
  reconciler = defer bounded. Rodada 2: **ALIGN** (1 hole não-alcançável
  notado: payload misto scheduled+released impossível pela transição atômica
  do coordinator). Verificação união: build/vet/vet-integration 0, api-lint
  0, boundaries OK, ladder pacotes tocados PASS, FE tsc/vitest/eslint verdes
  (41 testes adapter). Wiki 6 docs sincronizados (jobs 4→6 periodic;
  CLAUDE.md lista de jobs atualizada); drift pré-existente
  backend-target-architecture (narra scheduler de lease aposentado) flagado,
  fora de escopo. Próximo: commit → rebuild stack → backfill live 02bef5ae
  (quarentenados d18fbfdf/45c9e784 só na Etapa 5).

- 2026-07-29: **Backfill live executado + defeito latente Stage A achado por
  QA live + fix + RELEASE PROVADO.** Commit Stage C `1f2c6376` deployado;
  backfill aplicado (doc 02bef5ae, generation 1de1e704); worker materializou
  docx+pdf frescos (pipeline provado). Avaliação porém falhou 5 tentativas:
  `taxonomy: tenant context: tenant: not present in context` — preflight do
  coordinator (release_coordinator.go:219) consulta review interval via
  taxonomy GetByCode, que é HTTP-shaped em dobro: resolve tenant+actor do
  ctx Go (authz_guc.go) E exige authz.Require(CapTaxonomyView), que actor
  system nunca teria. Testes determinísticos mascaravam (suite do
  coordinator stubba a porta de cadência). Fix (padrão sancionado do próprio
  phase-1 do coordinator): taxonomy `GetByCodeSystem` — tx curta própria,
  SeedTxTenant do parâmetro EXPLÍCITO, BypassSystem fail-closed fora de
  WithBackgroundBypass + auditado F8; adapter de review-interval consome
  slice estreita nova (boundary intacto, SQL de document_profiles fica no
  taxonomy). 2 rodadas Codex: r1 NOT-ALIGN com Major real — tx só com
  rollback apagava o evento de auditoria F8 do bypass (sink escreve NA tx);
  fix = commit nos caminhos found e not-found + teste que assere linhas
  audit_events COMMITADAS via conexão metaldocs_ci NOBYPASSRLS separada
  (0→1→2; negative check confirmou que o guard morde); r2 **ALIGN**. Commit
  `f773b181`; jobs rebuild+deploy; retry forçado dos jobs 4097/4099 →
  **completed**. Prova live: generation released_at+last_evaluated_at
  carimbados, hold NULL; doc `published` effective_from=05:39:42Z; 1 evento
  governança `document_published` (system:release-coordinator); fanout
  `document.published` completed (job 4116); bypass audit commitado; e W2
  provado live — reconciler emitiu `release.generation.stuck_alert` às
  05:30 enquanto a generation estava presa >30min. **Etapa 2 FECHADA**
  (F-QA4-13/14 já estruturalmente fechados; quarentenados d18fbfdf/45c9e784
  → Etapa 5).
- 2026-07-29: **3.3 F-QA4-8 fechado + Etapa 3 FECHADA com QA live dos 3
  itens.** Inbox/dashboard mostravam UUID cru e avatar "?" porque o FE
  ignorava o roster de atores que o backend já resolve. Fix: contrato
  ApprovalInboxItem ganha `controlled_document_code` required+nullable
  (regen full); ListWorklist faz LEFT JOIN tenant-scoped em
  controlled_documents; handler mapeia vazio→null explícito (fail closed).
  FE: `mapApprovalChain` agora emite um item por entrada de
  `stage.actors[]` (display_name real, flowState por status do ator,
  signedAt casado por user id); `pickStageDecisiveItem` substitui
  `group[0]` cego no ArtifactDetailView (backend ordena signoffs por
  signed_at ascendente ⇒ head = decisor MAIS ANTIGO; prioridade: rejeitado
  > última aprovação assinada > ativo/pendente > head); render sites
  preferem `code ?? uuid` (degradação verdadeira). 2 rodadas Codex: r1
  NOT-ALIGN (P1 group[0] escolhia aprovador mais antigo em estágio
  rejeitado + 2 P2 de teste); fixes por agente opus; r2 **ALIGN**. Commit
  `8446f783`; api+web rebuild+deploy. **QA live (padrão curl dev-seed):**
  (3.3) submit de IT-USINAGEM-004 → inbox do approver-test devolve
  `controlled_document_code: "IT-USINAGEM-004"` e instance-detail devolve
  actors com `display_name: "Approver Test"`; instância QA cancelada após
  a prova (doc voltou a draft). (3.1 F-QA4-9) fluxo autosave completo
  acquire→presign→PUT MinIO→commit SEM form_data_snapshot → **200**
  revision 16 (antes 500). (3.2 F-QA4-4) GET /documents/{id} devolve
  identificação real (code, profile, área, revisão coerente com o
  autosave recém-commitado). Notas: senhas do wiki local-dev-startup
  estão stale para `admin`/`author-test` no stack compose (hashes de
  author-test/approver-test resetados para o valor dev conhecido via DB —
  prática igual ao seed 0159); If-Match `*` no cancel devolve 409 apesar
  de o contrato prometer skip (drift contrato×runtime, chip aberto).
- 2026-07-29: **Etapa 4.1 — jornada template-do-zero executada E2E pela
  primeira vez (curl dev-seed, 3 atores).** Fluxo provado: author-test cria
  template (perfil it) → autosave presign/PUT/commit registra content_hash →
  admin cria rota `subject_kind=template` via API crua (FE route-builder não
  suporta — fato D6) → submit → inbox do approver mostra item de template
  (code null verdadeiro) → signoff approver (password reauth) → publish
  admin (SoD 403 ISO_SEGREGATION_VIOLATION corretamente bloqueou publish
  pelo autor) → wizard lista template publicado (`?doc_type=it&published=
  true`) → IT-USINAGEM-005 criado com clone **byte-idêntico** (sha256 igual
  ao docx do template). Segunda rodada com v2 contendo tokens de sistema
  ({{doc_code}} etc.): validação de tokens no publish passou, IT-USINAGEM-006
  criado; fill-in-schema vazio é correto (tokens de sistema ≠ fill-in).
  **Achados (4.2):**
  - F-E4-1 (P1): submit com content_hash NULL → **500 cru** vazando
    constraint `chk_template_version_content_hash_non_draft`; caminho
    docx-upload-url deixa hash NULL (armadilha beco-sem-saída); precisa
    precondição 4xx problem+json ANTES do lock.
  - F-E4-2 (P2): createApprovalRoute 201 devolve projeção quase vazia
    (name "", version 0, active false, stages null) enquanto o DB está
    completo e ativo — runtime mente pro cliente.
  - F-E4-3 (P2): signoff de template devolve `signoff_id: ""` (mesma
    classe do gap VerdictID já corrigido).
  - F-E4-4 (P2): versão pós-aprovação com `approver_id: null` apesar de
    approved_at carimbado (atribuição eQMS ausente na superfície).
  - D6 (estrutural, ADR): rota por template_id obriga admin a criar rota
    nova via API crua PARA CADA template; gate existe só no submit (409
    APPROVAL_ROUTE_MISSING), não na criação (assimetria com D2 de
    documentos). Evidência viva pro ADR config-first re-key.
  - Notas: /documents/{id}/view em draft recém-criado → 404
    not_found.revision (editor usa docx-url; questão UX, não flag);
    dev-seed sem persona publicadora (approver não tem template.publish;
    jornada exigiu admin como publicador ⇒ autor teve de ser author-test).
- 2026-07-29: **D6 fechado — ADR 0086 aceito (Codex-aligned rev 4, commit
  `4eef7d77`).** Rotas de aprovação de template deixam de ser por
  template_id e passam a ser config-first por perfil (`subject_key =
  doc_type_code`), simétricas às rotas de documento: profile_code NOT NULL
  (substitui CHECK da migração 0297), RoutePolicy (ADR 0081) nos caminhos
  create E update, criação de template hard-gated 409
  APPROVAL_ROUTE_MISSING (simetria D2), `doc_type_code` obrigatório (422) —
  templates genéricos exterminados incluindo o ramo OR do list-by-profile e
  o create-sem-perfil (comportamento ativo de produto, quebra deliberada
  ratificada), keying de instância preservado (versão, ADR 0082), os 3
  pontos de resolução migram juntos (submit + preview via mesmo read port +
  seletor do handler HTTP; approval_version_reader projeta doc_type_code),
  mudanças contract-first declaradas (inversão profile_code no create-route
  + 422/409 no create-template), cutover duro sem fallback. Review: 4
  rodadas Codex (r1 6 achados — superfícies de resolução, genérico é
  produto ativo, gap de port, skip de policy no update; r2 3 — fonte da
  chave no preview, contratos; r3 1 — redação update; r4 ALIGN).
  **Implementação = unidade própria, enfileirada APÓS gate de push da
  Etapa 5.**
