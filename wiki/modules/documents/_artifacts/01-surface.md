# Documents Module Surface

## 1) File tree
- internal/modules/documents/
  - internal/modules/documents/module.go - (undocumented)
- internal/modules/documents/api/
  - internal/modules/documents/api/api.gen.go - (undocumented)
  - internal/modules/documents/api/gen.go - (undocumented)
  - internal/modules/documents/api/security_context.go - (undocumented)
- internal/modules/documents/application/
  - internal/modules/documents/application/cd_initializer.go - CDDocumentInitializer adapts the documents Service to the registry-owned
  - internal/modules/documents/application/context_builder.go - DocumentContextBuilder builds resolvers.ResolveInput for a document revision.
  - internal/modules/documents/application/draft_resolver_service.go - (undocumented)
  - internal/modules/documents/application/export_service.go - (undocumented)
  - internal/modules/documents/application/fillin_authz.go - requireDocEditDraft opens a short tx, sets authz GUCs, resolves the
  - internal/modules/documents/application/fillin_service.go - (undocumented)
  - internal/modules/documents/application/freeze_service.go - (undocumented)
  - internal/modules/documents/application/iam_user_options.go - (undocumented)
  - internal/modules/documents/application/list_options.go - ListOptions is an alias for repository.ListOptions so handlers depend
  - internal/modules/documents/application/ports.go - CapabilityChecker is the documents-module consumer port for tier-1
  - internal/modules/documents/application/reconstruct_service.go - (undocumented)
  - internal/modules/documents/application/service.go - Type aliases so handlers depend only on application types.
  - internal/modules/documents/application/snapshot_service.go - SnapshotTemplateReader loads a template's artifact data for snapshotting.
  - internal/modules/documents/application/view_service.go - ViewPresigner is implemented by objectstore helpers that presign a GET URL.
- internal/modules/documents/approval/application/
  - internal/modules/documents/approval/application/authz_guc.go - setAuthzGUC sets the tenant_id and actor_id GUC variables needed by authz.Require.
  - internal/modules/documents/approval/application/cancel_service.go - CancelService cancels an in-progress approval instance and reverts the
  - internal/modules/documents/approval/application/content_hash.go - ErrFloatInFormData returned when form_data contains float64 values (spec rejects floats).
  - internal/modules/documents/approval/application/cutover_service.go - ErrLegacyDocumentsRemain is returned by ValidateLegacyCutoverReady when one
  - internal/modules/documents/approval/application/decision_service.go - (undocumented)
  - internal/modules/documents/approval/application/events.go - GovernanceEvent mirrors the governance_events table columns.
  - internal/modules/documents/approval/application/idempotency.go - IdempotencyInput holds the fields for key derivation.
  - internal/modules/documents/approval/application/membership_tx.go - (undocumented)
  - internal/modules/documents/approval/application/obsolete_service.go - ObsoleteService marks a document as obsolete (end-of-life).
  - internal/modules/documents/approval/application/publish_service.go - PublishService handles transitioning an approved document to published state.
  - internal/modules/documents/approval/application/read_service.go - InboxView is the read-model projection for the inbox UI.
  - internal/modules/documents/approval/application/route_admin_service.go - RouteAdminService manages approval route configuration changes.
  - internal/modules/documents/approval/application/scheduler_service.go - SchedulerService processes scheduled publish jobs (F6 — ListScheduledDue).
  - internal/modules/documents/approval/application/services.go - Clock abstracts time so services can be tested deterministically.
  - internal/modules/documents/approval/application/submit_service.go - SubmitService handles document submission for approval.
  - internal/modules/documents/approval/application/supersede_service.go - SupersedeService marks a published document as superseded by a newer revision.
- internal/modules/documents/approval/domain/
  - internal/modules/documents/approval/domain/drift.go - DriftResult holds the output of ApplyEligibilityDrift.
  - internal/modules/documents/approval/domain/eligibility.go - ErrActorNotEligible is returned when an actor is not in the eligible_actor_ids
  - internal/modules/documents/approval/domain/instance.go - (undocumented)
  - internal/modules/documents/approval/domain/quorum.go - QuorumOutcome is the result of evaluating quorum for a stage.
  - internal/modules/documents/approval/domain/route.go - QuorumPolicy defines how many signoffs satisfy a stage.
  - internal/modules/documents/approval/domain/signoff.go - (undocumented)
  - internal/modules/documents/approval/domain/sod.go - (undocumented)
  - internal/modules/documents/approval/domain/state.go - ErrLegacyStateRejected returned when legacy state string (finalized, archived) parsed.
- internal/modules/documents/approval/http/
  - internal/modules/documents/approval/http/cancel_handler.go - (undocumented)
  - internal/modules/documents/approval/http/doc_approval_handler.go - GetInstanceByDocumentHandler handles GET /api/v1/documents/{id}/approval-instance.
  - internal/modules/documents/approval/http/errors.go - (undocumented)
  - internal/modules/documents/approval/http/get_instance_handler.go - (undocumented)
  - internal/modules/documents/approval/http/handler.go - (undocumented)
  - internal/modules/documents/approval/http/inbox_handler.go - (undocumented)
  - internal/modules/documents/approval/http/obsolete_handler.go - (undocumented)
  - internal/modules/documents/approval/http/publish_handler.go - (undocumented)
  - internal/modules/documents/approval/http/route_admin_handler.go - (undocumented)
  - internal/modules/documents/approval/http/router.go - RegisterRoutes wires all approval routes onto mux.
  - internal/modules/documents/approval/http/signoff_handler.go - (undocumented)
  - internal/modules/documents/approval/http/submit_handler.go - (undocumented)
  - internal/modules/documents/approval/http/supersede_handler.go - (undocumented)
