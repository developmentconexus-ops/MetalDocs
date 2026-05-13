## 1. HTTP Routes

### 1a. oapi-codegen ServerInterface (internal/modules/templates/api/api.gen.go:561-584)

ServerInterface method signatures:
- RedirectSignedUrlV2 - (GET /api/v1/signed) - internal/modules/templates/api/api.gen.go:564
- ListTemplatesV2 - (GET /api/v1/templates) - internal/modules/templates/api/api.gen.go:567
- CreateTemplateV2 - (POST /api/v1/templates) - internal/modules/templates/api/api.gen.go:570
- GetTemplateVersionV2 - (GET /api/v1/templates/{id}/versions/{n}) -
  internal/modules/templates/api/api.gen.go:573
- PresignTemplateDocxUploadUrlV2 -
  (POST /api/v1/templates/{id}/versions/{n}/docx-upload-url) -
  internal/modules/templates/api/api.gen.go:576
- SaveTemplateDraftV2 - (PUT /api/v1/templates/{id}/versions/{n}/draft) -
  internal/modules/templates/api/api.gen.go:579
- PublishTemplateVersionV2 - (POST /api/v1/templates/{id}/versions/{n}/publish) -
  internal/modules/templates/api/api.gen.go:582
- PresignTemplateSchemaUploadUrlV2 -
  (POST /api/v1/templates/{id}/versions/{n}/schema-upload-url) -
  internal/modules/templates/api/api.gen.go:585

Route mounts in lines 954-961:
- GET options.BaseURL+/api/v1/templates -> wrapper.ListTemplatesV2 -
  internal/modules/templates/api/api.gen.go:954
- POST options.BaseURL+/api/v1/templates -> wrapper.CreateTemplateV2 -
  internal/modules/templates/api/api.gen.go:955
- GET options.BaseURL+/api/v1/templates/{id}/versions/{n} -> wrapper.GetTemplateVersionV2 -
  internal/modules/templates/api/api.gen.go:956
- POST options.BaseURL+/api/v1/templates/{id}/versions/{n}/docx-upload-url ->
  wrapper.PresignTemplateDocxUploadUrlV2 - internal/modules/templates/api/api.gen.go:957
- PUT options.BaseURL+/api/v1/templates/{id}/versions/{n}/draft ->
  wrapper.SaveTemplateDraftV2 - internal/modules/templates/api/api.gen.go:958
- POST options.BaseURL+/api/v1/templates/{id}/versions/{n}/publish ->
  wrapper.PublishTemplateVersionV2 - internal/modules/templates/api/api.gen.go:959
- POST options.BaseURL+/api/v1/templates/{id}/versions/{n}/schema-upload-url ->
  wrapper.PresignTemplateSchemaUploadUrlV2 - internal/modules/templates/api/api.gen.go:960

### 1b. Hand-registered routes (internal/modules/templates/delivery/http/handler.go:39-61)

Routes in handler.go lines 39-61:
- GET /api/v1/signed -> generated.RedirectSignedUrlV2 -
  internal/modules/templates/delivery/http/handler.go:39
- GET /api/v1/templates -> generated.ListTemplatesV2 -
  internal/modules/templates/delivery/http/handler.go:40
- POST /api/v1/templates -> generated.CreateTemplateV2 -
  internal/modules/templates/delivery/http/handler.go:41
- GET /api/v1/templates/{id}/versions/{n} -> generated.GetTemplateVersionV2 -
  internal/modules/templates/delivery/http/handler.go:42
- POST /api/v1/templates/{id}/versions/{n}/docx-upload-url ->
  generated.PresignTemplateDocxUploadUrlV2 -
  internal/modules/templates/delivery/http/handler.go:43
- POST /api/v1/templates/{id}/versions/{n}/schema-upload-url ->
  generated.PresignTemplateSchemaUploadUrlV2 -
  internal/modules/templates/delivery/http/handler.go:44
- PUT /api/v1/templates/{id}/versions/{n}/draft -> generated.SaveTemplateDraftV2 -
  internal/modules/templates/delivery/http/handler.go:45
- POST /api/v1/templates/{id}/versions/{n}/publish -> generated.PublishTemplateVersionV2 -
  internal/modules/templates/delivery/http/handler.go:46
- POST /api/v1/templates/{id}/versions -> h.createNextVersion -
  internal/modules/templates/delivery/http/handler.go:48
- PUT /api/v1/templates/{id}/versions/{n}/schema -> h.updateSchemas -
  internal/modules/templates/delivery/http/handler.go:49
- POST /api/v1/templates/{id}/versions/{n}/autosave/presign -> h.presignAutosave -
  internal/modules/templates/delivery/http/handler.go:50
