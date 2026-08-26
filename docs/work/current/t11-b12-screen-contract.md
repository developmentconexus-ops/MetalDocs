# T11 — B12 Document Governance Administration — P9 Screen Contract

> **Status:** COMPLETE após LOCK do P8 R4 (operador, 2026-08-26). Temporary branch work, non-authoritative.
> **Base:** P8 R4 `t11-b12-document-governance-p8.html` (HTML único auto-contido).
> **Autoridade:** journeys §7/§22/§23/§24; wire-contract §3.1/§3.4 + tabela ops; `template-configuration-read.md`; `governance-step-deadline.md`; `access-assignment-read.md` (precedente de padrão apenas).
> **Implementation:** BLOCKED by `docs/roadmap.md`.

## 1. Rota / superfície

```text
GOAL      governance admin configura tipos de documento (base, rota, publicação, elegíveis) e modelos
ROUTE     /admin/document-governance — terceira seção do Admin Center (journeys §22), shell B01/B01N herdado
LENTES    Tipos de documento | Modelos (tabs locais, estado de UI apenas)
AUTHZ     superfície sob document_type.manage e/ou template_use.manage; 403 exibe painel de negação,
          nunca coleção vazia; co-location de UI nunca funde as duas permissões (journeys §22)
```

Classe de estado do cliente (global): página servida + ETags lidos + composição de formulário em edição. O frontend nunca possui verdade de negócio: sem cache cross-page, sem contagem total inventada, sem inferência de autorização, sem estado de lifecycle próprio.

## 2. Lente Tipos — lista (op34)

| Campo | Contrato |
|---|---|
| READ TRUTH | op34 `listDocumentTypes` → `DocumentTypePage`, ordem document_type_id ASC, página exata do servidor |
| IDENTIDADE | `document_type_id` de cada item; seleção alimenta o detail |
| WIRE | `SAFE_READ`, `PAGED` (cursor opaco, `has_more`), `JSON_NO_STORE` |
| FALHAS | falha de continuação → página atual preservada + retry explícito (P10-S3); 403 → painel de negação |
| SUCESSO | seleção re-lê o detail (ops 36/38/40) |
| PROIBIDO | crawl oculto de páginas; total fabricado |

## 3. Criar tipo (op35)

| Campo | Contrato |
|---|---|
| WRITE | op35 `createDocumentType` — `CreateDocumentTypeRequest { code:CodeInput, name:ShortText, numbering_scope, active, governance:GovernancePolicy, representation:RepresentationPolicy }` |
| WIRE | `IDEMPOTENT_CREATE`: X-CSRF-Token + Idempotency-Key UUID; 1 intenção lógica = 1 chave; retry ambíguo repete a MESMA chave; composição congelada durante ambiguidade (P10-S5) |
| VALIDAÇÃO LOCAL | CodeInput trim→uppercase→`^[A-Z0-9]+$` (1..32), “-” proibido; governança `use_governance_route` ⇒ ≥1 step; `due_in_days` inteiro ≥1 opcional |
| FALHAS | `U` 403; `J` 422 validation.failed; `I` conflito de chave; `S` 409 code duplicado na Company — mensagem nomeia o conflito, zero mutação |
| SUCESSO | `201 CreateDocumentTypeResult { document_type_id }` → tipo aparece na página do servidor que o contém; conjunto elegível inicia vazio |
| IDENTIDADE | `document_type_id` retornado; nunca inventado no cliente |

## 4. Detail do tipo — três domínios If-Match separados (journeys §24)

Cada seção lê seu próprio ETag e grava com If-Match próprio; um 412 numa seção nunca contamina as outras. Stale: `412 precondition.resource_changed`, zero mutação, ação explícita "Reler versão atual" e reaplicar.

### 4.1 Base (op36 read / op37 write)

