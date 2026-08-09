package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvalhttp "metaldocs/internal/modules/approval/http"
	approvalrepo "metaldocs/internal/modules/approval/infrastructure"
	approvalinfra "metaldocs/internal/modules/approval/infrastructure/idempotency"
	approvaljobs "metaldocs/internal/modules/approval/jobs"
	auditdomain "metaldocs/internal/modules/audit/domain"
	documents "metaldocs/internal/modules/documents"
	docapp "metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/documents/jobs"
	"metaldocs/internal/modules/jobs/maintenance"
	templatesapp "metaldocs/internal/modules/templates/application"
	templateshttp "metaldocs/internal/modules/templates/delivery/http"
	templatejobs "metaldocs/internal/modules/templates/jobs"

	"metaldocs/apps/api/internal/wiring"
	"metaldocs/internal/composition/tenantdata/registry"
	auditapp "metaldocs/internal/modules/audit/application"
	auditdelivery "metaldocs/internal/modules/audit/delivery/http"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	authapp "metaldocs/internal/modules/auth/application"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	controlleddocuments "metaldocs/internal/modules/controlleddocuments"
	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	distributionhttp "metaldocs/internal/modules/distribution/delivery/http"
	distributioninfra "metaldocs/internal/modules/distribution/infrastructure"
	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	iamjobs "metaldocs/internal/modules/iam/jobs"
	iampresence "metaldocs/internal/modules/iam/presence"
	notificationshttp "metaldocs/internal/modules/notifications/delivery/http"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/fanout/dispatchjobs"
	"metaldocs/internal/modules/render/fanout/retention"
	"metaldocs/internal/modules/render/resolvers"
	searchapp "metaldocs/internal/modules/search/application"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	searchdocs "metaldocs/internal/modules/search/infrastructure/v2documents"
	securityapp "metaldocs/internal/modules/security/application"
	securitydelivery "metaldocs/internal/modules/security/delivery/http"
	securitydomain "metaldocs/internal/modules/security/domain"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	"metaldocs/internal/modules/taxonomy"
	taxonomyapp "metaldocs/internal/modules/taxonomy/application"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	templatesinfra "metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/modules/tokens"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db"
	postgres "metaldocs/internal/platform/db/postgres"
	docgenv2 "metaldocs/internal/platform/docgenv2"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/formval"
	"metaldocs/internal/platform/httpclient"
	"metaldocs/internal/platform/httprouter"
	riverjobs "metaldocs/internal/platform/jobs/river"
	platformmw "metaldocs/internal/platform/middleware"
	"metaldocs/internal/platform/migrate"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/ratelimit"
	"metaldocs/internal/platform/security"
	e2etest "metaldocs/internal/test"
)

type fanoutComponents struct {
	client        *fanout.Client
	freezeService *docapp.FreezeService
}

// tenantCryptoKeyProvisioner adapts security's published TenantCrypto port
// to iam's TenantKeyProvisioner seam (M7 F7.3 Task B replaces F7.2's
// NoopTenantKeyProvisioner). Lives at the composition root, not inside
// iam or security, so neither module imports the other's internals.
type tenantCryptoKeyProvisioner struct {
	crypto securitydomain.TenantCrypto
}

func (p tenantCryptoKeyProvisioner) ProvisionTenantKey(ctx context.Context, tx *sql.Tx, tenantID string) error {
	return p.crypto.ProvisionTenantKeyTx(ctx, tx, tenantID)
}

// auditPayloadCryptoAdapter adapts security's published TenantCrypto port to
// the audit module's own narrow auditpg.PayloadCrypto port (M7 F7.3 item 2).
// Lives at the composition root so audit never imports security: it maps
// securitydomain's ErrKeyNotFound/ErrKeyDestroyed to auditpg's
// (_, encrypted=false, nil) fall-through-to-plaintext contract, and
// propagates any other (genuine, unexpected) error unchanged.
type auditPayloadCryptoAdapter struct {
	crypto securitydomain.TenantCrypto
}

func (a auditPayloadCryptoAdapter) EncryptForTenant(ctx context.Context, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenant(ctx, tenantID, plaintext)
	if err != nil {
		if errors.Is(err, securitydomain.ErrKeyNotFound) || errors.Is(err, securitydomain.ErrKeyDestroyed) {
			return "", false, nil
		}
		return "", false, err
	}
	return envelope, true, nil
}

// EncryptForTenantTx is the tx-aware variant RecordTx uses whenever it holds
// a live *sql.Tx (always, in practice) — it delegates to
// TenantCrypto.EncryptForTenantTx so the DEK lookup reads through the SAME
// transaction as the audit INSERT, closing the same-tx key-visibility gap
// (F7.3 defect: a tenant_keys row inserted earlier in that tx, still
// uncommitted, was invisible to a pool read, so onboarding's
// tenant.onboarded event always landed plaintext). Same sentinel-error
// mapping as EncryptForTenant.
func (a auditPayloadCryptoAdapter) EncryptForTenantTx(ctx context.Context, tx *sql.Tx, tenantID string, plaintext []byte) (string, bool, error) {
	envelope, err := a.crypto.EncryptForTenantTx(ctx, tx, tenantID, plaintext)
	if err != nil {
		if errors.Is(err, securitydomain.ErrKeyNotFound) || errors.Is(err, securitydomain.ErrKeyDestroyed) {
			return "", false, nil
		}
		return "", false, err
	}
	return envelope, true, nil
}

func (a auditPayloadCryptoAdapter) DecryptForTenant(ctx context.Context, tenantID, envelope string) ([]byte, error) {
	return a.crypto.DecryptForTenant(ctx, tenantID, envelope)
}

// e2eHandlersEnabled gates the test-only seed/reset/governance endpoints
// behind an explicit env flag. Defense-in-depth: internal/test/e2e_seed.go
// also early-returns on the same flag, but the call site must not register
// these routes either — permissions.go does not enumerate /internal/test/*,
// so any accidental mount is treated as fully public by newPublicPathChecker.
func e2eHandlersEnabled() bool {
	return e2etest.E2EEnabled()
}

//go:generate go run metaldocs/cmd/gen-http-surface -public ../../../../api/openapi/v1/openapi.yaml -e2e ../../../../api/openapi/internal-e2e.yaml -out-dir . -registry ../../../../internal/modules/iam/domain/model.go

func main() {
	os.Exit(run())
}

