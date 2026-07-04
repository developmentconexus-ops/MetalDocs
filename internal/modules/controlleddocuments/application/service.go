// Package application holds the controlleddocuments module's use-case
// service: the atomic clone-template-into-document flow that creates a
// ControlledDocument row plus its initial document revision in one
// transaction, delegating cross-module work to published ports.
package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	platformdb "metaldocs/internal/platform/db"
)

// TemplateVersionChecker is the cross-module read port into the templates
// module used to validate an override template version's status and
// profile ownership before it is bound to a controlled document.
type TemplateVersionChecker interface {
	GetTemplateVersionState(ctx context.Context, tenantID, templateVersionID string) (*string, string, error)
}

// ProfileReader is the cross-module read port into taxonomy used to
// validate a controlled document's profile is active before create.
type ProfileReader interface {
	GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)
}

// AreaReader is the cross-module read port into taxonomy used to
// validate a controlled document's process area is active before create.
type AreaReader interface {
	GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error)
}

// ControlledDocument aliases the domain aggregate for callers that only
// import the application package.
type ControlledDocument = controlleddocumentsdomain.ControlledDocument

// CDFilter aliases the domain list filter for callers that only import
// the application package.
type CDFilter = controlleddocumentsdomain.CDFilter

// ControlledDocumentService is the controlled-documents module's
// application-layer entry point: it orchestrates profile/area validation,
// code allocation (manual or auto-sequence), the atomic CD+first-revision
// create (ADR 0011), status transitions, and best-effort governance
// logging. It is the sole authz boundary for these operations — the
// repository layer does not re-check (F-CD6).
type ControlledDocumentService struct {
	// Core dependencies.
	runner    platformdb.TxRunner
	docs      controlleddocumentsdomain.ControlledDocumentRepository
	seq       controlleddocumentsdomain.SequenceAllocator
	tplCheck  TemplateVersionChecker
	profiles  ProfileReader
	areas     AreaReader
	govLogger taxonomydomain.GovernanceLogger
	docInit   controlleddocumentsdomain.DocumentInitializer

	// Runtime configuration.
	now func() time.Time
}

// CreateControlledDocumentCmd is the input to Create: either ManualCode
// (with ManualCodeReason) or an empty ManualCode to auto-allocate via the
// profile/area sequence. OverrideTemplateVersionID and
// OverrideTemplateReason are optional and validated together.
type CreateControlledDocumentCmd struct {
	TenantID                  string
	ProfileCode               string
	ProcessAreaCode           string
	DepartmentCode            *string
	Title                     string
	OwnerUserID               string
	ActorUserID               string
	ManualCode                *string
	ManualCodeReason          *string
	OverrideTemplateVersionID *string
	OverrideTemplateReason    *string
	TemplateVersionID         *string
	DocumentName              string
	FormData                  map[string]any
	VisibilityScope           string
	VisibilityAreaCodes       []string
	VisibilityUserIDs         []string
}

// CreateResult is the atomic-create return: the persisted ControlledDocument
// plus the DocumentRef returned by DocumentInitializer.
type CreateResult struct {
	ControlledDocument *controlleddocumentsdomain.ControlledDocument
	DocumentRef        *controlleddocumentsdomain.DocumentRef
}

// ErrTemplateArtifactMissing is returned when the resolved template's S3
// storage key is empty or does not exist, so Create/PreviewCode cannot
// proceed.
var ErrTemplateArtifactMissing = errors.New("template artifact missing")

// ErrTemplateArtifactInvariantUnconfigured is returned when
// ensureTemplateArtifact runs with no DocumentInitializer wired
// (s.docInit == nil) — a wiring-order bug, not a user error.
var ErrTemplateArtifactInvariantUnconfigured = errors.New("controlled_documents: template artifact invariant not configured")

// ErrActorMissing signals the request context lacked an authenticated
// principal where the service requires one. Both read and mutation paths
// fail-closed on this — see C5 in
// wiki/reviews/2026-05-21-go-backend-review/platform-2a-security.md.
var ErrActorMissing = errors.New("controlled_documents: actor user id missing in context")

