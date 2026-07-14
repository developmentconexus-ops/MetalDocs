# Approval Module Persistence Map (Phase 4)

## Migration discovery
- Requested glob root `db/migrations` for IDs `0016, 0134, 0135, 0140, 0141, 0142a, 0142b, 0144, 0145, 0146, 0151, 0160, 0167, 0173, 0180`: NOT FOUND (`db/migrations` path missing). [C:/Users/leandro.theodoro.MN-NTB-LEANDROT/Documents/MetalDocs/db/migrations]
- Located files under `migrations/`: `0016_init_workflow_approvals.sql`, `0134_approval_routes.sql`, `0135_approval_instances.sql`, `0140_revision_version_and_inbox_index.sql`, `0141_governance_events_dedupe_signoff_uniqueness.sql`, `0142a_role_capabilities_v2_additive.sql`, `0142b_role_capabilities_v2_enforce.sql`, `0144_cancel_state.sql`, `0145_route_config_immutable_trigger.sql`, `0146_approval_routes_active_column.sql`, `0151_seed_dev_tenant_approval_data.sql`, `0160_grant_metaldocs_app_schema_objects.sql`, `0167_documents_bridge_and_state_columns.sql`, `0173_signoff_actor_displayname_snapshot.sql`, `0180_signoff_eligibility_trigger.sql`. [migrations/0016_init_workflow_approvals.sql:1] [migrations/0180_signoff_eligibility_trigger.sql:1]

## 1) TABLES OWNED

### `approval_instances` (owned)
Created in `0135_approval_instances.sql`. [migrations/0135_approval_instances.sql:9]

| Column | Type | Constraints / FK |
|---|---|---|
| id | UUID | PK, default `gen_random_uuid()` [migrations/0135_approval_instances.sql:10] |
| tenant_id | UUID | NOT NULL [migrations/0135_approval_instances.sql:11] |
| document_v2_id | UUID | NOT NULL, FK `documents(id)` [migrations/0135_approval_instances.sql:12] |
| route_id | UUID | NOT NULL, FK `approval_routes(id)` [migrations/0135_approval_instances.sql:13] |
| route_version_snapshot | INT | NOT NULL [migrations/0135_approval_instances.sql:14] |
| status | TEXT | NOT NULL, CHECK enum [migrations/0135_approval_instances.sql:15] [migrations/0135_approval_instances.sql:16] |
| submitted_by | TEXT | NOT NULL [migrations/0135_approval_instances.sql:17] |
| submitted_at | TIMESTAMPTZ | NOT NULL, default `now()` [migrations/0135_approval_instances.sql:18] |
| completed_at | TIMESTAMPTZ | nullable [migrations/0135_approval_instances.sql:19] |
| content_hash_at_submit | TEXT | NOT NULL [migrations/0135_approval_instances.sql:20] |
| idempotency_key | TEXT | NOT NULL [migrations/0135_approval_instances.sql:21] |

Constraints/FKs:
- UNIQUE `(document_v2_id, idempotency_key)`. [migrations/0135_approval_instances.sql:22]
- FK `(tenant_id, submitted_by)` -> `metaldocs.iam_users(tenant_id, user_id)` (`NOT VALID`). [migrations/0135_approval_instances.sql:28] [migrations/0135_approval_instances.sql:30] [migrations/0135_approval_instances.sql:31]
- No FK to `process_areas` defined in this table DDL block. [migrations/0135_approval_instances.sql:9]

### `approval_stage_instances` (owned)
Created in `0135_approval_instances.sql`. [migrations/0135_approval_instances.sql:40]

| Column | Type | Constraints / FK |
|---|---|---|
| id | UUID | PK, default `gen_random_uuid()` [migrations/0135_approval_instances.sql:41] |
| approval_instance_id | UUID | NOT NULL, FK `approval_instances(id)` ON DELETE CASCADE [migrations/0135_approval_instances.sql:42] |
| stage_order | INT | NOT NULL, CHECK >= 1 [migrations/0135_approval_instances.sql:43] |
| name_snapshot | TEXT | NOT NULL [migrations/0135_approval_instances.sql:44] |
| required_role_snapshot | TEXT | NOT NULL [migrations/0135_approval_instances.sql:45] |
| required_capability_snapshot | TEXT | NOT NULL [migrations/0135_approval_instances.sql:46] |
| area_code_snapshot | TEXT | NOT NULL [migrations/0135_approval_instances.sql:47] |
| quorum_snapshot | TEXT | NOT NULL, CHECK enum [migrations/0135_approval_instances.sql:48] [migrations/0135_approval_instances.sql:49] |
| quorum_m_snapshot | INT | nullable [migrations/0135_approval_instances.sql:50] |
| on_eligibility_drift_snapshot | TEXT | NOT NULL, CHECK enum [migrations/0135_approval_instances.sql:51] [migrations/0135_approval_instances.sql:52] |
| eligible_actor_ids | JSONB | NOT NULL [migrations/0135_approval_instances.sql:53] |
| effective_denominator | INT | nullable [migrations/0135_approval_instances.sql:54] |
| status | TEXT | NOT NULL, CHECK enum [migrations/0135_approval_instances.sql:55] [migrations/0135_approval_instances.sql:56] |
| opened_at | TIMESTAMPTZ | nullable [migrations/0135_approval_instances.sql:57] |
| completed_at | TIMESTAMPTZ | nullable [migrations/0135_approval_instances.sql:58] |

