# Database Dictionary Index

> **Last verified:** 2026-05-15
> **Source:** `db/baseline/0001_current_schema.sql`

| Table | Schema | Owner | Page |
|---|---|---|---|
| `audit_events` | `metaldocs` | audit | `wiki/database/tables/audit_events.md` |
| `auth_identities` | `metaldocs` | auth | `wiki/database/tables/auth_identities.md` |
| `auth_sessions` | `metaldocs` | auth | `wiki/database/tables/auth_sessions.md` |
| `document_access_policies` | `metaldocs` | documents | `wiki/database/tables/document_access_policies.md` |
| `document_attachments` | `metaldocs` | documents | `wiki/database/tables/document_attachments.md` |
| `document_collaboration_presence` | `metaldocs` | documents | `wiki/database/tables/document_collaboration_presence.md` |
| `document_departments` | `metaldocs` | taxonomy | `wiki/database/tables/document_departments.md` |
| `document_edit_locks` | `metaldocs` | documents | `wiki/database/tables/document_edit_locks.md` |
| `document_families` | `metaldocs` | taxonomy | `wiki/database/tables/document_families.md` |
| `document_images` | `metaldocs` | documents | `wiki/database/tables/document_images.md` |
| `document_process_areas` | `metaldocs` | taxonomy | `wiki/database/tables/document_process_areas.md` |
| `document_profile_governance` | `metaldocs` | taxonomy | `wiki/database/tables/document_profile_governance.md` |
| `document_profile_schema_versions` | `metaldocs` | taxonomy | `wiki/database/tables/document_profile_schema_versions.md` |
| `document_profile_template_defaults` | `metaldocs` | taxonomy | `wiki/database/tables/document_profile_template_defaults.md` |
| `document_profiles` | `metaldocs` | taxonomy | `wiki/database/tables/document_profiles.md` |
| `document_sequences` | `metaldocs` | unknown-owner | `wiki/database/tables/document_sequences.md` |
| `document_subjects` | `metaldocs` | taxonomy | `wiki/database/tables/document_subjects.md` |
| `document_template_assignments` | `metaldocs` | templates | `wiki/database/tables/document_template_assignments.md` |
| `document_template_versions` | `metaldocs` | templates | `wiki/database/tables/document_template_versions.md` |
| `document_template_versions_mddm` | `metaldocs` | templates | `wiki/database/tables/document_template_versions_mddm.md` |
| `document_type_schema_versions` | `metaldocs` | taxonomy | `wiki/database/tables/document_type_schema_versions.md` |
| `document_types` | `metaldocs` | taxonomy | `wiki/database/tables/document_types.md` |
| `document_version_images` | `metaldocs` | documents | `wiki/database/tables/document_version_images.md` |
| `document_versions` | `metaldocs` | documents | `wiki/database/tables/document_versions.md` |
| `document_versions_mddm` | `metaldocs` | documents | `wiki/database/tables/document_versions_mddm.md` |
| `documents` | `metaldocs` | documents | `wiki/database/tables/documents.md` |
| `iam_group_members` | `metaldocs` | iam | `wiki/database/tables/iam_group_members.md` |
| `iam_group_roles` | `metaldocs` | iam | `wiki/database/tables/iam_group_roles.md` |
| `iam_groups` | `metaldocs` | iam | `wiki/database/tables/iam_groups.md` |
| `iam_user_roles` | `metaldocs` | iam | `wiki/database/tables/iam_user_roles.md` |
| `iam_users` | `metaldocs` | iam | `wiki/database/tables/iam_users.md` |
| `idempotency_keys` | `metaldocs` | registry | `wiki/database/tables/idempotency_keys.md` |
| `job_leases` | `metaldocs` | platform/workers | `wiki/database/tables/job_leases.md` |
| `mddm_shadow_diff_events` | `metaldocs` | documents | `wiki/database/tables/mddm_shadow_diff_events.md` |
| `notifications` | `metaldocs` | platform/workers | `wiki/database/tables/notifications.md` |
| `outbox_events` | `metaldocs` | platform/workers | `wiki/database/tables/outbox_events.md` |
| `pdf_dispatch_outbox` | `metaldocs` | platform/workers | `wiki/database/tables/pdf_dispatch_outbox.md` |
| `role_capabilities` | `metaldocs` | iam | `wiki/database/tables/role_capabilities.md` |
| `template_audit_log` | `metaldocs` | templates | `wiki/database/tables/template_audit_log.md` |
| `template_drafts` | `metaldocs` | templates | `wiki/database/tables/template_drafts.md` |
| `workflow_approvals` | `metaldocs` | approval | `wiki/database/tables/workflow_approvals.md` |
| `approval_instances` | `public` | approval | `wiki/database/tables/approval_instances.md` |
| `approval_route_stages` | `public` | approval | `wiki/database/tables/approval_route_stages.md` |
| `approval_routes` | `public` | approval | `wiki/database/tables/approval_routes.md` |
| `approval_signoffs` | `public` | approval | `wiki/database/tables/approval_signoffs.md` |
| `approval_stage_instances` | `public` | approval | `wiki/database/tables/approval_stage_instances.md` |
| `autosave_pending_uploads` | `public` | documents | `wiki/database/tables/autosave_pending_uploads.md` |
| `cd_sequence_counters` | `public` | registry | `wiki/database/tables/cd_sequence_counters.md` |
| `controlled_document_area_grants` | `public` | registry | `wiki/database/tables/controlled_document_area_grants.md` |
| `controlled_document_user_grants` | `public` | registry | `wiki/database/tables/controlled_document_user_grants.md` |
| `controlled_documents` | `public` | registry | `wiki/database/tables/controlled_documents.md` |
| `document_checkpoints` | `public` | documents | `wiki/database/tables/document_checkpoints.md` |
| `document_comments` | `public` | documents | `wiki/database/tables/document_comments.md` |
| `document_exports` | `public` | documents | `wiki/database/tables/document_exports.md` |
| `document_placeholder_values` | `public` | documents | `wiki/database/tables/document_placeholder_values.md` |
| `document_revisions` | `public` | documents | `wiki/database/tables/document_revisions.md` |
| `documents` | `public` | documents | `wiki/database/tables/documents.md` |
| `editor_sessions` | `public` | documents | `wiki/database/tables/editor_sessions.md` |
| `governance_events` | `public` | audit | `wiki/database/tables/governance_events.md` |
| `schema_migrations` | `public` | platform/db tooling | `wiki/database/tables/schema_migrations.md` |
| `template_audit_log` | `public` | templates | `wiki/database/tables/template_audit_log.md` |
| `template_versions` | `public` | templates | `wiki/database/tables/template_versions.md` |
| `templates` | `public` | templates | `wiki/database/tables/templates.md` |
| `templates_v2_approval_config` | `public` | templates | `wiki/database/tables/templates_v2_approval_config.md` |
| `templates_v2_audit_log` | `public` | templates | `wiki/database/tables/templates_v2_audit_log.md` |
| `templates_v2_template` | `public` | templates | `wiki/database/tables/templates_v2_template.md` |
| `templates_v2_template_version` | `public` | templates | `wiki/database/tables/templates_v2_template_version.md` |
| `user_process_areas` | `public` | iam | `wiki/database/tables/user_process_areas.md` |
