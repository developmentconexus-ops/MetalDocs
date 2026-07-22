package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"metaldocs/internal/platform/config"
	pgdb "metaldocs/internal/platform/db/postgres"
	"metaldocs/internal/platform/messaging"
	outboxpg "metaldocs/internal/platform/messaging/outbox/postgres"
	"metaldocs/internal/platform/render/gotenberg"
	"metaldocs/internal/platform/servicebus"
	miniostore "metaldocs/internal/platform/storage/minio"
)

type WorkerDependencies struct {
	Consumer     messaging.Consumer
	PDFConverter *servicebus.GotenbergPDFClient
	SQLDB        *sql.DB
	FanoutURL    string
	FanoutToken  string
	Cleanup      func()
}

func BuildWorkerDependencies(ctx context.Context, workerCfg config.WorkerConfig) (WorkerDependencies, error) {
	pgCfg, err := config.LoadPostgresConfig()
	if err != nil {
		return WorkerDependencies{}, fmt.Errorf("load postgres config: %w", err)
	}
	db, err := pgdb.Open(ctx, pgCfg.DSN)
	if err != nil {
		return WorkerDependencies{}, fmt.Errorf("open postgres: %w", err)
	}

	pdfConverter, err := buildWorkerPDFConverter()
	if err != nil {
		_ = closeDB(db)
		return WorkerDependencies{}, err
	}

	consumer := outboxpg.NewConsumer(db, workerClaimLease(workerCfg))

	// Reuse the shared loader so the worker enforces the same URL⇒token invariant
	// as the API (fail-fast if the renderer URL is set without an auth token),
	// instead of reading the env raw and silently building a tokenless client.
	fanoutCfg, err := config.LoadFanoutConfig()
	if err != nil {
		_ = closeDB(db)
		return WorkerDependencies{}, fmt.Errorf("load fanout config: %w", err)
	}

	return WorkerDependencies{
		Consumer:     consumer,
		PDFConverter: pdfConverter,
		SQLDB:        db,
		FanoutURL:    fanoutCfg.URL,
		FanoutToken:  fanoutCfg.ServiceToken,
		Cleanup:      func() { _ = closeDB(db) },
	}, nil
}

// buildWorkerPDFConverter wires the direct Gotenberg PDF path:
// docx (MinIO) -> Gotenberg LibreOffice -> PDF (MinIO).
// Returns nil when Gotenberg is not configured or storage is not MinIO. A nil
// converter yields a nil PDFJobRunner, and the worker service now FAILS LOUD on
// a PDFConvert event with no runner (F-QA2-1) — the event retries then
// dead-letters rather than being silently marked published. (The prior claim
// that "the outbox retries later" was false: the event was marked published and
// never retried.) PDF generation therefore does not run until the converter is
// configured; deploys that expect PDFs MUST wire Gotenberg + MinIO storage.
func buildWorkerPDFConverter() (*servicebus.GotenbergPDFClient, error) {
	gotenbergCfg, err := config.LoadGotenbergConfig()
	if err != nil {
		return nil, fmt.Errorf("load gotenberg config: %w", err)
	}
	if !gotenbergCfg.Enabled {
		return nil, nil
	}
	attachmentsCfg, err := config.LoadAttachmentsConfig()
	if err != nil {
		return nil, fmt.Errorf("load attachments config: %w", err)
	}
	if attachmentsCfg.Provider != config.StorageProviderMinIO {
		return nil, nil
	}
	minioClient, err := miniogo.New(attachmentsCfg.MinIOEndpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(attachmentsCfg.MinIOAccessKey, attachmentsCfg.MinIOSecretKey, ""),
		Secure: attachmentsCfg.MinIOUseSSL,
		Region: attachmentsCfg.MinIORegion,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	store := miniostore.NewStore(minioClient, attachmentsCfg)
	gotenbergClient, err := gotenberg.NewClient(gotenbergCfg.URL)
	if err != nil {
		return nil, fmt.Errorf("build gotenberg client: %w", err)
	}
	return servicebus.NewGotenbergPDFClient(store, gotenbergClient), nil
}

func workerClaimLease(cfg config.WorkerConfig) time.Duration {
	const minClaimLease = 5 * time.Minute

	lease := time.Duration(cfg.RetryMaxSeconds) * time.Second
	if lease < minClaimLease {
		return minClaimLease
	}
	return lease
}