```text
READ    op36 getDocumentType → DocumentTypeView { document_type_id, code, name, numbering_scope, active } + ETag
WRITE   op37 replaceDocumentType (PUT integral) — ReplaceDocumentTypeRequest
FALHAS  U/J/N/P(412)/S — S inclui 409 state.document_type_in_use: code/scope imutáveis após primeiro
        Documento comprometido usar o tipo (journeys §7); leitura não projeta o estado (B12-F1 REJECTED
        pelo operador: tentar-e-falhar; a UI apresenta a falha honestamente, sem flag simulado)
SUCESSO 200 DocumentTypeView + novo ETag; fatos re-renderizados
```

### 4.2 Rota de governança + Publicação oficial (op38 read / op39 write)

```text
READ    op38 getDocumentTypeGovernance → DocumentTypeGovernanceView { governance, representation } + ETag
WRITE   op39 replaceDocumentTypeGovernance (PUT integral If-Match) — UMA verdade de versão: rota e
        publicação gravam juntas; a separação visual "Publicação oficial" é apresentação (R2),
        o diálogo de edição declara o domínio único
POLICY  GovernancePolicy união fechada: no_human_approval (membro steps PROIBIDO) |
        use_governance_route (steps minItems=1); GovernanceRouteStep { label:ShortText,
        selector:{named_user,user_id}|{group,group_id}, due_in_days? inteiro ≥1 } — ordem do array = rota
REPR    RepresentationPolicy: {source_only} | {require_official_rendition, format:pdf}
EDITOR  reordenação por botões/teclado (nunca drag-only); remoção; validações locais antes do PUT
CONSEQ  attempts existentes preservam snapshot imutável da rota; só attempts futuros usam a nova config
FALHAS  U/J/N/P(412)/S
IDENT   user_id via op6 listUsers (picker paginado), group_id via op22 listGroups (picker paginado);
        rótulos exibidos são reconhecimento, nunca autoridade
```

### 4.3 Modelos elegíveis (op40 read / op41 write)

```text
READ    op40 getDocumentTypeEligibleTemplates → EligibleTemplatesView { templates:DocumentReference[] }
        ordem document.code ASC, document_id ASC + ETag; conjunto vazio é estado válido, não falha
WRITE   op41 replaceDocumentTypeEligibleTemplates — substituição INTEGRAL do conjunto
        { template_document_ids: unique Uuid[] } com If-Match; vazio válido
PICKER  candidatos vêm da projeção limitada op43 (paginada); sem conteúdo/histórico (journeys §23)
FALHAS  U/J/N/P(412)/S
SUCESSO 200 EligibleTemplatesView + novo ETag; chips re-renderizados
```

### 4.4 Preview de numeração (op42)

```text
READ    op42 getDocumentTypeNumberingPreview → NumberingPreviewView { preview_code, reservation:false }
QUERY   area_id opcional; scope=document_type_area sem area_id → 422 validation.failed apresentado
IDENT   area_id via op16 listAreas
VERDADE preview NUNCA reserva sequência; código final só existe na criação atômica do Documento;
        lacunas permitidas — o texto da UI afirma isso
FALHAS  A/N/validation.failed
```

## 5. Lente Modelos (op43 REFINADO / op50–51)

### 5.1 Coleção filtrada (op43 + `template-configuration-read.md`)

```text
READ    op43 listTemplateConfigurations → TemplateConfigurationPage; ordem document.code ASC, document_id ASC
FILTROS server-side ANTES da paginação (decisão ratificada): q (code + título EFETIVO, case-insensitive,
        título de rascunho nunca casa), eligible_document_type_id, template_role, has_effective_revision;
        conjuntivos; mudar qualquer filtro inicia NOVA identidade de primeira página; cursor autentica
        operationId + filtros normalizados + posição; repetir filtro com cursor = 400 request.invalid
PROJ    TemplateConfigurationItem { document, template_role, has_effective_revision,
        current_effective_title? (iff has_effective_revision), eligible_document_type_ids UUID ASC }
VAZIO   página vazia sob filtro = resultado ordinário do servidor, nunca erro nem prova de inexistência
GERAL   "template geral" é fato derivado (elegível em N tipos) — sem flag/conceito (operador, 2026-08-26)
PROIBIDO post-filter de cliente apresentado como completo; crawl para emular filtro
```

