package application

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	tmpldom "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

type FreezeFinalizer interface {
	WriteFreeze(ctx context.Context, tenantID, revisionID string, valuesHash []byte, frozenAt time.Time, q ...infrastructure.DBTX) error
}

type SnapshotReader interface {
	ReadSnapshotWithFreezeAt(ctx context.Context, tenantID, revisionID string, q ...infrastructure.DBTX) (v2dom.TemplateSnapshot, *time.Time, error)
	ReadFreezeAt(ctx context.Context, tenantID, revisionID string, q ...infrastructure.DBTX) (*time.Time, error)
	// ReadCurrentRevisionBodyKey returns the storage key of the document's
	// current editor revision — the body Materialize freezes (F-QA3-1,
	// option (a)). Empty string means "no current revision".
	ReadCurrentRevisionBodyKey(ctx context.Context, tenantID, revisionID string, q ...infrastructure.DBTX) (string, error)
}

type FanoutClient interface {
	Fanout(ctx context.Context, req fanout.FanoutRequest) (fanout.FanoutResponse, error)
}

// materializeDispatchEnqueuer is the minimal published interface for the
// staging materialize dispatch Enqueuer (render/fanout/dispatchjobs), owned
// here (the consumer) and satisfied by *dispatchjobs.Enqueuer. It inserts
// the paired (outbox row, River job) atomically inside tx (M5 F5.3 T3).
type materializeDispatchEnqueuer interface {
	EnqueueMaterializeTx(ctx context.Context, tx db.Tx, tenantID, revisionID string, contentHash []byte) error
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
		ListValues(ctx context.Context, tenantID, revisionID string) ([]infrastructure.PlaceholderValue, error)
	}
	resolvers         *resolvers.Registry
	finalize          FreezeFinalizer
	resolveCtx        ResolverContextBuilder
	snapshots         SnapshotReader
	fanout            FanoutClient
	materializeOutbox materializeDispatchEnqueuer
}

type ApproverContext struct {
	UserID       string
	Capabilities []string
}

type ResolverContextBuilder interface {
	Build(ctx context.Context, tenantID, revisionID string, approver ApproverContext) (resolvers.ResolveInput, error)
	BuildForDraft(ctx context.Context, tenantID, revisionID string) (resolvers.ResolveInput, error)
}

var _ FanoutClient = (*fanout.Client)(nil)

func NewFreezeService(
	schemas SchemaReader, values FillInWriter,
	valuesRead interface {
		ListValues(ctx context.Context, tenantID, revisionID string) ([]infrastructure.PlaceholderValue, error)
	},
	reg *resolvers.Registry, final FreezeFinalizer, ctxBuilder ResolverContextBuilder,
	snapshots SnapshotReader,
	fanoutClient FanoutClient,
) *FreezeService {
	return &FreezeService{
		schemas: schemas, values: values, valuesRead: valuesRead,
		resolvers: reg, finalize: final, resolveCtx: ctxBuilder,
		snapshots: snapshots,
		fanout:    fanoutClient,
	}
}

// WithMaterializeOutbox sets the transactional staging dispatch Enqueuer used
// by Pin. Takes the narrow materializeDispatchEnqueuer interface, satisfied
// by *dispatchjobs.Enqueuer (M5 F5.3 T3).
func (s *FreezeService) WithMaterializeOutbox(enqueuer materializeDispatchEnqueuer) *FreezeService {
	s.materializeOutbox = enqueuer
	return s
}

// pinValidateAndHash is the Pin setup path: validates required placeholders,
// resolves computed ones, computes values_hash, and writes the freeze marker
// inside tx. Returns the resolved valMap and schema.
func (s *FreezeService) pinValidateAndHash(
	ctx context.Context, tx db.Tx, tenantID, revisionID string, approver ApproverContext,
) (map[string]any, []tmpldom.Placeholder, error) {
	schema, err := s.schemas.LoadPlaceholderSchema(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	existing, err := s.valuesRead.ListValues(ctx, tenantID, revisionID)
	if err != nil {
		return nil, nil, err
	}
	byID := map[string]infrastructure.PlaceholderValue{}
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
		if err := s.values.UpsertValue(ctx, infrastructure.PlaceholderValue{
			TenantID: tenantID, RevisionID: revisionID, PlaceholderID: p.ID,
			ValueText: &strVal, Source: "computed",
			ComputedFrom: &key, ResolverVersion: &ver,
			InputsHash: rv.InputsHash,
		}, tx); err != nil {
			return nil, nil, err
		}
		byID[p.ID] = infrastructure.PlaceholderValue{ValueText: &strVal}
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
	if err := s.finalize.WriteFreeze(ctx, tenantID, revisionID, hashBytes, time.Now().UTC(), tx); err != nil {
		return nil, nil, err
	}
	return valMap, schema, nil
}

// Pin is the in-transaction half of the async freeze split (ADR 0015).
// It validates, resolves computed placeholders, writes values_hash + frozen_at,
// and enqueues a materialize_dispatch_outbox row — all inside tx.
// No network calls to docx-renderer. Fast and cheap.
// tx is mandatory (ADR 0015 amended by Wave Z Z-5).
func (s *FreezeService) Pin(ctx context.Context, tx db.Tx, tenantID, revisionID string, approver ApproverContext) error {
	if tx == nil {
		return fmt.Errorf("freeze_service: tx required (ADR 0015 amended by Wave Z Z-5)")
	}
	valuesFrozenAt, err := s.snapshots.ReadFreezeAt(ctx, tenantID, revisionID, tx)
	if err != nil {
		return fmt.Errorf("pin: read freeze_at: %w", err)
	}
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
	return s.materializeOutbox.EnqueueMaterializeTx(ctx, tx, tenantID, revisionID, hashBytes)
}

// Materialize is the async half of the freeze split (ADR 0015).
// It reads the already-pinned placeholder values, calls the docx-renderer fanout,
// and returns the result so the caller can persist it transactionally.
// The caller (MaterializeJobRunner) is responsible for WriteFinalDocx + PDF enqueue.
//
// The rendered BODY is the document's current editor revision, not the template
// snapshot (F-QA3-1, operator ruling option (a)): the approver signs the content
// they reviewed in the editor. The template snapshot still supplies the
// composition config and seeds the initial clone. Placeholder resolution is
// applied on top of the editor body, unchanged.
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

	// Editor truth is frozen truth. Fail closed when the document has no current
	// revision: rendering the template snapshot instead would produce a signed
	// artifact that does not carry the reviewed content (F-QA3-1).
	bodyKey, err := s.snapshots.ReadCurrentRevisionBodyKey(ctx, tenantID, revisionID)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: read current revision body: %w", err)
	}
	if bodyKey == "" {
		return MaterializeResult{}, fmt.Errorf(
			"materialize: document %s has no current revision body to freeze", revisionID)
	}

	resp, err := s.fanout.Fanout(ctx, fanout.FanoutRequest{
		TenantID:          tenantID,
		RevisionID:        revisionID,
		BodyDocxS3Key:     bodyKey,
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
