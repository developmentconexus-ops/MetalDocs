package http

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	registryapi "metaldocs/internal/modules/registry/api"
	"metaldocs/internal/modules/registry/application"
	registrydomain "metaldocs/internal/modules/registry/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/tenant"
)

type tenantContextKey struct{}

type registryService interface {
	Create(ctx context.Context, cmd application.CreateControlledDocumentCmd) (*application.CreateResult, error)
	List(ctx context.Context, tenantID string, filter application.CDFilter) ([]registrydomain.ControlledDocument, error)
	CreateRevision(ctx context.Context, cmd application.CreateRevisionCmd) (*registrydomain.DocumentRef, error)
	Get(ctx context.Context, tenantID, id string) (*registrydomain.ControlledDocument, error)
	Obsolete(ctx context.Context, tenantID, id string) error
	Supersede(ctx context.Context, tenantID, id string) error
	PeekSeq(ctx context.Context, tenantID, profileCode, areaCode string) (int, error)
}

type Handler struct {
	svc           registryService
	db            *sql.DB
	idempCreate   *idempotency.Store
	idempRevision *idempotency.Store
}

func NewHandler(svc *application.RegistryService, db *sql.DB) *Handler {
	return &Handler{
		svc:           svc,
		db:            db,
		idempCreate:   idempotency.New(db, "POST /api/v2/controlled-documents"),
		idempRevision: idempotency.New(db, "POST /api/v2/controlled-documents/{id}/revisions"),
	}
}

// injectTenant is a thin middleware that reads X-Tenant-ID from the request
// header and stores it in the context so the idempotency actor closure can
// access it without a reference to the *http.Request.
func injectTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
		if tid == "" {
			tid = tenant.DevTenantID
		}
		ctx := context.WithValue(r.Context(), tenantContextKey{}, tid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tenantIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(tenantContextKey{}).(string); ok && v != "" {
		return v
	}
	return tenant.DevTenantID
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	actorOf := func(ctx context.Context) (string, string) {
		return tenantIDFromContext(ctx), authn.UserIDFromContext(ctx)
	}

	generated := registryapi.ServerInterfaceWrapper{
		Handler: h,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			httpresponse.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		},
	}

	mux.Handle("POST /api/v2/controlled-documents",
		injectTenant(idempotency.Require(h.idempCreate, actorOf)(http.HandlerFunc(generated.AtomicCreateControlledDocument))))
	mux.Handle("POST /api/v2/controlled-documents/{id}/revisions",
		injectTenant(idempotency.Require(h.idempRevision, actorOf)(http.HandlerFunc(generated.CreateControlledDocumentRevision))))
	mux.HandleFunc("GET /api/v2/controlled-documents/preview-code", generated.PreviewControlledDocumentCode)
	mux.HandleFunc("GET /api/v2/controlled-documents", generated.ListControlledDocuments)
	mux.HandleFunc("GET /api/v2/controlled-documents/{id}", generated.GetControlledDocument)
	mux.HandleFunc("GET /api/v2/controlled-documents/{id}/active-document", generated.GetActiveDocument)
	mux.HandleFunc("PUT /api/v2/controlled-documents/{id}/obsolete", generated.ObsoleteControlledDocument)
	mux.HandleFunc("PUT /api/v2/controlled-documents/{id}/supersede", generated.SupersedeControlledDocument)
}
