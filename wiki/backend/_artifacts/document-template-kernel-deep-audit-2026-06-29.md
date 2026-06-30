# Auditoria de Engenharia — Kernel de Documentos e Templates (MetalDocs)

> Revisor sênior · 71 achados verificados adversarialmente · escopo: módulos `documents`, `templates`, `controlleddocuments`, `render`, plataforma `objectstore`/`idempotency`, contrato `api/openapi`, frontend `templates`/`approval`/`documents`, schema `db/baseline` + `db/migrations`.

---

## 1. Sumário Executivo

### Saúde geral do kernel

O kernel doc/template está **estruturalmente sólido mas com uma classe de defeito crítica ativa no caminho de produção** e várias lacunas de invariante DB e de fronteira de módulo. A arquitetura central (two-tier PDP, outbox transacional, store-then-reference, contract-first) está correta onde é honrada — o problema dominante não é arquitetura, é **adesão inconsistente**: o frontend de templates não cumpre um header que o backend exige em 100% das chamadas, e o DB delega ao app invariantes que deveriam ser enforced no schema.

Um único bug (Idempotency-Key ausente nos callers de lifecycle do FE de templates) **quebra completamente** o fluxo de aprovação de templates pela UI. Foi reportado por 5 achados independentes — é o achado de maior prioridade absoluta.

### Contagem por severidade efetiva (pós-verificação, deduplicada)

| Severidade | Achados distintos (deduplicados) |
|---|---|
| **Critical** | 1 (reportado por 5 achados) |
| **High** | 9 |
| **Medium** | 24 |
| **Low** | 19 |
| **Total bruto** | 71 |
| **Total deduplicado** | ~53 root-causes distintos |

### Os 5 bloqueadores de "documentos e templates operando sem problemas"

1. **[CRÍTICO] FE de templates omite `Idempotency-Key` em submit/review/approve** → todo o lifecycle de aprovação de template retorna 400 na UI, sempre. (5 achados convergentes.)
2. **[HIGH] Seed dev: usuário `approver` sem linha em `user_process_areas`** → tier-2 `authz.Require` falha sempre → 403 `FORBIDDEN_CAPABILITY` mesmo com `/auth/me` listando a capability. Bloqueia QA e2e do lifecycle.
3. **[HIGH] Contract drift de idempotência em 5 rotas de approval de documentos** → o handler exige `Idempotency-Key` que o OpenAPI não declara; qualquer SDK gerado pela spec recebe 400.
4. **[HIGH] `Approve` (caminho de publish com reviewer) sem gate de `content_hash`** → uma versão com hash vazio pode ser publicada (binário do docx não confirmado), violando o gate T-004 que os outros caminhos de publish aplicam.
5. **[HIGH] Lacunas de invariante DB em `templates_template_version`**: sem CHECK no `status`, sem partial-unique de "one-published-per-template" → race sob READ COMMITTED pode produzir duas versões publicadas; FSM corrompível por escrita direta.

---

## 2. Achados por Subsistema

> Convenção: cada achado lista **Severidade · Problema · Evidência · Invariante/ADR violado · Irmão canônico (file:line) + padrão SaaS maduro · Fix**.

---

### 2.1 Templates — Lifecycle de Aprovação (FE + delivery/http + application)

#### F-T1 · [CRÍTICO] FE omite `Idempotency-Key` em submit / review / approve — 400 garantido na UI

> **Deduplicação:** este é o mesmo root-cause reportado por 5 achados independentes (dimensões `templates-lifecycle`, `idempotency-and-lifecycle`, `contract-dto-drift`, `frontend-editor-ux`, `controlleddocuments-signoff`). Mérito de severidade reforçado pela convergência. Evidência mesclada abaixo.

- **Problema:** `submitForReview`, `reviewVersion`, `approveVersion` chamam `apiFetch` sem a opção `idempotencyKey`. O `apiFetch` só emite o header quando a opção está presente. As três rotas backend estão envolvidas em `h.idempotent()` → `idempotency.Require()`, que retorna 400 imediato quando o header falta. O OpenAPI declara `required: true` nas três operações. **As três ações nunca funcionaram pela UI.**
- **Evidência:**
  - `frontend/apps/web/src/features/templates/api/templates.ts:214-219` (submitForReview): `apiFetch(..., { method: 'POST' })` — sem `idempotencyKey`.
  - `:228-232` (reviewVersion) e `:251-259` (approveVersion): idem, só `{ method, body }`.
  - `frontend/apps/web/src/lib/api/client.ts:46`: `...(idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {})` — header condicional à opção.
  - `internal/modules/templates/delivery/http/handler.go:55-57`: submit/review/approve todos com `h.idempotent(...)`.
  - `internal/platform/idempotency/middleware.go:91-94`: `if key == "" { writeErrJSON(w, 400, problem.CodeIdempotencyKeyRequired, ...) }`.
  - `internal/modules/templates/api/api.gen.go:470-479` etc.: binding gerado retorna erro de header obrigatório ausente.
  - `api/openapi/v1/openapi.yaml:1451-1454, 1482-1485, 1513-1516`: `Idempotency-Key in: header, required: true`.
- **Invariante violado:** Contract-first (spec é a verdade da rota); cadeia de middleware fixa (camada de idempotência não-negociável); cliente DEVE cumprir headers obrigatórios declarados.
- **Irmão canônico:** `frontend/apps/web/src/features/templates/api/templates.ts:107` (`createTemplate` passa `idempotencyKey: cmd.idempotencyKey`); `frontend/apps/web/src/features/documents/api/documents.ts:43` (`finalizeDocument` gera `crypto.randomUUID()` inline); `frontend/apps/web/src/features/approval/api/mutationClient.ts:30` (`opts.idempotencyKey ?? uuidv4()` — sempre emite). **SaaS maduro (Stripe):** todo endpoint mutante carrega uma idempotency-key gerada pelo cliente por ação do usuário; o servidor deduplica numa janela temporal.
- **Fix:** adicionar `idempotencyKey: crypto.randomUUID()` nas opções de `apiFetch` dos três call-sites. Gerar **uma chave por ação do usuário** (não por render) — usar um `useRef` estável como `TemplateWizardPage.tsx:52` já faz para create, evitando chave duplicada em re-render do React. Não hardcodar; replicar o padrão já validado de `createTemplate`/`finalizeDocument`.

---

#### F-T2 · [HIGH] `Approve` (caminho reviewer→published) sem gate de `content_hash` T-004

- **Problema:** `Service.Approve` transiciona approved→published sem checar `version.ContentHash`. `SubmitForReview` (linha 35-37) e `PublishTemplateVersion` (linha 348-351) aplicam o gate; `Approve` é o único caminho de publish sem ele. Uma versão com `ContentHash=""` (correção direta no DB, bug de migração) seria publicada sem docx confirmado.
- **Evidência:** `internal/modules/templates/application/lifecycle.go:197-283` (Approve, branch accept chama `CanTransition(VersionStatusPublished)` e seta status sem checar hash); `:348-351` (gate explícito em PublishTemplateVersion); `internal/modules/templates/domain/version.go:78-104` (`CanTransition` é validador puro de grafo de status, não inspeciona hash).
- **Invariante violado:** gate de publish T-004 (citado inline no próprio código); "DB enforces invariants"; simetria entre todos os caminhos de publish.
- **Irmão canônico:** `internal/modules/templates/application/lifecycle.go:348-351` (mesmo módulo, mesmo gate). **SaaS/ISO 9001:** controle documental exige confirmação do binário antes do publish; todos os caminhos devem aplicar.
- **Fix:** adicionar no topo do branch accept de `Approve` (após buscar a versão, antes de `CanTransition`): `if version.ContentHash == "" { return nil, domain.ErrContentHashMismatch }`. Espelhar exatamente a linha 349. Adicionar teste unitário: Approve com `Accept=true, ContentHash=""` → `ErrContentHashMismatch`.

---

#### F-T3 · [MEDIUM] TOCTOU em SubmitForReview/Review/Approve — classificação de erro errada no perdedor da corrida

- **Problema:** `SubmitForReview` faz `GetTemplate`/`GetVersion`/`GetApprovalConfig` **fora** da tx e só depois abre tx + `authz.Require` + `UpdateVersionTx`. Sem `SELECT FOR UPDATE`. Duas chamadas concorrentes ambas leem `draft`, ambas passam a checagem de domínio. O CAS `lock_version` em `UpdateVersionTx` é o guard real (impede double-commit), mas o perdedor recebe `ErrStaleLockVersion` (412) em vez de `ErrInvalidStateTransition`/`ErrConcurrentTransition` (409). Idempotency é por `(tenant, actor, route, key)` → dois atores distintos com chaves próprias não são deduplicados.
- **Evidência:** `internal/modules/templates/application/lifecycle.go:20-42` (3 reads fora da tx, tx na linha 72); `internal/modules/templates/repository/postgres.go:362-363` (CAS `AND lock_version = $16`, sem FOR UPDATE), `:396` (retorna `ErrStaleLockVersion`); `internal/platform/idempotency/postgres_store.go:96` (conflito em `(tenant_id, actor_user_id, route_template, key)`).
- **Invariante violado:** classificação de erro de domínio na concorrência; reads que gateiam mutação deveriam estar no snapshot da tx.
- **Irmão canônico:** módulo `documents` usa `SELECT FOR UPDATE` para transições de lifecycle (global-maximum); o padrão CAS é variante aceita do MetalDocs.
- **Fix:** o CAS já previne corrupção — o defeito real é a **classificação de erro**. Remapear `ErrStaleLockVersion`→`ErrConcurrentTransition` quando a intenção é transição de status, em Submit/Review/Approve. (Defer do `SELECT FOR UPDATE` como melhoria de fundo, não bloqueante.)

---

#### F-T4 · [MEDIUM] `CreateTemplate` hardcoda `ApproverRole: "approver"`; spec não expõe o campo

> **Deduplicação:** mesmo root-cause em dois achados (dimensões `templates-lifecycle` e `authz-tiers`). Evidência mesclada.

