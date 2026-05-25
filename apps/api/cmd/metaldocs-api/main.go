package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	"metaldocs/apps/api/internal/wiring"
	auditapp "metaldocs/internal/modules/audit/application"
	auditdelivery "metaldocs/internal/modules/audit/delivery/http"
	auditdomain "metaldocs/internal/modules/audit/domain"
	authapp "metaldocs/internal/modules/auth/application"
	authdelivery "metaldocs/internal/modules/auth/delivery/http"
	controlleddocuments "metaldocs/internal/modules/controlleddocuments"
	controlleddocumentsapp "metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	controlleddocumentsinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	documents "metaldocs/internal/modules/documents"
	docapp "metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	approvalhttp "metaldocs/internal/modules/documents/approval/http"
	approvalinfra "metaldocs/internal/modules/documents/approval/infrastructure"
	approvaljobs "metaldocs/internal/modules/documents/approval/jobs"
	approvalrepo "metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/modules/documents/jobs"
	docrepo "metaldocs/internal/modules/documents/repository"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/modules/jobs/audit_integrity_validator"
	"metaldocs/internal/modules/jobs/idempotency_janitor"
	jobscheduler "metaldocs/internal/modules/jobs/scheduler"
	"metaldocs/internal/modules/jobs/stuck_instance_watchdog"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	searchapp "metaldocs/internal/modules/search/application"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	searchdocs "metaldocs/internal/modules/search/infrastructure/v2documents"
	"metaldocs/internal/modules/taxonomy"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	templatesapp "metaldocs/internal/modules/templates/application"
	templateshttp "metaldocs/internal/modules/templates/delivery/http"
	templatesrepo "metaldocs/internal/modules/templates/repository"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/bootstrap"
	"metaldocs/internal/platform/config"
	docgenv2 "metaldocs/internal/platform/docgenv2"
	"metaldocs/internal/platform/featureflags"
	"metaldocs/internal/platform/formval"
	"metaldocs/internal/platform/httpclient"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/internal/platform/migrate"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/security"
	e2etest "metaldocs/internal/test"
)

