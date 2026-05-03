# Runbook: Smoke Checklist (Local / Preview)

> **Reference:** [`wiki/workflows/user-onboarding.md`](../../wiki/workflows/user-onboarding.md) — read first for context on each step.
> **Goal:** Validar o fluxo completo MetalDocs antes de refactors e releases. Executar como usuário final, não como dev.
> **Tempo esperado:** 15–25 min (full run).

## Pré-requisitos

- API rodando: `.\scripts\start-api.ps1` (porta 8081). Frontend buildado.
- Postgres + MinIO + Gotenberg up via Docker Compose.
- Login admin: `admin` / `AdminMetalDocs123!`.
- Banco com pelo menos um perfil + área **OU** ambiente limpo para validar bootstrap.
- Bootstrap triggered quando `metaldocs.iam_user_roles` não tem nenhuma row com `role_code = 'system_admin'` (anteriormente `admin` — ver migration 0166).

---

## Setup test users (one-time per environment)

Para cobrir segregação ISO precisa de **três contas distintas**:

- `admin` — bootstrap (taxonomia + template publish + profile binding).
- `author@test` — autor de templates e de documentos.
- `approver@test` — aprovador (não pode ser o submitter).

Criar via `Configurações → Usuários` (admin). Conferir que `approver@test` tem role `approver` (capability `doc.signoff`) na área alvo.

---

## Routine A — Bootstrap (admin)

> Mapeia para Steps 1 + 4 do `user-onboarding.md`.

### A1. Áreas
- [ ] Login como `admin`.
- [ ] **Tipos Documentais → Áreas → Nova Área**: code `RH`, name `Recursos Humanos`.
- [ ] Confirmar a área aparece na listagem.
- [ ] **Esperado:** sem toast de erro. Persistir após F5.

### A2. Perfis
- [ ] **Tipos Documentais → Perfis → Novo Perfil**: code `DC`, name `Descrição de Cargo`.
- [ ] Confirmar perfil listado.

### A3. (Postergado — voltar após Routine B)
> Bind perfil ao template publicado. Ver A4 abaixo.

---

## Routine B — Template authoring (author + admin)

> Mapeia para Steps 2 + 3.

### B1. Criar template
- [ ] Logout admin → login como `author@test`.
- [ ] **Templates → Novo Template**: title `DC — Descrição de Cargo`, perfil-alvo `DC`.
- [ ] Confirmar criação. Editor eigenpal abre. Estado `v1 draft`.

### B2. Editar conteúdo
- [ ] Inserir cabeçalho com `{doc_code}` literal (chip).
- [ ] Inserir bloco de título com `{doc_title}`.
- [ ] Inserir linha de assinatura com `{author}` + `{effective_date}`.
- [ ] Inserir bloco "Aprovado por: `{approvers}`".
- [ ] Body: 2–3 parágrafos lorem ipsum.
- [ ] **Save**. F5 → conteúdo persiste.
- [ ] **Esperado:** todos os 7 tokens reconhecidos. Sem chips em vermelho/erro.

### B3. Submit → Approve → Publish
- [ ] No painel de versão: **Enviar para revisão** → estado `in_review`.
- [ ] Logout `author@test` → login `admin` (ou aprovador de templates).
- [ ] **Templates → DC → v1 → Aprovar**. Estado `approved`.
- [ ] **Publicar**. Estado `published`.
- [ ] **Esperado:** botão Publicar habilitado só após approve. Lista de templates mostra `v1 published`.

### B4. Tentar editar versão publicada (negative test)
- [ ] Tentar editar conteúdo de `v1 published`.
- [ ] **Esperado:** editor read-only ou botão Save desabilitado. Não pode mutar versão publicada.

---

## Routine A4 — Bind template ao perfil (CRÍTICO)

> Mapeia para Step 4. **Sem isto, Routine C falha.**

- [ ] Logado como `admin`.
- [ ] **Tipos Documentais → Perfis → DC → Editar**.
- [ ] **Template padrão**: selecionar `DC — Descrição de Cargo (v1)`.
- [ ] Salvar.
- [ ] Reabrir o perfil → confirmar binding persistido.

---

## Routine C — Document creation (author)

> Mapeia para Steps 5 + 6 + 7.

### C1. Registrar Documento Controlado
- [ ] Login `author@test`.
- [ ] **Documentos Controlados → Novo Documento Controlado**.
- [ ] Perfil: `DC`. Área: `RH`. Título: `Descrição de Cargo — Analista Fiscal`.
- [ ] Confirmar.
- [ ] **Esperado:** novo CD com código `DC-RH-001` (ou próximo da sequência).

### C2. Gerar versão de trabalho
- [ ] Abrir o CD recém-criado → **Gerar Documento**.
- [ ] Wizard mostra template já selecionado (binding do A4).
- [ ] **Gerar**. Editor abre.
- [ ] **Esperado:** layout idêntico ao template. Tokens visíveis como chips, **não** resolvidos ainda.