- POST /api/v1/templates/{id}/versions/{n}/autosave/commit -> h.commitAutosave -
  internal/modules/templates/delivery/http/handler.go:51
- POST /api/v1/templates/{id}/versions/{n}/submit -> h.submitForReview -
  internal/modules/templates/delivery/http/handler.go:52
- POST /api/v1/templates/{id}/versions/{n}/review -> h.review -
  internal/modules/templates/delivery/http/handler.go:53
- POST /api/v1/templates/{id}/versions/{n}/approve -> h.approve -
  internal/modules/templates/delivery/http/handler.go:54
- POST /api/v1/templates/{id}/archive -> h.archiveTemplate -
  internal/modules/templates/delivery/http/handler.go:55
- PUT /api/v1/templates/{id}/approval-config -> h.upsertApprovalConfig -
  internal/modules/templates/delivery/http/handler.go:56
- GET /api/v1/templates/{id} -> h.getTemplate -
  internal/modules/templates/delivery/http/handler.go:58
- GET /api/v1/templates/{id}/versions/{n}/docx-url -> h.getDocxURL -
  internal/modules/templates/delivery/http/handler.go:59
- GET /api/v1/templates/{id}/audit -> h.listAudit -
  internal/modules/templates/delivery/http/handler.go:60
- GET /api/v1/templates/v2/placeholder-catalog -> h.listPlaceholderCatalog -
  internal/modules/templates/delivery/http/handler.go:61

Additional route scan:
- internal/modules/templates/delivery/http/routes_autosave.go: no route registrations found
- internal/modules/templates/delivery/http/routes_catalog.go: no route registrations found
- internal/modules/templates/delivery/http/routes_create.go: no route registrations found
- internal/modules/templates/delivery/http/routes_generated.go: no route registrations found
- internal/modules/templates/delivery/http/routes_lifecycle.go: no route registrations found
- internal/modules/templates/delivery/http/routes_query.go: no route registrations found
- internal/modules/templates/delivery/http/routes_schema.go: no route registrations found

## 2. DB Migrations

- migrations/0101_docx_v2_templates.sql
  - CREATE TABLE: templates - migrations/0101_docx_v2_templates.sql:9
  - ALTER TABLE: none
- migrations/0102_docx_v2_template_versions.sql
  - CREATE TABLE: template_versions - migrations/0102_docx_v2_template_versions.sql:4
  - ALTER TABLE: templates - migrations/0102_docx_v2_template_versions.sql:27
- migrations/0108_docx_v2_template_audit_log.sql
  - CREATE TABLE: template_audit_log - migrations/0108_docx_v2_template_audit_log.sql:4
  - ALTER TABLE: none
- migrations/0120_templates_init.sql
  - CREATE TABLE: templates_template - migrations/0120_templates_init.sql:1
  - CREATE TABLE: templates_template_version - migrations/0120_templates_init.sql:19
  - CREATE TABLE: templates_approval_config - migrations/0120_templates_init.sql:47
  - CREATE TABLE: templates_audit_log - migrations/0120_templates_init.sql:53
  - ALTER TABLE: templates_template - migrations/0120_templates_init.sql:43
- migrations/0121_documents_v2_link_template_version.sql
  - CREATE TABLE: none
  - ALTER TABLE: documents_v2 - migrations/0121_documents_v2_link_template_version.sql:3
- migrations/0126_documents_v2_bridge_columns.sql
  - CREATE TABLE: none
  - ALTER TABLE: documents_v2 - migrations/0126_documents_v2_bridge_columns.sql:4
- migrations/0129_documents_v2_bridge_not_null.sql
  - CREATE TABLE: none
  - ALTER TABLE: documents_v2 - migrations/0129_documents_v2_bridge_not_null.sql:18
- migrations/0130_documents_drop_old_template_version_fk.sql
  - CREATE TABLE: none
  - ALTER TABLE: documents - migrations/0130_documents_drop_old_template_version_fk.sql:7
- migrations/0165_role_capabilities_reseed.sql
  - template.* capability inserts:
    - template.view - migrations/0165_role_capabilities_reseed.sql:5
    - template.view - migrations/0165_role_capabilities_reseed.sql:9
    - template.edit - migrations/0165_role_capabilities_reseed.sql:10
    - template.view - migrations/0165_role_capabilities_reseed.sql:16
    - template.create - migrations/0165_role_capabilities_reseed.sql:17
    - template.edit - migrations/0165_role_capabilities_reseed.sql:18
    - template.submit - migrations/0165_role_capabilities_reseed.sql:19
    - template.view - migrations/0165_role_capabilities_reseed.sql:26
    - template.approve - migrations/0165_role_capabilities_reseed.sql:27
    - template.view - migrations/0165_role_capabilities_reseed.sql:33
    - template.create - migrations/0165_role_capabilities_reseed.sql:34
    - template.edit - migrations/0165_role_capabilities_reseed.sql:35
    - template.submit - migrations/0165_role_capabilities_reseed.sql:36
    - template.approve - migrations/0165_role_capabilities_reseed.sql:37
    - template.publish - migrations/0165_role_capabilities_reseed.sql:38

