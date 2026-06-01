package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

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

	fanoutURL := strings.TrimSpace(os.Getenv("METALDOCS_FANOUT_URL"))
	fanoutToken := strings.TrimSpace(os.Getenv("METALDOCS_DOCX_RENDERER_SERVICE_TOKEN"))

	return WorkerDependencies{
		Consumer:     consumer,
		PDFConverter: pdfConverter,
		SQLDB:        db,
		FanoutURL:    fanoutURL,
		FanoutToken:  fanoutToken,
		Cleanup:      func() { _ = closeDB(db) },
	}, nil
}

// buildWorkerPDFConverter wires the direct Gotenberg PDF path:
// docx (MinIO) -> Gotenberg LibreOffice -> PDF (MinIO).
// Returns nil when Gotenberg is not configured or storage is not MinIO, in
// which case PDF generation simply does not run (the outbox retries later).
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
	store, err := miniostore.NewStore(attachmentsCfg)
	if err != nil {
		return nil, fmt.Errorf("build minio store: %w", err)
	}
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