### C3. Preencher conteúdo
- [ ] Adicionar 2–3 parágrafos descrevendo o cargo.
- [ ] Inserir tabela 2×3 com responsabilidades (validar render de tabelas).
- [ ] **Não tocar** nos chips dos 7 tokens fixos.
- [ ] **Save**. F5 → persiste.

### C4. Importar DOCX (opcional, gate sanity)
- [ ] Em outro CD de teste, gerar documento e importar `.docx` externo (ex.: `DC_Template_Descricao_Cargo.docx` se disponível).
- [ ] **Esperado:** import sem perda crítica de formatação. (Bug eigenpal de tabela em header está parked — anotar se reaparecer.)

---

## Routine D — Approval (author + approver)

> Mapeia para Step 8. Valida segregação ISO.

### D1. Submeter
- [ ] No editor → **Finalizar**.
- [ ] **Esperado:** estado `under_review`. Toast de sucesso.

### D2. ISO segregation (negative test)
- [ ] Ainda como `author@test`, ir em **Caixa de Entrada de Aprovação**.
- [ ] **Esperado:** o documento submetido **não aparece** para aprovação por ele mesmo. Se aparecer, é bug crítico.

### D3. Aprovar
- [ ] Logout → login `approver@test`.
- [ ] **Caixa de Entrada de Aprovação** → documento `DC-RH-001` listado.
- [ ] Abrir → revisar conteúdo → **Aprovar**.
- [ ] Modal pede senha → digitar senha de `approver@test` → confirmar.
- [ ] **Esperado:** signoff registrado. Se rota é `any_1`, estado vai a `approved` imediatamente. Se `m_of_n`, repetir com outro aprovador.

### D4. Rejeição (caminho alternativo, em outro CD)
- [ ] Repetir D1–D3 num CD separado, mas clicar **Rejeitar** com motivo.
- [ ] **Esperado:** estado vai a `rejected` (não volta a `draft` — author deve criar nova revisão). Author recebe notificação.

---

## Routine E — Freeze + Fanout (automático, mas observável)

> Mapeia para Step 9.

### E1. Freeze automático
- [ ] Logo após o último signoff em D3, abrir o detalhe do documento.
- [ ] **Esperado:** estado `approved` (a nomenclatura `frozen` foi removida — o estado `approved` é o estado pós-signoff imutável). Hashes `content_hash` visíveis, `revision_version` incrementado.
- [ ] Tentar editar → **Esperado:** editor read-only (sem botão Submeter/Editar disponível).

### E2. Token resolution
- [ ] Baixar o DOCX frozen → abrir no Word/LibreOffice.
- [ ] **Esperado:** todos os 7 tokens substituídos por valores reais (`DC-RH-001`, título real, autor, data, aprovadores, área, revisão `1`).

### E3. Fanout (PDF)
- [ ] Aguardar até 2 min. F5 na página do documento.
- [ ] **Esperado:** link de download de PDF disponível.
- [ ] Abrir PDF → tokens resolvidos, layout fiel ao DOCX.
- [ ] Se >2 min sem PDF, checar logs do worker (escalar para dev).

### E4. Idempotency
- [ ] Tentar reaprovar/refreezar via API ou re-clique → **Esperado:** no-op silencioso, sem corromper estado.
- [ ] **Nota (bug conhecido):** replay do mesmo `Idempotency-Key` depois que a instância já foi aprovada retorna HTTP 500 em vez de `was_replay: true`. Estado não é corrompido, mas a resposta de erro é incorreta. Ver issue de acompanhamento.

---

## Routine F — Revisão (segunda iteração)

> Smoke do fluxo de revisão. Reaproveitar o CD do Routine C.

- [ ] No CD `DC-RH-001` (já frozen) → **Nova Revisão**.
- [ ] **Esperado:** nova versão `v2 draft` clonada da v1 frozen.
- [ ] Editar diferença (1 parágrafo).
- [ ] Submit → Approve → Freeze → Fanout (Routines D + E).
- [ ] **Esperado:** `{revision_number}` agora resolve para `2`. Versão anterior continua acessível no histórico.

---

## Critérios de pass/fail

**Pass:** Todas as Routines A–E concluem sem erros. Checkboxes negative-test (B4, D2) confirmam bloqueio. PDF gerado com tokens resolvidos.

**Fail (qualquer um destes = bloqueia release):**
- Wizard do C2 não oferece template (bug de binding A4 ou publish).
- Author consegue aprovar próprio documento (D2 falha).
- Token não resolve no DOCX/PDF frozen (E2/E3).
- Fanout não produz PDF em <5 min.
- Editor permite edição de versão `frozen` ou `published`.

---

## Comandos rápidos

```powershell
# Start API
.\scripts\start-api.ps1

# Rebuild + start
.\scripts\start-api.ps1 -Build

# Frontend
cd frontend\apps\web; npm.cmd run build
cd frontend\apps\web; npm.cmd run dev

# Compose stack
docker compose up -d
docker compose logs -f api
docker compose logs -f docgen-v2
```

## Login API quick check
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"admin","password":"AdminMetalDocs123!"}'
```
