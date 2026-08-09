package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	v2dom "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/infrastructure"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	tmpldom "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/db"
)

// FreezeFinalizer writes the freeze marker (values_hash + values_frozen_at +
// the frozen_revision_id lineage pin) onto a document in a single update.
//
// NOTE on the `documentID` parameter naming below: the freeze pipeline's
// identity parameter has always carried documents.id, even where older code
// named it revisionID (see RepairMaterialization's note). These ports name it
// documentID so the ONE genuine document_revisions.id in this file — the
// frozen-revision pin — cannot be confused with it.
type FreezeFinalizer interface {
	// WriteFreeze stamps values_hash + values_frozen_at + the frozen_revision_id
	// lineage pin in a single update. frozenRevisionID is a
	// document_revisions.id; empty writes a NULL pin.
	WriteFreeze(ctx context.Context, tenantID, documentID string, valuesHash []byte, frozenRevisionID string, frozenAt time.Time, q ...infrastructure.DBTX) error
}

// SnapshotReader reads a document's template snapshot, freeze timestamp, and
// current/frozen revision references used by Pin and Materialize.
type SnapshotReader interface {
	ReadSnapshotWithFreezeAt(ctx context.Context, tenantID, documentID string, q ...infrastructure.DBTX) (v2dom.TemplateSnapshot, *time.Time, error)
	ReadFreezeAt(ctx context.Context, tenantID, documentID string, q ...infrastructure.DBTX) (*time.Time, error)
	// ReadCurrentRevisionRef returns the document's current editor revision —
	// the body Pin pins (F-QA3-1, option (a)). A zero RevisionRef means "no
	// current revision".
	ReadCurrentRevisionRef(ctx context.Context, tenantID, documentID string, q ...infrastructure.DBTX) (v2dom.RevisionRef, error)
	// ReadFrozenRevisionRef returns the revision PINNED at freeze time
	// (documents.frozen_revision_id). Materialize reads this one, never the
	// current head. A zero RevisionRef means "nothing was pinned".
	ReadFrozenRevisionRef(ctx context.Context, tenantID, documentID string, q ...infrastructure.DBTX) (v2dom.RevisionRef, error)
}

// RevisionBodyReader fetches the raw bytes of a stored revision body so
// Materialize can verify them against the pinned revision's recorded hash
// before anything is rendered. Satisfied by *objectstore.VerifiedStore; the
// tenant-asserted read is deliberate (defense in depth above the SQL tenant
// scoping — a mis-scoped key must never fetch another tenant's body).
type RevisionBodyReader interface {
	AssertedReadObject(ctx context.Context, tenantID, key string, maxBytes int64) ([]byte, error)
}

// maxRevisionBodyBytes bounds the verification read, mirroring the 25 MiB cap
// every composition root passes to objectstore.NewVerifiedStore.
const maxRevisionBodyBytes = 25 * 1024 * 1024

// ErrFrozenBodyHashMismatch is the distinct failure for the fail-closed lineage
// check: the bytes fetched for the pinned revision do not hash to that
// revision's recorded content_hash. It is NOT a transient error — retrying
// cannot fix substituted or corrupted source bytes — and no artifact is
// produced when it fires.
var ErrFrozenBodyHashMismatch = errors.New("frozen body hash mismatch")

// FanoutClient dispatches the frozen document body plus resolved placeholder
// values to the docx-renderer fanout for materialization.
type FanoutClient interface {
	Fanout(ctx context.Context, req fanout.Request) (fanout.Response, error)
}

// materializeDispatchEnqueuer is the minimal published interface for the
// staging materialize dispatch Enqueuer (render/fanout/dispatchjobs), owned
// here (the consumer) and satisfied by *dispatchjobs.Enqueuer. It inserts
// the paired (outbox row, River job) atomically inside tx (M5 F5.3 T3).
type materializeDispatchEnqueuer interface {
	EnqueueMaterializeTx(ctx context.Context, tx db.Tx, tenantID, documentID string, valuesHash []byte, releaseGenerationID string) error
}

// MaterializeResult is returned by Materialize after a successful fanout call.
type MaterializeResult struct {
	FinalDocxS3Key string
	ContentHash    []byte
}

