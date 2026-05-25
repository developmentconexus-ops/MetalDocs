package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
		panic("templates http: authz function is required")
	}
	return &Handler{svc: svc, authz: authz}
}

func (h *Handler) Register(mux *http.ServeMux) {
	generated := templatesapi.ServerInterfaceWrapper{
		Handler: h,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeErr(w, http.StatusBadRequest, codeTplInvalidRequest, err.Error())
		},
	}

	mux.HandleFunc("GET /api/v1/signed", generated.RedirectSignedUrl)
	mux.HandleFunc("GET /api/v1/templates", generated.ListTemplates)
	mux.Handle("POST /api/v1/templates", h.idempotent("POST /api/v1/templates", generated.CreateTemplate))
	mux.HandleFunc("GET /api/v1/templates/{id}/versions/{n}", generated.GetTemplateVersion)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/docx-upload-url", generated.PresignTemplateDocxUploadUrl)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/schema-upload-url", generated.PresignTemplateSchemaUploadUrl)
	mux.HandleFunc("PUT /api/v1/templates/{id}/versions/{n}/draft", generated.SaveTemplateDraft)
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/publish", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/publish", generated.PublishTemplateVersion))

	mux.HandleFunc("POST /api/v1/templates/{id}/versions", generated.CreateTemplateVersion)
	mux.HandleFunc("PUT /api/v1/templates/{id}/versions/{n}/schema", generated.UpdateTemplateSchema)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/autosave/presign", generated.PresignTemplateAutosave)
	mux.HandleFunc("POST /api/v1/templates/{id}/versions/{n}/autosave/commit", generated.CommitTemplateAutosave)
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/submit", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/submit", generated.SubmitTemplateVersion))
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/review", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/review", generated.ReviewTemplateVersion))
	mux.Handle("POST /api/v1/templates/{id}/versions/{n}/approve", h.idempotent("POST /api/v1/templates/{id}/versions/{n}/approve", generated.ApproveTemplateVersion))
	mux.HandleFunc("POST /api/v1/templates/{id}/archive", generated.ArchiveTemplate)
	mux.HandleFunc("PUT /api/v1/templates/{id}/approval-config", generated.UpsertTemplateApprovalConfig)

	mux.HandleFunc("GET /api/v1/templates/{id}", generated.GetTemplate)
	mux.HandleFunc("GET /api/v1/templates/system/blank", generated.GetSystemBlankTemplate)
	mux.HandleFunc("GET /api/v1/templates/{id}/versions/{n}/docx-url", generated.GetTemplateDocxUrl)
	mux.HandleFunc("GET /api/v1/templates/{id}/audit", generated.ListTemplateAudit)
	mux.HandleFunc("GET /api/v1/templates/placeholder-catalog", generated.ListTemplatePlaceholderCatalog)
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

func writeErr(w http.ResponseWriter, status int, code problem.Code, message string) {
	if err := problem.Write(w, problem.New(status, code, message)); err != nil {
		slog.Warn("templates http: failed to write problem response", "err", err, "status", status, "code", code)
	}
}

var friendlyMsg = map[problem.Code]string{
	codeTplUploadMissing: "DOCX file not yet uploaded. Please upload the template file before submitting for review.",
}

func writeMappedErr(w http.ResponseWriter, err error) {
	status, code := MapErr(err)
	msg := friendlyMsg[code]
	if msg == "" {
		msg = err.Error()
	}
	if msg == "" {
		msg = string(code)
	}
	writeErr(w, status, code, msg)
}
