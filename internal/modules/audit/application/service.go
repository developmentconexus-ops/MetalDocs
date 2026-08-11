package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"metaldocs/internal/modules/audit/domain"
)

var (
	// ErrTenantRequired is returned when a query or export request carries no
	// tenant id — the application layer refuses to run an unscoped audit query.
	ErrTenantRequired = errors.New("audit: tenant id is required")
	// ErrReaderRequired is returned by NewService (as a panic) and by any method
	// called on a Service constructed without a domain.Reader.
	ErrReaderRequired = errors.New("audit: reader is required")
	// ErrActorRequired is returned when an export or status/download call carries
	// no actor id — ownership scoping cannot be enforced without one.
	ErrActorRequired = errors.New("audit: actor id is required")
	// ErrInvalidFormat is returned when the requested export format is neither
	// csv nor jsonl.
	ErrInvalidFormat = errors.New("audit: invalid export format")
	// ErrExportTooLarge is returned when the estimated row count for an export
	// filter exceeds SyncExportRowLimit; the synchronous export path refuses
	// rather than running an unbounded query.
	ErrExportTooLarge = errors.New("audit: export result set too large for synchronous export")
	// ErrExportRepoMissing is returned when export operations are invoked on a
	// Service that has not been wired via WithExports.
	ErrExportRepoMissing = errors.New("audit: export job repository not configured")
	// ErrCounterMissing is returned when ExportEvents needs to size a query but
	// no domain.Counter has been wired via WithExports.
	ErrCounterMissing = errors.New("audit: counter not configured for export sizing")
	// ErrExportTokenMismatch is returned when a download request's token does not
	// match the export job's stored DownloadToken.
	ErrExportTokenMismatch = errors.New("audit: invalid export download token")
	// ErrExportsDisabled is returned when export operations are invoked on a
	// Service whose writer dependency has not been wired via WithExports.
	ErrExportsDisabled = errors.New("audit: export pipeline not configured")
	// ErrExportTenantErased is returned when ExportEvents re-checks erasure
	// status immediately before persisting the export job (see
	// refuseIfTenantErased) and finds the tenant has been erased — or the
	// check itself failed — since the export began. Fail-closed: "could not tell"
	// is treated the same as "erased". The accumulated payload is discarded
	// and no audit_export_jobs row is written (PR #121 review round 1, P1 —
	// export-persistence half).
	ErrExportTenantErased = errors.New("audit: tenant was erased during export; export discarded")
)

// SyncExportRowLimit is the threshold above which an export request is
// rejected — async worker lands in a later PR.
const SyncExportRowLimit int64 = 50_000

// ExportTTL is how long a generated export's signed download URL stays valid.
const ExportTTL = 15 * time.Minute

// SignedURLBuilder turns a stored export job into the externally visible signed
// download URL.
type SignedURLBuilder func(job domain.ExportJob) string

// Service implements the audit module's application use cases (list, export,
// export status/download) against the domain ports. The export pipeline
// (counter/exportRepo/writer/signedURL) is optional and only becomes active
// once WithExports has been called; until then export methods fail closed
// with ErrExportsDisabled / ErrExportRepoMissing / ErrCounterMissing.
type Service struct {
	reader     domain.Reader
	counter    domain.Counter
	exportRepo domain.ExportJobRepository
	writer     domain.Writer
	signedURL  SignedURLBuilder
	erasure    domain.ErasureChecker
	now        func() time.Time
}

// NewService constructs a Service backed by reader. Panics if reader is nil —
// a Service with no read port cannot serve any of its use cases.
func NewService(reader domain.Reader) *Service {
	if reader == nil {
		panic(ErrReaderRequired.Error())
	}
	return &Service{reader: reader, now: func() time.Time { return time.Now().UTC() }}
}

// WithExports wires the export pipeline. All four dependencies are required;
// passing nil for any of them panics.
// writer's embedded domain.ErasureChecker doubles as ExportEvents' pre-persist
// erasure re-check (see refuseIfTenantErased): since domain.Writer requires
// IsErased, wiring exports can never leave the erasure gate unset — there is
// no separate optional dependency for a composition root or test harness to
// omit (PR #121 review round 1, P1 — fail-open remediation superseding the
// prior WithErasureCheck setter, which this replaces).
func (s *Service) WithExports(counter domain.Counter, repo domain.ExportJobRepository, writer domain.Writer, urlBuilder SignedURLBuilder) *Service {
	if counter == nil || repo == nil || writer == nil || urlBuilder == nil {
		panic("audit.WithExports: all export dependencies are required")
	}
	s.counter = counter
	s.exportRepo = repo
	s.writer = writer
	s.erasure = writer
	s.signedURL = urlBuilder
	return s
}