### 5.2 Detalhe do modelo (projeção §23 + fronteiras)

```text
CONTEÚDO document id, code, título efetivo condicional, papel de modelo, indicador de revisão efetiva,
         tipos elegíveis — exatamente a projeção mínima do journeys §23
FRONTEIRA "Abrir documento ↗" → continua em B03 (Documento Oficial); "Histórico ↗" → continua em B07;
         terminadores método §14.3 no P8; navegação real no P11; acesso lá sob document.read_effective /
         document.read_history do usuário — template_use.manage NÃO concede leitura
PROIBIDO revisões/aprovação/conteúdo inline nesta lente (operador declinou reopen do §23, 2026-08-26)
```

### 5.3 Papel de modelo (op50 read / op51 write)

```text
READ    op50 getDocumentTemplateRole → TemplateRoleView { document_id, is_template } + ETag
WRITE   op51 replaceDocumentTemplateRole { is_template } com If-Match, sob template_use.manage
CONSEQ  atribuir: passa a ser administrável/elegível; opções de criação continuam exigindo revisão
        EFETIVA + elegibilidade no momento da criação. Remover: deixa de ser oferecido; documentos já
        derivados permanecem válidos, sem rebind
FALHAS  U/J/N/P(412)/S; stale → reler e reaplicar
```

## 6. Estados negativos globais

```text
403 permission.denied      painel de negação da superfície; dados/lentes/ações não renderizam como vazio (P10-S6)
404 resource.not_found     detail do tipo: ausência não prova existência oculta; instrução de recuperação
falha de continuação       página atual preservada + retry explícito (P10-S3)
412 stale                  por domínio, zero mutação, reler → reaplicar (journeys §24)
409 semânticos             code duplicado (op35), tipo em uso (op37) — mensagem nomeia a causa exata
resultado ambíguo (op35)   replay com mesma Idempotency-Key; nunca segunda intenção (P10-S5)
```

## 7. Acessibilidade / responsivo (estruturais)

```text
tabs/dialogs semânticos (role=tab/tablist, <dialog>, aria-labelledby, aria-live nos state-lines)
reordenação de steps por botões — nunca drag-only
navegação global vira drawer; painéis empilham; linhas viram blocos rotulados (mobile-hint demonstrado)
foco gerenciado nos diálogos; teclado cobre todos os caminhos materiais
```

## 8. Trace bidirecional — prova

Produto/backend → frontend (toda operação da família tem superfície ou fronteira):

```text
op34 → lista de tipos            op35 → diálogo criar tipo       op36/37 → seção Base
op38/39 → seção Rota+Publicação  op40/41 → seção Elegíveis        op42 → preview de numeração
op43 → lente Modelos (filtros ratificados) + picker de elegíveis  op50/51 → papel de modelo
op6/op22 → pickers de selector   op16 → seleção de Área do preview
```

Frontend → Produto/backend (todo controle material tem dono; nenhum inventado):

```text
22 regiões/controles materiais TRACED / 22
operações primárias B12: 12/12 BOUND (34–43, 50, 51)
leituras de apoio: op6, op16, op22
controles sem operação: 0 · operações humanas sem disposição: 0 · APIs em forma de tela: 0
```

## 9. Autoridade proibida ao frontend

```text
não decide imutabilidade de code/scope (verdade do op37)
não funde rota e publicação em domínios separados nem separa a gravação una
não computa elegibilidade efetiva de criação (verdade do op44 na criação)
não expõe conteúdo/revisões/histórico na administração (journeys §23)
não avalia autorização; não reordena/pagina no cliente; não cria conceito "template geral"
fixtures do P8 são Evidence, nunca Produto
```

## 10. Suficiência do backend

Toda necessidade material do bloco está PRESENT-IN-AUTHORITY após a ratificação de `template-configuration-read.md` (B12-F2). B12-F1 REJECTED pelo operador. Nenhuma contradição P9 → nenhum reopen de P7/P8.
