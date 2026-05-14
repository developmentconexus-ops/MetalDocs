package registry

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/registry/application"
	dhttp "metaldocs/internal/modules/registry/delivery/http"
	"metaldocs/internal/modules/registry/infrastructure"
	taxonomyapp "metaldocs/internal/modules/taxonomy/application"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

type Module struct {
	Handler *dhttp.Handler
	svc     *application.RegistryService
}

type Dependencies struct {
	DB          *sql.DB
	Logger      *slog.Logger
	AuditWriter auditdomain.Writer
}

func New(deps Dependencies) *Module {
	repo := infrastructure.NewPostgresControlledDocumentRepository(deps.DB)
	seq := infrastructure.NewPostgresSequenceAllocator(deps.DB)
	tplCheck := infrastructure.NewPostgresTemplateVersionChecker(deps.DB)
	profiles := infrastructure.NewTaxonomyProfileReader(deps.DB)
	areas := infrastructure.NewTaxonomyAreaReader(deps.DB)
	var govLogger taxonomydomain.GovernanceLogger
	if deps.AuditWriter != nil {
		govLogger = taxonomyapp.NewAuditGovernanceAdapter(deps.AuditWriter)
	} else {
		govLogger = taxonomyapp.NewDBGovernanceLogger(deps.DB)
	}
	svc := application.NewRegistryService(deps.DB, repo, seq, tplCheck, profiles, areas, govLogger, nil)
	h := dhttp.NewHandler(svc, deps.DB)
	return &Module{Handler: h, svc: svc}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}

// RunLegacyMaintenance executes legacy-only registry maintenance for older
// databases. On fresh curated baselines it is expected to no-op.
func (m *Module) RunLegacyMaintenance(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	return application.BackfillLegacyDocuments(ctx, db, logger)
}

func (m *Module) RunStartupMigrations(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	return m.RunLegacyMaintenance(ctx, db, logger)
}

func (m *Module) Service() *application.RegistryService { return m.svc }
