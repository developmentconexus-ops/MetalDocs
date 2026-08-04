// Package notificationshttp implements the HTTP delivery layer for the
// notifications module. It satisfies the generated notificationsapi.StrictServerInterface
// using the repository for data access. Auth (401/403) is produced upstream by
// tier-1 middleware (CapNotificationRead); this handler only sees authenticated,
// authorized requests. Self-scope (a caller sees only their own rows) is enforced
// by passing the authenticated user id into every repository call, where the SQL
// predicate filters on it.
package notificationshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	notificationsapi "metaldocs/internal/modules/notifications/api"
	notificationsdomain "metaldocs/internal/modules/notifications/domain"
	"metaldocs/internal/platform/authn"
	"metaldocs/internal/platform/pagination"
	"metaldocs/internal/platform/problem"
	"metaldocs/internal/platform/tenant"
)

// Repository is the minimal surface the handler needs from the infrastructure layer.
type Repository interface {
	List(ctx context.Context, tenantID, recipientUserID, statusFilter, cursor string, limit int) (notificationsdomain.NotificationsPage, error)
	UnreadCount(ctx context.Context, tenantID, recipientUserID string) (int, error)
	MarkRead(ctx context.Context, tenantID, notificationID, recipientUserID string) error
	MarkAllRead(ctx context.Context, tenantID, recipientUserID string) (int, error)
}

// Handler implements notificationsapi.StrictServerInterface.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler. repo must not be nil.
func NewHandler(repo Repository) *Handler {
	if repo == nil {
		panic("notificationshttp: repo is required")
	}
	return &Handler{repo: repo}
}

// ListNotifications handles GET /notifications. Self-scoped, newest first.
func (h *Handler) ListNotifications(
	ctx context.Context,
	req notificationsapi.ListNotificationsRequestObject,
) (notificationsapi.ListNotificationsResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.ListNotifications500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	statusFilter := ""
	if req.Params.Status != nil {
		statusFilter = string(*req.Params.Status)
	}
	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	limit := pagination.DefaultLimit
	if req.Params.Limit != nil && *req.Params.Limit > 0 {
		limit = *req.Params.Limit
	}
	limit = pagination.ClampLimit(limit)

	page, err := h.repo.List(ctx, tenantID, userID, statusFilter, cursor, limit)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			return notificationsapi.ListNotifications400ApplicationProblemPlusJSONResponse{
				BadRequestApplicationProblemPlusJSONResponse: notificationsapi.BadRequestApplicationProblemPlusJSONResponse(
					toProblem(problem.New(http.StatusBadRequest, problem.CodeInvalidCursor, "Invalid cursor")),
				),
			}, nil
		}
		slog.Error("notifications.ListNotifications", "error", err)
		return notificationsapi.ListNotifications500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	items := make([]notificationsapi.Notification, len(page.Items))
	for i, n := range page.Items {
		items[i] = toAPINotification(n)
	}

	var nextCursor *string
	if page.HasMore && page.NextCursor != "" {
		nc := page.NextCursor
		nextCursor = &nc
	}

	return notificationsapi.ListNotifications200JSONResponse{
		Items: items,
		Page: notificationsapi.CursorPage{
			HasMore:    page.HasMore,
			NextCursor: nextCursor,
		},
	}, nil
}

