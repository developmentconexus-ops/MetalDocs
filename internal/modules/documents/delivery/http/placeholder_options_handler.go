package http

import (
	"context"
	"net/http"

	documentsapi "metaldocs/internal/modules/documents/api"
	templatesdomain "metaldocs/internal/modules/templates/domain"
	"metaldocs/internal/platform/tenant"
)

type placeholderOptionsSchemaReader interface {
	LoadPlaceholderSchema(ctx context.Context, tenantID, docID string) ([]templatesdomain.Placeholder, error)
}

// UserOptionView is a local view model for user placeholder options.
type UserOptionView struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

type placeholderOptionsIAMReader interface {
	ListUserOptions(ctx context.Context, tenantID string) ([]UserOptionView, error)
}

// PlaceholderOptionsHandler serves the selectable-options route for a
// choice-typed (select/user) placeholder on a document.
type PlaceholderOptionsHandler struct {
	schema placeholderOptionsSchemaReader
	iam    placeholderOptionsIAMReader
}

// NewPlaceholderOptionsHandler constructs a PlaceholderOptionsHandler backed
// by the given schema reader and IAM user-options reader.
func NewPlaceholderOptionsHandler(schema placeholderOptionsSchemaReader, iam placeholderOptionsIAMReader) *PlaceholderOptionsHandler {
	return &PlaceholderOptionsHandler{schema: schema, iam: iam}
}

// HandleGetOptions returns the available options for a select- or
// user-typed placeholder on a document.
func (h *PlaceholderOptionsHandler) HandleGetOptions(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		writeFillInError(w, r, requestID(r), err)
		return
	}
	docID := r.PathValue("id")
	placeholderID := r.PathValue("pid")

	schema, err := h.schema.LoadPlaceholderSchema(r.Context(), tenantID, docID)
	if err != nil {
		writeFillInError(w, r, requestID(r), err)
		return
	}

	var ph *templatesdomain.Placeholder
	for i := range schema {
		if schema[i].ID == placeholderID {
			ph = &schema[i]
			break
		}
	}
	if ph == nil {
		writeFillInError(w, r, requestID(r), errNotChoicePlaceholder(placeholderID))
		return
	}

	// Typed envelope (M7 F7.4): the option item is polymorphic by placeholder type,
	// so items are boxed opaquely through []any — each keeps its own json tags,
	// preserving the prior wire {"options":[...]} byte-for-byte.
	switch ph.Type {
	case templatesdomain.PHSelect:
		writeFillInJSON(w, r, http.StatusOK, documentsapi.PlaceholderOptionsResponse{Options: toAnySlice(selectOptions(ph.Options))})
	case templatesdomain.PHUser:
		opts, err := h.iam.ListUserOptions(r.Context(), tenantID)
		if err != nil {
			writeFillInError(w, r, requestID(r), err)
			return
		}
		writeFillInJSON(w, r, http.StatusOK, documentsapi.PlaceholderOptionsResponse{Options: toAnySlice(opts)})
	default:
		writeFillInError(w, r, requestID(r), errNotChoicePlaceholder(placeholderID))
	}
}

func selectOptions(values []string) []map[string]string {
	out := make([]map[string]string, len(values))
	for i, v := range values {
		out[i] = map[string]string{"value": v, "display_name": v}
	}
	return out
}

type notChoicePlaceholderError struct{ id string }

func (e notChoicePlaceholderError) Error() string {
	return "not_a_choice_placeholder: " + e.id
}

func errNotChoicePlaceholder(id string) error { return notChoicePlaceholderError{id: id} }