// FreezeService implements the async freeze split (ADR 0015): Pin validates,
// resolves computed placeholders, and writes the freeze marker in-tx; the
// separate Materialize step later renders the pinned revision via the
// docx-renderer fanout.
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
	bodies            RevisionBodyReader
}

// ApproverContext carries the acting approver's identity and capabilities
// into computed-placeholder resolution during Pin.
type ApproverContext struct {
	UserID       string
	Capabilities []string
}

// ResolverContextBuilder builds the resolvers.ResolveInput used to resolve
// computed placeholders, either for a real approver (Build) or for draft-time
// preview resolution with no approver (BuildForDraft).
type ResolverContextBuilder interface {
	Build(ctx context.Context, tenantID, revisionID string, approver ApproverContext) (resolvers.ResolveInput, error)
	BuildForDraft(ctx context.Context, tenantID, revisionID string) (resolvers.ResolveInput, error)
}

var _ FanoutClient = (*fanout.Client)(nil)

// NewFreezeService constructs a FreezeService wired to its schema reader,
// value reader/writer, computed-placeholder resolver registry, freeze
// finalizer, resolver-context builder, snapshot reader, and fanout client.
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

// WithRevisionBodyReader wires the blob reader Materialize uses to verify the
// pinned revision's bytes against its recorded hash. Required by Materialize
// (which fails closed without it) and unused by Pin, so composition roots that
// only pin — the API and the release-backfill tool — need not supply one.
func (s *FreezeService) WithRevisionBodyReader(bodies RevisionBodyReader) *FreezeService {
	s.bodies = bodies
	return s
}