// run wires and serves the API process; it is main()'s former body,
// extracted so main can call os.Exit exactly once, after run returns. Every
// early-exit branch below now `return`s an exit code instead of calling
// os.Exit directly, so every defer registered along the way (deps.Cleanup,
// stop, otelShutdown, rlStore.Close, the background sweepers, ...) always
// unwinds — os.Exit used to skip them on every early-exit path (gocritic
// exitAfterDefer).
//
// run() is a composition root: it wires every module's dependencies in
// sequence. Each conditionally-constructed handler/service block below has
// been extracted into a cohesive, domain-named helper (buildXxx/loadXxx/
// wireXxx), mirroring the existing buildTaxonomyModule/buildTokensModule/
// buildFanoutComponents/startPresence convention — run() itself is now the
// sequencing of those helpers, not the inline logic.
func run() int {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelShutdownFn, otelEnabled, err := setupOTelTracing(ctx)
	if err != nil {
		return 1
	}
	defer otelShutdownFn()

	repoMode, corsCfg, attachmentsCfg, authCfg, featureFlagsCfg, err := loadBootConfig()
	if err != nil {
		return 1
	}

	deps, err := bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)
	if err != nil {
		slog.Error("build api dependencies", "err", err)
		return 1
	}
	defer deps.Cleanup()

	if err := applyStartupMigrations(ctx, deps); err != nil {
		return 1
	}

	authService, err := buildAuthService(ctx, deps, authCfg)
	if err != nil {
		return 1
	}

	// M7 F7.3: tenant DEK/KEK crypto-shred framework. tenantCrypto is nil when
	// METALDOCS_TENANT_KEK is unset — every consumer below falls back to its
	// own no-op path (F7.2's nil-safe pattern), so boot still works with
	// crypto disabled. Constructed here (before audit wiring) so the audit
	// writer below can be wired with the payload encryptor at boot.
	tenantCrypto, err := buildTenantCrypto(deps)
	if err != nil {
		return 1
	}

	// M7 F7.3 item 2: wire the audit payload envelope encryptor. deps.AuditWriter
	// is the same *auditpg.Writer instance backing AuditReader/AuditCounter (see
	// bootstrap.BuildAPIDependencies), so wiring it here via WithPayloadCrypto
	// (mutate-and-return-self) takes effect for every consumer of those three
	// fields. Nil tenantCrypto (KEK unset) leaves the writer's crypto nil —
	// RecordTx/ListEvents stay on the legacy plaintext path byte-for-byte.
	wireAuditPayloadCrypto(deps, tenantCrypto)

	auditService := buildAuditService(deps)

	auditHandler := auditdelivery.NewHandler(auditService).WithExporter(auditService)
	// ADR 0022 Phase 11 (F8): wire the tier-2 bypass audit sink so every
	// system_admin short-circuit and every background BypassSystem invocation is
	// recorded into the same audit pipe (audit.read surface). Set once, before serving.
	authz.SetBypassAuditSink(wiring.NewBypassAuditSink(deps.AuditWriter))
	// ADR 0038: search resolves document family_code through the taxonomy-owned
	// FamilyCodeResolver port instead of a raw metaldocs.document_profiles subquery.
	familyCodeResolver := taxonomyinfra.NewFamilyCodeResolverRepository(deps.SQLDB)
	searchService := searchapp.NewService(searchdocs.NewReader(deps.SQLDB, familyCodeResolver))
	searchHandler := searchdelivery.NewHandler(searchService)
	authHandler := authdelivery.NewHandler(authService).WithAudit(deps.AuditWriter)

	// mux is populated by one consolidated buildPublishers call (publishers.go)
	// once every handler below is constructed — not by scattered RegisterRoutes
	// call sites. Each publisher is mounted through its own httprouter.Recorder,
	// and the aggregate result is fed to assertSurface (surface.go), which
	// refuses to boot on any mismatch with the spec-generated surface table —
	// so a handler built here but never added to publisherDeps is a fatal at
	// startup, not a live discovery. iamAdminHandler, sessionsHandler,
	// observabilityHandler, securityHandler: mounted via buildPublishers below,
	// once peopleHandler/membershipHandler/rolesCapsHandler/presenceHandler are
	// also constructed.
	//
	// Declared here (rather than at its historical call site further down) so
	// permResolver below can close over it: newPermissionResolver reads
	// mux.Handler(r) per request, long after buildPublishers has populated it —
	// this declaration only needs the pointer to exist yet, not the routes.
	mux := http.NewServeMux()

	// surface is declared here, beside mux, for the identical reason: the
	// resolver closures below need to read it per-request through a pointer,
	// long before the E2E decision further down (main.go:~858) knows whether
	// to widen it. It starts as the generated httpSurface and — only if
	// useE2E — is reassigned (never shadowed) to mergedSurface(httpSurface,
	// httpSurfaceE2E) below, before buildPublishers ever mounts the first
	// route and before any request is served. Passing this map by value to
	// the resolvers would freeze it at construction time, before that
	// widening happens — they take *map[string]surfaceRule instead so they
	// keep reading through this same variable after it is reassigned.
	surface := httpSurface

	capabilityService := iamapp.NewCapabilityService(deps.SQLDB)
	// Wire optional capability hint into /auth/me + login responses.
	// Backend remains sole authz enforcer — FE consumes for UX hints only.
	authService.WithCapabilityProvider(capabilityService)
	cachedProvider := iamapp.NewCachedRoleProvider(ctx, deps.RoleProvider, authn.CacheTTL())
	// permResolver is the single authoritative source of truth for route
	// visibility: a mux-pattern lookup into the table `surface` points to —
	// the generated httpSurface table (produced from the OpenAPI spec by
	// cmd/gen-http-surface), possibly widened to include httpSurfaceE2E
	// below. It is shared with the auth middleware so that fully public
	// routes (no session required), the password-change-allowed exceptions,
	// and the IAM permission layer all read the same table — adding a new
	// public route requires a spec annotation, not a change in three places.
	// mux and surface are both declared above (mux historically at line
	// ~501; surface next to it) so the resolver can consult them per-request;
	// mux is only populated later by buildPublishers, and surface may still
	// be widened by the E2E decision below — both long after this closure is
	// constructed.
	permResolver := newPermissionResolver(mux, &surface)
	authMiddleware := authdelivery.NewMiddleware(authService, authCfg, authn.Enabled()).
		WithPublicPathChecker(newPublicPathChecker(permResolver)).
		WithPasswordChangeAllowedChecker(newPasswordChangeAllowedChecker(mux, &surface))
	iamMiddleware := iamdelivery.NewMiddleware(capabilityService, cachedProvider, authn.Enabled()).
		WithPermissionResolver(permResolver)
	originProtection := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           authCfg.OriginProtection,
		SessionCookieName: authCfg.SessionCookieName,
		TrustedOrigins:    authCfg.TrustedOrigins,
		TrustedProxyCIDRs: authCfg.TrustedProxyCIDRs,
	})

	// H-3b: TxRunner for IAM atomic audit writes (Site 2, 3, 4).
	iamTxRunner := buildIAMTxRunner(deps)
	iamAdminService := iamapp.NewAdminService(deps.RoleAdminRepo, cachedProvider, iamTxRunner, deps.AuditWriter)
	iamAdminHandler := iamdelivery.NewAdminHandler(iamAdminService, authService, deps.AuditWriter).
		WithAuditEventLister(auditService)

	// M7 F7.2: tenant onboarding (POST /tenants). tenantHandler is nil (Router
	// answers 501) on the SQLDB-less boot path, matching every other
	// conditionally-wired iam handler above/below.
	tenantHandler := buildTenantHandler(deps, iamTxRunner, tenantCrypto)

	// PR-7 Sessions & Security tab.
	sessionsHandler := buildSessionsHandler(deps)

	securityService, securityHandler := buildSecurityHandler(deps)

	// PR-8 Observability service.
	observabilityHandler := buildObservabilityHandler(deps, securityService, iamAdminHandler)

	featureFlagsHandler := featureflags.NewHandler(featureFlagsCfg)
	httpObs := observability.NewHTTPObservability(currentUserIDFromRequest, deps.StatusProvider)
	// healthHandler/observabilityHandlerMetrics are constructed after httpObs
	// (not at healthHandler's previous call site near authHandler) because
	// GetMetrics delegates to httpObs.MetricsHandler(). Task 15a (Ruling 1)
	// split the former combined HealthHandler into two SurfacePublishers —
	// one per OpenAPI tag ("health", "observability") — mounting two
	// generated packages (healthapi, observabilityapi) instead of one.
	healthHandler := observability.NewHealthHandler(deps.StatusProvider)
	observabilityHandlerMetrics := observability.NewMetricsHandler(httpObs.MetricsHandler())
	cors := security.NewCORS(corsCfg)

	// Rate-limit store backend selection (M8/F8.2): memory (default,
	// single-replica) or a shared Redis store so N api replicas enforce ONE
	// combined budget instead of N independent ones. The startup guard
	// refuses to boot when METALDOCS_MULTI_REPLICA=true with the in-memory
	// store (per-process counters would silently multiply the limits ×N).
	// nil for the memory backend — each limiter then builds its own private
	// in-memory store (unchanged behavior). Non-nil (redis) is shared by BOTH
	// limiter mounts below so the process holds exactly one Redis client.
	rlStore, rlCleanup, preAuthLimiter, err := buildRateLimiting(ctx, authCfg)
	if err != nil {
		return 1
	}
	defer rlCleanup()

	// Post-authn global envelope limiter (F-05/D-04, Wave 2.8): replaces the
	// legacy security.RateLimiter. Same identity precedence: authenticated user
	// → trusted-proxy-resolved client IP → fail-closed. Default: 120 req/min
	// per identity (matches old fixed-window default of 120 req / 60s window).
	// Env vars METALDOCS_RATE_LIMIT_ENABLED / _WINDOW_SECONDS / _MAX_REQUESTS
	// are now dead — see commit body for mapping.
	globalRateCfg := ratelimit.DefaultConfig()
	globalRateCfg.TrustedProxyCIDRs = authCfg.TrustedProxyCIDRs
	globalRateCfg.Store = rlStore // shared with preAuthLimiter (one Redis client when redis)
	globalLimiter := ratelimit.New(ctx, globalRateCfg)
	// userIDExtractor resolves the authenticated principal from context. Runs
	// after authn + iamMiddleware in the chain, so both auth and IAM user IDs
	// are available. Mirrors security.RateLimiter.requestIdentity without
	// importing domain packages (dependency injected via closure).
	userIDExtractor := resolveRateLimitIdentity

	presenceBump, presenceHub, presenceHandler := startPresence(ctx, deps, iamAdminHandler)

	taxonomyModule := buildTaxonomyModule(deps)

	tokensModule := buildTokensModule(deps)

	controlledDocumentsModule := buildControlledDocumentsModule(deps)
	controlledDocumentDuplicator := wiring.NewControlledDocumentDuplicator(controlledDocumentsModule.Service())

	peopleHandler, membershipHandler, rolesCapsHandler := buildPeopleHandlers(deps, authService, cachedProvider)

	// CON-07 (ADR 0012 / target-arch N2): mount the full generated IAM
	// ServerInterface — iamAdminHandler, sessionsHandler, observabilityHandler,
	// peopleHandler, membershipHandler, rolesCapsHandler, and presenceHandler's
	// HTTP snapshot route — in one call, replacing six independent
	// RegisterRoutes(mux) sites. Route shapes on the wire are unchanged (same
	// BaseURL "/api/v1" + spec-declared paths); tier-1 authz keys off
	// r.Method/r.URL.Path (permissions.go), not mux dispatch mechanics, so this
	// swap changes no auth behavior.
	iamRouter := iamdelivery.NewRouter(iamAdminHandler, peopleHandler, membershipHandler, rolesCapsHandler, sessionsHandler, observabilityHandler, presenceHandler).
		WithTenantHandler(tenantHandler)

	// Legacy templates module routes removed — templates owns /api/v1/templates/*

	// STO-02: single shared VerifiedStore instance for the composition root. The
	// documents and templates modules previously each constructed their own
	// NewVerifiedStore with identical args (same minio clients, same bucket, same
	// 25 MiB cap) — two kernel instances wrapping the same underlying bucket with
	// no shared state to justify separate objects. Construct once here and inject
	// into both consumers (docPresigner below; templatesPresigner via
	// buildTemplatesModule's presigner parameter).
	sharedPresigner := objectstore.NewVerifiedStore(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 25*1024*1024)
	docPresigner := sharedPresigner
	profileRepo := taxonomyinfra.NewProfileRepository(deps.SQLDB)

	// Fanout/eigenpal client — enabled when METALDOCS_FANOUT_URL is set.
	fanoutClientCfg, err := loadFanoutClientConfig()
	if err != nil {
		return 1
	}
	fanoutCfg := buildFanoutComponents(deps, fanoutClientCfg, controlledDocumentsModule)

	// M4/F4.1: construct the iam-owned display-name port once and inject into all
	// three consumers (approval repo, documents repo, approval handler).
	displayNameRepo := iampg.NewUserDisplayNameRepository(deps.SQLDB)

	docSnapshotReader := docgenv2.NewTemplatesSnapshotReader(deps.SQLDB)
	docDeps := documents.Dependencies{
		DB:      deps.SQLDB,
		Presign: docPresigner,
		// ARC-01/DB-01: legacy public.templates/template_versions fallback
		// reader removed (2026-07-03) after a full QA window proved zero
		// legacy-fallback reads; templates_template_version is now the only
		// read path.
		TplRead:                      docgenv2.NewTemplatesTemplateReader(deps.SQLDB),
		FormVal:                      formval.NewGojsonschema(),
		Audit:                        wiring.NewDocumentsAuditSink(deps.AuditWriter),
		ExportPresign:                docPresigner,
		ViewPresign:                  docPresigner,
		ControlledDocumentDuplicator: controlledDocumentDuplicator,
		Caps:                         wiring.NewCapabilityChecker(capabilityService),
		ProfileDefaults:              wiring.NewProfileDefaults(profileRepo, templatesinfra.NewTemplateVersionReader(deps.SQLDB)),
		SnapshotReader:               docSnapshotReader,
		DisplayNameReader:            displayNameRepo,
		IAMUserOptions:               wiring.NewDocumentsIAMUserOptions(authService),
	}
	// cdReader is the controlleddocuments-owned read-port for controlled_documents
	// fields (M2/F2.1; ADR-0039 D3(b)). One stateless instance serves documents'
	// profile_code read and the area-grade authz checks in the approval services.
	cdReader := cdinfra.NewCDFieldReaderPG()
	docDeps.CDFieldReader = cdReader
	// areaCatalog is the taxonomy-owned read-port for document_process_areas
	// (M2/F2.3; ADR-0039 D3(b)). documents reads the area name through it in-tx.
	docDeps.AreaCatalogReader = taxonomyinfra.NewAreaCatalogReaderPG()
	// TemplateVersionPort is the templates-owned read-port (ADR-0030; extended
	// M2/F2.4): documents reads fill-in placeholder schema through it instead of
	// joining templates_template_version to its own table.
	docDeps.TemplateVersionPort = templatesinfra.NewTemplateVersionReader(deps.SQLDB)
	// PDFOutboxReader is the render/fanout-owned liveness read-port (QA-1 F13):
	// pdf_status='failed' derives from dead-lettered materialize/pdf outbox events.
	docDeps.PDFOutboxReader = fanout.NewPDFPipelineStateReader(deps.SQLDB)
	finishDocumentsDependencies(&docDeps, deps, fanoutCfg)

	approvalRepo := approvalrepo.NewPostgresApprovalRepository(deps.SQLDB, displayNameRepo)
	approvalEmitter := approvalapp.NewSQLEmitter()
	approvalServices := approvalapp.NewServices(approvalRepo, approvalEmitter, approvalapp.RealClock{}, cdReader)
	// G1: wire the approval→taxonomy profile-policy reader so route-admin and
	// submit enforce the per-profile route-signature policy friendly-first. The
	// adapter reads through the taxonomy profileRepo (own short tx, CapTaxonomyView,
	// H-PRE-1). The DB deferrable trigger remains the authoritative last line.
	approvalServices = approvalServices.WithProfilePolicyReader(approvalrepo.NewProfilePolicyReader(profileRepo))
	// M3 P3.S2b-3b-ii: wire the approval-owned TemplateVersionReader port to
	// the templates-side adapter. approval never imports templates
	// infrastructure beyond this narrow interface satisfaction.
	approvalServices = approvalServices.WithTemplateVersionReader(templatesinfra.NewApprovalVersionReader())
	// M3 P3.S2b-3b-iii-b: wire the approval-owned TemplateCompletionWriter
	// port to the templates-side adapter — the ONLY seam a terminal
	// template-subject signoff decision crosses into templates_template_version.
	approvalServices = approvalServices.WithTemplateCompletionWriter(templatesinfra.NewApprovalCompletionWriter())

	// wireApprovalRuntime wires the River job runtime (which completes
	// approvalServices.Decision's ports) and the templates handler (whose
	// approval-kernel wiring requires Decision to be Ready()) behind one
	// call/error-check pair — neither depends on the documents module, so
	// both run before it without changing observable behavior.
	templatesModule, templatesRepo, templatesStore, approvalHandler, err := wireApprovalRuntime(ctx, deps, approvalServices, tenantHandler, iamTxRunner, tenantCrypto, sharedPresigner, fanoutCfg, cdReader, capabilityService, displayNameRepo)
	if err != nil {
		return 1
	}

	docMod := documents.New(docDeps)
	// SP-2: pin tenant dictionary values at document creation. tokensModule was
	// built at startup (line ~358), before docMod.
	docMod.Service.WithDictionaryReader(dictionaryValueReaderAdapter{reader: tokensModule.Reader})
	// Task 15a: Mount(Muxer) takes exactly one argument, so the rate limiter
	// and user-ID extractor (both constructed at line ~484, well before
	// docMod exists) move from a routeFamilies closure onto docMod as
	// constructor fields.
	docMod.WithRateLimit(globalLimiter, userIDExtractor)

	// Wire the documents-side adapter back into the controlled-documents service so atomic
	// CD-create can clone the initial document inside the same tx as the CD
	// insert. controlledDocumentsModule was constructed before docMod (because docMod
	// needs ControlledDocumentDuplicator), hence the post-construction setter.
	controlledDocumentsModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))

	// M2/F2.2: distribution module — read-only delivery + repository layer.
	// Reuses displayNameRepo (constructed above at line ~418); no new DB handle.
	distributionRepo := distributioninfra.NewCoverageRepository(deps.SQLDB, displayNameRepo)
	distributionHandler := distributionhttp.NewHandler(distributionRepo)

	// M3/F3.2: notifications module — read surface (list / unread-count / mark-read).
	// Self-scoped by CapNotificationRead (tier-1) + recipient_user_id SQL predicate.
	notificationsRepo := notificationsinfra.NewNotificationsRepository(deps.SQLDB)
	notificationsHandler := notificationshttp.NewHandler(notificationsRepo)

	// buildPublishers (publishers.go) is the composition root's route
	// inventory — a LIST, not a struct: a handler built above but never added
	// to publisherDeps is a missing list element, and assertSurface's check 1
	// (tag coverage) fires as a boot fatal below, not a live discovery.
	e2e := e2ePublisher(deps.SQLDB) // nil in every build without the handlers
	useE2E := decideE2E(e2e)        // ONE value decides both sides

	publishers := buildPublishers(publisherDeps{
		auth:                authHandler,
		health:              healthHandler,
		observability:       observabilityHandlerMetrics,
		featureFlags:        featureFlagsHandler,
		audit:               auditHandler,
		search:              searchHandler,
		security:            securityHandler,
		taxonomy:            taxonomyModule,
		tokens:              tokensModule,
		controlledDocuments: controlledDocumentsModule,
		iam:                 iamRouter,
		documents:           docMod,
		templates:           templatesModule,
		approval:            approvalHandler,
		distribution:        distributionHandler,
		notifications:       notificationsHandler,
	})
	// surface AND expectedTags are BOTH derived from useE2E — never from the
	// publisher list — but surface is an ASSIGNMENT into the variable
	// declared beside mux above, not a fresh `:=`: newPermissionResolver and
	// newPasswordChangeAllowedChecker already hold a pointer to that
	// variable, so reassigning it here (rather than shadowing it) is what
	// makes the widened table visible to every request the resolvers serve
	// from this point forward.
	expectedTags := specTags
	if useE2E {
		publishers = append(publishers, e2e)
		surface = mergedSurface(httpSurface, httpSurfaceE2E)
		expectedTags = union(specTags, specTagsE2E)
		slog.Warn("e2etest handlers mounted — destructive endpoints reachable without auth", "env", "METALDOCS_E2E=1")
	} else {
		slog.Info("e2etest handlers not mounted", "reason", "METALDOCS_E2E != 1")
	}

	mounted := mountPublishers(mux, publishers)

	// surface AND expectedTags are BOTH derived from useE2E — never from the
	// publisher list, which is the thing being audited: an expected set
	// derived from the list under audit makes check 1 vacuously true, which
	// is exactly how a missing publisher would pass a check written to catch
	// it.
	if err := assertSurface(mounted, surface, expectedTags, publishers); err != nil {
		slog.Error("http surface", "err", err)
		return 1
	}

	wireHTTPObservabilityDBPool(httpObs, deps)

	stopSessions := jobs.StartSessionSweeper(ctx, docMod.Repo(), 60*time.Second)
	stopOrphans := jobs.StartOrphanPendingSweeper(ctx, docMod.Repo(), time.Hour, 24*time.Hour)
	// F-T6(a): reconciliation janitor for template objects orphaned when
	// spawnNextDraft's pre-tx Copy survives a rolled-back publish tx. Mirrors the
	// documents orphan sweeper above; deletes only aged-out (>24h) objects absent
	// from the referenced docx∪schema key set.
	stopTemplateOrphans := templatejobs.StartTemplateOrphanSweeper(ctx, templatesRepo, templatesStore, time.Hour, 24*time.Hour)
	defer stopSessions()
	defer stopOrphans()
	defer stopTemplateOrphans()
	// Prometheus text-exposition scrape endpoint. Deliberately NOT mounted on
	// mux (which the API chain below wraps with authn/iam/httpObs/rate-limit)
	// — it is served from a top-level dispatch mux, ahead of and outside the
	// entire API chain (see rootMux below, after handler is built). Contract
	// §3.2: /metrics is a platform scrape surface, not a versioned product
	// route, so it is NOT declared in api/openapi/v1/openapi.yaml. Coexists
	// with the JSON endpoint (mounted via buildPublishers above, see
	// publishers.go); they read from separate storage (prometheus vecs vs.
	// byKey atomics).

	// Audit retention - AUDIT_RETENTION_DAYS=0 disables (default disabled).
	retentionCfg, err := config.LoadRetentionConfig()
	if err != nil {
		slog.Error("invalid retention config", "err", err)
		return 1
	}
	startAuditRetention(ctx, deps, retentionCfg.Days)

	// Canonical chain per backend-target-architecture.md §2.1 (F-01 fix,
	// REQ-MW-1/2/4/5): panic recovery + trace context outermost, access
	// log/metrics OUTSIDE authn so 401s and panics are observable,
	// pre-auth IP-keyed login limit before authn. presenceBump stays
	// after iamMiddleware (needs iamdomain.UserID in ctx, PR-9). Order is
	// asserted by chain_test.go (REQ-MW-7).
	// otel link is nil (skipped by buildChain) unless an exporter is configured;
	// recovery stays outermost, otel wraps everything else (Z-1, REQ-OBS-1).
	handler := buildChain(mux, apiChain(
		platformmw.Recovery,
		buildOtelWrap(otelEnabled),
		httpObs.Wrap,
		cors.Wrap,
		originProtection.Wrap,
		loginRateLimit(preAuthLimiter),
		authMiddleware.Wrap,
		iamMiddleware.Wrap,
		buildPresenceWrap(presenceBump),
		func(next http.Handler) http.Handler { return globalLimiter.GlobalEnvelopeWrap(userIDExtractor, next) },
		// Innermost (nearest the mux): rewrite the stdlib text/plain 404/405 the
		// method-routed ServeMux emits into problem+json, preserving Allow (D-03).
		platformmw.MethodNotAllowedJSON,
	))

	server, metricsServer, err := buildServers(handler, httpObs)
	if err != nil {
		return 1
	}

	// Z-22 / REQ-REL-2: drain live WS presence connections before the
	// HTTP server stops accepting.
	registerPresenceShutdown(server, presenceHub)

	slog.Info("MetalDocs API listening",
		"addr", server.Addr, "metrics_addr", metricsServer.Addr,
		"repository", repoMode, "auth_enabled", authn.Enabled(),
		"auth_cache_ttl", authn.CacheTTL(), "cors_enabled", corsCfg.Enabled,
		"cors_allowed_origins", len(corsCfg.AllowedOrigins))

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	metricsErr := make(chan error, 1)
	go func() {
		metricsErr <- metricsServer.ListenAndServe()
	}()

	// Returning (rather than os.Exit-ing here) lets every defer registered
	// above — deps.Cleanup, stop, otelShutdown, rlStore.Close, the session/
	// orphan/template sweepers — unwind on both the clean and the failed
	// shutdown path; main's single os.Exit(run()) applies the exit code
	// after that unwind completes.
	return shutdownServer(ctx, stop, server, metricsServer, serverErr, metricsErr)
}