Constraints/FKs:
- UNIQUE `(approval_instance_id, stage_order)`. [migrations/0135_approval_instances.sql:59]
- UNIQUE `(id, approval_instance_id)` (composite anchor). [migrations/0135_approval_instances.sql:62]

### `approval_signoffs` (owned)
Created in `0135_approval_instances.sql`; column added in `0173_signoff_actor_displayname_snapshot.sql`. [migrations/0135_approval_instances.sql:72] [migrations/0173_signoff_actor_displayname_snapshot.sql:6]

| Column | Type | Constraints / FK |
|---|---|---|
| id | UUID | PK, default `gen_random_uuid()` [migrations/0135_approval_instances.sql:73] |
| approval_instance_id | UUID | NOT NULL, FK `approval_instances(id)` [migrations/0135_approval_instances.sql:74] |
| stage_instance_id | UUID | NOT NULL, FK `approval_stage_instances(id)` [migrations/0135_approval_instances.sql:75] |
| actor_user_id | TEXT | NOT NULL [migrations/0135_approval_instances.sql:76] |
| actor_tenant_id | UUID | NOT NULL [migrations/0135_approval_instances.sql:77] |
| decision | TEXT | NOT NULL, CHECK enum [migrations/0135_approval_instances.sql:78] |
| comment | TEXT | nullable [migrations/0135_approval_instances.sql:79] |
| signed_at | TIMESTAMPTZ | NOT NULL, default `now()` [migrations/0135_approval_instances.sql:80] |
| signature_method | TEXT | NOT NULL [migrations/0135_approval_instances.sql:81] |
| signature_payload | JSONB | NOT NULL [migrations/0135_approval_instances.sql:82] |
| content_hash | TEXT | NOT NULL [migrations/0135_approval_instances.sql:83] |
| actor_display_name_snapshot | TEXT | nullable, added with `IF NOT EXISTS` [migrations/0173_signoff_actor_displayname_snapshot.sql:7] |

Constraints/FKs:
- UNIQUE `(approval_instance_id, actor_user_id)`. [migrations/0135_approval_instances.sql:84]
- UNIQUE `(stage_instance_id, actor_user_id)`. [migrations/0135_approval_instances.sql:85]
- FK `(stage_instance_id, approval_instance_id)` -> `approval_stage_instances(id, approval_instance_id)`. [migrations/0135_approval_instances.sql:87] [migrations/0135_approval_instances.sql:89]
- FK `(actor_tenant_id, actor_user_id)` -> `metaldocs.iam_users(tenant_id, user_id)` (`NOT VALID`). [migrations/0135_approval_instances.sql:95] [migrations/0135_approval_instances.sql:97] [migrations/0135_approval_instances.sql:98]
- No FK to `process_areas` defined in this table DDL block. [migrations/0135_approval_instances.sql:72]

### `approval_routes` (catalogue)
Created in `0134_approval_routes.sql`; `active` added in `0146_approval_routes_active_column.sql`. [migrations/0134_approval_routes.sql:3] [migrations/0146_approval_routes_active_column.sql:16]

| Column | Type | Constraints / FK |
|---|---|---|
| id | UUID | PK, default `gen_random_uuid()` [migrations/0134_approval_routes.sql:4] |
| tenant_id | UUID | NOT NULL [migrations/0134_approval_routes.sql:5] |
| profile_code | TEXT | NOT NULL [migrations/0134_approval_routes.sql:6] |
| name | TEXT | NOT NULL [migrations/0134_approval_routes.sql:7] |
| version | INT | NOT NULL, default 1 [migrations/0134_approval_routes.sql:8] |
| created_at | TIMESTAMPTZ | NOT NULL, default `now()` [migrations/0134_approval_routes.sql:9] |
| created_by | TEXT | NOT NULL [migrations/0134_approval_routes.sql:10] |
| active | BOOLEAN | NOT NULL, default TRUE [migrations/0146_approval_routes_active_column.sql:16] |

