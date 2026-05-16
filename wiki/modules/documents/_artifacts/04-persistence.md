# 04 — Persistence

## 1. Tables Owned
| Table | Created in | Notes |
|---|---|---|
| public.documents | 0110_docx_v2_documents.sql:14 | Replaced `documents_v2` workflow; later bridge/state fixes in 0167 and checks in 0183. |
| public.editor_sessions | 0110_docx_v2_documents.sql:34 | Session table used by repository methods; checklist name `public.document_sessions` not found (unclear: no migration defines that name). |
| public.document_revisions | 0110_docx_v2_documents.sql:49 | Revision ledger with unique `(document_id, content_hash)`. |
| public.document_checkpoints | 0110_docx_v2_documents.sql:90 | Checkpoint labels by `(document_id, version_num)`. |
| public.document_placeholder_values | 0152_placeholder_fillin_columns.sql:51 | Fill-in values keyed by `(tenant_id, revision_id, placeholder_id)`. |
| public.document_exports | 0111_docx_v2_exports.sql:3 | Export ledger keyed by composite hash. |
| public.document_comments | 0118_docx_v2_document_comments.sql:1 | Comment thread rows per document. |
| public.approval_routes | 0134_approval_routes.sql:3 | Approval route header table. |
| public.approval_route_stages | 0134_approval_routes.sql:14 | Stage definitions for routes. |
| public.approval_instances | 0135_approval_instances.sql:9 | Runtime approval instances. |
| public.approval_stage_instances | 0135_approval_instances.sql:40 | Runtime stage instances. |
| public.approval_signoffs | 0135_approval_instances.sql:72 | Immutable signoff rows (trigger-enforced). |
| public.governance_events | 0125_registry_iam_user_process_areas_governance_events.sql:24 | Governance audit/event log. |
| metaldocs.pdf_dispatch_outbox | 0176_pdf_dispatch_outbox.sql:2 | Exists in `metaldocs` schema; checklist `public.pdf_outbox` not found. |

```sql
-- public.documents (migrations/0110_docx_v2_documents.sql:14)
id UUID NOT NULL PRIMARY KEY
tenant_id UUID NOT NULL
template_version_id UUID NOT NULL REFERENCES template_versions(id)
name TEXT NOT NULL
status TEXT NOT NULL
current_revision_id UUID REFERENCES document_revisions(id)
active_session_id UUID REFERENCES editor_sessions(id)
created_by UUID NOT NULL
```

```sql
-- public.editor_sessions (migrations/0110_docx_v2_documents.sql:34)
id UUID NOT NULL PRIMARY KEY
document_id UUID NOT NULL REFERENCES documents(id)
user_id UUID NOT NULL
last_acknowledged_revision_id UUID NOT NULL
status TEXT NOT NULL
```

```sql
-- public.document_revisions (migrations/0110_docx_v2_documents.sql:49)
id UUID NOT NULL PRIMARY KEY
document_id UUID NOT NULL REFERENCES documents(id)
parent_revision_id UUID REFERENCES document_revisions(id)
session_id UUID NOT NULL REFERENCES editor_sessions(id)
storage_key TEXT NOT NULL
content_hash TEXT NOT NULL
```

```sql
-- public.document_checkpoints (migrations/0110_docx_v2_documents.sql:90)
id UUID NOT NULL PRIMARY KEY
document_id UUID NOT NULL REFERENCES documents(id)
revision_id UUID NOT NULL REFERENCES document_revisions(id)
version_num INT NOT NULL
created_by UUID NOT NULL
```

```sql
-- public.document_placeholder_values (migrations/0152_placeholder_fillin_columns.sql:51)
tenant_id UUID NOT NULL
revision_id UUID NOT NULL REFERENCES documents(id)
placeholder_id TEXT NOT NULL
source TEXT NOT NULL
PRIMARY KEY (tenant_id, revision_id, placeholder_id)
```

```sql
-- public.document_exports (migrations/0111_docx_v2_exports.sql:3)
id UUID NOT NULL PRIMARY KEY
document_id UUID NOT NULL REFERENCES documents(id)
revision_id UUID NOT NULL
composite_hash BYTEA NOT NULL
storage_key TEXT NOT NULL
size_bytes BIGINT NOT NULL
```

```sql
-- public.document_comments (migrations/0118_docx_v2_document_comments.sql:1)
id UUID NOT NULL PRIMARY KEY
tenant_id UUID NOT NULL
document_id UUID NOT NULL REFERENCES documents(id)
library_comment_id INTEGER NOT NULL
author_id TEXT NOT NULL
content_json JSONB NOT NULL
```