- **Problema:** o handler grava `ApproverRole: "approver"` na criação (em `routes_generated.go:57`), propagado tanto para `version.PendingApproverRole` quanto para `template_approval_configs` (via `UpsertApprovalConfigTx`). O OpenAPI de `POST /templates` só aceita `[key, name, description, doc_type_code]` — o caller não pode configurar o role. Todo template nasce com a string literal `"approver"`, corrigível apenas por uma chamada posterior a `PUT /approval-config`. `SubmitForReview` relê esse role sem validar que resolve a um binding real → causa raiz documental do 403 do KNOWN FACT do seed.
- **Evidência:** `internal/modules/templates/delivery/http/routes_generated.go:57`; `internal/modules/templates/application/create.go:60, 76-81`; `api/openapi/v1/openapi.yaml:1139-1145` (sem `approver_role`); `internal/modules/templates/application/lifecycle.go:39-45` (relê sem validar).
- **Invariante violado:** "nothing hardcoded" (CLAUDE.md); contract-first; ADR 0022 — bindings de role são decisão sensível, não devem ser defaultados em código.
- **Irmão canônico:** `internal/modules/templates/delivery/http/routes_lifecycle.go:221-224` / `api/openapi/v1/openapi.yaml:5555-5559` (`UpsertTemplateApprovalConfig` lê `approver_role`/`reviewer_role` do body). **SaaS maduro:** recursos são criados com bindings de aprovação configuráveis pelo caller.
- **Fix (global-maximum):** adicionar `approver_role` (opcional, default server-side `"approver"`) + `reviewer_role` ao request body de `POST /templates` no OpenAPI, regenerar, passar pelo `CreateTemplateCmd`. Torna o contrato de criação auto-contido. Severidade efetiva rebaixada para low por um dos verificadores (override path existe e é authz-guarded), mas é bloqueador de UX/SoD: tratar como **medium**.

---

#### F-T5 · [LOW] `PublishTemplateVersion` chama `UpdateTemplateTx` duas vezes na mesma tx

- **Problema:** escreve `template` com `LatestVersion` antigo (linha 415), depois seta `LatestVersion = nextNum` (424) e regrava (425). Primeira escrita é redundante. **Não há bug de correção** — ambas estão dentro do mesmo `runner.Do`; erro/panic faz rollback completo, nada commitado. Defeito é round-trip desperdiçado + inconsistência com `Approve`.
- **Evidência:** `internal/modules/templates/application/lifecycle.go:405-428` (duas escritas) vs `:246-251` (Approve seta antes da tx, escreve uma vez).
- **Fix:** mover `template.LatestVersion = nextNum` para antes do `runner.Do` (espelhando Approve:251), remover a primeira `UpdateTemplateTx`. Uma escrita por entidade por tx.

---

#### F-T6 · [LOW] `spawnNextDraft` pré-tx Copy → órfão em rollback; sem GC; `area "*"` em tier-1

> Três achados `by_design`/low agregados aqui (copy órfão, sweeper ausente, area-mismatch tier-1).

- **Problema (copy órfão):** `spawnNextDraft` faz `s.presign.Copy` antes da tx; um rollback deixa objeto órfão permanente. O comentário (`lifecycle.go:451-456`) aceita explicitamente o resultado ("safe orphan"). Não há GC/sweeper de objetos de template (o sweeper de `documents` só cobre rows pending-upload). A corrida de spawn concorrente é prevenida pelo `UNIQUE(docx_storage_key)` da migração 0250.
- **Problema (area tier-1):** `routes_lifecycle.go:29, 69, 121, 182, 216` passam area `"*"` no tier-1 enquanto tier-2 usa `"tenant"`. Inócuo hoje porque o `AuthzFunc` wired (`main.go:773`) **ignora** o argumento de area (param `_`) e chama `CanDo` sem area; todas as `CapTemplate*` são `ScopeTenant` por ADR 0022. É lacuna de documentação, não de comportamento.
- **Evidência:** `internal/modules/templates/application/lifecycle.go:457-471`; `internal/modules/templates/application/keys.go:8-10`; `db/migrations/0250` (UNIQUE); `internal/modules/iam/domain/capability_scope.go:52-59`; `apps/api/cmd/metaldocs-api/main.go:773`.
- **Fix:** (a) [defer pós-v1] adicionar janitor de reconciliação em `internal/modules/jobs` que diffa `docx_storage_key` no DB vs objectstore e apaga órfãos > 24h — irmão: `internal/modules/documents/jobs/orphan_pending_sweeper.go`. (b) Adicionar comentário inline explicando que a area é intencionalmente ignorada para caps tenant-grade. Ambos low.

---

### 2.2 Templates — Frontend Editor / UX

#### F-FE1 · [HIGH] `useTemplateAutosave.flush()` engole falha de upload e não bloqueia o submit

- **Problema:** o resultado do PUT no S3 nunca é checado por `!ok`; o `catch` seta `status='error'` mas retorna `void` (não relança). `handleSubmitForReview` aguarda `flush()` e prossegue incondicionalmente para `submitForReview`, submetendo conteúdo antigo. Se a aba fechar na janela de 15s, edições são perdidas silenciosamente.
- **Evidência:** `frontend/apps/web/src/features/templates/hooks/useTemplateAutosave.ts:22-33`; `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:153-157` (prossegue após flush).
- **Irmão canônico:** `frontend/apps/web/src/features/documents/hooks/editor/useDocumentAutosave.ts:63-67` (checa `uploadRes.ok` e `throw`) + `:38-109` (flush retorna `Promise<boolean>`, caller bloqueia em `false`).
- **Fix:** (1) checar `if (!uploadRes.ok) throw ...` após o PUT; (2) `flush()` retornar `Promise<boolean>`; (3) `handleSubmitForReview` abortar em `false` com erro visível. Replicar `useDocumentAutosave`. (Correção: o `importDocx` no mesmo hook **relança** exceções mas não checa `uploadRes.ok` — corrigir o `.ok` também, ver F-FE7.)

---

#### F-FE2 · [HIGH] Conflito de lock otimista do schema (412) nunca chega ao usuário

