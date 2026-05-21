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

	authdomain "metaldocs/internal/modules/auth/domain"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

type TemplateVersionChecker interface {
	GetTemplateVersionState(ctx context.Context, templateVersionID string) (*string, string, error)
}

type ProfileReader interface {
	GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)
}

type AreaReader interface {
	GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error)
}

type ControlledDocument = controlleddocumentsdomain.ControlledDocument
type CDFilter = controlleddocumentsdomain.CDFilter

type ControlledDocumentService struct {
	db        *sql.DB
	docs      controlleddocumentsdomain.ControlledDocumentRepository
	seq       controlleddocumentsdomain.SequenceAllocator
	tplCheck  TemplateVersionChecker
	profiles  ProfileReader
	areas     AreaReader
	govLogger taxonomydomain.GovernanceLogger
	docInit   controlleddocumentsdomain.DocumentInitializer
	now       func() time.Time
}

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

var ErrTemplateArtifactMissing = errors.New("template artifact missing")
var ErrTemplateArtifactInvariantUnconfigured = errors.New("controlled_documents: template artifact invariant not configured")

func NewControlledDocumentService(
	db *sql.DB,
	docs controlleddocumentsdomain.ControlledDocumentRepository,
	seq controlleddocumentsdomain.SequenceAllocator,
	tplCheck TemplateVersionChecker,
	profiles ProfileReader,
	areas AreaReader,
	govLogger taxonomydomain.GovernanceLogger,
	docInit controlleddocumentsdomain.DocumentInitializer,
) *ControlledDocumentService {
	if govLogger == nil {
		panic("controlled_documents: governance logger must not be nil")
	}
	svc := &ControlledDocumentService{
		db:        db,
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

func (s *ControlledDocumentService) Create(ctx context.Context, cmd CreateControlledDocumentCmd) (*CreateResult, error) {
	profile, err := s.profiles.GetByCode(ctx, cmd.TenantID, cmd.ProfileCode)
	if err != nil {
		return nil, err
	}
	if !profile.IsActive() {
		return nil, taxonomydomain.ErrProfileArchived
	}

	area, err := s.areas.GetByCode(ctx, cmd.TenantID, cmd.ProcessAreaCode)
	if err != nil {
		return nil, err
	}
	if !area.IsActive() {
		return nil, taxonomydomain.ErrAreaArchived
	}

	var (
		code       string
		sequence   *int
		events     []taxonomydomain.GovernanceEvent
		overrideID *string
		createTx   *sql.Tx
	)

	if cmd.ManualCode != nil {
		if !isReasonValid(cmd.ManualCodeReason) {
			return nil, controlleddocumentsdomain.ErrManualCodeReasonRequired
		}
		code = strings.TrimSpace(*cmd.ManualCode)
		taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, controlleddocumentsdomain.ErrCDCodeTaken
		}
		payload, _ := json.Marshal(map[string]string{"code": code})
		events = append(events, taxonomydomain.GovernanceEvent{
			TenantID:     cmd.TenantID,
			EventType:    "numbering.override",
			ActorUserID:  cmd.ActorUserID,
			ResourceType: "controlled_document",
			ResourceID:   code,
			Reason:       strings.TrimSpace(*cmd.ManualCodeReason),
			PayloadJSON:  payload,
		})
	} else {
		if s.db != nil {
			tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
			if err != nil {
				return nil, err
			}
			defer func() {
				if createTx != nil {
					_ = tx.Rollback()
				}
			}()
			createTx = tx
			if err := setAuthzGUC(ctx, createTx, cmd.TenantID, cmd.ActorUserID); err != nil {
				return nil, fmt.Errorf("controlled_documents: set authz context: %w", err)
			}
			if err := authz.Require(ctx, createTx, string(iamdomain.CapControlledDocumentCreate), "tenant"); err != nil {
				return nil, fmt.Errorf("controlled_documents: authz check sequence allocation: %w", err)
			}
		}
	}
	if cmd.OverrideTemplateVersionID != nil {
		if !isReasonValid(cmd.OverrideTemplateReason) {
			return nil, controlleddocumentsdomain.ErrOverrideReasonRequired
		}
		status, profileCode, err := s.tplCheck.GetTemplateVersionState(ctx, *cmd.OverrideTemplateVersionID)
		if err != nil {
			return nil, err
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
			return nil, err
		}
		overrideID = cmd.OverrideTemplateVersionID
	}
	if err := s.ensureTemplateArtifact(ctx, cmd); err != nil {
		return nil, err
	}

	if cmd.ManualCode == nil {
		if createTx != nil {
			next, err := s.seq.NextAndIncrement(ctx, createTx, cmd.TenantID, cmd.ProfileCode, cmd.ProcessAreaCode)
			if err != nil {
				return nil, err
			}
			code = controlleddocumentsdomain.AutoCode(cmd.ProfileCode, cmd.ProcessAreaCode, next)
			sequence = &next
			taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, controlleddocumentsdomain.ErrCDCodeTaken
			}
		} else {
			next, err := s.seq.NextAndIncrement(ctx, nil, cmd.TenantID, cmd.ProfileCode, cmd.ProcessAreaCode)
			if err != nil {
				return nil, err
			}
			code = controlleddocumentsdomain.AutoCode(cmd.ProfileCode, cmd.ProcessAreaCode, next)
			sequence = &next
			taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, controlleddocumentsdomain.ErrCDCodeTaken
			}
		}
	}

	if overrideID != nil {
		payload, _ := json.Marshal(map[string]string{"override_template_version_id": *cmd.OverrideTemplateVersionID})
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
		return nil, err
	}
	doc := &controlleddocumentsdomain.ControlledDocument{
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
	}
	var docRef *controlleddocumentsdomain.DocumentRef
	if createTx != nil {
		if err := s.docs.CreateTx(ctx, createTx, doc); err != nil {
			return nil, err
		}
		if s.docInit != nil {
			ref, err := s.docInit.CloneTemplate(ctx, createTx, doc, controlleddocumentsdomain.CloneTemplateRequest{
				TemplateVersionID: cmd.TemplateVersionID,
				Name:              cmd.DocumentName,
				FormData:          cmd.FormData,
			})
			if err != nil {
				return nil, err
			}
			docRef = ref
		}
		if err := createTx.Commit(); err != nil {
			return nil, err
		}
		createTx = nil
	} else {
		if err := s.docs.Create(ctx, doc); err != nil {
			return nil, err
		}
	}

	// Governance events are best-effort; document creation is already committed.
	for _, event := range events {
		if err := s.govLogger.Log(ctx, event); err != nil {
			slog.Warn("controlled documents governance event logging failed", "event_type", event.EventType, "resource_id", event.ResourceID, "error", err)
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

func setAuthzGUC(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error {
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.tenant_id', $1, true)", tenantID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "SELECT set_config('metaldocs.actor_id', $1, true)", actorID); err != nil {
		return err
	}
	return nil
}

// PreviewCode returns the next auto-allocated CD code for (profile, area)
// without consuming the sequence. Used by the wizard's preview endpoint.
func (s *ControlledDocumentService) PreviewCode(ctx context.Context, tenantID, profileCode, areaCode string) (string, error) {
	next, err := s.seq.Peek(ctx, tenantID, profileCode, areaCode)
	if err != nil {
		return "", err
	}
	return controlleddocumentsdomain.AutoCode(profileCode, areaCode, next), nil
}

// PeekSeq returns the next sequence number that NextAndIncrement would
// allocate for (profile, area). Used by handlers that need the raw integer.
func (s *ControlledDocumentService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	return s.seq.Peek(ctx, tenantID, profileCode, areaCode)
}

func (s *ControlledDocumentService) Obsolete(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, controlleddocumentsdomain.CDStatusObsolete, string(iamdomain.CapControlledDocumentObsolete))
}

func (s *ControlledDocumentService) Supersede(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, controlleddocumentsdomain.CDStatusSuperseded, string(iamdomain.CapControlledDocumentSupersede))
}

