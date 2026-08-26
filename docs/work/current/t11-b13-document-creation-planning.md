# T11 — B13 Document Creation — P6/P7 Planning

> **Status:** CANDIDATE / R1 — temporary branch work, non-authoritative.
> **Methods:** Engineering Method v1.0.0 + Frontend Product Experience Planning Method v2.3.
> **Trigger:** FP2-F1 (coverage gap found by the operator after P11).
> **Implementation:** BLOCKED by `docs/roadmap.md`.

## 1. Block ledger

| Field | Value |
|---|---|
| Block | B13 — Document Creation |
| Route | `/documents/new` (destino dono da criação; B02 já delega "Novo documento" / "Criar a partir de modelo" a uma rota/dono separado) |
| User goals | autor cria um Documento controlado — em branco, a partir de modelo elegível, ou anexando um arquivo-fonte — com tipo, área, título e responsável corretos |
| Authority pack | `product/contract.md` §4/§5A; `product/journeys.md` §14 blank/template create; `wire-contract.md` §2.7 (projeções completas de op44), ops 44/46 + 58/59/60; `decisions/content-format-vocabulary.md` (RATIFICADA); `architecture/responsible-owner.md`; `architecture/frontend.md` §19 |
| Primary operations | op44 `getDocumentCreationOptions`, op46 `createDocument` |
| Composed operations | op59 `startRevisionDraftUpload` → provider PUT → op60 `completeRevisionDraftUpload` → op58 `updateRevisionDraft` (anexo do arquivo-fonte na revisão criada) |
| Dependencies | B01 shell (P10-S1), B01N chrome (P10-S2), padrões P10-S3/S5/S6/S7; fronteira de saída para B04 (Documento em Edição) |
| Permission truth | `document.create`; `document.owner.manage` decide a presença de candidatos a responsável (§2.7) — co-location nunca funde permissões |

## 2. P6 — Reference study (gatilho real)

Ambiguidade consequente: a ordem das escolhas (tipo-primeiro vs área-primeiro), e onde o arquivo entra. Referências por padrão de tarefa:

```text
SOURCE OBSERVATION  SharePoint "New" abre pelo content type; o tipo determina o template aplicado
SOURCE OBSERVATION  M-Files "Create object" é class-first; a classe determina metadados obrigatórios
SOURCE OBSERVATION  Google Drive/Confluence separam "documento em branco" de "galeria de modelos"
                    como pontos de entrada distintos, não como um seletor escondido no formulário
SOURCE OBSERVATION  Drives maduros aceitam arrastar-e-soltar o arquivo como entrada de criação,
                    não como segundo passo

INFERENCE           o tipo é a escolha estruturante (governa rota, numeração, publicação e modelos
                    elegíveis); o modo de partida (branco / modelo / arquivo) é a segunda escolha
                    material e merece visibilidade própria

PRODUCT DECISION    tipo + área primeiro (verdade estruturante do servidor), depois modo de partida
                    explícito com três opções visíveis, depois título e responsável
```

Disposições de capacidades da classe de referência:

```text
criação em massa / múltiplos arquivos       REJECTED — journeys: um Documento = uma identidade
                                            governada; lote não é jornada aceita no Launch
pastas / hierarquia de arquivos             REJECTED — contract §1: não é file drive; Área é o
                                            escopo organizacional aceito
escolher o code manualmente                 REJECTED — journeys §7: numeração é server-owned
duplicar documento existente                REJECTED — Template é o mecanismo aceito de semente
salvar rascunho do formulário               REJECTED — nada é governado antes de op46; o formulário
                                            é estado de cliente puro
```

## 3. P7 — Hipóteses de layout

**H1 (leading): formulário único progressivo, dirigido pela verdade do servidor.**
Uma rota, uma coluna de decisão em ordem: (1) Tipo, (2) Área, (3) Modo de partida — em branco | de modelo | enviar arquivo, (4) Título, (5) Responsável. Cada seleção re-lê op44 com os filtros correspondentes, exatamente como a §2.7 prescreve.

