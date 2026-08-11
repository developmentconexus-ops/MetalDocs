package approvalhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metaldocs/internal/modules/approval/application"
	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/http/contracts"
)

// InboxHandler returns the paginated list of approval instances awaiting the
// requesting actor's signoff, along with the total count computed in the same
// query/snapshot (T-005) so total never drifts from the returned page.
func (h *Handler) InboxHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	// A3.3: the worklist IS the actor's own queue — a blank actor would ask the
	// repository "what is pending for nobody", which is a visibility question
	// with no correct answer. Fail closed instead.
	actorID, err := actorIDFromRequest(r)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	areaCode := strings.TrimSpace(r.URL.Query().Get("area_code"))

	limit, err := parseInboxLimit(r.URL.Query().Get("limit"))
	if err != nil {
		WriteError(w, r, NewValidationError(err.Error()))
		return
	}

	offset, err := parseInboxOffset(r.URL.Query().Get("offset"))
	if err != nil {
		WriteError(w, r, NewValidationError(err.Error()))
		return
	}

	filter, err := parseInboxFilter(r.URL.Query())
	if err != nil {
		WriteError(w, r, NewValidationError(err.Error()))
		return
	}

	if h.readSvc == nil {
		WriteError(w, r, errors.New("read service not configured"))
		return
	}

	// Single query/tx computes the page and the total together (T-005): the
	// prior two independent queries (ListInboxItems then CountPendingForActor)
	// could observe different snapshots if a signoff committed in between,
	// producing total < len(items) or vice versa on the wire. ListWorklist
	// (F8, spec.md §4/W4, §6.3 P2/P3/P8) is a superset of the pre-F8
	// ListInboxItemsWithTotal contract — an empty filter behaves identically.
	views, total, err := h.readSvc.ListWorklist(r.Context(), h.runner, tenantID, actorID, areaCode, filter, limit, offset)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	respItems := make([]contracts.InboxItem, 0, len(views))
	for i := range views {
		v := views[i]
		var dueAt *string
		if v.DueAt != nil {
			s := v.DueAt.UTC().Format(time.RFC3339)
			dueAt = &s
		}
		// controlled_document_id is required-and-nullable: emit the uuid for
		// document rows, explicit null (nil pointer) for template rows (empty).
		var controlledDocID *string
		if v.ControlledDocumentID != "" {
			cd := v.ControlledDocumentID
			controlledDocID = &cd
		}
		// controlled_document_code (F-QA4-8) is required-and-nullable on the
		// same terms: the canonical human code for document rows, explicit null
		// when the subject has no controlled document (template rows).
		var controlledDocCode *string
		if v.ControlledDocumentCode != "" {
			code := v.ControlledDocumentCode
			controlledDocCode = &code
		}
		respItems = append(respItems, contracts.InboxItem{
			InstanceID:             v.InstanceID,
			SubjectKind:            v.SubjectKind,
			SubjectKey:             v.SubjectKey,
			SubjectTitle:           v.SubjectTitle,
			SubjectRef:             v.SubjectRef,
			ControlledDocumentID:   controlledDocID,
			ControlledDocumentCode: controlledDocCode,
			AreaCode:               v.AreaCode,
			SubmittedBy:            v.SubmittedBy,
			SubmittedAt:            v.SubmittedAt.UTC().Format(time.RFC3339),
			StageLabel:             v.StageLabel,
			QuorumProgress:         v.QuorumProgress,
			StageKind:              string(v.StageKind),
			DueAt:                  dueAt,
		})
	}

	WriteJSON(w, r, http.StatusOK, contracts.InboxResponse{
		Items: respItems,
		Total: total,
	})
}

// parseInboxFilter (F8, spec.md §4/W4, §6.3 P2/P3/P8) parses the worklist
// query params into an application.InboxFilter. Validation mirrors the
// openapi enum constraints (stage_kind, scope) so a malformed value is
// rejected with 400 before reaching the service/repository layer, rather than
// silently matching zero rows.
func parseInboxFilter(q map[string][]string) (application.InboxFilter, error) {
	var filter application.InboxFilter

	if raw := strings.TrimSpace(firstQueryValue(q, "stage_kind")); raw != "" {
		kind := domain.StageKind(raw)
		if kind != domain.StageKindReview && kind != domain.StageKindApproval {
			return filter, fmt.Errorf("stage_kind must be one of: review, approval")
		}
		filter.StageKind = kind
	}

	if raw := strings.TrimSpace(firstQueryValue(q, "due_before")); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return filter, fmt.Errorf("due_before must be an RFC3339 timestamp")
		}
		t = t.UTC()
		filter.DueBefore = &t
	}

	if raw := strings.TrimSpace(firstQueryValue(q, "scope")); raw != "" {
		if raw != "oversee" {
			return filter, fmt.Errorf("scope must be: oversee")
		}
		filter.Oversee = true
	}

	return filter, nil
}

func firstQueryValue(q map[string][]string, key string) string {
	vals := q[key]
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func parseInboxLimit(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 25, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("limit must be an integer")
	}
	if v <= 0 || v > 100 {
		return 0, fmt.Errorf("limit must be between 1 and 100")
	}
	return v, nil
}

func parseInboxOffset(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("offset must be an integer")
	}
	if v < 0 {
		return 0, fmt.Errorf("offset must be >= 0")
	}
	return v, nil
}
