package backend

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/httpjson"
	"github.com/timo-42/rayboard/internal/backend/tracker"
)

type trackerRoute struct {
	tracker *tracker.Service
}

func registerTrackerRoutes(mux *http.ServeMux, authService *auth.Service, trackerService *tracker.Service) {
	authRoute := authRoute{auth: authService}
	route := trackerRoute{tracker: trackerService}

	mux.HandleFunc("GET /api/projects", authRoute.requireAuth(route.listProjects))
	mux.HandleFunc("POST /api/projects", authRoute.requireAuth(route.createProject))
	mux.HandleFunc("GET /api/projects/{project_id}", authRoute.requireAuth(route.getProject))
	mux.HandleFunc("GET /api/projects/{project_id}/backlog", authRoute.requireAuth(route.listBacklog))
	mux.HandleFunc("PATCH /api/projects/{project_id}/backlog", authRoute.requireAuth(route.reorderBacklog))
	mux.HandleFunc("GET /api/projects/{project_id}/statuses", authRoute.requireAuth(route.listProjectStatuses))
	mux.HandleFunc("PUT /api/projects/{project_id}/statuses", authRoute.requireAuth(route.replaceProjectStatuses))
	mux.HandleFunc("GET /api/projects/{project_id}/boards", authRoute.requireAuth(route.listBoards))
	mux.HandleFunc("POST /api/projects/{project_id}/boards", authRoute.requireAuth(route.createBoard))
	mux.HandleFunc("GET /api/projects/{project_id}/components", authRoute.requireAuth(route.listComponents))
	mux.HandleFunc("POST /api/projects/{project_id}/components", authRoute.requireAuth(route.createComponent))
	mux.HandleFunc("GET /api/projects/{project_id}/versions", authRoute.requireAuth(route.listVersions))
	mux.HandleFunc("POST /api/projects/{project_id}/versions", authRoute.requireAuth(route.createVersion))
	mux.HandleFunc("GET /api/projects/{project_id}/custom-fields", authRoute.requireAuth(route.listCustomFields))
	mux.HandleFunc("POST /api/projects/{project_id}/custom-fields", authRoute.requireAuth(route.createCustomField))
	mux.HandleFunc("GET /api/projects/{project_id}/tickets", authRoute.requireAuth(route.listTickets))
	mux.HandleFunc("POST /api/projects/{project_id}/tickets", authRoute.requireAuth(route.createTicket))
	mux.HandleFunc("GET /api/projects/{project_id}/roadmap", authRoute.requireAuth(route.listRoadmap))
	mux.HandleFunc("GET /api/projects/{project_id}/sprints", authRoute.requireAuth(route.listSprints))
	mux.HandleFunc("POST /api/projects/{project_id}/sprints", authRoute.requireAuth(route.createSprint))
	mux.HandleFunc("GET /api/tickets/{ticket_id}", authRoute.requireAuth(route.getTicket))
	mux.HandleFunc("PATCH /api/tickets/{ticket_id}", authRoute.requireAuth(route.updateTicket))
	mux.HandleFunc("GET /api/tickets/{ticket_id}/activity", authRoute.requireAuth(route.listTicketActivity))
	mux.HandleFunc("PUT /api/tickets/{ticket_id}/sprint", authRoute.requireAuth(route.assignTicketSprint))
	mux.HandleFunc("PATCH /api/tickets/{ticket_id}/sprint", authRoute.requireAuth(route.assignTicketSprint))
	mux.HandleFunc("DELETE /api/tickets/{ticket_id}/sprint", authRoute.requireAuth(route.removeTicketSprint))
	mux.HandleFunc("GET /api/sprints/{sprint_id}", authRoute.requireAuth(route.getSprint))
	mux.HandleFunc("PATCH /api/sprints/{sprint_id}", authRoute.requireAuth(route.updateSprint))
	mux.HandleFunc("DELETE /api/sprints/{sprint_id}", authRoute.requireAuth(route.deleteSprint))
	mux.HandleFunc("POST /api/sprints/{sprint_id}/start", authRoute.requireAuth(route.startSprint))
	mux.HandleFunc("POST /api/sprints/{sprint_id}/complete", authRoute.requireAuth(route.completeSprint))
	mux.HandleFunc("GET /api/boards/{board_id}", authRoute.requireAuth(route.getBoard))
	mux.HandleFunc("PATCH /api/boards/{board_id}", authRoute.requireAuth(route.updateBoard))
	mux.HandleFunc("DELETE /api/boards/{board_id}", authRoute.requireAuth(route.deleteBoard))
	mux.HandleFunc("GET /api/boards/{board_id}/tickets", authRoute.requireAuth(route.listBoardTickets))
	mux.HandleFunc("GET /api/components/{component_id}", authRoute.requireAuth(route.getComponent))
	mux.HandleFunc("PATCH /api/components/{component_id}", authRoute.requireAuth(route.updateComponent))
	mux.HandleFunc("DELETE /api/components/{component_id}", authRoute.requireAuth(route.deleteComponent))
	mux.HandleFunc("GET /api/versions/{version_id}", authRoute.requireAuth(route.getVersion))
	mux.HandleFunc("PATCH /api/versions/{version_id}", authRoute.requireAuth(route.updateVersion))
	mux.HandleFunc("DELETE /api/versions/{version_id}", authRoute.requireAuth(route.deleteVersion))
	mux.HandleFunc("GET /api/custom-fields/{field_id}", authRoute.requireAuth(route.getCustomField))
	mux.HandleFunc("PATCH /api/custom-fields/{field_id}", authRoute.requireAuth(route.updateCustomField))
	mux.HandleFunc("DELETE /api/custom-fields/{field_id}", authRoute.requireAuth(route.deleteCustomField))
}