// ListEvents normalizes query (trimming filters, clamping limit to [1,100],
// defaulting to 50) and delegates to the underlying domain.Reader. Returns
// ErrTenantRequired when query.TenantID is empty.
func (s *Service) ListEvents(ctx context.Context, query domain.ListEventsQuery) ([]domain.Event, bool, error) {
	if s == nil || s.reader == nil {
		return nil, false, ErrReaderRequired
	}

	normalized, err := normalizeQuery(query)
	if err != nil {
		return nil, false, err
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
	if normalized.Limit > 100 {
		normalized.Limit = 100
	}
	return normalized, nil
}

// ExportEvents runs the export inline. Refuses with ErrExportTooLarge when row
// count exceeds SyncExportRowLimit. Returns ErrExportsDisabled when the export
// pipeline has not been wired via WithExports.
func (s *Service) ExportEvents(ctx context.Context, actorID string, format domain.ExportFormat, filter domain.ListEventsQuery) (domain.ExportJob, error) {
	if s == nil {
		return domain.ExportJob{}, ErrReaderRequired
	}
	if s.writer == nil {
		return domain.ExportJob{}, ErrExportsDisabled
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

	normalizedSizing, estimatedRows, err := s.sizeExport(ctx, filter)
	if err != nil {
		return domain.ExportJob{}, err
	}

	payload, actualRows, err := s.renderExportPayload(ctx, normalizedSizing, format)
	if err != nil {
		return domain.ExportJob{}, err
	}

	job, err := s.buildExportJob(actorID, format, filter, normalizedSizing.TenantID, estimatedRows, actualRows, payload)
	if err != nil {
		return domain.ExportJob{}, err
	}

	// EXPLICITLY TRANSITIONAL — deleted by ROADMAP unit 4.10, as part of that
	// unit's definition-of-done, not as a follow-up.
	//
	// This re-check narrows the erasure race; it does not close it. A window
	// remains between this IsErased round trip and the Save below. The reason
	// it cannot be closed here is that the defect is not in this module:
	// iam's runErase raises the erasure signal in its LAST phase, after the
	// destructive phases have already run, so this gate is reading a signal
	// that is absent for the entire interval it needs to cover. Unit 4.10
	// moves that tombstone to its own transaction committed before phase 1,
	// at which point this check has nothing left to narrow and goes away.
	if err := s.refuseIfTenantErased(ctx, normalizedSizing.TenantID); err != nil {
		return domain.ExportJob{}, err
	}

	if err := s.exportRepo.Save(ctx, job); err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: persist export job: %w", err)
	}

	s.recordExportRequested(ctx, job, filter, estimatedRows)

	return job, nil
}

// sizeExport normalizes filter into a cursor/limit-stripped sizing query and
// estimates its row count via s.counter. Refuses with ErrExportTooLarge when
// the estimate exceeds SyncExportRowLimit.
func (s *Service) sizeExport(ctx context.Context, filter domain.ListEventsQuery) (domain.ListEventsQuery, int64, error) {
	sizingFilter := filter
	sizingFilter.Limit = 0
	sizingFilter.Cursor = domain.Cursor{}
	normalizedSizing, err := normalizeQuery(sizingFilter)
	if err != nil {
		return domain.ListEventsQuery{}, 0, err
	}
	estimatedRows, err := s.counter.CountEvents(ctx, normalizedSizing)
	if err != nil {
		return domain.ListEventsQuery{}, 0, fmt.Errorf("audit: estimate export rows: %w", err)
	}
	if estimatedRows > SyncExportRowLimit {
		return domain.ListEventsQuery{}, 0, fmt.Errorf("%w: %d rows exceeds limit %d", ErrExportTooLarge, estimatedRows, SyncExportRowLimit)
	}
	return normalizedSizing, estimatedRows, nil
}

// renderExportPayload fetches every event matching normalizedSizing and
// renders them in the requested format, returning the rendered payload and
// the actual row count.
func (s *Service) renderExportPayload(ctx context.Context, normalizedSizing domain.ListEventsQuery, format domain.ExportFormat) ([]byte, int64, error) {
	events, err := s.fetchAll(ctx, normalizedSizing)
	if err != nil {
		return nil, 0, fmt.Errorf("audit: fetch export rows: %w", err)
	}
	payload, err := render(format, events)
	if err != nil {
		return nil, 0, fmt.Errorf("audit: render export: %w", err)
	}
	return payload, int64(len(events)), nil
}

// refuseIfTenantErased re-checks tenant erasure status immediately before
// ExportEvents persists its export job. renderExportPayload already fetched
// every row via a separate, earlier read; if the tenant's erasure commits in
// the interval between that fetch and this call, the accumulated payload is
// stale plaintext that must never reach durable storage (PR #121 review
// round 1, P1). s.erasure is guaranteed non-nil here: WithExports sets it
// (from the same writer it requires) together with s.writer, and ExportEvents
// already returned ErrExportsDisabled above if s.writer were nil — there is
// no reachable path into this function with a nil checker. Fails closed: a
// lookup error is treated the same as "erased" (ErrExportTenantErased either
// way), because no caller downstream of this function can tell "confirmed
// erased" apart from "could not tell" and safely choose to persist anyway.
// EXPLICITLY TRANSITIONAL — this whole function is deleted by ROADMAP unit
// 4.10 (erasure gate ordering). It exists only because the signal it consults
// is currently raised too late to be trustworthy; once iam commits the
// tombstone before the first destructive phase, the condition this guards
// against is unrepresentable and the guard is dead weight.
func (s *Service) refuseIfTenantErased(ctx context.Context, tenantID string) error {
	erased, err := s.erasure.IsErased(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("%w: erasure check failed: %w", ErrExportTenantErased, err)
	}
	if erased {
		return ErrExportTenantErased
	}
	return nil
}

// buildExportJob assembles the ExportJob record: marshals the (unnormalized)
// filter for storage, parses tenantID (the normalized sizing filter's tenant)
// into a uuid, and stamps a fresh ID/token/timestamps.
func (s *Service) buildExportJob(actorID string, format domain.ExportFormat, filter domain.ListEventsQuery, tenantID string, estimatedRows, actualRows int64, payload []byte) (domain.ExportJob, error) {
	filterJSON, err := json.Marshal(filterPayload(filter))
	if err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: marshal filter: %w", err)
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return domain.ExportJob{}, fmt.Errorf("audit: invalid tenant id %q: %w", tenantID, err)
	}

	now := s.now()
	return domain.ExportJob{
		ID:            uuid.NewString(),
		TenantID:      tenantUUID,
		ActorID:       actorID,
		Format:        format,
		FilterJSON:    string(filterJSON),
		Status:        domain.ExportStatusReady,
		DownloadToken: randomToken(),
		ExpiresAt:     now.Add(ExportTTL),
		EstimatedRows: estimatedRows,
		ActualRows:    actualRows,
		Payload:       payload,
		CreatedAt:     now,
		CompletedAt:   now,
	}, nil
}