- **Problema:** `useTemplateSchemas` expõe `staleConflict` setado em 412 `CONCURRENT_MODIFICATION`. `TemplateEditorPage` nunca lê o campo. O usuário não sabe que suas mudanças de schema foram rejeitadas; saves seguintes continuam do `lock_version` velho e falham em loop até reload manual.
- **Evidência:** `frontend/apps/web/src/features/templates/hooks/useTemplateSchemas.ts:16, 54-56, 71`; `TemplateEditorPage.tsx` — `schemaState.staleConflict` não referenciado em nenhuma linha (só `.schemas`, `.save`, `.error`). O comentário do save debounced (`TemplateEditorPage.tsx:129`) é enganoso ("hook surfaces error state").
- **Irmão canônico:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:323-329` (banner de retry de `docxError` já existe no mesmo arquivo).
- **Fix:** ler `schemaState.staleConflict`, exibir banner não-dispensável ("Outro usuário editou os metadados. Recarregue para continuar.") com botão que chama `draft.refetch()` (re-semeia o lock_version conforme JSDoc do hook). Espelhar o padrão `docxError`.

---

#### F-FE3 · [HIGH] Template readonly com blob ausente (404) renderiza editor em branco silenciosamente

- **Problema:** `useTemplateDraft` trata 404 do blob como "template em branco" **independente do status da versão**. Para `in_review`/`approved`/`published`, 404 é falha de integridade, não estado vazio. `MetalDocsEditor` em `mode='readonly'` com `documentBuffer=undefined` e sem `blankDocument` (só criado para modos não-readonly) renderiza canvas em branco sem `parseError` nem loading. Agravado pelo KNOWN FACT do eigenpal que faz loop/hang quando o docx está ausente.
- **Evidência:** `frontend/apps/web/src/features/templates/hooks/useTemplateDraft.ts:51-55` (404 sempre silencioso, sem branch por status); `node_modules/@metaldocs/editor-ui/src/MetalDocsEditor.tsx:174-177` (`blankDocument` só se `mode !== 'readonly'`); `TemplateEditorPage.tsx:335-336` (mode readonly p/ não-draft, `documentBuffer={draft.docxBytes ?? undefined}`).
- **Irmão canônico:** `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx:122-129` (toda falha de blob, inclusive 404, vira `editorLoadError`).
- **Fix:** condicionar o silêncio do 404 ao status: tratar 404 como "rascunho em branco" **apenas** quando `version.status === 'draft' && !version.content_hash`; caso contrário setar `docxError`. (Nota do verificador: a versão sempre tem `docx_storage_key`, então o guard correto é por status, não por presença da key.)

---

#### F-FE4 · [MEDIUM] Autosave de template sem estado `dirty` — sem feedback de edição não-salva

- **Problema:** união de status `'idle'|'saving'|'saved'|'error'` sem `'dirty'`. Durante os 15s de debounce o indicador mostra "Salvo". `queueDocx` nunca chama `setStatus('dirty')`. **Segundo defeito composto:** `TemplateEditorPage.tsx:218-222` mapeia status com fallback `else→'idle'` — mesmo se o hook emitisse `'dirty'`, seria colapsado.
- **Evidência:** `useTemplateAutosave.ts:4 (DEBOUNCE_MS=15_000), :14, :41-47`; `TemplateEditorPage.tsx:218-222, :337`; `AutosaveStatus.tsx:3-4,13-21,56-57` (componente já trata `'dirty'`→"Editado"+dot).
- **Irmão canônico:** `useDocumentAutosave.ts:117-123` (`setStatus('dirty')` no `queue`).
- **Fix:** adicionar `'dirty'` à união + `setStatus('dirty')` no topo de `queueDocx`, **e** estender o mapeamento em `TemplateEditorPage.tsx:218-222` com o branch `'dirty'`. (Considerar alinhar `DEBOUNCE_MS` a 3000 como documents — defer.)

---

#### F-FE5 · [MEDIUM] `VersionActionPanel` botões sem `type='button'` + labels em inglês

- **Problema:** 4 botões de ação como `<button>` puro sem `type='button'` (default HTML `submit` — submeteria form pai). Labels em inglês ("Approve Review", "Reviewer actions", "Published", etc.) contra a convenção pt-BR do produto.
- **Evidência:** `frontend/apps/web/src/features/templates/VersionActionPanel.tsx:90,98,127,135` (sem type), `:67,71,75,78,115` (labels EN); contraste `TemplateEditorPage.tsx:292-308` (type='button', pt-BR).
- **Irmão canônico:** `TemplateEditorPage.tsx:285-308`; `DocumentEditorPage.tsx:437`.
- **Fix:** `type='button'` nos 4 botões; traduzir labels e toasts para pt-BR.

---

#### F-FE6 · [LOW] `rejectGate = acceptGate` (alias by-design)

- **Problema:** `VersionActionPanel.tsx:69` aliasa o gate de reject ao de accept. Intencional pelo SoD atual (mesmo ator governa accept/reject no estágio), mas se o backend evoluir para um gate separado de reject, o alias concede silenciosamente o botão accept.
- **Evidência:** `VersionActionPanel.tsx:68-69`; `canActOnVersion.ts` (sem `canReject`; header do arquivo confirma "UX hint only — backend é a fronteira de autorização").
- **Fix:** adicionar comentário documentando a justificativa SoD na linha do alias. Sem mudança de código.

---

#### F-FE7 · [MEDIUM] `importDocx` não checa `uploadRes.ok`; `importTemplateDocx` standalone diverge (não é dead code)

- **Problema:** `useTemplateAutosave.importDocx` descarta o resultado do PUT sem `.ok` → upload 4xx prossegue para `commitAutosave` com hash sem objeto de backing. Existe um `importTemplateDocx` standalone que **checa** `.ok`. **Correção ao achado original:** o standalone NÃO é dead code — é importado e usado por `TemplateWizardPage.tsx:12,140`. Os dois servem call-sites diferentes (editor vs wizard).
- **Evidência:** `useTemplateAutosave.ts:54-60` (sem `.ok`); `templates.ts:195-200` (com `.ok`); `TemplateWizardPage.tsx:140` (usa o standalone).
- **Fix:** adicionar `if (!uploadRes.ok) throw new Error(...)` em `useTemplateAutosave.importDocx` espelhando `templates.ts:200`. **Não apagar** `importTemplateDocx`. Opcional: extrair util de upload compartilhada para eliminar a duplicação de `sha256Hex` (definido em ambos os arquivos).

---

### 2.3 Templates — Contrato / DTO Drift (FE ↔ OpenAPI)

> Cluster relacionado: a spec de `GET /templates` está incompleta e o FE compensa com tipos hand-written. Fix raiz único, consequências múltiplas.

#### F-C1 · [MEDIUM] `listTemplates` envia query params não-declarados (`limit/offset/doc_type`) a endpoint sem params

- **Problema:** `GET /templates` é declarado **sem** bloco `parameters:` (e marcado `x-pagination-exempt` com justificativa "catálogo em um shot" — que o handler contradiz, pois filtra). O FE monta `?limit&offset&doc_type` à mão; o handler lê os três de `r.URL.Query()`. Invisível a consumidores/validadores; `doc_type` filtering invisível a qualquer cliente que use tipos gerados.
- **Evidência:** `api/openapi/v1/openapi.yaml:1105-1124` (sem parameters); `frontend/apps/web/src/features/templates/api/templates.ts:122-132` (URLSearchParams manual); `internal/modules/templates/delivery/http/routes_query.go:31-57`; `routes_generated.go:16-20` (ListTemplates sem Params struct, diferente de CreateTemplate).
- **Irmão canônico:** `frontend/apps/web/src/features/documents/api/library.ts:10-12, 30-39` (usa `operations['listDocuments']['parameters']['query']` + `api.GET`).
- **Fix:** adicionar `parameters:` (`limit` int default 50 max 200, `offset` int default 0, `doc_type` string opcional) ao spec, regenerar. Alternativamente, honrar a isenção e remover o filtering server-side. Decisão no nível da spec.

#### F-C2 · [MEDIUM] `listTemplates` usa tipo anônimo hand-written em vez do DTO gerado

- **Problema:** anota o retorno como `{ data?: { templates?: unknown; items?: unknown }; meta?: ... }` — todos opcionais e `unknown`, desconectando do `ListTemplatesResponse` gerado (`data.templates: TemplateDTO[]` required). Cast a `TemplateDTO[]` não-checado.
- **Evidência:** `templates.ts:129-132`; `frontend/apps/web/src/lib/api-types/index.d.ts:2827-2835`; `client.ts:130` (`api = createClient<paths>` já instanciado).
- **Fix:** migrar para `api.GET('/templates', { params: { query } })` tipado (após F-C1 registrar os params). Irmão: `library.ts:30-39`.

#### F-C3 · [LOW] `listTemplates` fallback a `data.items` (dead code, mascara regressão)

- **Problema:** fallback `data.templates ?? data.items` — `data.items` nunca é emitido por este endpoint (servidor sempre `data.templates`, `routes_query.go:75`). Retorna `[]` silencioso em vez de erro num mismatch futuro.
- **Evidência:** `templates.ts:134-140`; `internal/modules/templates/api/api.gen.go:124-133`; `openapi.yaml:5389-5406`.
- **Fix:** remover o fallback ao migrar para o tipo gerado (F-C2).

#### F-C4 · [LOW] `TemplateDTO`/`VersionDTO` locais com `Omit`+re-declare; `TemplateListRow` duplicado

> Dois achados low agregados. **Correção:** um verificador refutou a causa proposta (não é bug de config openapi-typescript — o gerador já emite `T | null`).

- **Problema real:** os wrappers `Omit` promovem campos opcionais a required (drop do `?`) e estreitam `placeholder_schema` a `WirePlaceholder[]`. Obrigação de sync manual com a spec, mas **com rede de segurança em compile-time** (mudança de tipo Omit'd dá erro de compilação, não drift silencioso). `TemplateListRow` é subset hand-written consumido por `taxonomy/ProfileEditDialog.tsx:4,25`.
- **Evidência:** `templates.ts:14-58, 66-77`; `index.d.ts:2836-2909`; `documents/api/documents.ts:8-24` (padrão validado sem Omit).
- **Fix:** [low/defer] para `TemplateListRow`, substituir por `TemplateDTO` direto em `ProfileEditDialog.tsx` (todos os campos acessados existem). Os wrappers `Omit` permanecem aceitáveis até os campos serem marcados `required` na spec.

#### F-C5 · [LOW] `commitAutosave` de template sem `requestBody` declarado na spec

- **Problema:** `POST /templates/{id}/versions/{n}/autosave/commit` sem `requestBody` no OpenAPI; handler lê `{ expected_content_hash }` via `readJSON` (não strict). Body é conhecimento de implementação invisível ao contrato.
- **Evidência:** `openapi.yaml:1417-1442`; `routes_autosave.go:68-73`; `templates.ts:177-180`; irmão `openapi.yaml:2714-2730` (documents autosave/commit declara body).
- **Fix:** adicionar `requestBody` required `expected_content_hash:string`, regenerar, trocar para `readStrictJSON`.

---

### 2.4 Documents — Instâncias (delivery/http + repository + application)

#### F-D1 · [HIGH] `authorizeDocumentScope` usa ownership (`IsDocumentOwner`) em vez de `authz.Require` em TODAS as rotas de leitura

- **Problema:** o helper concede acesso a não-admins via `IsDocumentOwner` (predicado de ownership SQL bruto), não via capability nomeada. 19 call-sites (getDocument, listCheckpoints, listRevisionHistory, autosave, comments, duplicate/rename/archive, etc.) burlam o PDP two-tier do ADR 0022. Consequências concretas: (a) usuário com `CapDocumentView` mas não-criador é rejeitado; (b) owner com capability revogada é aceito.
- **Evidência:** `internal/modules/documents/delivery/http/handler.go:1279-1306` (sem `SeedTxIdentity`/`Require`; `IsDocumentOwner` em `repository.go:1505`); 19 sites em `:391,495,...,1231`; `isSystemAdmin` (`:131-138`) consulta `iam_user_roles` (atalho admin permitido por ADR 0022, não é nova violação). Contraste correto: `view_service.go:76` (`authz.Require(CapDocumentView, "tenant")`).
- **Invariante violado:** ADR 0022 — AuthZ = capabilities, nunca ownership; tier-2 `authz.Require`.
- **Irmão canônico:** `internal/modules/documents/application/fillin_authz.go:21-40` (`requireDocEditDraft`: `runner.Do` + `SeedTxIdentity` + `authz.Require`); `view_service.go:76`.
- **Fix:** substituir o branch `IsDocumentOwner` por `SeedTxIdentity` + `authz.Require(CapDocumentView, areaCode)` numa tx RW curta. Manter `IsDocumentOwner` no máximo como tripwire DB atrás do gate de capability.

#### F-D2 · [HIGH] `CreateCheckpoint`/`ListCheckpoints` sem `authz.SeedTxIdentity` + `authz.Require`

> Severidade elevada de medium→high pelo verificador (escrita de estado em registro de documento controlado).

- **Problema:** `CreateCheckpoint` abre tx, lock FOR UPDATE, INSERT — **sem** authz. `ListCheckpoints` é `QueryContext` puro sem authz. `CommitUpload` no mesmo arquivo faz `SeedTxIdentity`+`Require(CapDocumentEdit)`. O handler chega via `authorizeDocumentScope` (ownership-only, F-D1) → não há tier-2 no lifecycle de checkpoint. O service layer (`service.go:772-782`) é passthrough trivial.
- **Evidência:** `internal/modules/documents/repository/repository.go:1281-1318, 1320-1339`; `CommitUpload :1017-1031` (padrão correto).
- **Irmão canônico:** `repository.go` `CommitUpload`; `fillin_authz.go` `requireDocEditDraft`.
- **Fix:** `SeedTxIdentity`+`authz.Require(CapDocumentEdit, area)` dentro da tx de `CreateCheckpoint`. Para `ListCheckpoints`, gate de read-capability na camada de application via `TxRunner` (não dentro de `QueryContext` bruto).

#### F-D3 · [MEDIUM] `validateValue` PHUser falha-aberto quando IAM reader não wired

- **Problema:** `validateValue` faz `return nil` quando `iam == nil` para placeholder `PHUser`. `NewFillInService` não exige IAM reader (opcional via `WithIAMReader`), e a composition root **nunca chama** `WithIAMReader` (apesar de `deps.IAMUserOptions` existir e ser usado p/ `PlaceholderOptionsHandler`). Em produção, todo placeholder PHUser aceita qualquer string. Não é bypass de authz (edit ainda gateado), mas é bypass de validação de integridade.
- **Evidência:** `internal/modules/documents/application/fillin_service.go:214-218, 45-46`; `module.go:118-120` (chain sem `WithIAMReader`), `:127` (IAMUserOptions usado p/ outro handler).
- **Fix:** (1) adicionar `.WithIAMReader(...)` na chain de `fillInSvc` em `module.go`; (2) mudar o branch para fail-closed: `if iam == nil { return fmt.Errorf("IAM reader not wired: cannot validate PHUser") }`; idealmente tornar o IAM reader parâmetro obrigatório de `NewFillInService`. Irmão: `fillin_authz.go` `requireDocEditDraft` (área nil → fail-closed).

#### F-D4 · [MEDIUM] `SubmitHandler` deriva idempotency key do relógio (segundo); sem header de cliente

> Severidade ajustada de high→medium (a spec **não** declara o header nesta rota; UNIQUE no DB é backstop parcial; irmãos publish/cancel também não têm replay-store).

- **Problema:** `SubmitHandler` lê `If-Match` mas não lê/exige `Idempotency-Key`. `SubmitRevisionForReview` deriva a key via `ComputeIdempotencyKey` truncando o relógio ao segundo. Retry >1s gera key distinta → submissão duplicada (`approval_instance` duplicada), burlando o UNIQUE do DB.
- **Evidência:** `internal/modules/documents/approval/http/submit_handler.go` (sem header); `application/submit_service.go:68-76`; `application/idempotency.go:28` (`Truncate(time.Second)`); `openapi.yaml:3207-3237` (só `id` + `If-Match`).
- **Irmão canônico:** `signoff_handler.go:68` (`BeginStageReplay`, key de cliente).
- **Fix:** ler/exigir `Idempotency-Key` no `SubmitHandler` (padrão `signoff_handler.go`), enfiar a key de cliente em `SubmitRevisionForReview` substituindo a derivação por relógio. Coordenar com F-CD2 (declarar o header na spec).

#### F-D5 · [LOW] `LoadPlaceholderSchema` param `revisionID` mas consulta `documents.id`

- **Problema:** param nomeado `revisionID` mas a SQL é `WHERE id=$2` em `documents` (documentID). Mismatch nome↔semântica. **Sem corrupção hoje** (toda a cadeia passa documentID; SQL semanticamente correta), risco é confusão de manutentor futuro.
- **Evidência:** `fillin_service.go:20-21, 70-76, 108`; irmão correto `TemplateVersionSchemaReader.LoadFillInSchema:255-278` (usa `docID`).
- **Fix:** renomear param para `documentID`/`docID` na interface e impl + call-sites. Sem mudança de SQL.

#### F-D6 · [LOW] Hash sintético (sha256 da string da key) na criação do documento — by_design

- **Problema:** `cloneIntoTx` semeia `content_hash` da revisão inicial como `sha256(docxKey)`, não dos bytes. Não é content-addressable; `RestoreCheckpoint` dedup por `content_hash` não funciona para a revisão de criação. Risco de colisão astronômico.
- **Evidência:** `service.go:335-337`; `repository.go:1461-1471` (ON CONFLICT dedup).
- **Fix:** [low/defer] hashear bytes reais do objeto na criação (padrão `spawnNextDraft`: copy pré-tx, depois hash), OU `content_hash` NULL como sentinela "não materializado".

---

### 2.5 Controlled Documents — Signoff / Module Boundaries

#### F-CD1 · [HIGH] Cinco handlers de approval de documentos exigem `Idempotency-Key` que o OpenAPI não declara

> **Deduplicação:** dois achados convergentes (`controlleddocuments-signoff` + `idempotency-and-lifecycle`). Evidência mesclada.

- **Problema:** `publish`, `schedule-publish`, `supersede`, `obsolete`, `cancel` (`/documents/{id}/...`) exigem o header no Go (`ErrIdempotencyRequired`/400) mas a spec só declara o path param `id`. SDK gerado pela spec recebe 400 inesperado; testes de contrato não verificam o enforcement. O FE só escapa porque `mutationClient.ts:30` gera UUID fallback universal.
- **Evidência:** `internal/modules/documents/approval/http/publish_handler.go:40-44, 88-92`; `obsolete_handler.go:30-34`; `supersede_handler.go:30-34`; `cancel_handler.go:30-34` (e `doc_approval_handler.go:214-218`, CancelByDocument); `openapi.yaml:3284,3313,3342,3371,3392/3503` (sem header). Precedente correto: `openapi.yaml:3481` (`recordDocumentSignoff` declara o header).
- **Invariante violado:** contract-first (spec é a verdade da rota).
- **Fix:** declarar `Idempotency-Key (header, required:true)` nas 5 operações no OpenAPI e regenerar; OU migrar os 5 handlers para o middleware `idempotency.Require()` (que valida UUID + dá replay) como as rotas de template fazem (`handler.go:68-74`). Irmão: `openapi.yaml:3481`.

#### F-CD2 · [HIGH] `Idempotency-Key` validado mas **descartado** em publish/schedule/obsolete/cancel — sem replay store

- **Problema:** os handlers leem a key, rejeitam se vazia, e descartam. Nenhum store consultado, nenhum replay-handle aberto. Retry re-executa a transição (risco de double-publish/double-obsolete sob erro de rede transiente). O signoff handler faz `BeginDocumentReplay`/`Complete`; estes não. (CancelByDocument tem a mesma lacuna; `SubmitHandler` nem lê o header — ver F-D4.) Mitigação parcial acidental via If-Match OCC, mas não é semântica de idempotência.
- **Evidência:** `publish_handler.go:40-44`; `cancel_handler.go:30-34`; `obsolete_handler.go:30-34`; `doc_approval_handler.go:138` (`BeginDocumentReplay`, só no signoff).
- **Invariante violado:** Async = outbox transacional; Idempotency-Key deve respaldar um replay store.
- **Irmão canônico:** `doc_approval_handler.go:138` (`BeginDocumentReplay`); `signoff_handler.go:68` (`BeginStageReplay`). **SaaS maduro (Stripe):** todo endpoint mutante armazena o outcome keyed em `(actor, key, request-hash)` e replaya no retry.
- **Fix:** abrir replay slot (`BeginStageReplay` ou novo `BeginLifecycleReplay`) em cada um dos 5 handlers; checar o registro antes de executar; `Complete`/`Fail` ao fim. (Migrar para o middleware `idempotency.Require()` resolve F-CD1 + F-CD2 + F-CD3 de uma vez.)

#### F-CD3 · [MEDIUM] Handlers de approval validam só presença, não formato UUID, do `Idempotency-Key`

- **Problema:** os 6 handlers de approval (publish/obsolete/supersede/cancel/signoff-by-document/cancel-by-document) checam só `!= ""`, não `IsValidKey()`. O middleware de plataforma valida UUID (`CodeIdempotencyKeyInvalid`/400). Uma string arbitrária ("retry-1") recebe 200 nas rotas de approval mas 400 nas de template para a mesma key malformada — contrato não-uniforme.
- **Evidência:** `doc_approval_handler.go:94-99`; demais handlers `:30-32`; `internal/platform/idempotency/middleware.go:22-24, 96-99`.
- **Irmão canônico:** `middleware.go:22-24` (`IsValidKey`) wired via `templates/delivery/http/handler.go:68-74`.
- **Fix:** migrar para `idempotency.Require()` (fix unificado com F-CD1/F-CD2) OU adicionar `IsValidKey` após o guard de vazio.

#### F-CD4 · [MEDIUM] `controlleddocuments/module.go` instancia infra de módulos irmãos (documents/templates/taxonomy)

> **Deduplicação:** quatro achados de fronteira de módulo, mesmo padrão (wiring no construtor em vez da composition root). Agregados.

- **Problema:** o construtor do módulo CD importa e instancia adaptadores concretos de infra de irmãos: `docrepo.NewActiveInstanceReaderPG` (`module.go:12,38`), `taxonomyinfra.NewProfileRepository`/`NewAreaRepository` + `taxonomyapp.NewAuditGovernanceAdapter` (`:13-14,48-50`), `templatesinfra.NewTemplateVersionReader` (`:15,43`). `Dependencies` (`:25-29`) só expõe DB/Logger/AuditWriter — sem pontos de injeção. Isolamento de módulo intestável sem schema real dos irmãos.
- **Evidência:** `internal/modules/controlleddocuments/module.go:12-15, 38, 43, 48-50`; composition root `apps/api/cmd/metaldocs-api/main.go:746-751` (passa só DB/Logger/AuditWriter).
- **Invariante violado:** acesso cross-module só via service/interface publicada; façade aceita colaboradores interface-typed, não constrói infra de irmão.
- **Irmão canônico:** `internal/modules/documents/module.go:61-65` (`CDFieldReader`/`AreaCatalogReader` injetados); `internal/modules/taxonomy/module.go:17-28` (`TplChecker` injetado + panic-on-nil) + `main.go:740`.
- **Fix:** adicionar a `controlleddocuments.Dependencies` os ports `ActiveInstanceReader documentsdomain.ActiveInstanceReader`, `ProfileReader`/`AreaReader`/`GovernanceLogger`, `TemplateVersionChecker`. A composition root wira os concretos. Nil-guard com Noop como os outros ports.

#### F-CD5 · [MEDIUM] `controlleddocuments/domain` importa `templates/domain` para pinar `VersionStatus`

> Severidade ajustada high→medium (valores são também enums DB, dupla-estáveis; rename improvável).

- **Problema:** `resolution.go:6,46,59,62` importa `templatesdomain` e compara contra `VersionStatusPublished/Obsolete` na camada **domain** de outro contexto. Domain de um BC não deve depender de domain de irmão.
- **Evidência:** `internal/modules/controlleddocuments/domain/resolution.go:6,46,59,62`; `resolution_test.go:7` (7 dos 8 testes já usam literais "published"/"obsolete").
- **Irmão canônico:** `internal/modules/documents/application/version_status_test.go:13-17` (pina o valor-de-wire num parity test em `_test.go`, sem importar o tipo no domain).
- **Fix:** definir constantes locais em `controlleddocuments/domain` (`const tplVersionPublished = "published"`) + parity test que importa `templatesdomain` só de `_test.go`.

#### F-CD6 · [MEDIUM] Double `authz.Require` no caminho de create do CD sem `WithCapCache`

- **Problema:** create CD (manual e auto-code) chama `authz.Require` no service (`:306/:350`) e de novo em `CreateTx` (`:373`) na mesma tx, mesma cap/area, sem `WithCapCache` → dobra o roundtrip DB. Não é defeito de correção/segurança (ambos concordam sempre), é anti-pattern de performance + ownership confuso da checagem.
- **Evidência:** `internal/modules/controlleddocuments/application/service.go:306,350`; `infrastructure/repository.go:373`; sem `WithCapCache` em todo o módulo.
- **Irmão canônico:** `internal/modules/documents/approval/application/decision_service.go:198` (`WithCapCache`); módulo documents mantém authz só na camada application.
- **Fix (global-maximum):** remover `authz.Require` de `CreateTx` e documentar que o service é o gate authz mandatório. Alternativa cheap: `authz.WithCapCache(ctx)` antes do `runner.Do`.

#### F-CD7 · [MEDIUM] `UpsertApprovalConfig` gateia por nomes de role IAM, não capabilities

> Severidade ajustada high→medium (há `authz.Require(CapTemplateEdit)` válido in-tx logo abaixo; o role-check é guard redundante → risco de falso-negativo, não escalonamento).

- **Problema:** computa `isOperator` via `containsRole('system_admin'||'qms_admin')` para decidir se pode mudar config de template publicado — raciocínio role-string explícito (ADR 0022: "nunca raciocine admin/author/editor pode X").
- **Evidência:** `internal/modules/templates/application/approval_config.go:36-46`; mas `:77` faz `authz.Require(CapTemplateEdit, "tenant")` in-tx (gate real).
- **Irmão canônico:** `internal/modules/documents/delivery/http/handler.go:137` (`caps.IsSystemAdmin`, capability-derived).
- **Fix:** substituir o bloco `isOperator` por `CanDo(CapTemplateManage)` (nova cap dedicada) antes da tx, ou um segundo `authz.Require` com cap distinta in-tx.

#### F-CD8 · [LOW] CD create governance loga post-commit (best-effort) vs `changeStatus` in-tx (atômico)

- **Problema:** o evento `template.override` (caminho auto-code com override) é logado post-commit `//cilint:allow-post-commit-audit`; `changeStatus` usa `LogTx` in-tx. Asimetria de durabilidade de auditoria. **Correção:** não é o caminho "manual-code" (esse não emite evento); rationale documentada (multi-leg create sem tx externa única, restrição H-PRE-1 advisory-lock). by_design.
- **Evidência:** `service.go:~466` (Log post-commit), `:682` (LogTx).
- **Fix:** [low/defer] se possível, mover o evento para dentro da tx de sequência e usar `LogTx`; senão manter a exceção documentada.

