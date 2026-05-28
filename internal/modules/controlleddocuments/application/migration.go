package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

const controlledDocumentsBackfillAdvisoryLock int64 = 903210421

func BackfillLegacyDocuments(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire backfill connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	var locked bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, controlledDocumentsBackfillAdvisoryLock).Scan(&locked); err != nil {
		return fmt.Errorf("acquire backfill advisory lock: %w", err)
	}
	if !locked {
		logger.Info("controlled-documents backfill skipped; advisory lock not acquired")
		return nil
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, controlledDocumentsBackfillAdvisoryLock)
	}()

	processed := 0
	skipped := 0
	errorsCount := 0
	for {
		rows, err := conn.QueryContext(ctx, `
			SELECT id::text, tenant_id::text
			FROM documents
			WHERE controlled_document_id IS NULL
			ORDER BY created_at ASC
			LIMIT 100`)
		if err != nil {
			return fmt.Errorf("query legacy documents: %w", err)
		}

		batchCount := 0
		for rows.Next() {
			batchCount++

			var docID string
			var tenantID string
			if err := rows.Scan(&docID, &tenantID); err != nil {
				errorsCount++
				logger.Error("controlled-documents backfill row scan failed", "event", "backfill", "error", err)
				continue
			}

			tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
			if err != nil {
				errorsCount++
				logger.Error("controlled-documents backfill begin tx failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", err)
				continue
			}

			legacyCode, err := legacyCodeFromDocumentID(docID)
			if err != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill legacy code generation failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", err)
				continue
			}
			insertRes, err := tx.ExecContext(ctx, `
				INSERT INTO controlled_documents
					(tenant_id, profile_code, process_area_code, code, sequence_num, title, owner_user_id, status)
				VALUES
					($1, 'unassigned', 'unassigned', $2, NULL, 'Legacy backfill', 'system', 'active')
				ON CONFLICT (tenant_id, profile_code, code)
				DO NOTHING`,
				tenantID, legacyCode,
			)
			if err != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill controlled_document insert failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", err)
				continue
			}
			insertedRows, rowsErr := insertRes.RowsAffected()
			if rowsErr != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill insert rows affected failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", rowsErr)
				continue
			}
			if insertedRows == 0 {
				skipped++
			}

			var controlledDocumentID string
			if err := tx.QueryRowContext(ctx, `
				SELECT id::text
				FROM controlled_documents
				WHERE tenant_id = $1
				  AND profile_code = 'unassigned'
				  AND code = $2`,
				tenantID, legacyCode,
			).Scan(&controlledDocumentID); err != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill controlled_document lookup failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", err)
				continue
			}

			res, err := tx.ExecContext(ctx, `
				UPDATE documents
				SET controlled_document_id = $1,
				    profile_code_snapshot = 'unassigned',
				    process_area_code_snapshot = 'unassigned'
				WHERE id = $2
				  AND controlled_document_id IS NULL`,
				controlledDocumentID, docID,
			)
			if err != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill document update failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", err)
				continue
			}
			updatedRows, rowsErr := res.RowsAffected()
			if rowsErr != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill update rows affected failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", rowsErr)
				continue
			}

			if err := tx.Commit(); err != nil {
				_ = tx.Rollback()
				errorsCount++
				logger.Error("controlled-documents backfill commit failed", "event", "backfill", "document_id", docID, "tenant_id", tenantID, "error", err)
				continue
			}
			if updatedRows > 0 {
				processed++
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return fmt.Errorf("iterate legacy documents: %w", err)
		}
		_ = rows.Close()
		if batchCount == 0 {
			break
		}
	}

	logger.Info("controlled-documents backfill completed", "event", "backfill", "processed", processed, "skipped", skipped, "errors", errorsCount)
	return nil
}

func legacyCodeFromDocumentID(docID string) (string, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(docID), "-", ""))
	if len(normalized) < 8 {
		return "", errors.New("invalid legacy document id")
	}
	return "MIG-" + normalized[:8], nil
}