// pinValidateAndHash is the Pin setup path: validates required placeholders,
// resolves computed ones, computes values_hash, and writes the freeze marker —
// including the frozen_revision_id lineage pin — inside tx. Returns the
// resolved valMap and schema.
func (s *FreezeService) pinValidateAndHash(
	ctx context.Context, tx db.Tx, tenantID, documentID string, approver ApproverContext,
) (map[string]any, []tmpldom.Placeholder, error) {
	schema, err := s.schemas.LoadPlaceholderSchema(ctx, tenantID, documentID)
	if err != nil {
		return nil, nil, err
	}
	existing, err := s.valuesRead.ListValues(ctx, tenantID, documentID)
	if err != nil {
		return nil, nil, err
	}
	byID := map[string]infrastructure.PlaceholderValue{}
	for _, v := range existing {
		byID[v.PlaceholderID] = v
	}

	if err := requireNonComputedPlaceholdersSet(schema, byID); err != nil {
		return nil, nil, err
	}

	resolveIn, err := s.resolveCtx.Build(ctx, tenantID, documentID, approver)
	if err != nil {
		return nil, nil, err
	}
	if err := s.resolveComputedPlaceholders(ctx, tx, tenantID, documentID, schema, byID, resolveIn); err != nil {
		return nil, nil, err
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
	// Lineage pin (F-QA3-1 remainder): record WHICH document_revisions row this
	// freeze covers, in the SAME update as values_hash. Read in-tx so the pin
	// and the hash describe one consistent instant — Materialize later renders
	// this exact revision instead of re-reading a head that may have moved.
	// A document with no current revision pins NULL: the honest record, which
	// Materialize refuses rather than substituting another body.
	current, err := s.snapshots.ReadCurrentRevisionRef(ctx, tenantID, documentID, tx)
	if err != nil {
		return nil, nil, fmt.Errorf("read current revision for freeze pin: %w", err)
	}
	if err := s.finalize.WriteFreeze(ctx, tenantID, documentID, hashBytes, current.ID, time.Now().UTC(), tx); err != nil {
		return nil, nil, err
	}
	return valMap, schema, nil
}

// isComputedPlaceholder reports whether p is a computed placeholder, either
// via the explicit Computed flag or the PHComputed type.
func isComputedPlaceholder(p tmpldom.Placeholder) bool {
	return p.Computed || p.Type == tmpldom.PHComputed
}

// requireNonComputedPlaceholdersSet enforces that every required,
// non-computed placeholder in schema already has a non-empty author value in
// byID, ahead of computed-placeholder resolution.
func requireNonComputedPlaceholdersSet(schema []tmpldom.Placeholder, byID map[string]infrastructure.PlaceholderValue) error {
	for _, p := range schema {
		if !p.Required || isComputedPlaceholder(p) {
			continue
		}
		v, ok := byID[p.ID]
		if !ok || v.ValueText == nil || *v.ValueText == "" {
			return fmt.Errorf("%w: placeholder %s required", v2dom.ErrValidationFailed, p.ID)
		}
	}
	return nil
}

// resolveComputedPlaceholders runs each computed placeholder's resolver,
// persists the resolved value (source="computed"), and updates byID in place
// so pinValidateAndHash's subsequent values_hash computation sees the
// resolved values.
func (s *FreezeService) resolveComputedPlaceholders(
	ctx context.Context, tx db.Tx, tenantID, documentID string,
	schema []tmpldom.Placeholder, byID map[string]infrastructure.PlaceholderValue, resolveIn resolvers.ResolveInput,
) error {
	for _, p := range schema {
		if !isComputedPlaceholder(p) {
			continue
		}
		if p.ResolverKey == nil {
			return fmt.Errorf("%w: placeholder %s computed without resolver_key",
				v2dom.ErrValidationFailed, p.ID)
		}
		r, ok := s.resolvers.Get(*p.ResolverKey)
		if !ok {
			return fmt.Errorf("%w: placeholder %s resolver %s",
				tmpldom.ErrUnknownResolver, p.ID, *p.ResolverKey)
		}
		rv, err := r.Resolve(ctx, resolveIn)
		if err != nil {
			return fmt.Errorf("resolver %s failed: %w", *p.ResolverKey, err)
		}
		strVal := fmt.Sprintf("%v", rv.Value)
		key, ver := *p.ResolverKey, rv.ResolverVer
		if err := s.values.UpsertValue(ctx, infrastructure.PlaceholderValue{
			TenantID: tenantID, RevisionID: documentID, PlaceholderID: p.ID,
			ValueText: &strVal, Source: "computed",
			ComputedFrom: &key, ResolverVersion: &ver,
			InputsHash: rv.InputsHash,
		}, tx); err != nil {
			return err
		}
		byID[p.ID] = infrastructure.PlaceholderValue{ValueText: &strVal}
	}
	return nil
}

// Pin is the in-transaction half of the async freeze split (ADR 0015).
// It validates, resolves computed placeholders, writes values_hash + frozen_at
// + the frozen_revision_id lineage pin, and enqueues a
// materialize_dispatch_outbox row — all inside tx.
// No network calls to docx-renderer. Fast and cheap.
// tx is mandatory (ADR 0015 amended by Wave Z Z-5).
//
// documentID is documents.id (the parameter the whole pin/fanout/worker chain
// calls "revisionID"; see RepairMaterialization's note below).
func (s *FreezeService) Pin(ctx context.Context, tx db.Tx, tenantID, documentID string, approver ApproverContext, releaseGenerationID string) error {
	if tx == nil {
		return fmt.Errorf("freeze_service: tx required (ADR 0015 amended by Wave Z Z-5)")
	}
	valuesFrozenAt, err := s.snapshots.ReadFreezeAt(ctx, tenantID, documentID, tx)
	if err != nil {
		return fmt.Errorf("pin: read freeze_at: %w", err)
	}
	if valuesFrozenAt != nil {
		return nil
	}

	valMap, _, err := s.pinValidateAndHash(ctx, tx, tenantID, documentID, approver)
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
	return s.materializeOutbox.EnqueueMaterializeTx(ctx, tx, tenantID, documentID, hashBytes, releaseGenerationID)
}

// RepairMaterialization re-pins the freeze lineage and re-enqueues the
// materialize dispatch for a document whose final artifact set is missing or
// incomplete (ADR 0085 Stage C, in-flight disposition). It is the operator-only
// repair path, reached exclusively through scripts/release-backfill.
//
// It does NOT re-run validation or placeholder resolution, and it does NOT
// recompute values_hash: that hash is what the approver signed, and recomputing
// it would silently replace approved values.
//
// It DOES take a fresh lineage pin (operator ruling, Etapa 5 §5.3): it re-reads
// the current revision ref in this same tx and writes frozen_revision_id
// through the same WriteFreeze seam Pin uses — no duplicated SQL. Repair is
// therefore an explicit, operator-sanctioned RE-FREEZE of the body from the
// current revision with TRUE lineage, never history fabrication: the pin names
// a revision that really is the body about to be rendered, and Materialize
// still verifies the fetched bytes against that revision's own content_hash.
// This is what makes the legacy NULL-pin documents (frozen before migration
// 0313) repairable at all; a migration-time backfill could only have guessed.
//
// A document that already carries a pin is re-pinned too: choosing to repair IS
// the decision to re-freeze, so the old pin is superseded rather than defended.
//
// values_frozen_at is carried over, not restamped: the approval instant is a
// fact about the past and repair does not get to rewrite it (and gates that
// read values_frozen_at as "is frozen" must not see this document flicker).
//
// It fails closed on a document that was never frozen (missing values_hash /
// values_frozen_at): there is no approved snapshot to render from, and the
// no-fallback principle forbids substituting a fresh freeze for one that was
// never taken. It also fails closed when the document has no current revision —
// there is no honest body to pin.
//
// The releaseGenerationID tags the dispatch so the produced artifacts are
// recorded against the right generation. Replaying with the SAME generation is
// a no-op: the staging outbox dedupe is generation-aware
// (render/fanout/staging_outbox.go), so no second dedupe is layered here.
//
// documentID is passed into the parameter named revisionID: that parameter
// carries the DOCUMENT id throughout the pin/fanout/worker chain (documents.id
// is the aggregate the freeze columns live on) — see the note at
// approval/application/release_terminal_approval.go:156-159. The convention is
// preserved verbatim, not "corrected", because the whole chain agrees on it.
func (s *FreezeService) RepairMaterialization(ctx context.Context, tx db.Tx, tenantID, documentID, releaseGenerationID string) error {
	if tx == nil {
		return fmt.Errorf("repair_materialization: tx required")
	}
	if s.materializeOutbox == nil {
		return fmt.Errorf("repair materialization: materialize outbox enqueuer not configured")
	}

	var (
		hashBytes []byte
		frozenAt  *time.Time
	)
	err := tx.QueryRowContext(ctx, `
		SELECT values_hash, values_frozen_at
		  FROM documents
		 WHERE id = $1::uuid AND tenant_id = $2::uuid`,
		documentID, tenantID,
	).Scan(&hashBytes, &frozenAt)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("repair materialization: document %s not found", documentID)
	}
	if err != nil {
		return fmt.Errorf("repair materialization: read freeze state: %w", err)
	}
	if frozenAt == nil || len(hashBytes) == 0 {
		return fmt.Errorf("repair materialization: document %s was never frozen", documentID)
	}

	// Fresh lineage pin, same tx, same seam as Pin.
	current, err := s.snapshots.ReadCurrentRevisionRef(ctx, tenantID, documentID, tx)
	if err != nil {
		return fmt.Errorf("repair materialization: read current revision for re-pin: %w", err)
	}
	if current.ID == "" {
		return fmt.Errorf("repair materialization: document %s has no current revision to pin", documentID)
	}
	if err := s.finalize.WriteFreeze(ctx, tenantID, documentID, hashBytes, current.ID, *frozenAt, tx); err != nil {
		return fmt.Errorf("repair materialization: re-pin frozen revision: %w", err)
	}

	return s.materializeOutbox.EnqueueMaterializeTx(ctx, tx, tenantID, documentID, hashBytes, releaseGenerationID)
}