- internal/modules/documents/approval/http/contracts/
  - internal/modules/documents/approval/http/contracts/cancel.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/errors.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/instance_read.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/obsolete.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/publish.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/route.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/signoff.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/strictjson.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/submit.go - (undocumented)
  - internal/modules/documents/approval/http/contracts/supersede.go - (undocumented)
- internal/modules/documents/approval/infra/signature/
  - internal/modules/documents/approval/infra/signature/password_reauth.go - (undocumented)
  - internal/modules/documents/approval/infra/signature/provider.go - ErrUnknownSignatureMethod returned when registry has no Provider for the given method.
- internal/modules/documents/approval/infrastructure/
  - internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go - (undocumented)
- internal/modules/documents/approval/repository/
  - internal/modules/documents/approval/repository/approval_repository.go - SignoffInsertResult returned by InsertSignoff.
  - internal/modules/documents/approval/repository/errors.go - (undocumented)
  - internal/modules/documents/approval/repository/postgres_approval_repository.go - (undocumented)
- internal/modules/documents/delivery/http/
  - internal/modules/documents/delivery/http/export_handler.go - (undocumented)
  - internal/modules/documents/delivery/http/handler.go - (undocumented)
- internal/modules/documents/domain/
  - internal/modules/documents/domain/comment.go - (undocumented)
  - internal/modules/documents/domain/composite_hash.go - RenderOptions captures all user-controlled PDF rendering knobs.
  - internal/modules/documents/domain/errors.go - (undocumented)
  - internal/modules/documents/domain/export.go - Export is an immutable row in document_exports representing one cached PDF.
  - internal/modules/documents/domain/model.go - (undocumented)
  - internal/modules/documents/domain/snapshot.go - (undocumented)
  - internal/modules/documents/domain/state.go - CanTransitionDocument returns true iff a document can move from cur to next.
  - internal/modules/documents/domain/values_hash.go - (undocumented)
- internal/modules/documents/http/
  - internal/modules/documents/http/fillin_handler.go - (undocumented)
  - internal/modules/documents/http/pdf_webhook_handler.go - PDFWriter persists PDF-completion columns on documents.
  - internal/modules/documents/http/placeholder_options_handler.go - (undocumented)
  - internal/modules/documents/http/reconstruct_handler.go - (undocumented)
  - internal/modules/documents/http/view_handler.go - (undocumented)
- internal/modules/documents/jobs/
  - internal/modules/documents/jobs/orphan_pending_sweeper.go - (undocumented)
  - internal/modules/documents/jobs/session_sweeper.go - (undocumented)
- internal/modules/documents/repository/
  - internal/modules/documents/repository/export_repository.go - (undocumented)
  - internal/modules/documents/repository/fillin_repository.go - FillInRepository manages document_placeholder_values rows.
  - internal/modules/documents/repository/repository.go - isInvalidUUID returns true when err is a Postgres error with SQLSTATE 22P02
  - internal/modules/documents/repository/resolver_readers.go - RevisionReader implements resolvers.RevisionReader backed by the documents table.
  - internal/modules/documents/repository/snapshot_repository.go - SnapshotRepository reads and writes the template snapshot columns on documents.

Test file counts by subpackage (excluded from tree rows):
- internal/modules/documents/application: 18
- internal/modules/documents/approval/application: 18
- internal/modules/documents/approval/domain: 9
- internal/modules/documents/approval/http: 11
- internal/modules/documents/approval/http/contracts: 1
- internal/modules/documents/approval/infra/signature: 2
- internal/modules/documents/approval/infrastructure: 1
- internal/modules/documents/approval/repository: 1
- internal/modules/documents/delivery/http: 4
- internal/modules/documents/domain: 4
- internal/modules/documents/http: 6
- internal/modules/documents/jobs: 1
- internal/modules/documents/repository: 10

