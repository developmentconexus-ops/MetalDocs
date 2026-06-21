package documents

import (
	"context"
	"database/sql"
	"net/http"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	approvalapp "metaldocs/internal/modules/documents/approval/application"
	dhttp "metaldocs/internal/modules/documents/delivery/http"
	"metaldocs/internal/modules/documents/repository"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/ratelimit"
)

type approvalSubmitter interface {
	SubmitRevisionForReview(ctx context.Context, runner db.TxRunner, req approvalapp.SubmitRequest) (approvalapp.SubmitResult, error)
}

type Module struct {
	Handler                   *dhttp.Handler
	Service                   *application.Service
	ExportHandler             *dhttp.ExportHandler
	FillInHandler             *dhttp.FillInHandler
	PlaceholderOptionsHandler *dhttp.PlaceholderOptionsHandler
	ViewHandler               *dhttp.ViewHandler
	ReconstructHandler        *dhttp.ReconstructHandler
	repo                      *repository.Repository
}

type Dependencies struct {
	DB                           *sql.DB
	Presign                      application.Presigner
	TplRead                      application.TemplateReader
	FormVal                      application.FormValidator
	Audit                        application.Audit
	ControlledDocumentDuplicator application.ControlledDocumentDuplicator
	Caps                         application.CapabilityChecker
	ProfileDefaults              application.ProfileDefaultTemplateReader
	SnapshotReader               application.SnapshotTemplateReader
	ExportPresign                application.ExportPresigner
	ExportDocgen                 application.DocgenPDFClient
	DocgenVer                    string
	GrammarVer                   string
	ReconstructRunner            application.ReconstructionRunner
	SubmitSvc                    approvalSubmitter
	IAMUserOptions               application.IAMUserOptionsReader
	// DisplayNameReader is the iam-owned port for reading display_name off
	// metaldocs.iam_users without crossing module boundaries (M4/F4.1).
	// When nil, iamdomain.NoopUserDisplayNameReader{} is used automatically.
	DisplayNameReader iamdomain.UserDisplayNameReader
	// CDFieldReader is the controlleddocuments-owned read-port for individual
	// controlled_documents fields (profile_code) consumed without crossing the
	// module boundary (M2/F2.1; ADR-0039 D3(b)). When nil,
	// controlleddocumentsdomain.NoopCDFieldReader{} is used automatically.
	CDFieldReader controlleddocumentsdomain.CDFieldReader
}

func New(deps Dependencies) *Module {
	if deps.Caps == nil {
		panic("documents.New: Caps (CapabilityChecker) is required for handler authorization")
	}
	displayNameReader := deps.DisplayNameReader
	if displayNameReader == nil {
		displayNameReader = iamdomain.NoopUserDisplayNameReader{}
	}
	cdFieldReader := deps.CDFieldReader
	if cdFieldReader == nil {
		cdFieldReader = controlleddocumentsdomain.NoopCDFieldReader{}
	}
	repo := repository.New(deps.DB, displayNameReader, cdFieldReader)
	var svc *application.Service
	if deps.SnapshotReader != nil {
		snapSvc := application.NewSnapshotService(deps.SnapshotReader)
		svc = application.NewServiceWithSnapshot(repo, deps.Presign, deps.TplRead, deps.FormVal, deps.Audit, deps.ProfileDefaults, snapSvc)
	} else {
		svc = application.NewService(repo, deps.Presign, deps.TplRead, deps.FormVal, deps.Audit, deps.ProfileDefaults)
	}
	svc.WithRunner(db.NewTxRunner(deps.DB))
	svc.WithControlledDocumentDuplicator(deps.ControlledDocumentDuplicator)
	h := dhttp.NewHandlerWithSubmit(svc, deps.DB, deps.SubmitSvc).WithCaps(deps.Caps)

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
	fillInSvc := application.NewFillInService(db.NewTxRunner(deps.DB), application.NewSnapshotSchemaReader(deps.DB), fillInRepo).
		WithReader(fillInRepo).
		WithTemplateSchemaReader(application.NewTemplateVersionSchemaReader(deps.DB)).
		WithCDFieldReader(cdFieldReader)
	fillInHandler := dhttp.NewFillInHandler(fillInSvc)
	placeholderOptionsHandler := dhttp.NewPlaceholderOptionsHandler(
		application.NewSnapshotSchemaReader(deps.DB),
		newPlaceholderOptionsIAMAdapter(deps.IAMUserOptions),
	)

	var viewHandler *dhttp.ViewHandler
	if deps.Presign != nil && deps.DB != nil {
		viewSvc := application.NewViewService(db.NewTxRunner(deps.DB), deps.Presign, nil)
		viewHandler = dhttp.NewViewHandler(viewSvc)
	}

	var reconstructHandler *dhttp.ReconstructHandler
	if deps.ReconstructRunner != nil && deps.DB != nil {
		reconstructSvc := application.NewReconstructionService(db.NewTxRunner(deps.DB), deps.ReconstructRunner, cdFieldReader)
		reconstructHandler = dhttp.NewReconstructHandler(reconstructSvc)
	}

	h.WithSubHandlers(exportHandler, fillInHandler, placeholderOptionsHandler, viewHandler, reconstructHandler)

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
	m.Handler.RegisterRoutes(mux)
}

func (m *Module) RegisterRoutesWithRateLimit(mux *http.ServeMux, rl *ratelimit.Middleware, userFn func(*http.Request) string) {
	m.Handler.RegisterRoutesWithRateLimit(mux, rl, userFn)
}

func (m *Module) Repo() *repository.Repository { return m.repo }

type placeholderOptionsIAMAdapter struct {
	reader application.IAMUserOptionsReader
}

func newPlaceholderOptionsIAMAdapter(reader application.IAMUserOptionsReader) *placeholderOptionsIAMAdapter {
	return &placeholderOptionsIAMAdapter{reader: reader}
}

func (a *placeholderOptionsIAMAdapter) ListUserOptions(ctx context.Context, tenantID string) ([]dhttp.UserOptionView, error) {
	if a.reader == nil {
		return []dhttp.UserOptionView{}, nil
	}
	opts, err := a.reader.ListUserOptions(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dhttp.UserOptionView, 0, len(opts))
	for _, opt := range opts {
		out = append(out, dhttp.UserOptionView{
			UserID:      opt.UserID,
			DisplayName: opt.DisplayName,
		})
	}
	return out, nil
}