// NewControlledDocumentService wires a ControlledDocumentService. It
// panics if any dependency except docInit is nil (fail-loud by design);
// docInit may be nil at construction and wired later via
// WithDocumentInitializer to break the controlled-documents<->documents
// module init cycle.
func NewControlledDocumentService(
	runner platformdb.TxRunner,
	docs controlleddocumentsdomain.ControlledDocumentRepository,
	seq controlleddocumentsdomain.SequenceAllocator,
	tplCheck TemplateVersionChecker,
	profiles ProfileReader,
	areas AreaReader,
	govLogger taxonomydomain.GovernanceLogger,
	docInit controlleddocumentsdomain.DocumentInitializer,
) *ControlledDocumentService {
	if docs == nil {
		panic("controlled_documents: repository must not be nil")
	}
	if seq == nil {
		panic("controlled_documents: sequence allocator must not be nil")
	}
	if tplCheck == nil {
		panic("controlled_documents: template checker must not be nil")
	}
	if profiles == nil {
		panic("controlled_documents: profile reader must not be nil")
	}
	if areas == nil {
		panic("controlled_documents: area reader must not be nil")
	}
	if govLogger == nil {
		panic("controlled_documents: governance logger must not be nil")
	}
	svc := &ControlledDocumentService{
		runner:    runner,
		docs:      docs,
		seq:       seq,
		tplCheck:  tplCheck,
		profiles:  profiles,
		areas:     areas,
		govLogger: govLogger,
		now:       time.Now,
	}
	if docInit != nil {
		svc.WithDocumentInitializer(docInit)
	}
	return svc
}

// WithDocumentInitializer wires the DocumentInitializer adapter post-construction.
// Used by the wiring layer to break the controlled-documents<->documents module cycle: the
// controlled-documents module is constructed first (because documents needs a
// ControlledDocumentDuplicator), then the documents module is built, then the
// initializer adapter is injected back here.
func (s *ControlledDocumentService) WithDocumentInitializer(d controlleddocumentsdomain.DocumentInitializer) *ControlledDocumentService {
	if d == nil {
		panic("controlled_documents: document initializer must not be nil")
	}
	s.docInit = d
	return s
}

