# T11 — Alinhamento de coesão pós-P11 (pré-P12)

> **Status:** OPEN — operator-initiated. Methods: Engineering v1.0.0 + Frontend v2.3 (§5.3 bounded rebaseline, §3.10A).
> **Implementation:** BLOCKED.

## FP2-F1 — Criação de Documento sem superfície própria (COVERAGE GAP)

- Evidence: operador + inspeção — B02 delega "Novo documento / Criar a partir de modelo" a um destino dono da criação ("separate create route/owner"), e nenhum bloco B01–B12 desenhou essa rota. Backend PRESENT-IN-AUTHORITY: op44 `getDocumentCreationOptions` + op46 `createDocument` (blank/template, journeys §14).
- Disposição (operador, 2026-08-26): **abrir B13 — Criação de Documento** pelo P6–P10; rebaseline bounded do inventário de blocos (método §5.3). Legado pré-reset pode ser consultado como referência P6 (proveniência, nunca autoridade; gate technical-baseline).
- Sequência: B13 e FP2-F2 fecham antes de abrir FP3/P12 (decisão do operador).

## FP2-F2 — Escopo de formatos de conteúdo (REALIZATION-LAYER CONTRADICTION)

### Correção de diagnóstico (leitura estratégica do Produto, 2026-08-26)

A busca inicial procurou "decisão multi-formato" e não achou. Errado o alvo: **o Contrato de Produto nunca fechou formato** — ele é deliberadamente agnóstico:

```text
contract §4 Rendition   "A Document Type may be source-only or require one derived official
                        representation SUCH AS PDF"        → "such as" = exemplo, não fechamento
contract §4 Submission  "freezes exact content"            → conteúdo, sem formato prescrito
contract §1 North Star  perguntas do produto não mencionam formato
contract §6 Launch Core "source/official representation read/download"  → sem lista de formatos
```

O fechamento `docx | pdf` existe **apenas na camada de realização**:

```text
wire-contract §2      ContentFormat = docx | pdf        (enum fechado)
content-integrity §3  content_format = closed vocabulary
journeys              "SourceOnly DOCX"
```

E a obrigação `CNT-14 — PRESERVE` diz: *"EigenPal may remain selected DOCX adapter/provider evidence; **it never owns semantic truth**"* — ou seja, o adaptador DOCX é mecanismo de **edição**, explicitamente proibido de virar verdade semântica.

**Classificação corrigida:** não é decisão de Produto perdida nem escopo novo. É **contradição da camada de realização com a autoridade de Produto** — exatamente o padrão `NO backend-shaped UX` (Frontend Method §3.10A): um vocabulário técnico fechado estreitou a intenção de Produto sem decisão de Produto que o autorizasse.

**Reopen necessário é menor do que o previsto:** precisão bounded do vocabulário `ContentFormat` (wire + content-integrity), preservando toda a lei de bytes exatos, malware gate, SHA-256 e admissão. Nenhum conceito novo de Produto.

### Adjudicações do operador (2026-08-26)

```text
F2a formatos-fonte     LISTA FECHADA AMPLIÁVEL — vocabulário maior (docx, xlsx, pptx, pdf,
                       imagens, txt/csv...), detecção real de formato server-side,
                       malware gate para todos, ampliação futura = decisão pequena
F2b publicação PDF     PDF oficial só onde há conversor; formatos não conversíveis usam
                       "somente fonte"; a regra fica visível na criação/config
F2c não controlado     RESOLVIDO PELA AUTORIDADE EXISTENTE — ver abaixo, sem reopen
F2d edição no app      DOCX apenas (CNT-14 EigenPal); demais formatos = upload/download
```

### F2c — resolvido pela autoridade existente (nenhuma decisão nova)

```text
contract §1 North Star   "Launch V1 is NOT a generic ECM/FILE DRIVE"    → classe não controlada
                         está explicitamente fora do produto
contract §4              todo documento é Controlled Document (identidade estável + revisões)
contract §4 Governance   NoHumanApproval cobre "não precisa de aprovação" SEM deixar de ser
                         controlado (numerado, efetivo, encontrável, auditável)
contract §6 Future       lista futura não contém classe de arquivo não controlado
```

Portanto: **não existe classe "documento não controlado"**; o trabalho humano correspondente é atendido por Document Type com rota `NoHumanApproval`. Isso já estava decidido; nada a reabrir.

## Próximos passos derivados

```text
1. redigir decisão bounded de precisão do vocabulário ContentFormat (F2a+F2b) → operador ratifica
2. B13 P6-P7 já pode citar a decisão pendente; P7 não fecha antes da ratificação (§13 blocking law)
3. B13 P8 inclui anexar arquivo-fonte na criação (adjudicado) + regra de publicação visível