type ControlledDocumentService interface {
	Get(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
	Create(ctx context.Context, cmd controlleddocumentsapp.CreateControlledDocumentCmd) (*controlleddocumentsapp.CreateResult, error)
}

type controlledDocumentDuplicatorAdapter struct {
	svc ControlledDocumentService
}

func newControlledDocumentDuplicatorAdapter(svc ControlledDocumentService) *controlledDocumentDuplicatorAdapter {
	if svc == nil {
		panic("newControlledDocumentDuplicatorAdapter: nil service")
	}
	return &controlledDocumentDuplicatorAdapter{svc: svc}
}

func (a controlledDocumentDuplicatorAdapter) DuplicateControlledDocument(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (*controlleddocumentsdomain.ControlledDocument, error) {
	source, err := a.svc.Get(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return nil, fmt.Errorf("duplicate controlled document %s: %w", controlledDocumentID, err)
	}

	var overrideReason *string
	if source.OverrideTemplateVersionID != nil {
		reason := "Duplicated from existing controlled document"
		overrideReason = &reason
	}

	res, err := a.svc.Create(ctx, controlleddocumentsapp.CreateControlledDocumentCmd{
		TenantID:                  tenantID,
		ProfileCode:               source.ProfileCode,
		ProcessAreaCode:           source.ProcessAreaCode,
		DepartmentCode:            source.DepartmentCode,
		Title:                     source.Title,
		OwnerUserID:               source.OwnerUserID,
		ActorUserID:               actorUserID,
		OverrideTemplateVersionID: source.OverrideTemplateVersionID,
		OverrideTemplateReason:    overrideReason,
	})
	if err != nil {
		return nil, fmt.Errorf("duplicate controlled document %s: %w", controlledDocumentID, err)
	}
	return res.ControlledDocument, nil
}

// e2eHandlersEnabled gates test-only seed/reset endpoints behind an explicit env flag.
func e2eHandlersEnabled() bool {
	return strings.TrimSpace(os.Getenv("METALDOCS_E2E")) == "1"
}

func mountE2EHandlersIfEnabled(mux *http.ServeMux, register func(*http.ServeMux)) {
	if !e2eHandlersEnabled() {
		slog.Info("e2etest handlers not mounted", "reason", "METALDOCS_E2E != 1")
		return
	}
	slog.Warn("e2etest handlers mounted - destructive endpoints reachable without auth", "env", "METALDOCS_E2E=1")
	register(mux)
}

type Config struct {
	RepoMode     string
	RateCfg      config.RateLimitConfig
	CORSCfg      config.CORSConfig
	Attachments  config.AttachmentsConfig
	AuthCfg      authapp.Config
	FeatureFlags config.FeatureFlagsConfig
	JobsCfg      config.JobsConfig
}

type Deps struct {
	bootstrap.APIDependencies
	JobsCfg          config.JobsConfig
	ApprovalServices *approvalapp.Services
	ApprovalEmitter  approvalapp.EventEmitter
	WorkerWG         *sync.WaitGroup
	SchedulerWG      *sync.WaitGroup
	StopSessions     func()
	StopOrphans      func()
	Handler          http.Handler
}

type Scheduler = jobscheduler.Scheduler

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	deps, err := buildDeps(ctx, cfg)
	if err != nil {
		log.Fatalf("build api dependencies: %v", err)
	}
	defer deps.Cleanup()

	if err := migrateStartup(ctx, deps); err != nil {
		log.Fatalf("apply startup migrations: %v", err)
	}

	mux := http.NewServeMux()
	if err := registerRoutes(ctx, mux, &deps, cfg); err != nil {
		log.Fatal(err)
	}

	scheduler, err := jobscheduler.New(deps.SQLDB, schedulerLeaderID())
	if err != nil {
		log.Fatalf("jobs scheduler configuration failed: %v", err)
	}
	registerScheduledJobs(scheduler, deps)

	deps.SchedulerWG.Add(1)
	go func() {
		defer deps.SchedulerWG.Done()
		scheduler.Start(ctx)
	}()
	defer deps.StopSessions()
	defer deps.StopOrphans()

	addr, err := resolveServerAddr()
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           deps.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("MetalDocs API listening on %s (repository=%s auth_enabled=%t auth_cache_ttl=%s rate_limit_enabled=%t rate_limit_window_s=%d rate_limit_max_requests=%d cors_enabled=%t cors_allowed_origins=%d)",
		addr, cfg.RepoMode, authn.Enabled(), authn.CacheTTL(), cfg.RateCfg.Enabled, cfg.RateCfg.WindowSeconds, cfg.RateCfg.MaxRequests, cfg.CORSCfg.Enabled, len(cfg.CORSCfg.AllowedOrigins))

	serverErr := make(chan error, 1)
	go func() { serverErr <- startAndWait(ctx, server) }()

	exitCode := shutdownServer(ctx, stop, server, serverErr, deps.SchedulerWG, deps.WorkerWG)
	if exitCode != 0 {
		deps.Cleanup()
		os.Exit(exitCode)
	}
}

func loadConfig() (Config, error) {
	repoMode, err := config.RepositoryMode()
	if err != nil {
		return Config{}, fmt.Errorf("invalid repository mode: %w", err)
	}
	rateCfg, err := config.LoadRateLimitConfig()
	if err != nil {
		return Config{}, fmt.Errorf("invalid rate limit config: %w", err)
	}
	corsCfg, err := config.LoadCORSConfig()
	if err != nil {
		return Config{}, fmt.Errorf("invalid cors config: %w", err)
	}
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		return Config{}, fmt.Errorf("invalid attachments config: %w", err)
	}
	authCfg, err := authn.LoadRuntimeConfig()
	if err != nil {
		return Config{}, fmt.Errorf("invalid auth config: %w", err)
	}
	featureFlagsCfg, err := config.LoadFeatureFlagsConfig()
	if err != nil {
		return Config{}, fmt.Errorf("invalid feature flags config: %w", err)
	}
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		return Config{}, fmt.Errorf("invalid jobs config: %w", err)
	}

	return Config{
		RepoMode:     repoMode,
		RateCfg:      rateCfg,
		CORSCfg:      corsCfg,
		Attachments:  attachmentsCfg,
		AuthCfg:      authCfg,
		FeatureFlags: featureFlagsCfg,
		JobsCfg:      jobsCfg,
	}, nil
}

