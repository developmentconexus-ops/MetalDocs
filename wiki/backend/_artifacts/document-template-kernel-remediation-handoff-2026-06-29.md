# Handoff de Execução — Remediação do Kernel de Documentos e Templates (Grade A)

> **Para a sessão `/goal` (autônoma, contexto-zero):** este documento é a sua ordem de missão.
> Você NÃO herda nenhum contexto de conversa anterior — isso é intencional (evitar viés).
> Toda a verdade técnica está em arquivos versionados. Leia-os antes de agir.

---

## 0. Objetivo (uma frase + critério de sucesso)

Implementar **100% do roadmap de remediação** da auditoria do kernel doc/template, em nível **Grade A**,
padrão de indústria, engenharia sênior — deixando **documento e template operando e funcionando sem
nenhum problema**, validados em API/integração e, ao final, em QA dirigido pelo Preview **como um
usuário real**. **Sem workaround, sem patch, sem hardcode, sem gap de engenharia.** Encerrar **somente**
quando todos os achados estiverem implementados, todos os gates verdes, e o QA final de Preview passar
sem erros.

---

## 1. Fonte de verdade — LEIA NESTA ORDEM antes de qualquer código

1. **Relatório de auditoria (defeitos + fixes):**
   `wiki/backend/_artifacts/document-template-kernel-deep-audit-2026-06-29.md`
   — 71 achados verificados adversarialmente. Cada achado tem: **Severidade · Problema · Evidência
   (`file:line`) · Invariante/ADR violado · Irmão canônico (`file:line`) + padrão SaaS maduro · Fix.**
   Seções: §2 achados por subsistema · **§3 roadmap priorizado (Onda 0→3)** · §4 o que já está correto
   (não regredir) · §5 limites/itens que precisam verificação runtime.
2. **`CLAUDE.md`** (raiz do repo, auto-carregado) — os 6 invariantes não-negociáveis, a regra de
   Orientação, a regra Global-Maximum-not-Local-Maximum, os comandos de build/start/test.
3. **`MEMORY.md`** (auto-carregado) — em especial:
   `dev-seed-template-approval-authz-gap`, `doc-template-kernel-audit-2026-06-29`,
   `adr0035-flat-envelope-drift`, `authz-root-cause-over-symptom`, `legacy-test-deletion`,
   `test-framework-hard-gate`, `advisory-lock-deadlock-constraint`.
4. **Skill `developing-new-work`** — gate de system-impact (ver §4).

**Não re-derive os defeitos.** O relatório é a fonte. Leia o achado, leia o irmão canônico que ele cita,
replique o padrão já validado.

---

## 2. Escopo — todos os 71 achados, na ordem das ondas (§3 do relatório)

### Onda 0 — Desbloqueio imediato (lifecycle quebrado / QA bloqueado) — FAZER PRIMEIRO
- **F-T1 [CRÍTICO]** — FE de templates omite `Idempotency-Key` em submit/review/approve → 400 sempre na
  UI. Adicionar `idempotencyKey: crypto.randomUUID()` (via `useRef` estável, **uma chave por ação do
  usuário, não por render**) aos 3 call-sites em `frontend/apps/web/src/features/templates/api/templates.ts`.
  Replicar o padrão já validado de `createTemplate` (mesmo arquivo, :107) e
  `approval/api/mutationClient.ts:30`.
- **F-IAM1 [HIGH]** — dev-seed `approver` sem linha em `user_process_areas` → tier-2 403. Adicionar linha
  UPA para `approver` na área `rh`, espelhando `approver-test`
  (`db/dev-seeds/0001_local_dev_seed.sql:149-157`). **Pré-requisito de TODO QA de approve/publish via API
  ao vivo** (KNOWN FACT do dev-seed) — sem isto, o QA final de Preview do lifecycle não roda.

### Onda 1 — High (integridade de publish, authz, contrato, DB, FE)
- **F-T2** — gate de `content_hash` em `Approve` (espelhar T-004 de `PublishTemplateVersion`,
  `lifecycle.go:349`).
- **F-CD1 + F-CD2 + F-CD3 (fix unificado)** — migrar os 5 handlers de approval de documentos
  (publish/schedule/supersede/obsolete/cancel) para o middleware `idempotency.Require()` — isso declara o
  header na spec OpenAPI + dá replay store + valida UUID de uma vez. Espelhar
  `templates/delivery/http/handler.go:68-74`. **(decisão spec-level + regen oapi-codegen).**
- **F-D1** — substituir `IsDocumentOwner` por `authz.Require(CapDocumentView, area)` nos ~19 read-sites de
  `documents` (ADR 0022 — capability, nunca ownership). Espelhar `view_service.go:76` /
  `fillin_authz.go:21-40`. Blast radius alto — cuidado.
- **F-D2** — authz in-tx em `CreateCheckpoint`/`ListCheckpoints` (espelhar `CommitUpload`).
- **F-DB1 + F-DB4 + F-DB7 (uma migração)** — CHECK de `status`, `content_hash` (=''/len 64),
  `version_number >= 1` em `templates_template_version`. Espelhar `documents` (`baseline:2025, 2030`).
