package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/idempotency"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

type AuthzFunc func(r *http.Request, tenantID, area string, action string) error

type Handler struct {
	svc   *application.Service
	authz AuthzFunc
}

func New(svc *application.Service, authz AuthzFunc) *Handler {
	if authz == nil {
		authz = func(*http.Request, string, string, string) error { return nil }
	}
	return &Handler{svc: svc, authz: authz}
}

func (h *Handler) Register(mux *http.ServeMux) {
	generated := templatesapi.ServerInterfaceWrapper{
		Handler: h,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeErr(w, http.StatusBadRequest, "invalid_request", err.Error())
		},
	}

	mux.HandleFunc("GET /api/v1/signed", generated.RedirectSignedUrlV2)
	mux.HandleFunc("GET /api/v1/templates", generated.ListTemplatesV2)
	mux.Handle("POST /api/v1/templates", h.idempotent("POST /api/v1/templates", generated.CreateTemplateV2))
	mux.HandleFunc("GET /api/v1/templates/{id}/versions/{n}", generated.GetTemplateVersionV2)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/docx-upload-url", generated.PresignTemplateDocxUploadUrlV2)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/schema-upload-url", generated.PresignTemplateSchemaUploadUrlV2)
	mux.HandleFunc("PUT /api/v1/templates/{id}/versions/{n}/draft", generated.SaveTemplateDraftV2)
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/publish", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/publish", generated.PublishTemplateVersionV2))

	mux.HandleFunc("POST /api/v1/templates/{id}/versions", generated.CreateTemplateVersionV2)
	mux.HandleFunc("PUT /api/v1/templates/{id}/versions/{n}/schema", generated.UpdateTemplateSchemaV2)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/autosave/presign", generated.PresignTemplateAutosaveV2)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/autosave/commit", generated.CommitTemplateAutosaveV2)
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/submit", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/submit", generated.SubmitTemplateVersionV2))
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/review", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/review", generated.ReviewTemplateVersionV2))
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/approve", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/approve", generated.ApproveTemplateVersionV2))
	mux.HandleFunc("POST /api/v1/templates/{id}/archive", generated.ArchiveTemplateV2)
	mux.HandleFunc("PUT /api/v1/templates/{id}/approval-config", generated.UpsertTemplateApprovalConfigV2)

	mux.HandleFunc("GET /api/v1/templates/{id}", generated.GetTemplateV2)
	mux.HandleFunc("GET /api/v1/templates/system/blank", generated.GetSystemBlankTemplate)
	mux.HandleFunc("GET /api/v1/templates/{id}/versions/{n}/docx-url", generated.GetTemplateDocxUrlV2)
	mux.HandleFunc("GET /api/v1/templates/{id}/audit", generated.ListTemplateAuditV2)
	mux.HandleFunc("GET /api/v1/templates/v2/placeholder-catalog", generated.ListTemplatePlaceholderCatalogV2)
}

func (h *Handler) idempotent(routeTemplate string, next http.HandlerFunc) http.Handler {
	db := h.svc.DB()
	if db == nil {
		return http.HandlerFunc(next)
	}
	store := idempotency.New(db, routeTemplate)
	return idempotency.Require(store, func(ctx context.Context) (string, string) {
		tenantID, _ := tenant.FromContext(ctx)
		return tenantID, iamdomain.UserIDFromContext(ctx)
	})(http.HandlerFunc(next))
}

var (
	writeJSON = httpresponse.WriteJSON
	readJSON  = httpresponse.ReadJSON
)

func readStrictJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}
		return err
	}
	return nil
}

func tenantIDFromReq(r *http.Request) (string, error) {
	return tenant.FromContext(r.Context())
}

func userIDFromReq(r *http.Request) string {
	return iamdomain.UserIDFromContext(r.Context())
}

func writeErr(w http.ResponseWriter, status int, code, message string) {
	_ = problem.Write(w, problem.New(status, code, message))
}

var friendlyMsg = map[string]string{
	"upload_missing": "DOCX file not yet uploaded. Please upload the template file before submitting for review.",
}

func writeMappedErr(w http.ResponseWriter, err error) {
	status, code := MapErr(err)
	msg := friendlyMsg[code]
	if msg == "" {
		msg = err.Error()
	}
	if msg == "" {
		msg = code
	}
	writeErr(w, status, code, msg)
}
