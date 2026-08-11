package http

import (
	"bytes"
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
	metadataSchema, placeholderSchema, perr := remarshalSchemas(bodyBytes)
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

// remarshalSchemas decodes the metadata_schema/placeholder_schema JSON
// sub-documents straight out of the raw request body into the domain's typed
// shapes (internal/modules/templates/domain/schemas.go), instead of going
// through the generated request type's freeform
// map[string]interface{}/[]map[string]interface{} representation of those
// fields (UpdateTemplateSchemaJSONBody).
//
// That freeform hop matters: decoding JSON into interface{} always collapses
// numbers to float64 (the same class of bug fixed for comment content in
// documents/delivery/http/handler.go's createComment/updateComment), and
// domain.Placeholder.Default is itself a freeform `any` — a placeholder's
// default value carrying a large external numeric id would silently lose
// precision above 2^53 the moment decodeUpdateSchemasBody's initial decode
// touched it. Reading directly off bodyBytes with UseNumber() keeps such
// values as json.Number, whose exact digit string round-trips through
// json.Marshal unchanged.
func remarshalSchemas(bodyBytes []byte) (domain.MetadataSchema, []domain.Placeholder, *problem.Problem) {
	var probe struct {
		MetadataSchema    json.RawMessage `json:"metadata_schema"`
		PlaceholderSchema json.RawMessage `json:"placeholder_schema"`
	}
	if err := json.Unmarshal(bodyBytes, &probe); err != nil {
		return domain.MetadataSchema{}, nil, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
	}

	var metadataSchema domain.MetadataSchema
	if len(probe.MetadataSchema) > 0 {
		dec := json.NewDecoder(bytes.NewReader(probe.MetadataSchema))
		dec.UseNumber()
		if err := dec.Decode(&metadataSchema); err != nil {
			return metadataSchema, nil, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
		}
	}

	var placeholderSchema []domain.Placeholder
	if len(probe.PlaceholderSchema) > 0 {
		dec := json.NewDecoder(bytes.NewReader(probe.PlaceholderSchema))
		dec.UseNumber()
		if err := dec.Decode(&placeholderSchema); err != nil {
			return metadataSchema, nil, problem.New(http.StatusBadRequest, codeTplInvalidBody, err.Error())
		}
	}
	return metadataSchema, placeholderSchema, nil
}