**H2: assistente multi-passo (wizard).** Rejeitada: a §2.7 já dá disclosure progressivo dentro de uma superfície; passos separados escondem decisões interdependentes (trocar o tipo depois de escolher o modelo invalida o modelo) e pioram a correção de erro.

**H3: três rotas separadas (novo em branco / de modelo / upload).** Rejeitada: fragmenta uma única intenção humana ("criar este documento") e triplica a superfície de tipo/área/título; a escolha de partida é um campo, não um destino.

### 3.1 Lei de disclosure progressivo (op44 §2.7) — dirige o layout

```text
sem filtros            → todas as Áreas e Tipos utilizáveis; templates=[]; candidatos ausentes
+ document_type_id     → templates elegíveis/efetivos/legíveis do tipo; candidatos ainda ausentes
+ area_id              → candidatos a responsável presentes IFF o ator tem owner.manage no escopo
ambos                  → Área/tipo exatos + templates do tipo + candidatos conforme escopo
```

Consequências obrigatórias de UX:

```text
a lista de modelos NÃO existe antes de um tipo escolhido — a UI não inventa lista vazia como
  "nenhum modelo disponível" antes da seleção; ela declara que a escolha do tipo a revela
o seletor de responsável só aparece quando a autoridade o fornece; ausência de candidatos
  significa "o servidor não delega essa escolha aqui", nunca "não há pessoas"
default_responsible_owner é o ator atual, vindo do servidor — nunca inferido no cliente
arrays completos (§2.7): sem paginação, sem truncamento silencioso, sem total inventado
```

### 3.2 Anexar arquivo na criação — composição honesta (adjudicado pelo operador)

A autoridade **não** tem uma operação "criar com arquivo". O trabalho humano único compõe-se de:

```text
op46 createDocument  → 201 { document_id, revision_id }   (Documento + REV000 DRAFT existem)
op59 startRevisionDraftUpload → provider PUT → op60 complete → op58 attach
```

Disposição: **PRESENT-IN-AUTHORITY por composição** — nenhuma operação nova é necessária; isto é exatamente `journeys §14` (blank create → REV000 DRAFT → WorkingContent). Mas a composição tem um estado intermediário real que o P8 deve representar com honestidade:

```text
op46 OK + upload falha  →  o Documento EXISTE como REV000 DRAFT vazio; nada é perdido nem
                           "revertido"; a recuperação continua na tela do Documento (B04)
                           A UI nunca finge que a criação falhou, nem oferece "tentar de novo"
                           que criaria um segundo Documento
op46 ambíguo            →  replay com a MESMA Idempotency-Key (frontend.md §19.2); nunca
                           uma segunda intenção
```

### 3.3 Requisitos → disposição

| Requisito | Fonte | Disposição |
|---|---|---|
| Áreas/Tipos utilizáveis completos, sem paginação | op44 + §2.7 | PRESENT-IN-AUTHORITY |
| Modelos elegíveis/efetivos do tipo, com identidade da revisão efetiva | op44 `TemplateCreationOption` | PRESENT-IN-AUTHORITY |
| Responsável: default do servidor + candidatos condicionados a `owner.manage` | op44 + §2.7 + `responsible-owner.md` | PRESENT-IN-AUTHORITY |
| Criar com tipo/área/título + modelo opcional + responsável opcional | op46 `CreateDocumentRequest` | PRESENT-IN-AUTHORITY |
| Criação idempotente com replay de resultado ambíguo | op46 `IDEMPOTENT_CREATE` + frontend §19.2 | PRESENT-IN-AUTHORITY |
| Anexar arquivo-fonte como parte do mesmo trabalho humano | ops 59/60/58 compostos | PRESENT-IN-AUTHORITY (composição) |
| Formatos aceitos + rejeição estrutural/malware nomeada | `content-format-vocabulary.md` §4/§4.1 | PRESENT-IN-AUTHORITY |
| Fonte não conversível sob tipo que exige PDF | `content-format-vocabulary.md` §6 | PRESENT-IN-AUTHORITY (ver questão aberta 3.4) |
| `code` alocado pelo servidor, nunca escolhido | journeys §7 | PRESENT-IN-AUTHORITY (fronteira negativa) |
| 404 para área/tipo pedido e não divulgável | §2.7 | PRESENT-IN-AUTHORITY |
| Saída para o documento criado | B04 | fronteira de bloco |

