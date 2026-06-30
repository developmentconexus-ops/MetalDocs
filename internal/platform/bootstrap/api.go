package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	auditdomain "metaldocs/internal/modules/audit/domain"
	auditpg "metaldocs/internal/modules/audit/infrastructure/postgres"
	authdomain "metaldocs/internal/modules/auth/domain"
	authpg "metaldocs/internal/modules/auth/infrastructure/postgres"
	iamdomain "metaldocs/internal/modules/iam/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/config"
	pgdb "metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/messaging"
	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/internal/platform/observability"
	"metaldocs/internal/platform/render/gotenberg"
	"metaldocs/internal/platform/servicebus"
	miniostore "metaldocs/internal/platform/storage/minio"
)

type APIDependencies struct {
	RoleProvider    iamdomain.RoleProvider
	RoleAdminRepo   iamdomain.RoleAdminRepository
	AuthRepo        authdomain.Repository
	AuditWriter     auditdomain.Writer
	AuditReader     auditdomain.Reader
	AuditCounter    auditdomain.Counter
	AuditExports    auditdomain.ExportJobRepository
	AuditValidator  auditdomain.IntegrityValidator
	Publisher       messaging.Publisher
	GotenbergClient *gotenberg.Client
	StatusProvider  observability.RuntimeStatusProvider
	// SQLDB is the raw *sql.DB used by modules that manage their own queries
	// (e.g. the templates module). Nil in memory/test mode.
	SQLDB *sql.DB
	// PDFConverter generates PDFs directly via Gotenberg (docx->PDF). Nil when
	// Gotenberg is not configured or storage is not MinIO.
	PDFConverter *servicebus.GotenbergPDFClient
	// MinioClient is the minio client for presigning. Nil when storage is not minio.
	MinioClient *miniogo.Client
	// MinIO wiring.
	// MinioPublicClient signs browser-facing URLs against a browser-reachable endpoint.
	MinioPublicClient *miniogo.Client
	MinioBucket       string
	Cleanup           func()
}

func BuildAPIDependencies(ctx context.Context, repoMode string, attachmentsCfg config.AttachmentsConfig) (APIDependencies, error) {
	gotenbergCfg, err := config.LoadGotenbergConfig()
	if err != nil {
		return APIDependencies{}, fmt.Errorf("load gotenberg config: %w", err)
	}
	var gotenbergClient *gotenberg.Client
	if gotenbergCfg.Enabled {
		gotenbergClient, err = gotenberg.NewClient(gotenbergCfg.URL)
		if err != nil {
			return APIDependencies{}, err
		}
	}

	switch repoMode {
	case config.RepositoryPostgres: //nolint:gocritic // single-case switch reserved for future modes
		pgCfg, err := config.LoadPostgresConfig()
		if err != nil {
			return APIDependencies{}, fmt.Errorf("load postgres config: %w", err)
		}
		db, err := pgdb.Open(ctx, pgCfg.DSN)
		if err != nil {
			return APIDependencies{}, fmt.Errorf("open postgres: %w", err)
		}
		authRepo := authpg.NewRepository(db, iampg.NewUserTenantRepository(db))
		var minioClient *miniogo.Client
		var minioPublicClient *miniogo.Client
		var minioBucket string
		var pdfConverter *servicebus.GotenbergPDFClient
		if attachmentsCfg.Provider == config.StorageProviderMinIO {
			var err error
			minioClient, minioPublicClient, minioBucket, err = buildMinioClients(attachmentsCfg)
			if err != nil {
				_ = closeDB(db)
				return APIDependencies{}, err
			}
			if gotenbergClient != nil {
				store := miniostore.NewStore(minioClient, attachmentsCfg)
				pdfConverter = servicebus.NewGotenbergPDFClient(store, gotenbergClient)
			}
		}
		auditStore := auditpg.NewWriter(db)
		auditExports := auditpg.NewExportJobRepository(db)
		return APIDependencies{
			RoleProvider:      iampg.NewRoleProvider(db),
			RoleAdminRepo:     iampg.NewRoleAdminRepository(db),
			AuthRepo:          authRepo,
			AuditWriter:       auditStore,
			AuditReader:       auditStore,
			AuditCounter:      auditStore,
			AuditExports:      auditExports,
			AuditValidator:    auditStore,
			Publisher:         outboxpg.NewPublisher(db),
			GotenbergClient:   gotenbergClient,
			StatusProvider:    observability.NewPostgresRuntimeStatusProvider(db, repoMode, string(attachmentsCfg.Provider), authn.Enabled(), gotenbergHealthCheck(gotenbergCfg)),
			SQLDB:             db,
			PDFConverter:      pdfConverter,
			MinioClient:       minioClient,
			MinioPublicClient: minioPublicClient,
			MinioBucket:       minioBucket,
			Cleanup:           func() { _ = closeDB(db) },
		}, nil
	default:
		return APIDependencies{}, fmt.Errorf("unsupported repository mode: %q", repoMode)
	}
}

func buildMinioClients(attachmentsCfg config.AttachmentsConfig) (*miniogo.Client, *miniogo.Client, string, error) {
	internalClient, err := miniogo.New(attachmentsCfg.MinIOEndpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(attachmentsCfg.MinIOAccessKey, attachmentsCfg.MinIOSecretKey, ""),
		Secure: attachmentsCfg.MinIOUseSSL,
		Region: attachmentsCfg.MinIORegion,
	})
	if err != nil {
		return nil, nil, "", fmt.Errorf("init minio internal client: %w", err)
	}

	publicClient, err := miniogo.New(attachmentsCfg.MinIOPublicEndpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(attachmentsCfg.MinIOAccessKey, attachmentsCfg.MinIOSecretKey, ""),
		Secure: attachmentsCfg.MinIOUseSSL,
		Region: attachmentsCfg.MinIORegion,
	})
	if err != nil {
		return nil, nil, "", fmt.Errorf("init minio public client: %w", err)
	}

	return internalClient, publicClient, attachmentsCfg.MinIOBucket, nil
}

func closeDB(db *sql.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}

func gotenbergHealthCheck(cfg config.GotenbergConfig) observability.DependencyCheck {
	return observability.DependencyCheck{
		Name: "gotenberg",
		Check: func(ctx context.Context) (observability.DependencyCheckResult, error) {
			if !cfg.Enabled || cfg.URL == "" {
				return observability.DependencyCheckResult{
					Status: "skipped",
					Detail: "gotenberg not configured",
				}, nil
			}
			client := &http.Client{Timeout: 2 * time.Second}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL+"/health", nil)
			if err != nil {
				return observability.DependencyCheckResult{}, err
			}
			resp, err := client.Do(req)
			if err != nil {
				return observability.DependencyCheckResult{}, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return observability.DependencyCheckResult{}, fmt.Errorf("gotenberg unhealthy: status %d", resp.StatusCode)
			}
			return observability.DependencyCheckResult{
				Status: "up",
				Detail: cfg.URL,
			}, nil
		},
	}
}
