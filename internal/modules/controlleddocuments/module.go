package controlleddocuments

import (
	"database/sql"
	"log/slog"
	"net/http"

	"metaldocs/internal/modules/controlleddocuments/application"
	dhttp "metaldocs/internal/modules/controlleddocuments/delivery/http"
	"metaldocs/internal/modules/controlleddocuments/infrastructure"
	documentsdomain "metaldocs/internal/modules/documents/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	platformdb "metaldocs/internal/platform/db"
)

type Module struct {
	Handler *dhttp.Handler
	svc     *application.ControlledDocumentService
	repo    *infrastructure.PostgresControlledDocumentRepository
}

// Dependencies carries all collaborator ports for the controlleddocuments module.
// Every sibling-module collaborator is expressed as an interface (never a
// concrete sibling infra type) — the composition root in main.go constructs the
// concretes and passes them here.
//
// Ports sourced from the sibling module's published domain/application interface:
//   - ActiveInstanceReader: documents/domain.ActiveInstanceReader (published)
//   - ProfileReader:        consumer-defined interface (application.ProfileReader)
//   - AreaReader:           consumer-defined interface (application.AreaReader)
//   - GovernanceLogger:     taxonomy/domain.GovernanceLogger (published)
//   - TemplateVersionChecker: consumer-defined interface (application.TemplateVersionChecker)
type Dependencies struct {
	DB     *sql.DB
	Logger *slog.Logger

	// ActiveInstanceReader is the documents-owned read-port for the active/
	// published document projection (ADR-0039 D3(b); M2/F2.2).
	// When nil, documentsdomain.NoopActiveInstanceReader{} is used.
	ActiveInstanceReader documentsdomain.ActiveInstanceReader

	// ProfileReader reads taxonomy document profiles without crossing the module
	// boundary. Satisfies application.ProfileReader. Must not be nil.
	ProfileReader application.ProfileReader

	// AreaReader reads taxonomy process areas without crossing the module
	// boundary. Satisfies application.AreaReader. Must not be nil.
	AreaReader application.AreaReader

	// GovernanceLogger writes governance audit events (taxonomy-owned port).
	// Must not be nil.
	GovernanceLogger taxonomydomain.GovernanceLogger

	// TemplateVersionChecker reads template-version state through the
	// templates-owned port (M4 F4.2). Must not be nil.
	TemplateVersionChecker application.TemplateVersionChecker
}

func New(deps Dependencies) *Module {
	if deps.ProfileReader == nil {
		panic("controlled_documents: ProfileReader is required (nil provided)")
	}
	if deps.AreaReader == nil {
		panic("controlled_documents: AreaReader is required (nil provided)")
	}
	if deps.GovernanceLogger == nil {
		panic("controlled_documents: GovernanceLogger is required (nil provided)")
	}
	if deps.TemplateVersionChecker == nil {
		panic("controlled_documents: TemplateVersionChecker is required (nil provided)")
	}

	// documents-owned active-instance read-port: use the injected port or the
	// published Noop default (mirrors documents.New Noop pattern).
	activeInstance := deps.ActiveInstanceReader
	if activeInstance == nil {
		activeInstance = documentsdomain.NoopActiveInstanceReader{}
	}

	repo := infrastructure.NewPostgresControlledDocumentRepository(deps.DB, activeInstance)
	seq := infrastructure.NewPostgresSequenceAllocator(deps.DB)
	svc := application.NewControlledDocumentService(
		platformdb.NewTxRunner(deps.DB),
		repo,
		seq,
		deps.TemplateVersionChecker,
		deps.ProfileReader,
		deps.AreaReader,
		deps.GovernanceLogger,
		nil,
	)
	if svc == nil {
		panic("controlled_documents: service construction returned nil")
	}
	h := dhttp.NewHandler(svc, deps.DB)
	if h == nil {
		panic("controlled_documents: handler construction returned nil")
	}
	return &Module{Handler: h, svc: svc, repo: repo}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}

func (m *Module) Service() *application.ControlledDocumentService { return m.svc }

// Repo returns the module's controlled-document repository so the composition
// root can wire it into other modules without constructing a second instance.
func (m *Module) Repo() *infrastructure.PostgresControlledDocumentRepository { return m.repo }
