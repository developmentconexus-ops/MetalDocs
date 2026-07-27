# QA-4 — Browser QA fim-a-fim "empresa do zero" (2026-07-24)

Operator-guided run (stack :80, HEAD 21f52f29 + fix noopPresigner). Auth: dev-seed
personas (`admin`, `author-test`, `approver-test`), senhas resetadas com aprovação
do operador; cookie injetado — nenhuma senha digitada em form (proibição vigente).

Cenário: montar empresa do zero via UI — taxonomia nova (família → perfil → área),
depois jornada documento completa no perfil/área novos.

## Setup empresa do zero (persona admin)

| Passo | Resultado |
|---|---|
| /admin/taxonomy → Nova Família `operacoes` "Operações" | PASS — criada, listada Ativa |
| Novo Perfil `it` "Instrução de Trabalho", família operacoes, role editora author | PASS — criado, listado Ativo |
| Nova Área `usinagem` "Usinagem", role aprovador approver | PASS — criada, listada Ativa |
| Wizard (author) mostra perfil `it` e área `usinagem` imediatamente | PASS |

## Findings

### F-QA4-1 — preview-code 403 engolido: wizard mostra `it-usinagem-???` sem erro

Wizard etapa 2 (perfil `it` + área `usinagem`, persona author-test): banner
"CÓDIGO GERADO · PRÓXIMO EM (IT, USINAGEM)" renderiza **`it-usinagem-???`** e
nada mais. Live:

```
GET /api/v1/controlled-documents/preview-code?profile_code=it&area_code=usinagem
→ 403 {"title":"you do not have the required capability in this area",
       "code":"FORBIDDEN_CAPABILITY"}
```

Dois defeitos:

1. **UX (erro silencioso).** `usePreviewCodeQuery` erro → `data` null →
   `formatCodePreview` (`frontend/apps/web/src/features/documents/lib/codePreview.ts:25`)
   faz `code ?? "<profile>-<area>-???"`. O estado de ERRO é indistinguível do
   estado "ainda não escolhi" para o usuário; nenhum toast/banner. Usuário avança
   até a confirmação e o create falhará com o mesmo 403 — tarde demais.
2. **Onboarding/gating inconsistente.** Dropdown de áreas do wizard é populado por
   `taxonomy.view` (global) e lista TODAS as áreas, mas preview-code/create são
   capability×área (tier-2, ADR 0022). Criar uma área nova não concede acesso a
   ninguém — fluxo "empresa do zero" leva o autor a um beco: seleciona a área
   recém-criada e não há sinal do porquê falhou. Relacionado ao flag anterior
   "preview-code 500→403" (doc+template kernel audit, task_7c4a984e).

**Melhoria proposta:** (a) surfacear erro do preview-code no banner (estado de
erro explícito: "Você não tem acesso a esta área"); (b) filtrar/anotar áreas sem
capability no dropdown (fonte: endpoint de áreas-elegíveis, não taxonomy.view);
(c) UX de onboarding: ao criar área, admin deve poder conceder acesso
(membership) no mesmo fluxo.

### F-QA4-2 — enums de role divergentes entre dialogs de taxonomia (hand-synced)

Dialog Novo Perfil (role editora): `viewer/editor/author/approver/system_admin`.
Dialog Nova Área (role aprovador): `viewer/editor/author/approver/signer/area_admin/qms_admin`.
Dois vocabulários de role hard-coded distintos na mesma tela. Instância do
meta-defeito "enumerações sincronizadas à mão" (final architecture review
2026-07-03). Melhoria: fonte única de roles gerada do contrato.

### F-QA4-3 — SUBMIT falha 409 silencioso: perfil novo não tem rota de aprovação

Documento `IT-USINAGEM-001` (perfil `it`, área `usinagem`), autor com membership
válida, conteúdo digitado e autosalvo ("Salvo", revisão 1854 B, page_count 1).
Clique em "Submeter para revisão" → **nada acontece na UI**. Rede:

```
POST /api/v1/documents/{id}/submit → 409
{"title":"no active approval route for this profile",
 "code":"state.approval_route_missing"}
```

Três cliques, três 409, **zero feedback visível** — sem toast, sem banner, sem
estado de erro no botão. Documento permanece `draft`. Mesma classe do F-QA4-1
(erro de API engolido pela UI), agora numa ação primária de escrita.

Segundo eixo (onboarding): criar um perfil pela tela de taxonomia **não** cria
nem exige rota de aprovação, e nada na UI de taxonomia sinaliza que o perfil
está incompleto. "Empresa do zero" trava aqui sem diagnóstico.

**Melhoria proposta:** (a) submit deve surfacear o problem+json (toast/banner com
o title e ação "configurar rota"); (b) taxonomia deve marcar perfil sem rota
ativa como incompleto (badge + link para /admin/routes); (c) considerar rota
default herdada da família na criação do perfil.

### F-QA4-4 — painel IDENTIFICAÇÃO com TIPO / ÁREA RESPONSÁVEL / VISIBILIDADE `---`

No workspace do documento recém-criado, o painel direito mostra CÓDIGO
corretamente (`IT-USINAGEM-001`), mas TIPO, ÁREA RESPONSÁVEL e VISIBILIDADE
renderizam `---`, embora a API devolva `profile_code_snapshot: "it"`,
`process_area_code_snapshot: "usinagem"` e a visibilidade tenha sido escolhida no
wizard. Dados existem, a tela não os liga. Cosmético/informacional, mas é a
ficha de identificação do documento controlado.

Observação adicional: chips do cabeçalho do workspace se sobrepõem
("RASCUNHO"/"Salvo"/"REV00" colididos) — mesma família do F23 (chip de header).

### F-QA4-5 — publicar é inalcançável pela UI: gate depende de `active-document`, que é escopo de área

Documento aprovado. Ficha completa (`/documents/{id}/details`) como **admin**
(system_admin, capability `document.publish`) mostra o botão desabilitado com o
rótulo **"Aguardando contexto ativo para publicar"**, em page load limpo. Causa:

```
GET /api/v1/controlled-documents/{cdId}/active-document
  como admin        → 404 {"code":"NO_ACTIVE_INSTANCE"}
  como approver-test→ 200 {"content_hash":"0e3936…","revision_version":1}
```

Não é estado — é escopo. Admin não tem membership na área `usinagem` (auto-grant
é corretamente recusado, SoD), então o contexto ativo "não existe" para ele e o
gate `canPublish` (`adapters/useDocumentArtifact.ts`, requer content_hash
confirmado) nunca abre. E quem TEM a área não tem a capability:

```
POST /documents/{id}/publish  como approver-test → 403 AUTH_FORBIDDEN
POST /documents/{id}/publish  como admin (via API, If-Match "v2") → 200 published
```

Ou seja: **o backend deixa o admin publicar, a UI não** — o gate de tela usa um
recurso escopado por área como pré-condição de uma ação escopada por capability
global. Na jornada "empresa do zero" ninguém consegue publicar pela tela.
Agravante de UX: o rótulo diz "aguardando", sugerindo espera transitória, quando
é permanente.

**Melhoria proposta:** (a) derivar o gate de publish do documento (status
`approved` + capability), não de `active-document`; (b) se o content_hash é
mesmo necessário para o If-Match, expor um endpoint de contexto legível por quem
tem `document.publish`; (c) trocar o rótulo por um motivo real.

**Reteste 2026-07-27 (QA-5) — afordância invertida, confirmada nos dois lados.**
Mesma ficha (`IT-USINAGEM-003`, `approved`), mesmo build, dois usuários:

