package controlleddocuments

import (
	"database/sql"
	"log/slog"
	"net/http"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/controlleddocuments/application"
	dhttp "metaldocs/internal/modules/controlleddocuments/delivery/http"
	"metaldocs/internal/modules/controlleddocuments/infrastructure"
	taxonomyapp "metaldocs/internal/modules/taxonomy/application"
)

type Module struct {
	Handler *dhttp.Handler
	svc     *application.ControlledDocumentService
}

type Dependencies struct {
	DB          *sql.DB
	Logger      *slog.Logger
	AuditWriter auditdomain.Writer
}

func New(deps Dependencies) *Module {
	if deps.AuditWriter == nil {
		panic("controlled_documents: AuditWriter is required (nil provided)")
	}
	repo := infrastructure.NewPostgresControlledDocumentRepository(deps.DB)
	seq := infrastructure.NewPostgresSequenceAllocator(deps.DB)
	tplCheck := infrastructure.NewPostgresTemplateVersionChecker(deps.DB)
	profiles := infrastructure.NewTaxonomyProfileReader(deps.DB)
	areas := infrastructure.NewTaxonomyAreaReader(deps.DB)
	govLogger := taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter)
	svc := application.NewControlledDocumentService(deps.DB, repo, seq, tplCheck, profiles, areas, govLogger, nil)
	h := dhttp.NewHandler(svc, deps.DB)
	if svc == nil {
		panic("controlled_documents: service construction returned nil")
	}
	if h == nil {
		panic("controlled_documents: handler construction returned nil")
	}
	return &Module{Handler: h, svc: svc}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}

func (m *Module) Service() *application.ControlledDocumentService { return m.svc }