// Create validates cmd's profile and area are active, resolves the CD
// code (manual, with uniqueness check, or auto-allocated inside the
// atomic tx), and — when a DocumentInitializer is wired — clones the
// effective template into the first document revision in the same tx
// (ADR 0011). Governance events are logged best-effort after commit
// (H-PRE-1: logging inside the tx risks an advisory-lock deadlock).
func (s *ControlledDocumentService) Create(ctx context.Context, cmd CreateControlledDocumentCmd) (*CreateResult, error) {
	ctx, span := otel.Tracer("metaldocs/controlleddocuments").Start(ctx, "cd.create",
		oteltrace.WithAttributes(attribute.String("document.profile_code", cmd.ProfileCode)),
	)
	defer span.End()

	profile, err := s.profiles.GetByCode(ctx, cmd.TenantID, cmd.ProfileCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("controlled_documents: get profile: %w", err)
	}
	if !profile.IsActive() {
		return nil, taxonomydomain.ErrProfileArchived
	}

	area, err := s.areas.GetByCode(ctx, cmd.TenantID, cmd.ProcessAreaCode)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("controlled_documents: get process area: %w", err)
	}
	if !area.IsActive() {
		return nil, taxonomydomain.ErrAreaArchived
	}

	var (
		code       string
		sequence   *int
		events     []taxonomydomain.GovernanceEvent
		overrideID *string
		doc        *controlleddocumentsdomain.ControlledDocument
		docRef     *controlleddocumentsdomain.DocumentRef
	)

	if cmd.ManualCode != nil {
		if !isReasonValid(cmd.ManualCodeReason) {
			return nil, controlleddocumentsdomain.ErrManualCodeReasonRequired
		}
		code = strings.TrimSpace(*cmd.ManualCode)
		taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: check manual code availability: %w", err)
		}
		if taken {
			return nil, controlleddocumentsdomain.ErrCDCodeTaken
		}
		payload, err := json.Marshal(map[string]string{"code": code})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: marshal numbering override payload: %w", err)
		}
		events = append(events, taxonomydomain.GovernanceEvent{
			TenantID:     cmd.TenantID,
			EventType:    "numbering.override",
			ActorUserID:  cmd.ActorUserID,
			ResourceType: "controlled_document",
			ResourceID:   code,
			Reason:       strings.TrimSpace(*cmd.ManualCodeReason),
			PayloadJSON:  payload,
		})

		// Override-template validation (manual path).
		if cmd.OverrideTemplateVersionID != nil {
			if !isReasonValid(cmd.OverrideTemplateReason) {
				return nil, controlleddocumentsdomain.ErrOverrideReasonRequired
			}
			status, profileCode, err := s.tplCheck.GetTemplateVersionState(ctx, cmd.TenantID, *cmd.OverrideTemplateVersionID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, fmt.Errorf("controlled_documents: get override template version state: %w", err)
			}
			_, err = controlleddocumentsdomain.Resolve(controlleddocumentsdomain.TemplateResolutionInput{
				ProfileCode: cmd.ProfileCode,
				OverrideTemplate: &controlleddocumentsdomain.TemplateVersionCandidate{
					ID:          *cmd.OverrideTemplateVersionID,
					ProfileCode: profileCode,
					Status:      status,
				},
			})
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, fmt.Errorf("controlled_documents: resolve template version: %w", err)
			}
			overrideID = cmd.OverrideTemplateVersionID
		}
		if err := s.ensureTemplateArtifact(ctx, cmd); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: ensure template artifact: %w", err)
		}

		if overrideID != nil {
			payload, err := json.Marshal(map[string]string{"override_template_version_id": *cmd.OverrideTemplateVersionID})
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, fmt.Errorf("controlled_documents: marshal template override payload: %w", err)
			}
			events = append(events, taxonomydomain.GovernanceEvent{
				TenantID:     cmd.TenantID,
				EventType:    "template.override",
				ActorUserID:  cmd.ActorUserID,
				ResourceType: "controlled_document",
				ResourceID:   code,
				Reason:       strings.TrimSpace(*cmd.OverrideTemplateReason),
				PayloadJSON:  payload,
			})
		}

		now := s.now().UTC()
		visibility, err := controlleddocumentsdomain.NewVisibility(
			cmd.VisibilityScope,
			cmd.VisibilityAreaCodes,
			cmd.VisibilityUserIDs,
			cmd.ProcessAreaCode,
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: build visibility: %w", err)
		}
		doc, err = controlleddocumentsdomain.NewControlledDocument(controlleddocumentsdomain.ControlledDocument{
			TenantID:                  cmd.TenantID,
			ProfileCode:               cmd.ProfileCode,
			ProcessAreaCode:           cmd.ProcessAreaCode,
			DepartmentCode:            cmd.DepartmentCode,
			Code:                      code,
			SequenceNum:               sequence,
			Title:                     cmd.Title,
			OwnerUserID:               cmd.OwnerUserID,
			OverrideTemplateVersionID: overrideID,
			Visibility:                visibility,
			Status:                    controlleddocumentsdomain.CDStatusActive,
			CreatedAt:                 now,
			UpdatedAt:                 now,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: build controlled document: %w", err)
		}
		if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
			// ADR 0022 Phase 7: area-scoped tier-2 — symmetric with the auto branch.
			// Closes B2 (F0.2): manual-code creation used to bypass tier-2 because the
			// branch had no tx/identity, so the repo's authz.Require failed-closed on
			// the missing actor_id GUC for every non-system-admin caller.
			// SERVICE IS THE AUTHZ BOUNDARY for CD create: this Require is the single
			// mandatory gate; the repository layer does not re-check (F-CD6).
			if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), cmd.ProcessAreaCode); err != nil {
				return fmt.Errorf("controlled_documents: authz check manual-code create: %w", err)
			}
			return s.docs.CreateTx(ctx, tx, doc)
		}); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: create controlled document (manual): %w", err)
		}
	} else {
		// Pre-flight OFF-TX: validate the template artifact and resolve the
		// effective template version id BEFORE opening the atomic tx. Both touch
		// an authz-recording taxonomy read (GetByCode); running them inside the tx
		// — which holds the audit hash-chain advisory lock once authz.Require
		// records the system_admin bypass — self-deadlocks. Pre-resolving keeps the
		// tx free of off-tx authz reads.
		if err := s.ensureTemplateArtifact(ctx, cmd); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: ensure template artifact: %w", err)
		}
		// ARC-05 hardening: ensureTemplateArtifact above already fails closed via
		// ErrTemplateArtifactInvariantUnconfigured when s.docInit is nil, so this
		// guard is unreachable on the current call order. It is kept explicit
		// (mirrors the :482 and :764 guards on the other two docInit call paths)
		// so a future reorder of these two calls fails with a named error instead
		// of a nil-pointer panic — the wiring-layer construction-order contract
		// (WithDocumentInitializer, see main.go) stays enforced at every call site,
		// not just this one incidentally.
		if s.docInit == nil {
			return nil, ErrTemplateArtifactInvariantUnconfigured
		}
		resolvedTemplateVersionID, err := s.docInit.ResolveTemplateVersionID(ctx, cmd.TenantID, cmd.ProfileCode, cmd.TemplateVersionID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("controlled_documents: resolve initial template version: %w", err)
		}
		var dictionaryValues map[string]string
		if s.docInit != nil {
			dictionaryValues, err = s.docInit.ResolveDictionaryValues(ctx, cmd.TenantID, resolvedTemplateVersionID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return nil, fmt.Errorf("controlled_documents: resolve dictionary values: %w", err)
			}
		}
		// Auto path: sequence allocation, authz, and persistence run atomically.
		if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
			// ADR 0022 Phase 7: area-scoped tier-2 — a CD is created INTO a process
			// area, so authorize against that area (least-privilege; system_admin
			// still bypasses). cmd.ProcessAreaCode validated active above.
			// SERVICE IS THE AUTHZ BOUNDARY for CD create: this Require is the single
			// mandatory gate; the repository layer does not re-check (F-CD6).
			if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), cmd.ProcessAreaCode); err != nil {
				return fmt.Errorf("controlled_documents: authz check sequence allocation: %w", err)
			}

			// Override-template validation (auto path).
			if cmd.OverrideTemplateVersionID != nil {
				if !isReasonValid(cmd.OverrideTemplateReason) {
					return controlleddocumentsdomain.ErrOverrideReasonRequired
				}
				status, profileCode, err := s.tplCheck.GetTemplateVersionState(ctx, cmd.TenantID, *cmd.OverrideTemplateVersionID)
				if err != nil {
					return fmt.Errorf("controlled_documents: get override template version state: %w", err)
				}
				_, err = controlleddocumentsdomain.Resolve(controlleddocumentsdomain.TemplateResolutionInput{
					ProfileCode: cmd.ProfileCode,
					OverrideTemplate: &controlleddocumentsdomain.TemplateVersionCandidate{
						ID:          *cmd.OverrideTemplateVersionID,
						ProfileCode: profileCode,
						Status:      status,
					},
				})
				if err != nil {
					return fmt.Errorf("controlled_documents: resolve template version: %w", err)
				}
				overrideID = cmd.OverrideTemplateVersionID
			}

			next, err := s.seq.NextAndIncrement(ctx, tx, cmd.TenantID, cmd.ProfileCode, cmd.ProcessAreaCode)
			if err != nil {
				return fmt.Errorf("controlled_documents: allocate sequence: %w", err)
			}
			autoCode := controlleddocumentsdomain.AutoCode(cmd.ProfileCode, cmd.ProcessAreaCode, next)
			taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, autoCode)
			if err != nil {
				return fmt.Errorf("controlled_documents: check auto code availability: %w", err)
			}
			if taken {
				return controlleddocumentsdomain.ErrCDCodeTaken
			}
			code = autoCode
			n := next
			sequence = &n

			if overrideID != nil {
				payload, err := json.Marshal(map[string]string{"override_template_version_id": *cmd.OverrideTemplateVersionID})
				if err != nil {
					return fmt.Errorf("controlled_documents: marshal template override payload: %w", err)
				}
				events = append(events, taxonomydomain.GovernanceEvent{
					TenantID:     cmd.TenantID,
					EventType:    "template.override",
					ActorUserID:  cmd.ActorUserID,
					ResourceType: "controlled_document",
					ResourceID:   code,
					Reason:       strings.TrimSpace(*cmd.OverrideTemplateReason),
					PayloadJSON:  payload,
				})
			}

			now := s.now().UTC()
			visibility, err := controlleddocumentsdomain.NewVisibility(
				cmd.VisibilityScope,
				cmd.VisibilityAreaCodes,
				cmd.VisibilityUserIDs,
				cmd.ProcessAreaCode,
			)
			if err != nil {
				return fmt.Errorf("controlled_documents: build visibility: %w", err)
			}
			built, err := controlleddocumentsdomain.NewControlledDocument(controlleddocumentsdomain.ControlledDocument{
				TenantID:                  cmd.TenantID,
				ProfileCode:               cmd.ProfileCode,
				ProcessAreaCode:           cmd.ProcessAreaCode,
				DepartmentCode:            cmd.DepartmentCode,
				Code:                      code,
				SequenceNum:               sequence,
				Title:                     cmd.Title,
				OwnerUserID:               cmd.OwnerUserID,
				OverrideTemplateVersionID: overrideID,
				Visibility:                visibility,
				Status:                    controlleddocumentsdomain.CDStatusActive,
				CreatedAt:                 now,
				UpdatedAt:                 now,
			})
			if err != nil {
				return fmt.Errorf("controlled_documents: build controlled document: %w", err)
			}
			doc = built

			if err := s.docs.CreateTx(ctx, tx, doc); err != nil {
				return fmt.Errorf("controlled_documents: create controlled document in tx: %w", err)
			}
			if s.docInit != nil {
				cloneReq, err := controlleddocumentsdomain.NewCloneTemplateRequest(&resolvedTemplateVersionID, cmd.DocumentName, cmd.FormData)
				if err != nil {
					return fmt.Errorf("controlled_documents: build clone template request: %w", err)
				}
				cloneReq = cloneReq.WithDictionaryValues(dictionaryValues)
				ref, err := s.docInit.CloneTemplate(ctx, tx, doc, cloneReq)
				if err != nil {
					return fmt.Errorf("controlled_documents: clone template for initial revision: %w", err)
				}
				docRef = ref
			}
			return nil
		}); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
	}

	// Governance events are best-effort; document creation is already committed.
	// Multi-leg Create has no single outer tx (no-db path + potential tx branch above),
	// and govLogger.Log issues an authz-recording read — which under H-PRE-1 must
	// never run inside a tx still holding the audit hash-chain advisory lock (see
	// the seedTxIdentity note at ~:320 / ~:734). Folding the event into the create
	// tx would therefore risk the advisory-lock deadlock, so post-commit best-effort
	// logging is accepted here by design (audit F-CD8 / item 2.11).
	for _, event := range events {
		if err := s.govLogger.Log(ctx, event); err != nil { //cilint:allow-post-commit-audit
			slog.WarnContext(ctx, "controlled documents governance event logging failed", "tenant_id", event.TenantID, "actor_user_id", event.ActorUserID, "event_type", event.EventType, "resource_id", event.ResourceID, "error", err)
		}
	}

	return &CreateResult{ControlledDocument: doc, DocumentRef: docRef}, nil
}

