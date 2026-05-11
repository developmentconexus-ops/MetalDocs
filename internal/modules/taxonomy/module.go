package taxonomy

import (
	"database/sql"
	"net/http"

	"metaldocs/internal/modules/taxonomy/application"
	thttp "metaldocs/internal/modules/taxonomy/delivery/http"
	"metaldocs/internal/modules/taxonomy/infrastructure"
)

type Module struct {
	Handler *thttp.Handler
}

type Dependencies struct {
	DB         *sql.DB
	TplChecker application.TemplateVersionChecker
}

func New(deps Dependencies) *Module {
	profileRepo := infrastructure.NewProfileRepository(deps.DB)
	areaRepo := infrastructure.NewAreaRepository(deps.DB)
	govLogger := application.NewDBGovernanceLogger(deps.DB)

	familyRepo := infrastructure.NewFamilyRepository(deps.DB)

	profileService := application.NewProfileService(profileRepo, deps.TplChecker, govLogger)
	areaService := application.NewAreaService(areaRepo, govLogger)
	familyService := application.NewFamilyService(familyRepo, govLogger)
	handler := thttp.NewHandler(profileService, areaService, familyService)

	return &Module{Handler: handler}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.Handler.RegisterRoutes(mux)
}