Constraints/FKs:
- UNIQUE `(tenant_id, profile_code)`. [migrations/0134_approval_routes.sql:11]
- Conditional FK `(tenant_id, profile_code)` -> `metaldocs.document_profiles(tenant_id, code)`. [migrations/0134_approval_routes.sql:56] [migrations/0134_approval_routes.sql:58]
- No FK to `documents`, `iam_users`, or `process_areas` declared on `approval_routes`. [migrations/0134_approval_routes.sql:3]

### `metaldocs.idempotency_keys`
Created in `0147_idempotency_keys.sql`. [migrations/0147_idempotency_keys.sql:6]

| Column | Type | Constraints / FK |
|---|---|---|
| tenant_id | UUID | NOT NULL [migrations/0147_idempotency_keys.sql:7] |
| actor_user_id | TEXT | NOT NULL [migrations/0147_idempotency_keys.sql:8] |
| route_template | TEXT | NOT NULL [migrations/0147_idempotency_keys.sql:9] |
| key | TEXT | NOT NULL [migrations/0147_idempotency_keys.sql:10] |
| payload_hash | TEXT | NOT NULL [migrations/0147_idempotency_keys.sql:11] |
| response_status | INT | NOT NULL [migrations/0147_idempotency_keys.sql:12] |
| response_body | JSONB | NOT NULL [migrations/0147_idempotency_keys.sql:13] |
| status | TEXT | NOT NULL, CHECK enum [migrations/0147_idempotency_keys.sql:14] |
| created_at | TIMESTAMPTZ | NOT NULL, default `now()` [migrations/0147_idempotency_keys.sql:15] |
| expires_at | TIMESTAMPTZ | NOT NULL [migrations/0147_idempotency_keys.sql:16] |

Constraints/FKs:
- PRIMARY KEY `(tenant_id, actor_user_id, route_template, key)`. [migrations/0147_idempotency_keys.sql:17]
- No FK to `documents`, `iam_users`, or `process_areas` declared. [migrations/0147_idempotency_keys.sql:6]

## 2) TRIGGERS

| Trigger name | Table | Event | Function | SECURITY DEFINER | Migration anchor |
|---|---|---|---|---|---|
| `trg_require_cap_asserted_instances` | `public.approval_instances` | `BEFORE INSERT` | `public.enforce_capability_asserted` | Yes (`SECURITY DEFINER` on function) | [migrations/0142b_role_capabilities_v2_enforce.sql:201] [migrations/0142b_role_capabilities_v2_enforce.sql:202] [migrations/0142b_role_capabilities_v2_enforce.sql:203] [migrations/0142b_role_capabilities_v2_enforce.sql:70] |
| `trg_require_cap_asserted_signoffs` | `public.approval_signoffs` | `BEFORE INSERT` | `public.enforce_capability_asserted` | Yes (`SECURITY DEFINER` on function) | [migrations/0142b_role_capabilities_v2_enforce.sql:207] [migrations/0142b_role_capabilities_v2_enforce.sql:208] [migrations/0142b_role_capabilities_v2_enforce.sql:209] [migrations/0142b_role_capabilities_v2_enforce.sql:70] |
| `enforce_signoff_eligibility_trg` | `approval_signoffs` | `BEFORE INSERT` | `enforce_signoff_eligibility` | No `SECURITY DEFINER` clause in function DDL | [migrations/0180_signoff_eligibility_trigger.sql:26] [migrations/0180_signoff_eligibility_trigger.sql:27] [migrations/0180_signoff_eligibility_trigger.sql:28] [migrations/0180_signoff_eligibility_trigger.sql:8] [migrations/0180_signoff_eligibility_trigger.sql:23] |
| `trg_route_config_immutable_upd` | `approval_routes` | `BEFORE UPDATE` | `enforce_route_immutable` | No `SECURITY DEFINER` clause in function DDL | [migrations/0145_route_config_immutable_trigger.sql:32] [migrations/0145_route_config_immutable_trigger.sql:33] [migrations/0145_route_config_immutable_trigger.sql:34] [migrations/0145_route_config_immutable_trigger.sql:13] |
| `trg_route_config_immutable_del` | `approval_routes` | `BEFORE DELETE` | `enforce_route_immutable` | No `SECURITY DEFINER` clause in function DDL | [migrations/0145_route_config_immutable_trigger.sql:37] [migrations/0145_route_config_immutable_trigger.sql:38] [migrations/0145_route_config_immutable_trigger.sql:39] [migrations/0145_route_config_immutable_trigger.sql:13] |
| `trg_documents_v2_legal_transition` (cancel-state migration) | `documents` | `BEFORE UPDATE` | `enforce_document_transition` (reads `metaldocs.cancel_in_progress`) | No `SECURITY DEFINER` clause in function DDL | [migrations/0144_cancel_state.sql:90] [migrations/0144_cancel_state.sql:91] [migrations/0144_cancel_state.sql:92] [migrations/0144_cancel_state.sql:61] |

