package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	controlleddocumentsapi "metaldocs/internal/modules/controlleddocuments/api"
	"metaldocs/internal/modules/controlleddocuments/application"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/iam/authz"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/httpresponse"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

const maxControlledDocumentsJSONBodyBytes int64 = 1 << 20 // 1 MiB
var errTenantIDInvalid = errors.New("controlled_documents: tenant id invalid")

// ListControlledDocuments handles GET /controlled-documents: lists
// controlled documents visible to the caller, filtered by params and
// forward-paginated via the opaque cursor (FD-2).
func (h *Handler) ListControlledDocuments(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.ListControlledDocumentsParams) {
	filter, err := filterFromListParams(params)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, err.Error())
		return
	}

	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	items, hasMore, err := h.svc.List(r.Context(), tenantID, filter)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "invalid cursor")
			return
		}
		h.writeDomainError(w, err)
		return
	}
	respItems, err := controlledDocumentResponses(items)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		c := pagination.EncodeCursor(last.CreatedAt.UTC().Format(time.RFC3339Nano), last.ID)
		nextCursor = &c
	}
	httpresponse.WriteJSON(w, http.StatusOK, controlleddocumentsapi.ListControlledDocuments200JSONResponse{
		Items: respItems,
		Page:  controlleddocumentsapi.CursorPage{NextCursor: nextCursor, HasMore: hasMore},
	})
}

// AtomicCreateControlledDocument handles POST /controlled-documents:
// creates a controlled document and its first document revision
// atomically (ADR 0011). Requires an Idempotency-Key header (enforced by
// the middleware wired in RegisterRoutes).
func (h *Handler) AtomicCreateControlledDocument(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.AtomicCreateControlledDocumentParams) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxControlledDocumentsJSONBodyBytes)
	var req controlleddocumentsapi.CreateAtomicRequest
	if err := decodeStrictJSON(r, &req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, err.Error())
		return
	}
	if field := missingAtomicCreateField(req); field != "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "field "+field+" is required")
		return
	}

	templateVersionID := uuidStringPtr(req.TemplateVersionId)
	overrideTemplateVersionID := uuidStringPtr(req.OverrideTemplateVersionId)
	formData := map[string]any(nil)
	if req.FormData != nil {
		formData = *req.FormData
	}

	actorUserID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		h.writeDomainError(w, application.ErrActorMissing)
		return
	}

	res, err := h.svc.Create(r.Context(), application.CreateControlledDocumentCmd{
		TenantID:                  tenantID,
		ProfileCode:               strings.TrimSpace(req.ProfileCode),
		ProcessAreaCode:           strings.TrimSpace(req.ProcessAreaCode),
		DepartmentCode:            req.DepartmentCode,
		Title:                     strings.TrimSpace(req.Title),
		OwnerUserID:               strings.TrimSpace(req.OwnerUserId),
		ActorUserID:               actorUserID,
		ManualCode:                req.ManualCode,
		ManualCodeReason:          req.ManualCodeReason,
		OverrideTemplateVersionID: overrideTemplateVersionID,
		OverrideTemplateReason:    req.OverrideTemplateReason,
		TemplateVersionID:         templateVersionID,
		DocumentName:              strings.TrimSpace(req.DocumentName),
		FormData:                  formData,
		VisibilityScope:           string(req.Visibility.Scope),
		VisibilityAreaCodes:       req.Visibility.AreaCodes,
		VisibilityUserIDs:         req.Visibility.UserIds,
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	cd, err := controlledDocumentResponse(*res.ControlledDocument)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	doc, err := documentRefResponse(res.DocumentRef)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, controlleddocumentsapi.AtomicCreateResponse{
		ControlledDocument: cd,
		Document:           doc,
	})
}

