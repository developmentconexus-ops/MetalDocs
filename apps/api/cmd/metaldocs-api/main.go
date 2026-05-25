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
	controlleddocuments "metaldocs/internal/modules/controlleddocuments"
	controlleddocumentsapp "metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	controlleddocumentsinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	iamapp "metaldocs/internal/modules/iam/application"
	iamdelivery "metaldocs/internal/modules/iam/delivery/http"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	searchapp "metaldocs/internal/modules/search/application"
	searchdelivery "metaldocs/internal/modules/search/delivery/http"
	searchdocs "metaldocs/internal/modules/search/infrastructure/v2documents"
	"metaldocs/internal/modules/taxonomy"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
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

type controlledDocumentDuplicatorAdapter struct {
	svc *controlleddocumentsapp.ControlledDocumentService
}

func (a controlledDocumentDuplicatorAdapter) DuplicateControlledDocument(ctx context.Context, tenantID, controlledDocumentID, actorUserID string) (*controlleddocumentsdomain.ControlledDocument, error) {
	if a.svc == nil {
		return nil, fmt.Errorf("controlled document service not configured")
	}
	source, err := a.svc.Get(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return res.ControlledDocument, nil
}

// e2eHandlersEnabled gates the test-only seed/reset/governance endpoints
// behind an explicit env flag. Defense-in-depth: internal/test/e2e_seed.go
// also early-returns on the same flag, but the call site must not register
// these routes either — permissions.go does not enumerate /internal/test/*,
// so any accidental mount is treated as fully public by newPublicPathChecker.
func e2eHandlersEnabled() bool {
	return strings.TrimSpace(os.Getenv("METALDOCS_E2E")) == "1"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repoMode, err := config.RepositoryMode()
	if err != nil {
		log.Fatalf("invalid repository mode: %v", err)
	}
	rateCfg, err := config.LoadRateLimitConfig()
	if err != nil {
		log.Fatalf("invalid rate limit config: %v", err)
	}
	corsCfg, err := config.LoadCORSConfig()
	if err != nil {
		log.Fatalf("invalid cors config: %v", err)
	}
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		log.Fatalf("invalid attachments config: %v", err)
	}
	authCfg, err := authn.LoadRuntimeConfig()
	if err != nil {
		log.Fatalf("invalid auth config: %v", err)
	}
	featureFlagsCfg, err := config.LoadFeatureFlagsConfig()
	if err != nil {
		log.Fatalf("invalid feature flags config: %v", err)
	}

	deps, err := bootstrap.BuildAPIDependencies(ctx, repoMode, attachmentsCfg)
	if err != nil {
		log.Fatalf("build api dependencies: %v", err)
	}
	defer deps.Cleanup()

	if deps.SQLDB != nil && !strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_SKIP_STARTUP_MIGRATIONS")), "true") {
		migrationsDir := strings.TrimSpace(os.Getenv("METALDOCS_MIGRATIONS_DIR"))
		if migrationsDir == "" {
			migrationsDir = "db/migrations"
		}
		if err := migrate.Apply(ctx, deps.SQLDB, migrationsDir, slog.Default()); err != nil {
			log.Fatalf("apply startup migrations: %v", err)
		}
	}

	authService, err := authapp.NewService(deps.AuthRepo, deps.RoleProvider, deps.RoleAdminRepo, authCfg)
	if err != nil {
		log.Fatalf("new auth service: %v", err)
	}
	if err := authService.BootstrapLocalAdmin(ctx); err != nil {
		log.Fatalf("bootstrap local admin: %v", err)
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
	// permResolver is the single authoritative source of truth for route
	// visibility. It is shared with the auth middleware so that fully public
	// routes (no session required) and the IAM permission layer stay in sync
	// automatically — adding a new public route requires one change here, not two.
	permResolver := newPermissionResolver()
	authMiddleware := authdelivery.NewMiddleware(authService, authCfg, authn.Enabled()).
		WithPublicPathChecker(newPublicPathChecker(permResolver))
	iamMiddleware := iamdelivery.NewMiddleware(capabilityService, cachedProvider, authn.Enabled(), authCfg.LegacyHeaderEnabled).
		WithPermissionResolver(permResolver)
	originProtection := security.NewOriginProtection(security.OriginProtectionConfig{
		Enabled:           authCfg.OriginProtection,
		SessionCookieName: authCfg.SessionCookieName,
		TrustedOrigins:    authCfg.TrustedOrigins,
		TrustedProxyCIDRs: authCfg.TrustedProxyCIDRs,
	})

	iamAdminService := iamapp.NewAdminService(deps.RoleAdminRepo, cachedProvider)
	iamAdminHandler := iamdelivery.NewAdminHandler(iamAdminService, authService, deps.AuditWriter).
		WithAuditReader(deps.AuditReader)
	featureFlagsHandler := featureflags.NewHandler(featureFlagsCfg)
	httpObs := observability.NewHTTPObservability(deps.StatusProvider)
	rateLimiter := security.NewRateLimiter(rateCfg)
	cors := security.NewCORS(corsCfg)

	mux := http.NewServeMux()
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
	controlledDocumentDuplicator := controlledDocumentDuplicatorAdapter{svc: controlledDocumentsModule.Service()}

	var membershipService *iamapp.AreaMembershipService
	if deps.SQLDB != nil {
		membershipService = iamapp.NewAreaMembershipService(iampg.NewUserAreaRepository(deps.SQLDB), nil)
	}
	iamdelivery.NewMembershipHandler(membershipService).RegisterRoutes(mux)

	// Legacy templates module routes removed — templates owns /api/v1/templates/*

	docPresigner := objectstore.NewDocumentPresigner(deps.MinioClient, deps.MinioPublicClient, deps.MinioBucket, 15*time.Minute, 25*1024*1024)
	controlledDocumentsRepo := controlleddocumentsinfra.NewPostgresControlledDocumentRepository(deps.SQLDB)
	profileRepo := taxonomyinfra.NewProfileRepository(deps.SQLDB)

	// Fanout/eigenpal client — enabled when METALDOCS_FANOUT_URL is set.
	fanoutURL := strings.TrimSpace(os.Getenv("METALDOCS_FANOUT_URL"))
	if fanoutURL == "" {
		if strings.EqualFold(strings.TrimSpace(os.Getenv("METALDOCS_REQUIRE_FANOUT")), "true") {
			log.Fatalf("METALDOCS_FANOUT_URL is required but not set")
		}
		slog.Warn("METALDOCS_FANOUT_URL not set; document approval will fail at freeze step")
	}
	serviceToken := strings.TrimSpace(os.Getenv("METALDOCS_DOCGEN_V2_SERVICE_TOKEN"))
	if fanoutURL != "" && serviceToken == "" {
		log.Fatalf("METALDOCS_DOCGEN_V2_SERVICE_TOKEN is required when METALDOCS_FANOUT_URL is set")
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

	// Approval services must be constructed before docMod so that
	// SubmitSvc can be wired into the finalize→submit flow.
	approvalRepo := approvalrepo.NewPostgresApprovalRepository(deps.SQLDB)
	approvalEmitter := approvalapp.NewSQLEmitter()
	approvalServices := approvalapp.NewServices(approvalRepo, approvalEmitter, approvalapp.RealClock{})
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		log.Fatalf("invalid jobs config: %v", err)
	}
	if deps.SQLDB != nil {
		if err := bootstrap.MigrateRiverSchema(ctx, deps.SQLDB, jobsCfg.RiverSchema); err != nil {
			log.Fatalf("migrate river schema: %v", err)
		}
		riverBundle, err := riverjobs.NewClientBundle(deps.SQLDB, riverjobs.Config{
			Queues:              jobsCfg.Queues,
			Schema:              jobsCfg.RiverSchema,
			SkipUnknownJobCheck: true,
		}, nil)
		if err != nil {
			log.Fatalf("build scheduled publish enqueuer client: %v", err)
		}
		approvalServices.WithScheduledPublishEnqueuer(approvaljobs.NewScheduledPublishEnqueuer(riverBundle.Client))
	}
	var effectiveFreezeInvoker approvalapp.FreezeInvoker = noopFreezeInvoker{}
	if freezeSvc != nil {
		effectiveFreezeInvoker = freezeSvc
	}
	pdfOutboxRepo := fanout.NewPDFOutboxRepository(deps.SQLDB)
	pdfOutboxWorker := fanout.NewPDFOutboxWorker(pdfOutboxRepo, deps.Publisher, slog.Default())
	var workerWG sync.WaitGroup
	workerWG.Add(1)
	go func() {
		defer workerWG.Done()
		// Restart loop with capped exponential backoff. The worker normally only
		// exits when ctx is cancelled; any non-nil return is treated as a crash
		// and the loop re-runs after a backoff so document approval doesn't
		// keep succeeding while PDF delivery silently stops.
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

	// Wire the documents-side adapter back into the controlled-documents service so atomic
	// CD-create can clone the initial document inside the same tx as the CD
	// insert. controlledDocumentsModule was constructed before docMod (because docMod
	// needs ControlledDocumentDuplicator), hence the post-construction setter.
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

	leaderID := schedulerLeaderID()
	s, err := jobscheduler.New(deps.SQLDB, leaderID)
	if err != nil {
		log.Fatalf("jobs scheduler configuration failed: %v", err)
	}
	if jobEnabled("ENABLE_JOB_STUCK_INSTANCE_WATCHDOG") {
		s.Register(jobscheduler.JobConfig{
			Name:     "stuck-instance-watchdog",
			Interval: 5 * time.Minute,
			Fn:       stuck_instance_watchdog.New(deps.SQLDB, approvalServices.Cancel, approvalEmitter),
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

	var schedulerWG sync.WaitGroup
	schedulerWG.Add(1)
	go func() {
		defer schedulerWG.Done()
		s.Start(ctx)
	}()

	stopSessions := jobs.StartSessionSweeper(ctx, docMod.Repo(), 60*time.Second)
	stopOrphans := jobs.StartOrphanPendingSweeper(ctx, docMod.Repo(), time.Hour)
	defer stopSessions()
	defer stopOrphans()
	mux.Handle("/api/v1/metrics", httpObs.MetricsHandler())

	// Audit retention - AUDIT_RETENTION_DAYS=0 disables (default disabled).
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

	handler := cors.Wrap(originProtection.Wrap(authMiddleware.Wrap(iamMiddleware.Wrap(httpObs.Wrap(rateLimiter.Wrap(mux))))))

	addr := ":8080"
	if appPort := os.Getenv("APP_PORT"); appPort != "" {
		port, convErr := strconv.Atoi(strings.TrimSpace(appPort))
		if convErr != nil || port < 1 || port > 65535 {
			log.Fatalf("invalid APP_PORT value")
		}
		addr = ":" + strconv.Itoa(port)
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("MetalDocs API listening on %s (repository=%s auth_enabled=%t auth_cache_ttl=%s rate_limit_enabled=%t rate_limit_window_s=%d rate_limit_max_requests=%d cors_enabled=%t cors_allowed_origins=%d)",
		addr, repoMode, authn.Enabled(), authn.CacheTTL(), rateCfg.Enabled, rateCfg.WindowSeconds, rateCfg.MaxRequests, corsCfg.Enabled, len(corsCfg.AllowedOrigins))

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
	return a.writer.RecordTx(ctx, tx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: "document",
		ResourceID:   docID,
		PayloadJSON:  string(raw),
		TraceID:      traceIDFromContext(ctx),
		TenantID:     tenantID,
	})
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

	if err := a.writer.Record(ctx, auditdomain.Event{
		ID:           uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       action,
		ResourceType: "document",
		ResourceID:   docID,
		PayloadJSON:  string(raw),
		TraceID:      traceIDFromContext(ctx),
		TenantID:     tenantID,
	}); err != nil {
		log.Printf("documents audit write failed: %v", err)
	}
}

func traceIDFromContext(_ context.Context) string {
	// Observability middleware stores the trace ID in the slog logger context, not in the
	// Go context value chain. Generating a fresh UUID per audit event is correct here.
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

// profileDefaultsAdapter bridges taxonomy ProfileRepository → documents module ProfileDefaultTemplateReader.
type profileDefaultsAdapter struct {
	profileRepo interface {
		GetByCode(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error)
	}
}

func (a *profileDefaultsAdapter) GetDefaultTemplateVersionID(ctx context.Context, tenantID, profileCode string) (*string, *string, error) {
	profile, err := a.profileRepo.GetByCode(ctx, tenantID, taxonomydomain.ProfileCode(profileCode))
	if err != nil {
		return nil, nil, err
	}
	if profile.DefaultTemplateVersionID == nil {
		return nil, nil, nil
	}
	status := "published"
	return profile.DefaultTemplateVersionID, &status, nil
}
