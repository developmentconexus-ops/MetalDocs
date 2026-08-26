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