#### F-CD9 · [LOW] Dupla extração de actor-ID em `changeStatus` governance

- **Problema:** `authn.UserIDFromContext` (guarded, `:621`) para authz vs `authdomain.CurrentUserFromContext` (não-guarded, `:679`) para `ActorUserID` do evento — em `!ok` o actorID vira `""`. **Sem divergência em produção** (middleware seta ambas as keys atomicamente, `auth/delivery/http/middleware.go:93-94`); unit tests sempre logam `""`.
- **Evidência:** `internal/modules/controlleddocuments/application/service.go:621, 678-681`.
- **Fix:** reusar `actorUserID` validado de `:621` no evento; remover o default `""`. One-liner.

---

### 2.6 Render — Pipeline / Fanout / Staging Outbox

#### F-R1 · [MEDIUM] `FreezeService.Freeze` faz HTTP ao docx-renderer dentro da tx aberta — dead code wired

> Severidade ajustada high→medium (em produção o branch é inalcançável: `recordSignoff` checa `pinInvoker != nil` primeiro, e `main.go:519` sempre seta `WithPinInvoker`). Defeito real = dead code que viola o invariante se re-wired + acoplamento mantido.

- **Problema:** `Freeze()` chama `s.fanout.Fanout()` (HTTP até 60s) com a tx do caller aberta. Viola "Async = outbox: tx de state-write nunca compartilha com network call". Superseded por Pin+Materialize (ADR 0015) — o próprio comentário do método diz isso. O construtor `NewDecisionService` ainda exige `FreezeInvoker`; o teste sp2 exercita `Freeze()` diretamente, preservando o dead code como "testado".
- **Evidência:** `internal/modules/documents/application/freeze_service.go:348` (Fanout in-tx); `decision_service.go:396-412` (branch else do guard pinInvoker); `main.go:517-521` (WithPinInvoker sempre); `sp2_dictionary_substitution_integration_test.go:167`.
- **Irmão canônico:** `freeze_service.go:188` (`Pin` — tx-only, zero network).
- **Fix:** deletar `Freeze()`; remover `FreezeInvoker` da assinatura de `NewDecisionService`; deletar o scaffolding de teste sp2 que chama `Freeze` direto.

