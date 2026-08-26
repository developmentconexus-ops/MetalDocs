# T11 — B12 Document Governance Administration — P6/P7 Planning

> **Status:** CANDIDATE / R1 — temporary branch work, non-authoritative.
> **Methods:** Engineering Method v1.0.0 + Frontend Product Experience Planning Method v2.3.
> **Implementation:** BLOCKED by `docs/roadmap.md`.

## 1. Block ledger

| Field | Value |
|---|---|
| Block | B12 — Document Governance Administration |
| Route | `/admin/document-governance` (terceira seção semântica do Admin Center — journeys §22) |
| User goals | governance admin configura Document Types (base, numeração, rota de governança, representação), elegibilidade de Templates por tipo e papel de Template por Documento |
| Authority pack | `product/journeys.md` §7, §22, §23, §24; `architecture/wire-contract.md` §3.1 (GovernancePolicy/RepresentationPolicy), §3 views ops 34–43, 50–51; `decisions/api-operation-census.md`; `decisions/governance-step-deadline.md` §4.1 (config `due_in_days?`); `decisions/governance-review-layer-seam.md` (nada promovido) |
| Primary operations | ops 34–43 (família Document Governance config) + ops 50–51 (Template role por Documento, `template_use.manage`) |
| Supporting reads | op6 `listUsers` (seletor NAMED_USER), op16 `listAreas` (numbering preview), op22 `listGroups` (seletor GROUP) |
| Dependencies | B01 shell LOCK (P10-S1), B01N chrome (P10-S2), padrões compartilhados B11 P10-S3…S7 |
| Permission truth | `document_type.manage` (config de tipo) e `template_use.manage` (papel/elegibilidade de Template) — co-location de UI nunca funde autoridade de permissão (journeys §22) |

## 2. P6 — Reference study (conditional)

Gate do método: referências apenas onde a ambiguidade é real e consequente.

- **Editor de rota sequencial + deadlines:** o estudo externo já ratificado em `governance-step-deadline.md` §2 (Camunda, Flowable, ProcessMaker, ServiceNow) cobre o espaço de decisão do padrão rota/step/deadline. Referência adicional não muda o espaço de decisão → **SATISFIED BY EXISTING EVIDENCE**.
- **Master-detail de configuração administrativa:** padrão de tarefa ordinário já operado e LOCKED em B10/B11 → sem ambiguidade consequente nova.

Disposições de capacidades da classe de referência (workflow admin):

```text
steps paralelos / quóruns              REJECTED — rota é sequencial por autoridade T2/journeys §22
calendário útil / feriados / horas     REJECTED — Launch: 1 dia = 24h corridas (governance-step-deadline §4.1)
escalação / reassignment automáticos   REJECTED — due_at é apresentação/atenção, nunca lifecycle
editor de Role/Permission custom       REJECTED — journeys §22: não existe
PolicyVersion / plataforma de workflow REJECTED — journeys §22: não existe
comentário/review inline               REJECTED — governance-review-layer-seam: delta de capability ZERO
priority manual em step                REJECTED — governance-step-deadline: não herdado por analogia
```

## 3. P7 — Layout hypotheses

**H1 (leading):** um workspace `/admin/document-governance` com duas lentes locais — **Tipos de documento** (master list op34 → detail com três seções de escrita separadas: base op36/37, governança op38/39, modelos elegíveis op40/41 + preview de numeração op42) e **Modelos** (lista transversal op43 com papel de Template op50/51).

**H2:** páginas separadas por família de recurso (tipos e modelos como rotas irmãs). Rejeitada: fragmenta a seção semântica única do Admin Center (journeys §22) sem ganho de tarefa; navegação cruzada elegibilidade↔tipo piora.

**H3:** editor de tipo em formulário único (base+governança+templates numa gravação). Rejeitada: funde três domínios ETag distintos (journeys §24: `DocumentType base`, `DocumentType governance`, `DocumentType eligible-template set` são verdades If-Match separadas); um 412 parcial ficaria irrepresentável com honestidade.

Critérios decisivos para H1: preservação de contexto, fronteiras de lost-update visíveis por seção, aderência ao frame Admin Center já LOCKED, escala (paginação de cursor server-owned).

### 3.1 Decisão de superfície de escrita da elegibilidade

