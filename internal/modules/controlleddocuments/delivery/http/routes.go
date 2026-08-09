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
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, err.Error())
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
			httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "invalid cursor")
			return
		}
		h.writeDomainError(w, err)
		return
	}
	respItems, err := controlledDocumentResponses(items)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
// the middleware wired in Mount).
func (h *Handler) AtomicCreateControlledDocument(w http.ResponseWriter, r *http.Request, params controlleddocumentsapi.AtomicCreateControlledDocumentParams) {
	tenantID, err := tenantIDFromRequest(r)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxControlledDocumentsJSONBodyBytes)
	var req controlleddocumentsapi.CreateAtomicRequest
	if err := decodeStrictJSON(r, &req); err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, err.Error())
		return
	}
	if field := missingAtomicCreateField(req); field != "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "field "+field+" is required")
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
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		return
	}
	doc, err := documentRefResponse(res.DocumentRef)
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "profile_code and area_code query parameters are required")
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
// Mount).
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
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, err.Error())
		return
	}
	if field := missingCreateRevisionField(body); field != "" {
		httpresponse.WriteError(w, http.StatusBadRequest, problem.CodeRequestInvalid, "field "+field+" is required")
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
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
		httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
		httpresponse.WriteError(w, http.StatusNotFound, codeCDNotFoundActiveInstance, "no active document instance for this controlled document")
		return
	}

	var resp controlleddocumentsapi.ActiveDocumentResponse
	if inst.DocumentID != nil {
		parsed, err := uuid.Parse(*inst.DocumentID)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
			return
		}
		resp.PublishedDocumentId = &parsed
	}
	if inst.ApprovalInstanceID != nil {
		parsed, err := uuid.Parse(*inst.ApprovalInstanceID)
		if err != nil {
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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
// ADR 0089 execution step 7 (annex §2.5, rows #126-#146). Before step 3 every
// one of these was a RAW STRING LITERAL at its emit site in writeDomainError
// below — 26 of them, in three competing naming conventions, legal only because
// the old `type Code string` accepted untyped string constants. Step 3 gave each
// one declaration site; this step gives each a semantic family from the closed
// set and BINDS its status to the code, so `RegisterLegacy` is gone from this
// file.
//
// Families are never module-named (ADR 0089 decision 4): `template.` and
// `creation_context.` were local dialects that told the client nothing about
// what to do, so they are re-homed to `state.` / `internal.` per annex §1.3.
//
// Three codes moved 400 -> 422 (annex R-20): manual-code reason, override reason
// and visibility scope all PARSE fine and fail a business rule over a supplied
// value, which is `validation.` @422, not a request-shape defect.
//
// state.approval_route_missing, notfound.document_profile,
// state.document_profile_archived, notfound.process_area,
// state.process_area_archived and validation.template_profile_mismatch are not
// declared here: they are wire strings the approval and taxonomy modules also
// emit, so their single registration lives in the platform catalog's shared
// block (annex §2.10 S-5..S-9 plus C-14) and the emit sites reference it
// directly.
var (
	codeCDNotFoundActiveInstance       = problem.Register("controlleddocuments", "notfound.active_document_instance", 404)
	codeCDNotFoundControlledDocument   = problem.Register("controlleddocuments", "notfound.controlled_document", 404)
	codeCDStateNotActive               = problem.Register("controlleddocuments", "state.controlled_document_not_active", 409)
	codeCDStateActiveRevisionExists    = problem.Register("controlleddocuments", "state.active_revision_exists", 409)
	codeCDConflictCodeTaken            = problem.Register("controlleddocuments", "conflict.controlled_document_code_taken", 409)
	codeCDConflictCodeArchived         = problem.Register("controlleddocuments", "conflict.controlled_document_code_archived", 409)
	codeCDValidationManualCodeReason   = problem.Register("controlleddocuments", "validation.manual_code_reason_required", 422)
	codeCDValidationOverrideReason     = problem.Register("controlleddocuments", "validation.override_reason_required", 422)
	codeCDValidationVisibilityScope    = problem.Register("controlleddocuments", "validation.visibility_scope_invalid", 422)
	codeCDStateOverrideTplDeleted      = problem.Register("controlleddocuments", "state.override_template_deleted", 409)
	codeCDStateOverrideTplNotPublished = problem.Register("controlleddocuments", "state.override_template_not_published", 409)
	codeCDValidationDictionaryToken    = problem.Register("controlleddocuments", "validation.dictionary_token_missing", 422)

	// annex §2.5 row #139 / C-14 / R-22: `template_invalid` renames onto
	// `validation.template_profile_mismatch`, which is ALSO the target of taxonomy
	// row #150 (`TEMPLATE_PROFILE_MISMATCH`). Two modules cannot both register one
	// wire string (record()'s duplicate guard panics at init) and a module may not
	// import another module's delivery package, so the collapsed code is declared
	// once in the platform catalog's shared block. Status is unchanged here (422);
	// taxonomy's site moves 409 -> 422.
	codeCDTemplateInvalid = problem.CodeValidationTemplateProfileMismatch

	codeCDStateTemplateArtifactMissing         = problem.Register("controlleddocuments", "state.template_artifact_missing", 409)
	codeCDInternalTemplateArtifactUnconfigured = problem.Register("controlleddocuments", "internal.template_artifact_invariant_unconfigured", 500)
	codeCDInternalCreationContextUnconfigured  = problem.Register("controlleddocuments", "internal.creation_context_unconfigured", 500)
	codeCDStateProfileNoDefaultTemplate        = problem.Register("controlleddocuments", "state.profile_no_default_template", 409)
	codeCDStateDefaultTemplateObsolete         = problem.Register("controlleddocuments", "state.default_template_obsolete", 409)

	codeCDAreaNotFound = problem.CodeNotFoundProcessArea
	codeCDAreaArchived = problem.CodeStateProcessAreaArchived
)

// cdDomainErrorEntry is one row of cdDomainErrorHandlers: match reports
// whether err is the sentinel/type this row handles, write emits that row's
// RFC 9457 response.
type cdDomainErrorEntry struct {
	match func(error) bool
	write func(w http.ResponseWriter, err error)
}

// cdDomainErrorHandlers is writeDomainError's dispatch table, in the same
// order as (and byte-identical in effect to) the switch it replaces. This is
// a 1:1 sentinel->response mapping with no real decision logic — the
// gocyclo cost was almost entirely the case count — so per the wave-2 lint
// brief's guidance ("a switch that drives dispatch can become a map"), it is
// data instead of 26 case branches in one function body. First match wins,
// same as the original switch; errors.As(capDenied) stays a dedicated check
// in writeDomainError itself since it needs a typed target, not an
// errors.Is sentinel.
var cdDomainErrorHandlers = []cdDomainErrorEntry{
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrNoActiveInstance) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusNotFound, codeCDNotFoundActiveInstance, "no active document instance for this controlled document")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrCDNotFound) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusNotFound, codeCDNotFoundControlledDocument, "controlled document not found")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrCDNotActive) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateNotActive, "controlled document is not active")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrActiveRevisionExists) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateActiveRevisionExists, "controlled document already has an active revision")
		},
	},
	{
		// Hard creation gate (D2): the profile has no active approval route, so the
		// document could never be submitted. Mirrors the SAME wire contract the
		// submit path already emits — 409 + "state.approval_route_missing"
		// (internal/modules/approval/http/errors.go) — so both surfaces are one
		// contract for the client.
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrApprovalRouteMissing) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, problem.CodeStateApprovalRouteMissing, "profile has no active approval route")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrCDCodeTaken) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDConflictCodeTaken, "controlled document code already taken")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrCDArchivedCodeReuse) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDConflictCodeArchived, "cannot reuse code from archived controlled document")
		},
	},
	// annex R-20: the three reason/scope rejections moved 400 -> 422 with the
	// rename, so the status now comes from the registration (NewFor) instead of
	// being restated at the call site — ADR 0089 decision 3.
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrManualCodeReasonRequired) },
		write: func(w http.ResponseWriter, _ error) {
			_ = problem.Write(w, problem.NewFor(codeCDValidationManualCodeReason, "manual code reason is required"))
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrOverrideReasonRequired) },
		write: func(w http.ResponseWriter, _ error) {
			_ = problem.Write(w, problem.NewFor(codeCDValidationOverrideReason, "override reason is required"))
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrVisibilityScopeInvalid) },
		write: func(w http.ResponseWriter, _ error) {
			_ = problem.Write(w, problem.NewFor(codeCDValidationVisibilityScope, "visibility scope is invalid"))
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrOverrideTemplateDeleted) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateOverrideTplDeleted, "override template deleted")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrOverrideNotPublished) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateOverrideTplNotPublished, "override template is not published")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrDictionaryTokenMissing) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusUnprocessableEntity, codeCDValidationDictionaryToken, "a referenced dictionary token does not exist")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrTemplateProfileMismatch) },
		write: func(w http.ResponseWriter, _ error) {
			_ = problem.Write(w, problem.NewFor(codeCDTemplateInvalid, "template version does not match the document profile"))
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, application.ErrTemplateArtifactMissing) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateTemplateArtifactMissing, "template artifact missing")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, application.ErrTemplateArtifactInvariantUnconfigured) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusInternalServerError, codeCDInternalTemplateArtifactUnconfigured, "template artifact invariant not configured")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, application.ErrCreationContextUnconfigured) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusInternalServerError, codeCDInternalCreationContextUnconfigured, "creation context is not configured")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, application.ErrActorMissing) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusUnauthorized, problem.CodeAuthUnauthenticated, "authentication required")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, errTenantIDInvalid) },
		write: func(w http.ResponseWriter, err error) {
			slog.Error("controlled-documents request has invalid tenant id in context",
				"route", "controlledDocuments.writeDomainError",
				"error", err,
			)
			httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrProfileHasNoDefaultTemplate) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateProfileNoDefaultTemplate, "profile has no default template")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, controlleddocumentsdomain.ErrDefaultObsolete) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDStateDefaultTemplateObsolete, "default template is obsolete")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, taxonomydomain.ErrProfileNotFound) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusNotFound, problem.CodeNotFoundDocumentProfile, "profile not found")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, taxonomydomain.ErrAreaNotFound) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusNotFound, codeCDAreaNotFound, "process area not found")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, taxonomydomain.ErrProfileArchived) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, problem.CodeStateDocumentProfileArchived, "profile is archived")
		},
	},
	{
		match: func(err error) bool { return errors.Is(err, taxonomydomain.ErrAreaArchived) },
		write: func(w http.ResponseWriter, _ error) {
			httpresponse.WriteError(w, http.StatusConflict, codeCDAreaArchived, "process area is archived")
		},
	},
}