#### F-R2 · [MEDIUM] `StagingOutboxWorker` hardcoda todos os parâmetros operacionais

- **Problema:** `pollEvery=5s, batchSize=10, maxAttempt=5, staleAfter=5m` literais, sem override env/construtor. O worker principal (`WorkerConfig`) expõe equivalentes via `METALDOCS_WORKER_*`. Roda dentro do processo da API; tuning exige code+deploy.
- **Evidência:** `internal/modules/render/fanout/staging_outbox_worker.go:44-45`; `internal/platform/config/worker.go:21-28` (padrão env); `main.go:862-895` (sem config injetada).
- **Irmão canônico:** `internal/platform/config/worker.go:21` (`WorkerConfig` + `LoadWorkerConfig`).
- **Fix:** `StagingOutboxWorkerConfig` espelhando `WorkerConfig`, carregada do env no bootstrap, passada a `NewStagingOutboxWorker`.

#### F-R3 · [MEDIUM] Staging outbox sem DLQ, sem alerta, sem visibilidade de rows permanentemente falhas

- **Problema:** após `maxAttempt`, seta `status='failed'` + log Error; sem DLQ, sem métrica, sem `dead_lettered_at`, sem API admin. Um materialize permanentemente falho = revisão congelada sem `final_docx_s3_key`/PDF, invisível sem query DB direta. O consumer de outbox principal escreve `dead_lettered_at` em `outbox_events` (tem observabilidade).
- **Evidência:** `staging_outbox_worker.go:94-98`; `staging_outbox.go:123-128` (MarkFailed sem DLQ); `internal/platform/messaging/outbox/postgres/consumer.go:152-177` (`dead_lettered_at`); grep DLQ/metric em `internal/modules/render` = 0.
- **Irmão canônico:** `consumer.go:152` (`dead_lettered_at` em `outbox_events`). **SaaS maduro:** DLQ (SQS/Kafka) torna mensagens permanentemente falhas visíveis e re-drivable.
- **Fix:** migração adicionando `dead_lettered_at` a `pdf_dispatch_outbox`/`materialize_dispatch_outbox`; setar no finalize; expor counter Prometheus / query de observabilidade.

#### F-R4 · [LOW] `StagingOutboxWorker.dispatchOne`: Publish OK + MarkDispatched fail deixa row em `processing` silenciosamente

- **Problema:** se `Publish` OK mas `MarkDispatched` falha, só loga Error; sem `MarkFailed` → `last_error` não atualizado, `attempts` não incrementado; row fica invisível em `processing` até `staleAfter` (5min). Eventualmente seguro (ON CONFLICT DO NOTHING no re-publish), mas estado intermediário invisível.
- **Evidência:** `staging_outbox_worker.go:80-86, 91` (MarkFailed existe no branch de Publish-fail).
- **Fix:** chamar `MarkFailed` no branch de `MarkDispatched`-fail (simétrico à linha 91).

#### F-R5 · [LOW] `fanout.NewClient` fallback a `http.DefaultClient` (sem timeout) quando `h == nil`

- **Problema:** nil guard instala client sem timeout (hang indefinido possível). Produção segura (`main.go:839` sempre passa `httpclient.NewInternalClient()`, 60s); risco latente p/ caller/teste futuro. Auditoria 2026-05-04 já flagou; fix aplicado no call-site, não no construtor.
- **Evidência:** `internal/modules/render/fanout/client.go:53-57`; `internal/platform/httpclient/internal_client.go:27`; `wiki/bugs/audit-2026-05-04.md:64-77`.
- **Irmão canônico:** `internal/platform/render/gotenberg/client.go:39-47` (sempre constrói client com timeout 30s, sem nil-path).
- **Fix:** remover o nil guard; exigir `*http.Client` não-nil (panic/error em nil = programmer error).

#### F-R6 · [LOW] `frozenDocxKey` hardcoded duplicado: renderer TS + fallback Go

- **Problema:** a key do frozen docx é computada em `apps/docx-renderer/src/routes/fanout.ts:25-26` e como fallback em `internal/platform/worker/pdf_job_runner.go:78`. `FinalDocxS3Key` é `omitempty` (`events.go:24`) → mensagem replay/legacy sem o campo cai no fallback hardcoded. Divergência silenciosa se a fórmula TS mudar.
- **Evidência:** `fanout.ts:25-27`; `pdf_job_runner.go:77-79`; `events.go:24`; irmão `pdf_dispatcher.go:39` (sempre popula `FinalDocxS3Key`).
- **Fix:** tornar `FinalDocxS3Key` required (dropar `omitempty`, validar não-vazio, erro em vez de adivinhar) e deletar o fallback.