// Materialize is the async half of the freeze split (ADR 0015).
// It reads the already-pinned placeholder values, calls the docx-renderer fanout,
// and returns the result so the caller can persist it transactionally.
// The caller (MaterializeJobRunner) is responsible for WriteFinalDocx + PDF enqueue.
//
// The rendered BODY is the revision PINNED at freeze time, not the template
// snapshot and not the current head (F-QA3-1, operator ruling option (a)): the
// approver signs the content they reviewed in the editor, and
// documents.frozen_revision_id records exactly which revision that was. The
// template snapshot still supplies the composition config and seeds the initial
// clone. Placeholder resolution is applied on top of that body, unchanged.
//
// Before anything is rendered, the fetched bytes are verified against the
// pinned revision's own document_revisions.content_hash. A mismatch fails the
// job with ErrFrozenBodyHashMismatch and produces NO artifact: silently
// rendering substituted bytes would put a signature on content nobody approved.
//
// documentID is documents.id (the parameter the chain calls "revisionID").
func (s *FreezeService) Materialize(ctx context.Context, tenantID, documentID string) (MaterializeResult, error) {
	if s.fanout == nil {
		return MaterializeResult{}, fmt.Errorf("materialize: fanout client not configured")
	}

	snap, frozenAt, err := s.snapshots.ReadSnapshotWithFreezeAt(ctx, tenantID, documentID)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: read snapshot: %w", err)
	}
	if frozenAt == nil {
		return MaterializeResult{}, fmt.Errorf("materialize: document %s not yet pinned", documentID)
	}

	placeholderVals, resolvedForSubblocks, err := s.resolveMaterializeValues(ctx, tenantID, documentID)
	if err != nil {
		return MaterializeResult{}, err
	}

	composition := snap.CompositionJSON
	if len(composition) == 0 {
		composition = json.RawMessage(`{}`)
	}

	// Pinned truth is frozen truth. Fail closed when nothing was pinned:
	// rendering the current head (or the template snapshot) instead would
	// produce a signed artifact whose provenance nobody can prove (F-QA3-1).
	pinned, err := s.snapshots.ReadFrozenRevisionRef(ctx, tenantID, documentID)
	if err != nil {
		return MaterializeResult{}, fmt.Errorf("materialize: read pinned revision: %w", err)
	}
	if pinned.ID == "" || pinned.StorageKey == "" {
		return MaterializeResult{}, fmt.Errorf(
			"materialize: document %s has no pinned frozen revision body to freeze", documentID)
	}
	if err := s.verifyPinnedBody(ctx, tenantID, documentID, pinned); err != nil {
		return MaterializeResult{}, err
	}

	resp, err := s.fanout.Fanout(ctx, fanout.Request{
		TenantID:          tenantID,
		RevisionID:        documentID,
		BodyDocxS3Key:     pinned.StorageKey,
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

// resolveMaterializeValues loads the frozen placeholder values for
// documentID and maps them into the two keying schemes Fanout expects:
// placeholderVals is keyed by placeholder NAME (falling back to id) for
// template substitution, resolvedForSubblocks stays keyed by placeholder ID
// for subblock resolution.
func (s *FreezeService) resolveMaterializeValues(ctx context.Context, tenantID, documentID string) (map[string]string, map[string]any, error) {
	schema, err := s.schemas.LoadPlaceholderSchema(ctx, tenantID, documentID)
	if err != nil {
		return nil, nil, fmt.Errorf("materialize: load schema: %w", err)
	}

	existing, err := s.valuesRead.ListValues(ctx, tenantID, documentID)
	if err != nil {
		return nil, nil, fmt.Errorf("materialize: list values: %w", err)
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

	return placeholderVals, resolvedForSubblocks, nil
}

// verifyPinnedBody fetches the pinned revision's stored bytes and proves they
// are the bytes that revision recorded at upload time. It is the fail-closed
// half of the lineage guarantee: the pin says WHICH revision was frozen, this
// says the object under that revision's key still IS that revision.
//
// A missing recorded hash is itself a failure, not a reason to skip the check —
// "no hash to compare against" is the exact shape of the silent-substitution
// hole this closes (no-fallback principle).
func (s *FreezeService) verifyPinnedBody(ctx context.Context, tenantID, documentID string, pinned v2dom.RevisionRef) error {
	if s.bodies == nil {
		return fmt.Errorf("materialize: revision body reader not configured (cannot verify frozen lineage)")
	}
	if pinned.ContentHash == "" {
		return fmt.Errorf("%w: document %s pinned revision %s has no recorded content_hash",
			ErrFrozenBodyHashMismatch, documentID, pinned.ID)
	}
	body, err := s.bodies.AssertedReadObject(ctx, tenantID, pinned.StorageKey, maxRevisionBodyBytes)
	if err != nil {
		return fmt.Errorf("materialize: read pinned revision body %s: %w", pinned.StorageKey, err)
	}
	sum := sha256.Sum256(body)
	actual := hex.EncodeToString(sum[:])
	if !strings.EqualFold(actual, pinned.ContentHash) {
		return fmt.Errorf("%w: document %s pinned revision %s key %s: computed %s, recorded %s",
			ErrFrozenBodyHashMismatch, documentID, pinned.ID, pinned.StorageKey, actual, pinned.ContentHash)
	}
	return nil
}
