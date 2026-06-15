# Feature F4b.1 — evidence (verified-dead manifest)

> **Outcome:** Cluster **verified dead**. HS-2 not tripped. Manifest below authorizes F4b.2's drop.

## Verified-dead manifest (10 objects, drop in this order — satellites before anchor)

| # | Object (schema-qualified) | Role in cluster | Runtime Go refs |
|---|---------------------------|-----------------|-----------------|
| 1 | `metaldocs.document_version_images` | 2nd-level satellite — FK → `document_versions_mddm`; leaf (nothing references it) | 0 |
| 2 | `metaldocs.document_attachments` | satellite — FK → `documents` | 0 |
| 3 | `metaldocs.document_collaboration_presence` | satellite — FK → `documents` | 0 |
| 4 | `metaldocs.document_edit_locks` | satellite — FK → `documents` | 0 |
| 5 | `metaldocs.document_template_assignments` | satellite — FK → `documents` | 0 |
| 6 | `metaldocs.document_versions` | satellite — FK → `documents` (holds dead `canonical_mddm_snapshot` col) | 0 |
| 7 | `metaldocs.document_versions_mddm` | satellite — FK → `documents` | 0 |
| 8 | `metaldocs.workflow_approvals` | satellite — FK → `documents` | 0 |
| 9 | `metaldocs.documents` | **anchor** of the legacy MDDM document model | 0 (only a comment at `internal/modules/search/infrastructure/v2documents/reader.go:25`) |
| 10 | `metaldocs.template_audit_log` | independent dead duplicate of `public.template_audit_log` | 0 |

Indexes, FK constraints, and any sequences attached to these tables drop with them (CASCADE-safe — see C).

## Gate proof

**A — manifest complete (inbound-FK closure).** Sweep of `REFERENCES metaldocs.<cluster>` over the
curated baseline returns exactly these edges, all originating **inside** the manifest:
- `document_attachments`, `document_collaboration_presence`, `document_edit_locks`,
  `document_template_assignments`, `document_versions`, `document_versions_mddm`, `workflow_approvals`
  → `metaldocs.documents(id)`
- `document_version_images` → `metaldocs.document_versions_mddm(id)`  ← **2nd-level satellite, caught
  during census** (operator's stop flagged MDDM; this child was added to the set).
No `REFERENCES metaldocs.document_version_images` exists → it is a leaf. Closure holds.

**B — zero runtime Go refs.** Per-table `grep` over `internal/`,`apps/` (non-test, non-vendor) = 0 for
all 10. (The only `metaldocs.documents` token in runtime is a comment calling the schema
"decommissioned".)

**C — no inbound FK from a kept table.** Every inbound FK in (A) originates from a manifest object.
Therefore drop ordering 1→10 (or a single `DROP TABLE ... CASCADE`) removes the cluster with no impact
on any kept table.

**D — no other dependent.** Targeted sweeps over baseline + migrations + reference-data:
- Views referencing cluster: **none** (only unrelated `metaldocs.user_process_areas` view exists).
- Triggers on cluster tables: **none**.
- RLS policies (incl. `0237_rls_all_tenant_tables.sql`) on cluster tables: **none** (the legacy tables
  lack `tenant_id`, so the tenant-RLS pass never touched them).
- Stored-function bodies reading/writing cluster tables: **none**.
- Sequences `OWNED BY` cluster tables: **none**.
- Reference-data seeds writing cluster tables: **none**.

**E — HS-2 not tripped.** A–D all clear.

## MDDM disambiguation (operator sanity check)

MDDM = the original *MetalDocs Document Model* (early editor era), superseded by the current
`public.documents` + `public.controlled_documents` governance model.
- **Legacy MDDM tables** (`metaldocs.document_versions_mddm`, `document_version_images`, and the whole
  `metaldocs.documents` cluster) → **dead** (0 runtime refs; this manifest).
- **MDDM export format** → **alive**, but only as a feature flag: all 8 non-test Go `mddm` refs are
  `MDDMNativeExportRolloutPercent` / `MDDM_NATIVE_EXPORT_ROLLOUT_PCT` (client-side DOCX export rollout)
  in `internal/platform/config/feature_flags.go` + `internal/platform/featureflags/handler.go`. **None
  reference the legacy tables.** The dead `canonical_mddm_snapshot` column rides on the dead
  `metaldocs.document_versions` table and drops with it.

Conclusion: dropping the cluster is safe and correct.

## Migration-ordering note (for F4b.2)

Migrations `0236` (CASCADE-dropped `fk_documents_subject_code`) and `0238` (dropped
`metaldocs.documents.subject_code`) still run **before** 4b's new migration and operate on the
still-present `metaldocs.documents` — no conflict; 4b's drop is the final state. New migration number
must be `> 0239`.