// GetNotificationsUnreadCount handles GET /notifications/unread-count.
func (h *Handler) GetNotificationsUnreadCount(
	ctx context.Context,
	req notificationsapi.GetNotificationsUnreadCountRequestObject,
) (notificationsapi.GetNotificationsUnreadCountResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.GetNotificationsUnreadCount500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	count, err := h.repo.UnreadCount(ctx, tenantID, userID)
	if err != nil {
		slog.Error("notifications.GetNotificationsUnreadCount", "error", err)
		return notificationsapi.GetNotificationsUnreadCount500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	return notificationsapi.GetNotificationsUnreadCount200JSONResponse{
		Count: count,
	}, nil
}

// MarkNotificationRead handles POST /notifications/{id}/read. Idempotent, self-scoped.
func (h *Handler) MarkNotificationRead(
	ctx context.Context,
	req notificationsapi.MarkNotificationReadRequestObject,
) (notificationsapi.MarkNotificationReadResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.MarkNotificationRead404ApplicationProblemPlusJSONResponse{
			NotFoundApplicationProblemPlusJSONResponse: notificationsapi.NotFoundApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusNotFound, problem.CodeNotFound, "Notification not found")),
			),
		}, nil
	}

	if err := h.repo.MarkRead(ctx, tenantID, req.Id.String(), userID); err != nil {
		slog.Error("notifications.MarkNotificationRead", "error", err)
		return notificationsapi.MarkNotificationRead500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	return notificationsapi.MarkNotificationRead204Response{}, nil
}

// MarkAllNotificationsRead handles POST /notifications/read-all. Idempotent,
// self-scoped: marks every one of the caller's PENDING/SENT rows READ and
// returns how many were transitioned.
func (h *Handler) MarkAllNotificationsRead(
	ctx context.Context,
	req notificationsapi.MarkAllNotificationsReadRequestObject,
) (notificationsapi.MarkAllNotificationsReadResponseObject, error) {
	tenantID, userID, ok := extractTenantAndUser(ctx)
	if !ok {
		return notificationsapi.MarkAllNotificationsRead500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	updated, err := h.repo.MarkAllRead(ctx, tenantID, userID)
	if err != nil {
		slog.Error("notifications.MarkAllNotificationsRead", "error", err)
		return notificationsapi.MarkAllNotificationsRead500ApplicationProblemPlusJSONResponse{
			InternalServerErrorApplicationProblemPlusJSONResponse: notificationsapi.InternalServerErrorApplicationProblemPlusJSONResponse(
				toProblem(problem.New(http.StatusInternalServerError, problem.CodeInternalError, "Internal server error")),
			),
		}, nil
	}

	return notificationsapi.MarkAllNotificationsRead200JSONResponse{
		Updated: updated,
	}, nil
}

// toAPINotification maps a stored row to the generated wire shape.
func toAPINotification(n notificationsdomain.NotificationRow) notificationsapi.Notification {
	var id openapi_types.UUID
	_ = id.UnmarshalText([]byte(n.ID))
	return notificationsapi.Notification{
		Id:              id,
		RecipientUserId: n.RecipientUserID,
		EventType:       n.EventType,
		ResourceType:    n.ResourceType,
		ResourceId:      n.ResourceID,
		Title:           n.Title,
		Message:         n.Message,
		Status:          notificationsapi.NotificationStatus(n.Status),
		CreatedAt:       n.CreatedAt,
		ReadAt:          n.ReadAt,
	}
}

// extractTenantAndUser reads the authenticated tenant + user from context. Returns
// ok=false when auth context is missing. 401/403 are handled upstream by tier-1
// middleware before the handler is invoked.
func extractTenantAndUser(ctx context.Context) (tenantID, userID string, ok bool) {
	userID, ok = authn.UserIDFromContext(ctx)
	if !ok {
		return "", "", false
	}
	tid, err := tenant.FromContext(ctx)
	if err != nil {
		return "", "", false
	}
	return tid, userID, true
}

// toProblem maps a *problem.Problem to the generated notificationsapi.Problem shape.
func toProblem(p *problem.Problem) notificationsapi.Problem {
	var detail *string
	if p.Detail != "" {
		d := p.Detail
		detail = &d
	}
	var instance *string
	if p.Instance != "" {
		inst := p.Instance
		instance = &inst
	}
	return notificationsapi.Problem{
		Status:   p.Status,
		Title:    p.Title,
		Code:     p.Code.String(),
		Detail:   detail,
		Instance: instance,
	}
}
