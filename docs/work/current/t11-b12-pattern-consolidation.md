# T11 — B12 Document Governance Administration — P10 Pattern Consolidation

> **Status:** COMPLETE após P9. Temporary branch work, non-authoritative.
> **Regra do método (§16/§3.12):** compartilhar apenas comportamento semântico/protegido repetido observado em blocos LOCKED; nada especulativo.

## 1. Padrões herdados reobservados (2ª+ ocorrência — reforçam graduação)

| Padrão | Origem | Ocorrência B12 |
|---|---|---|
| P10-S1 shell/rota Admin Center | B01/B10/B11 | `/admin/document-governance` terceira seção; shell intacto |
| P10-S2 chrome de notificações | B01N | herdado sem alteração |
| P10-S3 travessia de página do servidor visível | B10/B11 | listas op34/op43/pickers: página exata, cursor, falha de continuação preservando página + retry |
| P10-S5 criação idempotente + recuperação ambígua | B11 (op32) | op35: mesma Idempotency-Key no replay, composição congelada, mutação 1→1 |
| P10-S6 negação disclosure-safe | B10/B11 | 403 painel (nunca coleção vazia); 404 sem prova de existência |
| P10-S7 responsivo/teclado estrutural | B10/B11 | drawer, empilhamento, reordenação por botões, `<dialog>` semântico |
| 412 stale por domínio → reler/reaplicar | B10/B11 (journeys §24) | três domínios do tipo + papel de modelo, zero mutação |
| Filtro server-side antes da paginação + identidade de primeira página | B11 (op31) | op43 ratificado (`template-configuration-read.md`) — 2ª ocorrência da classe |

A 2ª ocorrência do padrão de filtro-servidor (op31 → op43) e a reobservação de S3/S5 fortalecem a obrigação de graduação aberta do B11 (roadmap: absorver a lei de realização coleção/idempotência/fixture-truthfulness em `docs/architecture/frontend.md` antes de FP2/P11). A graduação permanece tarefa da obrigação B11, agora com Evidence de dois blocos.

## 2. Padrões locais B12 — candidatos novos (1ª ocorrência, NÃO graduam)

| Candidato | Comportamento protegido | Disposição |
|---|---|---|
| Domínios If-Match separados por seção no mesmo recurso | um 412 por seção nunca contamina as irmãs; cada seção lê/grava seu ETag | CANDIDATE — aguardar 2ª ocorrência |
| Separação apresentação vs domínio de gravação (Publicação oficial dentro de op39) | seção de leitura própria + diálogo declara verdade una | CANDIDATE |
| Terminador de fronteira com contexto de permissão (detalhe do modelo → B03/B07) | navegação nomeia bloco destino E a permissão que governa lá | CANDIDATE — comparar com fronteiras B05/B08 no P11 |
| Preview sem reserva (op42 `reservation:false`) | leitura que simula efeito sem prometer efeito | CANDIDATE |
| Editor de rota sequencial (steps ordenados, selector, due_in_days) | específico do domínio; sem segunda superfície | LOCAL — não compartilhar |

## 3. Falsas abstrações verificadas

```text
nenhuma cópia cosmética promovida; chips de elegibilidade (B12) ≠ chips de permissões (B11): semânticas distintas, sem componente compartilhado implicado
diálogo de detalhe do modelo ≠ master-detail de tipo: um é projeção §23 bounded, outro é superfície de escrita — não unificar
```

## 4. Saída

Nenhuma alteração de autoridade nesta passada. Consolidação global/reconciliação final permanece no fechamento (P10 terminal) e a graduação da lei de realização permanece obrigação pré-FP2 do B11.