```sql
-- public.approval_routes (migrations/0134_approval_routes.sql:3)
id UUID NOT NULL PRIMARY KEY
tenant_id UUID NOT NULL
profile_code TEXT NOT NULL
name TEXT NOT NULL
created_by TEXT NOT NULL
```

```sql
-- public.approval_route_stages (migrations/0134_approval_routes.sql:14)
id UUID NOT NULL PRIMARY KEY
route_id UUID NOT NULL REFERENCES public.approval_routes(id)
stage_order INT NOT NULL
required_role TEXT NOT NULL
required_capability TEXT NOT NULL
area_code TEXT NOT NULL
```

```sql
-- public.approval_instances (migrations/0135_approval_instances.sql:9)
id UUID NOT NULL PRIMARY KEY
tenant_id UUID NOT NULL
document_id UUID NOT NULL REFERENCES documents(id)
route_id UUID NOT NULL REFERENCES approval_routes(id)
submitted_by TEXT NOT NULL
```

```sql
-- public.approval_stage_instances (migrations/0135_approval_instances.sql:40)
id UUID NOT NULL PRIMARY KEY
approval_instance_id UUID NOT NULL REFERENCES approval_instances(id)
stage_order INT NOT NULL
required_role_snapshot TEXT NOT NULL
required_capability_snapshot TEXT NOT NULL
```

```sql
-- public.approval_signoffs (migrations/0135_approval_instances.sql:72)
id UUID NOT NULL PRIMARY KEY
approval_instance_id UUID NOT NULL REFERENCES approval_instances(id)
stage_instance_id UUID NOT NULL REFERENCES approval_stage_instances(id)
actor_user_id TEXT NOT NULL
actor_tenant_id UUID NOT NULL
```

```sql
-- public.governance_events (migrations/0125_registry_iam_user_process_areas_governance_events.sql:24)
id UUID NOT NULL PRIMARY KEY
tenant_id UUID NOT NULL
event_type TEXT NOT NULL
actor_user_id TEXT NOT NULL
resource_type TEXT NOT NULL
resource_id TEXT NOT NULL
payload_json JSONB NOT NULL
```

```sql
-- metaldocs.pdf_dispatch_outbox (migrations/0176_pdf_dispatch_outbox.sql:2)
id UUID NOT NULL PRIMARY KEY
tenant_id UUID NOT NULL
revision_id UUID NOT NULL
content_hash BYTEA NOT NULL
status TEXT NOT NULL
```

## 2. Tables Read/Written but Not Owned
| Table | Owner module | Read/Write | Operations |
|---|---|---|---|
| metaldocs.iam_users | IAM | Read | `SELECT display_name` in create/signoff snapshot paths (`repository.go:89`, `0173_signoff_actor_displayname_snapshot.sql:9`). |
| metaldocs.document_process_areas | Taxonomy | Read | `SELECT name` in document snapshot path (`repository.go:95`). |
| template_versions | Templates | Read via FK | FK on `documents.template_version_id` (`0110_docx_v2_documents.sql:17`). |
| metaldocs.document_profiles | Taxonomy | Read via FK | FK on `approval_routes(tenant_id, profile_code)` (`0134_approval_routes.sql:57-58`). |

## 3. Triggers, GUCs, Functions
| Object | Kind | File:line | Purpose |
|---|---|---|---|
| `public.enforce_capability_asserted` | Function (`SECURITY DEFINER`) | `migrations/0142b_role_capabilities_v2_enforce.sql:67` | Checks `metaldocs.asserted_caps` for INSERT into approval tables. |
| `trg_require_cap_asserted_instances` | Trigger | `migrations/0142b_role_capabilities_v2_enforce.sql:201` | BEFORE INSERT on `public.approval_instances`. |
| `trg_require_cap_asserted_signoffs` | Trigger | `migrations/0142b_role_capabilities_v2_enforce.sql:207` | BEFORE INSERT on `public.approval_signoffs`. |
| `metaldocs.asserted_caps` | GUC read | `migrations/0142b_role_capabilities_v2_enforce.sql:139` | Capability assertions source. |
| `metaldocs.actor_id` | GUC read | `migrations/0137_db_roles_security_definer.sql:68` | Actor identity inside SECURITY DEFINER membership functions. |
| `metaldocs.tenant_id` | GUC | (unclear: no read/write found in scanned migrations) | No direct match found by search. |
| `enforce_snapshot_on_submit` | Function | `migrations/0152_placeholder_fillin_columns.sql:29` | Blocks state transitions when snapshot columns are NULL. |
| `enforce_snapshot_on_submit_trg` | Trigger | `migrations/0152_placeholder_fillin_columns.sql:47` | BEFORE INSERT/UPDATE on `documents`. |
| `enforce_signoff_tenant_consistent` + trigger | Function + Trigger | `migrations/0135_approval_instances.sql:104`, `:125` | Rejects cross-tenant signoffs. |
| `enforce_signoff_sod` + trigger | Function + Trigger | `migrations/0135_approval_instances.sql:131`, `:151` | Blocks author self-signoff. |