- **F-DB2 + F-DB6** — partial-unique `one-published-per-template` + `(template_id, revision_number)`
  unique. Fecha race de double-publish sob READ COMMITTED. Espelhar
  `ux_approval_instances_active_document_id`.
- **F-O4** — `tenant_id` + RLS em `document_exports` (invariante multi-tenant). Espelhar padrão 0237.
- **F-FE1, F-FE2, F-FE3** — upload-fail bloqueia submit (espelhar `useDocumentAutosave`); surfacing de
  412 stale-conflict no editor; 404 readonly por status (não silenciar para in_review/approved/published).
- **F-DB3** — regenerar `db/baseline/0001_current_schema.sql` via `pg_dump --schema-only --no-owner
  --no-privileges` de DB totalmente migrado (NÃO mascarar com mais `IF NOT EXISTS`).

### Onda 2 — Medium (fronteiras de módulo, hygiene, config, contrato)
F-CD4 (injetar ports via Dependencies + composition root) · F-CD5 (constantes locais + parity test) ·
F-CD6 (remover authz duplicado em CreateTx) · **F-CD7 (nova capability `template.manage` — precisa ADR +
gate)** · F-O1/F-O2/F-O3/F-O5/F-O6 (tenant-guards e helpers no kernel objectstore — F-O2 **antes** de
wirar a rota webhook) · F-O7 (`MinIORegion` env-configurável) · F-R1 (deletar `Freeze()` dead-code) ·
F-R2/F-R3 (config env do staging worker + DLQ/observabilidade) · F-D3 (wirar `WithIAMReader` +
fail-closed PHUser) · F-D4 (`Idempotency-Key` de cliente no `SubmitHandler`) · F-C1→F-C2→F-C3 (query
params de `GET /templates` na spec → tipo gerado no FE → remover fallback) · F-T3 (remapear 412→409 na
concorrência) · F-T4 (`approver_role` no contrato de create) · F-DB5 (trigger de consistência de tenant
+ predicado em `ObsoletePreviousPublished`) · F-FE4/F-FE5/F-FE7 (dirty-state, type=button + pt-BR,
`importDocx` `.ok`).

### Onda 3 — Low (hygiene, docs, defer pós-v1 — implementar todos que couberem)
F-T5 · **F-T6 (GC sweeper de órfãos — novo job em `jobs`, precisa gate)** · F-D5 · F-D6 · F-CD8 · F-CD9 ·
F-R4 · F-R5 · F-R6 · F-O8 · F-C4 · F-C5 · F-IAM2 · F-IAM3 · F-TST1 · F-TST2.

> Os achados completos (evidência + irmão canônico + fix detalhado) estão no relatório §2.
> O usuário quer **tudo** finalizado; defers só onde o relatório marca explicitamente "pós-v1" **e** o
> item for genuinamente fora de escopo v1 — mesmo esses, implementar se couber sem violar invariante.

---

## 3. Ordem e dependências (não-negociável)

1. **Onda 0 antes de tudo.** F-T1 + F-IAM1 desbloqueiam o lifecycle e o QA e2e. Nada de Preview QA de
   approve/publish funciona antes de F-IAM1.
2. Depois **Onda 1**, depois **2**, depois **3**.
3. **Fixes unificados** (uma mudança resolve vários): F-CD1+2+3 (um middleware), F-DB1+4+7 (uma migração),
   F-DB2+6 (uma migração), F-O8+F-R6 (um helper de keys), F-DB8⊆F-DB5.
4. **Coordene fixes que tocam o mesmo arquivo/spec** para não conflitar (ex.: vários itens FE em
   `templates.ts`; vários itens de spec em `api/openapi/v1/openapi.yaml` → regenerar uma vez por lote).

---

## 4. Processo obrigatório — COMO construir (regras de engenharia)

- **Gate `developing-new-work` ANTES de desenhar** qualquer item que seja **trabalho novo**: nova
  capability **F-CD7** (`template.manage`) e novo job **F-T6** (GC sweeper). Emite system-impact analysis
  + veredito Green/Yellow/Red. **Red bloqueia o design** — pare e resolva o gate. Para os demais (fixes
  dentro de módulos existentes), aplique a **regra de Orientação** do CLAUDE.md: declare (a) módulo dono,
  (b) invariantes tocados, (c) leia `wiki/modules/<nome>.md` antes de tocar.
- **Espelhe o irmão canônico** que o achado cita. NÃO invente solução. Compare com o módulo já validado
  que usa o mesmo padrão/lógica. **Nada hardcoded. Sem workaround. Sem patch.** Se a base do que você ia
  tocar for legacy/patch, **não otimize dentro dela** — proponha a estrutura global-maximum e pare
  (regra Global-Maximum do CLAUDE.md).
- **Contract-first:** rotas mudam SÓ via `api/openapi/v1/openapi.yaml` + `oapi-codegen` (regenerar). A
  spec é a verdade da rota.
- **DB:** migrações forward-only, auto-registrantes, espelhando os siblings da baseline; o app-check é só
  a primeira linha amigável, o DB enforça o invariante.
