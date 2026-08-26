# T11 — FP2 / P11 — Integrated Low-Fidelity Product

> **Status:** R1 CANDIDATE — temporary branch work, non-authoritative.
> **Method:** Frontend Method v2.3 §17 — P11 assembles already-LOCKED blocks; it does not redesign them.
> **Implementation:** BLOCKED by `docs/roadmap.md`.

## 1. Assembly identity

Todos os 13 blocos LOCKED foram recuperados **byte a byte** dos refs de Evidence (nenhuma modificação):

| Bloco | Artefato | Origem (exact commit) |
|---|---|---|
| B01 | t11-b01-app-shell-wireframe.html | adf58e4 (evidence/t11-pr162-b01-b09-locks-20260824) |
| B01N | t11-b01-notifications-wireframe.html | adf58e4 |
| B02 | t11-b02-library-wireframe.html | adf58e4 |
| B03 | t11-b03-document-official-functional-wireframe.html | adf58e4 |
| B04 | t11-b04-document-work-functional-wireframe.html | adf58e4 |
| B05 | t11-b05-my-work-functional-wireframe.html | adf58e4 |
| B06 | t11-b06-governance-case-functional-wireframe.html | adf58e4 |
| B07 | t11-b07-document-history-functional-wireframe.html | adf58e4 |
| B08 | t11-b08-notifications-full-inbox-functional-wireframe.html | adf58e4 |
| B09 | t11-b09-audit-functional-wireframe.html | adf58e4 |
| B10 | t11-b10-organization-administration-p8.html | b8c607c (evidence/t11-pr170-b10-locks-20260824) |
| B11 | t11-b11-access-administration-p8.{html,css,js} | 75e00f3 (evidence/t11-b11-clean-r3-locks-20260825) |
| B12 | t11-b12-document-governance-p8.html | c9a4414 (evidence/t11-b12-p8-r4-locks-20260826) |

## 2. Integrador (`t11-p11-integrated-product.html`)

- Navegação global com a IA aceita do B01 (Início / Minha caixa / Documentos / Governança / Gestão / Evidência), cada entrada abrindo o P8 LOCKED do bloco em um stage isolado (iframe).
- **Deep links**: hash por rota (`#/library`, `#/admin/access`, …); reload preserva o bloco ativo — provado.
- Painel de **jornadas cross-block** (consumo, autoria, aprovação, administração, evidência) com atalhos.
- Teclado: navegação por âncoras focáveis; `aria-current` na rota ativa.

Prova automatizada (Chromium headless): 13/13 rotas renderizam o bloco, estado de navegação correto, deep-link reload OK, zero erros de console.

## 3. Costura conhecida R1 (para adjudicação do operador)

**S1 — terminadores de fronteira internos:** dentro de cada bloco, botões "continua em Bxx" permanecem os terminadores do Evidence (fidelidade byte-a-byte). A travessia real entre blocos usa a navegação global do integrador. Alternativa R2: shim de integração por bloco (postMessage → navegação do pai), ao custo de tocar as cópias dos artefatos LOCKED. Decisão do operador: fidelidade absoluta (R1) vs. travessia embutida (R2).

## 4. Fora de escopo P11

Sem redesenho de blocos; sem P12 (adversarial), P13/P14; sem implementação. Findings de integração reabrem apenas o menor escopo afetado (método §17).

## 5. Aceite do operador

**ACCEPTED — R1, 2026-08-26.** O operador operou o produto integrado e aceitou o P11 R1. Costura S1 adjudicada **como está**: os terminadores de fronteira internos dos blocos permanecem os do Evidence LOCKED; a travessia cross-block é da navegação global do integrador. Nenhum finding de integração aberto. FP2/P11 completo; FP3/P12 permanece decisão de roadmap do operador.

## 6. Prova de fluxos negativos / recuperação cross-block (método §17)

Executada via integrador em Chromium headless, operando os controles de fixture dos blocos LOCKED (sem modificação de bloco):

```text
N1a 403 na superfície de governança → painel de negação (nunca coleção vazia)   PASS
N1b recuperação cross-block via navegação global após negação                    PASS
N1c reentrada no bloco carrega contexto autorizado fresco                        PASS
N2  404 não-divulgante + instrução de recuperação no bloco                       PASS
N3a falha de continuação preserva página + retry explícito                       PASS
N3b navegação cross-block permanece operável após falha dentro do bloco          PASS
N4  deep link desconhecido cai na rota shell (navegação negativa)                PASS
N5  reload de deep link direto em superfície admin permanece operável            PASS
```

Zero erros de console no shell integrador. Estados negativos intra-bloco permanecem provados pelos P8/P9 de cada bloco LOCKED; esta prova cobre a dimensão cross-block exigida pelo método §17.
