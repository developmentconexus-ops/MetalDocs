# T11 — B01 App Shell + Global Information Architecture

> **Status:** LOCKED / OPERATOR-RATIFIED on 2026-08-22.  
> **Scope:** B01 only. This lock freezes the current structural planning baseline; it is not Product implementation, visual-design finalization, or a ban on later evidence-backed reopen.  
> **Rendered structural artifact:** `docs/work/current/t11-b01-app-shell-wireframe.html`

## 1. Boundary

B01 owns only the global application frame, top-level task/findability model, the default `/work` home composition, global navigation placement, and responsive transformation of that frame.

B01 does **not** lock Library collection structure, Document Official composition, Document Work/editor structure, Governance Case layout, History/Audit detail, Administration detail, final visual styling, or production component boundaries.

Fixed inherited invariants remain:

```text
stable Product SPA paths             10
application operations               78
operation 79                         ABSENT
frontend semantic owner              none
frontend Authorization engine        absent
Product implementation               BLOCKED
T12                                  NOT OPEN
```

## 2. Operator adjudication

The first rendered hypothesis — a sparse shell centered on four top-level links — was rejected because it understated the amount of work users must understand and made MetalDocs feel like a thin route launcher instead of a complete controlled-document workspace.

The revised **Home A — Central operacional** was explicitly approved by the operator as the current baseline, with the expectation that later evidence or stronger ideas may trigger the normal smallest-scope reopen process.

Decision:

```text
initial sparse sidebar hypothesis     REJECTED
Home A / operational home direction   LOCKED
B01                                    LOCKED
B02                                    NOT OPEN in this record
```

## 3. User mental model locked by B01

The global frame is organized around human work rather than backend nouns:

```text
INÍCIO
  what needs my attention now

MINHA CAIXA
  Para aprovação
  Em edição

DOCUMENTOS
  Biblioteca
  Novo documento
  Criar a partir de modelo

GESTÃO
  Governança documental
  Organização
  Acesso

EVIDÊNCIA
  Auditoria
```

This is presentation/IA. It does not create peer Product owners named Home, Approval, Templates, Inbox, or Management.

## 4. Reference evidence used for the revision

Reference evidence informed the hypothesis but did not become Product authority.

| Reference | Source observation | Product inference / decision |
|---|---|---|
| Qualio | Home exposes `My actions` / outstanding work and can lead directly into document review/approval. | MetalDocs should surface actor work on entry instead of making users hunt for it. |
| Veeva Vault | Home commonly exposes assigned tasks; Library remains a distinct document context. | Work queue and official-document discovery should remain distinct mental models. |
| M-Files | Home/quick views expose user-relevant work such as `Assigned to Me` and recently accessed contexts. | A home can be task-oriented without becoming lifecycle authority. |

The B01 candidate deliberately did **not** copy configurable workflow trees, generic dashboards, custom tab collections, favorites, recents, or process taxonomies because current MetalDocs authority does not admit them.

## 5. Locked global IA and route mapping

### Início

`Início` is **not** a new `/home` Product route. The default authenticated landing experience is the existing `/work` lens composed as an operational home.

It may read both accepted work projections:

```text
listGovernanceWork  → Para aprovação
listAuthoringWork   → Em edição
```

The home does not own lifecycle truth. Each item navigates to its current owner lens, which rechecks current server truth.

### Minha Caixa

```text
Para aprovação  → /work governance presentation
Em edição       → /work authoring presentation
```

Any lane/tab/query key is browser presentation state only and must not become `/api/v1` semantics.

### Documentos

```text
Biblioteca               → /documents
Novo documento           → existing creation entry within /documents
Criar a partir de modelo → existing creation entry using admitted eligible Template options
```

`Criar a partir de modelo` is a task entry, not an independent Template semantic owner or a newly admitted Template-browse route.

The Home search field is an entry to accepted Library search:

```text
input q
→ navigate /documents?q=<q>
→ listDocuments at destination
```

No global body/full-text/vector search is introduced.

### Gestão

```text
Governança documental → /admin/document-governance
Organização            → /admin/organization
Acesso                 → /admin/access
```

`Gestão` is visual grouping only; there is no `/admin` Product landing route implied by B01.

### Evidência

```text
Auditoria → /audit
```

Audit remains evidence, never current-state authority.

## 6. Locked structural wireframe decisions

Desktop baseline:

```text
utility header
+ persistent left global navigation
+ wide primary content region
```

The first fold of `/work` Home contains, in reading order:

```text
1. page identity + primary create entry
2. official-document search entry
3. "Precisa da sua atenção"
   - Para aprovação
   - Em edição
4. fixed quick actions
```

