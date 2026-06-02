package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/audit/domain"
)

var (
	ErrTenantRequired      = errors.New("audit: tenant id is required")
	ErrReaderRequired      = errors.New("audit: reader is required")
	ErrActorRequired       = errors.New("audit: actor id is required")
	ErrInvalidFormat       = errors.New("audit: invalid export format")
	ErrExportTooLarge      = errors.New("audit: export result set too large for synchronous export")
	ErrExportNotImpl       = errors.New("audit: async export not implemented")
	ErrExportRepoMissing   = errors.New("audit: export job repository not configured")
	ErrCounterMissing      = errors.New("audit: counter not configured for export sizing")
	ErrExportTokenMismatch = errors.New("audit: invalid export download token")
)

// SyncExportRowLimit is the threshold above which an export request is
// rejected — async worker lands in a later PR.
const SyncExportRowLimit int64 = 50_000

// ExportTTL is how long a generated export's signed download URL stays valid.
const ExportTTL = 15 * time.Minute

// SignedURLBuilder turns a stored export job into the externally visible signed
// download URL.
type SignedURLBuilder func(job domain.ExportJob) string

type Service struct {
	reader     domain.Reader
	counter    domain.Counter
	exportRepo domain.ExportJobRepository
	writer     domain.Writer
	signedURL  SignedURLBuilder
	now        func() time.Time
}

func NewService(reader domain.Reader) *Service {
	if reader == nil {
		panic(ErrReaderRequired.Error())
	}
	return &Service{reader: reader, now: func() time.Time { return time.Now().UTC() }}
}

// WithExports wires the export pipeline.
func (s *Service) WithExports(counter domain.Counter, repo domain.ExportJobRepository, writer domain.Writer, urlBuilder SignedURLBuilder) *Service {
	s.counter = counter
	s.exportRepo = repo
	s.writer = writer
	s.signedURL = urlBuilder
	return s
}

func (s *Service) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, error) {
	if s == nil || s.reader == nil {
		return nil, ErrReaderRequired
	}

	normalized, err := normalizeQuery(query)
	if err != nil {
		return nil, err
	}

	return s.reader.ListEvents(ctx, normalized)
}

func normalizeQuery(query domain.ListEventsQuery) (domain.ListEventsQuery, error) {
	normalized := domain.ListEventsQuery{
		TenantID:       strings.TrimSpace(query.TenantID),
		ResourceType:   strings.TrimSpace(query.ResourceType),
		ResourceID:     strings.TrimSpace(query.ResourceID),
		ActorID:        strings.TrimSpace(query.ActorID),
		Action:         strings.TrimSpace(query.Action),
		Query:          strings.TrimSpace(query.Query),
		OccurredAfter:  query.OccurredAfter,
		OccurredBefore: query.OccurredBefore,
		Cursor:         query.Cursor,
		Limit:          query.Limit,
	}
	if normalized.TenantID == "" {
		return domain.ListEventsQuery{}, ErrTenantRequired
	}
	if normalized.Limit <= 0 {
		normalized.Limit = 50
	}
	if normalized.Limit > 200 {
		normalized.Limit = 200
	}
	return normalized, nil
}

// ExportEvents runs the export inline. Refuses with ErrExportTooLarge when row
// count exceeds SyncExportRowLimit.
func (s *Service) ExportEvents(ctx context.Context, actorID string, format domain.ExportFormat, filter domain.ListEventsQuery) (domain.ExportJob, error) {
	if s == nil {
		return domain.ExportJob{}, ErrReaderRequired
	}
	if s.exportRepo == nil {
		return domain.ExportJob{}, ErrExportRepoMissing
	}
	if s.counter == nil {
		return domain.ExportJob{}, ErrCounterMissing
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return domain.ExportJob{}, ErrActorRequired
	}
	if !format.Valid() {
		return domain.ExportJob{}, ErrInvalidFormat
	}

	sizingFilter := filter
	sizingFilter.Limit = 0
	sizingFilter.Cursor = domain.Cursor{}
	normalizedSizing, err := normalizeQuery(sizingFilter)
	if err != nil {
		return domain.ExportJob{}, err
	}
	estimatedRows, err := s.counter.CountEvents(ctx, normalizedSizing)
	if err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: estimate export rows: %w", err)
	}
	if estimatedRows > SyncExportRowLimit {
		return domain.ExportJob{}, fmt.Errorf("%w: %d rows exceeds limit %d", ErrExportTooLarge, estimatedRows, SyncExportRowLimit)
	}

	events, err := s.fetchAll(ctx, normalizedSizing)
	if err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: fetch export rows: %w", err)
	}

	payload, err := render(format, events)
	if err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: render export: %w", err)
	}

	filterJSON, err := json.Marshal(filterPayload(filter))
	if err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: marshal filter: %w", err)
	}

	now := s.now()
	job := domain.ExportJob{
		ID:            uuid.NewString(),
		TenantID:      normalizedSizing.TenantID,
		ActorID:       actorID,
		Format:        format,
		FilterJSON:    string(filterJSON),
		Status:        domain.ExportStatusReady,
		DownloadToken: randomToken(),
		ExpiresAt:     now.Add(ExportTTL),
		EstimatedRows: estimatedRows,
		ActualRows:    int64(len(events)),
		Payload:       payload,
		CreatedAt:     now,
		CompletedAt:   now,
	}
	if err := s.exportRepo.Save(ctx, job); err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: persist export job: %w", err)
	}

	if s.writer != nil {
		summary := map[string]any{
			"format":        string(format),
			"filterSummary": filterSummary(filter),
			"estimatedRows": estimatedRows,
			"actualRows":    job.ActualRows,
			"exportId":      job.ID,
		}
		if event, evErr := domain.NewEvent(job.TenantID, "audit_export", job.ID, actorID, "audit.export.requested", summary); evErr == nil {
			_ = s.writer.Record(ctx, event)
		}
	}

	return job, nil
}