## 2) Public surface
| File:line | Kind | Name | Signature / receiver | Doc comment first line |
|---|---|---|---|---|
| internal/modules/documents/application/cd_initializer.go:14 | type | CDDocumentInitializer | type CDDocumentInitializer struct { | CDDocumentInitializer adapts the documents Service to the registry-owned |
| internal/modules/documents/application/cd_initializer.go:18 | func | NewCDDocumentInitializer | func NewCDDocumentInitializer(svc *Service) *CDDocumentInitializer { | (undocumented) |
| internal/modules/documents/application/cd_initializer.go:26 | var | formData | var formData json.RawMessage | (undocumented) |
| internal/modules/documents/application/context_builder.go:14 | type | DocumentContextBuilder | type DocumentContextBuilder struct { | DocumentContextBuilder builds resolvers.ResolveInput for a document revision. |
| internal/modules/documents/application/context_builder.go:24 | func | NewDocumentContextBuilder | func NewDocumentContextBuilder( | NewDocumentContextBuilder wires a DocumentContextBuilder. |
| internal/modules/documents/application/context_builder.go:48 | const | activeInstanceSQL | const activeInstanceSQL = ` | (undocumented) |
| internal/modules/documents/application/context_builder.go:60 | var | instanceID | var instanceID sql.NullString | (undocumented) |
| internal/modules/documents/application/draft_resolver_service.go:13 | type | DraftResolverService | type DraftResolverService struct { | (undocumented) |
| internal/modules/documents/application/draft_resolver_service.go:23 | func | NewDraftResolverService | func NewDraftResolverService( | (undocumented) |
| internal/modules/documents/application/export_service.go:13 | type | ExportRepo | type ExportRepo interface { | (undocumented) |
| internal/modules/documents/application/export_service.go:20 | type | ExportPresigner | type ExportPresigner interface { | (undocumented) |
| internal/modules/documents/application/export_service.go:26 | type | DocgenPDFClient | type DocgenPDFClient interface { | (undocumented) |
| internal/modules/documents/application/export_service.go:30 | type | ExportService | type ExportService struct { | (undocumented) |
| internal/modules/documents/application/export_service.go:39 | func | NewExportService | func NewExportService(repo ExportRepo, presigner ExportPresigner, docgen DocgenPDFClient, audit Audit, docgenVer, grammarVer string) *ExportService { | (undocumented) |
| internal/modules/documents/application/fillin_authz.go:16 | func | requireDocEditDraft | func requireDocEditDraft(ctx context.Context, db *sql.DB, tenantID, actorID, docID string) error { | requireDocEditDraft opens a short tx, sets authz GUCs, resolves the |
| internal/modules/documents/application/fillin_authz.go:35 | func | setAuthzGUC | func setAuthzGUC(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error { | (undocumented) |
| internal/modules/documents/application/fillin_authz.go:45 | func | loadDocumentAreaCode | func loadDocumentAreaCode(ctx context.Context, tx *sql.Tx, tenantID, documentID string) (string, error) { | (undocumented) |
| internal/modules/documents/application/fillin_authz.go:46 | var | areaCode | var areaCode string | (undocumented) |
| internal/modules/documents/application/fillin_service.go:151 | func | findPlaceholder | func findPlaceholder(phs []templatesdomain.Placeholder, id string) (templatesdomain.Placeholder, bool) { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:160 | func | validateValue | func validateValue(ctx context.Context, tenantID string, p templatesdomain.Placeholder, raw string, iam IAMUserOptionsReader) error { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:18 | type | SchemaReader | type SchemaReader interface { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:22 | type | FillInWriter | type FillInWriter interface { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:227 | type | FillInReader | type FillInReader interface { | FillInReader reads current fill-in values from the DB. |
| internal/modules/documents/application/fillin_service.go:233 | type | TemplateVersionSchemaReader | type TemplateVersionSchemaReader struct { | TemplateVersionSchemaReader reads fill-in schema from the template version |
| internal/modules/documents/application/fillin_service.go:237 | func | NewTemplateVersionSchemaReader | func NewTemplateVersionSchemaReader(db *sql.DB) *TemplateVersionSchemaReader { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:242 | var | pRaw | var pRaw []byte | (undocumented) |
| internal/modules/documents/application/fillin_service.go:256 | var | placeholders | var placeholders []templatesdomain.Placeholder | (undocumented) |
| internal/modules/documents/application/fillin_service.go:26 | type | draftResolver | type draftResolver interface { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:30 | type | FillInService | type FillInService struct { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:42 | func | NewFillInService | func NewFillInService(db *sql.DB, s SchemaReader, w FillInWriter) *FillInService { | NewFillInService wires the service with a DB handle for authz enforcement. |
| internal/modules/documents/application/fillin_service.go:48 | func | NewFillInServiceNoAuthz | func NewFillInServiceNoAuthz(s SchemaReader, w FillInWriter) *FillInService { | NewFillInServiceNoAuthz is a TEST-ONLY constructor that skips capability checks. |
| internal/modules/documents/application/fillin_service.go:65 | type | SnapshotSchemaReader | type SnapshotSchemaReader struct { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:69 | func | NewSnapshotSchemaReader | func NewSnapshotSchemaReader(db *sql.DB) *SnapshotSchemaReader { | (undocumented) |
| internal/modules/documents/application/fillin_service.go:74 | var | raw | var raw []byte | (undocumented) |
| internal/modules/documents/application/fillin_service.go:88 | func | parsePlaceholderSchema | func parsePlaceholderSchema(raw []byte) ([]templatesdomain.Placeholder, error) { | parsePlaceholderSchema handles both storage formats: |
| internal/modules/documents/application/fillin_service.go:93 | var | phs | var phs []templatesdomain.Placeholder | Try raw array first (eigenpal format). |
| internal/modules/documents/application/fillin_service.go:98 | var | wrapped | var wrapped struct { | Fallback: wrapped object format. |
| internal/modules/documents/application/freeze_service.go:18 | type | FreezeFinalizer | type FreezeFinalizer interface { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:22 | type | SnapshotReader | type SnapshotReader interface { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:26 | type | FinalDocxWriter | type FinalDocxWriter interface { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:30 | type | FanoutClient | type FanoutClient interface { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:34 | type | FreezeService | type FreezeService struct { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:48 | type | ApproverContext | type ApproverContext struct { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:53 | type | ResolverContextBuilder | type ResolverContextBuilder interface { | (undocumented) |
| internal/modules/documents/application/freeze_service.go:58 | func | NewFreezeService | func NewFreezeService( | (undocumented) |
| internal/modules/documents/application/iam_user_options.go:14 | type | IAMUserOptionsReader | type IAMUserOptionsReader interface { | IAMUserOptionsReader is the consumer-defined port for listing user options. |
| internal/modules/documents/application/iam_user_options.go:20 | type | IAMUser | type IAMUser struct { | IAMUser is a minimal user record from the auth system. |
| internal/modules/documents/application/iam_user_options.go:27 | type | IAMUserLister | type IAMUserLister interface { | IAMUserLister is the narrow auth-system port consumed by IAMUserOptionsAdapter. |
| internal/modules/documents/application/iam_user_options.go:33 | type | IAMUserOptionsAdapter | type IAMUserOptionsAdapter struct { | IAMUserOptionsAdapter adapts an IAMUserLister to IAMUserOptionsReader. |
| internal/modules/documents/application/iam_user_options.go:37 | func | NewIAMUserOptionsAdapter | func NewIAMUserOptionsAdapter(lister IAMUserLister) *IAMUserOptionsAdapter { | (undocumented) |
| internal/modules/documents/application/iam_user_options.go:8 | type | UserOption | type UserOption struct { | (undocumented) |
| internal/modules/documents/application/list_options.go:8 | type | ListOptions | type ListOptions = repository.ListOptions | ListOptions is an alias for repository.ListOptions so handlers depend |
| internal/modules/documents/application/ports.go:12 | type | CapabilityChecker | type CapabilityChecker interface { | CapabilityChecker is the documents-module consumer port for tier-1 |
| internal/modules/documents/application/reconstruct_service.go:12 | type | ReconstructionRunner | type ReconstructionRunner interface { | (undocumented) |
| internal/modules/documents/application/reconstruct_service.go:16 | type | ReconstructionService | type ReconstructionService struct { | (undocumented) |
| internal/modules/documents/application/reconstruct_service.go:21 | func | NewReconstructionService | func NewReconstructionService(db *sql.DB, runner ReconstructionRunner) *ReconstructionService { | (undocumented) |
| internal/modules/documents/application/service.go:113 | func | New | func New(r Repository, d DocgenRenderer, p Presigner, t TemplateReader, fv FormValidator, a Audit) *Service { | (undocumented) |
| internal/modules/documents/application/service.go:124 | func | NewService | func NewService( | (undocumented) |
| internal/modules/documents/application/service.go:150 | func | NewServiceWithSnapshot | func NewServiceWithSnapshot( | NewServiceWithSnapshot is like NewService but also wires a SnapshotService |
| internal/modules/documents/application/service.go:181 | var | ErrControlledDocumentRequired | var ErrControlledDocumentRequired = errors.New("controlled_document_id is required") | (undocumented) |
| internal/modules/documents/application/service.go:182 | var | errRegistryReaderNotConfigured | var errRegistryReaderNotConfigured = errors.New("registry reader not configured") | (undocumented) |
| internal/modules/documents/application/service.go:183 | var | errRegistryDuplicatorNotConfigured | var errRegistryDuplicatorNotConfigured = errors.New("registry duplicator not configured") | (undocumented) |
| internal/modules/documents/application/service.go:184 | var | errCapabilityCheckerNotConfigured | var errCapabilityCheckerNotConfigured = errors.New("capability checker not configured") | (undocumented) |
| internal/modules/documents/application/service.go:185 | var | errProfileTemplateReaderNotConfigured | var errProfileTemplateReaderNotConfigured = errors.New("profile default template reader not configured") | (undocumented) |
| internal/modules/documents/application/service.go:187 | type | CreateDocumentInput | type CreateDocumentInput struct { | (undocumented) |
| internal/modules/documents/application/service.go:196 | type | CreateDocumentCmd | type CreateDocumentCmd = CreateDocumentInput | (undocumented) |
| internal/modules/documents/application/service.go:198 | type | CreateDocumentResult | type CreateDocumentResult struct { | (undocumented) |
| internal/modules/documents/application/service.go:204 | func | buildDocumentForCreate | func buildDocumentForCreate(cmd CreateDocumentInput, cd *registrydomain.ControlledDocument, resolvedTemplateVersionID string) domain.Document { | (undocumented) |
| internal/modules/documents/application/service.go:23 | type | PendingCommitMeta | type PendingCommitMeta = repository.PendingCommitMeta | Type aliases so handlers depend only on application types. |
| internal/modules/documents/application/service.go:24 | type | CommitResult | type CommitResult = repository.CommitResult | (undocumented) |
| internal/modules/documents/application/service.go:25 | type | RestoreResult | type RestoreResult = repository.RestoreResult | (undocumented) |
| internal/modules/documents/application/service.go:253 | var | overrideTemplate | var overrideTemplate *registrydomain.TemplateVersionCandidate | (undocumented) |
| internal/modules/documents/application/service.go:263 | var | defaultTemplate | var defaultTemplate *registrydomain.TemplateVersionCandidate | (undocumented) |
| internal/modules/documents/application/service.go:27 | type | Repository | type Repository interface { | (undocumented) |
| internal/modules/documents/application/service.go:298 | var | snap | var snap domain.TemplateSnapshot | Resolve template snapshot pre-INSERT so snapshot columns are written |
| internal/modules/documents/application/service.go:299 | var | phs | var phs []templatesdomain.Placeholder | (undocumented) |
| internal/modules/documents/application/service.go:301 | var | resolveErr | var resolveErr error | (undocumented) |
| internal/modules/documents/application/service.go:308 | var | contentHash | var contentHash, finalKey string | (undocumented) |
| internal/modules/documents/application/service.go:311 | var | err | var err error | (undocumented) |
| internal/modules/documents/application/service.go:369 | type | cloneIntoTxInput | type cloneIntoTxInput struct { | cloneIntoTxInput is the internal payload for cloneIntoTx. It mirrors the |
| internal/modules/documents/application/service.go:406 | var | overrideTemplate | var overrideTemplate *registrydomain.TemplateVersionCandidate | (undocumented) |
| internal/modules/documents/application/service.go:416 | var | defaultTemplate | var defaultTemplate *registrydomain.TemplateVersionCandidate | (undocumented) |
| internal/modules/documents/application/service.go:442 | var | snap | var snap domain.TemplateSnapshot | Resolve template snapshot for the same atomicity guarantees as |
| internal/modules/documents/application/service.go:443 | var | phs | var phs []templatesdomain.Placeholder | (undocumented) |
| internal/modules/documents/application/service.go:520 | type | DocumentStats | type DocumentStats struct { | (undocumented) |
| internal/modules/documents/application/service.go:61 | type | DocgenRenderer | type DocgenRenderer interface { | (undocumented) |
| internal/modules/documents/application/service.go:623 | type | PresignAutosaveCmd | type PresignAutosaveCmd struct { | (undocumented) |
| internal/modules/documents/application/service.go:627 | type | PresignAutosaveResult | type PresignAutosaveResult struct { | (undocumented) |
| internal/modules/documents/application/service.go:65 | type | Presigner | type Presigner interface { | (undocumented) |
| internal/modules/documents/application/service.go:653 | type | CommitAutosaveCmd | type CommitAutosaveCmd struct { | (undocumented) |
| internal/modules/documents/application/service.go:73 | type | TemplateReader | type TemplateReader interface { | (undocumented) |
| internal/modules/documents/application/service.go:77 | type | FormValidator | type FormValidator interface { | (undocumented) |
| internal/modules/documents/application/service.go:81 | type | Audit | type Audit interface { | (undocumented) |
| internal/modules/documents/application/service.go:86 | type | RegistryReader | type RegistryReader interface { | RegistryReader loads a ControlledDocument for validation at create time. |
| internal/modules/documents/application/service.go:90 | type | RegistryDuplicator | type RegistryDuplicator interface { | (undocumented) |
| internal/modules/documents/application/service.go:94 | type | ProfileDefaultTemplateReader | type ProfileDefaultTemplateReader interface { | (undocumented) |
| internal/modules/documents/application/service.go:99 | type | Service | type Service struct { | (undocumented) |
| internal/modules/documents/application/snapshot_service.go:12 | type | SnapshotTemplateReader | type SnapshotTemplateReader interface { | SnapshotTemplateReader loads a template's artifact data for snapshotting. |
| internal/modules/documents/application/snapshot_service.go:17 | type | SnapshotWriter | type SnapshotWriter interface { | SnapshotWriter persists snapshot columns on a document. |
| internal/modules/documents/application/snapshot_service.go:22 | type | PlaceholderValueSeeder | type PlaceholderValueSeeder interface { | PlaceholderValueSeeder seeds default placeholder value rows for a revision. |
| internal/modules/documents/application/snapshot_service.go:28 | type | SnapshotService | type SnapshotService struct { | SnapshotService copies template artifacts onto a document at creation time |
| internal/modules/documents/application/snapshot_service.go:35 | func | NewSnapshotService | func NewSnapshotService(t SnapshotTemplateReader, w SnapshotWriter) *SnapshotService { | NewSnapshotService constructs a SnapshotService without a seeder. |
| internal/modules/documents/application/snapshot_service.go:41 | func | NewSnapshotServiceWithSeeder | func NewSnapshotServiceWithSeeder(t SnapshotTemplateReader, w SnapshotWriter, s PlaceholderValueSeeder) *SnapshotService { | NewSnapshotServiceWithSeeder constructs a SnapshotService with a seeder |
| internal/modules/documents/application/snapshot_service.go:87 | func | parseRequiredPlaceholders | func parseRequiredPlaceholders(schemaJSON []byte) ([]templatesdomain.Placeholder, error) { | parseRequiredPlaceholders extracts placeholders with Required=true from |
| internal/modules/documents/application/snapshot_service.go:92 | var | out | var out []templatesdomain.Placeholder | (undocumented) |
| internal/modules/documents/application/view_service.go:15 | type | ViewPresigner | type ViewPresigner interface { | ViewPresigner is implemented by objectstore helpers that presign a GET URL. |
| internal/modules/documents/application/view_service.go:21 | type | PDFOutboxStateReader | type PDFOutboxStateReader interface { | PDFOutboxStateReader returns the latest pdf_outbox state for a revision/document. |
| internal/modules/documents/application/view_service.go:27 | type | ViewService | type ViewService struct { | ViewService serves viewer requests by checking area-scoped RBAC, validating |
| internal/modules/documents/application/view_service.go:33 | func | NewViewService | func NewViewService(db *sql.DB, presigner ViewPresigner, outbox PDFOutboxStateReader) *ViewService { | (undocumented) |
| internal/modules/documents/application/view_service.go:37 | var | viewableStatuses | var viewableStatuses = map[string]struct{}{ | (undocumented) |
| internal/modules/documents/application/view_service.go:55 | var | status | var status, areaCode string | (undocumented) |
| internal/modules/documents/application/view_service.go:56 | var | pdfKey | var pdfKey sql.NullString | (undocumented) |
| internal/modules/documents/approval/application/authz_guc.go:11 | func | setAuthzGUC | func setAuthzGUC(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error { | setAuthzGUC sets the tenant_id and actor_id GUC variables needed by authz.Require. |
| internal/modules/documents/approval/application/cancel_service.go:17 | type | CancelService | type CancelService struct { | CancelService cancels an in-progress approval instance and reverts the |
| internal/modules/documents/approval/application/cancel_service.go:189 | func | newCancelService | func newCancelService(repo repository.ApprovalRepository, emitter EventEmitter, clock Clock) *CancelService { | newCancelService constructs a CancelService (wired by NewServices). |
| internal/modules/documents/approval/application/cancel_service.go:24 | var | ErrReasonRequired | var ErrReasonRequired = errors.New("cancel: reason must not be empty") | ErrReasonRequired is returned when CancelInput.Reason is empty. |
| internal/modules/documents/approval/application/cancel_service.go:27 | type | CancelInput | type CancelInput struct { | CancelInput carries all inputs for CancelService.CancelInstance. |
| internal/modules/documents/approval/application/cancel_service.go:40 | type | CancelResult | type CancelResult struct { | CancelResult is returned on a successful cancellation. |
| internal/modules/documents/approval/application/cancel_service.go:77 | var | areaCode | var areaCode string | Fetch document area_code for authz check. FOR UPDATE locks the document row |
| internal/modules/documents/approval/application/content_hash.go:117 | func | validateNoFloats | func validateNoFloats(m map[string]any) error { | (undocumented) |
| internal/modules/documents/approval/application/content_hash.go:121 | func | walkAny | func walkAny(v any) error { | (undocumented) |
| internal/modules/documents/approval/application/content_hash.go:15 | var | ErrFloatInFormData | var ErrFloatInFormData = errors.New("content hash: float64 values are not allowed in form_data") | ErrFloatInFormData returned when form_data contains float64 values (spec rejects floats). |
| internal/modules/documents/approval/application/content_hash.go:18 | type | ContentHashInput | type ContentHashInput struct { | ContentHashInput holds the fields canonicalized into the content hash. |
| internal/modules/documents/approval/application/content_hash.go:35 | func | ComputeContentHash | func ComputeContentHash(input ContentHashInput) (string, error) { | ComputeContentHash returns the lowercase hex SHA-256 of the canonical JSON encoding. |
| internal/modules/documents/approval/application/content_hash.go:57 | func | canonicalize | func canonicalize(v any) ([]byte, error) { | canonicalize produces deterministic, whitespace-free JSON with sorted keys. |
| internal/modules/documents/approval/application/cutover_service.go:14 | var | ErrLegacyDocumentsRemain | var ErrLegacyDocumentsRemain = errors.New("cutover: legacy documents remain with status 'finalized' or 'archived'") | ErrLegacyDocumentsRemain is returned by ValidateLegacyCutoverReady when one |
| internal/modules/documents/approval/application/cutover_service.go:22 | type | CutoverService | type CutoverService struct { | CutoverService validates preconditions for the Phase 5.10 legacy cutover. |
| internal/modules/documents/approval/application/cutover_service.go:28 | func | NewCutoverService | func NewCutoverService(emitter EventEmitter, clock Clock) *CutoverService { | NewCutoverService constructs a CutoverService. |
| internal/modules/documents/approval/application/cutover_service.go:40 | var | count | var count int64 | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:19 | type | FreezeInvoker | type FreezeInvoker interface { | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:190 | var | actorDisplayName | var actorDisplayName sql.NullString | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:23 | type | PDFDispatchInvoker | type PDFDispatchInvoker interface { | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:253 | var | result | var result SignoffResult | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:254 | var | shouldDispatchPDF | var shouldDispatchPDF bool | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:255 | var | pdfTenantID | var pdfTenantID string | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:256 | var | pdfRevisionID | var pdfRevisionID string | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:28 | type | PDFOutboxEnqueuer | type PDFOutboxEnqueuer interface { | PDFOutboxEnqueuer enqueues a PDF dispatch inside the approval transaction. |
| internal/modules/documents/approval/application/decision_service.go:33 | type | DecisionService | type DecisionService struct { | DecisionService handles approver approve/reject decisions. |
| internal/modules/documents/approval/application/decision_service.go:43 | func | NewDecisionService | func NewDecisionService( | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:448 | func | scanSignoffs | func scanSignoffs(rows *sql.Rows) ([]domain.Signoff, error) { | scanSignoffs reads rows into domain.Signoff slice. |
| internal/modules/documents/approval/application/decision_service.go:449 | var | signoffs | var signoffs []domain.Signoff | (undocumented) |
| internal/modules/documents/approval/application/decision_service.go:493 | func | splitSignoffs | func splitSignoffs(all []domain.Signoff) (approvals, rejections []domain.Signoff) { | splitSignoffs partitions a slice of Signoff into approvals and rejections. |
| internal/modules/documents/approval/application/decision_service.go:507 | func | marshalSignaturePayload | func marshalSignaturePayload(payload map[string]any) (json.RawMessage, error) { | marshalSignaturePayload converts the map to json.RawMessage. |
| internal/modules/documents/approval/application/decision_service.go:66 | type | SignoffRequest | type SignoffRequest struct { | SignoffRequest carries all inputs for RecordSignoff. |
| internal/modules/documents/approval/application/decision_service.go:80 | type | SignoffResult | type SignoffResult struct { | SignoffResult is returned by RecordSignoff. |
| internal/modules/documents/approval/application/events.go:11 | type | GovernanceEvent | type GovernanceEvent struct { | GovernanceEvent mirrors the governance_events table columns. |
| internal/modules/documents/approval/application/events.go:24 | type | EventEmitter | type EventEmitter interface { | EventEmitter writes governance events within the caller's transaction. |
| internal/modules/documents/approval/application/events.go:29 | type | sqlEmitter | type sqlEmitter struct{} | sqlEmitter is the default production implementation. |
| internal/modules/documents/approval/application/events.go:32 | func | NewSQLEmitter | func NewSQLEmitter() EventEmitter { return &sqlEmitter{} } | NewSQLEmitter returns the production event emitter. |
| internal/modules/documents/approval/application/events.go:34 | const | insertEventSQL | const insertEventSQL = ` | (undocumented) |
| (truncated) | note | exports | domain/repository/jobs/root partially omitted to keep <400 lines | 367 rows omitted |

## 3) HTTP operations
Mount context from API main:
- apps/api/cmd/metaldocs-api/main.go:318: docMod := documents.New(docDeps)
- apps/api/cmd/metaldocs-api/main.go:319: docMod.RegisterRoutes(mux)
- apps/api/cmd/metaldocs-api/main.go:332: approvalHandler.RegisterRoutes(mux)

OperationIDs from internal/modules/documents/api/api.gen.go:
- (unclear: none found)

| Method | Path | Handler symbol | Source file:line |
|---|---|---|---|
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1487 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1531 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1532 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1533 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1534 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1535 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1536 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1537 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1538 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1539 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1540 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1541 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1542 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1543 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1544 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1545 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1546 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1547 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1548 |
| HANDLEFUNC | (unclear: parse failed) | HandleFunc(...) | internal/modules/documents/api/api.gen.go:1549 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/cancel_handler.go:27 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/cancel_handler.go:33 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/contracts/strictjson.go:28 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/doc_approval_handler.go:151 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/doc_approval_handler.go:56 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/handler.go:82 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/handler.go:93 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/obsolete_handler.go:27 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/obsolete_handler.go:33 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/publish_handler.go:36 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/publish_handler.go:42 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/publish_handler.go:75 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/publish_handler.go:81 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/route_admin_handler.go:113 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/route_admin_handler.go:118 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/route_admin_handler.go:20 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/route_admin_handler.go:64 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/route_admin_handler.go:69 |
| HANDLEFUNC | POST /api/v1/documents/{id}/publish | h.PublishHandler | internal/modules/documents/approval/http/router.go:10 |
| HANDLEFUNC | POST /api/v1/documents/{id}/schedule-publish | h.SchedulePublishHandler | internal/modules/documents/approval/http/router.go:11 |
| HANDLEFUNC | POST /api/v1/documents/{id}/supersede | h.SupersedeHandler | internal/modules/documents/approval/http/router.go:12 |
| HANDLEFUNC | POST /api/v1/documents/{id}/obsolete | h.ObsoleteHandler | internal/modules/documents/approval/http/router.go:13 |
| HANDLEFUNC | POST /api/v1/approval/instances/{instance_id}/cancel | h.CancelHandler | internal/modules/documents/approval/http/router.go:14 |
| HANDLEFUNC | GET /api/v1/approval/instances/{instance_id} | h.GetInstanceHandler | internal/modules/documents/approval/http/router.go:17 |
| HANDLEFUNC | GET /api/v1/documents/{id}/approval-instance | h.GetInstanceByDocumentHandler | internal/modules/documents/approval/http/router.go:18 |
| HANDLEFUNC | GET /api/v1/approval/inbox | h.InboxHandler | internal/modules/documents/approval/http/router.go:19 |
| HANDLEFUNC | POST /api/v1/documents/{id}/signoff | h.SignoffByDocumentHandler | internal/modules/documents/approval/http/router.go:22 |
| HANDLEFUNC | POST /api/v1/documents/{id}/cancel | h.CancelByDocumentHandler | internal/modules/documents/approval/http/router.go:23 |
| HANDLEFUNC | POST /api/v1/approval/routes | h.CreateRouteHandler | internal/modules/documents/approval/http/router.go:26 |
| HANDLEFUNC | PUT /api/v1/approval/routes/{id} | h.UpdateRouteHandler | internal/modules/documents/approval/http/router.go:27 |
| HANDLEFUNC | DELETE /api/v1/approval/routes/{id} | h.DeactivateRouteHandler | internal/modules/documents/approval/http/router.go:28 |
| HANDLEFUNC | GET /api/v1/approval/routes | h.ListRoutesHandler | internal/modules/documents/approval/http/router.go:29 |
| HANDLEFUNC | POST /api/v1/documents/{id}/submit | h.SubmitHandler | internal/modules/documents/approval/http/router.go:8 |
| HANDLEFUNC | POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/signoffs | h.SignoffHandler | internal/modules/documents/approval/http/router.go:9 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/signoff_handler.go:24 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/signoff_handler.go:30 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/submit_handler.go:19 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/submit_handler.go:25 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/supersede_handler.go:27 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/approval/http/supersede_handler.go:33 |
| HANDLEFUNC | POST /api/v1/documents/{id}/export/pdf | h.exportPDF | internal/modules/documents/delivery/http/export_handler.go:40 |
| HANDLEFUNC | GET /api/v1/documents/{id}/export/docx-url | h.exportDocxURL | internal/modules/documents/delivery/http/export_handler.go:41 |
| HANDLEFUNC | GET /api/v1/documents/{id}/export/docx-url | h.exportDocxURL | internal/modules/documents/delivery/http/export_handler.go:49 |
| HANDLEFUNC | POST /api/v1/documents/{id}/checkpoints | h.createCheckpoint | internal/modules/documents/delivery/http/handler.go:100 |
| HANDLEFUNC | POST /api/v1/documents/{id}/checkpoints/{version}/restore | h.restoreCheckpoint | internal/modules/documents/delivery/http/handler.go:101 |
| HANDLEFUNC | GET /api/v1/documents/{id}/revisions/{rid}/url | h.signedRevisionURL | internal/modules/documents/delivery/http/handler.go:103 |
| HANDLEFUNC | GET /api/v1/documents/{id}/comments | h.listComments | internal/modules/documents/delivery/http/handler.go:104 |
| HANDLEFUNC | POST /api/v1/documents/{id}/comments | h.createComment | internal/modules/documents/delivery/http/handler.go:105 |
| HANDLEFUNC | PATCH /api/v1/documents/{id}/comments/{libraryID} | h.updateComment | internal/modules/documents/delivery/http/handler.go:106 |
| HANDLEFUNC | DELETE /api/v1/documents/{id}/comments/{libraryID} | h.deleteComment | internal/modules/documents/delivery/http/handler.go:107 |
| HANDLEFUNC | GET /api/v1/documents | h.listDocuments | internal/modules/documents/delivery/http/handler.go:111 |
| HANDLEFUNC | GET /api/v1/documents/stats | h.documentStats | internal/modules/documents/delivery/http/handler.go:112 |
| HANDLEFUNC | GET /api/v1/documents/{id} | h.getDocument | internal/modules/documents/delivery/http/handler.go:114 |
| HANDLEFUNC | PATCH /api/v1/documents/{id} | h.renameDocument | internal/modules/documents/delivery/http/handler.go:115 |
| HANDLEFUNC | POST /api/v1/documents/{id}/finalize | h.finalizeDocument | internal/modules/documents/delivery/http/handler.go:116 |
| HANDLEFUNC | POST /api/v1/documents/{id}/archive | h.archiveDocument | internal/modules/documents/delivery/http/handler.go:117 |
| HANDLEFUNC | POST /api/v1/documents/{id}/duplicate | h.duplicateDocument | internal/modules/documents/delivery/http/handler.go:118 |
| HANDLEFUNC | POST /api/v1/documents/{id}/session/acquire | h.acquireSession | internal/modules/documents/delivery/http/handler.go:120 |
| HANDLEFUNC | POST /api/v1/documents/{id}/session/heartbeat | h.heartbeatSession | internal/modules/documents/delivery/http/handler.go:121 |
| HANDLEFUNC | POST /api/v1/documents/{id}/session/release | h.releaseSession | internal/modules/documents/delivery/http/handler.go:122 |
| HANDLEFUNC | POST /api/v1/documents/{id}/session/force-release | h.forceReleaseSession | internal/modules/documents/delivery/http/handler.go:123 |
| HANDLEFUNC | GET /api/v1/documents/{id}/checkpoints | h.listCheckpoints | internal/modules/documents/delivery/http/handler.go:134 |
| HANDLEFUNC | POST /api/v1/documents/{id}/checkpoints | h.createCheckpoint | internal/modules/documents/delivery/http/handler.go:135 |
| HANDLEFUNC | POST /api/v1/documents/{id}/checkpoints/{version}/restore | h.restoreCheckpoint | internal/modules/documents/delivery/http/handler.go:136 |
| HANDLEFUNC | GET /api/v1/documents/{id}/revisions/{rid}/url | h.signedRevisionURL | internal/modules/documents/delivery/http/handler.go:138 |
| HANDLEFUNC | GET /api/v1/documents/{id}/comments | h.listComments | internal/modules/documents/delivery/http/handler.go:139 |
| HANDLEFUNC | POST /api/v1/documents/{id}/comments | h.createComment | internal/modules/documents/delivery/http/handler.go:140 |
| HANDLEFUNC | PATCH /api/v1/documents/{id}/comments/{libraryID} | h.updateComment | internal/modules/documents/delivery/http/handler.go:141 |
| HANDLEFUNC | DELETE /api/v1/documents/{id}/comments/{libraryID} | h.deleteComment | internal/modules/documents/delivery/http/handler.go:142 |
| HANDLEFUNC | GET /api/v1/documents | h.listDocuments | internal/modules/documents/delivery/http/handler.go:82 |
| HANDLEFUNC | GET /api/v1/documents/stats | h.documentStats | internal/modules/documents/delivery/http/handler.go:83 |
| HANDLEFUNC | GET /api/v1/documents/{id} | h.getDocument | internal/modules/documents/delivery/http/handler.go:85 |
| HANDLEFUNC | PATCH /api/v1/documents/{id} | h.renameDocument | internal/modules/documents/delivery/http/handler.go:86 |
| HANDLEFUNC | POST /api/v1/documents/{id}/finalize | h.finalizeDocument | internal/modules/documents/delivery/http/handler.go:87 |
| HANDLEFUNC | POST /api/v1/documents/{id}/archive | h.archiveDocument | internal/modules/documents/delivery/http/handler.go:88 |
| HANDLEFUNC | POST /api/v1/documents/{id}/duplicate | h.duplicateDocument | internal/modules/documents/delivery/http/handler.go:89 |
| GET | (unclear: parse failed) | (unclear: parse failed) | internal/modules/documents/delivery/http/handler.go:899 |

## 4) Migration list
| Filename | Verb | Tables touched |
|---|---|---|
| 0001_init_documents.sql | CREATE | (unclear: not in first 3 lines) |
| 0006_grant_documents_bootstrap_privileges.sql | INSERT | metaldocs.document_versions, metaldocs.documents |
| 0009_init_document_types_and_metadata.sql | CREATE | IF |
| 0010_grant_document_types_privileges.sql | INSERT | metaldocs.document_types, metaldocs.documents, ON |
| 0011_init_document_access_policies.sql | CREATE | IF |
| 0012_grant_document_access_policies_privileges.sql | INSERT | metaldocs.document_access_policies |