The following are deliberately absent from the locked baseline because current authority does not prove them:

```text
home KPI cards / totals
favorites
recently accessed documents
personalized saved shortcuts
configurable dashboard widgets
standalone Template browse page
```

Fixed shortcuts are allowed because they only navigate to already-admitted tasks. Persisted user-customizable shortcuts would require a new durable preference consumer and remain outside B01 authority.

## 7. Responsive and accessibility structure

Narrow-width transformation:

```text
persistent sidebar → drawer
Para aprovação / Em edição → stacked sections
quick actions → 4 columns → 2 → 1 as width requires
```

Semantic order remains:

```text
Início → Minha Caixa → Documentos → Gestão → Evidência
```

The primary workflow must remain keyboard reachable; no material destination depends on hover-only disclosure. Drawer open/close must preserve a plausible focus return path. Navigation grouping is conveyed by labels/structure, not color alone.

## 8. P9 — B01 Screen Contract

| Region / control | Goal | Accepted truth / operation | Identity / navigation source | Material failure intent | Forbidden frontend authority |
|---|---|---|---|---|---|
| Session shell | establish authenticated frame | `getSession` | HttpOnly session + returned SessionView | unauthenticated means sign-in state; dependency/internal stays sanitized | local session/role authority |
| Sign out | end current session | `endSession` | current session + CSRF | failure must not pretend logout succeeded | client-only logout truth |
| Home — Para aprovação | expose work waiting for actor governance action | `listGovernanceWork` | returned `attempt_id` | stale row is only a projection; destination rechecks current case | inbox row as participation/lifecycle authority |
| Home — Em edição | expose actor authoring work | `listAuthoringWork` | returned `document_id` | stale row routes through current Work resolver | row as current Revision authority |
| Home search | find official document | destination `listDocuments(q)` | entered query text becomes admitted Library URL state | normal Library denied/notfound/validation semantics | global/full-text search authority |
| Biblioteca navigation | reach official discovery | `/documents` + `listDocuments` | stable route | server owns disclosure | DRAFT-as-official inference |
| Novo documento | begin accepted create journey | existing Library creation surface | stable `/documents` destination; exact local presentation state remains B02-owned | destination handles current options/permissions | separate create owner/route |
| Criar a partir de modelo | begin create journey with eligible Template option | existing creation options/create flow | current eligible Template references are supplied by accepted creation options | absence means no usable option, not hidden global Template catalog | Template peer lifecycle/catalog invention |
| Admin destinations | reach accepted admin lenses | three stable `/admin/*` paths | stable routes | destination 403/404 remains authoritative | shell permission matrix |
| Audit destination | inspect evidence | `/audit` | stable route | destination 403/404 remains authoritative | Audit as current-state resolver |

## 9. Bidirectional trace

Product/backend → B01:

```text
getSession          → authenticated shell
endSession          → sign-out control
listGovernanceWork  → Home / Para aprovação
listAuthoringWork   → Home / Em edição
listDocuments       → Home search destination + Biblioteca
accepted create flow→ Novo documento / Criar a partir de modelo task entries
stable admin paths  → Gestão group
/audit              → Evidência group
```

B01 → Product/backend:

```text
shell session        → Authentication / getSession
sign out             → Authentication / endSession
approval item        → Controlled Documents work projection → Governance Case
editing item         → Controlled Documents work projection → Document Work
search               → Controlled Documents / listDocuments
create shortcuts     → accepted Library creation journey
management nav       → accepted Administration lenses
audit nav            → Audit lens
```

Reconciliation:

```text
invented application operations   0
new semantic owners               0
new stable Product paths          0
operation 79                      ABSENT
```

## 10. P10 bounded pattern consolidation

B01 provides evidence for local patterns only:

```text
global grouped navigation
operational-home attention section
fixed quick-action entry
responsive sidebar/drawer transformation
```

No shared Product pattern graduates yet because B01 is the only locked material block. Repetition must be observed across later locked blocks before semantic component/pattern consolidation.

## 11. P11 disposition

A bounded P8 HTML artifact is locked as structural evidence. Full B01 interaction realization is deferred until enough destination blocks are locked to make navigation behavior realistically testable without inventing their structures in advance.

This is not permission to pre-generate B02–B08.

## 12. Reopen law for B01

B01 may be reopened on concrete evidence such as:

```text
operator/direct-user evidence that findability fails
later locked block proves the global grouping obstructs a core journey
responsive/accessibility realization falsifies the structure
accepted Product authority changes a top-level user need or route meaning
new admitted preference capability justifies customizable shortcuts
```

Aesthetic preference alone can be handled at visual-design stage when it does not change navigation, reading order, grouping, density class, or workflow meaning.
