# T11 — Alinhamento de coesão pós-P11 (pré-P12)

> **Status:** OPEN — operator-initiated. Methods: Engineering v1.0.0 + Frontend v2.3 (§5.3 bounded rebaseline, §3.10A).
> **Implementation:** BLOCKED.

## FP2-F1 — Criação de Documento sem superfície própria (COVERAGE GAP)

- Evidence: operador + inspeção — B02 delega "Novo documento / Criar a partir de modelo" a um destino dono da criação ("separate create route/owner"), e nenhum bloco B01–B12 desenhou essa rota. Backend PRESENT-IN-AUTHORITY: op44 `getDocumentCreationOptions` + op46 `createDocument` (blank/template, journeys §14).
- Disposição (operador, 2026-08-26): **abrir B13 — Criação de Documento** pelo P6–P10; rebaseline bounded do inventário de blocos (método §5.3). Legado pré-reset pode ser consultado como referência P6 (proveniência, nunca autoridade; gate technical-baseline).
- Sequência: B13 e FP2-F2 fecham antes de abrir FP3/P12 (decisão do operador).

## FP2-F2 — Escopo de formatos de conteúdo (UPSTREAM FINDING — MATERIAL)

- Evidence do operador: a plataforma não deve se restringir a .docx/pdf; edição dentro do app apenas para .docx; demais formatos por upload.
- Autoridade atual CONTRADIZ: `ContentFormat = docx | pdf` fechado (wire §2/§3, content-integrity §3/§18); journeys "SourceOnly DOCX"; rendição oficial pdf a partir de docx.
- Busca exaustiva (docs atuais + histórico Git completo): nenhuma decisão multi-formato sobreviveu ao clean-slate. Classificação: **decisão perdida no reset → exige re-ratificação como bounded Product reopen** (não é reuso silencioso de memória).
- Aguardando adjudicação de escopo do operador (perguntas registradas na sessão) antes de redigir a decisão bounded.