## 3. Public Types

Exported type declarations (struct/interface):
- DocumentTemplateFieldSlotNodeResponse (struct) - internal/modules/templates/api/api.gen.go:193
- DocumentTemplateLabelNodeResponse (struct) - internal/modules/templates/api/api.gen.go:207
- DocumentTemplateNodeResponse (struct) - internal/modules/templates/api/api.gen.go:217
- DocumentTemplatePageNodeResponse (struct) - internal/modules/templates/api/api.gen.go:222
- DocumentTemplateRepeatSlotNodeResponse (struct) - internal/modules/templates/api/api.gen.go:232
- DocumentTemplateRichSlotNodeResponse (struct) - internal/modules/templates/api/api.gen.go:246
- DocumentTemplateSectionFrameNodeResponse (struct) - internal/modules/templates/api/api.gen.go:260
- DocumentTemplateTableSlotNodeResponse (struct) - internal/modules/templates/api/api.gen.go:271
- RedirectSignedUrlV2Params (struct) - internal/modules/templates/api/api.gen.go:285
- CreateTemplateV2JSONBody (struct) - internal/modules/templates/api/api.gen.go:290
- SaveTemplateDraftV2JSONBody (struct) - internal/modules/templates/api/api.gen.go:298
- PublishTemplateVersionV2JSONBody (struct) - internal/modules/templates/api/api.gen.go:307
- ServerInterface (interface) - internal/modules/templates/api/api.gen.go:561
- ServerInterfaceWrapper (struct) - internal/modules/templates/api/api.gen.go:589
- UnescapedCookieParamError (struct) - internal/modules/templates/api/api.gen.go:833
- UnmarshalingParamError (struct) - internal/modules/templates/api/api.gen.go:846
- RequiredParamError (struct) - internal/modules/templates/api/api.gen.go:859
- RequiredHeaderError (struct) - internal/modules/templates/api/api.gen.go:867
- InvalidParamFormatError (struct) - internal/modules/templates/api/api.gen.go:880
- TooManyValuesForParamError (struct) - internal/modules/templates/api/api.gen.go:893
- ServeMux (interface) - internal/modules/templates/api/api.gen.go:908
- StdHTTPServerOptions (struct) - internal/modules/templates/api/api.gen.go:913
- RedirectSignedUrlV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:965
- RedirectSignedUrlV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:969
- RedirectSignedUrlV2302Response (struct) - internal/modules/templates/api/api.gen.go:973
- ListTemplatesV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:981
- ListTemplatesV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:984
- ListTemplatesV2403Response (struct) - internal/modules/templates/api/api.gen.go:1009
- CreateTemplateV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:1017
- CreateTemplateV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:1021
- CreateTemplateV2201JSONResponse (struct) - internal/modules/templates/api/api.gen.go:1025
- GetTemplateVersionV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:1042
- GetTemplateVersionV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:1047
- GetTemplateVersionV2200Response (struct) - internal/modules/templates/api/api.gen.go:1051
- GetTemplateVersionV2404Response (struct) - internal/modules/templates/api/api.gen.go:1059
- PresignTemplateDocxUploadUrlV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:1067
- PresignTemplateDocxUploadUrlV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:1072
- PresignTemplateDocxUploadUrlV2200JSONResponse (struct) - internal/modules/templates/api/api.gen.go:1076
- SaveTemplateDraftV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:1093
- SaveTemplateDraftV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:1099
- SaveTemplateDraftV2204Response (struct) - internal/modules/templates/api/api.gen.go:1103
- SaveTemplateDraftV2409Response (struct) - internal/modules/templates/api/api.gen.go:1111
- PublishTemplateVersionV2RequestObject (struct) - internal/modules/templates/api/api.gen.go:1119
- PublishTemplateVersionV2ResponseObject (interface) - internal/modules/templates/api/api.gen.go:1125
- PublishTemplateVersionV2200JSONResponse (struct) - internal/modules/templates/api/api.gen.go:1129
- PublishTemplateVersionV2422JSONResponse (struct) - internal/modules/templates/api/api.gen.go:1147
- PresignTemplateSchemaUploadUrlV2RequestObject (struct) -
  internal/modules/templates/api/api.gen.go:1166
