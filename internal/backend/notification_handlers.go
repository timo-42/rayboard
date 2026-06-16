package backend

import (
	"errors"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
	"github.com/timo-42/rayboard/internal/backend/notifications"
)

type notificationRoute struct {
	notifications *notifications.Service
}

func registerNotificationRoutes(mux *http.ServeMux, authService *auth.Service, notificationService *notifications.Service) {
	authRoute := authRoute{auth: authService}
	route := notificationRoute{notifications: notificationService}

	mux.HandleFunc("GET /api/notifications", authRoute.requireAuth(route.listNotifications))
	mux.HandleFunc("POST /api/notifications/read-all", authRoute.requireAuth(route.markAllRead))
	mux.HandleFunc("POST /api/notifications/{notification_id}/read", authRoute.requireAuth(route.markRead))
	mux.HandleFunc("POST /api/notifications/{notification_id}/unread", authRoute.requireAuth(route.markUnread))
}

func (route notificationRoute) listNotifications(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	limit, offset, ok := optionalIntWindow(w, r)
	if !ok {
		return
	}
	items, err := route.notifications.List(r.Context(), principal, notifications.ListInput{
		UnreadOnly: r.URL.Query().Get("unread") == "true",
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": items})
}

func (route notificationRoute) markRead(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	notification, err := route.notifications.SetRead(r.Context(), principal, r.PathValue("notification_id"), true)
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, notification)
}

func (route notificationRoute) markUnread(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	notification, err := route.notifications.SetRead(r.Context(), principal, r.PathValue("notification_id"), false)
	if err != nil {
		writeNotificationError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, notification)
}

func (route notificationRoute) markAllRead(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.notifications.MarkAllRead(r.Context(), principal); err != nil {
		writeNotificationError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeNotificationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, notifications.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, notifications.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
}
