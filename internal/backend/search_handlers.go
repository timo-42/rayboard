package backend

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
	"github.com/timo-42/rayboard/internal/backend/search"
)

type searchRoute struct {
	search *search.Service
}

func registerSearchRoutes(mux *http.ServeMux, authService *auth.Service, searchService *search.Service) {
	authRoute := authRoute{auth: authService}
	route := searchRoute{search: searchService}

	mux.HandleFunc("POST /api/search", authRoute.requireAuth(route.searchTickets))
	mux.HandleFunc("GET /api/saved-views", authRoute.requireAuth(route.listSavedViews))
	mux.HandleFunc("POST /api/saved-views", authRoute.requireAuth(route.createSavedView))
	mux.HandleFunc("GET /api/saved-views/{view_id}", authRoute.requireAuth(route.getSavedView))
	mux.HandleFunc("PATCH /api/saved-views/{view_id}", authRoute.requireAuth(route.updateSavedView))
	mux.HandleFunc("DELETE /api/saved-views/{view_id}", authRoute.requireAuth(route.deleteSavedView))
}

func (route searchRoute) searchTickets(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request search.SearchTicketsInput
	if !decodeJSON(w, r, &request) {
		return
	}
	result, err := route.search.SearchTickets(r.Context(), principal, request)
	if err != nil {
		writeSearchError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, result)
}

func (route searchRoute) listSavedViews(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	limit, offset, ok := savedViewWindowFromQuery(w, r)
	if !ok {
		return
	}
	views, err := route.search.ListSavedViews(r.Context(), principal, search.ListSavedViewsInput{
		ProjectID: r.URL.Query().Get("project_id"),
		Pinned:    r.URL.Query().Get("pinned") == "true",
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		writeSearchError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": views})
}

func (route searchRoute) createSavedView(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request search.CreateSavedViewInput
	if !decodeJSON(w, r, &request) {
		return
	}
	view, err := route.search.CreateSavedView(r.Context(), principal, request)
	if err != nil {
		writeSearchError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, view)
}

func (route searchRoute) getSavedView(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	view, err := route.search.GetSavedView(r.Context(), principal, r.PathValue("view_id"))
	if err != nil {
		writeSearchError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, view)
}

func (route searchRoute) updateSavedView(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request search.UpdateSavedViewInput
	if !decodeJSON(w, r, &request) {
		return
	}
	view, err := route.search.UpdateSavedView(r.Context(), principal, r.PathValue("view_id"), request)
	if err != nil {
		writeSearchError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, view)
}

func (route searchRoute) deleteSavedView(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.search.DeleteSavedView(r.Context(), principal, r.PathValue("view_id")); err != nil {
		writeSearchError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func savedViewWindowFromQuery(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, ok := optionalIntQuery(w, r, "limit")
	if !ok {
		return 0, 0, false
	}
	offset, ok := optionalIntQuery(w, r, "offset")
	if !ok {
		return 0, 0, false
	}
	return limit, offset, true
}

func optionalIntQuery(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Query parameter must be an integer", map[string]string{name: "Must be an integer"})
		return 0, false
	}
	return parsed, true
}

func writeSearchError(w http.ResponseWriter, err error) {
	var validation *search.ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, search.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, search.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, search.ErrConflict):
		httpjson.Error(w, http.StatusConflict, "conflict", "Resource already exists", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
}
