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
| 2.3 | F-QA4-11 `values_frozen_at` sempre null | Incluir coluna no SELECT de `GetDocument`. |
| 2.4 | F-QA4-7 `signoff_id` vazio | Devolver o id persistido. |
| 2.5 | F-QA4-6 Idempotency-Key inconsistente | Um formato só (decidir UUID vs livre) em todas as rotas irmãs. |

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
| 1 | EM ANDAMENTO | — |
| 2 | FILA | — |
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

## Log

- 2026-07-28: tracker criado; D1–D5 ratificadas; Etapa 1 iniciada.