func buildDeps(ctx context.Context, cfg Config) (Deps, error) {
	baseDeps, err := bootstrap.BuildAPIDependencies(ctx, cfg.RepoMode, cfg.Attachments)
	if err != nil {
		return Deps{}, err
	}
	return Deps{
		APIDependencies: baseDeps,
		JobsCfg:         cfg.JobsCfg,
		WorkerWG:        &sync.WaitGroup{},
		SchedulerWG:     &sync.WaitGroup{},
		StopSessions:    func() {},
		StopOrphans:     func() {},
	}, nil
}

func migrateStartup(ctx context.Context, deps Deps) error {
	if deps.SQLDB == nil || strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_SKIP_STARTUP_MIGRATIONS")), "true") {
		return nil
	}
	migrationsDir := strings.TrimSpace(os.Getenv("METALDOCS_MIGRATIONS_DIR"))
	if migrationsDir == "" {
		migrationsDir = "db/migrations"
	}
	return migrate.Apply(ctx, deps.SQLDB, migrationsDir, slog.Default())
}

func registerRoutes(ctx context.Context, mux *http.ServeMux, deps *Deps, cfg Config) error {
	authService, err := authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, cfg.AuthCfg)
	if err != nil {
		return fmt.Errorf("new auth service: %w", err)
	}
	if err := authService.BootstrapLocalAdmin(ctx); err != nil {
		return fmt.Errorf("bootstrap local admin: %w", err)
	}

	auditService := auditapp.NewService(deps.AuditReader)
	auditHandler := auditdelivery.NewHandler(auditService)
	searchService := searchapp.NewService(searchdocs.NewReader(deps.SQLDB))
	searchHandler := searchdelivery.NewHandler(searchService)
	authHandler := authdelivery.NewHandler(authService).WithAudit(deps.AuditWriter)
	healthHandler := observability.NewHealthHandler(deps.StatusProvider)

	var capabilityService *iamapp.CapabilityService
	if deps.SQLDB != nil {
		capabilityService = iamapp.NewCapabilityService(deps.SQLDB)
	}
	cachedProvider := iamapp.NewCachedRoleProvider(ctx, deps.RoleProvider, authn.CacheTTL())
	permResolver := newPermissionResolver()
	authMiddleware := authdelivery.NewMiddleware(authService, cfg.AuthCfg, authn.Enabled()).
		WithPublicPathChecker(newPublicPathChecker(permResolver))
	iamMiddleware := iamdelivery.NewMiddleware(capabilityService, cachedProvider, authn.Enabled(), cfg.AuthCfg.LegacyHeaderEnabled).
		WithPermissionResolver(permResolver)
	originProtection := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           cfg.AuthCfg.OriginProtection,
		SessionCookieName: cfg.AuthCfg.SessionCookieName,
		TrustedOrigins:    cfg.AuthCfg.TrustedOrigins,
		TrustedProxyCIDRs: cfg.AuthCfg.TrustedProxyCIDRs,
	})

	iamAdminService := iamapp.NewAdminService(deps.RoleAdminRepo, cachedProvider)
	iamAdminHandler := iamdelivery.NewAdminHandler(iamAdminService, authService, deps.AuditWriter).
		WithAuditReader(deps.AuditReader)
	featureFlagsHandler := featureflags.NewHandler(cfg.FeatureFlags)
	httpObs := observability.NewHTTPObservability(deps.StatusProvider)
	rateLimiter := security.NewRateLimiter(cfg.RateCfg)
	cors := security.NewCORS(cfg.CORSCfg)

	authHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)
	featureFlagsHandler.RegisterRoutes(mux)
	auditHandler.RegisterRoutes(mux)
	searchHandler.RegisterRoutes(mux)
	iamAdminHandler.RegisterRoutes(mux)

	taxonomyModule := taxonomy.New(taxonomy.Dependencies{
		DB:          deps.SQLDB,
		TplChecker:  taxonomyinfra.NewTemplateVersionChecker(deps.SQLDB),
		AuditWriter: deps.AuditWriter,
	})
	taxonomyModule.RegisterRoutes(mux)

	controlledDocumentsModule := controlleddocuments.New(controlleddocuments.Dependencies{
		DB:          deps.SQLDB,
		Logger:      slog.Default(),
		AuditWriter: deps.AuditWriter,
	})
	controlledDocumentsModule.RegisterRoutes(mux)
	controlledDocumentDuplicator := newControlledDocumentDuplicatorAdapter(controlledDocumentsModule.Service())

	var membershipService *iamapp.AreaMembershipService
	if deps.SQLDB != nil {
		membershipService = iamapp.NewAreaMembershipService(iampg.NewUserAreaRepository(deps.SQLDB), nil)
	}
	iamdelivery.NewMembershipHandler(membershipService).RegisterRoutes(mux)

	docPresigner := objectstore.NewDocumentPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 15*time.Minute, 25*1024*1024)
	controlledDocumentsRepo := controlleddocumentsinfra.NewPostgresControlledDocumentRepository(deps.SQLDB)
	profileRepo := taxonomyinfra.NewProfileRepository(deps.SQLDB)

	fanoutURL := strings.TrimSpace(os.Getenv("METALDOCS_FANOUT_URL"))
	if fanoutURL == "" {
		if strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_REQUIRE_FANOUT")), "true") {
			return fmt.Errorf("METALDOCS_FANOUT_URL is required but not set")
		}
		slog.Warn("METALDOCS_FANOUT_URL not set; document approval will fail at freeze step")
	}
	serviceToken := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN"))
	if fanoutURL != "" && serviceToken == "" {
		return fmt.Errorf("METALDOCS_DOCGEN_V2_SERVICE_TOKEN is required when METALDOCS_FANOUT_URL is set")
	}

	var fanoutCli *fanout.Client
	var freezeSvc *docapp.FreezeService
	var pdfDispatchAdapter approvalapp.PDFDispatchInvoker
	if fanoutURL != "" && deps.SQLDB != nil {
		fanoutCli = fanout.NewClient(fanoutURL, serviceToken, httpclient.NewInternalClient())
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		fillInRepo := docrepo.NewFillInRepository(deps.SQLDB)
		schemaReader := docapp.NewSnapshotSchemaReader(deps.SQLDB)
		revReader := docrepo.NewRevisionReader(deps.SQLDB)
		wfReader := docrepo.NewWorkflowReader(deps.SQLDB)
		ctxBuilder := docapp.NewDocumentContextBuilder(deps.SQLDB, revReader, wfReader,
			controlledDocumentsReaderAdapter{controlledDocumentsRepo}, revReader)
		resolverReg := resolvers.NewRegistry()
		resolvers.RegisterBuiltins(resolverReg)
		freezeSvc = docapp.NewFreezeService(
			schemaReader, fillInRepo, fillInRepo,
			resolverReg, snapRepo, ctxBuilder,
			snapRepo, snapRepo, fillInRepo, fanoutCli,
		)
		if deps.Publisher != nil {
			pdfDispatcher := fanout.NewPDFDispatcher(deps.Publisher)
			pdfDispatchAdapter = fanout.NewPDFDispatchAdapter(pdfDispatcher, snapRepo)
		}
	}

	docSnapshotReader := docgenv2.NewTemplatesSnapshotReader(deps.SQLDB)
	docSnapshotWriter := docrepo.NewSnapshotRepository(deps.SQLDB)
	docDeps := documents.Dependencies{
		DB:      deps.SQLDB,
		Docgen:  nil,
		Presign: docPresigner,
		TplRead: docgenv2.NewFanoutTemplateReader(
			docgenv2.NewTemplateReader(deps.SQLDB, deps.MinioClient, deps.MinioBucket),
			docgenv2.NewTemplatesTemplateReader(deps.SQLDB),
		),
		FormVal:                      formval.NewGojsonschema(),
		Audit:                        newDocumentsAuditAdapter(deps.AuditWriter),
		ExportPresign:                docPresigner,
		ControlledDocumentReader:     controlledDocumentsRepo,
		ControlledDocumentDuplicator: controlledDocumentDuplicator,
		Caps:                         wiring.NewCapabilityChecker(capabilityService),
		ProfileDefaults:              &profileDefaultsAdapter{profileRepo: profileRepo},
		SnapshotReader:               docSnapshotReader,
		SnapshotWriter:               docSnapshotWriter,
	}
	if deps.DocgenV2Client != nil {
		docDeps.ExportDocgen = deps.DocgenV2Client
	}
	if fanoutCli != nil && deps.SQLDB != nil {
		snapRepo := docrepo.NewSnapshotRepository(deps.SQLDB)
		inputsReader := docrepo.NewFanoutInputsReader(deps.SQLDB)
		docDeps.ReconstructRunner = fanout.NewReconstructService(
			inputsReader, fanoutCli, snapRepo,
			fanout.EngineVersions{EigenpalVer: "local", DocxtemplaterVer: "local"},
			nil,
		)
	}

	approvalRepo := approvalrepo.NewPostgresApprovalRepository(deps.SQLDB)
	approvalEmitter := approvalapp.NewSQLEmitter()
	approvalServices := approvalapp.NewServices(approvalRepo, approvalEmitter, approvalapp.RealClock{})
	if deps.SQLDB != nil {
		if err := bootstrap.MigrateRiverSchema(ctx, deps.SQLDB, deps.JobsCfg.RiverSchema); err != nil {
			return fmt.Errorf("migrate river schema: %w", err)
		}
		riverBundle, err := riverjobs.NewClientBundle(deps.SQLDB, riverjobs.Config{
			Queues:              deps.JobsCfg.Queues,
			Schema:              deps.JobsCfg.RiverSchema,
			SkipUnknownJobCheck: true,
		}, nil)
		if err != nil {
			return fmt.Errorf("build scheduled publish enqueuer client: %w", err)
		}
		approvalServices.WithScheduledPublishEnqueuer(approvaljobs.NewScheduledPublishEnqueuer(riverBundle.Client))
	}

	var effectiveFreezeInvoker approvalapp.FreezeInvoker = noopFreezeInvoker{}
	if freezeSvc != nil {
		effectiveFreezeInvoker = freezeSvc
	}
	pdfOutboxRepo := fanout.NewPDFOutboxRepository(deps.SQLDB)
	pdfOutboxWorker := fanout.NewPDFOutboxWorker(pdfOutboxRepo, deps.Publisher, slog.Default())
	deps.WorkerWG.Add(1)
	go func() {
		defer deps.WorkerWG.Done()
		backoff := time.Second
		for ctx.Err() == nil {
			err := pdfOutboxWorker.Run(ctx)
			if err == nil {
				return
			}
			slog.Error("pdf outbox worker exited; restarting", "err", err, "backoff", backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < time.Minute {
				backoff *= 2
				if backoff > time.Minute {
					backoff = time.Minute
				}
			}
		}
	}()
	approvalServices.Decision = approvalapp.NewDecisionService(
		approvalRepo, approvalEmitter, approvalapp.RealClock{}, effectiveFreezeInvoker, pdfDispatchAdapter,
	).WithPDFOutbox(pdfOutboxRepo)
	docDeps.SubmitSvc = approvalServices.Submit

	docMod := documents.New(docDeps)
	docMod.RegisterRoutes(mux)
	controlledDocumentsModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))

	templatesPresigner := objectstore.NewTemplatesPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 25*1024*1024)
	templatesSvc := templatesapp.New(templatesrepo.New(deps.SQLDB).WithAudit(deps.AuditWriter), templatesPresigner, realClock{}, realUUIDGen{}).WithDB(deps.SQLDB)
	templatesAuthzFn := func(r *http.Request, tenantID, _ string, action string) error {
		userID := iamdomain.UserIDFromContext(r.Context())
		return capabilityService.CanDo(r.Context(), userID, tenantID, action)
	}
	templateshttp.New(templatesSvc, templatesAuthzFn).Register(mux)

	signoffIdempStore := approvalinfra.NewPostgresSignoffIdempStore(deps.SQLDB)
	approvalHandler := approvalhttp.NewHandler(approvalServices, deps.SQLDB, signoffIdempStore)
	approvalHandler.RegisterRoutes(mux)

	mountE2EHandlersIfEnabled(mux, func(m *http.ServeMux) {
		e2etest.RegisterE2EHandlers(m, deps.SQLDB, nil)
	})

	deps.ApprovalServices = approvalServices
	deps.ApprovalEmitter = approvalEmitter
	deps.StopSessions = jobs.StartSessionSweeper(ctx, docMod.Repo(), 60*time.Second)
	deps.StopOrphans = jobs.StartOrphanPendingSweeper(ctx, docMod.Repo(), time.Hour)
	mux.Handle("/api/v1/metrics", httpObs.MetricsHandler())

	if retentionDays, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("AUDIT_RETENTION_DAYS"))); retentionDays > 0 && deps.SQLDB != nil {
		go func() {
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
					if _, err := deps.SQLDB.ExecContext(ctx,
						`DELETE FROM metaldocs.audit_events WHERE occurred_at < $1`, cutoff,
					); err != nil {
						slog.Warn("audit retention purge failed", "error", err)
					}
				}
			}
		}()
	}

	deps.Handler = cors.Wrap(originProtection.Wrap(authMiddleware.Wrap(iamMiddleware.Wrap(httpObs.Wrap(rateLimiter.Wrap(mux))))))
	return nil
}