```
admin        (TEM document.publish)  → <button aria-disabled="true"
                                        title="Aguardando contexto ativo para publicar">
                                       "A publicação está bloqueada porque o contexto
                                        ativo desta revisão ainda não foi confirmado."
author-test  (NÃO tem a capability)  → <button aria-disabled="false">  ← habilitado
                                       clique → NENHUMA requisição de rede,
                                       nenhum diálogo, nenhum toast: no-op mudo
```

Ou seja, o gate está **exatamente invertido**: quem pode publicar vê o botão
morto; quem não pode vê o botão vivo — e o clique não faz absolutamente nada
(nem 403, nem feedback). O gate atual só consulta `active-document`; a
capability do usuário não entra na decisão de renderizar/habilitar. Reforça a
melhoria (a): o gate tem de ser `status == approved && capability
document.publish`, e a ação sem capability nem deve ser renderizada.

### F-QA4-6 — validação de `Idempotency-Key` inconsistente entre rotas irmãs

Mesma família de rotas de mutação de documento, mesmo cliente:

```
POST /documents/{id}/signoff  Idempotency-Key: qa4-signoff-01 → 200 (aceito)
POST /documents/{id}/publish  Idempotency-Key: qa4-publish-01 → 400 IDEMPOTENCY_KEY_INVALID
                                                    ("Idempotency-Key must be a UUID")
```

Uma exige UUID, a irmã aceita string livre. Contrato de idempotência precisa ser
uniforme (ou o formato é livre, ou é UUID em todas).

### F-QA4-7 — resposta de signoff devolve `signoff_id` vazio

```
POST /documents/{id}/signoff → 200 {"signoff_id":"","was_replay":false,"outcome":"approved"}
```

A assinatura foi persistida (o dossiê a exibe corretamente), mas o id volta como
string vazia — cliente não tem como referenciar a assinatura recém-criada.
Campo obrigatório no contrato com valor não-preenchido.

### F-QA4-8 — inbox/dossiê exibem identificadores crus

1. Card da fila do aprovador identifica o item pelo **UUID**
   (`18d6bed1-fdf0-4690-b0b5-375fe925c3d8`) e não pelo código
   `IT-USINAGEM-001` — o código existe e é mostrado no workspace.
2. Painel `PRÓXIMOS APROVADORES` mostra **`?`** no lugar do nome do aprovador
   enquanto a decisão está pendente; após a aprovação passa a mostrar
   "AT · Approver Test" corretamente. O selector é `role_in_fixed_area
   approver@usinagem` e o usuário elegível é resolvível antes da decisão.

### F-QA3-1 — REPRODUZIDO com conteúdo real (blocker de release confirmado)

Jornada completa executada num documento com conteúdo autoral real (IT de torno
CNC, 1854 B, 1 página, autosalvo e verificado na tela do aprovador). Após
aprovação + publicação, os artefatos congelados:

Extração direta do objeto no MinIO
(`tenants/{t}/revisions/ba24c4f2…/frozen.docx`):

```
frozen.docx entries: ['_rels/.rels','word/document.xml','word/','[Content_Types].xml']
word/document.xml → texto: 18 chars (apenas whitespace)  ← ZERO conteúdo
final.pdf: 6118 bytes, em branco
```

Comparação de controle — export do PDF a partir da revisão viva
(`POST /documents/{id}/export/pdf`): **17726 bytes** com o conteúdo correto.

Confirma a causa já diagnosticada em QA-3: `FreezeService.Materialize` monta o
artefato a partir do snapshot do template e ignora `current_revision_id`. O
documento assinado e publicado — o artefato que vale legalmente — não contém o
texto aprovado. **Blocker de release, sem workaround.** Correção acordada:
opção (a) do verdict 2026-07-23.

### F-QA4-9 — `form_data_snapshot` opcional no contrato, NOT NULL no banco (500)

`POST /documents/{id}/autosave/commit` sem `form_data_snapshot` (campo **não**
listado em `required` no contrato, `api/openapi/v1/openapi.yaml:2890`) →

```
500 null value in column "form_data_json" of relation "documents"
    violates not-null constraint (SQLSTATE 23502)
```