// setupOTelTracing wraps observability.SetupOTel with the deferred-shutdown
// closure and the "tracing enabled" log line run() previously inlined
// (Z-1, REQ-OBS-3). otelShutdown is a no-op when disabled; the returned
// bool gates the otel chain link (buildOtelWrap).
func setupOTelTracing(ctx context.Context) (func(), bool, error) {
	otelShutdown, otelEnabled, err := observability.SetupOTel(ctx, "metaldocs-api")
	if err != nil {
		slog.Error("setup otel", "err", err)
		return nil, false, err
	}
	if otelEnabled {
		slog.Info("OpenTelemetry tracing enabled", "exporter", os.Getenv("OTEL_TRACES_EXPORTER")) //nolint:gosec // G706: slog default is JSONHandler (set at process start) — control chars are JSON-escaped, log-line injection not possible
	}
	return wrapOTelShutdown(ctx, otelShutdown), otelEnabled, nil
}

// wrapOTelShutdown wraps the OTel SetupOTel shutdown func with the 5s
// shutdown timeout and warning log run() previously inlined into its own
// defer closure. parentCtx is run()'s signal-context, propagated here (rather
// than a bare context.Background()) so the shutdown deadline stays anchored
// to the process's own context chain; it is deliberately detached from
// cancellation via context.WithoutCancel, since the returned closure only
// runs from run()'s own deferred cleanup, at which point parentCtx is always
// already Done() (shutdownServer has already returned by then) — deriving
// the timeout from it directly would cancel instantly, defeating the 5s
// grace period below.
func wrapOTelShutdown(parentCtx context.Context, shutdown func(context.Context) error) func() {
	return func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(parentCtx), 5*time.Second)
		defer cancel()
		if err := shutdown(shutdownCtx); err != nil {
			slog.Warn("otel shutdown", "err", err)
		}
	}
}