Nenhum requisito material carece de autoridade. **Nenhum finding bloqueante aberto → P8 pode iniciar.**

### 3.4 B13-Q1 — questão UX aberta (adjudicação por walkthrough)

`content-format-vocabulary.md` §6 fixa a lei semântica e deixa o **ponto de realização** para este bloco:

```text
A  restringir na criação  — o tipo que exige PDF oficial só oferece/aceita fontes conversíveis;
   o autor descobre antes de enviar bytes; menos frustração, mais acoplamento tipo↔upload na tela

B  falhar em gate posterior — aceita o arquivo, falha com problema nomeado ao submeter;
   permite rascunhar livremente; a descoberta é tardia
```

O P8 R1 apresenta **A** como hipótese líder (custo de erro menor, coerente com o disclosure progressivo já dirigido pelo servidor) e mantém **B** operável via controle de fixture, para o operador comparar as duas experiências antes de decidir.

## 4. P8 candidate scope

Interações materiais: seleção progressiva tipo→área com re-leitura de op44; três modos de partida; galeria de modelos condicionada ao tipo; seletor de responsável condicionado a escopo/permissão; upload com progresso, rejeição de formato/estrutura/malware nomeadas, expiração de upload (410) e recuperação; criação idempotente com 409/ambíguo/replay; falha parcial pós-criação apresentada honestamente; 403/404 disclosure-safe; responsivo/teclado.

Fora de escopo: edição de conteúdo (B04), governança/aprovação (B06), administração de tipos/modelos (B12), descoberta (B02).

## 5. P8 — artefato, conformidade de padrão e prova

Artefato: `t11-b13-document-creation-p8.html` — HTML único auto-contido (CSS/JS inline).

### 5.1 R1 REJEITADO — desvio de padrão

**REJECTED — wrong representation medium (método §4).** O R1 inventou vocabulário visual próprio (cartões numerados de passo, `.step`/`.modes`/`.opt`/`.drop`, tokens de cor próprios, `window.prompt` como seletor de arquivo) em vez de realizar o padrão low-fi já LOCKED em B10/B11/B12. Achado do operador. A rejeição atinge o artefato, não o requisito de Produto.

### 5.2 R2 — conformidade ao padrão canônico

O R2 reusa **verbatim** a folha de estilo e o esqueleto dos blocos LOCKED (recuperados de `evidence/t11-b12-p8-r4-locks-20260826`):

```text
tokens          --bg --paper --ink --muted --line --soft --soft-strong (idênticos)
esqueleto       review > review-head > note > fixture > shell(top/side/main) > content
                > page-head > split > panel(panel-head/panel-body) > section(section-head)
componentes     list/list-button, mode-choice, picker, form-grid, field-hint, read-only,
                review-box, state-line, pill, empty, result-box, ambiguous, actions,
                boundary-note, foot, mobile-hint, failure-panel
escrita          <dialog> canônico (dialog-head/body/foot + data-close) no lugar de window.prompt
negação          a regra canônica enumera IDs de painel por bloco; B13 estende a enumeração
                exatamente como B12 estendeu a de B11
navegação        aside/side canônico reusado, com "Novo documento" como item ativo
deltas locais    apenas .drop, .hypo e .tpl-row — sem redefinir token ou componente existente
```