Origem: `internal/modules/documents/infrastructure/repository.go:1158` grava
`form_data_json = $2` com o snapshot nil. Contrato diz opcional, escrita exige
valor. Um cliente conforme ao contrato derruba a rota de autosave com 500 (não
400). Melhoria: ou tornar o campo `required` no contrato, ou defaultar para
`'{}'::jsonb` na escrita — e devolver 400 problem+json, nunca 500.

### F-QA4-10 — `*_dispatch_outbox.content_hash` carrega values_hash (nome enganoso)

A coluna `content_hash` de `materialize_dispatch_outbox` / `pdf_dispatch_outbox`
**não** é o hash do artefato: é o `values_hash` dos placeholders resolvidos
(`freeze_service.go:206-218` → `ComputeValuesHash(valMap)` →
`EnqueueMaterializeTx(..., hashBytes)`). Num template sem placeholders o valor é
sempre `e3b0c442…` (SHA-256 do mapa vazio) para todos os tenants e documentos —
inclusive no run VERDE pós-correção. Custou uma leitura errada de evidência
neste próprio relatório (o `e3b0c442…` foi lido como "artefato vazio"; o
artefato vazio era real, mas a prova é o objeto no MinIO, não essa coluna).
Melhoria: renomear para `values_hash` (o hash do artefato já vive em
`documents.content_hash`) ou passar a gravar o hash real do artefato.

### F-QA4-11 — `values_frozen_at` sempre `null` no contrato (coluna nunca lida)

```
GET /api/v1/documents/{id}  →  "values_frozen_at": null
SELECT values_frozen_at ... →  2026-07-27 22:31:17.148056+00
```

`Repository.GetDocument`
(`internal/modules/documents/infrastructure/repository.go:291-304`) não inclui
`d.values_frozen_at` na lista de colunas do SELECT, então o campo do contrato
(`api.gen.go:250`) é sempre `null` para todo documento — inclusive congelados.
Campo publicado no contrato que nunca carrega valor: consumidor que decidir
"está congelado?" por ele erra sempre. Melhoria: incluir a coluna no SELECT (é
a única correção necessária; o mapeamento em `handler.go:481/532` já existe).

### F-QA4-12 — nav primário: item morto ("Auditoria") e telas de admin sem entrada

`Rail.tsx:19` declara `{ label: 'Auditoria', path: '/audit' }`, mas **nenhuma
rota `/audit` existe** — não há feature `audit` em
`frontend/apps/web/src/features/` e o `AppRouter.tsx:35` tem
`{ path: '*', element: <Navigate to="/" replace /> }`. Clicar em "Auditoria" no
nav primário devolve silenciosamente o usuário à home, sem erro nem 404.

Espelho do mesmo problema: as telas de administração que a jornada "empresa do
zero" exige — `/admin/taxonomy`, `/admin/memberships`, `/admin/routes` — **não
têm nenhuma entrada de navegação**; só se chega a elas digitando a URL. O nav
mostra o que não existe e esconde o que existe.

Melhoria: (a) remover o item morto ou implementar a tela de auditoria (o feed de
eventos já existe na home, seção MURMÚRIOS); (b) adicionar um grupo
"Administração" no rail, filtrado por capability.

## F-QA3-1 — CORRIGIDO e verificado live (2026-07-27)

Correção implementada (opção (a) do verdict 2026-07-23), commit `e1c0ea28`:

- `SnapshotReader.ReadCurrentRevisionBodyKey` (novo seam) — lê
  `document_revisions.storage_key` via `documents.current_revision_id`, join
  tenant-predicado (`snapshot_repository.go:112-133`).
- `FreezeService.Materialize` passa esse key como `BodyDocxS3Key` no fanout, em
  vez de `snap.BodyDocxS3Key` (snapshot do template); revisão ausente →
  **falha fechada** (`materialize: document %s has no current revision body to
  freeze`), sem fallback silencioso.
