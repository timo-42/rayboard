package backend

import (
	"errors"
	"net/http"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
)

type cronRoute struct {
	cron *cronjobs.Service
}

func registerCronRoutes(mux *http.ServeMux, authService *auth.Service, cronService *cronjobs.Service) {
	authRoute := authRoute{auth: authService}
	route := cronRoute{cron: cronService}

	mux.HandleFunc("GET /api/cron-jobs", authRoute.requireAuth(route.listJobs))
	mux.HandleFunc("POST /api/cron-jobs", authRoute.requireAuth(route.createJob))
	mux.HandleFunc("GET /api/cron-jobs/{job_id}", authRoute.requireAuth(route.getJob))
	mux.HandleFunc("PATCH /api/cron-jobs/{job_id}", authRoute.requireAuth(route.updateJob))
	mux.HandleFunc("DELETE /api/cron-jobs/{job_id}", authRoute.requireAuth(route.deleteJob))
	mux.HandleFunc("POST /api/cron-jobs/{job_id}/run", authRoute.requireAuth(route.runJob))
	mux.HandleFunc("GET /api/cron-jobs/{job_id}/runs", authRoute.requireAuth(route.listRuns))
}

func (route cronRoute) listJobs(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	limit, offset, ok := optionalIntWindow(w, r)
	if !ok {
		return
	}
	jobs, err := route.cron.List(r.Context(), principal, cronjobs.ListInput{
		ProjectID: r.URL.Query().Get("project_id"),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		writeCronError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": jobs})
}

func (route cronRoute) createJob(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request cronjobs.CreateInput
	if !decodeJSON(w, r, &request) {
		return
	}
	job, err := route.cron.Create(r.Context(), principal, request)
	if err != nil {
		writeCronError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, job)
}

func (route cronRoute) getJob(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	job, err := route.cron.Get(r.Context(), principal, r.PathValue("job_id"))
	if err != nil {
		writeCronError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, job)
}

func (route cronRoute) updateJob(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request cronjobs.UpdateInput
	if !decodeJSON(w, r, &request) {
		return
	}
	job, err := route.cron.Update(r.Context(), principal, r.PathValue("job_id"), request)
	if err != nil {
		writeCronError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, job)
}

func (route cronRoute) deleteJob(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.cron.Delete(r.Context(), principal, r.PathValue("job_id")); err != nil {
		writeCronError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route cronRoute) runJob(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	run, err := route.cron.RunNow(r.Context(), principal, r.PathValue("job_id"))
	if err != nil && run.ID == "" {
		writeCronError(w, err)
		return
	}
	httpjson.Write(w, http.StatusAccepted, run)
}

func (route cronRoute) listRuns(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	limit, offset, ok := optionalIntWindow(w, r)
	if !ok {
		return
	}
	runs, err := route.cron.ListRuns(r.Context(), principal, r.PathValue("job_id"), limit, offset)
	if err != nil {
		writeCronError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": runs})
}

func optionalIntWindow(w http.ResponseWriter, r *http.Request) (int, int, bool) {
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

func writeCronError(w http.ResponseWriter, err error) {
	var validation *cronjobs.ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, cronjobs.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, cronjobs.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
}