// GetExportStatus returns an export job, enforcing tenant + actor scoping.
func (s *Service) GetExportStatus(ctx context.Context, tenantID, actorID, exportID string) (domain.ExportJob, error) {
	if s == nil || s.exportRepo == nil {
		return domain.ExportJob{}, ErrExportRepoMissing
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return domain.ExportJob{}, ErrTenantRequired
	}
	job, err := s.exportRepo.Get(ctx, tenantID, strings.TrimSpace(exportID))
	if err != nil {
		return domain.ExportJob{}, err
	}
	if a := strings.TrimSpace(actorID); a != "" && job.ActorID != a {
		return domain.ExportJob{}, domain.ErrExportJobNotFound
	}
	return job, nil
}

// LoadExportPayload validates the download token + returns persisted bytes.
func (s *Service) LoadExportPayload(ctx context.Context, exportID, token string) (domain.ExportJob, error) {
	if s == nil || s.exportRepo == nil {
		return domain.ExportJob{}, ErrExportRepoMissing
	}
	job, err := s.exportRepo.GetByDownloadToken(ctx, strings.TrimSpace(exportID), strings.TrimSpace(token))
	if err != nil {
		return domain.ExportJob{}, err
	}
	if !job.ExpiresAt.IsZero() && s.now().After(job.ExpiresAt) {
		return domain.ExportJob{}, domain.ErrExportJobNotFound
	}
	return job, nil
}

// BuildSignedURL renders the externally visible download URL for a job.
func (s *Service) BuildSignedURL(job domain.ExportJob) string {
	if s == nil || s.signedURL == nil {
		return ""
	}
	return s.signedURL(job)
}

func (s *Service) fetchAll(ctx context.Context, baseQuery domain.ListEventsQuery) ([]domain.Event, error) {
	const pageSize = 200
	var all []domain.Event
	cursor := domain.Cursor{}
	for {
		q := baseQuery
		q.Limit = pageSize
		q.Cursor = cursor
		page, err := s.reader.ListEvents(ctx, q)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		all = append(all, page...)
		if len(page) < pageSize {
			break
		}
		last := page[len(page)-1]
		cursor = domain.Cursor{OccurredAt: last.OccurredAt, ID: last.ID}
	}
	return all, nil
}

func filterPayload(q domain.ListEventsQuery) map[string]any {
	out := map[string]any{}
	if v := strings.TrimSpace(q.ActorID); v != "" {
		out["actorId"] = v
	}
	if v := strings.TrimSpace(q.Action); v != "" {
		out["action"] = v
	}
	if v := strings.TrimSpace(q.ResourceType); v != "" {
		out["resourceType"] = v
	}
	if v := strings.TrimSpace(q.ResourceID); v != "" {
		out["resourceId"] = v
	}
	if !q.OccurredAfter.IsZero() {
		out["occurredAfter"] = q.OccurredAfter.UTC().Format(time.RFC3339)
	}
	if !q.OccurredBefore.IsZero() {
		out["occurredBefore"] = q.OccurredBefore.UTC().Format(time.RFC3339)
	}
	if v := strings.TrimSpace(q.Query); v != "" {
		out["q"] = v
	}
	return out
}

func filterSummary(q domain.ListEventsQuery) string {
	parts := []string{}
	for k, v := range filterPayload(q) {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	if len(parts) == 0 {
		return "(no filter)"
	}
	return strings.Join(parts, ",")
}

func randomToken() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return strings.ReplaceAll(uuid.NewString(), "-", "")
	}
	return hex.EncodeToString(buf)
}