- PresignTemplateSchemaUploadUrlV2ResponseObject (interface) -
  internal/modules/templates/api/api.gen.go:1171
- PresignTemplateSchemaUploadUrlV2200JSONResponse (struct) -
  internal/modules/templates/api/api.gen.go:1175
- StrictServerInterface (interface) - internal/modules/templates/api/api.gen.go:1193
- StrictHTTPServerOptions (struct) - internal/modules/templates/api/api.gen.go:1223
- strictHandler (struct) - internal/modules/templates/api/api.gen.go:1243
- UpsertApprovalConfigCmd (struct) - internal/modules/templates/application/approval_config.go:9
- PresignAutosaveCmd (struct) - internal/modules/templates/application/autosave.go:13
- PresignAutosaveResult (struct) - internal/modules/templates/application/autosave.go:18
- PresignTemplateUploadCmd (struct) - internal/modules/templates/application/autosave.go:24
- CommitAutosaveCmd (struct) - internal/modules/templates/application/autosave.go:81
- SaveTemplateDraftCmd (struct) - internal/modules/templates/application/autosave.go:87
- CreateTemplateCmd (struct) - internal/modules/templates/application/create.go:11
- CreateTemplateResult (struct) - internal/modules/templates/application/create.go:25
- CreateVersionCmd (struct) - internal/modules/templates/application/create.go:109
- SubmitForReviewCmd (struct) - internal/modules/templates/application/lifecycle.go:9
- ReviewCmd (struct) - internal/modules/templates/application/lifecycle.go:66
- ApproveCmd (struct) - internal/modules/templates/application/lifecycle.go:150
- ArchiveCmd (struct) - internal/modules/templates/application/lifecycle.go:249
- PublishTemplateVersionCmd (struct) - internal/modules/templates/application/lifecycle.go:253
- PublishTemplateVersionResult (struct) - internal/modules/templates/application/lifecycle.go:260
- Repository (interface) - internal/modules/templates/application/ports.go:10
- Presigner (interface) - internal/modules/templates/application/ports.go:30
- Clock (interface) - internal/modules/templates/application/ports.go:37
- UUIDGen (interface) - internal/modules/templates/application/ports.go:38
- ResolverRegistryReader (interface) - internal/modules/templates/application/ports.go:39
- ListFilter (struct) - internal/modules/templates/application/ports.go:41
- GetDocxURLCmd (struct) - internal/modules/templates/application/queries.go:41
- UpdateSchemasCmd (struct) - internal/modules/templates/application/schema.go:12
- Service (struct) - internal/modules/templates/application/service.go:3
- Handler (struct) - internal/modules/templates/delivery/http/handler.go:19
- catalogEntry (struct) - internal/modules/templates/delivery/http/routes_catalog.go:7
- ApprovalConfig (struct) - internal/modules/templates/domain/approval.go:3
- AuditEvent (struct) - internal/modules/templates/domain/audit.go:21
- MetadataSchema (struct) - internal/modules/templates/domain/schemas.go:3
- VisibilityCondition (struct) - internal/modules/templates/domain/schemas.go:22
- Placeholder (struct) - internal/modules/templates/domain/schemas.go:28
- CompositionConfig (struct) - internal/modules/templates/domain/schemas.go:48
- Template (struct) - internal/modules/templates/domain/template.go:16
- TemplateVersion (struct) - internal/modules/templates/domain/version.go:18
- rowScanner (interface) - internal/modules/templates/repository/mappers.go:10
- Repository (struct) - internal/modules/templates/repository/postgres.go:21

Exported top-level funcs:
- Handler - func - internal/modules/templates/api/api.gen.go:903
- HandlerFromMux - func - internal/modules/templates/api/api.gen.go:921
- HandlerFromMuxWithBaseURL - func - internal/modules/templates/api/api.gen.go:927
- HandlerWithOptions - func - internal/modules/templates/api/api.gen.go:935
- NewStrictHandler - func - internal/modules/templates/api/api.gen.go:1228
- NewStrictHandlerWithOptions - func - internal/modules/templates/api/api.gen.go:1239
- PathToRawSpec - func - internal/modules/templates/api/api.gen.go:1553
- GetSpec - func - internal/modules/templates/api/api.gen.go:1567
- GetSpecJSON - func - internal/modules/templates/api/api.gen.go:1599
- GetSwagger - func - internal/modules/templates/api/api.gen.go:1609
- ValidatePlaceholders - func - internal/modules/templates/application/schema.go:84
- New - func - internal/modules/templates/application/service.go:11
- DetectVisibilityCycle - func - internal/modules/templates/application/visibility_graph.go:16
- MapErr - func - internal/modules/templates/delivery/http/errors.go:10
- New - func - internal/modules/templates/delivery/http/handler.go:24
- CheckSegregation - func - internal/modules/templates/domain/approval.go:17
- New - func - internal/modules/templates/repository/postgres.go:25