Elegibilidade é estado do **tipo** (`eligible-templates`, ETag do tipo). A lente Modelos apresenta `eligible_document_type_ids` como leitura + navegação para o tipo dono; a escrita ocorre apenas na seção do tipo (op41). Evita mutação fan-out multi-recurso sem dono e mantém um único domínio If-Match por gravação. A lente Modelos escreve somente o papel de Template do Documento (op51, recurso do Documento).

### 3.2 Requisitos → disposição

| Requisito | Fonte | Disposição |
|---|---|---|
| Lista de tipos: code, name, scope, active; cursor visível | op34 `DocumentTypePage` | PRESENT-IN-AUTHORITY |
| Criar tipo: code, name, scope, active + governance + representation | op35 `CreateDocumentTypeRequest` / `IDEMPOTENT_CREATE` | PRESENT-IN-AUTHORITY |
| Editar base com If-Match; 412 stale | op36/37 + journeys §24 | PRESENT-IN-AUTHORITY |
| Governança: `no_human_approval` sem steps; `use_governance_route` ≥1 step ordenado; selector NAMED_USER/GROUP; `due_in_days?` positivo inteiro | wire §3.1 + governance-step-deadline §4.1 | PRESENT-IN-AUTHORITY |
| Representação: `source_only` / `require_official_rendition(pdf)` | wire §3.1 | PRESENT-IN-AUTHORITY |
| Modelos elegíveis: conjunto de refs estáveis, vazio válido | op40/41 | PRESENT-IN-AUTHORITY |
| Preview de numeração sem reserva; `area_id` quando scope=DOCUMENT_TYPE_AREA; `validation.failed` | op42 + journeys §7 | PRESENT-IN-AUTHORITY |
| Lista transversal de Modelos: role flag, efetivo, título condicional, tipos elegíveis | op43 `TemplateConfigurationItem` (título presente iff `has_effective_revision`) | PRESENT-IN-AUTHORITY |
| Papel de Template por Documento com If-Match | op50/51 | PRESENT-IN-AUTHORITY |
| Sem vazamento de conteúdo/histórico na administração de Template | journeys §23 | PRESENT-IN-AUTHORITY (fronteira negativa) |
| Pickers de User/Group/Area | op6/op22/op16 | PRESENT-IN-AUTHORITY |

### 3.3 B12-F1 — imutabilidade de code/scope não é legível antes da escrita

Journeys §7: `DocumentType code + numbering scope` tornam-se imutáveis após o primeiro Documento comprometido usar o tipo. `DocumentTypeView` não projeta nenhum indicador de uso/imutabilidade; o frontend não pode inferir (fixture não pode mascarar — método §14.1).

- Necessidade humana: o admin saber **antes** de editar que code/scope estão congelados.
- Autoridade atual: a tentativa de escrita falha com problema de estado (`S`) do op37 — recuperável, não destrutivo, zero mutação.
- Classificação proposta: **NÃO-BLOQUEANTE para P8** — o P8 apresenta honestamente o caminho de falha de estado e não simula um flag inexistente.
- **Adjudicação do operador (2026-08-26): `REJECTED` — tentar-e-falhar basta.** O 409 é recuperável, não destrutivo e zero-mutação; editar code/scope de tipo já em uso é raro. Nenhum reopen upstream; o P8 mantém o caminho honesto de falha de estado.

Nenhum outro requisito material carece de autoridade. **Nenhum finding bloqueante aberto → P8 pode iniciar.**

## 4. P8 candidate scope

Interações materiais operáveis: lentes; master-detail; criação (op35) com conflito 409 de code duplicado e recuperação de resultado ambíguo (P10-S5); edição base/governança/elegíveis com 412 stale (journeys §24); editor de steps com reordenação por teclado (sem drag), validações locais (≥1 step, dias inteiros positivos); preview de numeração com seleção de Área e `validation.failed`; lente Modelos com toggle de papel via op51; travessia de cursor visível (P10-S3); 403/404 disclosure-safe (P10-S6); responsivo/teclado (P10-S7).

Fora de escopo (fronteiras explícitas): conteúdo/histórico de documento (B03/B07), casos de governança (B06), Organização (B10), Acesso (B11).

Artefato: `t11-b12-document-governance-p8.html` — HTML único auto-contido (CSS/JS inline), operável direto no navegador. Evidence temporário; nunca entra em candidato de merge/`main`.

## 5. Walkthrough R1 → revisão R2

Operador operou o R1 (2026-08-26). Resultado: estrutura aprovada em geral; dois findings de P8 (interação/clareza, categoria 1 do §26 — nenhum upstream):

