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

	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	registrydomain "metaldocs/internal/modules/registry/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

type TemplateVersionChecker interface {
	GetTemplateVersionState(ctx context.Context, templateVersionID string) (*string, string, error)
}

type TemplateArtifactChecker interface {
	Exists(ctx context.Context, storageKey string) (bool, error)
}

type templateArtifactResolver interface {
	ResolveTemplateStorageKey(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error)
}

type ProfileReader interface {
	GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.DocumentProfile, error)
}

type AreaReader interface {
	GetByCode(ctx context.Context, tenantID, code string) (*taxonomydomain.ProcessArea, error)
}

type ControlledDocument = registrydomain.ControlledDocument
type CDFilter = registrydomain.CDFilter

type RegistryService struct {
	db                       *sql.DB
	docs                     registrydomain.ControlledDocumentRepository
	seq                      registrydomain.SequenceAllocator
	tplCheck                 TemplateVersionChecker
	profiles                 ProfileReader
	areas                    AreaReader
	govLogger                taxonomydomain.GovernanceLogger
	docInit                  registrydomain.DocumentInitializer
	templateArtifactResolver templateArtifactResolver
	templateArtifactChecker  TemplateArtifactChecker
	now                      func() time.Time
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
// plus the optional DocumentRef returned by DocumentInitializer (nil when no
// initializer is wired).
type CreateResult struct {
	ControlledDocument *registrydomain.ControlledDocument
	DocumentRef        *registrydomain.DocumentRef
}

var ErrTemplateArtifactMissing = errors.New("template artifact missing")

func NewRegistryService(
	db *sql.DB,
	docs registrydomain.ControlledDocumentRepository,
	seq registrydomain.SequenceAllocator,
	tplCheck TemplateVersionChecker,
	profiles ProfileReader,
	areas AreaReader,
	govLogger taxonomydomain.GovernanceLogger,
	docInit registrydomain.DocumentInitializer,
) *RegistryService {
	if govLogger == nil {
		panic("registry: governance logger must not be nil")
	}
	return &RegistryService{
		db:        db,
		docs:      docs,
		seq:       seq,
		tplCheck:  tplCheck,
		profiles:  profiles,
		areas:     areas,
		govLogger: govLogger,
		docInit:   docInit,
		now:       time.Now,
	}
}

// WithDocumentInitializer wires the DocumentInitializer adapter post-construction.
// Used by the wiring layer to break the registry<->documents module cycle: the
// registry module is constructed first (because documents needs a
// RegistryDuplicator), then the documents module is built, then the
// initializer adapter is injected back here.
func (s *RegistryService) WithDocumentInitializer(d registrydomain.DocumentInitializer) *RegistryService {
	s.docInit = d
	if resolver, ok := d.(templateArtifactResolver); ok {
		s.templateArtifactResolver = resolver
	}
	if checker, ok := d.(TemplateArtifactChecker); ok {
		s.templateArtifactChecker = checker
	}
	return s
}

func (s *RegistryService) withTemplateArtifactResolver(resolver templateArtifactResolver) *RegistryService {
	s.templateArtifactResolver = resolver
	return s
}

func (s *RegistryService) withTemplateArtifactChecker(checker TemplateArtifactChecker) *RegistryService {
	s.templateArtifactChecker = checker
	return s
}

func (s *RegistryService) Create(ctx context.Context, cmd CreateControlledDocumentCmd) (*CreateResult, error) {
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

	if err := s.ensureTemplateArtifact(ctx, cmd); err != nil {
		return nil, err
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
			return nil, registrydomain.ErrManualCodeReasonRequired
		}
		code = strings.TrimSpace(*cmd.ManualCode)
		taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, registrydomain.ErrCDCodeTaken
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
				return nil, fmt.Errorf("registry: set authz context: %w", err)
			}
			if err := authz.Require(ctx, createTx, string(iamdomain.CapRegistryCreate), "tenant"); err != nil {
				return nil, fmt.Errorf("registry: authz check sequence allocation: %w", err)
			}

			next, err := s.seq.NextAndIncrement(ctx, tx, cmd.TenantID, cmd.ProfileCode, cmd.ProcessAreaCode)
			if err != nil {
				return nil, err
			}
			code = registrydomain.AutoCode(cmd.ProfileCode, cmd.ProcessAreaCode, next)
			sequence = &next
			taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, registrydomain.ErrCDCodeTaken
			}
		} else {
			next, err := s.seq.NextAndIncrement(ctx, nil, cmd.TenantID, cmd.ProfileCode, cmd.ProcessAreaCode)
			if err != nil {
				return nil, err
			}
			code = registrydomain.AutoCode(cmd.ProfileCode, cmd.ProcessAreaCode, next)
			sequence = &next
			taken, err := s.docs.CodeExists(ctx, cmd.TenantID, cmd.ProfileCode, code)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, registrydomain.ErrCDCodeTaken
			}
		}
	}

	if cmd.OverrideTemplateVersionID != nil {
		if !isReasonValid(cmd.OverrideTemplateReason) {
			return nil, registrydomain.ErrOverrideReasonRequired
		}
		status, profileCode, err := s.tplCheck.GetTemplateVersionState(ctx, *cmd.OverrideTemplateVersionID)
		if err != nil {
			return nil, err
		}
		_, err = registrydomain.Resolve(registrydomain.TemplateResolutionInput{
			ProfileCode: cmd.ProfileCode,
			OverrideTemplate: &registrydomain.TemplateVersionCandidate{
				ID:          *cmd.OverrideTemplateVersionID,
				ProfileCode: profileCode,
				Status:      status,
			},
		})
		if err != nil {
			return nil, err
		}
		overrideID = cmd.OverrideTemplateVersionID
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
	visibility, err := registrydomain.NewVisibility(
		cmd.VisibilityScope,
		cmd.VisibilityAreaCodes,
		cmd.VisibilityUserIDs,
		cmd.ProcessAreaCode,
	)
	if err != nil {
		return nil, err
	}
	doc := &registrydomain.ControlledDocument{
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
		Status:                    registrydomain.CDStatusActive,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	var docRef *registrydomain.DocumentRef
	if createTx != nil {
		if err := s.docs.CreateTx(ctx, createTx, doc); err != nil {
			return nil, err
		}
		if s.docInit != nil {
			ref, err := s.docInit.CloneTemplate(ctx, createTx, doc, registrydomain.CloneTemplateRequest{
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
			slog.Warn("registry governance event logging failed", "event_type", event.EventType, "resource_id", event.ResourceID, "error", err)
		}
	}

	return &CreateResult{ControlledDocument: doc, DocumentRef: docRef}, nil
}

func (s *RegistryService) ensureTemplateArtifact(ctx context.Context, cmd CreateControlledDocumentCmd) error {
	if s.templateArtifactResolver == nil || s.templateArtifactChecker == nil {
		return nil
	}
	storageKey, err := s.templateArtifactResolver.ResolveTemplateStorageKey(ctx, cmd.TenantID, cmd.ProfileCode, cmd.TemplateVersionID)
	if err != nil {
		return fmt.Errorf("resolve template artifact: %w", err)
	}
	if strings.TrimSpace(storageKey) == "" {
		return ErrTemplateArtifactMissing
	}
	exists, err := s.templateArtifactChecker.Exists(ctx, storageKey)
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
func (s *RegistryService) PreviewCode(ctx context.Context, tenantID, profileCode, areaCode string) (string, error) {
	next, err := s.seq.Peek(ctx, tenantID, profileCode, areaCode)
	if err != nil {
		return "", err
	}
	return registrydomain.AutoCode(profileCode, areaCode, next), nil
}

// PeekSeq returns the next sequence number that NextAndIncrement would
// allocate for (profile, area). Used by handlers that need the raw integer.
func (s *RegistryService) PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error) {
	return s.seq.Peek(ctx, tenantID, profileCode, areaCode)
}

func (s *RegistryService) Obsolete(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, registrydomain.CDStatusObsolete, string(iamdomain.CapRegistryObsolete))
}

func (s *RegistryService) Supersede(ctx context.Context, tenantID, controlledDocumentID string) error {
	return s.changeStatus(ctx, tenantID, controlledDocumentID, registrydomain.CDStatusSuperseded, string(iamdomain.CapRegistrySupersede))
}

func (s *RegistryService) List(ctx context.Context, tenantID string, filter CDFilter) ([]ControlledDocument, error) {
	if actorUserID := strings.TrimSpace(authn.UserIDFromContext(ctx)); actorUserID != "" {
		filter.ActorUserID = &actorUserID
	}
	return s.docs.List(ctx, tenantID, filter)
}

func (s *RegistryService) Get(ctx context.Context, tenantID, id string) (*ControlledDocument, error) {
	actorUserID := strings.TrimSpace(authn.UserIDFromContext(ctx))
	canRead, err := s.docs.CanRead(ctx, tenantID, id, actorUserID)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, registrydomain.ErrCDNotFound
	}
	return s.docs.GetByID(ctx, tenantID, id)
}

func (s *RegistryService) changeStatus(ctx context.Context, tenantID, controlledDocumentID string, next registrydomain.CDStatus, cap string) error {
	doc, err := s.docs.GetByID(ctx, tenantID, controlledDocumentID)
	if err != nil {
		return err
	}
	if !doc.IsActive() {
		return registrydomain.ErrCDNotActive
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin changeStatus tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := authz.Require(ctx, tx, cap, "tenant"); err != nil {
		return fmt.Errorf("registry: authz check changeStatus: %w", err)
	}
	if err := s.docs.UpdateStatusTx(ctx, tx, tenantID, controlledDocumentID, next, s.now().UTC()); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	eventType := "registry.cd.obsoleted"
	if next == registrydomain.CDStatusSuperseded {
		eventType = "registry.cd.superseded"
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
		slog.Warn("registry governance event logging failed", "event_type", eventType, "resource_id", controlledDocumentID, "error", err)
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
func (s *RegistryService) CreateRevision(ctx context.Context, cmd CreateRevisionCmd) (*registrydomain.DocumentRef, error) {
	cd, err := s.docs.GetByID(ctx, cmd.TenantID, cmd.CDID)
	if err != nil {
		return nil, err
	}
	if !cd.IsActive() {
		return nil, registrydomain.ErrCDNotActive
	}
	if s.docInit == nil {
		return nil, errors.New("registry: document initializer not configured")
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

	ref, txErr := s.docInit.CloneTemplate(ctx, tx, cd, registrydomain.CloneTemplateRequest{
		TemplateVersionID: cmd.TemplateVersionID,
		Name:              cmd.Name,
		FormData:          cmd.FormData,
	})
	if txErr != nil {
		return nil, txErr
	}
	txErr = tx.Commit()
	return ref, txErr
}

func isReasonValid(reason *string) bool {
	if reason == nil {
		return false
	}
	return len(strings.TrimSpace(*reason)) >= 10
}