## 4. Dependencies

stdlib:
- bytes
- compress/flate
- context
- database/sql
- encoding/base64
- encoding/json
- errors
- fmt
- io
- net/http
- net/url
- path
- regexp
- strconv
- strings
- time

internal (metaldocs/internal/...):
- metaldocs/internal/modules/iam/domain -
  internal/modules/templates/delivery/http/handler.go:10,
  internal/modules/templates/delivery/http/routes_lifecycle.go:8
- metaldocs/internal/modules/templates/api -
  internal/modules/templates/delivery/http/handler.go:11,
  internal/modules/templates/delivery/http/routes_generated.go:8
- metaldocs/internal/modules/templates/application -
  internal/modules/templates/delivery/http/handler.go:12,
  internal/modules/templates/delivery/http/routes_autosave.go:8,
  internal/modules/templates/delivery/http/routes_create.go:7,
  internal/modules/templates/delivery/http/routes_generated.go:9,
  internal/modules/templates/delivery/http/routes_lifecycle.go:9,
  internal/modules/templates/delivery/http/routes_query.go:9,
  internal/modules/templates/delivery/http/routes_schema.go:7,
  internal/modules/templates/repository/postgres.go:10
- metaldocs/internal/modules/templates/domain -
  internal/modules/templates/application/approval_config.go:6,
  internal/modules/templates/application/autosave.go:8,
  internal/modules/templates/application/create.go:8,
  internal/modules/templates/application/lifecycle.go:6,
  internal/modules/templates/application/ports.go:7,
  internal/modules/templates/application/queries.go:7,
  internal/modules/templates/application/schema.go:9,
  internal/modules/templates/application/visibility_graph.go:7,
  internal/modules/templates/delivery/http/errors.go:7,
  internal/modules/templates/delivery/http/routes_create.go:8,
  internal/modules/templates/delivery/http/routes_generated.go:10,
  internal/modules/templates/delivery/http/routes_schema.go:8,
  internal/modules/templates/repository/mappers.go:7,
  internal/modules/templates/repository/postgres.go:11
- metaldocs/internal/platform/httpresponse -
  internal/modules/templates/delivery/http/handler.go:13
- metaldocs/internal/platform/tenant -
  internal/modules/templates/delivery/http/handler.go:14

external:
- github.com/getkin/kin-openapi/openapi3
- github.com/jackc/pgx/v5/pgconn
- github.com/oapi-codegen/runtime
- github.com/oapi-codegen/runtime/types

## 5. Key Constants and Enums

Const blocks and identifiers:
- internal/modules/templates/domain/audit.go:7
  - AuditCreated, AuditSaved, AuditSubmitted, AuditReviewed, AuditApproved, AuditRejected,
    AuditPublished, AuditObsoleted, AuditArchived, AuditRestored, AuditApprovalConfigUpdated
- internal/modules/templates/domain/template.go:10
  - VisibilityPublic, VisibilityInternal, VisibilitySpecific
- internal/modules/templates/domain/schemas.go:12
  - PHText, PHDate, PHNumber, PHSelect, PHUser, PHPicture, PHComputed
- internal/modules/templates/domain/version.go:10
  - VersionStatusDraft, VersionStatusInReview, VersionStatusApproved,
    VersionStatusPublished, VersionStatusObsolete
- internal/modules/templates/application/visibility_graph.go:10
  - nodeWhite, nodeGray, nodeBlack
- internal/modules/templates/api/api.gen.go:28
  - Scalar
- internal/modules/templates/api/api.gen.go:43
  - FieldSlot
- internal/modules/templates/api/api.gen.go:58
  - Label
- internal/modules/templates/api/api.gen.go:73
  - Page
- internal/modules/templates/api/api.gen.go:88
  - Repeat
- internal/modules/templates/api/api.gen.go:103
  - RepeatSlot
- internal/modules/templates/api/api.gen.go:118
  - Rich
- internal/modules/templates/api/api.gen.go:133
  - RichSlot
- internal/modules/templates/api/api.gen.go:148
  - SectionFrame
- internal/modules/templates/api/api.gen.go:163
  - Table
- internal/modules/templates/api/api.gen.go:178
  - TableSlot