## 3) GUCs

### Read by SQL
- `metaldocs.asserted_caps` via `current_setting(..., true)` in capability trigger function. [migrations/0142b_role_capabilities_v2_enforce.sql:139]
- `metaldocs.cancel_in_progress` via `current_setting(..., true)` in document transition function. [migrations/0144_cancel_state.sql:61]
- `metaldocs.tenant_id` and `metaldocs.actor_id` are read by authz SQL path in module transactions; module tests assert `current_setting` reads for both in authz flow. [internal/modules/documents/approval/application/submit_service_test.go:168] [internal/modules/documents/approval/application/submit_service_test.go:171]

### Written by Go
- `setAuthzGUC` writes `metaldocs.tenant_id` and `metaldocs.actor_id` with `set_config`. [internal/modules/documents/approval/application/authz_guc.go:12] [internal/modules/documents/approval/application/authz_guc.go:15]
- Reject path (`RecordSignoff`) writes `metaldocs.cancel_in_progress` with `set_config`. [internal/modules/documents/approval/application/decision_service.go:334]
- Cancel path (`CancelInstance`) writes `metaldocs.cancel_in_progress` with `set_config`. [internal/modules/documents/approval/application/cancel_service.go:111]
- `ReadService.LoadInstance` now resolves the actor from request session context before writing authz GUCs; callers no longer pass `actorID` as an explicit parameter. [internal/modules/documents/approval/application/read_service.go:38] [internal/modules/documents/approval/application/read_service.go:47]

## 3.1 Governance event persistence note

- `sqlEmitter.Emit` now persists `governance_events.occurred_at` explicitly with `COALESCE($8::timestamptz, now())`, so approval event rows keep service-supplied timestamps when present and still fall back safely for zero values. [internal/modules/documents/approval/application/events.go:43] [internal/modules/documents/approval/application/events.go:52]

## 4) INDEXES (migration 0140)

| Index name | Table | Columns | Type |
|---|---|---|---|
| `ix_approval_instances_inbox` | `approval_instances` | `(tenant_id, submitted_by, status, submitted_at DESC)` | B-tree (`CREATE INDEX`) |

Anchors: [migrations/0140_revision_version_and_inbox_index.sql:23] [migrations/0140_revision_version_and_inbox_index.sql:24]

## 5) IDEMPOTENCY GRANTS (migration 0160)

- `GRANT SELECT, INSERT, UPDATE ON metaldocs.idempotency_keys TO metaldocs_app;` [migrations/0160_grant_metaldocs_app_schema_objects.sql:27] [migrations/0160_grant_metaldocs_app_schema_objects.sql:28] [migrations/0160_grant_metaldocs_app_schema_objects.sql:29]

## 6) TRIPWIRE AUDIT (`internal/modules/documents/approval/**/*.go`)

Scope query (non-test files): `INSERT/UPDATE/DELETE` on `approval_instances` and `approval_signoffs`. [internal/modules/documents/approval/repository/postgres_approval_repository.go:34] [internal/modules/documents/approval/repository/postgres_approval_repository.go:127] [internal/modules/documents/approval/repository/postgres_approval_repository.go:471] [internal/modules/documents/approval/application/obsolete_service.go:111]

| File:line | Operation | Statement target | `authz.Require` precedes in same function | Result |
|---|---|---|---|---|
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:34` | INSERT | `approval_instances` (`InsertInstance`) | No `authz.Require` call in `InsertInstance`. [internal/modules/documents/approval/repository/postgres_approval_repository.go:32] [internal/modules/documents/approval/repository/postgres_approval_repository.go:53] | FAIL |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:127` | INSERT | `approval_signoffs` (`InsertSignoff`) | No `authz.Require` call in `InsertSignoff`. [internal/modules/documents/approval/repository/postgres_approval_repository.go:114] [internal/modules/documents/approval/repository/postgres_approval_repository.go:173] | FAIL |
| `internal/modules/documents/approval/repository/postgres_approval_repository.go:471` | UPDATE | `approval_instances` (`UpdateInstanceStatus`) | No `authz.Require` call in `UpdateInstanceStatus`. [internal/modules/documents/approval/repository/postgres_approval_repository.go:469] [internal/modules/documents/approval/repository/postgres_approval_repository.go:486] | FAIL |
| `internal/modules/documents/approval/application/obsolete_service.go:111` | UPDATE | `approval_instances` (`MarkObsolete`) | Yes: `authz.Require(..., "doc.obsolete", areaCode)` at line 79 before line 111 in same function. [internal/modules/documents/approval/application/obsolete_service.go:79] [internal/modules/documents/approval/application/obsolete_service.go:111] | PASS |
