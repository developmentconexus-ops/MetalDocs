package approvalhttp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metaldocs/internal/modules/documents/approval/http/contracts"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// InboxHandler returns the paginated list of approval instances awaiting the
// requesting actor's signoff, along with the total count computed in the same
// query/snapshot (T-005) so total never drifts from the returned page.
func (h *Handler) InboxHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenantIDFromReq(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	actorID := iamdomain.UserIDFromContext(r.Context())
	areaCode := strings.TrimSpace(r.URL.Query().Get("area_code"))

	limit, err := parseInboxLimit(r.URL.Query().Get("limit"))
	if err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	offset, err := parseInboxOffset(r.URL.Query().Get("offset"))
	if err != nil {
		WriteError(w, NewValidationError(err.Error()))
		return
	}

	if h.readSvc == nil {
		WriteError(w, errors.New("read service not configured"))
		return
	}

	// Single query/tx computes the page and the total together (T-005): the
	// prior two independent queries (ListInboxItems then CountPendingForActor)
	// could observe different snapshots if a signoff committed in between,
	// producing total < len(items) or vice versa on the wire.
	views, total, err := h.readSvc.ListInboxItemsWithTotal(r.Context(), h.runner, tenantID, actorID, areaCode, limit, offset)
	if err != nil {
		WriteError(w, err)
		return
	}

	respItems := make([]contracts.InboxItem, 0, len(views))
	for i := range views {
		v := views[i]
		respItems = append(respItems, contracts.InboxItem{
			InstanceID:           v.InstanceID,
			DocumentID:           v.DocumentID,
			ControlledDocumentID: v.ControlledDocumentID,
			DocumentTitle:        v.DocumentTitle,
			AreaCode:             v.AreaCode,
			SubmittedBy:          v.SubmittedBy,
			SubmittedAt:          v.SubmittedAt.UTC().Format(time.RFC3339),
			StageLabel:           v.StageLabel,
			QuorumProgress:       v.QuorumProgress,
		})
	}

	WriteJSON(w, http.StatusOK, contracts.InboxResponse{
		Items: respItems,
		Total: total,
	})
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
