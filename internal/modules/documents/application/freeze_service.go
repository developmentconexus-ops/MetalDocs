package application

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	tmpldom "metaldocs/internal/modules/templates/domain"
)

type FreezeFinalizer interface {
	WriteFreeze(ctx context.Context, tenantID, revisionID string, valuesHash []byte, frozenAt time.Time, q ...repository.DBTX) error
}

type SnapshotReader interface {
	ReadSnapshotWithFreezeAt(ctx context.Context, tenantID, revisionID string, q ...repository.DBTX) (v2dom.TemplateSnapshot, *time.Time, error)
}

type FinalDocxWriter interface {
	WriteFinalDocx(ctx context.Context, tenantID, revisionID, s3Key string, contentHash []byte, q ...repository.DBTX) error
}

type FanoutClient interface {
	Fanout(ctx context.Context, req fanout.FanoutRequest) (fanout.FanoutResponse, error)
}

// MaterializeOutboxEnqueuer enqueues an async materialize job inside the Pin transaction.
type MaterializeOutboxEnqueuer interface {
	Enqueue(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, contentHash []byte) error
}

// MaterializeResult is returned by Materialize after a successful fanout call.
type MaterializeResult struct {
	FinalDocxS3Key string
	ContentHash    []byte
}

type FreezeService struct {
	schemas    SchemaReader
	values     FillInWriter
	valuesRead interface {
		ListValues(ctx context.Context, tenantID, revisionID string) ([]repository.PlaceholderValue, error)
	}
	resolvers        *resolvers.Registry
	finalize         FreezeFinalizer
	resolveCtx       ResolverContextBuilder
	snapshots        SnapshotReader
	finalDocx        FinalDocxWriter
	fanout           FanoutClient
	materializeOutbox MaterializeOutboxEnqueuer
}

type ApproverContext struct {
	UserID       string
	Capabilities []string
}

type ResolverContextBuilder interface {
	Build(ctx context.Context, tenantID, revisionID string, approver ApproverContext) (resolvers.ResolveInput, error)
	BuildForDraft(ctx context.Context, tenantID, revisionID string) (resolvers.ResolveInput, error)
}

func NewFreezeService(
	schemas SchemaReader, values FillInWriter,
	valuesRead interface {
		ListValues(ctx context.Context, tenantID, revisionID string) ([]repository.PlaceholderValue, error)
	},
	reg *resolvers.Registry, final FreezeFinalizer, ctxBuilder ResolverContextBuilder,
	snapshots SnapshotReader, finalDocx FinalDocxWriter,
	fanoutClient any, legacyFanout ...FanoutClient,
) *FreezeService {
	fc, _ := fanoutClient.(FanoutClient)
	if len(legacyFanout) > 0 {
		fc = legacyFanout[0]
	}
	return &FreezeService{
		schemas: schemas, values: values, valuesRead: valuesRead,
		resolvers: reg, finalize: final, resolveCtx: ctxBuilder,
		snapshots: snapshots, finalDocx: finalDocx,
		fanout: fc,
	}
}

// WithMaterializeOutbox sets the transactional outbox enqueuer used by Pin.
func (s *FreezeService) WithMaterializeOutbox(enqueuer MaterializeOutboxEnqueuer) *FreezeService {
	s.materializeOutbox = enqueuer
	return s
}

