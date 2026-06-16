package backend

import (
	"errors"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/comments"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
)

type commentRoute struct {
	comments *comments.Service
}

type createCommentRequest struct {
	Body string `json:"body"`
}

func registerCommentRoutes(mux *http.ServeMux, authService *auth.Service, commentService *comments.Service) {
	authRoute := authRoute{auth: authService}
	route := commentRoute{comments: commentService}

	mux.HandleFunc("GET /api/tickets/{ticket_id}/comments", authRoute.requireAuth(route.list))
	mux.HandleFunc("POST /api/tickets/{ticket_id}/comments", authRoute.requireAuth(route.create))
	mux.HandleFunc("DELETE /api/comments/{comment_id}", authRoute.requireAuth(route.delete))
}

func (route commentRoute) list(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	items, err := route.comments.List(r.Context(), principal, r.PathValue("ticket_id"))
	if err != nil {
		writeCommentError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": items})
}

func (route commentRoute) create(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request createCommentRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	comment, err := route.comments.Create(r.Context(), principal, comments.CreateInput{
		TicketID: r.PathValue("ticket_id"),
		Body:     request.Body,
	})
	if err != nil {
		writeCommentError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, comment)
}

func (route commentRoute) delete(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.comments.Delete(r.Context(), principal, r.PathValue("comment_id")); err != nil {
		writeCommentError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeCommentError(w http.ResponseWriter, err error) {
	var validation *comments.ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, comments.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, comments.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
}