// loadBootConfig loads the fixed set of process-lifetime config values run()
// needs before it can build bootstrap.APIDependencies: repository mode (with
// the metaldocs-api-requires-postgres guard), CORS, attachments, auth runtime
// config, and feature flags. Each load keeps its own "invalid X config" log
// line, unchanged from their previous inline call sites.
func loadBootConfig() (repoMode string, corsCfg config.CORSConfig, attachmentsCfg config.AttachmentsConfig, authCfg authapp.Config, featureFlagsCfg config.FeatureFlagsConfig, err error) {
	repoMode, err = config.RepositoryMode()
	if err != nil {
		slog.Error("invalid repository mode", "err", err)
		return
	}
	if err = requirePostgresRepositoryMode(repoMode); err != nil {
		slog.Error("unsupported repository mode", "err", err)
		return
	}
	corsCfg, err = config.LoadCORSConfig()
	if err != nil {
		slog.Error("invalid cors config", "err", err)
		return
	}
	attachmentsCfg, err = config.LoadAttachmentsConfig()
	if err != nil {
		slog.Error("invalid attachments config", "err", err)
		return
	}
	authCfg, err = authn.LoadRuntimeConfig()
	if err != nil {
		slog.Error("invalid auth config", "err", err)
		return
	}
	featureFlagsCfg, err = config.LoadFeatureFlagsConfig()
	if err != nil {
		slog.Error("invalid feature flags config", "err", err)
		return
	}
	return
}

// applyStartupMigrations applies the db/grants stage followed by the
// migration ledger, when a database connection is configured and migrations
// are not skipped. The grants stage runs BEFORE migrations and
// unconditionally (no ledger): db/grants carries the privilege/role posture
// pg_dump --no-privileges cannot put in the baseline, and it was previously
// applied only at fresh bootstrap — so an existing volume never saw an edit.
// Every file is idempotent by construction; a missing/empty dir is fatal.
// See internal/platform/migrate.ApplyGrants.
func applyStartupMigrations(ctx context.Context, deps bootstrap.APIDependencies) error {
	migrationCfg, err := config.LoadMigrationConfig()
	if err != nil {
		slog.Error("invalid migration config", "err", err)
		return err
	}
	if deps.SQLDB == nil || migrationCfg.Skip {
		return nil
	}
	if err := migrate.ApplyGrants(ctx, deps.SQLDB, migrationCfg.GrantsDir, slog.Default()); err != nil {
		slog.Error("apply grants stage", "err", err)
		return err
	}
	if err := migrate.Apply(ctx, deps.SQLDB, migrationCfg.Dir, slog.Default()); err != nil {
		slog.Error("apply startup migrations", "err", err)
		return err
	}
	return nil
}

// buildAuthService constructs the auth application service and runs its
// local-admin bootstrap.
func buildAuthService(ctx context.Context, deps bootstrap.APIDependencies, authCfg authapp.Config) (*authapp.Service, error) {
	authService, err := authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, iampg.NewLoginContextRepository(deps.SQLDB), authCfg, deps.AuditWriter)
	if err != nil {
		slog.Error("new auth service", "err", err)
		return nil, err
	}
	if err := authService.BootstrapLocalAdmin(ctx); err != nil {
		slog.Error("bootstrap local admin", "err", err)
		return nil, err
	}
	return authService, nil
}

// buildTenantCrypto constructs the M7 F7.3 tenant DEK/KEK crypto-shred
// service when a database connection is available and METALDOCS_TENANT_KEK
// is configured. Returns (nil, nil) in either the SQLDB-less boot path or
// when the KEK is unset — every consumer falls back to its own no-op path
// (F7.2's nil-safe pattern), so boot still works with crypto disabled.
func buildTenantCrypto(deps bootstrap.APIDependencies) (securitydomain.TenantCrypto, error) {
	if deps.SQLDB == nil {
		return nil, nil
	}
	kek, kekConfigured, kekErr := config.LoadTenantKEK()
	if kekErr != nil {
		slog.Error("invalid tenant crypto KEK", "err", kekErr)
		return nil, kekErr
	}
	if !kekConfigured {
		slog.Info("tenant crypto disabled: METALDOCS_TENANT_KEK not set")
		return nil, nil
	}
	svc, err := securityapp.NewTenantCryptoService(securitypg.NewTenantKeyRepository(deps.SQLDB), kek)
	if err != nil {
		slog.Error("construct tenant crypto service", "err", err)
		return nil, err
	}
	return svc, nil
}

// wireAuditPayloadCrypto wires the M7 F7.3 item 2 payload envelope encryptor
// into deps.AuditWriter when tenant crypto is enabled. deps.AuditWriter is
// the same *auditpg.Writer instance backing AuditReader/AuditCounter (see
// bootstrap.BuildAPIDependencies), so wiring it here via WithPayloadCrypto
// (mutate-and-return-self) takes effect for every consumer of those three
// fields. A nil tenantCrypto (KEK unset) is a no-op — RecordTx/ListEvents
// stay on the legacy plaintext path byte-for-byte.
func wireAuditPayloadCrypto(deps bootstrap.APIDependencies, tenantCrypto securitydomain.TenantCrypto) {
	if tenantCrypto == nil {
		return
	}
	if auditWriter, ok := deps.AuditWriter.(*auditpg.Writer); ok {
		auditWriter.WithPayloadCrypto(auditPayloadCryptoAdapter{crypto: tenantCrypto})
	}
}

// auditExportDownloadURL builds the export-download URL audit.Service.WithExports
// stamps onto a completed export job, or "" for a job missing its ID/token.
func auditExportDownloadURL(job auditdomain.ExportJob) string {
	if job.ID == "" || job.DownloadToken == "" {
		return ""
	}
	return fmt.Sprintf("/api/v1/audit/events/export/%s/download?token=%s", job.ID, job.DownloadToken)
}

// buildAuditService constructs the audit application service, wiring export
// support when both the counter and export repositories are available.
func buildAuditService(deps bootstrap.APIDependencies) *auditapp.Service {
	auditService := auditapp.NewService(deps.AuditReader)
	if deps.AuditCounter != nil && deps.AuditExports != nil {
		auditService.WithExports(deps.AuditCounter, deps.AuditExports, deps.AuditWriter, auditExportDownloadURL)
	}
	return auditService
}

// buildIAMTxRunner returns a TxRunner over deps.SQLDB for IAM's atomic audit
// writes (H-3b, Site 2/3/4), or nil in the SQLDB-less in-memory boot path.
func buildIAMTxRunner(deps bootstrap.APIDependencies) db.TxRunner {
	if deps.SQLDB == nil {
		return nil
	}
	return db.NewTxRunner(deps.SQLDB)
}