- Testes de fake/asserção atualizados; `go build ./...`,
  `go vet -tags integration ./...` e testes do módulo `documents` verdes.
- Imagens `metaldocs-api:dev` / `metaldocs-worker:dev` reconstruídas do commit e
  containers recriados antes da verificação.

Jornada de verificação (documento novo `IT-USINAGEM-003`, perfil `it` / área
`usinagem`, id `2930e658-4e5b-43d4-b13a-e45d4463cd02`), com o **mesmo corpo
autoral** do run que falhou (1854 B, IT de torno CNC):

| Passo | Resultado |
|---|---|
| Criar CD + upload do corpo (acquire → presign → PUT → commit) | PASS — `revision_id 2db48d32…`, `file_size_bytes 1854`, `page_count 1` |
| Submeter (`If-Match: "v0"`) | PASS — `instance_id bc22bb29…`, etag `"v1"` |
| `active-document` (approver) | PASS — `content_hash b8d4be68…`, `revision_version 1` |
| Signoff aprovar+assinar (`If-Match: "v1"`) | PASS — `outcome: approved` |
| Materialize + PDF (worker) | PASS — `pdf_generated_at 22:31:42` |

Artefatos congelados, extraídos direto do MinIO:

```
documents.content_hash = 8586754f7494761143ace5a5fb30416f800c10867f3f7f5f54612e9f434179fd
frozen.docx : 1854 bytes  (= byte-a-byte o tamanho do corpo autorado)
  word/document.xml: 2827 chars; texto extraído: 535 chars
  "1. OBJETIVO Padronizar o ajuste do torno CNC para usinagem da peca Flange A320.
   2. ESCOPO … 3. PROCEDIMENTO 3.1 Verificar o desenho tecnico DT-A320-REV3 …
   4. REGISTROS Ficha de setup FR-USI-002 preenchida a cada troca de lote."
final.pdf   : 17726 bytes
```

Comparação com o run pré-correção e com o controle:

| Métrica | Pré-fix (QA-4) | Pós-fix (QA-5) | Controle (export da revisão viva) |
|---|---|---|---|
| texto em `frozen.docx` | 18 chars (whitespace) | **535 chars reais** | — |
| `final.pdf` | 6118 B, em branco | **17726 B** | 17726 B |

O PDF congelado agora é **byte-idêntico em tamanho** ao export de controle. O
blocker de release F-QA3-1 está fechado no eixo "conteúdo assinado ≠ conteúdo
congelado".

Escopo restante da opção (a), **não** implementado (defer explícito):
hashes de linhagem revisão → frozen.docx → PDF, e expurgo dos artefatos
inválidos já produzidos (`ba24c4f2…`, `45c9e784…`, `d18fbfdf…`).

## Jornada

| Passo | Resultado |
|---|---|
| Wizard 4 etapas (perfil it → área usinagem → Em branco → confirmar) | PASS — IT-USINAGEM-001 criado, editor abre |
| Preview de código com membership | PASS — `IT-USINAGEM-001` (confirma causa do F-QA4-1) |
| Digitação no editor + autosave | PASS — "Salvo", revisão 1854 B, 1 página |
| Criar rota de aprovação para o perfil novo (UI route-admin) | PASS — "Rota IT — Usinagem" ativa, 1 estágio |
| Submeter para revisão (1ª tentativa, sem rota) | **FAIL** — F-QA4-3 (409 silencioso) |
| Submeter para revisão (com rota) | PASS — status `under_review` |
| Inbox do aprovador mostra a pendência | PASS com defeito — F-QA4-8 (UUID no card) |
| Workspace "Aprovando": conteúdo + painel de decisão | PASS — conteúdo íntegro na tela do aprovador |
| Signoff aprovar+assinar (executado via API, senha nunca digitada em form) | PASS — 200 `outcome: approved`; dossiê registra "Reautenticação por senha" |
| Publicar pela UI | **FAIL** — F-QA4-5 (gate morto); publicado via API para seguir |
| Artefato congelado (frozen.docx / final.pdf) | **FAIL** — F-QA3-1 reproduzido: conteúdo vazio |

