# System-impact analysis — Materialização de conteúdo na criação de template em branco

**Date:** 2026-08-04
**Intent (one line):** criar template "do zero" deve materializar o objeto DOCX e o `content_hash` na própria operação de criação, tornando **inalcançável** o estado "versão de template sem conteúdo" que hoje produz 409 `UPLOAD_MISSING` no submit.
**Work type:** feature
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow *(see §10 — ADR required: o significado de `content_hash` como gate é uma política vigente e será redefinido)*

---

## 1. Classify & own

- **Work type:** feature (dentro de um módulo existente; nenhum módulo novo, nenhuma capability nova).
- **Owning module:** `templates` — dono da criação de template/versão, do `Presigner` port
  (`internal/modules/templates/application/ports.go:47-58`), das chaves canônicas de objeto
  (`application/keys.go`) e do gate de publicação (`application/lifecycle.go:72-74`).
- **Explicitly NOT owning:**
  - `controlleddocuments` — hoje é o único consumidor de `system/templates/blank.docx`, mas por um
    caminho próprio (bootstrap de *documento* em branco). Não é dependência deste trabalho.
  - `render` / `docx-renderer` — **não** será usado para gerar DOCX vazio (ver §2).
  - `approval` — o submit apenas *lê* o hash via porta publicada
    (`LoadTemplateVersionContentHash`); nada muda lá.
- **Cross-module edges:** nenhuma nova. O trabalho é interno a `templates` + object store.
- **Ambiguity?** Nenhuma. Sem AS-3.

## 2. Foundation verdict