Semântica de rádio (grupos de escolha única) mantida do R1: `role="radiogroup"`/`role="radio"`, `aria-checked`, seleção idempotente, tabindex rotativo e navegação por setas — acessibilidade é estrutural (§3.13/§24).

### 5.3 Prova automatizada (Chromium headless) — 30/30

```text
conformidade de padrão (7)   shell canônico, split com dois painéis, sections com section-head,
                             mode-choice, form-grid, <dialog> nativo, ausência de classes ad-hoc
disclosure progressivo (5)   gate antes do tipo, modos após o tipo, responsável somente-leitura
                             antes da área, seletor após a área, ausência de owner.manage
seleção e teclado (2)        3 modelos para FRM; navegação por setas nos tipos
criação (4)                  habilitação, código alocado pelo servidor + REV000, fronteira B04,
                             congelamento na ambiguidade
idempotência (1)             replay com a mesma chave mantém mutação 1 → 1
conteúdo (4)                 lista de aceitos, malware nomeado, estrutura inválida com macro
                             nomeada, formato derivado dos bytes
falha parcial (1)            documento existe, sem rollback falso, sem repetição que duplicaria
hipóteses B13-Q1 (4)         A exclui não conversível na lista e no diálogo; B aceita e avisa;
                             voltar para A limpa o arquivo incompatível com explicação
negativos (2)                403 substitui a superfície; 404 não divulgante
```

Zero erros de console.

Status: **R2 CANDIDATE — aguardando operação do operador.** Decisão pendente: **B13-Q1** (hipótese A vs B), comparável no próprio wireframe.

## 6. Validação de cobertura pedida pelo operador (2026-08-26)

### 6.1 B13-F1 — violação de fixture-truthfulness no R2 (achado próprio, MATERIAL)

Ao validar "todas as informações necessárias estão na página", conferi a projeção exata do op44:

```text
DocumentTypeReference { document_type_id, code, name }
AreaReference         { area_id, code, name }
TemplateCreationOption { document:DocumentReference, effective_revision:RevisionReference }
DocumentReference     { document_id, code }        // título pertence à Revisão
RevisionReference     { revision:RevisionIdentity, title }
```

O R2 exibia, na lista de tipos, `numeração por tipo + Área` e `exige PDF oficial`, e na lista de modelos, `fonte Excel (.xlsx)`. **Nenhum desses campos existe na projeção.** Isso viola a lei de realização já graduada em `architecture/frontend.md` §19.3 (fixture nunca simula verdade que o contrato não fornece) — a mesma lei que este programa acabou de absorver. Corrigido no R3.

### 6.2 B13-F2 — hipótese A do B13-Q1 não é realizável com a autoridade atual (UPSTREAM FINDING)

Consequência direta de 6.1: para **restringir na criação** (hipótese A) o cliente precisaria saber (a) que o tipo exige PDF oficial e (b) qual o formato-fonte de cada modelo. `op44` não projeta nem um nem outro.

```text
hipótese A   exige precisão de leitura no op44:
             + representation policy no tipo projetado
             + content_format em TemplateCreationOption
             classificação: UPSTREAM FINDING (leitura), espelhando op31/op43
hipótese B   realizável hoje sem nenhuma mudança: o servidor rejeita no gate com
             problema nomeado (content-format-vocabulary.md §6 já garante a lei semântica)
```

O R3 passa a operar **B como padrão realizável**, e mantém A operável apenas sob um controle de fixture que declara explicitamente que simula leitura ainda não ratificada. O operador decide: ratificar a precisão de leitura (habilita A) ou aceitar B.

### 6.3 Visibilidade / audiência do documento — Global Maximum

Evidence do operador: o legado permitia escolher, por documento, quem veria — Área, empresa toda, externo, pessoas específicas.

Mapeamento contra a autoridade atual (`architecture/authorization-and-audit.md` §2/§9, `domain-model.md`, T3):