- **R1-W1 — representação confusa dentro da rota:** `source_only`/PDF não era compreensível junto dos steps. R2: seção própria "Publicação oficial" no detalhe do tipo (leitura + botão de edição) e subseção rotulada no editor, com explicação em linguagem humana ("o que os leitores veem quando o documento vale"). Honestidade preservada: representação e rota permanecem **um único domínio If-Match (op38/39)** e gravam juntas; a separação é apenas de apresentação.
- **R1-W2 — "Modelos" sem definição:** a lente não explicava o que é um modelo. R2: explicação em linguagem humana no topo da lente (documento do acervo com papel de modelo; ponto de partida na criação; utilizável quando tem revisão efetiva e elegibilidade no tipo).

Layout/estética foram notados como mutáveis — P8 não congela paleta/tipografia (§14.4); ficam para P13.

Status R2: aprovado pelo operador **exceto a lente Modelos** (finding abaixo).

## 6. Walkthrough R2 — lente Modelos (P6/P7 dedicado)

### 6.1 Evidence do operador

Repositório antigo (Evidence de referência, não autoridade): área de templates navegável, distinguindo **templates gerais** vs **templates específicos de um tipo de documento** (ex.: template só para PO). No R2 a lente é uma lista plana paginada; o operador não reconhece o trabalho ("estou vinculando template a tipo?") e questiona a organização em escala (muitos templates × muitos tipos), pedindo padrão de plataforma madura.

### 6.2 P6 — referências (classe: galeria/administração de templates)

Padrão dominante em plataformas maduras (Confluence template gallery, SharePoint content types + templates, Google Workspace template gallery, suites de documento controlado): a coleção de templates é **organizada pela categoria de uso** (aqui: tipo de documento), com **busca por nome** e distinção visível entre template de uso amplo e template restrito. Navegação plana por código só funciona em acervos pequenos.

```text
SOURCE OBSERVATION  galerias agrupam por categoria/tipo + busca por nome
INFERENCE           o job do admin é "ver os templates de um tipo" e "achar um template pelo nome"
PRODUCT DECISION    a lente Modelos deve permitir recorte por tipo elegível + busca, servidos pelo servidor
```

### 6.3 P7 — hipóteses da lente Modelos

- **H1 (leading):** lente Modelos com **recorte por tipo de documento** ("todos os tipos" | tipo específico | "sem elegibilidade"), **busca por code/título efetivo** e filtro por papel/revisão efetiva — tudo **server-side** sobre op43. Generalidade apresentada como fato derivado: "elegível em N tipos" (chips), sem novo conceito de Produto.
- **H2 (rejeitada):** manter lista plana e organizar no cliente — desonesta: o cliente só vê a página corrente do cursor; agrupar/filtrar client-side mascara o conjunto real (proibido: fixture como verdade).
- **H3 (rejeitada):** navegar templates somente via conjunto elegível de cada tipo (op40) — cobre a direção tipo→templates mas perde o job transversal (achar template pelo nome, ver templates sem elegibilidade nenhuma, administrar papel).

### 6.4 B12-F2 — UPSTREAM FINDING (bloqueante para a lente Modelos)

```text
necessidade humana provada    admin encontra/organiza templates por tipo e por nome em escala
autoridade atual              op43 = PAGED, ordem document.code, SEM filtro/busca server-side
classificação                 INSUFICIENTE → UPSTREAM FINDING (método §25: não degradar a coleção
                              ao que o endpoint atual suporta)
menor reopen proposto         precisão de leitura do op43 (espelho do precedente B11-F1/op31):
                              filtros server-side q (code/título efetivo),
                              eligible_document_type_id, template_role, has_effective_revision.
                              Nenhuma operação nova, nenhuma escrita nova, censo permanece 89.
conceito "template geral"     REJECTED como semântica nova — generalidade permanece fato derivado
                              da elegibilidade explícita por tipo (modelo de controle do Produto);
                              a UI a apresenta, não a inventa
templates por Área            REJECTED — templates não são recurso de Área na autoridade atual;
                              Evidence antigo referia-se a tipos/uso, não a escopo de Área
```

**Adjudicação do operador (2026-08-26): RATIFICADO.** O reopen de leitura do op43 é autoridade bounded em `docs/decisions/template-configuration-read.md`; "template geral" permanece fato derivado (sem conceito/flag novo). O P8 R3 refaz a lente Modelos com recorte por tipo elegível + busca `q` + filtros de papel/efetividade, todos server-side sobre fixtures determinísticas.
