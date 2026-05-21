package documents

import (
	"context"
	"database/sql"
	"net/http"

	documentsapi "metaldocs/internal/modules/documents/api"
	"metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	dhttp "metaldocs/internal/modules/documents/delivery/http"
	documentshttp "metaldocs/internal/modules/documents/http"
	"metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/ratelimit"
)

type approvalSubmitter interface {
	SubmitRevisionForReview(ctx context.Context, db *sql.DB, req approvalapp.SubmitRequest) (approvalapp.SubmitResult, error)
}

type Module struct {
	Handler                   *dhttp.Handler
	Service                   *application.Service
	ExportHandler             *dhttp.ExportHandler
	FillInHandler             *documentshttp.FillInHandler
	PlaceholderOptionsHandler *documentshttp.PlaceholderOptionsHandler
	ViewHandler               *documentshttp.ViewHandler
	ReconstructHandler        *documentshttp.ReconstructHandler
	repo                      *repository.Repository
}

type Dependencies struct {
	DB                           *sql.DB
	Docgen                       application.DocgenRenderer
	Presign                      application.Presigner
	TplRead                      application.TemplateReader
	FormVal                      application.FormValidator
	Audit                        application.Audit
	ControlledDocumentReader     application.ControlledDocumentReader
	ControlledDocumentDuplicator application.ControlledDocumentDuplicator
	Caps                         application.CapabilityChecker
	ProfileDefaults              application.ProfileDefaultTemplateReader
	SnapshotReader               application.SnapshotTemplateReader
	SnapshotWriter               application.SnapshotWriter
	ExportPresign                application.ExportPresigner
	ExportDocgen                 application.DocgenPDFClient
	DocgenVer                    string
	GrammarVer                   string
	ReconstructRunner            application.ReconstructionRunner
	SubmitSvc                    approvalSubmitter
	IAMUserOptions               application.IAMUserOptionsReader
}

func New(deps Dependencies) *Module {
	repo := repository.New(deps.DB)
	var svc *application.Service
	if deps.SnapshotReader != nil && deps.SnapshotWriter != nil {
		snapSvc := application.NewSnapshotService(deps.SnapshotReader, deps.SnapshotWriter)
		svc = application.NewServiceWithSnapshot(repo, deps.Docgen, deps.Presign, deps.TplRead, deps.FormVal, deps.Audit, deps.ControlledDocumentReader, deps.Caps, deps.ProfileDefaults, snapSvc)
	} else {
		svc = application.NewService(repo, deps.Docgen, deps.Presign, deps.TplRead, deps.FormVal, deps.Audit, deps.ControlledDocumentReader, deps.Caps, deps.ProfileDefaults)
	}
	svc.WithDB(deps.DB)
	svc.WithControlledDocumentDuplicator(deps.ControlledDocumentDuplicator)
	h := dhttp.NewHandlerWithSubmit(svc, deps.DB, deps.SubmitSvc)

	var exportHandler *dhttp.ExportHandler
	if deps.ExportPresign != nil && deps.ExportDocgen != nil {
		docgenVer := deps.DocgenVer
		if docgenVer == "" {
			docgenVer = "docgen-v2@0.4.0"
		}
		grammarVer := deps.GrammarVer
		if grammarVer == "" {
			grammarVer = "grammar-v1"
		}
		exportSvc := application.NewExportService(repo, deps.ExportPresign, deps.ExportDocgen, deps.Audit, docgenVer, grammarVer)
		exportHandler = dhttp.NewExportHandler(exportSvc)
	}

	fillInRepo := repository.NewFillInRepository(deps.DB)
	fillInSvc := application.NewFillInService(deps.DB, application.NewSnapshotSchemaReader(deps.DB), fillInRepo).
		WithReader(fillInRepo).
		WithTemplateSchemaReader(application.NewTemplateVersionSchemaReader(deps.DB))
	fillInHandler := documentshttp.NewFillInHandler(fillInSvc)
	placeholderOptionsHandler := documentshttp.NewPlaceholderOptionsHandler(
		application.NewSnapshotSchemaReader(deps.DB),
		newPlaceholderOptionsIAMAdapter(deps.IAMUserOptions),
	)

	var viewHandler *documentshttp.ViewHandler
	if deps.Presign != nil && deps.DB != nil {
		viewSvc := application.NewViewService(deps.DB, deps.Presign, nil)
		viewHandler = documentshttp.NewViewHandler(viewSvc)
	}

	var reconstructHandler *documentshttp.ReconstructHandler
	if deps.ReconstructRunner != nil && deps.DB != nil {
		reconstructSvc := application.NewReconstructionService(deps.DB, deps.ReconstructRunner)
		reconstructHandler = documentshttp.NewReconstructHandler(reconstructSvc)
	}

	return &Module{
		Handler:                   h,
		Service:                   svc,
		ExportHandler:             exportHandler,
		FillInHandler:             fillInHandler,
		PlaceholderOptionsHandler: placeholderOptionsHandler,
		ViewHandler:               viewHandler,
		ReconstructHandler:        reconstructHandler,
		repo:                      repo,
	}
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	legacyMux := m.buildLegacyMux(nil, nil)
	documentsapi.HandlerWithOptions(
		dhttp.NewGeneratedServerAdapter(legacyMux),
		documentsapi.StdHTTPServerOptions{
			BaseRouter: mux,
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
			},
		},
	)
}