## 4. Indexes
| Index | Table | Columns | Unique? | Purpose |
|---|---|---|---|---|
| `idx_documents_tenant_status` | `documents` | `(tenant_id, status)` | No | list/filter by tenant+status (`0110:30`). |
| `idx_documents_template_version` | `documents` | `(template_version_id)` | No | template-version lookup (`0110:31`). |
| `idx_documents_form_data_gin` | `documents` | `form_data_json` (GIN) | No | JSONB path queries (`0110:32`). |
| `idx_one_active_session_per_doc` | `editor_sessions` | `(document_id) WHERE status='active'` | Yes | single active writer (`0110:46`). |
| `idx_revisions_doc_num` | `document_revisions` | `(document_id, revision_num DESC)` | No | revision history scans (`0110:62`). |
| `document_exports_doc_hash_uidx` | `document_exports` | `(document_id, composite_hash)` | Yes | idempotent export insert (`0111:16`). |
| `idx_document_comments_doc_lib` | `document_comments` | `(document_id, library_comment_id)` | Yes | unique library comment id per doc (`0118:16`). |
| `approval_routes_tenant_profile_key` | `approval_routes` | `(tenant_id, profile_code)` | Yes | one route per tenant/profile (`0134:11`). |
| `ux_approval_instances_active` | `approval_instances` | `(document_id) WHERE status='in_progress'` | Yes | one in-progress instance per doc (`0135:33`). |
| `ix_stage_instances_active` | `approval_stage_instances` | `(approval_instance_id, stage_order) WHERE status='active'` | No | active stage lookup (`0135:65`). |
| `ix_signoffs_stage` | `approval_signoffs` | `(stage_instance_id)` | No | stage signoff lookup (`0135:100`). |
| `ix_governance_events_tenant_type` | `governance_events` | `(tenant_id, event_type, created_at DESC)` | No | governance feed lookup (`0125:36`). |
| `ix_pdf_dispatch_outbox_pending` | `metaldocs.pdf_dispatch_outbox` | `(next_retry_at) WHERE status IN (...)` | No | pending dispatch polling (`0176:18`). |

## 5. Tripwire Pairing Audit
| Method (file:line) | Authz.Require called? | Cap + area arg | SQL verb | Table |
|---|---|---|---|---|
| `InsertInstance` (`postgres_approval_repository.go:32`) | No | (unclear: no `authz.Require` call in file) | INSERT | `approval_instances` |
| `InsertSignoff` (`postgres_approval_repository.go:114`) | No | (unclear: no `authz.Require` call in file) | INSERT | `approval_signoffs` |
| `UpdateInstanceStatus` (`postgres_approval_repository.go:469`) | No | (unclear: no `authz.Require` call in file) | UPDATE | `approval_instances` |
| `CreateDocumentTx` (`repository.go:76`) | Yes | `document.create` before INSERT; `document.edit` before pointer/snapshot UPDATEs; area=`tenant` | INSERT/UPDATE | `documents` |
| `UpdateDocumentName` (`repository.go:216`) | No | (unclear: no `authz.Require` call in file) | UPDATE | `documents` |
| `UpdateDocumentStatus` (`repository.go:428`) | No | (unclear: no `authz.Require` call in file) | UPDATE | `documents` |
| `MarkArchived` (`repository.go:1071`) | No | (unclear: no `authz.Require` call in file) | UPDATE | `public.documents` |
| `Unarchive` (`repository.go:1082`) | No | (unclear: no `authz.Require` call in file) | UPDATE | `public.documents` |

Tripwire rule evaluation facts:
- `approval_instances` / `approval_signoffs` writes above have no `authz.Require` call in the scanned repository files; DB tripwire trigger exists in `0142b_role_capabilities_v2_enforce.sql:201-209`.
- `CreateDocumentTx` now asserts `document.create` and `document.edit` in order inside the caller-owned transaction; the remaining `documents` rows in this audit are pre-existing scanned findings outside this sync scope.
- Reads were not audited per rule.