- **Cross-module:** acesso só via service/interface publicada — nunca repo/SQL/domain de outro módulo.
- **TDD:** teste que falha → fix mínimo → verde. Use o **framework de fixtures canônico** (`test-discipline`
  / ADR 0034). Testes legacy de scaffolding que quebrarem: deletar, não manter (memória
  `legacy-test-deletion`); reparar só guards de contrato/invariante.
- **Execução com subagents de sessão fresca, um por task** (sem herdar contexto → sem viés). **Two-stage
  review** por task: spec-compliance primeiro, depois code-quality. Modelo: implement/review em sonnet,
  mecânico em haiku, **nunca fable**, ≤15 agentes concorrentes (memória `workflow-model-balancing`).
  Skills sugeridas: `writing-plans` (plano por onda) → `subagent-driven-development` (executa task-a-task
  com review). Para a escala do programa, `mission`/`milestone` é aceitável se preferir gates de
  separação-de-poderes — mas o relatório já é a pesquisa+decomposição; não re-planeje do zero.
- **ADR quando exigido:** F-CD7 (nova cap) e qualquer desvio MUST do target-spec
  (`wiki/architecture/backend-target-architecture.md`) precisam de ADR citando REQ IDs.

---

## 5. Definition of Done — validação (evidência antes de fechar)

**Por achado:** reportar comando + outcome + disposição. Espelhou o irmão canônico? Sem hardcode/patch?

**Gates por onda (todos verdes):**
- `go build ./...` e `go test ./...` verdes.
- Frontend: `make test` verde + `npm run typecheck:docx-v2` (e typecheck do app web) limpos.
- Contrato: `oapi-codegen` regenerado **sem drift** após edição de spec.
- DB: migrações aplicam limpo; `.\scripts\check-system-runnable.ps1` OK; baseline regenerada (F-DB3)
  bate com migrado.
- QA por finding via **suíte de integração / API** (o relatório §5 recomenda rodar a suíte após Onda
  0+1 para confirmar F-T1, F-IAM1/copy-on-spawn e2e, F-DB2 sob concorrência).

**Gate FINAL de aceitação (rodar SÓ no final, depois de tudo implementado) — QA real-user via Preview:**
- Subir o app (`.\scripts\start-api.ps1`; frontend dev server via Preview tools).
- **Dirigir o app como um usuário real** (clicar, preencher, usar o editor eigenpal de verdade) — **NÃO
  fabricar bytes de docx nem injetar via fetch**. Fluxo de template ponta-a-ponta:
  criar template → editor → digitar conteúdo → autosave (docx válido na key canônica) → submit → review →
  approve → publish. Fluxo de documento: preencher placeholders → submit → signoff → publish.
- **Login com a conta `approver` do wiki QA** (`wiki/references/local-dev-credentials.md`) para o
  caminho de aprovação (não admin — SoD). Isso só funciona depois de F-IAM1.
- **ZERO erros** = PASS. Qualquer erro = FAIL honesto (não declarar PASS sobre erros).
- Atenção ao **KNOWN FACT eigenpal** (F-FE3): editor entrava em loop/hang quando o docx estava ausente
  (janela 404 pré-autosave). F-FE3 corrige; validar que não loopa mais.
- Capturar evidência (screenshot/network/logs) do fluxo passando.

**Critério de término:** NÃO encerrar até 71 achados implementados Grade A + todos os gates verdes +
Preview QA final sem erros. Reportar evidência por onda.

---

## 6. Guardrails (segurança e escopo)

- **Commit após trabalho verificado é permitido** (autorização permanente, CLAUDE.md §5.0).
  **NUNCA dar push sem permissão explícita do operador.**
- **Nunca** ler/printar/commitar/expor segredos `.env`. Startup local **só** via scripts PowerShell —
  não bash, não `source .env`.
- **Mantenha o escopo nos achados.** Não refatore código adjacente fora da lista. Não reverta trabalho do
  usuário.
- **Pare em contradição de arquitetura** — surface o problema, não patcheie em volta (regra CLAUDE.md +
  memória `authz-root-cause-over-symptom`). Authz é capability, nunca role.
- **Não force-grant authz** em DB para "destravar" QA — F-IAM1 é o fix correto (corrigir o seed), não um
  grant manual.
- Restrição advisory-lock (memória `advisory-lock-deadlock-constraint`): nunca chamar read que grava
  authz dentro de tx que segura lock (H-PRE-1).
- Heavy-write (builds/caches) preferir disco D: se aplicável (memória `machine-ssd-degraded-writes`).

---

## 7. O que NÃO regredir (§4 do relatório — já correto)

Copy-on-spawn store-then-reference · migração 0250 (UNIQUE docx_storage_key) · PDP two-tier no design ·
CAS optimistic-lock em `UpdateVersionTx` · DI por composition-root em `documents`/`taxonomy` ·
`view_service.go:76` (padrão authz correto) · Pin tx-only (ADR 0015) · `render/domain` como contrato
publicado (ADR 0050). Os fixes devem **alinhar a esses padrões**, não substituí-los.