func (s *ControlledDocumentService) ensureTemplateArtifact(ctx context.Context, cmd CreateControlledDocumentCmd) error {
	if s.docInit == nil {
		return ErrTemplateArtifactInvariantUnconfigured
	}
	storageKey, err := s.docInit.ResolveTemplateStorageKey(ctx, cmd.TenantID, cmd.ProfileCode, cmd.TemplateVersionID)
	if err != nil {
		return fmt.Errorf("resolve template artifact: %w", err)
	}
	if strings.TrimSpace(storageKey) == "" {
		return ErrTemplateArtifactMissing
	}
	exists, err := s.docInit.Exists(ctx, storageKey)
	if err != nil {
		return fmt.Errorf("check template artifact: %w", err)
	}
	if !exists {
		return ErrTemplateArtifactMissing
	}
	return nil
}

// PreviewCode returns the next auto-allocated CD code for (profile, area)
// without consuming the sequence. Used by the wizard's preview endpoint.
func (s *ControlledDocumentService) PreviewCode(ctx context.Context, tenantID, profileCode, areaCode string) (string, error) {
	next, err := s.seq.Peek(ctx, tenantID, profileCode, areaCode)
	if err != nil {
		return "", fmt.Errorf("controlled_documents: peek sequence for preview code: %w", err)
	}
	return controlleddocumentsdomain.AutoCode(profileCode, areaCode, next), nil
}