## 6. Migration History
| Order | Filename | Verb summary | Date |
|---|---|---|---|
| 1 | 0001_init_documents.sql | CREATE schema/tables/indexes (`metaldocs.documents`, `metaldocs.document_versions`) | (unclear: filename has no date) |
| 2 | 0103_docx_v2_documents.sql | CREATE `documents_v2` + indexes | (unclear: filename has no date) |
| 3 | 0104_docx_v2_editor_sessions.sql | CREATE `editor_sessions` (W1) + indexes | (unclear: filename has no date) |
| 4 | 0105_docx_v2_document_revisions.sql | CREATE `document_revisions` (W1) + FK wiring | (unclear: filename has no date) |
| 5 | 0107_docx_v2_document_checkpoints.sql | CREATE `document_checkpoints` (W1) + index | (unclear: filename has no date) |
| 6 | 0110_docx_v2_documents.sql | DROP W1 tables; CREATE `documents`, `editor_sessions`, `document_revisions`, `autosave_pending_uploads`, `document_checkpoints` | (unclear: filename has no date) |
| 7 | 0111_docx_v2_exports.sql | CREATE `document_exports` + unique/index | (unclear: filename has no date) |
| 8 | 0118_docx_v2_document_comments.sql | CREATE `document_comments` + indexes | (unclear: filename has no date) |
| 9 | 0125_registry_iam_user_process_areas_governance_events.sql | CREATE `user_process_areas`, `governance_events` + indexes | (unclear: filename has no date) |
| 10 | 0126_documents_v2_bridge_columns.sql | ALTER `documents_v2` add bridge cols + index | (unclear: filename has no date) |
| 11 | 0129_documents_v2_bridge_not_null.sql | ALTER `documents_v2` tighten nullability (bridge columns) | (unclear: filename has no date) |
| 12 | 0131_documents_v2_state_columns.sql | ALTER `documents` state columns + unique indexes | (unclear: filename has no date) |
| 13 | 0134_approval_routes.sql | CREATE `approval_routes`, `approval_route_stages` + profile FK wiring | (unclear: filename has no date) |
| 14 | 0135_approval_instances.sql | CREATE approval runtime tables/triggers/indexes | (unclear: filename has no date) |
| 15 | 0139_governance_events_caps_bump.sql | CREATE unique spec-version index on governance events payload | (unclear: filename has no date) |
| 16 | 0141_governance_events_dedupe_signoff_uniqueness.sql | governance event dedupe/signoff uniqueness changes | (unclear: filename has no date) |
| 17 | 0142b_role_capabilities_v2_enforce.sql | CREATE capability tripwire function/triggers (`SECURITY DEFINER`) | (unclear: filename has no date) |
| 18 | 0142b_down.sql | DROP capability tripwire function/triggers (down migration) | (unclear: filename has no date) |
| 19 | 0146_approval_routes_active_column.sql | ALTER `approval_routes` add `active` | (unclear: filename has no date) |
| 20 | 0151_seed_dev_tenant_approval_data.sql | ALTER role constraint + seed capability/dev approval data | (unclear: filename has no date) |
| 21 | 0152_placeholder_fillin_columns.sql | ALTER `documents`; CREATE fill-in tables; CREATE snapshot trigger | (unclear: filename has no date) |
| 22 | 0153_placeholder_values_tenant_consistency.sql | CREATE tenant consistency trigger(s) | (unclear: filename has no date) |
| 23 | 0164_documents_v2_visibility.sql | ALTER `documents_v2` add visibility | (unclear: filename has no date) |
| 24 | 0167_documents_bridge_and_state_columns.sql | ALTER `documents` bridge/state repair + indexes | (unclear: filename has no date) |
| 25 | 0168_drop_documents_v2_orphan.sql | DROP `public.documents_v2` orphan | (unclear: filename has no date) |
| 26 | 0174_documents_created_by_displayname_snapshot.sql | ALTER/UPDATE `public.documents` snapshot field | (unclear: filename has no date) |
| 27 | 0175_documents_area_name_snapshot.sql | ALTER/UPDATE `public.documents` snapshot field | (unclear: filename has no date) |
| 28 | 0176_pdf_dispatch_outbox.sql | CREATE `metaldocs.pdf_dispatch_outbox` + index | (unclear: filename has no date) |
| 29 | 0181_drop_documents_locked_at.sql | ALTER `public.documents` drop `locked_at` | (unclear: filename has no date) |
| 30 | 0182_cd_sequence_per_area.sql | DROP old counter; CREATE `cd_sequence_counters` | (unclear: filename has no date) |
| 31 | 0183_documents_name_not_empty.sql | UPDATE + ALTER `documents` NOT NULL/CHECK name | (unclear: filename has no date) |