func registerScheduledJobs(scheduler *Scheduler, deps Deps) {
	if jobEnabled("ENABLE_JOB_STUCK_INSTANCE_WATCHDOG") {
		scheduler.Register(jobscheduler.JobConfig{
			Name:     "stuck-instance-watchdog",
			Interval: 5 * time.Minute,
			Fn:       stuck_instance_watchdog.New(deps.SQLDB, deps.ApprovalServices.Cancel, deps.ApprovalEmitter),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
	if jobEnabled("ENABLE_JOB_IDEMPOTENCY_JANITOR") {
		scheduler.Register(jobscheduler.JobConfig{
			Name:     "idempotency-janitor",
			Interval: 15 * time.Minute,
			Fn:       idempotency_janitor.New(deps.SQLDB),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
	if jobEnabled("ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR") && deps.AuditValidator != nil {
		scheduler.Register(jobscheduler.JobConfig{
			Name:     "audit-integrity-validator",
			Interval: time.Hour,
			Fn:       audit_integrity_validator.New(deps.AuditValidator),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
	if jobEnabled("ENABLE_JOB_LEASE_REAPER") {
		scheduler.Register(jobscheduler.JobConfig{
			Name:     "lease-reaper",
			Interval: 10 * time.Minute,
			Fn:       jobscheduler.RunLeaseReaper(deps.SQLDB),
			Policy:   jobscheduler.SkipOnPressure,
		})
	}
}

func resolveServerAddr() (string, error) {
	addr := ":8080"
	if appPort := os.Getenv("APP_PORT"); appPort != "" {
		port, convErr := strconv.Atoi(strings.TrimSpace(appPort))
		if convErr != nil || port < 1 || port > 65535 {
			return "", fmt.Errorf("invalid APP_PORT value")
		}
		addr = ":" + strconv.Itoa(port)
	}
	return addr, nil
}

func startAndWait(_ context.Context, srv *http.Server) error {
	return srv.ListenAndServe()
}

// shutdownServer waits for either a server-listen error or ctx cancellation,
// then runs the same graceful-teardown sequence on both paths: server.Shutdown,
// stop signal handler, scheduler join, worker join. Returns a non-zero exit
// code only if a real failure occurred (genuine ListenAndServe error or a
// Shutdown that did not drain cleanly).
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

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// noopFreezeInvoker is used when METALDOCS_FANOUT_URL is unset.
// Approval completes locally without calling the fanout service.
type noopFreezeInvoker struct{}

func (noopFreezeInvoker) Freeze(_ context.Context, _ *sql.Tx, _, _ string, _ docapp.ApproverContext) error {
	slog.Warn("freeze skipped: METALDOCS_FANOUT_URL not configured")
	return nil
}

type realUUIDGen struct{}

func (realUUIDGen) New() string { return uuid.NewString() }

type documentsAuditAdapter struct {
	writer auditdomain.Writer
}

func newDocumentsAuditAdapter(writer auditdomain.Writer) *documentsAuditAdapter {
	if writer == nil {
		panic("documents audit writer is nil")
	}
	return &documentsAuditAdapter{writer: writer}
}

func (a *documentsAuditAdapter) WriteTx(ctx context.Context, tx *sql.Tx, tenantID, actorID, action, docID string, meta any) error {
	payload := map[string]any{"tenant_id": tenantID}
	if meta != nil {
		payload["meta"] = meta
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}
	event, err := auditdomain.NewEvent(tenantID, actorID, action, "document", docID, time.Now().UTC())
	if err != nil {
		return err
	}
	event.ID = uuid.NewString()
	event.PayloadJSON = string(raw)
	event.TraceID = traceIDFromContext(ctx)
	return a.writer.RecordTx(ctx, tx, event)
}

func (a *documentsAuditAdapter) Write(ctx context.Context, tenantID, actorID, action, docID string, meta any) {
	payload := map[string]any{"tenant_id": tenantID}
	if meta != nil {
		payload["meta"] = meta
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte("{}")
	}

	event, err := auditdomain.NewEvent(tenantID, actorID, action, "document", docID, time.Now().UTC())
	if err != nil {
		log.Printf("documents audit event invalid: %v", err)
		return
	}
	event.ID = uuid.NewString()
	event.PayloadJSON = string(raw)
	event.TraceID = traceIDFromContext(ctx)

	if err := a.writer.Record(ctx, event); err != nil {
		log.Printf("documents audit write failed: %v", err)
	}
}

func traceIDFromContext(_ context.Context) string {
	// Observability middleware stores the trace ID in slog context, not in context values.
	return uuid.NewString()
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

// controlledDocumentsReaderAdapter bridges the controlled-document repository
// to resolvers.RegistryReader.
type controlledDocumentsReaderAdapter struct {
	repo interface {
		GetByID(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ControlledDocument, error)
	}
}

func (a controlledDocumentsReaderAdapter) GetControlledDocument(ctx context.Context, tenantID, controlledDocumentID string) (resolvers.ControlledDocumentInfo, error) {
	cd, err := a.repo.GetByID(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return resolvers.ControlledDocumentInfo{}, err
	}
	return resolvers.ControlledDocumentInfo{DocCode: cd.Code}, nil
}

// profileDefaultsAdapter bridges taxonomy ProfileRepository to documents ProfileDefaultTemplateReader.
type profileDefaultsAdapter struct {
	profileRepo interface {
		GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)
	}
}

func (a *profileDefaultsAdapter) GetDefaultTemplateVersionID(ctx context.Context, tenantID, profileCode string) (*string, *string, error) {
	profile, err := a.profileRepo.GetByCode(ctx, tenantID, profileCode)
	if err != nil {
		return nil, nil, err
	}
	if profile.DefaultTemplateVersionID == nil {
		return nil, nil, nil
	}
	status := "published"
	return profile.DefaultTemplateVersionID, &status, nil
}
