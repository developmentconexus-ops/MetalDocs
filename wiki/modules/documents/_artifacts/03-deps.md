# Documents Import Graph (03-deps)

- Module prefix: 1:module metaldocs

## 1) OUT imports (from internal/modules/documents/)

| Import | Present | Evidence |
|---|---|---|
| metaldocs/internal/modules/iam/domain | present | internal/modules/documents\http\fillin_handler.go:16 |
| metaldocs/internal/modules/iam/application | present | internal/modules/documents\delivery\http\handler.go:17 |
| metaldocs/internal/modules/iam/authz | present | internal/modules/documents\http\fillin_handler.go:15 |
| metaldocs/internal/modules/registry/domain | present | internal/modules/documents\application\create_document_snapshot_integration_test.go:15 |
| metaldocs/internal/modules/templates/domain | present | internal/modules/documents\repository\fillin_repository_integration_test.go:14 |
| metaldocs/internal/modules/render | present | internal/modules/documents\repository\resolver_readers.go:9 |
| metaldocs/internal/modules/audit/domain | absent | absent |
| metaldocs/internal/platform/idempotency | present | internal/modules/documents\approval\infrastructure\postgres_signoff_idemp_store.go:9 |
| metaldocs/internal/platform/objectstore | absent | absent |
| metaldocs/internal/platform/tenant | present | internal/modules/documents\repository\fillin_repository_integration_test.go:13 |
| metaldocs/internal/modules/auth | absent | absent |

Other internal OUT imports observed (non-self):
- metaldocs/internal/platform/docgenv2 (internal/modules/documents\application\create_document_snapshot_integration_test.go:16)
- metaldocs/internal/platform/httpresponse (internal/modules/documents\delivery\http\handler.go:20)
- metaldocs/internal/platform/ratelimit (internal/modules/documents\module.go:13)
- metaldocs/internal/platform/servicebus (internal/modules/documents\application\export_service.go:10)
- metaldocs/internal/test (internal/modules/documents\approval\application\services.go:8)

## 2) IN importers (outside internal/modules/documents/)

| Importer file | Evidence |
|---|---|
| .\\apps\\api\\cmd\\metaldocs-api\\main.go | .\\apps\\api\\cmd\\metaldocs-api\\main.go:22 |
| .\\apps\\api\\internal\\wiring\\documents.go | .\\apps\\api\\internal\\wiring\\documents.go:6 |
| .\\apps\\worker\\cmd\\metaldocs-worker\\main.go | .\\apps\\worker\\cmd\\metaldocs-worker\\main.go:8 |
| .\\internal\\modules\\iam\\integration_test.go | .\\internal\\modules\\iam\\integration_test.go:16 |
| .\\apps\\jobs\\cmd\\metaldocs-jobs\\main.go | .\\apps\\jobs\\cmd\\metaldocs-jobs\\main.go:14 |
| .\\internal\\modules\\documents\\approval\\jobs\\scheduled_publish_job.go | .\\internal\\modules\\documents\\approval\\jobs\\scheduled_publish_job.go:8 |
| .\\internal\\modules\\documents\\approval\\jobs\\scheduled_publish_job_test.go | .\\internal\\modules\\documents\\approval\\jobs\\scheduled_publish_job_test.go:14 |
| .\\internal\\modules\\jobs\\stuck_instance_watchdog\\job.go | .\\internal\\modules\\jobs\\stuck_instance_watchdog\\job.go:10 |
| .\\internal\\modules\\jobs\\stuck_instance_watchdog\\job_test.go | .\\internal\\modules\\jobs\\stuck_instance_watchdog\\job_test.go:14 |
| .\\internal\\platform\\docgenv2\\templates_snapshot_reader.go | .\\internal\\platform\\docgenv2\\templates_snapshot_reader.go:8 |
| .\\internal\\platform\\docgenv2\\templates_snapshot_reader_test.go | .\\internal\\platform\\docgenv2\\templates_snapshot_reader_test.go:6 |
| .\\internal\\platform\\objectstore\\document_presigner.go | .\\internal\\platform\\objectstore\\document_presigner.go:17 |