// buildTenantHandler wires the M7 F7.2 tenant onboarding handler (POST
// /tenants). Returns nil (Router answers 501) on the SQLDB-less boot path,
// matching every other conditionally-wired IAM handler.
func buildTenantHandler(deps bootstrap.APIDependencies, iamTxRunner db.TxRunner, tenantCrypto securitydomain.TenantCrypto) *iamdelivery.TenantHandler {
	if deps.SQLDB == nil {
		return nil
	}
	var keyProvisioner iamapp.TenantKeyProvisioner
	if tenantCrypto != nil {
		// Composition-root adapter: iam depends only on its own
		// TenantKeyProvisioner port; it never imports security's
		// internals. This wraps security's published TenantCrypto port
		// (F7.3 replaces F7.2's NoopTenantKeyProvisioner).
		keyProvisioner = tenantCryptoKeyProvisioner{crypto: tenantCrypto}
	}
	onboardTenantService := iamapp.NewOnboardTenantService(
		iampg.NewTenantRepository(deps.SQLDB),
		authpg.NewRepository(deps.SQLDB, iampg.NewUserTenantRepository(deps.SQLDB)),
		iamTxRunner,
		deps.AuditWriter,
		keyProvisioner, // nil -> NoopTenantKeyProvisioner when tenant crypto is disabled
		authapp.HashPassword,
	)
	return iamdelivery.NewTenantHandler(onboardTenantService)
}

// buildSessionsHandler wires the PR-7 Sessions & Security tab handler.
// Returns nil on the SQLDB-less boot path.
func buildSessionsHandler(deps bootstrap.APIDependencies) *iamdelivery.SessionsHandler {
	sqlDB := deps.SQLDB
	if sqlDB == nil {
		return nil
	}
	authRepo := authpg.NewRepository(sqlDB, iampg.NewUserTenantRepository(sqlDB))
	sessionSvc := iamapp.NewSessionService(db.NewTxRunner(sqlDB), deps.AuditWriter, authRepo)
	// M4/F4.4: auth returns auth-owned session rows; the iam consumer enriches
	// display names via the iam-owned port (read off-tx on the pool).
	return iamdelivery.NewSessionsHandler(authRepo, deps.AuditWriter).
		WithSessionService(sessionSvc).
		WithDisplayNameReader(iampg.NewUserDisplayNameRepository(sqlDB))
}

// buildSecurityHandler wires the security-report service. Security reports on
// auth_identities/auth_sessions but does not own iam_users or
// iam_user_roles; it resolves tenant membership, display names, and
// admin-role membership via iam-owned ports (M4/F4.2) instead of JOINing
// iam's tables. The returned handler is always non-nil (nil service on the
// SQLDB-less boot path, matching the original inline nil guard).
func buildSecurityHandler(deps bootstrap.APIDependencies) (*securityapp.Service, *securitydelivery.Handler) {
	sqlDB := deps.SQLDB
	if sqlDB == nil {
		return nil, securitydelivery.NewHandler(nil)
	}
	securityService := securityapp.NewService(securitypg.NewRepository(
		sqlDB,
		iampg.NewUserDisplayNameRepository(sqlDB),
		iampg.NewTenantUserRepository(sqlDB),
		iampg.NewAdminRoleMemberRepository(sqlDB),
		iampg.NewMfaUserRepository(sqlDB),
	))
	return securityService, securitydelivery.NewHandler(securityService)
}

// buildObservabilityHandler wires the PR-8 Observability service and mutates
// iamAdminHandler in place (WithObservabilityService is a mutate-and-
// return-self pointer method, so the caller's iamAdminHandler sees the
// wiring without needing the return value threaded back). Returns nil on the
// SQLDB-less boot path, leaving iamAdminHandler untouched.
func buildObservabilityHandler(deps bootstrap.APIDependencies, securityService *securityapp.Service, iamAdminHandler *iamdelivery.AdminHandler) *iamdelivery.ObservabilityHandler {
	sqlDB := deps.SQLDB
	if sqlDB == nil {
		return nil
	}
	observabilityRepo := iampg.NewObservabilityRepository(sqlDB)
	observabilityService := iamapp.NewObservabilityService(observabilityRepo, wiring.NewMfaCoveragePctReader(securityService))
	observabilityHandler := iamdelivery.NewObservabilityHandler(observabilityService)
	iamAdminHandler.WithObservabilityService(observabilityService)
	return observabilityHandler
}

// currentUserIDFromRequest resolves the authenticated principal's user ID
// for httpObs's per-request logging, or "" when unauthenticated.
func currentUserIDFromRequest(r *http.Request) string {
	if currentUser, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
		return currentUser.UserID
	}
	return ""
}

// resolveRateLimitIdentity resolves the global envelope limiter's identity
// key: authenticated user → IAM user ID → "" (fail-closed to the shared
// anonymous bucket). Runs after authn + iamMiddleware in the chain, so both
// are available. Mirrors the retired security.RateLimiter.requestIdentity
// without importing domain packages.
func resolveRateLimitIdentity(r *http.Request) string {
	if currentUser, ok := authdomain.CurrentUserFromContext(r.Context()); ok && strings.TrimSpace(currentUser.UserID) != "" {
		return strings.TrimSpace(currentUser.UserID)
	}
	if userID := strings.TrimSpace(iamdomain.UserIDFromContext(r.Context())); userID != "" {
		return userID
	}
	return ""
}

// buildRateLimitStore loads the rate-limit store backend config and
// constructs the store: nil for the in-memory backend (each limiter then
// builds its own private in-memory store), or a shared Redis-backed store so
// N api replicas enforce ONE combined budget instead of N independent ones.
func buildRateLimitStore() (ratelimit.Store, error) {
	rlStoreCfg, err := ratelimit.LoadStoreConfig(os.Getenv)
	if err != nil {
		slog.Error("rate limit store config", "err", err)
		return nil, err
	}
	rlStore, err := ratelimit.NewStoreFromConfig(rlStoreCfg, slog.Default())
	if err != nil {
		slog.Error("rate limit store init", "err", err)
		return nil, err
	}
	return rlStore, nil
}

// buildPreAuthLimiter builds the pre-auth IP-keyed login rate limiter
// (REQ-MW-5): 10 attempts/min per IP, sharing rlStore with the post-auth
// global envelope limiter when a Redis backend is configured.
func buildPreAuthLimiter(ctx context.Context, authCfg authapp.Config, rlStore ratelimit.Store) (*ratelimit.Middleware, error) {
	loginRateCfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteAuthLogin: 10})
	if err != nil {
		slog.Error("login rate limit config", "err", err)
		return nil, err
	}
	loginRateCfg.TrustedProxyCIDRs = authCfg.TrustedProxyCIDRs
	loginRateCfg.Store = rlStore
	return ratelimit.New(ctx, loginRateCfg), nil
}

// buildRateLimiting wires the M8/F8.2 rate-limit store backend (memory or
// shared Redis) together with the pre-auth IP-keyed login limiter (REQ-MW-5)
// that shares it. The returned cleanup func closes the store for the Redis
// backend and is a no-op for the memory backend (nil store) — the caller
// unconditionally defers it, matching the original conditional
// `if rlStore != nil { defer rlStore.Close() }`.
func buildRateLimiting(ctx context.Context, authCfg authapp.Config) (ratelimit.Store, func(), *ratelimit.Middleware, error) {
	rlStore, err := buildRateLimitStore()
	if err != nil {
		return nil, nil, nil, err
	}
	// nil for the memory backend — each limiter then builds its own private
	// in-memory store (unchanged behavior). Non-nil (redis) is shared by BOTH
	// limiter mounts (this one and the post-auth global envelope limiter in
	// run()) so the process holds exactly one Redis client.
	cleanup := func() {}
	if rlStore != nil {
		cleanup = rlStore.Close
	}
	// Pre-auth IP-keyed rate limit for the login endpoint (REQ-MW-5). Runs
	// before authn in the chain; always keys by client IP. 10 attempts/min
	// per IP — brute force is additionally bounded by account lockout.
	preAuthLimiter, err := buildPreAuthLimiter(ctx, authCfg, rlStore)
	if err != nil {
		return nil, nil, nil, err
	}
	return rlStore, cleanup, preAuthLimiter, nil
}

// buildPeopleHandlers wires the PR-4/PR-1/PR-5 People-tab handler trio
// (People, Membership, Roles & Capabilities) and their shared
// AreaMembershipService/PeopleService orchestrators. Every SQLDB-gated
// dependency degrades to nil in the in-memory boot path, matching every
// other conditionally-wired IAM handler.
func buildPeopleHandlers(deps bootstrap.APIDependencies, authService *authapp.Service, cachedProvider *iamapp.CachedRoleProvider) (*iamdelivery.PeopleHandler, *iamdelivery.MembershipHandler, *iamdelivery.RolesCapsHandler) {
	var membershipService *iamapp.AreaMembershipService
	if deps.SQLDB != nil {
		// WithRoleCacheInvalidator: grant/revoke must flush the cached role set so a
		// changed area membership stops authorizing immediately, not after the TTL (A3).
		membershipService = iamapp.NewAreaMembershipService(
			iampg.NewUserAreaRepository(deps.SQLDB),
			iamapp.NewAuditMembershipLogger(deps.AuditWriter),
		).WithRoleCacheInvalidator(cachedProvider)
	}

	// PR-4: People-tab orchestrator. AreaCatalogReader validates an invite's
	// areaCode against the process-area SSOT (metaldocs.document_process_areas)
	// up front, so an unknown area is a clean boundary error instead of a
	// downstream FK violation. In-memory mode (no SQLDB) leaves it nil, which
	// NewPeopleService resolves to the permissive catalog.
	var areaCatalog iamapp.AreaCatalogReader
	if deps.SQLDB != nil {
		areaCatalog = iampg.NewProcessAreaCatalog(deps.SQLDB, taxonomyinfra.NewAreaCatalogReaderPG())
	}
	peopleService := iamapp.NewPeopleService(authService, cachedProvider, deps.RoleAdminRepo, membershipService, areaCatalog, cachedProvider)
	// H-3b Site 3: wire atomic PatchAtomic (UpdateUserTx + ReplaceUserRolesTx + RecordTx).
	// authpg.Repository satisfies the userUpdaterTx port (UpdateUserTx method).
	if deps.SQLDB != nil {
		peopleService.WithTxAudit(db.NewTxRunner(deps.SQLDB), deps.AuditWriter, authpg.NewRepository(deps.SQLDB, iampg.NewUserTenantRepository(deps.SQLDB)))
	}
	peopleHandler := iamdelivery.NewPeopleHandler(peopleService, authService, deps.AuditWriter)

	// PR-1 (area-memberships rebuild): MembershipHandler now takes a
	// cross-tenant verifier (PeopleService.VerifyUserInTenant) so cross-tenant
	// probes return 404. Grant/revoke audit rows are written in-tx by the
	// service's AuditMembershipLogger (wired above), not by the handler (H-3a).
	membershipHandler := iamdelivery.NewMembershipHandler(membershipService, peopleService)

	// PR-5: IAM Admin Center "Roles & Capabilities" tab: read-only matrix.
	var roleCapsReader iamdelivery.RoleCapabilitiesReader
	if deps.SQLDB != nil {
		roleCapsReader = iampg.NewRoleCapabilitiesRepository(deps.SQLDB)
	}
	rolesCapsHandler := iamdelivery.NewRolesCapsHandler(roleCapsReader)

	return peopleHandler, membershipHandler, rolesCapsHandler
}

