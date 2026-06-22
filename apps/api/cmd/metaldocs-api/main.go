package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	documents "metaldocs/internal/modules/documents"
	docapp "metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvalhttp "metaldocs/internal/modules/documents/approval/http"
	approvalinfra "metaldocs/internal/modules/documents/approval/infrastructure"
	approvaljobs "metaldocs/internal/modules/documents/approval/jobs"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/documents/jobs"
	docrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/jobs/audit_integrity_validator"
	"metaldocs/internal/modules/jobs/idempotency_janitor"
	jobscheduler "metaldocs/internal/modules/jobs/scheduler"
	"metaldocs/internal/modules/jobs/stuck_instance_watchdog"
	templatesapp "metaldocs/internal/modules/templates/application"
	templateshttp "metaldocs/internal/modules/templates/delivery/http"
	templatesrepo "metaldocs/internal/modules/templates/repository"

	"metaldocs/apps/api/internal/wiring"
	auditapp "metaldocs/internal/modules/audit/application"
	auditdelivery "metaldocs/internal/modules/audit/delivery/http"
	authapp "metaldocs/internal/modules/auth/application"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	controlleddocuments "metaldocs/internal/modules/controlleddocuments"
	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	distributionhttp "metaldocs/internal/modules/distribution/delivery/http"
	distributioninfra "metaldocs/internal/modules/distribution/infrastructure"
	notificationshttp "metaldocs/internal/modules/notifications/delivery/http"
	notificationsinfra "metaldocs/internal/modules/notifications/infrastructure"
	iamapp "metaldocs/internal/modules/iam/application"
	"metaldocs/internal/modules/iam/authz"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	iampresence "metaldocs/internal/modules/iam/presence"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	searchapp "metaldocs/internal/modules/search/application"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	searchdocs "metaldocs/internal/modules/search/infrastructure/v2documents"
	securityapp "metaldocs/internal/modules/security/application"
	securitydelivery "metaldocs/internal/modules/security/delivery/http"
	securitypg "metaldocs/internal/modules/security/infrastructure/postgres"
	"metaldocs/internal/modules/taxonomy"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	templatesinfra "metaldocs/internal/modules/templates/infrastructure"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	"metaldocs/internal/platform/db"
	postgres "metaldocs/internal/platform/db/postgres"
	docgenv2 "metaldocs/internal/platform/docgenv2"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/formval"
	"metaldocs/internal/platform/httpclient"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/internal/platform/messaging"
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

// e2eHandlersEnabled gates the test-only seed/reset/governance endpoints
// behind an explicit env flag. Defense-in-depth: internal/test/e2e_seed.go
// also early-returns on the same flag, but the call site must not register
// these routes either — permissions.go does not enumerate /internal/test/*,
// so any accidental mount is treated as fully public by newPublicPathChecker.
func e2eHandlersEnabled() bool {
	return e2etest.E2EEnabled()
}