---

### 2.7 Object Store — Kernel (`VerifiedStore` + callers)

> Cluster de tenant-guard: o kernel guarda o destino mas não a origem/leitura, com um caminho de webhook (atualmente unwired) que amplifica o risco.

#### F-O1 · [MEDIUM] `VerifiedStore.Copy` não assere prefixo de tenant na `srcKey`

> by_design (comentado). Severidade ajustada high→medium (sem caminho exploitável atual: o único caller `spawnNextDraft` recebe `source.DocxStorageKey` de `GetVersion`, que JOINa em `tenant_id`; UUID PK impede compartilhamento cross-tenant de key).

- **Problema:** `Copy` guarda `dstKey` via `assertTenant` mas passa `srcKey` direto a `CopyObject`. Comentário (`:106-108`) documenta src como "DB-sourced / não-guarded". Defense-in-depth ausente para caller futuro.
- **Evidência:** `internal/platform/objectstore/verified_store.go:109-124, 106-108`; caller `lifecycle.go:459`; `repository/postgres.go:259-278` (GetVersion JOIN tenant).
- **Irmão canônico:** `verified_store.go:46` (`assertTenant`). **AWS S3:** server-side copy exige permissão na origem; aplicar o mesmo escopo de tenant na src é o controle equivalente in-process.
- **Fix:** assertar `srcKey` contra `tenants/{tenantID}/` em `Copy`, com exceção explícita para keys do tenant de sistema (`ffffffff-...`), não por omissão da checagem.

#### F-O2 · [MEDIUM] `pdf_webhook_handler.isValidFinalPDFS3Key` não exige prefixo de tenant

> Severidade ajustada high→medium (rota **unwired** — `RegisterRoutes` nunca chamado; nota H-1e na linha 26; zero exposição runtime hoje). Defeito latente pré-wire.

- **Problema:** o webhook HMAC aceita `final_pdf_s3_key` que só rejeita `..` e `\x00`. Worker comprometido/forja de key escreveria key arbitrária em `documents.final_pdf_s3_key`, depois passada a `PresignGet` → URL presigned cross-tenant.
- **Evidência:** `internal/modules/documents/delivery/http/pdf_webhook_handler.go:26 (UNWIRED), :149-155`; `view_service.go:84-86`; `snapshot_repository.go:154-165` (WritePDF grava verbatim).
- **Irmão canônico:** `verified_store.go:46-51` (`assertTenant`).
- **Fix:** **antes de wirar**, adicionar `strings.HasPrefix(trimmed, "tenants/"+canonicalTenantID+"/")` (o `canonicalTenantID` está disponível via `ResolveTenantByDocumentID`, `:91`) — validação completa tenant-bound.

#### F-O3 · [LOW] `PresignGet` sem tenant-guard (amplificador de F-O2)

> by_design; severidade ajustada medium→low (caminho de poluição primário dormant via F-O2 unwired).

- **Problema:** `PresignGet` não assere prefixo (`:126` "NOT guarded: DB-sourced"). Risco sistêmico se uma key poluída chega ao DB.
- **Evidência:** `verified_store.go:126-134`; `view_service.go:60-85`; callers adicionais `export_service.go:125-147`, `templates/application/queries.go:57`.
- **Irmão canônico:** `verified_store.go:55-63` (`PresignPut` recebe tenantID + `assertTenant`).
- **Fix:** corrigir F-O2 fecha a lacuna primária. Defense-in-depth: variante `AssertedPresignGet(ctx, tenantID, key, ttl)` usada pelo view service.

#### F-O4 · [MEDIUM] `document_exports` sem coluna `tenant_id` — invariante multi-tenant violado no schema

> Severidade ajustada critical→high→medium na cadeia de verificação (PK uuid de `document_id` bloqueia colisão cross-tenant prática). Listado medium pelo achado mais conservador; tratar como **high** por invariante quebrado + RLS ausente.

- **Problema:** `document_exports` (colunas reais: `id, document_id, revision_id, composite_hash, storage_key, size_bytes, ...`) sem `tenant_id`, ausente do RLS 0237. `GetExportByHash`/`InsertExport` sem predicado tenant. Violação do invariante "toda tabela tenant tem tenant_id"; sem defense-in-depth RLS. (Snippet do achado original citava colunas stale `content_hash`/`pdf_s3_key`.)
- **Evidência:** `db/baseline/0001_current_schema.sql:1906-1919`; `db/migrations/0237_rls_all_tenant_tables.sql` (ausente); `internal/modules/documents/repository/export_repository.go`.
- **Irmão canônico:** `internal/modules/documents/repository/repository.go` (`CreateDocumentTx` enfia `tenant_id` em toda escrita); `document_comments:1514-1528`.
- **Fix:** migração: `ADD COLUMN tenant_id uuid NOT NULL REFERENCES tenants(id)`; `UNIQUE(tenant_id, document_id, composite_hash)`; habilitar RLS (padrão 0237); incluir `tenant_id` em `InsertExport`/`GetExportByHash`.

#### F-O5 · [MEDIUM] `export_service.go` constrói key PDF com `fmt.Sprintf` inline em vez de helper

- **Problema:** `ExportService.ExportPDF` monta a key inline (`:77`) em vez de helper dedicado, criando segundo site ad-hoc no mesmo módulo (que já centraliza `documentRevisionKey`).
- **Evidência:** `internal/modules/documents/application/export_service.go:77`; `internal/modules/documents/application/keys.go:8-10`.
- **Irmão canônico:** `keys.go:8` (`documentRevisionKey`).
- **Fix:** adicionar `documentExportKey(tenantID, documentID string, compositeHash []byte) string` a `keys.go`, substituir o inline.

#### F-O6 · [MEDIUM] `TemplateReader` (docgenv2) usa `minio.Client` cru, limite de tamanho hardcoded, sem helper de key

- **Problema:** `GetPublishedVersion` faz `t.client.GetObject` direto (bypass do `VerifiedStore`), com `maxSchemaBytes=1MiB` hardcoded (independente do `defaultMaxObjectBytes=25MiB` configurável). Sem `assertTenant` Go-level (mitigado pelo JOIN tenant na SQL). Caminho de leitura divergente do kernel.
- **Evidência:** `internal/platform/docgenv2/template_reader.go:44-63, 55`; `internal/platform/objectstore/verified_store.go:16`.
- **Irmão canônico:** `verified_store.go:68-100` (`Confirm` — read + size-limit + hash).
- **Fix:** adicionar `ReadObject(ctx, key, maxBytes) ([]byte, error)` ao `VerifiedStore` e usar em `TemplateReader`; limite passado pelo caller.

#### F-O7 · [MEDIUM] Region `"us-east-1"` hardcoded em dois bootstraps

- **Problema:** `Region: "us-east-1"` em `api.go:127,136` e `worker.go:89`. Em S3 não-US/MinIO regional → assinatura V4 errada → `SignatureDoesNotMatch` em todo presigned PUT/GET. `AttachmentsConfig` já lê os demais campos MinIO do env; region é o único faltante.
- **Evidência:** `internal/platform/bootstrap/api.go:127,136`; `internal/platform/bootstrap/worker.go:89`; `internal/platform/config/attachments.go:27-40,88-103`.
- **Irmão canônico:** `attachments.go:89-95` (padrão `METALDOCS_MINIO_*`).
- **Fix:** `MinIORegion` em `AttachmentsConfig` (env `METALDOCS_MINIO_REGION`, default `"us-east-1"`); substituir os 3 literais.

#### F-O8 · [LOW] `pdf_job_runner` monta keys com `fmt.Sprintf` cru (sem helper)

> Severidade ajustada high→low (`tenants/%s/revisions/%s/frozen.docx` **É** a key canônica do renderer, não um namespace errado; o achado original confundiu objetos distintos). Defeito real = falta de helper centralizado.

- **Problema:** `:78` (docx fallback) e `:81` (output PDF) são `fmt.Sprintf` inline sem helper. Hygiene, não locatabilidade.
- **Evidência:** `internal/platform/worker/pdf_job_runner.go:77-81`; renderer `fanout.ts:25-26`.
- **Fix:** extrair `internal/platform/worker/keys.go` com `workerFrozenDocxKey`/`workerPDFKey`; combinar com F-R6.

#### F-O9 · [LOW] Copy-on-spawn sem GC (ver F-T6) — agregado em F-T6.

---

### 2.8 Banco de Dados — Constraints / Migração / Multi-tenant

> A maior densidade de invariantes delegados-ao-app está em `templates_template_version`. Migração 0250 está correta; a baseline está stale.

#### F-DB1 · [HIGH] `templates_template_version.status` sem CHECK constraint

- **Problema:** `status text NOT NULL` sem CHECK. Domain define 5 valores; DB aceita qualquer string. Literal errado em `UpdateVersionTx` ou escrita SQL direta corrompe o FSM sem rejeição DB. A tabela legacy `template_versions` (`:2120`) e `documents` (`:2030`) têm o CHECK.
- **Evidência:** `db/baseline/0001_current_schema.sql:2214-2235` (sem CHECK); `internal/modules/templates/domain/version.go:11-15`; siblings `:2120, :2030`.
- **Invariante violado:** "DB enforces invariants; app checks são só a primeira linha amigável".
- **Fix:** `ALTER TABLE public.templates_template_version ADD CONSTRAINT chk_template_version_status CHECK (status = ANY (ARRAY['draft','in_review','approved','published','obsolete']))`. Espelhar `:2120`.

#### F-DB2 · [HIGH] `templates_template_version` sem partial-unique "one-published-per-template"