// PeekSeq returns the next sequence number that NextAndIncrement would
// allocate for (profile, area). Used by handlers that need the raw integer.
func (s *ControlledDocumentService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	if err := s.validateSequenceSeries(ctx, tenantID, profileCode, areaCode); err != nil {
		return 0, fmt.Errorf("controlled_documents: validate sequence series: %w", err)
	}
	if _, ok := authn.UserIDFromContext(ctx); !ok {
		return 0, ErrActorMissing
	}

	var n int
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		// ADR 0022 Phase 7: preview-code allocation is part of CD create — authorize
		// against the target area, not tenant-wide.
		if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), areaCode); err != nil {
			return fmt.Errorf("controlled_documents: authz check preview code: %w", err)
		}

		v, err := s.seq.Peek(ctx, tenantID, profileCode, areaCode)
		if err != nil {
			return fmt.Errorf("controlled_documents: peek sequence for preview code: %w", err)
		}
		n = v
		return nil
	}); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *ControlledDocumentService) validateSequenceSeries(ctx context.Context, tenantID, profileCode, areaCode string) error {
	profile, err := s.profiles.GetByCode(ctx, tenantID, profileCode)
	if err != nil {
		return fmt.Errorf("controlled_documents: get profile for sequence validation: %w", err)
	}
	if !profile.IsActive() {
		return taxonomydomain.ErrProfileArchived
	}
	area, err := s.areas.GetByCode(ctx, tenantID, areaCode)
	if err != nil {
		return fmt.Errorf("controlled_documents: get process area for sequence validation: %w", err)
	}
	if !area.IsActive() {
		return taxonomydomain.ErrAreaArchived
	}
	return nil
}