func mountE2EHandlersIfEnabled(mux *http.ServeMux, register func(*http.ServeMux)) {
	if !e2eHandlersEnabled() {
		slog.Info("e2etest handlers not mounted", "reason", "METALDOCS_E2E != 1")
		return
	}
	slog.Warn("e2etest handlers mounted — destructive endpoints reachable without auth", "env", "METALDOCS_E2E=1")
	register(mux)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OpenTelemetry: inert unless an exporter is configured (Z-1, REQ-OBS-3).
	// otelShutdown is a no-op when disabled; otelEnabled gates the chain link.
	otelShutdown, otelEnabled, err := observability.SetupOTel(ctx, "metaldocs-api")
	if err != nil {
		slog.Error("setup otel", "err", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			slog.Warn("otel shutdown", "err", err)
		}
	}()
	if otelEnabled {
		slog.Info("OpenTelemetry tracing enabled", "exporter", os.Getenv("OTEL_TRACES_EXPORTER"))
	}

	repoMode, err := config.RepositoryMode()
	if err != nil {
		slog.Error("invalid repository mode", "err", err)
		os.Exit(1)
	}
	if err := requirePostgresRepositoryMode(repoMode); err != nil {
		slog.Error("unsupported repository mode", "err", err)
		os.Exit(1)
	}
	corsCfg, err := config.LoadCORSConfig()
	if err != nil {
		slog.Error("invalid cors config", "err", err)
		os.Exit(1)
	}
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		slog.Error("invalid attachments config", "err", err)
		os.Exit(1)
	}
	authCfg, err := authn.LoadRuntimeConfig()
	if err != nil {
		slog.Error("invalid auth config", "err", err)
		os.Exit(1)
	}
	featureFlagsCfg, err := config.LoadFeatureFlagsConfig()
	if err != nil {
		slog.Error("invalid feature flags config", "err", err)
		os.Exit(1)
	}

	deps, err := bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)
	if err != nil {
		slog.Error("build api dependencies", "err", err)
		os.Exit(1)
	}
	defer deps.Cleanup()

	migrationCfg, err := config.LoadMigrationConfig()
	if err != nil {
		slog.Error("invalid migration config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	if deps.SQLDB != nil && !migrationCfg.Skip {
		if err := migrate.Apply(ctx, deps.SQLDB, migrationCfg.Dir, slog.Default()); err != nil {
			slog.Error("apply startup migrations", "err", err)
			deps.Cleanup()
			os.Exit(1)
		}
	}

	authService, err := authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, iampg.NewLoginContextRepository(deps.SQLDB), authCfg, deps.AuditWriter)
	if err != nil {
		slog.Error("new auth service", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	if err := authService.BootstrapLocalAdmin(ctx); err != nil {
		slog.Error("bootstrap local admin", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}

	auditService := auditapp.NewService(deps.AuditReader)
	if deps.AuditCounter != nil && deps.AuditExports != nil {
		auditService.WithExports(deps.AuditCounter, deps.AuditExports, deps.AuditWriter, func(job auditdomain.ExportJob) string {
			if job.ID == "" || job.DownloadToken == "" {
				return ""
			}
			return fmt.Sprintf("/api/v1/audit/events/export/%s/download?token=%s", job.ID, job.DownloadToken)
		})
	}

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
	healthHandler := observability.NewHealthHandler(deps.StatusProvider)

	capabilityService := iamapp.NewCapabilityService(deps.SQLDB)
	// Wire optional capability hint into /auth/me + login responses.
	// Backend remains sole authz enforcer — FE consumes for UX hints only.
	authService.WithCapabilityProvider(capabilityService)
	cachedProvider := iamapp.NewCachedRoleProvider(ctx, deps.RoleProvider, authn.CacheTTL())
	// permResolver is the single authoritative source of truth for route
	// visibility. It is shared with the auth middleware so that fully public
	// routes (no session required) and the IAM permission layer stay in sync
	// automatically — adding a new public route requires one change here, not two.
	permResolver := newPermissionResolver()
	authMiddleware := authdelivery.NewMiddleware(authService, authCfg, authn.Enabled()).
		WithPublicPathChecker(newPublicPathChecker(permResolver))
	iamMiddleware := iamdelivery.NewMiddleware(capabilityService, cachedProvider, authn.Enabled()).
		WithPermissionResolver(permResolver)
	originProtection := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           authCfg.OriginProtection,
		SessionCookieName: authCfg.SessionCookieName,
		TrustedOrigins:    authCfg.TrustedOrigins,
		TrustedProxyCIDRs: authCfg.TrustedProxyCIDRs,
	})

	// H-3b: TxRunner for IAM atomic audit writes (Site 2, 3, 4).
	var iamTxRunner db.TxRunner
	if deps.SQLDB != nil {
		iamTxRunner = db.NewTxRunner(deps.SQLDB)
	}
	iamAdminService := iamapp.NewAdminService(deps.RoleAdminRepo, cachedProvider, iamTxRunner, deps.AuditWriter)
	iamAdminHandler := iamdelivery.NewAdminHandler(iamAdminService, authService, deps.AuditWriter).
		WithAuditEventLister(auditService)

	// PR-7 Sessions & Security tab.
	var sessionsHandler *iamdelivery.SessionsHandler
	if sqlDB := deps.SQLDB; sqlDB != nil {
		authRepo := authpg.NewRepository(sqlDB, iampg.NewUserTenantRepository(sqlDB))
		sessionSvc := iamapp.NewSessionService(db.NewTxRunner(sqlDB), deps.AuditWriter, authRepo)
		// M4/F4.4: auth returns auth-owned session rows; the iam consumer enriches
		// display names via the iam-owned port (read off-tx on the pool).
		sessionsHandler = iamdelivery.NewSessionsHandler(authRepo, deps.AuditWriter).
			WithSessionService(sessionSvc).
			WithDisplayNameReader(iampg.NewUserDisplayNameRepository(sqlDB))
	}
	var securityService *securityapp.Service
	var securityHandler *securitydelivery.Handler
	if sqlDB := deps.SQLDB; sqlDB != nil {
		// Security reports on auth_identities/auth_sessions but does not own
		// iam_users or iam_user_roles; it resolves tenant membership, display names,
		// and admin-role membership via iam-owned ports (M4/F4.2) instead of JOINing
		// iam's tables.
		securityService = securityapp.NewService(securitypg.NewRepository(
			sqlDB,
			iampg.NewUserDisplayNameRepository(sqlDB),
			iampg.NewTenantUserRepository(sqlDB),
			iampg.NewAdminRoleMemberRepository(sqlDB),
			iampg.NewMfaUserRepository(sqlDB),
		))
		securityHandler = securitydelivery.NewHandler(securityService)
	} else {
		securityHandler = securitydelivery.NewHandler(nil)
	}

	// PR-8 Observability service.
	var observabilityHandler *iamdelivery.ObservabilityHandler
	if sqlDB := deps.SQLDB; sqlDB != nil {
		observabilityRepo := iampg.NewObservabilityRepository(sqlDB)
		observabilityService := iamapp.NewObservabilityService(observabilityRepo, wiring.NewMfaCoveragePctReader(securityService))
		observabilityHandler = iamdelivery.NewObservabilityHandler(observabilityService)
		iamAdminHandler = iamAdminHandler.WithObservabilityService(observabilityService)
	}
	featureFlagsHandler := featureflags.NewHandler(featureFlagsCfg)
	httpObs := observability.NewHTTPObservability(
		func(r *http.Request) string {
			if currentUser, ok := authdomain.CurrentUserFromContext(r.Context()); ok {
				return currentUser.UserID
			}
			return ""
		},
		deps.StatusProvider,
	)
	cors := security.NewCORS(corsCfg)

	// Pre-auth IP-keyed rate limit for the login endpoint (REQ-MW-5). Runs
	// before authn in the chain; always keys by client IP. 10 attempts/min
	// per IP — brute force is additionally bounded by account lockout.
	loginRateCfg, err := ratelimit.NewConfig(map[ratelimit.RouteKey]int{ratelimit.RouteAuthLogin: 10})
	if err != nil {
		slog.Error("login rate limit config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	loginRateCfg.TrustedProxyCIDRs = authCfg.TrustedProxyCIDRs
	preAuthLimiter := ratelimit.New(ctx, loginRateCfg)

	// Post-authn global envelope limiter (F-05/D-04, Wave 2.8): replaces the
	// legacy security.RateLimiter. Same identity precedence: authenticated user
	// → trusted-proxy-resolved client IP → fail-closed. Default: 120 req/min
	// per identity (matches old fixed-window default of 120 req / 60s window).
	// Env vars METALDOCS_RATE_LIMIT_ENABLED / _WINDOW_SECONDS / _MAX_REQUESTS
	// are now dead — see commit body for mapping.
	globalRateCfg := ratelimit.DefaultConfig()
	globalRateCfg.TrustedProxyCIDRs = authCfg.TrustedProxyCIDRs
	globalLimiter := ratelimit.New(ctx, globalRateCfg)
	// userIDExtractor resolves the authenticated principal from context. Runs
	// after authn + iamMiddleware in the chain, so both auth and IAM user IDs
	// are available. Mirrors security.RateLimiter.requestIdentity without
	// importing domain packages (dependency injected via closure).
	userIDExtractor := func(r *http.Request) string {
		if currentUser, ok := authdomain.CurrentUserFromContext(r.Context()); ok && strings.TrimSpace(currentUser.UserID) != "" {
			return strings.TrimSpace(currentUser.UserID)
		}
		if userID := strings.TrimSpace(iamdomain.UserIDFromContext(r.Context())); userID != "" {
			return userID
		}
		return ""
	}

	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)
	featureFlagsHandler.RegisterRoutes(mux)
	auditHandler.RegisterRoutes(mux)
	searchHandler.RegisterRoutes(mux)
	iamAdminHandler.RegisterRoutes(mux)
	if sessionsHandler != nil {
		sessionsHandler.RegisterRoutes(mux)
	}
	if securityHandler != nil {
		securityHandler.RegisterRoutes(mux)
	}
	if observabilityHandler != nil {
		observabilityHandler.RegisterRoutes(mux)
	}

	presenceBump, presenceHub := startPresence(ctx, deps, mux, iamAdminHandler)

	taxonomyModule := buildTaxonomyModule(deps)
	taxonomyModule.RegisterRoutes(mux)

	controlledDocumentsModule := buildControlledDocumentsModule(deps)
	controlledDocumentsModule.RegisterRoutes(mux)
	controlledDocumentDuplicator := wiring.NewControlledDocumentDuplicator(controlledDocumentsModule.Service())

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
	iamdelivery.NewPeopleHandler(peopleService, authService, deps.AuditWriter).RegisterRoutes(mux)

	// PR-1 (area-memberships rebuild): MembershipHandler now takes a
	// cross-tenant verifier (PeopleService.VerifyUserInTenant) so cross-tenant
	// probes return 404. Grant/revoke audit rows are written in-tx by the
	// service's AuditMembershipLogger (wired above), not by the handler (H-3a).
	iamdelivery.NewMembershipHandler(membershipService, peopleService).RegisterRoutes(mux)

	// PR-5: IAM Admin Center "Roles & Capabilities" tab: read-only matrix.
	var roleCapsReader iamdelivery.RoleCapabilitiesReader
	if deps.SQLDB != nil {
		roleCapsReader = iampg.NewRoleCapabilitiesRepository(deps.SQLDB)
	}
	iamdelivery.NewRolesCapsHandler(roleCapsReader).RegisterRoutes(mux)

	// Legacy templates module routes removed — templates owns /api/v1/templates/*

	docPresigner := objectstore.NewDocumentPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 15*time.Minute, 25*1024*1024)
	profileRepo := taxonomyinfra.NewProfileRepository(deps.SQLDB)

	// Fanout/eigenpal client — enabled when METALDOCS_FANOUT_URL is set.
	fanoutClientCfg, err := config.LoadFanoutConfig()
	if err != nil {
		slog.Error("invalid fanout config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	if err := requireApprovalRuntimeSupport(fanoutClientCfg.URL); err != nil {
		slog.Error("approval runtime unavailable", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	fanoutCfg := buildFanoutComponents(deps, fanoutClientCfg, controlledDocumentsModule)

	// M4/F4.1: construct the iam-owned display-name port once and inject into all
	// three consumers (approval repo, documents repo, approval handler).
	displayNameRepo := iampg.NewUserDisplayNameRepository(deps.SQLDB)

	docSnapshotReader := docgenv2.NewTemplatesSnapshotReader(deps.SQLDB)
	docDeps := documents.Dependencies{
		DB:      deps.SQLDB,
		Presign: docPresigner,
		TplRead: docgenv2.NewFanoutTemplateReader(
			docgenv2.NewTemplateReader(deps.SQLDB, deps.MinioClient, deps.MinioBucket),
			docgenv2.NewTemplatesTemplateReader(deps.SQLDB),
		),
		FormVal:                      formval.NewGojsonschema(),
		Audit:                        wiring.NewDocumentsAuditSink(deps.AuditWriter),
		ExportPresign:                docPresigner,
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

	// Approval services must be constructed before docMod so that
	// SubmitSvc can be wired into the finalize→submit flow.
	approvalRepo := approvalrepo.NewPostgresApprovalRepository(deps.SQLDB, displayNameRepo)
	approvalEmitter := approvalapp.NewSQLEmitter()
	approvalServices := approvalapp.NewServices(approvalRepo, approvalEmitter, approvalapp.RealClock{}, cdReader)
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		slog.Error("invalid jobs config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	if deps.SQLDB != nil {
		if err := bootstrap.MigrateRiverSchema(ctx, deps.SQLDB, jobsCfg.RiverSchema); err != nil {
			slog.Error("migrate river schema", "err", err)
			deps.Cleanup()
			os.Exit(1)
		}
		riverBundle, err := riverjobs.NewClientBundle(deps.SQLDB, riverjobs.Config{
			Queues:              jobsCfg.Queues,
			Schema:              jobsCfg.RiverSchema,
			SkipUnknownJobCheck: true,
		}, nil)
		if err != nil {
			slog.Error("build scheduled publish enqueuer client", "err", err)
			deps.Cleanup()
			os.Exit(1)
		}
		approvalServices.WithScheduledPublishEnqueuer(approvaljobs.NewScheduledPublishEnqueuer(riverBundle.Client))
		approvalServices.WithLifecycleEnqueuer(approvaljobs.NewLifecycleEventEnqueuer(riverBundle.Client))
	}
	if fanoutCfg.freezeService == nil {
		slog.Error("approval runtime requires configured freeze service")
		deps.Cleanup()
		os.Exit(1)
	}
	pdfOutboxRepo := fanout.NewPDFOutboxRepository(deps.SQLDB)
	materializeOutboxRepo := fanout.NewMaterializeOutboxRepository(deps.SQLDB)

	// Wire materialize outbox into the freeze service so Pin can enqueue async jobs.
	fanoutCfg.freezeService.WithMaterializeOutbox(materializeOutboxRepo)

	// StagingOutboxWorker.Run() only returns nil (context cancellation); no restart loop needed.
	var workerWG sync.WaitGroup
	startOutboxWorkers(ctx, &workerWG, deps.Publisher, pdfOutboxRepo, materializeOutboxRepo)

	approvalServices.Decision = approvalapp.NewDecisionService(
		approvalRepo, approvalEmitter, approvalapp.RealClock{}, fanoutCfg.freezeService,
	).WithPDFOutbox(pdfOutboxRepo).WithPinInvoker(fanoutCfg.freezeService).
		WithSignatureRegistry(newSignoffReauthRegistry(deps.AuthRepo, deps.SQLDB)).
		WithCDFieldReader(cdReader)
	docDeps.SubmitSvc = approvalServices.Submit

	docMod := documents.New(docDeps)
	docMod.RegisterRoutesWithRateLimit(mux, globalLimiter, userIDExtractor)

	// Wire the documents-side adapter back into the controlled-documents service so atomic
	// CD-create can clone the initial document inside the same tx as the CD
	// insert. controlledDocumentsModule was constructed before docMod (because docMod
	// needs ControlledDocumentDuplicator), hence the post-construction setter.
	controlledDocumentsModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))

	templatesModule, err := buildTemplatesModule(deps, capabilityService)
	if err != nil {
		slog.Error("build templates module", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	templatesModule.Register(mux)
	signoffIdempStore := approvalinfra.NewPostgresSignoffIdempStore(deps.SQLDB)
	routeAdminIdempStore := approvalinfra.NewPostgresRouteAdminIdempStore(deps.SQLDB)
	approvalServices = approvalServices.WithRouteAdminIdempStore(routeAdminIdempStore)
	approvalHandler := approvalhttp.NewHandler(approvalServices, deps.SQLDB, signoffIdempStore, displayNameRepo)
	approvalHandler.RegisterRoutes(mux)

	// M2/F2.2: distribution module — read-only delivery + repository layer.
	// Reuses displayNameRepo (constructed above at line ~418); no new DB handle.
	distributionRepo := distributioninfra.NewCoverageRepository(deps.SQLDB, displayNameRepo)
	distributionHandler := distributionhttp.NewHandler(distributionRepo)
	distributionhttp.RegisterRoutes(distributionHandler, mux)

	// M3/F3.2: notifications module — read surface (list / unread-count / mark-read).
	// Self-scoped by CapNotificationRead (tier-1) + recipient_user_id SQL predicate.
	notificationsRepo := notificationsinfra.NewNotificationsRepository(deps.SQLDB)
	notificationsHandler := notificationshttp.NewHandler(notificationsRepo)
	notificationshttp.RegisterRoutes(notificationsHandler, mux)

	mountE2EHandlersIfEnabled(mux, func(m *http.ServeMux) {
		e2etest.RegisterE2EHandlers(m, deps.SQLDB, nil)
		// Runtime probe for REQ-MW-1: a deliberate handler panic that the
		// platform recovery middleware must convert into a 500 problem+json
		// without killing the process. Mounted only when METALDOCS_E2E=1;
		// touches no data.
		m.HandleFunc("GET /internal/test/panic", func(http.ResponseWriter, *http.Request) {
			panic("e2e panic probe: must be recovered by platform/middleware.Recovery (REQ-MW-1)")
		})
	})

	leaderID := schedulerLeaderID()
	s, err := jobscheduler.New(deps.SQLDB, leaderID, slog.Default())
	if err != nil {
		slog.Error("jobs scheduler configuration failed", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	registerScheduledJobs(s, deps, approvalServices.Cancel, approvalEmitter)
	httpObs.SetSchedulerMetrics(s)
	if deps.SQLDB != nil {
		httpObs.SetDBPool(postgres.NewPoolStatsAdapter(deps.SQLDB))
	}

	var schedulerWG sync.WaitGroup
	schedulerWG.Add(1)
	go func() {
		defer schedulerWG.Done()
		s.Start(ctx)
	}()

	stopSessions := jobs.StartSessionSweeper(ctx, docMod.Repo(), 60*time.Second)
	stopOrphans := jobs.StartOrphanPendingSweeper(ctx, docMod.Repo(), time.Hour, 24*time.Hour)
	defer stopSessions()
	defer stopOrphans()
	mux.Handle("/api/v1/metrics", httpObs.MetricsHandler())

	// Audit retention - AUDIT_RETENTION_DAYS=0 disables (default disabled).
	retentionCfg, err := config.LoadRetentionConfig()
	if err != nil {
		slog.Error("invalid retention config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	startAuditRetention(ctx, deps, retentionCfg.Days)

	// Canonical chain per backend-target-architecture.md §2.1 (F-01 fix,
	// REQ-MW-1/2/4/5): panic recovery + trace context outermost, access
	// log/metrics OUTSIDE authn so 401s and panics are observable,
	// pre-auth IP-keyed login limit before authn. presenceBump stays
	// after iamMiddleware (needs iamdomain.UserID in ctx, PR-9). Order is
	// asserted by chain_test.go (REQ-MW-7).
	var presenceWrap func(http.Handler) http.Handler
	if presenceBump != nil {
		presenceWrap = presenceBump.Wrap
	}
	// otel link is nil (skipped by buildChain) unless an exporter is configured;
	// recovery stays outermost, otel wraps everything else (Z-1, REQ-OBS-1).
	var otelWrap func(http.Handler) http.Handler
	if otelEnabled {
		otelWrap = observability.OTelMiddleware()
	}
	handler := buildChain(mux, apiChain(
		platformmw.Recovery,
		otelWrap,
		httpObs.Wrap,
		cors.Wrap,
		originProtection.Wrap,
		loginRateLimit(preAuthLimiter),
		authMiddleware.Wrap,
		iamMiddleware.Wrap,
		presenceWrap,
		func(next http.Handler) http.Handler { return globalLimiter.GlobalEnvelopeWrap(userIDExtractor, next) },
		// Innermost (nearest the mux): rewrite the stdlib text/plain 404/405 the
		// method-routed ServeMux emits into problem+json, preserving Allow (D-03).
		platformmw.MethodNotAllowedJSON,
	))

	serverCfg, err := config.LoadServerConfig()
	if err != nil {
		slog.Error("invalid server config", "err", err)
		deps.Cleanup()
		os.Exit(1)
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

	// Z-22 / REQ-REL-2: drain live WS presence connections before the
	// HTTP server stops accepting. RegisterOnShutdown runs synchronously
	// inside server.Shutdown after the listener is closed but before
	// Shutdown returns, so the shutdown context (15s) covers the drain.
	if presenceHub != nil {
		server.RegisterOnShutdown(func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			presenceHub.CloseAll(shutdownCtx)
		})
	}

	slog.Info("MetalDocs API listening",
		"addr", serverCfg.Addr, "repository", repoMode, "auth_enabled", authn.Enabled(),
		"auth_cache_ttl", authn.CacheTTL(), "cors_enabled", corsCfg.Enabled,
		"cors_allowed_origins", len(corsCfg.AllowedOrigins))

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	exitCode := shutdownServer(ctx, stop, server, serverErr, &schedulerWG, &workerWG)
	if exitCode != 0 {
		// os.Exit skips deferred functions, including deps.Cleanup. Invoke
		// cleanup explicitly so DB / object-store handles are released on
		// the error path too. closeDB swallows close-on-closed, so calling
		// twice is safe.
		deps.Cleanup()
		os.Exit(exitCode)
	}
}

// shutdownServer waits for either a server-listen error or ctx cancellation,
// then runs the same graceful-teardown sequence on both paths: server.Shutdown,
// stop signal handler, scheduler join, worker join. Returns a non-zero exit
// code only if a real failure occurred (genuine ListenAndServe error or a
// Shutdown that didn't drain cleanly).
func shutdownServer(
	ctx context.Context,
	stop context.CancelFunc,
	server *http.Server,
	serverErr <-chan error,
	schedulerWG, workerWG *sync.WaitGroup,
) int {
	exitCode := 0
	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			exitCode = 1
		}
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown incomplete", "err", err)
		if exitCode == 0 {
			exitCode = 1
		}
	} else {
		slog.Info("graceful shutdown complete")
	}
	stop()
	schedulerWG.Wait()
	workerWG.Wait()
	slog.Info("scheduler stopped")
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
	})
}

func buildControlledDocumentsModule(deps bootstrap.APIDependencies) *controlleddocuments.Module {
	return controlleddocuments.New(controlleddocuments.Dependencies{
		DB:          deps.SQLDB,
		Logger:      slog.Default(),
		AuditWriter: deps.AuditWriter,
	})
}

func buildTemplatesModule(deps bootstrap.APIDependencies, capabilityService *iamapp.CapabilityService) (*templateshttp.Handler, error) {
	if capabilityService == nil {
		return nil, errors.New("templates capability service is required")
	}
	templatesPresigner := objectstore.NewTemplatesPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 25*1024*1024)
	templatesSvc := templatesapp.New(templatesrepo.New(deps.SQLDB).WithAudit(deps.AuditWriter), templatesPresigner, wiring.Clock{}, wiring.UUIDGen{}).WithRunner(db.NewTxRunner(deps.SQLDB))
	templatesAuthzFn := func(r *http.Request, tenantID, _ string, action string) error {
		userID := iamdomain.UserIDFromContext(r.Context())
		return capabilityService.CanDo(r.Context(), userID, tenantID, action)
	}
	return templateshttp.New(templatesSvc, templatesAuthzFn, deps.SQLDB), nil
}

func requireApprovalRuntimeSupport(fanoutURL string) error {
	if strings.TrimSpace(fanoutURL) == "" {
		return errors.New("approval runtime requires METALDOCS_FANOUT_URL; startup without freeze support is not allowed")
	}
	return nil
}

func jobEnabled(envName string) bool {
	return !strings.EqualFold(strings.TrimSpace(os.Getenv(envName)), "false")
}

func schedulerLeaderID() string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s:%d", hostname, os.Getpid())
}

// startPresence initialises the PR-9 presence subsystem: hub goroutines, WebSocket
// handler, HTTP snapshot fallback, bump middleware with cleanup, and wires the
// presence reader into iamAdminHandler. Returns (nil, nil) when deps.SQLDB is nil
// (in-memory mode). The returned Hub is captured by main for shutdown drain
// (Z-22, REQ-REL-2); the BumpMiddleware is wrapped into the outer request chain
// so authenticated requests refresh last_seen_at (debounced 60s per user, PR-9).
func startPresence(
	ctx context.Context,
	deps bootstrap.APIDependencies,
	mux *http.ServeMux,
	iamAdminHandler *iamdelivery.AdminHandler,
) (*iampresence.BumpMiddleware, *iampresence.Hub) {
	if deps.SQLDB == nil {
		return nil, nil
	}
	presenceRepo := iampresence.NewPostgresRepository(deps.SQLDB)
	presenceHub := iampresence.NewHub(presenceRepo, slog.Default())
	go presenceHub.Run(ctx)
	go presenceHub.RunHeartbeat(ctx)
	iampresence.NewHandler(presenceHub, presenceRepo, slog.Default()).RegisterRoutes(mux)
	presenceBump := iampresence.NewBumpMiddleware(presenceRepo, slog.Default())
	presenceBump.StartCleanup(ctx)
	iamAdminHandler.WithPresenceReader(presenceRepo)
	return presenceBump, presenceHub
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
		snapRepo, snapRepo, client,
	)
	return fanoutComponents{client: client, freezeService: freezeService}
}

// startOutboxWorkers builds and starts the PDF and materialize staging outbox
// workers. Each worker polls its outbox table and publishes events via publisher;
// goroutine lifetimes are tracked in wg so shutdownServer can join them cleanly.
// StagingOutboxWorker.Run() returns only on ctx cancellation (nil error);
// no restart loop is needed.
func startOutboxWorkers(
	ctx context.Context,
	wg *sync.WaitGroup,
	publisher messaging.Publisher,
	pdfOutboxRepo, materializeOutboxRepo *fanout.StagingOutboxRepository,
) {
	start := func(w *fanout.StagingOutboxWorker) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = w.Run(ctx)
		}()
	}

	pdfOutboxWorker := fanout.NewStagingOutboxWorker(pdfOutboxRepo, publisher, func(r fanout.OutboxRow) messaging.Event {
		return messaging.Event{
			EventID:        messaging.EventID(uuid.NewString()),
			EventType:      messaging.EventTypePDFConvert,
			AggregateType:  messaging.AggregateType("document_revision"),
			AggregateID:    messaging.AggregateID(r.RevisionID),
			IdempotencyKey: messaging.IdempotencyKey("docgen_v2_pdf:" + r.TenantID + ":" + r.RevisionID),
			Payload: messaging.PDFConvertPayload{
				TenantID:    r.TenantID,
				RevisionID:  r.RevisionID,
				ContentHash: hex.EncodeToString(r.ContentHash),
			},
		}
	}, slog.Default())
	start(pdfOutboxWorker)

	materializeOutboxWorker := fanout.NewStagingOutboxWorker(materializeOutboxRepo, publisher, func(r fanout.OutboxRow) messaging.Event {
		return messaging.Event{
			EventID:        messaging.EventID(uuid.NewString()),
			EventType:      messaging.EventTypeMaterializeFanout,
			AggregateType:  messaging.AggregateType("document_revision"),
			AggregateID:    messaging.AggregateID(r.RevisionID),
			IdempotencyKey: messaging.IdempotencyKey("materialize_fanout:" + r.TenantID + ":" + r.RevisionID),
			Payload: messaging.MaterializeFanoutPayload{
				TenantID:   r.TenantID,
				RevisionID: r.RevisionID,
			},
		}
	}, slog.Default())
	start(materializeOutboxWorker)
}