func decodeStrictJSON(r *http.Request, dst any) error {
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

func missingAtomicCreateField(req controlleddocumentsapi.CreateAtomicRequest) string {
	switch {
	case strings.TrimSpace(req.DocumentName) == "":
		return "document_name"
	case strings.TrimSpace(req.ProfileCode) == "":
		return "profile_code"
	case strings.TrimSpace(req.ProcessAreaCode) == "":
		return "process_area_code"
	case strings.TrimSpace(req.Title) == "":
		return "title"
	case strings.TrimSpace(req.OwnerUserId) == "":
		return "owner_user_id"
	case strings.TrimSpace(string(req.Visibility.Scope)) == "":
		return "visibility.scope"
	default:
		return ""
	}
}

func uuidStringPtr(id *openapi_types.UUID) *string {
	if id == nil {
		return nil
	}
	value := id.String()
	return &value
}

// PreviewControlledDocumentCode handles GET
// /controlled-documents/preview-code: returns the next auto-allocated
// code for (profile_code, area_code) without consuming the sequence.
func (h *Handler) PreviewControlledDocumentCode(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.PreviewControlledDocumentCodeParams) {
	profileCode := strings.TrimSpace(params.ProfileCode)
	areaCode := strings.TrimSpace(params.AreaCode)
	if profileCode == "" || areaCode == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "profile_code and area_code query parameters are required")
		return
	}
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	next, err := h.svc.PeekSeq(r.Context(), tenantID, profileCode, areaCode)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, controlleddocumentsapi.PreviewCodeResponse{
		ProfileCode: strings.ToUpper(profileCode),
		AreaCode:    strings.ToUpper(areaCode),
		NextSeq:     next,
		Code:        controlleddocumentsdomain.AutoCode(profileCode, areaCode, next),
	})
}

// GetControlledDocumentCreationContext handles GET
// /controlled-documents/creation-context: the create form's pre-flight read
// model — active profiles annotated with approval-route readiness, plus only
// the process areas the CALLER holds controlled_documents.create in. The area
// narrowing is server-side (see application.CreationContext); the client never
// gets the full catalog to filter.
func (h *Handler) GetControlledDocumentCreationContext(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	cc, err := h.svc.CreationContext(r.Context(), tenantID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	resp := controlleddocumentsapi.ControlledDocumentCreationContextResponse{
		Profiles: make([]controlleddocumentsapi.CreationContextProfileItem, 0, len(cc.Profiles)),
		Areas:    make([]controlleddocumentsapi.CreationContextAreaItem, 0, len(cc.Areas)),
	}
	for _, p := range cc.Profiles {
		resp.Profiles = append(resp.Profiles, controlleddocumentsapi.CreationContextProfileItem{
			Code:           p.Code,
			Name:           p.Name,
			HasActiveRoute: p.HasActiveRoute,
		})
	}
	for _, a := range cc.Areas {
		resp.Areas = append(resp.Areas, controlleddocumentsapi.CreationContextAreaItem{
			Code: a.Code,
			Name: a.Name,
		})
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

// CreateControlledDocumentRevision handles POST
// /controlled-documents/{id}/revisions: creates a new document revision
// for an existing active controlled document. Requires an
// Idempotency-Key header (enforced by the middleware wired in
// RegisterRoutes).
func (h *Handler) CreateControlledDocumentRevision(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params controlleddocumentsapi.CreateControlledDocumentRevisionParams) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	cdID := r.PathValue("id")
	r.Body = http.MaxBytesReader(w, r.Body, maxControlledDocumentsJSONBodyBytes)
	var body controlleddocumentsapi.CreateRevisionRequest
	if err := decodeStrictJSON(r, &body); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, err.Error())
		return
	}
	if field := missingCreateRevisionField(body); field != "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeValidationError, "field "+field+" is required")
		return
	}
	formData := map[string]any(nil)
	if body.FormData != nil {
		formData = *body.FormData
	}
	ref, err := h.svc.CreateRevision(r.Context(), application.CreateRevisionCmd{
		TenantID:          tenantID,
		CDID:              cdID,
		Name:              strings.TrimSpace(body.Name),
		FormData:          formData,
		TemplateVersionID: uuidStringPtr(body.TemplateVersionId),
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	respRef, err := documentRefResponse(ref)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, controlleddocumentsapi.RevisionResponse{Document: respRef})
}

func missingCreateRevisionField(req controlleddocumentsapi.CreateRevisionRequest) string {
	if strings.TrimSpace(req.Name) == "" {
		return "name"
	}
	return ""
}

// GetControlledDocument handles GET /controlled-documents/{id}: returns
// the controlled document by id, after verifying the caller can read it.
func (h *Handler) GetControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	doc, err := h.svc.Get(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	resp, err := controlledDocumentResponse(*doc)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
		return
	}
	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