// Obsolete transitions the controlled document to CDStatusObsolete.
// Requires the document to currently be active; see changeStatus.
func (s *ControlledDocumentService) Obsolete(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, controlleddocumentsdomain.CDStatusObsolete, string(iamdomain.CapControlledDocumentObsolete))
}

// Supersede transitions the controlled document to CDStatusSuperseded.
// Requires the document to currently be active; see changeStatus.
func (s *ControlledDocumentService) Supersede(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, controlleddocumentsdomain.CDStatusSuperseded, string(iamdomain.CapControlledDocumentSupersede))
}

// List returns controlled documents matching filter, scoped to what the
// caller in ctx is authorized to read (filter.ActorUserID is set from
// context, overriding any caller-supplied value).
func (s *ControlledDocumentService) List(ctx context.Context, tenantID string, filter CDFilter) ([]ControlledDocument, bool, error) {
	actorUserID, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return nil, false, ErrActorMissing
	}
	filter.ActorUserID = &actorUserID
	docs, hasMore, err := s.docs.List(ctx, tenantID, filter)
	if err != nil {
		return nil, false, fmt.Errorf("controlled_documents: list controlled documents: %w", err)
	}
	return docs, hasMore, nil
}

// Get returns the controlled document by id after verifying the caller
// in ctx can read it (CanRead); returns ErrCDNotFound otherwise, so
// existence is not leaked to unauthorized callers.
func (s *ControlledDocumentService) Get(ctx context.Context, tenantID, id string) (*ControlledDocument, error) {
	actorUserID, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrActorMissing
	}
	canRead, err := s.docs.CanRead(ctx, tenantID, id, actorUserID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: check read access: %w", err)
	}
	if !canRead {
		return nil, controlleddocumentsdomain.ErrCDNotFound
	}
	doc, err := s.docs.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: get controlled document: %w", err)
	}
	return doc, nil
}