// registerScheduledJobs registers the four optional background jobs with the
// scheduler. Each job is gated on its ENABLE_JOB_* env var (default enabled).
// The audit-integrity job additionally requires deps.AuditValidator to be non-nil.
// Scheduler startup and the schedulerWG goroutine remain in main.
func registerScheduledJobs(
	s *jobscheduler.Scheduler,
	deps bootstrap.APIDependencies,
	cancelSvc *approvalapp.CancelService,
	emitter approvalapp.EventEmitter,
) {
	if jobEnabled("ENABLE_JOB_STUCK_INSTANCE_WATCHDOG") {
		s.Register(jobscheduler.JobConfig{
			Name:     "stuck-instance-watchdog",
			Interval: 5 * time.Minute,
			Fn:       stuck_instance_watchdog.New(deps.SQLDB, cancelSvc, emitter),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
	if jobEnabled("ENABLE_JOB_IDEMPOTENCY_JANITOR") {
		s.Register(jobscheduler.JobConfig{
			Name:     "idempotency-janitor",
			Interval: 15 * time.Minute,
			Fn:       idempotency_janitor.New(deps.SQLDB),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
	if jobEnabled("ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR") && deps.AuditValidator != nil {
		s.Register(jobscheduler.JobConfig{
			Name:     "audit-integrity-validator",
			Interval: time.Hour,
			Fn:       audit_integrity_validator.New(deps.AuditValidator),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
	if jobEnabled("ENABLE_JOB_LEASE_REAPER") {
		s.Register(jobscheduler.JobConfig{
			Name:     "lease-reaper",
			Interval: 10 * time.Minute,
			Fn:       jobscheduler.RunLeaseReaper(deps.SQLDB),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
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