Expected-path verification:
- internal/modules/registry: absent
- internal/modules/approval (top-level): absent (unclear: path does not exist in workspace)
- apps/api/cmd/metaldocs-api/main.go: present (apps/api/cmd/metaldocs-api/main.go:22)
- apps/api/internal/wiring/documents.go: present (apps/api/internal/wiring/documents.go:6)
- internal/modules/render: absent
- internal/modules/templates: absent
- internal/modules/search: absent

## 3) DI / wiring touchpoints

| File:line | Symbol | Category |
|---|---|---|
| apps/api/internal/wiring/documents.go:24 | NewCapabilityChecker | constructor |
| apps/api/cmd/metaldocs-api/main.go:246 | docapp.NewSnapshotSchemaReader | application.New* |
| apps/api/cmd/metaldocs-api/main.go:249 | docapp.NewDocumentContextBuilder | application.New* |
| apps/api/cmd/metaldocs-api/main.go:253 | docapp.NewFreezeService | application.New* |
| apps/api/cmd/metaldocs-api/main.go:279 | wiring.NewCapabilityChecker | capability wiring |
| apps/api/cmd/metaldocs-api/main.go:300 | approvalapp.NewSQLEmitter | approval.New* |
| apps/api/cmd/metaldocs-api/main.go:301 | approvalapp.NewServices | approval.New* |
| apps/api/cmd/metaldocs-api/main.go:313 | approvalapp.NewDecisionService | approval.New* |
| apps/api/cmd/metaldocs-api/main.go:318 | documents.New | documents.New |
| apps/api/cmd/metaldocs-api/main.go:319 | docMod.RegisterRoutes | RegisterRoutes |
| apps/api/cmd/metaldocs-api/main.go:325 | docapp.NewCDDocumentInitializer | application.New* |
| apps/api/cmd/metaldocs-api/main.go:331 | approvalhttp.NewHandler | approval.New* |
| apps/api/cmd/metaldocs-api/main.go:332 | approvalHandler.RegisterRoutes | RegisterRoutes |

## 4) Config surface (internal/modules/documents/)

| Pattern | Result | Evidence |
|---|---|---|
| os.Getenv | absent | absent |
| viper. | absent | absent |
| cfg. | absent | absent |
| flag. | absent | absent |
| other config-like hit | present | internal/modules/documents\api\gen.go:3://go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml ../../../../api/openapi/v1/openapi.yaml |

## 5) Tests under internal/modules/documents/

