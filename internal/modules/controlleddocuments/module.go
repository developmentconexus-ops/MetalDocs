// Package controlleddocuments is the composition root for the
// controlled-documents bounded context: it owns the numbered
// ControlledDocument slot (metaldocs.controlled_documents) that binds a
// (profile, area) pair to a chain of documents-module revisions. New
// wires the repository, sequence allocator, application service, and
// HTTP handler together; RegisterRoutes mounts the resulting HTTP
// surface. The module depends on taxonomy (profile/area validation),
// documents (revision content via DocumentInitializer /
// ActiveInstanceReader), and templates (template version state) purely
// through published ports — never their repositories or SQL.
package controlleddocuments

import (
	"database/sql"
	"log/slog"
	"net/http"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/controlleddocuments/application"
	dhttp "metaldocs/internal/modules/controlleddocuments/delivery/http"
	"metaldocs/internal/modules/controlleddocuments/infrastructure"
	documentsdomain "metaldocs/internal/modules/documents/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	platformdb "metaldocs/internal/platform/db"
)

// Module is the assembled controlled-documents module: its HTTP Handler
// plus the service and repository closed over it.
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

	// RouteReadinessReader answers "does this profile have an active approval
	// route?" through the approval-owned port
	// (approval/domain.RouteReadinessReader). It backs the hard creation gate
	// (D2) and the creation-context read model. Must not be nil — a missing
	// reader would make Create fail closed for every request.
	RouteReadinessReader approvaldomain.RouteReadinessReader

	// ProfileLister / AreaLister are the taxonomy catalog-read ports backing the
	// creation-context read model. Satisfy application.ProfileLister /
	// application.AreaLister. Must not be nil.
	ProfileLister application.ProfileLister
	AreaLister    application.AreaLister

	// AreaCapabilityReader answers "in which areas does this actor hold
	// controlled_documents.create?" through the iam-owned port
	// (iam/domain.AreaCapabilityReader), so the creation context narrows areas
	// server-side instead of shipping the full catalog. Must not be nil.
	AreaCapabilityReader iamdomain.AreaCapabilityReader
}

// New builds a Module: the Postgres repository and sequence allocator
// over deps.DB, the application service, and the HTTP handler. It panics
// if ProfileReader, AreaReader, GovernanceLogger, or
// TemplateVersionChecker is nil (fail-loud by design); ActiveInstanceReader
// defaults to documentsdomain.NoopActiveInstanceReader when nil. The
// service is constructed with a nil DocumentInitializer — callers must
// wire it post-construction via Service().WithDocumentInitializer to
// break the controlled-documents<->documents module init cycle.
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
	if deps.RouteReadinessReader == nil {
		panic("controlled_documents: RouteReadinessReader is required (nil provided)")
	}
	if deps.ProfileLister == nil {
		panic("controlled_documents: ProfileLister is required (nil provided)")
	}
	if deps.AreaLister == nil {
		panic("controlled_documents: AreaLister is required (nil provided)")
	}
	if deps.AreaCapabilityReader == nil {
		panic("controlled_documents: AreaCapabilityReader is required (nil provided)")
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
	svc.WithRouteReadinessReader(deps.RouteReadinessReader)
	svc.WithCreationContextReaders(deps.ProfileLister, deps.AreaLister, deps.AreaCapabilityReader)
	h := dhttp.NewHandler(svc, deps.DB)
	if h == nil {
		panic("controlled_documents: handler construction returned nil")
	}
	return &Module{Handler: h, svc: svc, repo: repo}
}

// RegisterRoutes mounts the module's HTTP handler onto mux.
func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}

// Service returns the module's application service so the composition
// root can wire a DocumentInitializer post-construction (see New) or
// call the service directly from another module's wiring code.
func (m *Module) Service() *application.ControlledDocumentService { return m.svc }

// Repo returns the module's controlled-document repository so the composition
// root can wire it into other modules without constructing a second instance.
func (m *Module) Repo() *infrastructure.PostgresControlledDocumentRepository { return m.repo }
