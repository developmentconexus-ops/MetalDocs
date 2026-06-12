package taxonomy

import (
	"database/sql"
	"net/http"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/taxonomy/application"
	thttp "metaldocs/internal/modules/taxonomy/delivery/http"
	"metaldocs/internal/modules/taxonomy/infrastructure"
)

type Module struct {
	Handler *thttp.Handler
}

type Dependencies struct {
	DB          *sql.DB
	TplChecker  application.TemplateVersionChecker
	AuditWriter auditdomain.Writer
}

func New(deps Dependencies) *Module {
	profileRepo := infrastructure.NewProfileRepository(deps.DB)
	areaRepo := infrastructure.NewAreaRepository(deps.DB)
	tplChecker := deps.TplChecker
	if tplChecker == nil {
		tplChecker = infrastructure.NewTemplateVersionChecker(deps.DB)
	}
	if deps.AuditWriter == nil {
		panic("taxonomy.module: AuditWriter is required; nil fallback to DBGovernanceLogger was removed (FE / Wave 2.12)")
	}
	logger := application.NewAuditGovernanceAdapter(deps.AuditWriter)

	familyRepo := infrastructure.NewFamilyRepository(deps.DB)

	profileService := application.NewProfileService(profileRepo, tplChecker, logger)
	areaService := application.NewAreaService(areaRepo, logger)
	familyService := application.NewFamilyService(familyRepo, logger)
	handler := thttp.NewHandler(profileService, areaService, familyService)

	return &Module{Handler: handler}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}