| Test file | Class | Reason |
|---|---|---|
| internal/modules/documents\application\autosave_commit_branches_test.go | unit | default classification |
| internal/modules/documents\application\cd_initializer_test.go | unit | default classification |
| internal/modules/documents\application\context_builder_test.go | unit | default classification |
| internal/modules/documents\application\create_document_snapshot_integration_test.go | integration | filename contains integration_test.go |
| internal/modules/documents\application\draft_resolver_service_test.go | unit | default classification |
| internal/modules/documents\application\export_service_test.go | unit | default classification |
| internal/modules/documents\application\fillin_service_test.go | unit | default classification |
| internal/modules/documents\application\freeze_idempotency_test.go | unit | default classification |
| internal/modules/documents\application\freeze_service_test.go | unit | default classification |
| internal/modules/documents\application\iam_user_options_test.go | unit | default classification |
| internal/modules/documents\application\service_caps_test.go | unit | default classification |
| internal/modules/documents\application\service_cd_test.go | unit | default classification |
| internal/modules/documents\application\service_pagination_test.go | unit | default classification |
| internal/modules/documents\application\service_test.go | unit | default classification |
| internal/modules/documents\application\snapshot_resolver_test.go | unit | default classification |
| internal/modules/documents\application\snapshot_seeder_test.go | unit | default classification |
| internal/modules/documents\application\snapshot_service_test.go | unit | default classification |
| internal/modules/documents\application\snapshot_wire_test.go | unit | default classification |
| internal/modules/documents\approval\application\cancel_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\content_hash_test.go | unit | default classification |
| internal/modules/documents\approval\application\coverage_boost_test.go | unit | default classification |
| internal/modules/documents\approval\application\cutover_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\decision_service_freeze_test.go | unit | default classification |
| internal/modules/documents\approval\application\decision_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\events_test.go | unit | default classification |
| internal/modules/documents\approval\application\idempotency_test.go | unit | default classification |
| internal/modules/documents\approval\application\membership_tx_test.go | unit | default classification |
| internal/modules/documents\approval\application\obsolete_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\phase5_integration_test.go | integration | filename contains integration_test.go |
| internal/modules/documents\approval\application\publish_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\read_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\route_admin_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\scheduler_test_helpers_test.go | unit | default classification |
| internal/modules/documents\approval\application\submit_eligible_actors_test.go | unit | default classification |
| internal/modules/documents\approval\application\submit_service_test.go | unit | default classification |
| internal/modules/documents\approval\application\supersede_service_test.go | unit | default classification |
| internal/modules/documents\approval\domain\drift_test.go | unit | default classification |
| internal/modules/documents\approval\domain\eligibility_test.go | unit | default classification |
| internal/modules/documents\approval\domain\instance_test.go | unit | default classification |
| internal/modules/documents\approval\domain\integration_test.go | integration | filename contains integration_test.go |
| internal/modules/documents\approval\domain\quorum_test.go | unit | default classification |
| internal/modules/documents\approval\domain\route_test.go | unit | default classification |
| internal/modules/documents\approval\domain\signoff_test.go | unit | default classification |
| internal/modules/documents\approval\domain\sod_test.go | unit | default classification |
| internal/modules/documents\approval\domain\state_test.go | unit | default classification |
| internal/modules/documents\approval\http\cancel_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\contracts\contracts_test.go | unit | default classification |
| internal/modules/documents\approval\http\errors_test.go | unit | default classification |
| internal/modules/documents\approval\http\get_instance_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\inbox_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\obsolete_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\publish_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\route_admin_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\router_test.go | unit | default classification |
| internal/modules/documents\approval\http\signoff_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\submit_handler_test.go | unit | default classification |
| internal/modules/documents\approval\http\supersede_handler_test.go | unit | default classification |
| internal/modules/documents\approval\infra\signature\password_reauth_test.go | unit | default classification |
| internal/modules/documents\approval\infra\signature\registry_test.go | unit | default classification |
| internal/modules/documents\approval\infrastructure\postgres_signoff_idemp_store_test.go | unit | default classification |
| internal/modules/documents\approval\repository\postgres_approval_repository_test.go | unit | default classification |
| internal/modules/documents\delivery\http\finalize_wiring_test.go | unit | default classification |
| internal/modules/documents\delivery\http\handler_comments_test.go | unit | default classification |
| internal/modules/documents\delivery\http\handler_pagination_test.go | unit | default classification |
| internal/modules/documents\delivery\http\handler_test.go | unit | default classification |
| internal/modules/documents\domain\composite_hash_test.go | unit | default classification |
| internal/modules/documents\domain\model_test.go | unit | default classification |
| internal/modules/documents\domain\snapshot_test.go | unit | default classification |
| internal/modules/documents\domain\values_hash_test.go | unit | default classification |
| internal/modules/documents\http\fillin_handler_test.go | unit | default classification |
| internal/modules/documents\http\pdf_webhook_handler_test.go | unit | default classification |
| internal/modules/documents\http\placeholder_options_handler_test.go | unit | default classification |
| internal/modules/documents\http\rbac_test.go | unit | default classification |
| internal/modules/documents\http\reconstruct_handler_test.go | unit | default classification |
| internal/modules/documents\http\view_handler_test.go | unit | default classification |
| internal/modules/documents\jobs\jobs_test.go | unit | default classification |
| internal/modules/documents\repository\fillin_repository_integration_test.go | integration | filename contains integration_test.go |
| internal/modules/documents\repository\fillin_repository_test.go | unit | default classification |
| internal/modules/documents\repository\repository_archive_test.go | unit | default classification |
| internal/modules/documents\repository\repository_create_integration_test.go | integration | filename contains integration_test.go |
| internal/modules/documents\repository\repository_create_test.go | unit | default classification |
| internal/modules/documents\repository\repository_integration_test.go | integration | filename contains integration_test.go |
| internal/modules/documents\repository\repository_list_test.go | unit | default classification |
| internal/modules/documents\repository\repository_pagination_test.go | unit | default classification |
| internal/modules/documents\repository\repository_stats_test.go | unit | default classification |
| internal/modules/documents\repository\snapshot_repository_test.go | unit | default classification |
