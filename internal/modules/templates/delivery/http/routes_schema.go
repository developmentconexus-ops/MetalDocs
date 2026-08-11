package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	iamdomain "metaldocs/internal/modules/iam/domain"
	templatesapi "metaldocs/internal/modules/templates/api"
	"metaldocs/internal/modules/templates/application"
	"metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/problem"
)

func (h *Handler) updateSchemas(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	actorID, err := userIDFromReq(r)
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}
	templateID := r.PathValue("id")
	versionNum, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidParam, "version must be an integer"))
		return
	}

	if err := h.authz(r, tenantID, "*", string(iamdomain.CapTemplateEdit)); err != nil {
		writeMappedErr(w, r, err)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error()))
		return
	}
	req, perr := decodeUpdateSchemasBody(bodyBytes)
	if perr != nil {
		problem.Respond(w, r, perr)
		return
	}
	metadataSchema, placeholderSchema, perr := remarshalSchemas(req)
	if perr != nil {
		problem.Respond(w, r, perr)
		return
	}

	v, err := h.svc.UpdateSchemas(r.Context(), application.UpdateSchemasCmd{
		TenantID:            tenantID,
		ActorUserID:         actorID,
		TemplateID:          templateID,
		VersionNumber:       versionNum,
		MetadataSchema:      metadataSchema,
		PlaceholderSchema:   placeholderSchema,
		ExpectedLockVersion: int(req.ExpectedLockVersion),
	})
	if err != nil {
		writeMappedErr(w, r, err)
		return
	}

	dto, err := toAPIVersionDTO(v)
	if err != nil {
		problem.Respond(w, r, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error"))
		return
	}
	var resp templatesapi.UpdateTemplateSchema200JSONResponse
	resp.Data.Version = dto
	writeJSON(w, http.StatusOK, resp)
}

// decodeUpdateSchemasBody decodes the generated request type and a raw
// presence probe from the same body bytes: the generated ExpectedLockVersion
// field is a spec-required int32 value type — oapi-codegen never emits
// pointers for required fields, so a present-but-omitted key and an explicit
// 0 both decode to the same zero value. The contract still needs the two
// distinguished (absent must 400 "required", not silently compare-and-swap
// against lock_version 0), hence the second raw-key decode.
func decodeUpdateSchemasBody(bodyBytes []byte) (templatesapi.UpdateTemplateSchemaJSONRequestBody, *problem.Problem) {
	var req templatesapi.UpdateTemplateSchemaJSONRequestBody
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return req, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
	}
	var presence struct {
		ExpectedLockVersion *json.RawMessage `json:"expected_lock_version"`
	}
	if err := json.Unmarshal(bodyBytes, &presence); err != nil {
		return req, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
	}
	if presence.ExpectedLockVersion == nil {
		return req, problem.New(http.StatusBadRequest, codeTplInvalidBody, "expected_lock_version is required")
	}
	if req.ExpectedLockVersion < 0 {
		return req, problem.New(http.StatusBadRequest, codeTplInvalidBody, "expected_lock_version must be >= 0")
	}
	return req, nil
}

// remarshalSchemas converts the generated MetadataSchema/PlaceholderSchema
// freeform map[string]interface{} shapes (the spec models them as open
// objects) into the domain's typed shapes, whose JSON tags match the wire
// field names exactly (internal/modules/templates/domain/schemas.go).
func remarshalSchemas(req templatesapi.UpdateTemplateSchemaJSONRequestBody) (domain.MetadataSchema, []domain.Placeholder, *problem.Problem) {
	var metadataSchema domain.MetadataSchema
	mb, err := json.Marshal(req.MetadataSchema)
	if err != nil {
		return metadataSchema, nil, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error")
	}
	if err := json.Unmarshal(mb, &metadataSchema); err != nil {
		return metadataSchema, nil, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
	}
	var placeholderSchema []domain.Placeholder
	pb, err := json.Marshal(req.PlaceholderSchema)
	if err != nil {
		return metadataSchema, nil, problem.New(http.StatusInternalServerError, codeTplInternalError, "internal server error")
	}
	if err := json.Unmarshal(pb, &placeholderSchema); err != nil {
		return metadataSchema, nil, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
	}
	return metadataSchema, placeholderSchema, nil
}