- **Base:** `spawnNextDraft` (`application/lifecycle.go:166-189`) já resolve exatamente este problema
  para o caminho "nova revisão": copia o objeto **PRE-TX** ("store-then-reference: the object exists
  before the referencing row commits; the only crash outcome is a safe orphan") na chave canônica do
  próprio draft, nunca na do source. É um padrão sólido, deliberado e documentado no próprio código.
- **Sound or patchy?** A fundação é **sólida** — e é justamente por isso que o defeito fica evidente:
  o caminho "em branco" é o **único** caminho de criação de versão que não executa o
  store-then-reference. Não há nada a otimizar dentro de um remendo; há um caminho que não segue o
  padrão que o módulo já tem. Fechar o buraco = aplicar o padrão existente ao caminho que o pulou.
- **Alternativa descartada — gerar DOCX vazio no docx-renderer:** introduziria uma chamada de rede a
  outro serviço no caminho de criação, uma dependência `templates → render` que não existe hoje, e um
  segundo gerador de "DOCX vazio" concorrendo com o asset determinístico que o repo já semeia
  (`deploy/assets/system-blank.docx` via `minio-init`). `Presigner.Copy` já existe e é o primitivo
  certo. Copiar > gerar.
- **Smell de fundação encontrado (o achado real):** `content_hash` está **sobrecarregado** com dois
  significados incompatíveis:
  1. em `autosave.go:151-165` e no CHECK `chk_template_version_content_hash_non_draft`
     (`db/baseline/0001_current_schema.sql:2536`) = *"hash verificado do objeto que esta versão
     aponta"*;
  2. em `lifecycle.go:72-74` e em `spawnNextDraft` ("ContentHash is left empty ... so the publish gate
     still forces a real edit") = *"o usuário efetivamente editou"*.
  Enquanto os dois sentidos moram no mesmo campo, "versão sem conteúdo" continua sendo um estado
  alcançável **por construção** — é assim que o sentido (2) é expresso. Fechar o buraco na origem
  exige desfazer a sobrecarga, não só copiar bytes a mais.
- **AS-2?** Não — a fundação é boa; a sobrecarga é um item de design a resolver no ADR, não um
  workaround dentro do qual estaríamos otimizando.

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | Não muda | `template.create` já é exigida no create (`application/create.go:67`, ScopeTenant → areaCode `"tenant"`). Nenhuma capability nova. | `authz.SeedTxIdentity`, `authz.Require` |
| Contract-first | Talvez | Se o create precisar aceitar `starting_point` explícito (hoje o backend não sabe se é blank ou docx — o wizard decide no cliente), entra por `api/openapi/v1/openapi.yaml` + regen completo. | oapi-codegen, `go generate ./...` |
| Multi-tenant pooled | **Sim** | A cópia vai para a chave canônica tenant-namespaced via `templateDocxKey(tenantID, templateID, n)`; o source é o asset de sistema. `Presigner.Copy(ctx, tenantID, src, dst)` já carrega o tenant. | `tenant.FromContext`, `keys.go` |
| Async = transactional outbox | **Sim — núcleo** | A cópia é chamada de rede: vai **PRE-TX**, exatamente como `spawnNextDraft` (store-then-reference). Nunca dentro da tx. Outbox **não** serve aqui: materializar depois do commit recriaria a janela de "versão sem conteúdo" que este trabalho existe para eliminar. | padrão de `lifecycle.go:173-177` |
| DB enforces invariants | **Sim** | Se `content_hash` passar a ser sempre presente, o CHECK `chk_template_version_content_hash_non_draft` aperta para NOT NULL/length=64 incondicional, via **migração forward 0317** (baseline congelado pós-fold 2026-07-29). Requer decidir o destino do sentido (2) antes. | migração + CHECK |
| Cross-module via published interface | Não | Nenhuma nova aresta entre módulos; `Presigner` é porta do próprio `templates`. | — |

Sem AS-1.

## 4. Capability wiring

**N/A** — nenhuma capability adicionada ou alterada. `template.create` cobre a operação.

## 5. Module wiring

**N/A** — feature dentro do módulo `templates`.

## 6. Frameworks to reuse, not reinvent

- `Presigner.Copy` / `Presigner.Confirm` (`application/ports.go:47-58`) — cópia do asset e obtenção do
  hash verificado. `Confirm` devolve `objectstore.VerifiedPointer`, ou seja, o hash real do objeto:
  não inventar cálculo de hash paralelo.
- `templateDocxKey` (`application/keys.go`) — chave canônica; nunca montar chave à mão.
- `TxRunner.Do` — o create já roda numa tx; a cópia fica fora dela.
- `audit.RecordTx` / `newAuditEvent` — a materialização é mudança de estado registrável (o módulo já
  tem `AuditSaved`; avaliar um evento próprio de criação-em-branco).
- `problem.Write` + códigos — qualquer erro novo é problem+json; hoje `UPLOAD_MISSING` é 409
  (`delivery/http/errors.go:67-68`).
- `testdb` factory + `//go:build integration` — todos os testes de integração.

## 7. Contract & data

- **OpenAPI:** provável adição de um discriminador de ponto de partida no create de template (hoje o
  backend não distingue blank de docx — o wizard faz duas chamadas e simplesmente omite a segunda no
  caminho blank, `TemplateWizardPage.tsx:139-161`). Se o backend passar a materializar, ele precisa
  saber a intenção. Entra por spec + regen completo.
- **Migração 0317** (forward-only; baseline congelado): aperto do CHECK de `content_hash` **se** o
  ADR concluir que o campo passa a ser sempre presente. Backfill: versões draft existentes sem hash
  precisam de decisão (materializar a partir do objeto existente quando houver; para as que não têm
  objeto algum, há um conjunto pequeno e identificável em dev — em produção o sistema ainda não está
  liberado).
- **Destructive change?** Não há quebra de wire. O aperto de CHECK é o único ponto que exige backfill.

## 8. Test & QA plan

- **Framework:** `testdb` factory, `//go:build integration`.
- **Gates aplicáveis:** DB-invariant (novo CHECK + backfill), contract (spec↔gerado↔handlers se o
  create mudar), multi-tenant (chave de destino namespaced; cópia não vaza entre tenants),
  async/idempotência (recriação do mesmo template não duplica objeto; falha na cópia não deixa linha
  órfã — só objeto órfão, que é o resultado seguro por desenho).
- **Casos-chave:** criar template em branco → objeto existe na chave canônica **e** a linha nasce com
  hash; abrir o editor não dá 404; submeter sem editar tem comportamento **definido pelo ADR** (não
  mais um 409 acidental); cópia falha → nenhuma linha criada; tenant B não enxerga o objeto de A.
- **Evidence:** `go build ./...`, `go vet -tags integration ./...`, suites de templates + migrations,
  QA vivo em `:80` (criar template em branco pela UI → editor abre → submeter funciona), disposição
  de review e defers.

## 9. Docs / ADR

- **Wiki:** `wiki/modules/templates.md` — hoje **não descreve** nem o fluxo em branco nem a
  precondição de `content_hash` (lacuna confirmada). Documentar ambos + `Last verified`.
- **REQ IDs:** classe REQ-DB-INVARIANT (CHECK como linha autoritativa) e REQ-CONTRACT; citar de
  `wiki/architecture/backend-target-architecture.md` no review.
- **ADR required?** **Sim.** O trabalho redefine o significado de `content_hash` como gate de
  submit/publish — uma política vigente, deliberada e documentada em código
  (`lifecycle.go:170-172`). O ADR deve decidir explicitamente:
  1. `content_hash` passa a significar **um único** conceito: hash verificado do objeto apontado;
  2. o sentido "o usuário editou de verdade", se ainda desejado para revisões, é expresso de forma
     honesta (hash ≠ hash do source) em vez de por ausência de valor;
  3. se um template em branco **não editado** pode ser submetido (decisão de produto).

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — segue para design; ADR obrigatório; fundação sólida.
- **Open hard-stops:** nenhum (AS-1/AS-2/AS-3 livres).
- **Locked constraints para o design:**
  1. **Store-then-reference, PRE-TX** — a cópia do objeto acontece antes da tx, no padrão idêntico ao
     `spawnNextDraft`. Nunca chamada de rede dentro da tx; nunca outbox pós-commit (recriaria a
     janela sem conteúdo).
  2. **Copiar o asset de sistema, não gerar** — `Presigner.Copy` a partir de
     `system/templates/blank.docx`; proibido introduzir dependência de `templates` em `render`/
     docx-renderer ou um segundo gerador de DOCX vazio.
  3. **Chave canônica sempre** — destino via `templateDocxKey(tenantID, templateID, n)`;
     tenant-namespaced; nunca a chave do source.
  4. **Desfazer a sobrecarga de `content_hash`** — o ADR decide o significado único do campo antes de
     qualquer código; o aperto do CHECK vai em **migração forward 0317**, baseline intocado.
  5. **Sem capability nova, sem módulo novo, sem nova aresta entre módulos** — se o design exigir
     qualquer um dos três, o gate deve ser reaberto.