## Jornada QA-5 — pós-publicação (2026-07-27)

| Passo | Resultado |
|---|---|
| Ficha completa do documento publicado (`/details`) | PASS — TIPO/ÁREA/TAMANHO/cadeia de aprovação corretos; TAMANHO 1,8 KB = artefato real |
| Workspace "Visualizando" (mesmo documento) | PASS com defeito — F-QA4-4 persiste: TIPO/ÁREA/VISIBILIDADE `---` na mesma tela em que a ficha os mostra |
| Botão "Publicar / Agendar" | **FAIL** — F-QA4-5 confirmado nos dois lados (afordância invertida) |
| Publicar (via API, admin, `If-Match "v2"`) | PASS — `new_status: published` |
| Biblioteca (`/documents`) | PASS — IT-USINAGEM-003 `PUBLICADO`; contagens coerentes (7 = 1 rascunho + 3 aprovados + 3 publicados) |
| Distribuição (`/details/distribution`) | PASS (gap declarado) — todas as ações `aria-disabled` com motivo explícito apontando `wiki/backlog/document-distribution-mission.md`. **Este é o padrão correto de gate** — exatamente o que falta em F-QA4-5 |
| Auditoria pelo nav primário | **FAIL** — F-QA4-12: rota inexistente, redirect mudo para a home |
| Feed de eventos na home (MURMÚRIOS) | PASS — registra `authz bypass system admin document.publish`, autosave e session acquired com autor e alvo corretos |

## Verdicto QA-4

**FAIL.** A jornada fim-a-fim "empresa do zero" chega ao fim, mas o produto final
— o documento assinado e publicado — sai em branco (F-QA3-1, blocker já
conhecido, agora com prova de conteúdo real). Além dele, dois defeitos de fluxo
bloqueiam a operação sem intervenção por API (F-QA4-3 submit silencioso,
F-QA4-5 publicar inalcançável) e cinco de contrato/UX (F-QA4-1, -2, -4, -6, -7,
-8) degradam o onboarding de um cliente novo.

Ordem de correção sugerida: F-QA3-1 → F-QA4-5 → F-QA4-3 → F-QA4-1 → resto.

### Estado dos achados (atualizado 2026-07-27)

| ID | Classe | Estado |
|---|---|---|
| F-QA3-1 | Blocker — artefato congelado vazio | **CORRIGIDO** (`e1c0ea28`), verificado live QA-5 |
| F-QA4-5 | Bloqueio de fluxo — publicar inalcançável na UI | ABERTO |
| F-QA4-3 | Bloqueio de fluxo — submit 409 silencioso | ABERTO |
| F-QA4-1 | UX/gating — preview-code 403 engolido (`it-usinagem-???`) | ABERTO |
| F-QA4-9 | Contrato⊥banco — autosave/commit 500 | ABERTO (novo) |
| F-QA4-2 | Enums de role hand-synced | ABERTO |
| F-QA4-4 | Painel IDENTIFICAÇÃO `---` + chips sobrepostos | ABERTO |
| F-QA4-6 | Idempotency-Key inconsistente | ABERTO |
| F-QA4-7 | `signoff_id` vazio | ABERTO |
| F-QA4-8 | UUID no card do inbox / `?` em próximos aprovadores | ABERTO |
| F-QA4-10 | `content_hash` do outbox = values_hash (nome enganoso) | ABERTO (novo) |
| F-QA4-11 | `values_frozen_at` sempre null (coluna fora do SELECT) | ABERTO (novo) |
| F-QA4-12 | Nav: item "Auditoria" morto; telas de admin sem entrada | ABERTO (novo) |

O veredito QA-4 permanece **FAIL** — o blocker caiu, mas F-QA4-5 e F-QA4-3 ainda
impedem a jornada de completar apenas pela tela.