// pinValidateAndHash is the shared setup path for both Pin and Freeze:
// validates required placeholders, resolves computed ones, computes values_hash,
// and writes the freeze marker inside tx. Returns the resolved valMap and schema.
func (s *FreezeService) pinValidateAndHash(
	ctx context.Context, tx *sql.Tx, tenantID, revisionID string, approver ApproverContext,
) (map[string]any, []tmpldom.Placeholder, error) {
	schema, err := s.schemas.LoadPlaceholderSchema(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	existing, err := s.valuesRead.ListValues(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	byID := map[string]repository.PlaceholderValue{}
	for _, v := range existing {
		byID[v.PlaceholderID] = v
	}

	isComputed := func(p tmpldom.Placeholder) bool {
		return p.Computed || p.Type == tmpldom.PHComputed
	}

	for _, p := range schema {
		if !p.Required || isComputed(p) {
			continue
		}
		v, ok := byID[p.ID]
		if !ok || v.ValueText == nil || *v.ValueText == "" {
			return nil, nil, fmt.Errorf("%w: placeholder %s required", v2dom.ErrValidationFailed, p.ID)
		}
	}

	resolveIn, err := s.resolveCtx.Build(ctx, tenantID, revisionID, approver)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range schema {
		if !isComputed(p) {
			continue
		}
		if p.ResolverKey == nil {
			return nil, nil, fmt.Errorf("%w: placeholder %s computed without resolver_key",
				v2dom.ErrValidationFailed, p.ID)
		}
		r, ok := s.resolvers.Get(*p.ResolverKey)
		if !ok {
			return nil, nil, fmt.Errorf("%w: placeholder %s resolver %s",
				tmpldom.ErrUnknownResolver, p.ID, *p.ResolverKey)
		}
		rv, err := r.Resolve(ctx, resolveIn)
		if err != nil {
			return nil, nil, fmt.Errorf("resolver %s failed: %w", *p.ResolverKey, err)
		}
		strVal := fmt.Sprintf("%v", rv.Value)
		key, ver := *p.ResolverKey, rv.ResolverVer
		if err := s.values.UpsertValue(ctx, repository.PlaceholderValue{
			TenantID: tenantID, RevisionID: revisionID, PlaceholderID: p.ID,
			ValueText: &strVal, Source: "computed",
			ComputedFrom: &key, ResolverVersion: &ver,
			InputsHash: rv.InputsHash,
		}, tx); err != nil {
			return nil, nil, err
		}
		byID[p.ID] = repository.PlaceholderValue{ValueText: &strVal}
	}

	valMap := make(map[string]any, len(byID))
	for _, p := range schema {
		if v, ok := byID[p.ID]; ok && v.ValueText != nil {
			valMap[p.ID] = *v.ValueText
		}
	}
	hashHex, err := v2dom.ComputeValuesHash(valMap)
	if err != nil {
		return nil, nil, fmt.Errorf("compute values_hash: %w", err)
	}
	hashBytes, err := hex.DecodeString(hashHex)
	if err != nil {
		return nil, nil, fmt.Errorf("decode values_hash: %w", err)
	}
	if tx != nil {
		if err := s.finalize.WriteFreeze(ctx, tenantID, revisionID, hashBytes, time.Now().UTC(), tx); err != nil {
			return nil, nil, err
		}
	} else {
		if err := s.finalize.WriteFreeze(ctx, tenantID, revisionID, hashBytes, time.Now().UTC()); err != nil {
			return nil, nil, err
		}
	}
	return valMap, schema, nil
}

// Pin is the in-transaction half of the async freeze split (ADR 0015).
// It validates, resolves computed placeholders, writes values_hash + frozen_at,
// and enqueues a materialize_dispatch_outbox row — all inside tx.
// No network calls to docx-renderer. Fast and cheap.
func (s *FreezeService) Pin(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, approver ApproverContext) error {
	snap, valuesFrozenAt, err := s.snapshots.ReadSnapshotWithFreezeAt(ctx, tenantID, revisionID, tx)
	if err != nil {
		return fmt.Errorf("pin: read snapshot: %w", err)
	}
	_ = snap
	if valuesFrozenAt != nil {
		return nil
	}

	valMap, _, err := s.pinValidateAndHash(ctx, tx, tenantID, revisionID, approver)
	if err != nil {
		return fmt.Errorf("pin: %w", err)
	}

	hashHex, err := v2dom.ComputeValuesHash(valMap)
	if err != nil {
		return fmt.Errorf("pin: compute values_hash for outbox: %w", err)
	}
	hashBytes, err := hex.DecodeString(hashHex)
	if err != nil {
		return fmt.Errorf("pin: decode values_hash for outbox: %w", err)
	}

	if s.materializeOutbox == nil {
		return fmt.Errorf("pin: materialize outbox enqueuer not configured")
	}
	return s.materializeOutbox.Enqueue(ctx, tx, tenantID, revisionID, hashBytes)
}

// Materialize is the async half of the freeze split (ADR 0015).
// It reads the already-pinned placeholder values, calls the docx-renderer fanout,
// and returns the result so the caller can persist it transactionally.
// The caller (MaterializeJobRunner) is responsible for WriteFinalDocx + PDF enqueue.
func (s *FreezeService) Materialize(ctx context.Context, tenantID, revisionID string) (MaterializeResult, error) {
	if s.fanout == nil {
		return MaterializeResult{}, fmt.Errorf("materialize: fanout client not configured")
	}

	snap, frozenAt, err := s.snapshots.ReadSnapshotWithFreezeAt(ctx, tenantID, revisionID)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: read snapshot: %w", err)
	}
	if frozenAt == nil {
		return MaterializeResult{}, fmt.Errorf("materialize: revision %s not yet pinned", revisionID)
	}

	schema, err := s.schemas.LoadPlaceholderSchema(ctx, tenantID, revisionID)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: load schema: %w", err)
	}

	existing, err := s.valuesRead.ListValues(ctx, tenantID, revisionID)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: list values: %w", err)
	}

	byID := map[string]string{}
	for _, v := range existing {
		if v.ValueText != nil {
			byID[v.PlaceholderID] = *v.ValueText
		}
	}

	idToName := make(map[string]string, len(schema))
	for _, p := range schema {
		if p.Name != "" {
			idToName[p.ID] = p.Name
		}
	}

	placeholderVals := map[string]string{}
	resolvedForSubblocks := map[string]any{}
	for id, sv := range byID {
		key := id
		if n, ok := idToName[id]; ok {
			key = n
		}
		placeholderVals[key] = sv
		resolvedForSubblocks[id] = sv
	}

	composition := snap.CompositionJSON
	if len(composition) == 0 {
		composition = json.RawMessage(`{}`)
	}

	resp, err := s.fanout.Fanout(ctx, fanout.FanoutRequest{
		TenantID:          tenantID,
		RevisionID:        revisionID,
		BodyDocxS3Key:     snap.BodyDocxS3Key,
		PlaceholderValues: placeholderVals,
		Composition:       json.RawMessage(composition),
		ResolvedValues:    resolvedForSubblocks,
	})
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: fanout: %w", err)
	}

	contentHashBytes, err := hex.DecodeString(resp.ContentHash)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: decode content_hash: %w", err)
	}
	return MaterializeResult{
		FinalDocxS3Key: resp.FinalDocxS3Key,
		ContentHash:    contentHashBytes,
	}, nil
}