// loadFanoutClientConfig loads the fanout client config and asserts the
// approval runtime's hard dependency on METALDOCS_FANOUT_URL (freeze support
// is not optional).
func loadFanoutClientConfig() (config.FanoutConfig, error) {
	fanoutClientCfg, err := config.LoadFanoutConfig()
	if err != nil {
		slog.Error("invalid fanout config", "err", err)
		return config.FanoutConfig{}, err
	}
	if err := requireApprovalRuntimeSupport(fanoutClientCfg.URL); err != nil {
		slog.Error("approval runtime unavailable", "err", err)
		return config.FanoutConfig{}, err
	}
	return fanoutClientCfg, nil
}

// finishDocumentsDependencies wires the two documents.Dependencies fields
// that depend on deps/fanoutCfg being fully constructed: the direct Gotenberg
// PDF export path (when configured) and the fanout reconstruct runner (when
// both a fanout client and a database connection are available).
func finishDocumentsDependencies(docDeps *documents.Dependencies, deps bootstrap.APIDependencies, fanoutCfg fanoutComponents) {
	if deps.PDFConverter != nil {
		docDeps.ExportDocgen = deps.PDFConverter
	}
	if fanoutCfg.client != nil && deps.SQLDB != nil {
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		inputsReader := docrepo.NewFanoutInputsReader(deps.SQLDB)
		docDeps.ReconstructRunner = fanout.NewReconstructService(
			inputsReader, fanoutCfg.client, snapRepo,
			fanout.EngineVersions{EigenpalVer: "local", DocxtemplaterVer: "local"},
			nil,
		)
	}
}

// wireApprovalJobRuntime wires the River job-queue runtime the approval
// kernel depends on: the lifecycle/notification enqueuers, the tenant
// export/erase lifecycle service, the staging-outbox dispatch enqueuer, and
// the terminal-approval release recorder (ADR 0085 / F-QA4-14). riverBundle
// is nil when deps.SQLDB is nil, in which case every River-dependent wiring
// below is skipped — approvalServices stays on its NewServices defaults
// (domain.Noop* enqueuers) exactly as it did inline. The freeze-service
// presence check runs unconditionally (matching its original position,
// outside the SQLDB guard): approval requires a configured freeze service
// regardless of repository mode.
func wireApprovalJobRuntime(
	ctx context.Context,
	deps bootstrap.APIDependencies,
	approvalServices *approvalapp.Services,
	tenantHandler *iamdelivery.TenantHandler,
	iamTxRunner db.TxRunner,
	tenantCrypto securitydomain.TenantCrypto,
	sharedPresigner *objectstore.VerifiedStore,
	fanoutCfg fanoutComponents,
	cdReader *cdinfra.CDFieldReaderPG,
) error {
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		slog.Error("invalid jobs config", "err", err)
		return err
	}
	var riverBundle *riverjobs.ClientBundle
	if deps.SQLDB != nil {
		if err := bootstrap.MigrateRiverSchema(ctx, deps.SQLDB, jobsCfg.RiverSchema); err != nil {
			slog.Error("migrate river schema", "err", err)
			return err
		}
		riverBundle, err = riverjobs.NewClientBundle(deps.SQLDB, riverjobs.Config{
			Queues: jobsCfg.Queues,
			// PeriodicJobs is defined here too (not just in metaldocs-jobs) because
			// River only enqueues periodic jobs from the elected leader's own
			// Config.PeriodicJobs; metaldocs-api joins the same leader election but
			// does NOT subscribe the "maintenance" queue and has nil Workers here,
			// so it enqueues-when-leader but never executes these jobs (ADR 0067
			// dual-define, jobs-only execute topology). The staging-outbox purge
			// job (M5 F5.4 T2) follows the exact same pattern: its periodic-job
			// definition is appended here for leader-election parity, but api
			// never subscribes "maintenance" and never registers the PurgeWorker.
			PeriodicJobs:        append(maintenance.PeriodicJobs(), retention.PeriodicJob()),
			Schema:              jobsCfg.RiverSchema,
			SkipUnknownJobCheck: true,
		}, nil)
		if err != nil {
			slog.Error("build river enqueuer client", "err", err)
			return err
		}
		approvalServices.WithLifecycleEnqueuer(approvaljobs.NewLifecycleEventEnqueuer(riverBundle.Client))
		// Task 6: wire the approval-owned notification port into BOTH submit
		// routes (Submit, TemplateSubmit) so a real submission tells its
		// eligible approvers, not just the audit trail. Without this call the
		// services stay fail-closed to domain.NoopApprovalNotificationEnqueuer —
		// same silent-no-notification defect this feature exists to remove.
		approvalServices.WithApprovalNotificationEnqueuer(approvaljobs.NewApprovalNotificationEnqueuer(riverBundle.Client))

		// M7 F7.3 Task E: tenant export/erase orchestrator. Constructed here
		// (not alongside tenantHandler/onboardTenantService above) because it
		// needs riverBundle.Client for the paired same-tx River enqueue —
		// riverBundle does not exist yet at the point tenantHandler is built.
		// WithLifecycle mutates the already-constructed tenantHandler in
		// place; iamRouter (built later, but referencing the SAME
		// tenantHandler pointer via WithTenantHandler) sees the wiring once
		// Mount's wrapped handlers are invoked at request time —
		// route mounting only captures the Router struct, not a snapshot of
		// tenantHandler's fields.
		if tenantHandler != nil {
			tenantLifecycleService := iamapp.NewTenantLifecycleService(
				iampg.NewTenantRepository(deps.SQLDB),
				iampg.NewTenantLifecycleRepository(deps.SQLDB),
				iampg.NewTenantLifecycleRepository(deps.SQLDB),
				iamjobs.NewTenantLifecycleEnqueuer(riverBundle.Client),
				iamTxRunner,
				deps.AuditWriter,
				deps.SQLDB,
				registry.AllTenantDataPorts(deps.SQLDB),
				sharedPresigner,
				tenantCrypto,
			)
			tenantHandler.WithLifecycle(tenantLifecycleService)
		}
	}
	if fanoutCfg.freezeService == nil {
		slog.Error("approval runtime requires configured freeze service")
		return errors.New("approval runtime requires configured freeze service")
	}
	pdfOutboxRepo := fanout.NewPDFOutboxRepository(deps.SQLDB)
	materializeOutboxRepo := fanout.NewMaterializeOutboxRepository(deps.SQLDB)

	stagingOutboxWorkerCfg, err := config.LoadStagingOutboxWorkerConfig()
	if err != nil {
		slog.Error("invalid staging outbox worker config", "err", err)
		return err
	}

	// pdfDispatchEnqueuer produces the paired (outbox row, River job) write for
	// both staging dispatch kinds inside the caller's business tx (M5 F5.3 T3).
	// riverBundle.Client is enqueue-only here (never Started in this binary);
	// the temporal-queue dispatch workers that consume these jobs run in
	// metaldocs-jobs. riverBundle is nil when deps.SQLDB is nil, so this wiring
	// is guarded the same way the release recorder is below.
	if riverBundle != nil {
		pdfDispatchEnqueuer := dispatchjobs.NewEnqueuer(riverBundle.Client, pdfOutboxRepo, materializeOutboxRepo, stagingOutboxWorkerCfg.MaxAttempts)

		// Wire materialize outbox into the freeze service so Pin can enqueue async jobs.
		fanoutCfg.freezeService.WithMaterializeOutbox(pdfDispatchEnqueuer)
	}

	// Wire the remaining ports directly onto the existing Decision pointer
	// (in place — With* return the same *DecisionService) rather than
	// rebuilding it: pdfDispatchEnqueuer and fanoutCfg.freezeService only
	// exist from this point on in the composition root, but Decision already
	// carries the profile-policy/template ports (above) and the lifecycle
	// enqueuer, and FastForward (built in NewServices) holds this same
	// pointer. Rebuilding here would silently drop all of that wiring from
	// Decision while leaving FastForward on the original, divergent instance.
	approvalServices.Decision = approvalServices.Decision.
		WithSignatureRegistry(newSignoffReauthRegistry(deps.AuthRepo, deps.SQLDB)).
		WithCDFieldReader(cdReader)
	// ADR 0085 / F-QA4-14: ONE terminal-approval recorder — approval fact +
	// async-freeze pin + coordinator evaluation, all on the signoff/verdict
	// transaction — wired into BOTH terminal-approval routes from this single
	// call. A document approved through a review verdict must freeze,
	// materialize and release exactly like one approved through a signoff.
	if riverBundle != nil {
		approvalServices.WithReleaseRecorder(approvalapp.NewReleaseFactRecorder(
			fanoutCfg.freezeService,
			approvaljobs.NewReleaseEvaluationEnqueuer(riverBundle.Client),
		))
	}
	return nil
}