// GetActiveInstance returns the active document instance for the controlled
// document identified by id, after verifying that the caller can read it.
// Returns ErrNoActiveInstance when the caller cannot read the document (restored contract).
// Returns nil, nil when the document is readable but has no active/published instance.
//
// SEC-03 (T-006): this is a sensitive read — it returns content hashes,
// approval state, and published-revision IDs. Tier-1 (permissions.go) already
// requires document.view for the GET route; this adds the tier-2 in-tx
// capability check so the same read-policy floor documents/approval's
// ReadService.LoadInstance and documents/application ViewService.GetViewURL
// apply (ADR 0022: reason in capabilities, never roles — document.view is the
// tenant-grade *.view read capability, so callers pass the "tenant" sentinel
// to intentionally disable the area filter, exactly as those siblings do).
// The visibility CanRead EXISTS check above stays first — it narrows by
// document visibility_scope/grants; authz.Require then gates by capability.
// Neither check holds a lock, so this does not trip H-PRE-1 (authz-recording
// read inside a lock-holding tx).
func (s *ControlledDocumentService) GetActiveInstance(ctx context.Context, tenantID, id string) (*controlleddocumentsdomain.ActiveDocumentInstance, error) {
	actorUserID, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrActorMissing
	}
	canRead, err := s.docs.CanRead(ctx, tenantID, id, actorUserID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: check read access: %w", err)
	}
	if !canRead {
		return nil, controlleddocumentsdomain.ErrNoActiveInstance
	}

	ctx = authz.WithCapCache(ctx)
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		// document.view is tenant-grade (iam/domain/capability_scope.go) — pass the
		// "tenant" sentinel so the area filter is intentionally OFF, mirroring
		// documents/approval/application/read_service.go LoadInstance and
		// documents/application/view_service.go GetViewURL.
		if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant"); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	inst, err := s.docs.GetActiveInstance(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: get active instance: %w", err)
	}
	return inst, nil
}

func (s *ControlledDocumentService) changeStatus(ctx context.Context, tenantID, controlledDocumentID string, next controlleddocumentsdomain.CDStatus, cap string) error {
	actorUserID, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return ErrActorMissing
	}
	canRead, err := s.docs.CanRead(ctx, tenantID, controlledDocumentID, actorUserID)
	if err != nil {
		return fmt.Errorf("controlled_documents: check read access before status change: %w", err)
	}
	if !canRead {
		return controlleddocumentsdomain.ErrCDNotFound
	}

	return s.runner.Do(ctx, func(tx *sql.Tx) error {
		// ADR 0022 Phase 7: load the CD's process area (FOR UPDATE — read, not a
		// mutation) BEFORE the tier-2 check so obsolete/supersede are authorized
		// against the CD's real area, not tenant-wide. authz.Require still precedes
		// the UpdateStatusTx mutation, so the tripwire pairing holds.
		var (
			currentStatus controlleddocumentsdomain.CDStatus
			areaCode      string
		)
		if err := tx.QueryRowContext(ctx, `
SELECT status, process_area_code
  FROM controlled_documents
 WHERE tenant_id = $1
   AND id = $2
 FOR UPDATE`, tenantID, controlledDocumentID).Scan(&currentStatus, &areaCode); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return controlleddocumentsdomain.ErrCDNotFound
			}
			return fmt.Errorf("controlled_documents: lock controlled document for status change: %w", err)
		}

		if err := authz.Require(ctx, tx, cap, areaCode); err != nil {
			return fmt.Errorf("controlled_documents: authz check changeStatus: %w", err)
		}

		if currentStatus != controlleddocumentsdomain.CDStatusActive {
			return controlleddocumentsdomain.ErrCDNotActive
		}

		if err := s.docs.UpdateStatusTx(ctx, tx, tenantID, controlledDocumentID, next, s.now().UTC()); err != nil {
			return fmt.Errorf("controlled_documents: update controlled document status: %w", err)
		}

		eventType := "controlled_documents.cd.obsoleted"
		if next == controlleddocumentsdomain.CDStatusSuperseded {
			eventType = "controlled_documents.cd.superseded"
		}
		payload, err := json.Marshal(map[string]string{"status": string(next)})
		if err != nil {
			return fmt.Errorf("controlled_documents: marshal changeStatus payload: %w", err)
		}
		// Reuse the actor validated at the top of changeStatus (:625). The
		// middleware seeds authn (iam) and authdomain context keys atomically
		// from the same currentUser.UserID, so this is the same actor — no need
		// to re-extract via the unguarded authdomain path (which defaulted to "").
		if err := s.govLogger.LogTx(ctx, tx, taxonomydomain.GovernanceEvent{
			TenantID:     tenantID,
			EventType:    taxonomydomain.GovernanceEventType(eventType),
			ActorUserID:  actorUserID,
			ResourceType: "controlled_document",
			ResourceID:   controlledDocumentID,
			PayloadJSON:  payload,
		}); err != nil {
			return fmt.Errorf("controlled_documents: governance event for status change: %w", err)
		}
		return nil
	})
}