| Opção do legado | Autoridade atual | Disposição |
|---|---|---|
| Da Área | O Documento pertence a uma Área; a leitura exige `document.read_effective` **naquela Área**. A escolha da Área **é** a decisão de audiência | PRESENT-IN-AUTHORITY (por outro mecanismo) |
| Empresa toda | Concessões de escopo Company leem todas as Áreas | PRESENT-IN-AUTHORITY (administrado em B11) |
| Pessoas específicas | Exigiria ACL por documento. T3: sujeito = `User\|Group`, escopo = `Company\|Area`; §2 lista `materialized ACL` como **sem autoridade semântica persistida** | **CONTRADIZ A AUTORIDADE** — só entraria por reopen material de Produto |
| Documento externo | Ambíguo: leitor externo (não existe identidade externa no Launch — uma empresa, usuários autenticados) ou documento de origem externa (classificação, não acesso). `External Repository` é Future | **AGUARDA DESAMBIGUAÇÃO DO OPERADOR** |

**Global Maximum.** O caminho tentador — um seletor de audiência por documento — criaria uma segunda autoridade de autorização em conflito com T3, permitiria que autores contornassem o acesso administrado centralmente (regressão de governança num sistema de documento controlado), exigiria operação nova (censo proíbe op90+ sem reopen) e romperia `no materialized ACL`. Custo alto, ganho negativo.

A necessidade humana real provada é **informacional**, não de capacidade: *o autor escolhe a Área sem saber que essa é a decisão de quem poderá ler*. A menor correção sustentável:

```text
declarar a consequência na própria tela de criação, com verdade já disponível:
  "Quem poderá ler quando efetivo: quem tiver permissão de leitura na Área X
   ou em toda a Empresa. O acesso é administrado em Acesso, não escolhido aqui."
zero operação nova · zero conceito novo · zero widening de disclosure
```

Se o operador provar que o autor precisa da **lista concreta** de leitores, isso vira upstream finding próprio (leitura disclosure-safe de audiência), com custo real: expor atribuições de acesso a não-administradores. Não é assumido aqui.

Implementado no R3: seção "Quem poderá ler" derivada da Área escolhida, mais o pop-up canônico de resultado da criação pedido pelo operador.

### 6.4 Auditoria de completude da tela (pedido do operador)

Confronto entre `CreateDocumentRequest` + a jornada aceita e o que a tela apresenta:

```text
document_type_id            presente (lista op44)
area_id                     presente (lista op44)
title                       presente (metadado da revisão, explicado)
template_document_id?       presente (modo "a partir de modelo", lista op44)
responsible_owner_user_id?  presente, condicionado a owner.manage (op44 §2.7)
arquivo-fonte               presente por composição (ops 59→60→58), com estado intermediário honesto
code                        AUSENTE por autoridade — alocado pelo servidor; a tela declara isso
audiência/leitores          NÃO É CAMPO — é consequência da Área; agora declarada explicitamente
numbering scope             não projetado por op44 → não exibido (B13-F1)
representation policy       não projetado por op44 → não exibido (B13-F1)
formato do modelo           não projetado por op44 → não exibido (B13-F1)
rota de governança          op38 é superfície de document_type.manage; um autor comum não a lê →
                            deliberadamente ausente, sem inventar leitura
```

Nenhum campo material do comando aceito falta na tela. Os ausentes são ausências **de autoridade**, declaradas em vez de simuladas.

### 6.5 Prova do R3 (Chromium headless) — 17/17

```text
fidelidade de fixture     linhas de tipo e de modelo sem campos não projetados          PASS (4)
leitura simulada          revela campos com rótulo explícito, habilita A, desligar volta a B  PASS (4)
audiência                 pede Área primeiro; declara consequência + administração central PASS (2)
pop-up de resultado       diálogo canônico com código, REV000, audiência, fronteira B04  PASS (4)
regressões                replay ambíguo 1→1, honestidade da falha parcial, 403           PASS (3)
```

Zero erros de console.