// GetActiveDocument handles GET /controlled-documents/{id}/active-document:
// returns the active (or, absent that, most recently published) document
// instance for the controlled document, including approval state and
// in-progress approval instance id when applicable (SEC-03/T-006).
func (h *Handler) GetActiveDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	cdID := r.PathValue("id")

	inst, err := h.svc.GetActiveInstance(r.Context(), tenantID, cdID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	if inst == nil {
		httpresponse.WriteError(w, http.StatusNotFound, codeCDNoActiveInstance, "no active document instance for this controlled document")
		return
	}

	var resp controlleddocumentsapi.ActiveDocumentResponse
	if inst.DocumentID != nil {
		parsed, err := uuid.Parse(*inst.DocumentID)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
			return
		}
		resp.DocumentId = &parsed
	}
	if inst.ContentHash != nil {
		resp.ContentHash = inst.ContentHash
	}
	if inst.RevisionVersion != nil {
		resp.RevisionVersion = inst.RevisionVersion
	}
	if inst.ApprovalState != nil {
		state := controlleddocumentsapi.ActiveDocumentResponseApprovalState(*inst.ApprovalState)
		resp.ApprovalState = &state
	}
	if inst.PublishedDocumentID != nil {
		parsed, err := uuid.Parse(*inst.PublishedDocumentID)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
			return
		}
		resp.PublishedDocumentId = &parsed
	}
	if inst.ApprovalInstanceID != nil {
		parsed, err := uuid.Parse(*inst.ApprovalInstanceID)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
			return
		}
		resp.ApprovalInstanceId = &parsed
	}

	httpresponse.WriteJSON(w, http.StatusOK, resp)
}

