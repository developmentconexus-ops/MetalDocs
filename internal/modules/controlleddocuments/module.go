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
	taxonomyinfra "metaldocs/internal/modules/taxonomy/infrastructure"
	platformdb "metaldocs/internal/platform/db"
)

type Module struct {
	Handler *dhttp.Handler
	svc     *application.ControlledDocumentService
	repo    *infrastructure.PostgresControlledDocumentRepository
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
	// Read profiles/areas through the canonical taxonomy repositories so the
	// authz GUC and CapTaxonomyView check run on every lookup (H-1b). Every role
	// that can create a controlled document already holds taxonomy.view, so the
	// enforced capability is one the request actor already has.
	profiles := infrastructure.NewTaxonomyProfileReader(taxonomyinfra.NewProfileRepository(deps.DB))
	areas := infrastructure.NewTaxonomyAreaReader(taxonomyinfra.NewAreaRepository(deps.DB))
	govLogger := taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter)
	svc := application.NewControlledDocumentService(platformdb.NewTxRunner(deps.DB), repo, seq, tplCheck, profiles, areas, govLogger, nil)
	h := dhttp.NewHandler(svc, deps.DB)
	if svc == nil {
		panic("controlled_documents: service construction returned nil")
	}
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