func (m *Module) RegisterRoutesWithRateLimit(mux *http.ServeMux, rl *ratelimit.Middleware, userFn func(*http.Request) string) {
	legacyMux := m.buildLegacyMux(rl, userFn)
	documentsapi.HandlerWithOptions(
		dhttp.NewGeneratedServerAdapter(legacyMux),
		documentsapi.StdHTTPServerOptions{
			BaseRouter: mux,
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				_ = problem.Write(w, problem.New(http.StatusBadRequest, "VALIDATION_ERROR", err.Error()))
			},
		},
	)
}

func (m *Module) Repo() *repository.Repository { return m.repo }

func (m *Module) buildLegacyMux(rl *ratelimit.Middleware, userFn func(*http.Request) string) *http.ServeMux {
	legacyMux := http.NewServeMux()
	if rl == nil || userFn == nil {
		m.Handler.RegisterRoutes(legacyMux)
		if m.ExportHandler != nil {
			m.ExportHandler.RegisterRoutes(legacyMux)
		}
	} else {
		m.Handler.RegisterRoutesWithRateLimit(legacyMux, rl, userFn)
		if m.ExportHandler != nil {
			m.ExportHandler.RegisterRoutesWithRateLimit(legacyMux, rl, userFn)
		}
	}

	m.FillInHandler.RegisterRoutes(legacyMux)
	if m.PlaceholderOptionsHandler != nil {
		m.PlaceholderOptionsHandler.RegisterRoutes(legacyMux)
	}
	if m.ViewHandler != nil {
		m.ViewHandler.RegisterRoutes(legacyMux)
	}
	if m.ReconstructHandler != nil {
		m.ReconstructHandler.RegisterRoutes(legacyMux)
	}
	return legacyMux
}

type placeholderOptionsIAMAdapter struct {
	reader application.IAMUserOptionsReader
}

func newPlaceholderOptionsIAMAdapter(reader application.IAMUserOptionsReader) *placeholderOptionsIAMAdapter {
	return &placeholderOptionsIAMAdapter{reader: reader}
}

func (a *placeholderOptionsIAMAdapter) ListUserOptions(ctx context.Context, tenantID string) ([]documentshttp.UserOptionView, error) {
	if a.reader == nil {
		return []documentshttp.UserOptionView{}, nil
	}
	opts, err := a.reader.ListUserOptions(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]documentshttp.UserOptionView, 0, len(opts))
	for _, opt := range opts {
		out = append(out, documentshttp.UserOptionView{
			UserID:      opt.UserID,
			DisplayName: opt.DisplayName,
		})
	}
	return out, nil
}
