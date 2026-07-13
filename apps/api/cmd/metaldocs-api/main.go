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
	templatesrepo "metaldocs/internal/modules/templates/infrastructure"
	templatejobs "metaldocs/internal/modules/templates/jobs"

	"metaldocs/apps/api/internal/wiring"
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
	riverjobs "metaldocs/internal/platform/jobs/river"
	platformmw "metaldocs/internal/platform/middleware"
	"metaldocs/internal/platform/migrate"
	"metaldocs/internal/platform/objectstore"
	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/ratelimit"
	"metaldocs/internal/platform/security"
	"metaldocs/internal/platform/tenantdata/registry"
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

	// M7 F7.3: tenant DEK/KEK crypto-shred framework. tenantCrypto is nil when
	// METALDOCS_TENANT_KEK is unset — every consumer below falls back to its
	// own no-op path (F7.2's nil-safe pattern), so boot still works with
	// crypto disabled. Constructed here (before audit wiring) so the audit
	// writer below can be wired with the payload encryptor at boot.
	var tenantCrypto securitydomain.TenantCrypto
	if deps.SQLDB != nil {
		kek, kekConfigured, kekErr := config.LoadTenantKEK()
		if kekErr != nil {
			slog.Error("invalid tenant crypto KEK", "err", kekErr)
			os.Exit(1)
		}
		if kekConfigured {
			svc, err := securityapp.NewTenantCryptoService(securitypg.NewTenantKeyRepository(deps.SQLDB), kek)
			if err != nil {
				slog.Error("construct tenant crypto service", "err", err)
				os.Exit(1)
			}
			tenantCrypto = svc
		} else {
			slog.Info("tenant crypto disabled: METALDOCS_TENANT_KEK not set")
		}
	}

	// M7 F7.3 item 2: wire the audit payload envelope encryptor. deps.AuditWriter
	// is the same *auditpg.Writer instance backing AuditReader/AuditCounter (see
	// bootstrap.BuildAPIDependencies), so wiring it here via WithPayloadCrypto
	// (mutate-and-return-self) takes effect for every consumer of those three
	// fields. Nil tenantCrypto (KEK unset) leaves the writer's crypto nil —
	// RecordTx/ListEvents stay on the legacy plaintext path byte-for-byte.
	if tenantCrypto != nil {
		if auditWriter, ok := deps.AuditWriter.(*auditpg.Writer); ok {
			auditWriter.WithPayloadCrypto(auditPayloadCryptoAdapter{crypto: tenantCrypto})
		}
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

	// M7 F7.2: tenant onboarding (POST /tenants). tenantHandler is nil (Router
	// answers 501) on the SQLDB-less boot path, matching every other
	// conditionally-wired iam handler above/below.
	var tenantHandler *iamdelivery.TenantHandler
	if deps.SQLDB != nil {
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
		tenantHandler = iamdelivery.NewTenantHandler(onboardTenantService)
	}

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

	// Rate-limit store backend selection (M8/F8.2): memory (default,
	// single-replica) or a shared Redis store so N api replicas enforce ONE
	// combined budget instead of N independent ones. The startup guard
	// refuses to boot when METALDOCS_MULTI_REPLICA=true with the in-memory
	// store (per-process counters would silently multiply the limits ×N).
	rlStoreCfg, err := ratelimit.LoadStoreConfig(os.Getenv)
	if err != nil {
		slog.Error("rate limit store config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	// nil for the memory backend — each limiter then builds its own private
	// in-memory store (unchanged behavior). Non-nil (redis) is shared by BOTH
	// limiter mounts below so the process holds exactly one Redis client.
	rlStore, err := ratelimit.NewStoreFromConfig(rlStoreCfg, slog.Default())
	if err != nil {
		slog.Error("rate limit store init", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	if rlStore != nil {
		defer rlStore.Close()
	}

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
	loginRateCfg.Store = rlStore
	preAuthLimiter := ratelimit.New(ctx, loginRateCfg)

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
	// iamAdminHandler, sessionsHandler, observabilityHandler: mounted below via
	// iamdelivery.Router.RegisterGenerated (CON-07 codegen rollout), once
	// peopleHandler/membershipHandler/rolesCapsHandler/presenceHandler are also
	// constructed. securityHandler is a separate module (not IAM) and keeps its
	// own hand-written mount here.
	if securityHandler != nil {
		securityHandler.RegisterRoutes(mux)
	}

	presenceBump, presenceHub, presenceHandler := startPresence(ctx, deps, mux, iamAdminHandler)

	taxonomyModule := buildTaxonomyModule(deps)
	taxonomyModule.RegisterRoutes(mux)

	tokensModule := buildTokensModule(deps)
	tokensModule.RegisterRoutes(mux)

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
	iamRouter.RegisterGenerated(mux)

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
	jobsCfg, err := config.LoadJobsConfig()
	if err != nil {
		slog.Error("invalid jobs config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	var riverBundle *riverjobs.ClientBundle
	if deps.SQLDB != nil {
		if err := bootstrap.MigrateRiverSchema(ctx, deps.SQLDB, jobsCfg.RiverSchema); err != nil {
			slog.Error("migrate river schema", "err", err)
			deps.Cleanup()
			os.Exit(1)
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
			slog.Error("build scheduled publish enqueuer client", "err", err)
			deps.Cleanup()
			os.Exit(1)
		}
		approvalServices.WithScheduledPublishEnqueuer(approvaljobs.NewScheduledPublishEnqueuer(riverBundle.Client))
		approvalServices.WithLifecycleEnqueuer(approvaljobs.NewLifecycleEventEnqueuer(riverBundle.Client))

		// M7 F7.3 Task E: tenant export/erase orchestrator. Constructed here
		// (not alongside tenantHandler/onboardTenantService above) because it
		// needs riverBundle.Client for the paired same-tx River enqueue —
		// riverBundle does not exist yet at the point tenantHandler is built.
		// WithLifecycle mutates the already-constructed tenantHandler in
		// place; iamRouter (built later, but referencing the SAME
		// tenantHandler pointer via WithTenantHandler) sees the wiring once
		// RegisterGenerated's wrapped handlers are invoked at request time —
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
		deps.Cleanup()
		os.Exit(1)
	}
	pdfOutboxRepo := fanout.NewPDFOutboxRepository(deps.SQLDB)
	materializeOutboxRepo := fanout.NewMaterializeOutboxRepository(deps.SQLDB)

	stagingOutboxWorkerCfg, err := config.LoadStagingOutboxWorkerConfig()
	if err != nil {
		slog.Error("invalid staging outbox worker config", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}

	// pdfDispatchEnqueuer produces the paired (outbox row, River job) write for
	// both staging dispatch kinds inside the caller's business tx (M5 F5.3 T3).
	// riverBundle.Client is enqueue-only here (never Started in this binary);
	// the temporal-queue dispatch workers that consume these jobs run in
	// metaldocs-jobs.
	pdfDispatchEnqueuer := dispatchjobs.NewEnqueuer(riverBundle.Client, pdfOutboxRepo, materializeOutboxRepo, stagingOutboxWorkerCfg.MaxAttempts)

	// Wire materialize outbox into the freeze service so Pin can enqueue async jobs.
	fanoutCfg.freezeService.WithMaterializeOutbox(pdfDispatchEnqueuer)

	approvalServices.Decision = approvalapp.NewDecisionService(
		approvalRepo, approvalEmitter, approvalapp.RealClock{},
	).WithPDFOutbox(pdfDispatchEnqueuer).WithPinInvoker(fanoutCfg.freezeService).
		WithSignatureRegistry(newSignoffReauthRegistry(deps.AuthRepo, deps.SQLDB)).
		WithCDFieldReader(cdReader)

	docMod := documents.New(docDeps)
	// SP-2: pin tenant dictionary values at document creation. tokensModule was
	// built at startup (line ~358), before docMod.
	docMod.Service.WithDictionaryReader(dictionaryValueReaderAdapter{reader: tokensModule.Reader})
	docMod.RegisterRoutesWithRateLimit(mux, globalLimiter, userIDExtractor)

	// Wire the documents-side adapter back into the controlled-documents service so atomic
	// CD-create can clone the initial document inside the same tx as the CD
	// insert. controlledDocumentsModule was constructed before docMod (because docMod
	// needs ControlledDocumentDuplicator), hence the post-construction setter.
	controlledDocumentsModule.Service().WithDocumentInitializer(docapp.NewCDDocumentInitializer(docMod.Service))

	templatesModule, templatesRepo, templatesStore, err := buildTemplatesModule(deps, capabilityService, sharedPresigner, displayNameRepo)
	if err != nil {
		slog.Error("build templates module", "err", err)
		deps.Cleanup()
		os.Exit(1)
	}
	// M3 P3.S2b-4 (R2a): wire the approval kernel's published services into
	// the templates HTTP handler so its two thin kernel routes
	// (submit-for-approval, signoff) can delegate. Must happen after
	// approvalServices.Decision is finalized above (line ~737-741).
	templatesModule.WithApprovalKernel(approvalServices.TemplateSubmit, approvalServices.Decision, approvalServices.Read, db.NewTxRunner(deps.SQLDB))
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

	if deps.SQLDB != nil {
		httpObs.SetDBPool(postgres.NewPoolStatsAdapter(deps.SQLDB))
	}

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
	mux.Handle("/api/v1/metrics", httpObs.MetricsHandler())
	// Prometheus text-exposition scrape endpoint. Deliberately NOT mounted on
	// mux (which the API chain below wraps with authn/iam/httpObs/rate-limit)
	// — it is served from a top-level dispatch mux, ahead of and outside the
	// entire API chain (see rootMux below, after handler is built). Contract
	// §3.2: /metrics is a platform scrape surface, not a versioned product
	// route, so it is NOT declared in api/openapi/v1/openapi.yaml. Coexists
	// with the JSON endpoint above; they read from separate storage
	// (prometheus vecs vs. byKey atomics).

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

	// /metrics is served on a DEDICATED listener (METRICS_ADDR, default :9090),
	// never on the public API server above. This isolates the scrape surface by
	// process topology: the public port structurally cannot serve /metrics, so
	// exposure no longer depends on ops/ingress discipline (F-R1, Dim-9). The
	// scrape stays credential-less (bypasses authn/iam) and self-scrapes never
	// feed httpObs.Wrap counters or the global rate limiter, because this mux is
	// not part of the API chain. Panic recovery still wraps it. Not in openapi
	// (contract §3.2). Compose does not host-publish this port — infra-network
	// reachable only (see ops/DEPLOY.md).
	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics", httpObs.PrometheusHandler())
	metricsServer := &http.Server{
		Addr:              serverCfg.MetricsAddr,
		Handler:           platformmw.Recovery(metricsMux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("MetalDocs API listening",
		"addr", serverCfg.Addr, "metrics_addr", serverCfg.MetricsAddr,
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

	exitCode := shutdownServer(ctx, stop, server, metricsServer, serverErr, metricsErr)
	if exitCode != 0 {
		// os.Exit skips deferred functions, including deps.Cleanup. Invoke
		// cleanup explicitly so DB / object-store handles are released on
		// the error path too. closeDB swallows close-on-closed, so calling
		// twice is safe.
		deps.Cleanup()
		os.Exit(exitCode)
	}
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
func buildTemplatesModule(deps bootstrap.APIDependencies, capabilityService *iamapp.CapabilityService, presigner *objectstore.VerifiedStore, displayNameReader iamdomain.UserDisplayNameReader) (*templateshttp.Handler, *templatesrepo.Repository, *objectstore.VerifiedStore, error) {
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
	templatesRepo := templatesrepo.New(deps.SQLDB).WithAudit(deps.AuditWriter)
	templatesSvc := templatesapp.New(templatesRepo, templatesPresigner, wiring.Clock{}, wiring.UUIDGen{}, templatesResolverReg).WithRunner(db.NewTxRunner(deps.SQLDB))
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
// CON-07: the presence Handler's WebSocket /iam/presence/stream route is still
// registered here directly (RegisterRoutes only mounts /stream — see
// presence/handler.go), because streamPresence is excluded from server codegen
// (cfg.yaml exclude-operation-ids). The HTTP-fallback snapshot route is NOT
// mounted here anymore: the caller mounts it via iamdelivery.Router.
// RegisterGenerated using the returned *iampresence.Handler (ServeSnapshot).
func startPresence(
	ctx context.Context,
	deps bootstrap.APIDependencies,
	mux *http.ServeMux,
	iamAdminHandler *iamdelivery.AdminHandler,
) (*iampresence.BumpMiddleware, *iampresence.Hub, *iampresence.Handler) {
	if deps.SQLDB == nil {
		return nil, nil, nil
	}
	presenceRepo := iampresence.NewPostgresRepository(deps.SQLDB)
	presenceHub := iampresence.NewHub(presenceRepo, slog.Default())
	go presenceHub.Run(ctx)
	go presenceHub.RunHeartbeat(ctx)
	presenceHandler := iampresence.NewHandler(presenceHub, presenceRepo, slog.Default())
	presenceHandler.RegisterRoutes(mux)
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