// ObsoleteControlledDocument handles PUT
// /controlled-documents/{id}/obsolete: transitions the controlled
// document from active to obsolete. params.IdempotencyKey is consumed by
// the router-level idempotency.Require middleware (handler.go); it is not
// re-read here, mirroring the atomic-create/create-revision routes in this
// same package.
func (h *Handler) ObsoleteControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, _ controlleddocumentsapi.ObsoleteControlledDocumentParams) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	if err := h.svc.Obsolete(r.Context(), tenantID, r.PathValue("id")); err != nil {
		h.writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SupersedeControlledDocument handles PUT
// /controlled-documents/{id}/supersede: transitions the controlled
// document from active to superseded. params.IdempotencyKey is consumed by
// the router-level idempotency.Require middleware (handler.go); it is not
// re-read here, mirroring the atomic-create/create-revision routes in this
// same package.
func (h *Handler) SupersedeControlledDocument(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, _ controlleddocumentsapi.SupersedeControlledDocumentParams) {
	r.SetPathValue("id", id.String())
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}
	if err := h.svc.Supersede(r.Context(), tenantID, r.PathValue("id")); err != nil {
		h.writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func controlledDocumentResponses(docs []controlleddocumentsdomain.ControlledDocument) ([]controlleddocumentsapi.ControlledDocument, error) {
	items := make([]controlleddocumentsapi.ControlledDocument, 0, len(docs))
	for _, doc := range docs {
		item, err := controlledDocumentResponse(doc)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func controlledDocumentResponse(doc controlleddocumentsdomain.ControlledDocument) (controlleddocumentsapi.ControlledDocument, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return controlleddocumentsapi.ControlledDocument{}, err
	}
	tenantID, err := uuid.Parse(doc.TenantID)
	if err != nil {
		return controlleddocumentsapi.ControlledDocument{}, err
	}
	overrideTemplateVersionID, err := optionalUUID(doc.OverrideTemplateVersionID)
	if err != nil {
		return controlleddocumentsapi.ControlledDocument{}, err
	}
	return controlleddocumentsapi.ControlledDocument{
		Id:                        id,
		TenantId:                  tenantID,
		ProfileCode:               doc.ProfileCode,
		ProcessAreaCode:           doc.ProcessAreaCode,
		DepartmentCode:            doc.DepartmentCode,
		Code:                      doc.Code,
		SequenceNum:               doc.SequenceNum,
		Title:                     doc.Title,
		OwnerUserId:               doc.OwnerUserID,
		OverrideTemplateVersionId: overrideTemplateVersionID,
		Status:                    controlleddocumentsapi.ControlledDocumentStatus(doc.Status),
		CreatedAt:                 doc.CreatedAt,
		UpdatedAt:                 doc.UpdatedAt,
		Visibility: controlleddocumentsapi.ControlledDocumentVisibility{
			Scope:     controlleddocumentsapi.ControlledDocumentVisibilityScope(doc.Visibility.Scope),
			AreaCodes: doc.Visibility.AreaCodes,
			UserIds:   doc.Visibility.UserIDs,
		},
	}, nil
}

func documentRefResponse(ref *controlleddocumentsdomain.DocumentRef) (controlleddocumentsapi.DocumentRef, error) {
	id, err := uuid.Parse(ref.ID)
	if err != nil {
		return controlleddocumentsapi.DocumentRef{}, err
	}
	return controlleddocumentsapi.DocumentRef{
		Id:          id,
		ContentHash: ref.ContentHash,
	}, nil
}

func optionalUUID(value *string) (*openapi_types.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// Controlled-documents domain codes.
//
// ADR 0089 step 3: every one of these was a RAW STRING LITERAL at its emit site
// in writeDomainError below — 26 of them, in three competing naming conventions,
// legal only because the old `type Code string` accepted untyped string
// constants. That is the exact hole the closed Code type closes: a code now has
// one declaration site and comes from the registry.
//
// Wire strings and statuses are unchanged here; annex §2.5 renames them in
// execution step 7. RegisterLegacy skips the semantic-family validation these
// pre-taxonomy names would fail.
//
// state.approval_route_missing, PROFILE_NOT_FOUND and PROFILE_ARCHIVED are not
// in this block: they are wire strings the approval and taxonomy modules also
// emit, so their single registration lives in the platform catalog's shared
// block (collisions C-9 and C-15) and the emit sites reference it directly.
var (
	codeCDNoActiveInstance                      = problem.RegisterLegacy("controlleddocuments", "NO_ACTIVE_INSTANCE", 404)
	codeCDNotFound                              = problem.RegisterLegacy("controlleddocuments", "CONTROLLED_DOCUMENT_NOT_FOUND", 404)
	codeCDNotActive                             = problem.RegisterLegacy("controlleddocuments", "CONTROLLED_DOCUMENT_NOT_ACTIVE", 409)
	codeCDActiveRevisionExists                  = problem.RegisterLegacy("controlleddocuments", "ACTIVE_REVISION_ALREADY_EXISTS", 409)
	codeCDCodeTaken                             = problem.RegisterLegacy("controlleddocuments", "CONTROLLED_DOCUMENT_CODE_TAKEN", 409)
	codeCDCodeArchived                          = problem.RegisterLegacy("controlleddocuments", "CONTROLLED_DOCUMENT_CODE_ARCHIVED", 409)
	codeCDManualCodeReasonRequired              = problem.RegisterLegacy("controlleddocuments", "MANUAL_CODE_REASON_REQUIRED", 400)
	codeCDOverrideReasonRequired                = problem.RegisterLegacy("controlleddocuments", "OVERRIDE_REASON_REQUIRED", 400)
	codeCDVisibilityScopeInvalid                = problem.RegisterLegacy("controlleddocuments", "VISIBILITY_SCOPE_INVALID", 400)
	codeCDOverrideTemplateDeleted               = problem.RegisterLegacy("controlleddocuments", "OVERRIDE_TEMPLATE_DELETED", 409)
	codeCDOverrideTemplateNotPublished          = problem.RegisterLegacy("controlleddocuments", "OVERRIDE_TEMPLATE_NOT_PUBLISHED", 409)
	codeCDDictionaryTokenMissing                = problem.RegisterLegacy("controlleddocuments", "DICTIONARY_TOKEN_MISSING", 422)
	codeCDTemplateInvalid                       = problem.RegisterLegacy("controlleddocuments", "template_invalid", 422)
	codeCDTemplateArtifactMissing               = problem.RegisterLegacy("controlleddocuments", "template.artifact_missing", 409)
	codeCDTemplateArtifactInvariantUnconfigured = problem.RegisterLegacy("controlleddocuments", "template.artifact_invariant_unconfigured", 500)
	codeCDCreationContextUnconfigured           = problem.RegisterLegacy("controlleddocuments", "creation_context.unconfigured", 500)
	codeCDProfileNoDefaultTemplate              = problem.RegisterLegacy("controlleddocuments", "PROFILE_NO_DEFAULT_TEMPLATE", 409)
	codeCDDefaultTemplateObsolete               = problem.RegisterLegacy("controlleddocuments", "DEFAULT_TEMPLATE_OBSOLETE", 409)
	codeCDAreaNotFound                          = problem.CodeSharedAreaNotFound
	codeCDAreaArchived                          = problem.CodeSharedAreaArchived
)

func (h *Handler) writeDomainError(w http.ResponseWriter, err error) {
	var capDenied authz.ErrCapDenied
	switch {
	// ADR 0022 tier-2: an in-tx authz.Require denial (e.g. PeekSeq's preview-code
	// create check) is "you lack this capability" — surface it as 403
	// FORBIDDEN_CAPABILITY problem+json, the same client-visible code the documents
	// module emits, never the default 500. (Without this case the wrapped denial
	// fell through to INTERNAL_ERROR.)
	case errors.As(err, &capDenied):
		httpresponse.WriteError(w, http.StatusForbidden, problem.CodeForbiddenCapability, "you do not have the required capability in this area")
	case errors.Is(err, controlleddocumentsdomain.ErrNoActiveInstance):
		httpresponse.WriteError(w, http.StatusNotFound, codeCDNoActiveInstance, "no active document instance for this controlled document")
	case errors.Is(err, controlleddocumentsdomain.ErrCDNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, codeCDNotFound, "controlled document not found")
	case errors.Is(err, controlleddocumentsdomain.ErrCDNotActive):
		httpresponse.WriteError(w, http.StatusConflict, codeCDNotActive, "controlled document is not active")
	case errors.Is(err, controlleddocumentsdomain.ErrActiveRevisionExists):
		httpresponse.WriteError(w, http.StatusConflict, codeCDActiveRevisionExists, "controlled document already has an active revision")
	// Hard creation gate (D2): the profile has no active approval route, so the
	// document could never be submitted. Mirrors the SAME wire contract the
	// submit path already emits — 409 + "state.approval_route_missing"
	// (internal/modules/approval/http/errors.go) — so both surfaces are one
	// contract for the client.
	case errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing):
		httpresponse.WriteError(w, http.StatusConflict, problem.CodeSharedStateApprovalRouteMissing, "profile has no active approval route")
	case errors.Is(err, controlleddocumentsdomain.ErrCDCodeTaken):
		httpresponse.WriteError(w, http.StatusConflict, codeCDCodeTaken, "controlled document code already taken")
	case errors.Is(err, controlleddocumentsdomain.ErrCDArchivedCodeReuse):
		httpresponse.WriteError(w, http.StatusConflict, codeCDCodeArchived, "cannot reuse code from archived controlled document")
	case errors.Is(err, controlleddocumentsdomain.ErrManualCodeReasonRequired):
		httpresponse.WriteError(w, http.StatusBadRequest, codeCDManualCodeReasonRequired, "manual code reason is required")
	case errors.Is(err, controlleddocumentsdomain.ErrOverrideReasonRequired):
		httpresponse.WriteError(w, http.StatusBadRequest, codeCDOverrideReasonRequired, "override reason is required")
	case errors.Is(err, controlleddocumentsdomain.ErrVisibilityScopeInvalid):
		httpresponse.WriteError(w, http.StatusBadRequest, codeCDVisibilityScopeInvalid, "visibility scope is invalid")
	case errors.Is(err, controlleddocumentsdomain.ErrOverrideTemplateDeleted):
		httpresponse.WriteError(w, http.StatusConflict, codeCDOverrideTemplateDeleted, "override template deleted")
	case errors.Is(err, controlleddocumentsdomain.ErrOverrideNotPublished):
		httpresponse.WriteError(w, http.StatusConflict, codeCDOverrideTemplateNotPublished, "override template is not published")
	case errors.Is(err, controlleddocumentsdomain.ErrDictionaryTokenMissing):
		httpresponse.WriteError(w, http.StatusUnprocessableEntity, codeCDDictionaryTokenMissing, "a referenced dictionary token does not exist")
	case errors.Is(err, controlleddocumentsdomain.ErrTemplateProfileMismatch):
		_ = problem.Write(w, problem.New(http.StatusUnprocessableEntity, codeCDTemplateInvalid, "template version does not match the document profile"))
	case errors.Is(err, application.ErrTemplateArtifactMissing):
		httpresponse.WriteError(w, http.StatusConflict, codeCDTemplateArtifactMissing, "template artifact missing")
	case errors.Is(err, application.ErrTemplateArtifactInvariantUnconfigured):
		httpresponse.WriteError(w, http.StatusInternalServerError, codeCDTemplateArtifactInvariantUnconfigured, "template artifact invariant not configured")
	case errors.Is(err, application.ErrCreationContextUnconfigured):
		httpresponse.WriteError(w, http.StatusInternalServerError, codeCDCreationContextUnconfigured, "creation context is not configured")
	case errors.Is(err, application.ErrActorMissing):
		httpresponse.WriteError(w, http.StatusUnauthorized, problem.CodeUnauthenticated, "authentication required")
	case errors.Is(err, errTenantIDInvalid):
		slog.Error("controlled-documents request has invalid tenant id in context",
			"route", "controlledDocuments.writeDomainError",
			"error", err,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
	case errors.Is(err, controlleddocumentsdomain.ErrProfileHasNoDefaultTemplate):
		httpresponse.WriteError(w, http.StatusConflict, codeCDProfileNoDefaultTemplate, "profile has no default template")
	case errors.Is(err, controlleddocumentsdomain.ErrDefaultObsolete):
		httpresponse.WriteError(w, http.StatusConflict, codeCDDefaultTemplateObsolete, "default template is obsolete")
	case errors.Is(err, taxonomydomain.ErrProfileNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, problem.CodeSharedProfileNotFound, "profile not found")
	case errors.Is(err, taxonomydomain.ErrAreaNotFound):
		httpresponse.WriteError(w, http.StatusNotFound, codeCDAreaNotFound, "process area not found")
	case errors.Is(err, taxonomydomain.ErrProfileArchived):
		httpresponse.WriteError(w, http.StatusConflict, problem.CodeSharedProfileArchived, "profile is archived")
	case errors.Is(err, taxonomydomain.ErrAreaArchived):
		httpresponse.WriteError(w, http.StatusConflict, codeCDAreaArchived, "process area is archived")
	default:
		slog.Error("controlled-documents request failed",
			"route", "controlledDocuments.writeDomainError",
			"error", err,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalError, "internal server error")
	}
}

func tenantIDFromRequest(r *http.Request) (string, error) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		return "", err
	}
	if strings.EqualFold(strings.TrimSpace(tenantID), uuid.Nil.String()) {
		return "", errTenantIDInvalid
	}
	return tenantID, nil
}

func filterFromListParams(params controlleddocumentsapi.ListControlledDocumentsParams) (application.CDFilter, error) {
	filter := application.CDFilter{}

	if params.ProfileCode != nil {
		value := strings.TrimSpace(*params.ProfileCode)
		if value != "" {
			filter.ProfileCode = &value
		}
	}
	if params.ProcessAreaCode != nil {
		value := strings.TrimSpace(*params.ProcessAreaCode)
		if value != "" {
			filter.ProcessAreaCode = &value
		}
	}
	if params.DepartmentCode != nil {
		value := strings.TrimSpace(*params.DepartmentCode)
		if value != "" {
			filter.DepartmentCode = &value
		}
	}
	if params.OwnerUserId != nil {
		value := strings.TrimSpace(*params.OwnerUserId)
		if value != "" {
			filter.OwnerUserID = &value
		}
	}
	if params.Q != nil {
		value := strings.TrimSpace(*params.Q)
		if value != "" {
			filter.Query = &value
		}
	}
	if params.Status != nil {
		if !params.Status.Valid() {
			return application.CDFilter{}, errors.New("invalid status value")
		}
		status := controlleddocumentsdomain.CDStatus(*params.Status)
		filter.Status = &status
	}
	if params.Limit != nil {
		if *params.Limit < 1 || *params.Limit > 100 {
			return application.CDFilter{}, errors.New("limit must be between 1 and 100")
		}
		filter.Limit = *params.Limit
	}
	if params.Cursor != nil {
		filter.Cursor = strings.TrimSpace(*params.Cursor)
	}

	return filter, nil
}