// recordExportRequested emits the best-effort "audit.export.requested"
// governance event for job. The export job is already persisted by the time
// this runs, so a construction/record failure here is logged, not returned —
// but a silent drop would lose an audit-trail record with no trace.
func (s *Service) recordExportRequested(ctx context.Context, job domain.ExportJob, filter domain.ListEventsQuery, estimatedRows int64) {
	summary := map[string]any{
		"format":        string(job.Format),
		"filterSummary": filterSummary(filter),
		"estimatedRows": estimatedRows,
		"actualRows":    job.ActualRows,
		"export_id":     job.ID,
	}
	event, evErr := domain.NewEvent(job.TenantID.String(), "audit_export", job.ID, job.ActorID, "audit.export.requested", summary)
	if evErr != nil {
		return
	}
	if recErr := s.writer.Record(ctx, event); recErr != nil {
		slog.Warn("audit export governance event dropped", "export_id", job.ID, "err", recErr)
	}
}

// GetExportStatus returns an export job, enforcing tenant + actor scoping.
// Fail-closed on empty actorID — the prior "if non-empty, then check" pattern
// silently bypassed ownership when an internal caller forgot to thread the
// actor through. Edge layer rejects empty userIDs today; this is defense in
// depth.
func (s *Service) GetExportStatus(ctx context.Context, tenantID, actorID, exportID string) (domain.ExportJob, error) {
	if s == nil || s.exportRepo == nil {
		return domain.ExportJob{}, ErrExportRepoMissing
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return domain.ExportJob{}, ErrTenantRequired
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return domain.ExportJob{}, ErrActorRequired
	}
	job, err := s.exportRepo.Get(ctx, tenantID, strings.TrimSpace(exportID))
	if err != nil {
		return domain.ExportJob{}, err
	}
	if job.ActorID != actorID {
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
		page, hasMore, err := s.reader.ListEvents(ctx, q)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		all = append(all, page...)
		if !hasMore {
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
		out["actor_id"] = v
	}
	if v := strings.TrimSpace(q.Action); v != "" {
		out["action"] = v
	}
	if v := strings.TrimSpace(q.ResourceType); v != "" {
		out["resource_type"] = v
	}
	if v := strings.TrimSpace(q.ResourceID); v != "" {
		out["resource_id"] = v
	}
	if !q.OccurredAfter.IsZero() {
		out["occurred_after"] = q.OccurredAfter.UTC().Format(time.RFC3339)
	}
	if !q.OccurredBefore.IsZero() {
		out["occurred_before"] = q.OccurredBefore.UTC().Format(time.RFC3339)
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
