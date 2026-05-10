# Phase 1 Map — documento-publicado

**Completed:** 2026-05-08  
**No open questions at checkpoint.**

---

## 1.1 Reusability scan — backward

| Design element | Existing primitive | Status | Notes |
|---|---|---|---|
| Status badge ("vigente") | `components/ui/StatusPill` | ✅ Direct use | `status='published'` → "Publicado"; design said "vigente" but "Publicado" is the correct system term |
| Code badge (PR-EHS-014) | `components/ui/CodeChip` | ✅ Direct use | |
| Icon (copy, view, arrow, check) | `components/ui/Icon` | ⚠️ Partial | `link` ✅, `arrow` ✅, `check` ✅, `docs` ✅; `eye` missing from `IconName` — use `docs` for "Visualizar" or add `eye` |
| Avatar (owner initials) | `components/ui/Avatar` | ✅ Direct use | |
| Signoff pipeline stepper | `components/ui/Stepper` | ⚠️ Insufficient | Stepper is single-level; SignoffPipeline needs stages + signoffs per stage. New domain component required. |
| Breadcrumb | None | ❌ Missing | Build inline in hero; not generic enough yet for `components/ui/` |

---

## 1.2 Reusability scan — forward (new components)

| New component | Placement | Reason |
|---|---|---|
| `DocumentPublishedPage` | `features/documents/pages/` | Route-level page |
| `DocumentPublishedHero` | `features/documents/components/` | Domain-specific hero (breadcrumb + card + badges + actions) |
| `DocCardMini` | `features/documents/components/` | Decorative 3D card; documents-only |
| `DocumentPublishedSidebar` | `features/documents/components/` | Sidebar shell (AboutCard + SignoffPipeline layout) |
| `DocumentSignoffPipeline` | `features/documents/components/` | Approval stages + signoffs display; domain-specific |
| `DocumentObsoleteBanner` | `features/documents/components/` | Full-width obsolete stamp; conditional on status |

No new generic primitives — all new components are domain-specific to `features/documents/`.

---

## 1.3 Decomposition tree

```
DocumentPublishedPage
├── [conditional] DocumentObsoleteBanner (status === 'obsolete')
├── DocumentPublishedHero
│   ├── <nav> Breadcrumb (Biblioteca → Área → Tipo → Código)
│   ├── DocCardMini (decorative 3D card — CSS only)
│   ├── CodeChip (controlled_document.code)
│   ├── StatusPill (document.status)
│   ├── <span> Version badge (document.revision_version)
│   ├── <span> Type label (profile label from profiles query)
│   ├── <h1> document.name
│   └── <div> Action buttons
│       ├── "Visualizar documento" → navigate(`/documents-v2/${id}`)
│       ├── "Iniciar revisão" → POST /api/v2/controlled-documents/:cdId/revisions (RBAC-gated, disabled if not author)
│       └── "Copiar link" → navigator.clipboard.writeText(href)
├── <section> KPI strip
│   └── Versão atual: document.revision_version
└── <aside> DocumentPublishedSidebar
    ├── <section> AboutCard
    │   ├── Avatar (document.created_by)
    │   ├── owner name (document.created_by)
    │   ├── created_at date
    │   ├── Fact: Tipo (profile label)
    │   └── Fact: Área (area label)
    └── DocumentSignoffPipeline
        └── For each stage in approval_instance.stages:
            ├── Stage header (stage name + status)
            └── For each signoff in stage.signoffs:
                ├── Avatar (actor display name)
                ├── actor name
                └── signoff status icon
```

---

## 1.4 Status / enum meta SSOT

**Reuse existing:** `features/documents/lib/libraryStatus.ts` has `LIBRARY_STATUSES` + `DocumentStatus` type from `StatusPill`.

**New file needed:** `features/documents/lib/documentDetailMeta.ts` — for:
- Profile code snapshot → human label lookup util: `resolveProfileLabel(code, profiles[])`
- Area code snapshot → human label lookup util: `resolveAreaLabel(code, areas[])`
- Signoff status → icon + color: `SIGNOFF_STATUS_META` record

No other meta files. No inline string duplications in components.

---

## 1.5 State design

| State | Type | Hook / location |
|---|---|---|
| Document detail | Server | `useDocumentDetailQuery(id)` — new query hook wrapping `getDocument` |
| Approval instance | Server | `useApprovalInstanceQuery(id)` — new query hook |
| Profiles (for label) | Server | `useProfilesQuery()` — existing |
| Areas (for label) | Server | `useAreasQuery()` — existing |
| Copy link feedback | Local | `useState<boolean>(false)` in page — resets after 2s via `setTimeout` |
| "Iniciar revisão" loading | Local | Derived from `isPending` of mutation — no extra state |

No localStorage state. No debounced inputs. No persisted state.

---

## 1.6 Backend contract

### Existing — use as-is
| Endpoint | Frontend function | Query key |
|---|---|---|
| `GET /api/v2/documents/:id` | `documentsV2.ts#getDocument` | `QK.documents.detail(id)` |
| `GET /api/v2/profiles` | `taxonomy.ts#fetchProfiles` via `useProfilesQuery` | `QK.taxonomy.profiles()` |
| `GET /api/v2/areas` | `taxonomy.ts#fetchAreas` via `useAreasQuery` | `QK.taxonomy.areas()` |

### New — add to documentsV2.ts
| Endpoint | New function | Query key |
|---|---|---|
| `GET /api/v2/documents/:id/approval-instance` | `getApprovalInstance(id)` | `QK.approval.instance(id)` |

### New — mutation
| Endpoint | New function | Notes |
|---|---|---|
| `POST /api/v2/controlled-documents/:cdId/revisions` | `createRevision(cdId)` in new `controlledDocumentsApi.ts` or existing registry api | RBAC-gated; disabled when user is not author |

### Approval instance response types (to add to documentsV2.ts or new approvalApi.ts)
```ts
type SignoffRecord = {
  actor_user_id: string;    // actually display name snapshot (confirmed in handler)
  status: 'pending' | 'approved' | 'rejected' | 'abstained';
  signed_at: string | null;
  comment: string | null;
};

type StageInstance = {
  stage_id: string;
  name: string;
  status: 'pending' | 'approved' | 'rejected' | 'skipped';
  signoffs: SignoffRecord[];
};

type ApprovalInstanceResponse = {
  id: string;
  document_id: string;
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  stages: StageInstance[];
  created_at: string;
  updated_at: string;
};
```

### Mock fallback strategy
- If `getApprovalInstance` returns 404 (no approval instance for this doc), render nothing for SignoffPipeline (show null/empty state, not an error).
- Document detail 404 → error state with retry.

### Deferred to backlog
- PDF download endpoint
- Fanout coverage API
- Revision list endpoint (VersionTimeline)
- Comments architecture (storage brainstorm)
- `values_hash` in document response (AuditCard)

---

## 1.7 Route registration

Add to `features/documents/routes.tsx`:
```ts
{
  path: 'documents/:documentId',
  handle: { workspaceView: 'library' },
  lazy: () => import('./pages/DocumentPublishedPage').then(m => ({ Component: m.DocumentPublishedPage })),
},
```

**Placement:** After `documents/all`, `documents/area/:areaCode` etc. (React Router v6 prefers static over dynamic — no conflict risk with existing routes.)

---

## Open questions at checkpoint

None. Phase 1 complete.