func (route trackerRoute) listProjects(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	limit, offset, ok := listWindowFromQuery(w, r)
	if !ok {
		return
	}
	projects, err := route.tracker.ListProjects(r.Context(), principal, tracker.ListProjectsInput{
		IncludeArchived: r.URL.Query().Get("include_archived") == "true",
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": projects})
}

func (route trackerRoute) createProject(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateProjectInput
	if !decodeJSON(w, r, &request) {
		return
	}
	project, err := route.tracker.CreateProject(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, project)
}

func (route trackerRoute) getProject(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	project, err := route.tracker.GetProject(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, project)
}

func (route trackerRoute) listBacklog(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	tickets, err := route.tracker.ListBacklog(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": tickets})
}

func (route trackerRoute) reorderBacklog(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.ReorderBacklogInput
	if !decodeJSON(w, r, &request) {
		return
	}
	tickets, err := route.tracker.ReorderBacklog(r.Context(), principal, r.PathValue("project_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": tickets})
}

func (route trackerRoute) listProjectStatuses(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	statuses, err := route.tracker.ListProjectStatuses(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": statuses})
}

func (route trackerRoute) replaceProjectStatuses(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.ReplaceProjectStatusesInput
	if !decodeJSON(w, r, &request) {
		return
	}
	statuses, err := route.tracker.ReplaceProjectStatuses(r.Context(), principal, r.PathValue("project_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": statuses})
}

func (route trackerRoute) listBoards(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	boards, err := route.tracker.ListBoards(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": boards})
}

func (route trackerRoute) createBoard(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateBoardInput
	if !decodeJSON(w, r, &request) {
		return
	}
	request.ProjectID = r.PathValue("project_id")
	board, err := route.tracker.CreateBoard(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, board)
}

func (route trackerRoute) getBoard(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	board, err := route.tracker.GetBoard(r.Context(), principal, r.PathValue("board_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, board)
}

func (route trackerRoute) updateBoard(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.UpdateBoardInput
	if !decodeJSON(w, r, &request) {
		return
	}
	board, err := route.tracker.UpdateBoard(r.Context(), principal, r.PathValue("board_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, board)
}

func (route trackerRoute) deleteBoard(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.tracker.DeleteBoard(r.Context(), principal, r.PathValue("board_id")); err != nil {
		writeTrackerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route trackerRoute) listBoardTickets(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	boardTickets, err := route.tracker.ListBoardTickets(r.Context(), principal, r.PathValue("board_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, boardTickets)
}

func (route trackerRoute) listComponents(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	components, err := route.tracker.ListComponents(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": components})
}

func (route trackerRoute) createComponent(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateComponentInput
	if !decodeJSON(w, r, &request) {
		return
	}
	request.ProjectID = r.PathValue("project_id")
	component, err := route.tracker.CreateComponent(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, component)
}

func (route trackerRoute) getComponent(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	component, err := route.tracker.GetComponent(r.Context(), principal, r.PathValue("component_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, component)
}

func (route trackerRoute) updateComponent(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.UpdateComponentInput
	if !decodeJSON(w, r, &request) {
		return
	}
	component, err := route.tracker.UpdateComponent(r.Context(), principal, r.PathValue("component_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, component)
}

func (route trackerRoute) deleteComponent(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.tracker.DeleteComponent(r.Context(), principal, r.PathValue("component_id")); err != nil {
		writeTrackerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route trackerRoute) listVersions(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	versions, err := route.tracker.ListVersions(r.Context(), principal, r.PathValue("project_id"), r.URL.Query().Get("status"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": versions})
}

func (route trackerRoute) createVersion(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateVersionInput
	if !decodeJSON(w, r, &request) {
		return
	}
	request.ProjectID = r.PathValue("project_id")
	version, err := route.tracker.CreateVersion(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, version)
}

func (route trackerRoute) getVersion(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	version, err := route.tracker.GetVersion(r.Context(), principal, r.PathValue("version_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, version)
}

func (route trackerRoute) updateVersion(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.UpdateVersionInput
	if !decodeJSON(w, r, &request) {
		return
	}
	version, err := route.tracker.UpdateVersion(r.Context(), principal, r.PathValue("version_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, version)
}

func (route trackerRoute) deleteVersion(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.tracker.DeleteVersion(r.Context(), principal, r.PathValue("version_id")); err != nil {
		writeTrackerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route trackerRoute) listCustomFields(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	fields, err := route.tracker.ListCustomFields(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": fields})
}

func (route trackerRoute) createCustomField(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateCustomFieldInput
	if !decodeJSON(w, r, &request) {
		return
	}
	request.ProjectID = r.PathValue("project_id")
	field, err := route.tracker.CreateCustomField(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, field)
}

func (route trackerRoute) getCustomField(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	field, err := route.tracker.GetCustomField(r.Context(), principal, r.PathValue("field_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, field)
}

func (route trackerRoute) updateCustomField(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.UpdateCustomFieldInput
	if !decodeJSON(w, r, &request) {
		return
	}
	field, err := route.tracker.UpdateCustomField(r.Context(), principal, r.PathValue("field_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, field)
}

func (route trackerRoute) deleteCustomField(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.tracker.DeleteCustomField(r.Context(), principal, r.PathValue("field_id")); err != nil {
		writeTrackerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route trackerRoute) listTickets(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	limit, offset, ok := listWindowFromQuery(w, r)
	if !ok {
		return
	}
	tickets, err := route.tracker.ListTickets(r.Context(), principal, tracker.ListTicketsInput{
		ProjectID:   r.PathValue("project_id"),
		Status:      r.URL.Query().Get("status"),
		AssigneeID:  r.URL.Query().Get("assignee_id"),
		SprintID:    r.URL.Query().Get("sprint_id"),
		ComponentID: r.URL.Query().Get("component_id"),
		VersionID:   r.URL.Query().Get("version_id"),
		Label:       r.URL.Query().Get("label"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": tickets})
}

func (route trackerRoute) createTicket(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateTicketInput
	if !decodeJSON(w, r, &request) {
		return
	}
	request.ProjectID = r.PathValue("project_id")
	ticket, err := route.tracker.CreateTicket(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, ticket)
}

func (route trackerRoute) getTicket(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	ticket, err := route.tracker.GetTicket(r.Context(), principal, r.PathValue("ticket_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, ticket)
}

func (route trackerRoute) updateTicket(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.UpdateTicketInput
	if !decodeJSON(w, r, &request) {
		return
	}
	ticket, err := route.tracker.UpdateTicket(r.Context(), principal, r.PathValue("ticket_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, ticket)
}

func (route trackerRoute) listTicketActivity(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	activities, err := route.tracker.ListTicketActivity(r.Context(), principal, r.PathValue("ticket_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": activities})
}

func (route trackerRoute) listRoadmap(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	items, err := route.tracker.ListRoadmap(r.Context(), principal, r.PathValue("project_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": items})
}

func (route trackerRoute) listSprints(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	sprints, err := route.tracker.ListSprints(r.Context(), principal, r.PathValue("project_id"), r.URL.Query().Get("state"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, map[string]any{"items": sprints})
}

func (route trackerRoute) createSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.CreateSprintInput
	if !decodeJSON(w, r, &request) {
		return
	}
	request.ProjectID = r.PathValue("project_id")
	sprint, err := route.tracker.CreateSprint(r.Context(), principal, request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusCreated, sprint)
}

func (route trackerRoute) getSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	sprint, err := route.tracker.GetSprint(r.Context(), principal, r.PathValue("sprint_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, sprint)
}

func (route trackerRoute) updateSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request tracker.UpdateSprintInput
	if !decodeJSON(w, r, &request) {
		return
	}
	sprint, err := route.tracker.UpdateSprint(r.Context(), principal, r.PathValue("sprint_id"), request)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, sprint)
}

func (route trackerRoute) deleteSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	if err := route.tracker.DeleteSprint(r.Context(), principal, r.PathValue("sprint_id")); err != nil {
		writeTrackerError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (route trackerRoute) startSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	sprint, err := route.tracker.StartSprint(r.Context(), principal, r.PathValue("sprint_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, sprint)
}

func (route trackerRoute) completeSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	sprint, err := route.tracker.CompleteSprint(r.Context(), principal, r.PathValue("sprint_id"))
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, sprint)
}

type assignTicketSprintRequest struct {
	SprintID string `json:"sprint_id"`
}

func (route trackerRoute) assignTicketSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	var request assignTicketSprintRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	ticket, err := route.tracker.SetTicketSprint(r.Context(), principal, r.PathValue("ticket_id"), request.SprintID)
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, ticket)
}

func (route trackerRoute) removeTicketSprint(w http.ResponseWriter, r *http.Request, principal authz.Principal, _ auth.User) {
	ticket, err := route.tracker.SetTicketSprint(r.Context(), principal, r.PathValue("ticket_id"), "")
	if err != nil {
		writeTrackerError(w, err)
		return
	}
	httpjson.Write(w, http.StatusOK, ticket)
}

func listWindowFromQuery(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, ok := intQuery(w, r, "limit")
	if !ok {
		return 0, 0, false
	}
	offset, ok := intQuery(w, r, "offset")
	if !ok {
		return 0, 0, false
	}
	return limit, offset, true
}

func intQuery(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
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

func writeTrackerError(w http.ResponseWriter, err error) {
	var validation *tracker.ValidationError
	switch {
	case errors.As(err, &validation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", validation.Message, validation.Fields)
	case errors.Is(err, tracker.ErrValidation):
		httpjson.Error(w, http.StatusBadRequest, "validation_failed", "Validation failed", nil)
	case errors.Is(err, tracker.ErrNotFound):
		httpjson.Error(w, http.StatusNotFound, "not_found", "Resource was not found", nil)
	case errors.Is(err, tracker.ErrConflict):
		httpjson.Error(w, http.StatusConflict, "conflict", "Resource already exists", nil)
	case errors.Is(err, authz.ErrForbidden):
		httpjson.Error(w, http.StatusForbidden, "forbidden", "Permission denied", nil)
	default:
		httpjson.Error(w, http.StatusInternalServerError, "internal_error", "Request failed", nil)
	}
}