- **Problema:** sem `UNIQUE (template_id) WHERE status='published'`. Invariante (ADR 0013) só no app (`ObsoletePreviousPublishedTx` antes do set). Sob READ COMMITTED, dois publish concorrentes podem ler zero publicadas, ambos prosseguir → duas versões publicadas, corrompendo `published_version_id`. `ObsoletePreviousPublishedTx` é UPDATE simples (`postgres.go:509-522`).
- **Evidência:** `db/baseline:3527` (índice não-unique); `:3485` (unique mas na tabela legacy `template_versions`); `lifecycle.go:264`.
- **Irmão canônico:** `db/baseline:3625` (`ux_approval_instances_active_document_id`); `:3289` (`idx_one_active_draft_per_doc`).
- **Fix:** `CREATE UNIQUE INDEX uq_one_published_per_template ON public.templates_template_version (template_id) WHERE status = 'published'`. Considerar `uq_one_active_draft_per_template` se o workflow garante ≤1 in-flight.

#### F-DB3 · [HIGH] Baseline stale relativa a migrações 0211/0213/0233

> Severidade ajustada high→medium pelo verificador (guards `IF NOT EXISTS`/no-op ALTER previnem broken bootstrap). Listado como **medium** efetivo, mas com impacto direto na confiabilidade de análise de schema (esta auditoria).

- **Problema:** baseline atrás de 0211 (`editor_sessions.tenant_id`), 0213 (`templates_template.tenant_id` TEXT→UUID; `templates_audit_log` idem), 0233 (`revision_number` + `ux_templates_version_revision`). `pg_dump` de DB migrado diverge da baseline+tail. Baseline é fonte de verdade para review de DDL — análise de constraint dela é incompleta.
- **Evidência:** `db/baseline:2193, 2039-2049, 2214-2235`; `db/migrations/0213:40-42`, `0211:4-14`, `0233:15-36`.
- **Invariante violado:** `wiki/database/migration-policy.md` (baseline = estado cumulativo); runtime/schema truth beats docs.
- **Fix:** regenerar `db/baseline/0001_current_schema.sql` via `pg_dump --schema-only --no-owner --no-privileges` de DB totalmente migrado. Não mascarar com mais `IF NOT EXISTS`.

#### F-DB4 · [MEDIUM] `templates_template_version.content_hash=''` permitido pelo DB; gate de publish só no app

- **Problema:** `content_hash text NOT NULL` aceita `''`. Gate (`== ""`) só em `lifecycle.go:35, 349`. Versão sem docx pode ser submetida/publicada via escrita direta. `documents` tem `documents_content_hash_len CHECK (octet_length = 32)` (`:2025`) + trigger `enforce_snapshot_on_submit` (`:3751`).
- **Evidência:** `db/baseline:2219-2220`; `lifecycle.go:35-36`.
- **Irmão canônico:** `db/baseline:2025, 3751`.
- **Fix:** `CHECK (content_hash = '' OR length(content_hash) = 64)` + trigger/CHECK rejeitando `content_hash=''` quando `status NOT IN ('draft')`.

#### F-DB5 · [MEDIUM] `templates_template_version` sem `tenant_id` → RLS não protege a tabela (by_design)

- **Problema:** sem `tenant_id`; 0237 explicitamente pula a tabela; trigger `enforce_capability_asserted` seta `v_tenant_id := NULL`. Isolamento depende 100% do JOIN a `templates_template`. **Lacuna concreta subvalorizada:** `ObsoletePreviousPublished(Tx)` (`postgres.go:494-516`) UPDATE por `template_id`+`status` **sem nenhuma checagem de tenant na SQL** — depende de validação upstream.
- **Evidência:** `db/baseline:482-486, 2214-2235`; `0237` (ausente); `postgres.go:494-516`.
- **Irmão canônico:** `db/baseline:381-405` (`check_document_tenant_consistency`); `:729-747` (`enforce_signoff_tenant_consistent`).
- **Fix:** [médio prazo] adicionar `tenant_id` a `templates_template_version` para paridade RLS. [curto prazo] trigger DB de consistência de tenant por JOIN no parent em qualquer escrita; adicionar predicado tenant à SQL de `ObsoletePreviousPublished`.

#### F-DB6 · [MEDIUM] `ux_templates_version_revision` (one-revision-per-template) só em 0233, ausente da baseline

- **Problema:** 0233 adiciona `UNIQUE (template_id, revision_number)` (invariante ADR 0013). Baseline não tem (nem a coluna `revision_number`). O INSERT usa `COALESCE(MAX(revision_number)+1)` (`postgres.go:195-211`) — sem o UNIQUE, dois inserts concorrentes alocam o mesmo número.
- **Evidência:** `db/migrations/0233:35-37`; baseline grep = 0; `postgres.go:195-211`.
- **Irmão canônico:** `db/baseline:3639` (`ux_documents_cd_revision`).
- **Fix:** o índice em 0233 É o fix; avançar a baseline (F-DB3) p/ ambientes que bootstrapam só da baseline.

#### F-DB7 · [LOW] `templates_template_version.version_number` sem CHECK (>= 1)

- **Problema:** sem CHECK de positividade. App sempre passa ≥1, mas sem guard DB. `document_versions_mddm` (`:1328`) e `approval_route_stages` (`:1721`) têm.
- **Evidência:** `db/baseline:2217`; siblings `:1328`.
- **Fix:** `ALTER TABLE ... ADD CONSTRAINT chk_template_version_number_positive CHECK (version_number >= 1)`.

#### F-DB8 · [LOW] `ObsoletePreviousPublishedTx` sem predicado `tenant_id`

> Severidade ajustada high→low (UUID PK isola; call-sites precedidos de GetTemplate/GetVersion JOIN tenant + authz.Require). Defesa-em-profundidade/policy, não exploitable. Sobrepõe-se a F-DB5.

- **Fix:** adicionar `AND template_id IN (SELECT id FROM templates_template WHERE tenant_id = $3::uuid)` + param tenantID. Irmão: `postgres.go:528-529` (GetApprovalConfig já usa o padrão).

---

### 2.9 IAM / AuthZ / Seeds / Reference Data

#### F-IAM1 · [HIGH] Seed dev `approver` sem linha em `user_process_areas` → tier-2 falha sempre (403)

- **Problema:** o usuário `approver` tem `iam_user_roles` mas zero `user_process_areas`. `authz.Require` consulta **só** `user_process_areas` (JOIN `role_capabilities`); nunca lê `iam_user_roles`. Tier-1 `CanDo` passa (lê ambos), tier-2 retorna `ErrCapDenied` → 403 `FORBIDDEN_CAPABILITY`. **O bug é o SEED, não o código** (PDP two-tier funcionando por design). Bloqueia QA e2e do lifecycle de template via API ao vivo.
- **Evidência:** `db/dev-seeds/0001_local_dev_seed.sql:36-42` (insere em iam_user_roles), `:118-194` (`approver` ausente do bloco UPA — só admin/author-test/approver-test/reviewer-1); `internal/modules/iam/authz/authz.go:129-141` (JOIN só UPA); `internal/modules/iam/application/capability_service.go:62-67` (CanDo cobre ambos).
- **Invariante violado:** ADR 0022 — tier-1 grant precisa de membership tier-2 correspondente.
- **Irmão canônico:** `db/dev-seeds/0001_local_dev_seed.sql:149-157` (padrão `approver-test` na área `rh`).
- **Fix:** adicionar linha `user_process_areas` para `approver` na área `rh` (role='approver'), espelhando approver-test. Como caps de template usam `areaCode='tenant'` no tier-2 (filtro de área OFF), qualquer UPA ativa com role que carrega `template.approve` satisfaz o EXISTS.

#### F-IAM2 · [LOW] Comentário stale em `capability_scope.go` (e ADR 0022) afirma gap já fechado

- **Problema:** NOTE (`:31-34`) diz que `document.create`/`controlled_documents.*` ainda passam `"tenant"` no tier-2 — falso desde Phase 7. Todos os 4 call-sites passam area real. Texto stale também em `wiki/decisions/0022-...md:101`. Mislead de manutentor.
- **Evidência:** `internal/modules/iam/domain/capability_scope.go:31-35`; `documents/repository/repository.go:145` (docArea); `controlleddocuments/application/service.go:306,350,523` (ProcessAreaCode); `wiki/decisions/0022-authz-capability-coherence.md:101`.
- **Fix:** remover/atualizar o NOTE ("Phase 3 gap CLOSED em Phase 7") em ambos os locais.

#### F-IAM3 · [LOW] Role `approver` sem `template.archive` (by_design, não documentado)

- **Problema:** matriz concede `template.archive` só a `qms_admin` (+ system_admin bypass). `approver` tem approve/review/view. Padrão consistente com `template.publish` também retido → SoD intencional (só QMS admin executa transições terminais), mas não documentado.
- **Evidência:** `db/reference-data/0001_product_reference_data.sql:121, 23-25`; `capability_scope.go:59` (ScopeTenant).
- **Irmão canônico:** `reference_data.sql:110-114` (grant com comentário explicativo).
- **Fix:** documentar a fronteira SoD em `wiki/modules/iam.md` ou comentário no reference data. Se a intenção for permitir, adicionar o grant.

---

### 2.10 Testes — Fronteiras de Módulo

#### F-TST1 · [LOW] Testes de integração de `documents` importam `controlleddocuments/infrastructure` diretamente

- **Problema:** 3-4 arquivos de parity test importam `cdinfra` e chamam `NewCDFieldReaderPG()`. Test code obedece às mesmas regras de fronteira que produção.
- **Evidência:** `internal/modules/documents/application/document_area_parity_integration_test.go:12,86`; `documents/approval/application/read_service_area_parity_integration_test.go:12,130`; `documents/approval/jobs/scheduled_publish_job_test.go:22`; interface publicada `controlleddocuments/domain/cd_field_reader_port.go`.
- **Fix:** centralizar o cruzamento numa factory `testdb.NewCDFieldReader(t)` que internamente chama `NewCDFieldReaderPG()` (parity tests precisam de reader PG real). Remover conforme `legacy-test-deletion` quando o port for verificado (são gates D6 temporários).

#### F-TST2 · [LOW] `templates` importa `render/domain` + singleton `sync.Once` global