func (s *ControlledDocumentService) List(ctx context.Context, tenantID string, filter CDFilter) ([]ControlledDocument, error) {
	if actorUserID := strings.TrimSpace(authn.UserIDFromContext(ctx)); actorUserID != "" {
		filter.ActorUserID = &actorUserID
	}
	return s.docs.List(ctx, tenantID, filter)
}

func (s *ControlledDocumentService) Get(ctx context.Context, tenantID, id string) (*ControlledDocument, error) {
	actorUserID := strings.TrimSpace(authn.UserIDFromContext(ctx))
	canRead, err := s.docs.CanRead(ctx, tenantID, id, actorUserID)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, controlleddocumentsdomain.ErrCDNotFound
	}
	return s.docs.GetByID(ctx, tenantID, id)
}

func (s *ControlledDocumentService) changeStatus(ctx context.Context, tenantID, controlledDocumentID string, next controlleddocumentsdomain.CDStatus, cap string) error {
	doc, err := s.docs.GetByID(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return err
	}
	if !doc.IsActive() {
		return controlleddocumentsdomain.ErrCDNotActive
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin changeStatus tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, cap, "tenant"); err != nil {
		return fmt.Errorf("controlled_documents: authz check changeStatus: %w", err)
	}
	if err := s.docs.UpdateStatusTx(ctx, tx, tenantID, controlledDocumentID, next, s.now().UTC()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	eventType := "controlled_documents.cd.obsoleted"
	if next == controlleddocumentsdomain.CDStatusSuperseded {
		eventType = "controlled_documents.cd.superseded"
	}
	payload, _ := json.Marshal(map[string]string{"status": string(next)})
	actorID := ""
	if u, ok := authdomain.CurrentUserFromContext(ctx); ok {
		actorID = u.UserID
	}
	if err := s.govLogger.Log(ctx, taxonomydomain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    eventType,
		ActorUserID:  actorID,
		ResourceType: "controlled_document",
		ResourceID:   controlledDocumentID,
		PayloadJSON:  payload,
	}); err != nil {
		slog.Warn("controlled documents governance event logging failed", "event_type", eventType, "resource_id", controlledDocumentID, "error", err)
	}
	return nil
}

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
	cd, err := s.docs.GetByID(ctx, cmd.TenantID, cmd.CDID)
	if err != nil {
		return nil, err
	}
	if !cd.IsActive() {
		return nil, controlleddocumentsdomain.ErrCDNotActive
	}
	if s.docInit == nil {
		return nil, errors.New("controlled_documents: document initializer not configured")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	var txErr error
	defer func() {
		if txErr != nil {
			_ = tx.Rollback()
		}
	}()

	actorUserID := strings.TrimSpace(authn.UserIDFromContext(ctx))
	if actorUserID == "" {
		txErr = errors.New("controlled_documents: actor user id missing in context")
		return nil, txErr
	}
	if txErr = setAuthzGUC(ctx, tx, cmd.TenantID, actorUserID); txErr != nil {
		return nil, fmt.Errorf("controlled_documents: set authz guc for create revision: %w", txErr)
	}
	if txErr = authz.Require(ctx, tx, string(iamdomain.CapControlledDocumentCreate), "tenant"); txErr != nil {
		return nil, fmt.Errorf("controlled_documents: authz check create revision: %w", txErr)
	}

	ref, txErr := s.docInit.CloneTemplate(ctx, tx, cd, controlleddocumentsdomain.CloneTemplateRequest{
		TemplateVersionID: cmd.TemplateVersionID,
		Name:              cmd.Name,
		FormData:          cmd.FormData,
	})
	if txErr != nil {
		return nil, mapCreateRevisionError(txErr)
	}
	txErr = tx.Commit()
	return ref, txErr
}

func mapCreateRevisionError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_documents_cd_active" {
		return controlleddocumentsdomain.ErrActiveRevisionExists
	}
	return err
}

func isReasonValid(reason *string) bool {
	if reason == nil {
		return false
	}
	return len(strings.TrimSpace(*reason)) >= 10
}