// wireTemplatesApprovalKernel builds the templates module, verifies the
// approval kernel's Decision service is fully wired, and wires the kernel's
// published services into both the templates HTTP handler (its two thin
// kernel routes: submit-for-approval, signoff) and the approval HTTP
// handler. approvalServices.Decision is wired in place throughout (never
// rebuilt), so this observes the same fully-ported instance FastForward and
// approvalHandler do.
func wireTemplatesApprovalKernel(
	deps bootstrap.APIDependencies,
	capabilityService *iamapp.CapabilityService,
	sharedPresigner *objectstore.VerifiedStore,
	displayNameRepo iamdomain.UserDisplayNameReader,
	approvalServices *approvalapp.Services,
) (*templateshttp.Handler, *templatesinfra.Repository, *objectstore.VerifiedStore, *approvalhttp.Handler, error) {
	templatesModule, templatesRepo, templatesStore, err := buildTemplatesModule(deps, capabilityService, sharedPresigner, displayNameRepo)
	if err != nil {
		slog.Error("build templates module", "err", err)
		return nil, nil, nil, nil, err
	}
	if err := approvalServices.Decision.Ready(); err != nil {
		slog.Error("approval decision service not fully wired", "err", err)
		return nil, nil, nil, nil, err
	}
	templatesModule.WithApprovalKernel(approvalServices.TemplateSubmit, approvalServices.Decision, approvalServices.Read, db.NewTxRunner(deps.SQLDB))
	signoffIdempStore := approvalinfra.NewPostgresSignoffIdempStore(deps.SQLDB)
	routeAdminIdempStore := approvalinfra.NewPostgresRouteAdminIdempStore(deps.SQLDB)
	approvalServices.WithRouteAdminIdempStore(routeAdminIdempStore)
	approvalHandler := approvalhttp.NewHandler(approvalServices, deps.SQLDB, signoffIdempStore, displayNameRepo)
	return templatesModule, templatesRepo, templatesStore, approvalHandler, nil
}

// wireApprovalRuntime wires the two approval-kernel dependent construction
// steps that must run in sequence — the River job runtime (wireApprovalJobRuntime,
// which completes approvalServices.Decision's remaining ports) and the
// templates handler (wireTemplatesApprovalKernel, whose approval-kernel
// wiring requires Decision to be Ready()) — behind a single call/error-check
// pair. Neither step depends on the documents module's construction, so both
// run before it (run() now sequences docMod after this call) without
// changing observable behavior: no logging or other side effect occurs on
// either step's success path, only on its own failure.
func wireApprovalRuntime(
	ctx context.Context,
	deps bootstrap.APIDependencies,
	approvalServices *approvalapp.Services,
	tenantHandler *iamdelivery.TenantHandler,
	iamTxRunner db.TxRunner,
	tenantCrypto securitydomain.TenantCrypto,
	sharedPresigner *objectstore.VerifiedStore,
	fanoutCfg fanoutComponents,
	cdReader *cdinfra.CDFieldReaderPG,
	capabilityService *iamapp.CapabilityService,
	displayNameRepo iamdomain.UserDisplayNameReader,
) (*templateshttp.Handler, *templatesinfra.Repository, *objectstore.VerifiedStore, *approvalhttp.Handler, error) {
	if err := wireApprovalJobRuntime(ctx, deps, approvalServices, tenantHandler, iamTxRunner, tenantCrypto, sharedPresigner, fanoutCfg, cdReader); err != nil {
		return nil, nil, nil, nil, err
	}
	return wireTemplatesApprovalKernel(deps, capabilityService, sharedPresigner, displayNameRepo, approvalServices)
}

// decideE2E reports whether the E2E-only test handlers should be mounted:
// both a non-nil publisher (built from the running binary) AND the explicit
// env-flag gate (e2eHandlersEnabled) must hold. Extracted from run() as its
// own boolean expression so run()'s cyclomatic complexity does not carry the
// short-circuit && in addition to the `if useE2E` branch that consumes it.
func decideE2E(e2e httprouter.SurfacePublisher) bool {
	return e2e != nil && e2eHandlersEnabled()
}

// registerPresenceShutdown wires the Z-22/REQ-REL-2 WS presence drain into
// server's shutdown hook when presence is enabled (presenceHub non-nil).
// RegisterOnShutdown runs synchronously inside server.Shutdown after the
// listener is closed but before Shutdown returns, so the 15s shutdown
// context (shutdownServer) covers the drain; this closure keeps its own
// independent 5s budget exactly as the original inline closure did.
func registerPresenceShutdown(server *http.Server, presenceHub *iampresence.Hub) {
	if presenceHub == nil {
		return
	}
	server.RegisterOnShutdown(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		presenceHub.CloseAll(shutdownCtx)
	})
}

// mountPublishers mounts every publisher onto mux through its own
// httprouter.Recorder (one recorder PER publisher — required by assertSurface
// check 4, surface.go) and returns the per-publisher mounted-pattern map.
func mountPublishers(mux *http.ServeMux, publishers []httprouter.SurfacePublisher) map[string][]string {
	mounted := map[string][]string{}
	for _, p := range publishers {
		rec := httprouter.NewRecorder(mux) // one recorder PER publisher — check 4
		p.Mount(rec)
		mounted[p.Name()] = rec.Patterns()
	}
	return mounted
}

// wireHTTPObservabilityDBPool wires the DB pool-stats adapter into httpObs
// when a database connection is available (SQLDB nil in in-memory mode).
func wireHTTPObservabilityDBPool(httpObs *observability.HTTPObservability, deps bootstrap.APIDependencies) {
	if deps.SQLDB != nil {
		httpObs.SetDBPool(postgres.NewPoolStatsAdapter(deps.SQLDB))
	}
}

// buildPresenceWrap returns presenceBump.Wrap when presence is enabled
// (deps.SQLDB non-nil at startup), or nil otherwise — apiChain skips a nil
// link.
func buildPresenceWrap(presenceBump *iampresence.BumpMiddleware) func(http.Handler) http.Handler {
	if presenceBump != nil {
		return presenceBump.Wrap
	}
	return nil
}

// buildOtelWrap returns the OTel middleware when tracing is enabled (an
// exporter is configured), or nil otherwise — apiChain skips a nil link
// (Z-1, REQ-OBS-1).
func buildOtelWrap(otelEnabled bool) func(http.Handler) http.Handler {
	if otelEnabled {
		return observability.OTelMiddleware()
	}
	return nil
}

// buildServers constructs the public API server (wired to handler) and the
// dedicated metrics server (wired to httpObs.PrometheusHandler, wrapped only
// in panic recovery) from ServerConfig.
//
// /metrics is served on a DEDICATED listener (METRICS_ADDR, default :9090),
// never on the public API server. This isolates the scrape surface by
// process topology: the public port structurally cannot serve /metrics, so
// exposure no longer depends on ops/ingress discipline (F-R1, Dim-9). The
// scrape stays credential-less (bypasses authn/iam) and self-scrapes never
// feed httpObs.Wrap counters or the global rate limiter, because this mux is
// not part of the API chain. Not in openapi (contract §3.2). Compose does
// not host-publish this port — infra-network reachable only (see
// ops/DEPLOY.md).
func buildServers(handler http.Handler, httpObs *observability.HTTPObservability) (*http.Server, *http.Server, error) {
	serverCfg, err := config.LoadServerConfig()
	if err != nil {
		slog.Error("invalid server config", "err", err)
		return nil, nil, err
	}

	server := &http.Server{
		Addr:              serverCfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		// REQ-REL-1/2 (F-16): bound slow-read/slow-write clients so they
		// cannot hold connections indefinitely. WriteTimeout is sized for
		// the slowest synchronous responses (PDF export, render).
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  90 * time.Second,
	}

	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics", httpObs.PrometheusHandler())
	metricsServer := &http.Server{
		Addr:              serverCfg.MetricsAddr,
		Handler:           platformmw.Recovery(metricsMux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server, metricsServer, nil
}

// shutdownServer waits for a listen error on EITHER server or ctx cancellation,
// then drains both listeners: the public API server and the dedicated metrics
// server. A fatal bind error on the metrics listener is treated the same as one
// on the public listener (fail-fast) — a misconfigured METRICS_ADDR must not
// silently leave the scrape surface down. Returns a non-zero exit code only if
// a real failure occurred (genuine ListenAndServe error or a Shutdown that
// didn't drain cleanly).
func shutdownServer(
	ctx context.Context,
	stop context.CancelFunc,
	server *http.Server,
	metricsServer *http.Server,
	serverErr <-chan error,
	metricsErr <-chan error,
) int {
	exitCode := 0
	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			exitCode = 1
		}
	case err := <-metricsErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server failed", "err", err)
			exitCode = 1
		}
	case <-ctx.Done():
	}

	// shutdownCtx is intentionally detached from ctx via context.WithoutCancel
	// rather than context.Background(): ctx is the process signal-context, and
	// by the time we reach here it is already Done() (that is precisely why
	// shutdownServer was called — either a listen error occurred or the signal
	// fired). Deriving the 15s shutdown deadline from ctx directly would cancel
	// it instantly, defeating the graceful-drain budget below. WithoutCancel
	// keeps ctx in the propagation chain (any values it carries survive) while
	// detaching only from its cancellation.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown incomplete", "err", err)
		if exitCode == 0 {
			exitCode = 1
		}
	} else {
		slog.Info("graceful shutdown complete")
	}
	// Drain the metrics listener too; best-effort within the same budget. A
	// failure here is logged but does not by itself fail an otherwise-clean
	// shutdown of the public server.
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("metrics server shutdown incomplete", "err", err)
	}
	stop()
	slog.Info("workers stopped")
	return exitCode
}

func requirePostgresRepositoryMode(repoMode string) error {
	if repoMode != config.RepositoryPostgres {
		return fmt.Errorf("metaldocs api requires %q repository mode; got %q", config.RepositoryPostgres, repoMode)
	}
	return nil
}

func buildTaxonomyModule(deps bootstrap.APIDependencies) *taxonomy.Module {
	return taxonomy.New(taxonomy.Dependencies{
		DB:          deps.SQLDB,
		TplChecker:  templatesinfra.NewTemplateVersionReader(deps.SQLDB),
		AuditWriter: deps.AuditWriter,
		// Approval-owned readiness port for the profile listing's
		// has_active_route badge (approval owns approval_routes).
		RouteReadinessReader: approvalrepo.NewRouteReadinessReaderPG(deps.SQLDB),
	})
}