func (h *Handler) writeDomainError(w http.ResponseWriter, err error) {
	// ADR 0022 tier-2: an in-tx authz.Require denial (e.g. PeekSeq's preview-code
	// create check) is "you lack this capability" — surface it as 403
	// permission.capability_denied problem+json, the same client-visible code the documents
	// module emits, never the default 500. (Without this case the wrapped denial
	// fell through to internal.unknown.)
	var capDenied authz.ErrCapDenied
	if errors.As(err, &capDenied) {
		httpresponse.WriteError(w, http.StatusForbidden, problem.CodePermissionCapabilityDenied, "you do not have the required capability in this area")
		return
	}
	for _, entry := range cdDomainErrorHandlers {
		if entry.match(err) {
			entry.write(w, err)
			return
		}
	}
	slog.Error("controlled-documents request failed",
		"route", "controlledDocuments.writeDomainError",
		"error", err,
	)
	httpresponse.WriteError(w, http.StatusInternalServerError, problem.CodeInternalUnknown, "internal server error")
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

// trimmedOptionalFilter trims value and returns a pointer to the trimmed
// string, or nil when value is nil or blank after trimming. Shared by every
// plain string filter field in filterFromListParams — extracted so each
// field no longer needs its own nested nil/blank-check pair.
func trimmedOptionalFilter(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// applyListLimit validates limit (1-100) and, when present, copies it onto
// filter.Limit. Extracted from filterFromListParams; behavior unchanged.
func applyListLimit(filter *application.CDFilter, limit *int) error {
	if limit == nil {
		return nil
	}
	if *limit < 1 || *limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	filter.Limit = *limit
	return nil
}

func filterFromListParams(params controlleddocumentsapi.ListControlledDocumentsParams) (application.CDFilter, error) {
	filter := application.CDFilter{
		ProfileCode:     trimmedOptionalFilter(params.ProfileCode),
		ProcessAreaCode: trimmedOptionalFilter(params.ProcessAreaCode),
		DepartmentCode:  trimmedOptionalFilter(params.DepartmentCode),
		OwnerUserID:     trimmedOptionalFilter(params.OwnerUserId),
		Query:           trimmedOptionalFilter(params.Q),
	}

	if params.Status != nil {
		if !params.Status.Valid() {
			return application.CDFilter{}, errors.New("invalid status value")
		}
		status := controlleddocumentsdomain.CDStatus(*params.Status)
		filter.Status = &status
	}
	if err := applyListLimit(&filter, params.Limit); err != nil {
		return application.CDFilter{}, err
	}
	if params.Cursor != nil {
		filter.Cursor = strings.TrimSpace(*params.Cursor)
	}

	return filter, nil
}