- **Problema:** import de `render/domain` é by_design (ADR 0050, contrato publicado cross-module). O defeito real é o `placeholderCatalogSetOnce` (`sync.Once`) cacheando um slice estático de 8 tokens — estado global mutável intestável, por zero ganho (ComputedCatalog é função pura). `routes_catalog.go:22` já chama inline (padrão correto).
- **Evidência:** `internal/modules/templates/application/schema.go:13,106-119`; `internal/modules/render/domain/computed_catalog.go:17-28`; `routes_catalog.go:7,22`.
- **Fix:** chamar `renderdomain.ComputedCatalog()` inline em cada `ValidatePlaceholders` (ou injetar via construtor p/ override em teste). Remover o singleton.

---

## 3. Roadmap de Remediação Priorizado

> Ordenado por severidade efetiva + dependência. Módulo dono por `developing-new-work`. **(ADR)** = precisa decisão ADR; **(gate)** = precisa system-impact gate de new-work.

### Onda 0 — Desbloqueio imediato (lifecycle quebrado / QA bloqueado)

| # | Item | Módulo | Notas |
|---|---|---|---|
| 0.1 | **F-T1** — adicionar `idempotencyKey: crypto.randomUUID()` (via ref estável) aos 3 callers FE de templates | `frontend/templates` | Bloqueador crítico; mecânico; replica `createTemplate`. |
| 0.2 | **F-IAM1** — seed `user_process_areas` p/ `approver` na área `rh` | `iam` / db dev-seeds | Desbloqueia QA e2e do lifecycle. |

### Onda 1 — High (integridade de publish, authz, contrato)

| # | Item | Módulo | Notas |
|---|---|---|---|
| 1.1 | **F-T2** — gate `content_hash` em `Approve` | `templates` | Espelhar T-004 de PublishTemplateVersion. |
| 1.2 | **F-CD1 + F-CD2 + F-CD3 (unificado)** — migrar os 5 handlers de approval para `idempotency.Require()` (declara header na spec + replay store + validação UUID) | `documents/approval` | **(ADR-leve)** decisão spec-level; regenerar oapi-codegen. Resolve 3 achados. |
| 1.3 | **F-D1** — substituir `IsDocumentOwner` por `authz.Require(CapDocumentView)` nos 19 read-sites | `documents` | ADR 0022; blast radius alto. |
| 1.4 | **F-D2** — authz in-tx em `CreateCheckpoint`/`ListCheckpoints` | `documents` | Espelhar `CommitUpload`. |
| 1.5 | **F-DB1 + F-DB4 + F-DB7 (uma migração)** — CHECK de `status`, `content_hash`, `version_number>=1` | `templates` / db | Forward-only. |
| 1.6 | **F-DB2 + F-DB6** — partial-unique de published + revisão | `templates` / db | Fecha race de double-publish. |
| 1.7 | **F-O4** — `tenant_id` + RLS em `document_exports` | `documents` / db | Invariante multi-tenant. |
| 1.8 | **F-FE1, F-FE2, F-FE3** — upload-fail bloqueia submit; surfacing de stale-conflict; 404 readonly por status | `frontend/templates` | Integridade de dados na UI. |
| 1.9 | **F-DB3** — regenerar baseline via pg_dump | db | Pré-requisito de confiabilidade p/ futuros reviews de schema. |

### Onda 2 — Medium (fronteiras de módulo, hygiene, config)

| # | Item | Módulo |
|---|---|---|
| 2.1 | **F-CD4** — injetar ports (ActiveInstanceReader/Profile/Area/Governance/TemplateVersion) via Dependencies + composition root | `controlleddocuments` |
| 2.2 | **F-CD5** — constantes locais + parity test p/ VersionStatus | `controlleddocuments` |
| 2.3 | **F-CD6** — remover authz duplicado de `CreateTx` (service é o gate) | `controlleddocuments` |
| 2.4 | **F-CD7** — `UpsertApprovalConfig` capability em vez de role-string **(ADR — nova cap `template.manage`)** | `templates` |
| 2.5 | **F-O2** (antes de wirar) + **F-O1** + **F-O3** + **F-O5** + **F-O6** — tenant-guards e helpers no kernel objectstore | `objectstore` / `documents` / `docgenv2` |
| 2.6 | **F-O7** — `MinIORegion` env-configurável | platform/config |
| 2.7 | **F-R1** — deletar `Freeze()` dead-code + scaffolding | `documents`/`render` |
| 2.8 | **F-R2, F-R3** — config env do staging worker + DLQ/observabilidade | `render` |
| 2.9 | **F-D3** — wirar `WithIAMReader` + fail-closed PHUser | `documents` |
| 2.10 | **F-D4** — `Idempotency-Key` de cliente no `SubmitHandler` | `documents/approval` |
| 2.11 | **F-C1→F-C2→F-C3** — declarar query params de `GET /templates`, migrar FE p/ tipo gerado, remover fallback | `templates` + `frontend/templates` |
| 2.12 | **F-T3** — remapear erro de concorrência (412→409) | `templates` |
| 2.13 | **F-T4** — `approver_role` no contrato de create | `templates` |
| 2.14 | **F-DB5** — trigger de consistência de tenant + predicado em ObsoletePreviousPublished (sobrepõe F-DB8) | `templates`/db |
| 2.15 | **F-FE4, F-FE5, F-FE7** — dirty-state, type=button/pt-BR, importDocx .ok | `frontend/templates` |

### Onda 3 — Low (hygiene, docs, defer pós-v1)

F-T5, F-T6 (GC sweeper — **(gate)** se novo job em `jobs`), F-D5, F-D6, F-CD8, F-CD9, F-R4, F-R5, F-R6, F-O8, F-C4, F-C5, F-IAM2, F-IAM3, F-TST1, F-TST2.

---

## 4. O Que Já Está Correto / Validado

- **Copy-on-spawn store-then-reference** (`lifecycle.go:451-471`): semântica correta — objeto existe antes da row referenciar; `ContentHash` deixado vazio força edição real (gate de publish). Único modo de crash é órfão seguro.
- **Migração 0250** — `UNIQUE(docx_storage_key)` + de-share de keys colididas: previne a corrida de spawn concorrente no nível DB; documentação da migração descreve as duas classes de falha.
- **PDP two-tier corretamente implementado no design** — a falha do `approver` é puramente de seed (F-IAM1); o split `iam_user_roles` (tier-1) vs `user_process_areas` (tier-2) está exatamente conforme ADR 0022.
- **Contract-first honrado** onde aplicado: `recordDocumentSignoff` (`openapi.yaml:3481`) declara `Idempotency-Key` corretamente; `createTemplate`/`finalizeDocument` no FE passam a key.
- **Optimistic-lock CAS** em `UpdateVersionTx` (`postgres.go:362`) — previne double-commit (só a classificação de erro precisa ajuste, F-T3).
- **`view_service.go:76`** — `authz.Require(CapDocumentView, "tenant")` in-tx é o padrão correto que os read-routes (F-D1) deveriam seguir.
- **Pin (`freeze_service.go:188`)** — tx-only, zero network, conforme ADR 0015; é o caminho de produção real.
- **Padrões DB de constraint validados existentes** — `documents_status_check`, `documents_content_hash_len`, partial-uniques (`ux_approval_instances_active_document_id`, `idx_one_active_draft_per_doc`, `ux_documents_cd_revision`): os fixes DB têm irmãos canônicos diretos no mesmo schema.
- **Composition-root DI correto** em `documents` (`CDFieldReader`/`AreaCatalogReader`) e `taxonomy` (`TplChecker`) — exatamente o padrão que CD (F-CD4) deve adotar.
- **`render/domain`** como contrato publicado cross-module (ADR 0050) — o import é legítimo; só o singleton (F-TST2) é o smell.

---

## 5. Cobertura e Limites desta Auditoria

**Verificado por leitura direta de código/schema** (file:line confirmados): todos os 71 achados passaram verificação adversarial; 18 tiveram severidade/escopo ajustados (ex.: F-O4 critical→high, F-CD5/CD7 high→medium, F-O1/O2/O3 ajustados por unwired/UUID-isolation, F-DB8/F-T7-família high→low por isolamento UUID, F-DB3 high→medium por guards idempotentes).

**NÃO coberto / precisa verificação runtime mais profunda:**

1. **Comportamento real do eigenpal** (`@eigenpal/docx-editor-react`) no caminho readonly de F-FE3 — o loop/hang no parse de docx ausente foi observado em preview QA mas não isolado; o guard `blankDocument` em nível de componente pode divergir do comportamento da lib. Precisa repro runtime.
2. **Race de double-publish (F-DB2) sob carga real** — confirmada por análise estática (READ COMMITTED + UPDATE não-locking); não reproduzida com clientes concorrentes reais.
3. **`document_exports` colisão cross-tenant (F-O4)** — a mitigação por UUID PK foi raciocinada, não testada com fixtures multi-tenant.
4. **Migração de baseline (F-DB3)** — divergência confirmada por diff de DDL; um `pg_dump` real de DB totalmente migrado vs baseline+tail não foi executado (a auditoria não rodou migrações).
5. **Janela de 404 pré-autosave** (F-FE3 + KNOWN FACT eigenpal) — a frequência real em produção depende do timing spawn→submit, não medida.
6. **Flaky test conhecido** (`TestSweeper_ExitsOnContextCancel`, ratelimit) — fora do escopo doc/template; mencionado por completude.
7. **Frontend além de `templates`/`approval`/`documents`** — outros features (distribution, search) não auditados.
8. **Caminhos de worker/jobs além de staging-outbox e pdf_job_runner** — não exaustivamente cobertos.

**Recomendação de seguimento:** rodar a suíte de integração (`go test ./...` + framework de fixtures canônico) após Onda 0+1 para confirmar empiricamente F-T1 (lifecycle desbloqueado), F-IAM1 (copy-on-spawn e2e via API ao vivo), e F-DB2 (sob teste de concorrência). O KNOWN FACT do dev-seed indica que approve/publish e2e via API ao vivo só é verificável após F-IAM1 — até lá, validar via suíte de integração.