// Freeze is the original synchronous implementation kept for backward compatibility.
// New code should use Pin (in-tx) + Materialize (async worker) instead.
func (s *FreezeService) Freeze(ctx context.Context, tx *sql.Tx, tenantID, revisionID string, approver ApproverContext) error {
	var (
		snap           v2dom.TemplateSnapshot
		valuesFrozenAt *time.Time
		err            error
	)
	if tx != nil {
		snap, valuesFrozenAt, err = s.snapshots.ReadSnapshotWithFreezeAt(ctx, tenantID, revisionID, tx)
	} else {
		snap, valuesFrozenAt, err = s.snapshots.ReadSnapshotWithFreezeAt(ctx, tenantID, revisionID)
	}
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}
	if valuesFrozenAt != nil {
		return nil
	}

	valMap, schema, err := s.pinValidateAndHash(ctx, tx, tenantID, revisionID, approver)
	if err != nil {
		return err
	}

	idToName := make(map[string]string, len(schema))
	for _, p := range schema {
		if p.Name != "" {
			idToName[p.ID] = p.Name
		}
	}

	placeholderVals := map[string]string{}
	resolvedForSubblocks := map[string]any{}
	for id, v := range valMap {
		if sv, ok := v.(string); ok {
			key := id
			if n, ok := idToName[id]; ok {
				key = n
			}
			placeholderVals[key] = sv
			resolvedForSubblocks[id] = sv
		}
	}

	composition := snap.CompositionJSON
	if len(composition) == 0 {
		composition = json.RawMessage(`{}`)
	}

	if s.fanout == nil {
		return fmt.Errorf("freeze: fanout client not configured")
	}

	resp, err := s.fanout.Fanout(ctx, fanout.FanoutRequest{
		TenantID:          tenantID,
		RevisionID:        revisionID,
		BodyDocxS3Key:     snap.BodyDocxS3Key,
		PlaceholderValues: placeholderVals,
		Composition:       json.RawMessage(composition),
		ResolvedValues:    resolvedForSubblocks,
	})
	if err != nil {
		return fmt.Errorf("fanout: %w", err)
	}

	contentHashBytes, err := hex.DecodeString(resp.ContentHash)
	if err != nil {
		return fmt.Errorf("decode content_hash: %w", err)
	}
	if tx != nil {
		return s.finalDocx.WriteFinalDocx(ctx, tenantID, revisionID, resp.FinalDocxS3Key, contentHashBytes, tx)
	}
	return s.finalDocx.WriteFinalDocx(ctx, tenantID, revisionID, resp.FinalDocxS3Key, contentHashBytes)
}