// CreateRevisionCmd is the input to CreateRevision: TemplateVersionID is
// optional (nil resolves to the CD's profile default / existing override).
type CreateRevisionCmd struct {
	TenantID          string
	CDID              string
	Name              string
	FormData          map[string]any
	TemplateVersionID *string
}

// CreateRevision creates a new document revision for an existing controlled
// document. It requires a DocumentInitializer to be wired (see WithDocumentInitializer).
func (s *ControlledDocumentService) CreateRevision(ctx context.Context, cmd CreateRevisionCmd) (*controlleddocumentsdomain.DocumentRef, error) {
	actorUserID, ok := authn.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrActorMissing
	}
	canRead, err := s.docs.CanRead(ctx, cmd.TenantID, cmd.CDID, actorUserID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: check read access before revision create: %w", err)
	}
	if !canRead {
		return nil, controlleddocumentsdomain.ErrCDNotFound
	}

	cd, err := s.docs.GetByID(ctx, cmd.TenantID, cmd.CDID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: get controlled document for revision: %w", err)
	}
	if !cd.IsActive() {
		return nil, controlleddocumentsdomain.ErrCDNotActive
	}
	if s.docInit == nil {
		return nil, errors.New("controlled_documents: document initializer not configured")
	}

	// Pre-resolve the effective template version id OFF-TX so the in-tx clone does
	// not issue an authz-recording taxonomy read while holding the audit advisory
	// lock (self-deadlock). See ResolveTemplateVersionID on the port.
	resolvedTemplateVersionID, err := s.docInit.ResolveTemplateVersionID(ctx, cmd.TenantID, cd.ProfileCode, cmd.TemplateVersionID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: resolve revision template version: %w", err)
	}

	dictionaryValues, err := s.docInit.ResolveDictionaryValues(ctx, cmd.TenantID, resolvedTemplateVersionID)
	if err != nil {
		return nil, fmt.Errorf("controlled_documents: resolve revision dictionary values: %w", err)
	}

	var ref *controlleddocumentsdomain.DocumentRef
	if err := s.runner.Do(ctx, func(tx *sql.Tx) error {
		// ADR 0022 Phase 7: a revision is created within the CD's own process area —
		// authorize against cd.ProcessAreaCode (loaded above), not tenant-wide.
		if err := authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), cd.ProcessAreaCode); err != nil {
			return fmt.Errorf("controlled_documents: authz check create revision: %w", err)
		}

		cloneReq, err := controlleddocumentsdomain.NewCloneTemplateRequest(&resolvedTemplateVersionID, cmd.Name, cmd.FormData)
		if err != nil {
			return fmt.Errorf("controlled_documents: build revision clone template request: %w", err)
		}
		cloneReq = cloneReq.WithDictionaryValues(dictionaryValues)
		r, err := s.docInit.CloneTemplate(ctx, tx, cd, cloneReq)
		if err != nil {
			return fmt.Errorf("controlled_documents: clone template for revision: %w", mapCreateRevisionError(err))
		}
		ref = r
		return nil
	}); err != nil {
		return nil, err
	}
	return ref, nil
}

func mapCreateRevisionError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_documents_cd_active" {
		return controlleddocumentsdomain.ErrActiveRevisionExists
	}
	return fmt.Errorf("create controlled document revision: %w", err)
}

func isReasonValid(reason *string) bool {
	if reason == nil {
		return false
	}
	return len(strings.TrimSpace(*reason)) >= 10
}