func buildControlledDocumentsModule(deps bootstrap.APIDependencies) *controlleddocuments.Module {
	// Construct sibling-module collaborator concretes here (composition root),
	// not inside the CD module constructor. Each concrete satisfies the
	// interface-typed port declared in controlleddocuments.Dependencies (F-CD4).
	profileRepo := taxonomyinfra.NewProfileRepository(deps.SQLDB)
	areaRepo := taxonomyinfra.NewAreaRepository(deps.SQLDB)
	return controlleddocuments.New(controlleddocuments.Dependencies{
		DB:     deps.SQLDB,
		Logger: slog.Default(),
		// documents-owned active-instance read-port (ADR-0039 D3(b); M2/F2.2).
		ActiveInstanceReader: docrepo.NewActiveInstanceReaderPG(deps.SQLDB),
		// Taxonomy profile/area readers: adapters in CD infrastructure wrap the
		// canonical taxonomy repositories so authz GUC + CapTaxonomyView run on
		// every lookup (H-1b). The adapters satisfy application.ProfileReader /
		// application.AreaReader — consumer-defined interfaces.
		ProfileReader: cdinfra.NewTaxonomyProfileReader(profileRepo),
		AreaReader:    cdinfra.NewTaxonomyAreaReader(areaRepo),
		// GovernanceLogger routes governance events to the canonical audit sink
		// (taxonomy/application.AuditGovernanceAdapter satisfies
		// taxonomydomain.GovernanceLogger).
		GovernanceLogger: taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter),
		// TemplateVersionChecker reads template-version state through the
		// templates-owned port (M4 F4.2 — H-G reach closed).
		TemplateVersionChecker: templatesinfra.NewTemplateVersionReader(deps.SQLDB),
		// RouteReadinessReader is the approval-owned port backing the hard
		// creation gate (D2) and the creation-context read model. approval owns
		// approval_routes; CD never touches that table.
		RouteReadinessReader: approvalrepo.NewRouteReadinessReaderPG(deps.SQLDB),
		// Creation-context read model: taxonomy catalog listers (same adapter
		// rationale as ProfileReader/AreaReader above) plus the iam-owned
		// capability-reach reader that narrows areas server-side.
		ProfileLister:        cdinfra.NewTaxonomyProfileLister(profileRepo),
		AreaLister:           cdinfra.NewTaxonomyAreaLister(areaRepo),
		AreaCapabilityReader: iampg.NewAreaCapabilityReaderPG(deps.SQLDB),
	})
}

func buildTokensModule(deps bootstrap.APIDependencies) *tokens.Module {
	return tokens.New(tokens.Dependencies{
		DB:            deps.SQLDB,
		AuditWriter:   deps.AuditWriter,
		ReservedNames: newReservedNamesFromRegistry(),
	})
}

// buildTemplatesModule returns the templates HTTP handler plus the repository and
// object store it is built on, so the composition root can wire the background
// orphan-object sweeper (F-T6) against the same instances the request path uses.
// presigner is the shared VerifiedStore constructed once at the composition root
// (STO-02) — templates does not construct its own instance.
func buildTemplatesModule(deps bootstrap.APIDependencies, capabilityService *iamapp.CapabilityService, presigner *objectstore.VerifiedStore, displayNameReader iamdomain.UserDisplayNameReader) (*templateshttp.Handler, *templatesinfra.Repository, *objectstore.VerifiedStore, error) {
	if capabilityService == nil {
		return nil, nil, nil, errors.New("templates capability service is required")
	}
	if presigner == nil {
		return nil, nil, nil, errors.New("templates object store presigner is required")
	}
	templatesPresigner := presigner
	// Build a dedicated builtins registry so the D5 reserved-name guard fires in
	// production. Mirrors the pattern in reserved_names.go (SP-2 §5.1). The 8 static
	// builtins are cheap; a separate instance here keeps this independent of the
	// resolverReg built later at the composition root for runtime resolution.
	templatesResolverReg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(templatesResolverReg)
	// SEC-08 fail-fast: PHComputed.resolver_key validation (schema.go ValidatePlaceholders)
	// silently no-ops when the ResolverRegistryReader is nil or empty, letting an
	// arbitrary resolver_key persist into published templates and every document
	// instantiated from them (wiki/modules/templates-tech-debt.md T-008). Assert the
	// registry is wired and non-empty at boot instead of trusting the call graph.
	if templatesResolverReg == nil || len(templatesResolverReg.Known()) == 0 {
		return nil, nil, nil, errors.New("templates resolver registry is nil or empty; resolver_key validation would be silently skipped (SEC-08 / T-008)")
	}
	templatesRepo := templatesinfra.New(deps.SQLDB).WithAudit(deps.AuditWriter)
	// ADR 0086: the config-first creation gate reads template-route readiness
	// through the approval-owned port, in the create tx. Without it CreateTemplate
	// fails closed, so it is wired here (never left nil in production).
	templatesSvc := templatesapp.New(templatesRepo, templatesPresigner, wiring.Clock{}, wiring.UUIDGen{}, templatesResolverReg).
		WithRunner(db.NewTxRunner(deps.SQLDB)).
		WithRouteReadinessReader(approvalrepo.NewRouteReadinessReaderPG(deps.SQLDB))
	templatesAuthzFn := func(r *http.Request, tenantID, _ string, action string) error {
		userID := iamdomain.UserIDFromContext(r.Context())
		return capabilityService.CanDo(r.Context(), userID, tenantID, action)
	}
	// FE-08: resolve TemplateDTO.created_by_display_name via the iam-owned
	// UserDisplayNameReader port (M4/F4.1) instead of templates querying
	// iam_users directly.
	templatesHandler := templateshttp.New(templatesSvc, templatesAuthzFn, deps.SQLDB).WithDisplayNameReader(displayNameReader)
	return templatesHandler, templatesRepo, templatesPresigner, nil
}

func requireApprovalRuntimeSupport(fanoutURL string) error {
	if strings.TrimSpace(fanoutURL) == "" {
		return errors.New("approval runtime requires METALDOCS_FANOUT_URL; startup without freeze support is not allowed")
	}
	return nil
}

// startPresence initialises the PR-9 presence subsystem: hub goroutines, WebSocket
// handler, HTTP snapshot fallback, bump middleware with cleanup, and wires the
// presence reader into iamAdminHandler. Returns (nil, nil, nil) when deps.SQLDB is
// nil (in-memory mode). The returned Hub is captured by main for shutdown drain
// (Z-22, REQ-REL-2); the BumpMiddleware is wrapped into the outer request chain
// so authenticated requests refresh last_seen_at (debounced 60s per user, PR-9).
//
// CON-07: the presence Handler's WebSocket /iam/presence/stream route mounts
// via presence.MountStream (see presence/handler.go), because streamPresence
// is excluded from server codegen (cfg.yaml exclude-operation-ids). Task 15a
// (Ruling 2) folded that mount into iamdelivery.Router.Mount, which also
// mounts the HTTP-fallback snapshot route via the returned *iampresence.
// Handler (ServeSnapshot) — neither route is mounted here at construction
// time; both are mounted by buildPublishers (publishers.go) via iamRouter,
// which is itself the SurfacePublisher for the "iam" tag.
func startPresence(
	ctx context.Context,
	deps bootstrap.APIDependencies,
	iamAdminHandler *iamdelivery.AdminHandler,
) (*iampresence.BumpMiddleware, *iampresence.Hub, *iampresence.Handler) {
	if deps.SQLDB == nil {
		return nil, nil, nil
	}
	presenceRepo := iampresence.NewPostgresRepository(deps.SQLDB)
	presenceHub := iampresence.NewHub(presenceRepo, slog.Default())
	go presenceHub.Run(ctx)
	go presenceHub.RunHeartbeat(ctx)
	// presenceHandler mounts via buildPublishers (publishers.go), inside
	// iamRouter's Mount, alongside every other route family — not here at
	// construction time.
	presenceHandler := iampresence.NewHandler(presenceHub, presenceRepo, slog.Default())
	presenceBump := iampresence.NewBumpMiddleware(presenceRepo, slog.Default())
	presenceBump.StartCleanup(ctx)
	iamAdminHandler.WithPresenceReader(presenceRepo)
	return presenceBump, presenceHub, presenceHandler
}

// buildFanoutComponents constructs the fanout client and freeze service when
// METALDOCS_FANOUT_URL is set and a database connection is available. Returns
// an empty fanoutComponents when either condition is not met. The materialize
// outbox is NOT wired here — the caller does that via
// fanoutCfg.freezeService.WithMaterializeOutbox after pdfOutboxRepo /
// materializeOutboxRepo are created.
func buildFanoutComponents(
	deps bootstrap.APIDependencies,
	cfg config.FanoutConfig,
	ctlDocs *controlleddocuments.Module,
) fanoutComponents {
	if cfg.URL == "" || deps.SQLDB == nil {
		return fanoutComponents{}
	}
	client := fanout.NewClient(cfg.URL, cfg.ServiceToken, httpclient.NewInternalClient())
	snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
	fillInRepo := docrepo.NewFillInRepository(deps.SQLDB)
	schemaReader := docapp.NewSnapshotSchemaReader(deps.SQLDB)
	revReader := docrepo.NewRevisionReader(deps.SQLDB)
	wfReader := docrepo.NewWorkflowReader(deps.SQLDB)
	ctxBuilder := docapp.NewDocumentContextBuilder(
		deps.SQLDB,
		wiring.NewSearchRevisionReader(revReader),
		wiring.NewSearchWorkflowReader(wfReader),
		wiring.NewControlledDocumentsReader(ctlDocs.Repo()),
		wiring.NewSearchDocumentReader(revReader),
	)
	resolverReg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(resolverReg)
	freezeService := docapp.NewFreezeService(
		schemaReader, fillInRepo, fillInRepo,
		resolverReg, snapRepo, ctxBuilder,
		snapRepo, client,
	)
	return fanoutComponents{client: client, freezeService: freezeService}
}

// startAuditRetention launches a background goroutine that purges audit_events
// older than `days` days on a 24-hour tick. Skips when days <= 0 or
// deps.SQLDB is nil (AUDIT_RETENTION_DAYS=0 disables, default disabled).
func startAuditRetention(ctx context.Context, deps bootstrap.APIDependencies, days int) {
	if days <= 0 || deps.SQLDB == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().UTC().AddDate(0, 0, -days)
				if _, err := deps.SQLDB.ExecContext(ctx,
					`DELETE FROM metaldocs.audit_events WHERE occurred_at < $1`, cutoff,
				); err != nil {
					slog.Warn("audit retention purge failed", "error", err)
				}
			}
		}
	}()
}
